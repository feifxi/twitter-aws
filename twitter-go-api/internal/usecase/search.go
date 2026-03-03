package usecase

import (
	"context"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/db"
)

func (u *SearchUsecase) SearchUsers(ctx context.Context, query string, page, size int32, viewerID *int64) ([]UserItem, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []UserItem{}, nil
	}

	rows, err := u.store.SearchUsers(ctx, db.SearchUsersParams{
		Column1:  &trimmed,
		Limit:    size,
		Offset:   page * size,
		ViewerID: viewerID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, newUserItemFromDB(r.User, r.IsFollowing))
	}
	return items, nil
}

func (u *SearchUsecase) CountSearchUsers(ctx context.Context, query string) (int64, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return 0, nil
	}
	return u.store.CountSearchUsers(ctx, &trimmed)
}

func (u *SearchUsecase) SearchTweets(ctx context.Context, query string, page, size int32, viewerID *int64) ([]TweetItem, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []TweetItem{}, nil
	}

	if strings.HasPrefix(trimmed, "#") {
		hashtag := strings.TrimSpace(strings.ToLower(strings.TrimLeft(trimmed, "#")))
		if hashtag == "" {
			return []TweetItem{}, nil
		}

		rows, err := u.store.SearchTweetsByHashtag(ctx, db.SearchTweetsByHashtagParams{
			Lower:    hashtag,
			Limit:    size,
			Offset:   page * size,
			ViewerID: viewerID,
		})
		if err != nil {
			return nil, err
		}

		inputs := mapTweetHydrationRows(
			rows,
			func(r db.SearchTweetsByHashtagRow) db.Tweet { return r.Tweet },
			func(r db.SearchTweetsByHashtagRow) bool { return r.IsLiked },
			func(r db.SearchTweetsByHashtagRow) bool { return r.IsRetweeted },
			func(r db.SearchTweetsByHashtagRow) bool { return r.IsFollowing },
		)
		return populateTweetItems(ctx, u.store, inputs, viewerID)
	}

	tsQuery := buildTSQuery(trimmed)
	if tsQuery == "" {
		return []TweetItem{}, nil
	}

	rows, err := u.store.SearchTweetsFullText(ctx, db.SearchTweetsFullTextParams{
		ToTsquery: tsQuery,
		Limit:     size,
		Offset:    page * size,
		ViewerID:  viewerID,
	})
	if err != nil {
		return nil, err
	}

	inputs := mapTweetHydrationRows(
		rows,
		func(r db.SearchTweetsFullTextRow) db.Tweet { return r.Tweet },
		func(r db.SearchTweetsFullTextRow) bool { return r.IsLiked },
		func(r db.SearchTweetsFullTextRow) bool { return r.IsRetweeted },
		func(r db.SearchTweetsFullTextRow) bool { return r.IsFollowing },
	)
	return populateTweetItems(ctx, u.store, inputs, viewerID)
}

func (u *SearchUsecase) CountSearchTweets(ctx context.Context, query string) (int64, error) {
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

func (u *SearchUsecase) SearchHashtags(ctx context.Context, query string, limit int32) ([]db.Hashtag, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []db.Hashtag{}, nil
	}

	prefix := strings.TrimPrefix(trimmed, "#")
	return u.store.SearchHashtagsByPrefix(ctx, db.SearchHashtagsByPrefixParams{
		Column1: &prefix,
		Limit:   limit,
	})
}
