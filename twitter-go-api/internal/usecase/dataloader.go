package usecase

import (
	"context"

	"github.com/chanombude/twitter-go-api/internal/db"
)

// populateTweetItems acts as a simple DataLoader to batch fetch authors and parent/original tweets,
// resolving the N+1 query problem.
func (u *Usecase) populateTweetItems(ctx context.Context, inputs []TweetHydrationInput, viewerID *int64) ([]TweetItem, error) {
	if len(inputs) == 0 {
		return []TweetItem{}, nil
	}

	vID := nullViewerID(viewerID)

	// 1. Collect unique IDs
	userIDsMap := make(map[int64]bool)
	tweetIDsMap := make(map[int64]bool)

	for _, in := range inputs {
		t := in.Tweet
		userIDsMap[t.UserID] = true
		if t.ParentID.Valid {
			tweetIDsMap[t.ParentID.Int64] = true
		}
		if t.RetweetID.Valid {
			tweetIDsMap[t.RetweetID.Int64] = true
		}
	}

	// Convert to slices
	userIDs := make([]int64, 0, len(userIDsMap))
	for id := range userIDsMap {
		userIDs = append(userIDs, id)
	}

	tweetIDs := make([]int64, 0, len(tweetIDsMap))
	for id := range tweetIDsMap {
		tweetIDs = append(tweetIDs, id)
	}

	// 2. Fetch Authors
	users, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		UserIds:  userIDs,
		ViewerID: vID,
	})
	if err != nil {
		return nil, err
	}

	usersMap := make(map[int64]UserItem)
	for _, rawUser := range users {
		usersMap[rawUser.User.ID] = UserItem{User: rawUser.User, IsFollowing: rawUser.IsFollowing}
	}

	// 3. Fetch Referenced Tweets (Parents / Originals)
	var refTweetsMap map[int64]TweetItem
	if len(tweetIDs) > 0 {
		rawRefTweets, err := u.store.GetTweetsByIDs(ctx, db.GetTweetsByIDsParams{
			TweetIds: tweetIDs,
			ViewerID: vID,
		})
		if err != nil {
			return nil, err
		}

		// Because a referenced tweet also needs an author, we recursively call ourselves!
		// But to prevent infinite recursion, we assume ref tweets don't need *their* parents fully populated
		// down an infinite tree, just the immediate parent/original.
		// For a simple DataLoader, one level of depth is usually enough.
		refTweetsSlice := make([]db.Tweet, 0, len(rawRefTweets))
		for _, rt := range rawRefTweets {
			refTweetsSlice = append(refTweetsSlice, rt.Tweet)
		}

		// To avoid deep recursion, we'll manually attach authors to the ref tweets here.
		// We need to fetch any missing authors first.
		missingAuthorsMap := make(map[int64]bool)
		for _, rt := range refTweetsSlice {
			if _, ok := usersMap[rt.UserID]; !ok {
				missingAuthorsMap[rt.UserID] = true
			}
		}

		if len(missingAuthorsMap) > 0 {
			missingAuthorIDs := make([]int64, 0, len(missingAuthorsMap))
			for id := range missingAuthorsMap {
				missingAuthorIDs = append(missingAuthorIDs, id)
			}
			moreUsers, err := u.store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
				UserIds:  missingAuthorIDs,
				ViewerID: vID,
			})
			if err == nil {
				for _, rawUser := range moreUsers {
					usersMap[rawUser.User.ID] = UserItem{User: rawUser.User, IsFollowing: rawUser.IsFollowing}
				}
			}
		}

		refTweetsMap = make(map[int64]TweetItem)
		for i, rt := range refTweetsSlice {
			raw := rawRefTweets[i]
			item := TweetItem{
				Tweet:       rt,
				Author:      usersMap[rt.UserID],
				IsLiked:     raw.IsLiked,
				IsRetweeted: raw.IsRetweeted,
				IsFollowing: raw.IsFollowing,
			}
			refTweetsMap[rt.ID] = item
		}
	}

	// 4. Assemble final result
	result := make([]TweetItem, 0, len(inputs))
	for _, in := range inputs {
		t := in.Tweet
		item := TweetItem{
			Tweet:       t,
			Author:      usersMap[t.UserID],
			IsLiked:     in.IsLiked,
			IsRetweeted: in.IsRetweeted,
			IsFollowing: in.IsFollowing,
		}

		if t.ParentID.Valid {
			if parentTweet, ok := refTweetsMap[t.ParentID.Int64]; ok {
				item.ParentUsername = &parentTweet.Author.Username
			}
		}

		if t.RetweetID.Valid {
			if originalTweet, ok := refTweetsMap[t.RetweetID.Int64]; ok {
				item.OriginalTweet = &originalTweet
			}
		}

		result = append(result, item)
	}

	return result, nil
}
