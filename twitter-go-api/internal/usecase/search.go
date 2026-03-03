package usecase

import (
	"context"
	"database/sql"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/db"
)

func (u *Usecase) SearchUsers(ctx context.Context, query string, page, size int32, viewerID *int64) ([]UserItem, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []UserItem{}, nil
	}

	rows, err := u.store.SearchUsers(ctx, db.SearchUsersParams{
		Column1:  sql.NullString{String: trimmed, Valid: true},
		Limit:    size,
		Offset:   page * size,
		ViewerID: nullViewerID(viewerID),
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

func (u *Usecase) CountSearchUsers(ctx context.Context, query string) (int64, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return 0, nil
	}
	return u.store.CountSearchUsers(ctx, sql.NullString{String: trimmed, Valid: true})
}

func (u *Usecase) SearchTweets(ctx context.Context, query string, page, size int32, viewerID *int64) ([]TweetItem, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []TweetItem{}, nil
	}

	vID := nullViewerID(viewerID)
	if strings.HasPrefix(trimmed, "#") {
		hashtag := strings.TrimSpace(strings.ToLower(strings.TrimLeft(trimmed, "#")))
		if hashtag == "" {
			return []TweetItem{}, nil
		}

		rows, err := u.store.SearchTweetsByHashtag(ctx, db.SearchTweetsByHashtagParams{
			Lower:    hashtag,
			Limit:    size,
			Offset:   page * size,
			ViewerID: vID,
		})
		if err != nil {
			return nil, err
		}

		return u.populateTweetItems(ctx, mapHashtagSearchRows(rows), viewerID)
	}

	tsQuery := buildTSQuery(trimmed)
	if tsQuery == "" {
		return []TweetItem{}, nil
	}

	rows, err := u.store.SearchTweetsFullText(ctx, db.SearchTweetsFullTextParams{
		ToTsquery: tsQuery,
		Limit:     size,
		Offset:    page * size,
		ViewerID:  vID,
	})
	if err != nil {
		return nil, err
	}

	return u.populateTweetItems(ctx, mapFullTextSearchRows(rows), viewerID)
}

func (u *Usecase) CountSearchTweets(ctx context.Context, query string) (int64, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return 0, nil
	}

	if strings.HasPrefix(trimmed, "#") {
		return u.store.CountSearchTweetsByHashtag(ctx, strings.ToLower(trimmed[1:]))
	}

	tsQuery := buildTSQuery(trimmed)
	return u.store.CountSearchTweetsFullText(ctx, tsQuery)
}

func (u *Usecase) SearchHashtags(ctx context.Context, query string, limit int32) ([]db.Hashtag, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []db.Hashtag{}, nil
	}

	return u.store.SearchHashtagsByPrefix(ctx, db.SearchHashtagsByPrefixParams{
		Column1: sql.NullString{String: strings.TrimPrefix(trimmed, "#"), Valid: true},
		Limit:   limit,
	})
}

func mapHashtagSearchRows(rows []db.SearchTweetsByHashtagRow) []TweetHydrationInput {
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

func mapFullTextSearchRows(rows []db.SearchTweetsFullTextRow) []TweetHydrationInput {
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
