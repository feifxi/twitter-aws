package usecase

import (
	"context"

	"github.com/chanombude/twitter-go-api/internal/db"
)

func (u *FeedUsecase) GetGlobalFeed(ctx context.Context, page, size int32, viewerID *int64) ([]TweetItem, error) {
	rows, err := u.store.ListForYouFeed(ctx, db.ListForYouFeedParams{
		Limit:    size,
		Offset:   page * size,
		ViewerID: viewerID,
	})
	if err != nil {
		return nil, err
	}

	inputs := mapTweetHydrationRows(
		rows,
		func(r db.ListForYouFeedRow) db.Tweet { return r.Tweet },
		func(r db.ListForYouFeedRow) bool { return r.IsLiked },
		func(r db.ListForYouFeedRow) bool { return r.IsRetweeted },
		func(r db.ListForYouFeedRow) bool { return r.IsFollowing },
	)
	return populateTweetItems(ctx, u.store, inputs, viewerID)
}

func (u *FeedUsecase) GetFollowingFeed(ctx context.Context, userID int64, page, size int32) ([]TweetItem, error) {
	rows, err := u.store.ListFollowingFeed(ctx, db.ListFollowingFeedParams{
		FollowerID: userID,
		Limit:      size,
		Offset:     page * size,
		ViewerID:   &userID,
	})
	if err != nil {
		return nil, err
	}

	inputs := mapTweetHydrationRows(
		rows,
		func(r db.ListFollowingFeedRow) db.Tweet { return r.Tweet },
		func(r db.ListFollowingFeedRow) bool { return r.IsLiked },
		func(r db.ListFollowingFeedRow) bool { return r.IsRetweeted },
		func(r db.ListFollowingFeedRow) bool { return r.IsFollowing },
	)
	return populateTweetItems(ctx, u.store, inputs, &userID)
}

func (u *FeedUsecase) GetUserFeed(ctx context.Context, userID int64, page, size int32, viewerID *int64) ([]TweetItem, error) {
	if _, err := u.store.GetUser(ctx, db.GetUserParams{ID: userID, ViewerID: viewerID}); err != nil {
		return nil, err
	}

	rows, err := u.store.ListUserTweets(ctx, db.ListUserTweetsParams{
		UserID:   userID,
		Limit:    size,
		Offset:   page * size,
		ViewerID: viewerID,
	})
	if err != nil {
		return nil, err
	}

	inputs := mapTweetHydrationRows(
		rows,
		func(r db.ListUserTweetsRow) db.Tweet { return r.Tweet },
		func(r db.ListUserTweetsRow) bool { return r.IsLiked },
		func(r db.ListUserTweetsRow) bool { return r.IsRetweeted },
		func(r db.ListUserTweetsRow) bool { return r.IsFollowing },
	)
	return populateTweetItems(ctx, u.store, inputs, viewerID)
}
