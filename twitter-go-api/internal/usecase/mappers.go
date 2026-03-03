package usecase

import "github.com/chanombude/twitter-go-api/internal/db"

func newUserItemFromDB(user db.User, isFollowing bool) UserItem {
	return UserItem{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		DisplayName:    user.DisplayName,
		Bio:            user.Bio,
		AvatarUrl:      user.AvatarUrl,
		Role:           user.Role,
		Provider:       user.Provider,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		IsFollowing:    isFollowing,
	}
}

func newTweetItemFromDB(tweet db.Tweet) TweetItem {
	return TweetItem{
		ID:           tweet.ID,
		UserID:       tweet.UserID,
		Content:      tweet.Content,
		MediaType:    tweet.MediaType,
		MediaUrl:     tweet.MediaUrl,
		ParentID:     tweet.ParentID,
		RetweetID:    tweet.RetweetID,
		ReplyCount:   tweet.ReplyCount,
		RetweetCount: tweet.RetweetCount,
		LikeCount:    tweet.LikeCount,
		CreatedAt:    tweet.CreatedAt,
		UpdatedAt:    tweet.UpdatedAt,
	}
}
