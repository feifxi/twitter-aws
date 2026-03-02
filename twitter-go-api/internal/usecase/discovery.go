package usecase

import (
	"context"
	"database/sql"

	"github.com/chanombude/twitter-go-api/internal/db"
)

func (u *Usecase) GetTrendingHashtags(ctx context.Context, limit int32) ([]db.Hashtag, error) {
	hashtags, err := u.store.GetTrendingHashtagsLast24h(ctx, limit)
	if err != nil {
		return nil, err
	}
	if len(hashtags) == 0 {
		return u.store.GetTopHashtagsAllTime(ctx, limit)
	}

	return hashtags, nil
}

func (u *Usecase) GetSuggestedUsers(ctx context.Context, page, size int32, viewerID *int64) ([]UserItem, error) {
	if viewerID != nil {
		rows, err := u.store.ListSuggestedUsers(ctx, db.ListSuggestedUsersParams{
			FollowerID: *viewerID,
			Limit:      size,
			Offset:     page * size,
			ViewerID:   sql.NullInt64{Int64: *viewerID, Valid: true},
		})
		if err != nil {
			return nil, err
		}

		items := make([]UserItem, 0, len(rows))
		for _, r := range rows {
			items = append(items, UserItem{User: r.User, IsFollowing: r.IsFollowing})
		}
		return items, nil
	}

	users, err := u.store.ListTopUsers(ctx, db.ListTopUsersParams{Limit: size, Offset: page * size})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(users))
	for _, r := range users {
		items = append(items, UserItem{User: r, IsFollowing: false})
	}
	return items, nil
}

func (u *Usecase) CountSuggestedUsers(ctx context.Context, viewerID *int64) (int64, error) {
	if viewerID != nil {
		return u.store.CountSuggestedUsers(ctx, *viewerID)
	}
	return u.store.CountTopUsers(ctx)
}
