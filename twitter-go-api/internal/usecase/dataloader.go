package usecase

import (
	"context"

	"github.com/chanombude/twitter-go-api/internal/db"
)

// populateTweetItems acts as a simple DataLoader to batch fetch authors and parent/original tweets,
// resolving the N+1 query problem.
func populateTweetItems(ctx context.Context, store db.Store, inputs []TweetHydrationInput, viewerID *int64) ([]TweetItem, error) {
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
		if t.ParentID != nil {
			tweetIDsMap[*t.ParentID] = true
		}
		if t.RetweetID != nil {
			tweetIDsMap[*t.RetweetID] = true
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
	users, err := store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
		UserIds:  userIDs,
		ViewerID: vID,
	})
	if err != nil {
		return nil, err
	}

	usersMap := make(map[int64]UserItem)
	for _, rawUser := range users {
		usersMap[rawUser.User.ID] = newUserItemFromDB(rawUser.User, rawUser.IsFollowing)
	}

	// 3. Fetch Referenced Tweets (Parents / Originals)
	var refTweetsMap map[int64]TweetItem
	if len(tweetIDs) > 0 {
		rawRefTweets, err := store.GetTweetsByIDs(ctx, db.GetTweetsByIDsParams{
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
			moreUsers, err := store.GetUsersByIDs(ctx, db.GetUsersByIDsParams{
				UserIds:  missingAuthorIDs,
				ViewerID: vID,
			})
			if err == nil {
				for _, rawUser := range moreUsers {
					usersMap[rawUser.User.ID] = newUserItemFromDB(rawUser.User, rawUser.IsFollowing)
				}
			}
		}

		refTweetsMap = make(map[int64]TweetItem)
		for i, rt := range refTweetsSlice {
			raw := rawRefTweets[i]
			item := newTweetItemFromDB(rt)
			item.Author = usersMap[rt.UserID]
			item.IsLiked = raw.IsLiked
			item.IsRetweeted = raw.IsRetweeted
			item.IsFollowing = raw.IsFollowing
			refTweetsMap[rt.ID] = item
		}
	}

	// 4. Assemble final result
	result := make([]TweetItem, 0, len(inputs))
	for _, in := range inputs {
		t := in.Tweet
		item := newTweetItemFromDB(t)
		item.Author = usersMap[t.UserID]
		item.IsLiked = in.IsLiked
		item.IsRetweeted = in.IsRetweeted
		item.IsFollowing = in.IsFollowing

		if t.ParentID != nil {
			if parentTweet, ok := refTweetsMap[*t.ParentID]; ok {
				item.ParentUsername = &parentTweet.Author.Username
			}
		}

		if t.RetweetID != nil {
			if originalTweet, ok := refTweetsMap[*t.RetweetID]; ok {
				item.OriginalTweet = &originalTweet
			}
		}

		result = append(result, item)
	}

	return result, nil
}
