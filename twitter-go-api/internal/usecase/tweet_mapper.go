package usecase

import db "github.com/chanombude/twitter-go-api/internal/db"

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

func mapForYouFeedRows(rows []db.ListForYouFeedRow) []TweetHydrationInput {
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

func mapFollowingFeedRows(rows []db.ListFollowingFeedRow) []TweetHydrationInput {
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

func mapUserTweetRows(rows []db.ListUserTweetsRow) []TweetHydrationInput {
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
