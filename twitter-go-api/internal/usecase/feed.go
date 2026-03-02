package usecase

import (
	"context"
	"database/sql"

	"github.com/chanombude/twitter-go-api/internal/db"
)

func (u *Usecase) GetGlobalFeed(ctx context.Context, page, size int32, viewerID *int64) ([]TweetItem, error) {
	rows, err := u.store.ListForYouFeed(ctx, db.ListForYouFeedParams{
		Limit:    size,
		Offset:   page * size,
		ViewerID: nullViewerID(viewerID),
	})
	if err != nil {
		return nil, err
	}

	return u.populateTweetItems(ctx, mapForYouFeedRows(rows), viewerID)
}

func (u *Usecase) CountGlobalFeed(ctx context.Context) (int64, error) {
	return u.store.CountForYouFeed(ctx)
}

func (u *Usecase) GetFollowingFeed(ctx context.Context, userID int64, page, size int32) ([]TweetItem, error) {
	rows, err := u.store.ListFollowingFeed(ctx, db.ListFollowingFeedParams{
		FollowerID: userID,
		Limit:      size,
		Offset:     page * size,
		ViewerID:   sql.NullInt64{Int64: userID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return u.populateTweetItems(ctx, mapFollowingFeedRows(rows), &userID)
}

func (u *Usecase) CountFollowingFeed(ctx context.Context, userID int64) (int64, error) {
	return u.store.CountFollowingFeed(ctx, userID)
}

func (u *Usecase) GetUserFeed(ctx context.Context, userID int64, page, size int32, viewerID *int64) ([]TweetItem, error) {
	rows, err := u.store.ListUserTweets(ctx, db.ListUserTweetsParams{
		UserID:   userID,
		Limit:    size,
		Offset:   page * size,
		ViewerID: nullViewerID(viewerID),
	})
	if err != nil {
		return nil, err
	}

	return u.populateTweetItems(ctx, mapUserTweetRows(rows), viewerID)
}

func (u *Usecase) CountUserFeed(ctx context.Context, userID int64) (int64, error) {
	return u.store.CountUserTweets(ctx, userID)
}
