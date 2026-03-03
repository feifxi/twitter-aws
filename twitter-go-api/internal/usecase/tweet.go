package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/db"
)

type MediaUpload struct {
	Filename    string
	ContentType string
	Reader      interface {
		Read(p []byte) (n int, err error)
	}
}

type CreateTweetInput struct {
	UserID   int64
	Content  *string
	ParentID *int64
	Media    *MediaUpload
}

func (u *Usecase) CreateTweet(ctx context.Context, input CreateTweetInput) (TweetItem, error) {
	trimmedContent := ""
	if input.Content != nil {
		trimmedContent = strings.TrimSpace(*input.Content)
	}

	parentID := sql.NullInt64{Valid: false}
	if input.ParentID != nil {
		if _, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: *input.ParentID}); err != nil {
			return TweetItem{}, err
		}
		parentID = sql.NullInt64{Int64: *input.ParentID, Valid: true}
	}

	mediaType := sql.NullString{String: MediaTypeNone, Valid: true}
	mediaURL := sql.NullString{Valid: false}
	if input.Media != nil {
		contentType := strings.ToLower(input.Media.ContentType)
		switch {
		case strings.HasPrefix(contentType, "image/"):
			mediaType = sql.NullString{String: MediaTypeImage, Valid: true}
		case strings.HasPrefix(contentType, "video/"):
			mediaType = sql.NullString{String: MediaTypeVideo, Valid: true}
		default:
			return TweetItem{}, apperr.BadRequest("only images or videos are allowed")
		}

		uploadedURL, err := u.storage.UploadFile(ctx, input.Media.Reader, input.Media.Filename, contentType)
		if err != nil {
			return TweetItem{}, err
		}
		mediaURL = sql.NullString{String: uploadedURL, Valid: true}
	}

	if trimmedContent == "" && !mediaURL.Valid {
		return TweetItem{}, apperr.BadRequest("tweet must include text or media")
	}

	content := sql.NullString{Valid: false}
	if trimmedContent != "" {
		content = sql.NullString{String: trimmedContent, Valid: true}
	}

	var createdTweet db.Tweet
	var pendingNotification db.Notification
	err := u.store.ExecTxAfterCommit(ctx, func(q *db.Queries) error {
		var err error
		createdTweet, err = q.CreateTweet(ctx, db.CreateTweetParams{
			UserID:    input.UserID,
			Content:   content,
			MediaType: mediaType,
			MediaUrl:  mediaURL,
			ParentID:  parentID,
			RetweetID: sql.NullInt64{Valid: false},
		})
		if err != nil {
			return err
		}

		if parentID.Valid {
			if err := q.IncrementParentReplyCount(ctx, parentID.Int64); err != nil {
				return err
			}

			parentTweet, err := q.GetTweet(ctx, db.GetTweetParams{ID: parentID.Int64})
			if err == nil {
				tweetID := createdTweet.ID
				pendingNotification, _ = u.createNotification(ctx, q, parentTweet.Tweet.UserID, input.UserID, &tweetID, NotifTypeReply)
			}
		}

		if content.Valid {
			tags := extractHashtags(content.String)
			for _, tag := range tags {
				h, err := q.UpsertHashtag(ctx, tag)
				if err != nil {
					return err
				}
				if err := q.LinkTweetHashtag(ctx, db.LinkTweetHashtagParams{TweetID: createdTweet.ID, HashtagID: h.ID}); err != nil {
					return err
				}
			}
		}

		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			u.dispatchNotification(pendingNotification)
		}
	})
	if err != nil {
		if mediaURL.Valid {
			_ = u.storage.DeleteFile(ctx, mediaURL.String)
		}
		return TweetItem{}, err
	}

	return u.GetTweet(ctx, createdTweet.ID, &input.UserID)
}

func (u *Usecase) DeleteTweet(ctx context.Context, userID, tweetID int64) error {
	tweet, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID})
	if err != nil {
		return err
	}
	if tweet.Tweet.UserID != userID {
		return apperr.Forbidden("you can only delete your own tweets")
	}

	mediaURLs, err := u.store.ListMediaUrlsInThread(ctx, tweetID)
	if err != nil {
		return err
	}

	err = u.store.ExecTxAfterCommit(ctx, func(q *db.Queries) error {
		// Collect hashtag usage impact for the full cascade set (root tweet + replies + retweets)
		// before deletion, because tweet_hashtags rows are removed via ON DELETE CASCADE.
		hashtagUsage, err := q.ListHashtagUsageToDecrementForDeleteRoot(ctx, tweetID)
		if err != nil {
			return err
		}

		if tweet.Tweet.RetweetID.Valid {
			_, err := q.DeleteRetweetByUser(ctx, db.DeleteRetweetByUserParams{
				UserID:    userID,
				RetweetID: sql.NullInt64{Int64: tweet.Tweet.RetweetID.Int64, Valid: true},
			})
			if err != nil {
				return err
			}
		} else {
			_, err := q.DeleteTweetByOwner(ctx, db.DeleteTweetByOwnerParams{ID: tweetID, UserID: userID})
			if err != nil {
				return err
			}
		}

		if tweet.Tweet.ParentID.Valid {
			if err := q.DecrementParentReplyCount(ctx, tweet.Tweet.ParentID.Int64); err != nil {
				return err
			}
		}

		for _, impact := range hashtagUsage {
			if err := q.DecrementHashtagUsageBy(ctx, db.DecrementHashtagUsageByParams{
				ID:         impact.HashtagID,
				UsageCount: impact.DecrementBy,
			}); err != nil {
				return err
			}
			if err := q.DeleteUnusedHashtag(ctx, impact.HashtagID); err != nil {
				return err
			}
		}

		return nil
	}, func() {
		seen := make(map[string]struct{}, len(mediaURLs))
		for _, url := range mediaURLs {
			if !url.Valid || url.String == "" {
				continue
			}
			if _, exists := seen[url.String]; exists {
				continue
			}
			seen[url.String] = struct{}{}
			_ = u.storage.DeleteFile(ctx, url.String)
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *Usecase) GetTweet(ctx context.Context, tweetID int64, viewerID *int64) (TweetItem, error) {
	r, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID, ViewerID: nullViewerID(viewerID)})
	if err != nil {
		return TweetItem{}, err
	}
	items, err := u.populateTweetItems(ctx, []TweetHydrationInput{mapGetTweetRow(r)}, viewerID)
	if err != nil || len(items) == 0 {
		return TweetItem{}, err
	}
	return items[0], nil
}

func (u *Usecase) ListReplies(ctx context.Context, tweetID int64, page, size int32, viewerID *int64) ([]TweetItem, error) {
	vID := nullViewerID(viewerID)

	_, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID, ViewerID: vID})
	if err != nil {
		return nil, err
	}

	rows, err := u.store.ListTweetReplies(ctx, db.ListTweetRepliesParams{
		ParentID: sql.NullInt64{Int64: tweetID, Valid: true},
		Limit:    size,
		Offset:   page * size,
		ViewerID: vID,
	})
	if err != nil {
		return nil, err
	}

	return u.populateTweetItems(ctx, mapTweetReplyRows(rows), viewerID)
}

func mapGetTweetRow(row db.GetTweetRow) TweetHydrationInput {
	return TweetHydrationInput{
		Tweet:       row.Tweet,
		IsLiked:     row.IsLiked,
		IsRetweeted: row.IsRetweeted,
		IsFollowing: row.IsFollowing,
	}
}

func mapTweetReplyRows(rows []db.ListTweetRepliesRow) []TweetHydrationInput {
	items := make([]TweetHydrationInput, len(rows))
	for i := range rows {
		items[i] = TweetHydrationInput{
			Tweet:       rows[i].Tweet,
			IsLiked:     rows[i].IsLiked,
			IsRetweeted: rows[i].IsRetweeted,
			IsFollowing: rows[i].IsFollowing,
		}
	}
	return items
}

func (u *Usecase) CountReplies(ctx context.Context, tweetID int64) (int64, error) {
	return u.store.CountTweetReplies(ctx, sql.NullInt64{Int64: tweetID, Valid: true})
}

func (u *Usecase) LikeTweet(ctx context.Context, userID, tweetID int64) error {
	tweet, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID})
	if err != nil {
		return err
	}

	var pendingNotification db.Notification
	err = u.store.ExecTxAfterCommit(ctx, func(q *db.Queries) error {
		liked, err := q.LikeTweet(ctx, db.LikeTweetParams{UserID: userID, TweetID: tweetID})
		if err != nil {
			return err
		}

		if liked {
			id := tweet.Tweet.ID
			pendingNotification, _ = u.createNotification(ctx, q, tweet.Tweet.UserID, userID, &id, NotifTypeLike)
		}
		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			u.dispatchNotification(pendingNotification)
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *Usecase) UnlikeTweet(ctx context.Context, userID, tweetID int64) error {
	if _, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID}); err != nil {
		return err
	}

	_, err := u.store.UnlikeTweet(ctx, db.UnlikeTweetParams{UserID: userID, TweetID: tweetID})
	return err
}

func (u *Usecase) Retweet(ctx context.Context, userID, tweetID int64) (TweetItem, error) {
	targetTweet, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID})
	if err != nil {
		return TweetItem{}, err
	}

	originalTweet := targetTweet
	if targetTweet.Tweet.RetweetID.Valid {
		originalTweet, err = u.store.GetTweet(ctx, db.GetTweetParams{ID: targetTweet.Tweet.RetweetID.Int64})
		if err != nil {
			return TweetItem{}, err
		}
	}

	var created db.CreateRetweetRow
	var pendingNotification db.Notification
	err = u.store.ExecTxAfterCommit(ctx, func(q *db.Queries) error {
		var err error
		created, err = q.CreateRetweet(ctx, db.CreateRetweetParams{
			UserID:    userID,
			RetweetID: sql.NullInt64{Int64: originalTweet.Tweet.ID, Valid: true},
		})
		if err != nil {
			// ON CONFLICT DO NOTHING returns no row for existing retweet.
			if errors.Is(err, sql.ErrNoRows) {
				existing, getErr := q.GetUserRetweet(ctx, db.GetUserRetweetParams{
					UserID:    userID,
					RetweetID: sql.NullInt64{Int64: originalTweet.Tweet.ID, Valid: true},
				})
				if getErr != nil {
					return getErr
				}
				created = db.CreateRetweetRow(existing)
				return nil
			}
			return err
		}

		id := originalTweet.Tweet.ID
		pendingNotification, _ = u.createNotification(ctx, q, originalTweet.Tweet.UserID, userID, &id, NotifTypeRetweet)
		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			u.dispatchNotification(pendingNotification)
		}
	})
	if err != nil {
		return TweetItem{}, err
	}

	return u.GetTweet(ctx, created.ID, &userID)
}

func (u *Usecase) UndoRetweet(ctx context.Context, userID, tweetID int64) error {
	targetTweet, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID})
	if err != nil {
		return err
	}

	originalID := targetTweet.Tweet.ID
	if targetTweet.Tweet.RetweetID.Valid {
		originalID = targetTweet.Tweet.RetweetID.Int64
	}

	_, err = u.store.DeleteRetweetByUser(ctx, db.DeleteRetweetByUserParams{
		UserID:    userID,
		RetweetID: sql.NullInt64{Int64: originalID, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}
