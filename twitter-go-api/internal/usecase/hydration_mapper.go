package usecase

import "github.com/chanombude/twitter-go-api/internal/db"

func mapTweetHydrationRows[T any](
	rows []T,
	tweetFn func(T) db.Tweet,
	likedFn func(T) bool,
	retweetedFn func(T) bool,
	followingFn func(T) bool,
) []TweetHydrationInput {
	items := make([]TweetHydrationInput, len(rows))
	for i := range rows {
		row := rows[i]
		items[i] = TweetHydrationInput{
			Tweet:       tweetFn(row),
			IsLiked:     likedFn(row),
			IsRetweeted: retweetedFn(row),
			IsFollowing: followingFn(row),
		}
	}
	return items
}
