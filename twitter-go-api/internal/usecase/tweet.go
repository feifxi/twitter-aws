package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/jackc/pgx/v5"
)

type CreateTweetInput struct {
	UserID   int64
	Content  *string
	ParentID *int64
	Media    *FileUpload
}

func (u *TweetUsecase) CreateTweet(ctx context.Context, input CreateTweetInput) (TweetItem, error) {
	trimmedContent := ""
	if input.Content != nil {
		trimmedContent = strings.TrimSpace(*input.Content)
	}

	var mediaType *string
	var mediaURL *string
	if input.Media != nil {
		contentType := strings.ToLower(input.Media.ContentType)
		switch {
		case strings.HasPrefix(contentType, "image/"):
			img := MediaTypeImage
			mediaType = &img
		case strings.HasPrefix(contentType, "video/"):
			vid := MediaTypeVideo
			mediaType = &vid
		default:
			return TweetItem{}, apperr.BadRequest("only images or videos are allowed")
		}

		uploadedURL, err := u.storage.UploadFile(ctx, input.Media.Reader, input.Media.Filename, contentType)
		if err != nil {
			return TweetItem{}, err
		}
		mediaURL = &uploadedURL
	}

	if trimmedContent == "" && mediaURL == nil {
		return TweetItem{}, apperr.BadRequest("tweet must include text or media")
	}

	var content *string
	if trimmedContent != "" {
		content = &trimmedContent
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
			ParentID:  input.ParentID,
			RetweetID: nil,
		})
		if err != nil {
			return err
		}

		if input.ParentID != nil {
			if err := q.IncrementParentReplyCount(ctx, *input.ParentID); err != nil {
				return err
			}

			parentTweet, err := q.GetTweet(ctx, db.GetTweetParams{ID: *input.ParentID})
			if err == nil {
				tweetID := createdTweet.ID
				pendingNotification, _ = createNotification(ctx, q, parentTweet.Tweet.UserID, input.UserID, &tweetID, NotifTypeReply)
			}
		}

		if content != nil {
			tags := extractHashtags(*content)
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
			dispatchNotification(u.publishNotification, pendingNotification)
		}
	})
	if err != nil {
		if mediaURL != nil {
			_ = u.storage.DeleteFile(ctx, *mediaURL)
		}
		return TweetItem{}, err
	}

	return u.GetTweet(ctx, createdTweet.ID, &input.UserID)
}

func (u *TweetUsecase) DeleteTweet(ctx context.Context, userID, tweetID int64) error {
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

		if tweet.Tweet.RetweetID != nil {
			_, err := q.DeleteRetweetByUser(ctx, db.DeleteRetweetByUserParams{
				UserID:    userID,
				RetweetID: tweet.Tweet.RetweetID,
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

		if tweet.Tweet.ParentID != nil {
			if err := q.DecrementParentReplyCount(ctx, *tweet.Tweet.ParentID); err != nil {
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
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		seen := make(map[string]struct{}, len(mediaURLs))
		for _, url := range mediaURLs {
			if url == nil || *url == "" {
				continue
			}
			if _, exists := seen[*url]; exists {
				continue
			}
			seen[*url] = struct{}{}
			_ = u.storage.DeleteFile(cleanupCtx, *url)
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *TweetUsecase) GetTweet(ctx context.Context, tweetID int64, viewerID *int64) (TweetItem, error) {
	r, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID, ViewerID: viewerID})
	if err != nil {
		return TweetItem{}, err
	}
	inputs := mapTweetHydrationRows(
		[]db.GetTweetRow{r},
		func(row db.GetTweetRow) db.Tweet { return row.Tweet },
		func(row db.GetTweetRow) bool { return row.IsLiked },
		func(row db.GetTweetRow) bool { return row.IsRetweeted },
		func(row db.GetTweetRow) bool { return row.IsFollowing },
	)
	items, err := populateTweetItems(ctx, u.store, inputs, viewerID)
	if err != nil || len(items) == 0 {
		return TweetItem{}, err
	}
	return items[0], nil
}

func (u *TweetUsecase) ListReplies(ctx context.Context, tweetID int64, page, size int32, viewerID *int64) ([]TweetItem, error) {
	_, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID, ViewerID: viewerID})
	if err != nil {
		return nil, err
	}

	rows, err := u.store.ListTweetReplies(ctx, db.ListTweetRepliesParams{
		ParentID: &tweetID,
		Limit:    size,
		Offset:   page * size,
		ViewerID: viewerID,
	})
	if err != nil {
		return nil, err
	}

	inputs := mapTweetHydrationRows(
		rows,
		func(row db.ListTweetRepliesRow) db.Tweet { return row.Tweet },
		func(row db.ListTweetRepliesRow) bool { return row.IsLiked },
		func(row db.ListTweetRepliesRow) bool { return row.IsRetweeted },
		func(row db.ListTweetRepliesRow) bool { return row.IsFollowing },
	)
	return populateTweetItems(ctx, u.store, inputs, viewerID)
}

func (u *TweetUsecase) LikeTweet(ctx context.Context, userID, tweetID int64) error {
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
			pendingNotification, _ = createNotification(ctx, q, tweet.Tweet.UserID, userID, &id, NotifTypeLike)
		}
		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			dispatchNotification(u.publishNotification, pendingNotification)
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *TweetUsecase) UnlikeTweet(ctx context.Context, userID, tweetID int64) error {
	if _, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID}); err != nil {
		return err
	}

	_, err := u.store.UnlikeTweet(ctx, db.UnlikeTweetParams{UserID: userID, TweetID: tweetID})
	return err
}

func (u *TweetUsecase) Retweet(ctx context.Context, userID, tweetID int64) (TweetItem, error) {
	targetTweet, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID})
	if err != nil {
		return TweetItem{}, err
	}

	originalTweet := targetTweet
	if targetTweet.Tweet.RetweetID != nil {
		originalTweet, err = u.store.GetTweet(ctx, db.GetTweetParams{ID: *targetTweet.Tweet.RetweetID})
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
			RetweetID: &originalTweet.Tweet.ID,
		})
		if err != nil {
			// ON CONFLICT DO NOTHING returns no row for existing retweet.
			if errors.Is(err, pgx.ErrNoRows) {
				existing, getErr := q.GetUserRetweet(ctx, db.GetUserRetweetParams{
					UserID:    userID,
					RetweetID: &originalTweet.Tweet.ID,
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
		pendingNotification, _ = createNotification(ctx, q, originalTweet.Tweet.UserID, userID, &id, NotifTypeRetweet)
		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			dispatchNotification(u.publishNotification, pendingNotification)
		}
	})
	if err != nil {
		return TweetItem{}, err
	}

	return u.GetTweet(ctx, created.ID, &userID)
}

func (u *TweetUsecase) UndoRetweet(ctx context.Context, userID, tweetID int64) error {
	targetTweet, err := u.store.GetTweet(ctx, db.GetTweetParams{ID: tweetID})
	if err != nil {
		return err
	}

	originalID := targetTweet.Tweet.ID
	if targetTweet.Tweet.RetweetID != nil {
		originalID = *targetTweet.Tweet.RetweetID
	}

	_, err = u.store.DeleteRetweetByUser(ctx, db.DeleteRetweetByUserParams{
		UserID:    userID,
		RetweetID: &originalID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}
