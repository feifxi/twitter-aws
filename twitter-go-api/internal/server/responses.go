package server

import (
	"time"

	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/usecase"
)

type authResponse struct {
	AccessToken string       `json:"accessToken"`
	User        userResponse `json:"user"`
}

type userResponse struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	DisplayName    *string `json:"displayName"`
	Bio            *string `json:"bio"`
	AvatarUrl      *string `json:"avatarUrl"`
	IsFollowing    bool    `json:"isFollowing"`
	FollowersCount int32   `json:"followersCount"`
	FollowingCount int32   `json:"followingCount"`
}

func newUserResponse(user usecase.UserItem) userResponse {
	var displayName, bio, avatarUrl *string
	if user.DisplayName.Valid {
		displayName = &user.DisplayName.String
	}
	if user.Bio.Valid {
		bio = &user.Bio.String
	}
	if user.AvatarUrl.Valid {
		avatarUrl = &user.AvatarUrl.String
	}

	return userResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		DisplayName:    displayName,
		Bio:            bio,
		AvatarUrl:      avatarUrl,
		IsFollowing:    user.IsFollowing,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
	}
}

type tweetResponse struct {
	ID            int64          `json:"id"`
	Content       *string        `json:"content"`
	MediaType     *string        `json:"mediaType"`
	MediaUrl      *string        `json:"mediaUrl"`
	User          userResponse   `json:"user"`
	ReplyCount    int32          `json:"replyCount"`
	LikeCount     int32          `json:"likeCount"`
	RetweetCount  int32          `json:"retweetCount"`
	IsLiked       bool           `json:"isLiked"`
	IsRetweeted   bool           `json:"isRetweeted"`
	RetweetedTweet *tweetResponse `json:"retweetedTweet,omitempty"`
	ReplyToTweetID *int64         `json:"replyToTweetId"`
	ReplyToUsername *string       `json:"replyToUsername"`
	CreatedAt     time.Time      `json:"createdAt"`
}

func newTweetResponse(tweet usecase.TweetItem) tweetResponse {
	var content, mediaType, mediaURL *string
	if tweet.Content.Valid {
		content = &tweet.Content.String
	}
	if tweet.MediaType.Valid {
		mediaType = &tweet.MediaType.String
	}
	if tweet.MediaUrl.Valid {
		mediaURL = &tweet.MediaUrl.String
	}

	var parentID *int64
	if tweet.ParentID.Valid {
		parentID = &tweet.ParentID.Int64
	}

	var original *tweetResponse
	if tweet.OriginalTweet != nil {
		r := newTweetResponse(*tweet.OriginalTweet)
		original = &r
	}

	return tweetResponse{
		ID:             tweet.ID,
		Content:        content,
		MediaType:      mediaType,
		MediaUrl:       mediaURL,
		User:           newUserResponse(tweet.Author),
		ReplyCount:     tweet.ReplyCount,
		LikeCount:      tweet.LikeCount,
		RetweetCount:   tweet.RetweetCount,
		IsLiked:        tweet.IsLiked,
		IsRetweeted:    tweet.IsRetweeted,
		RetweetedTweet: original,
		ReplyToTweetID: parentID,
		ReplyToUsername: tweet.ParentUsername,
		CreatedAt:      tweet.CreatedAt,
	}
}

type hashtagResponse struct {
	ID         int64     `json:"id"`
	Text       string    `json:"text"`
	UsageCount int32     `json:"usageCount"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

func newHashtagResponse(tag db.Hashtag) hashtagResponse {
	return hashtagResponse{
		ID:         tag.ID,
		Text:       tag.Text,
		UsageCount: tag.UsageCount,
		LastUsedAt: tag.LastUsedAt,
		CreatedAt:  tag.CreatedAt,
	}
}

type notificationResponse struct {
	ID                   int64        `json:"id"`
	Actor                userResponse `json:"actor"`
	TweetID              *int64       `json:"tweetId"`
	TweetContent         *string      `json:"tweetContent"`
	TweetMediaUrl        *string      `json:"tweetMediaUrl"`
	OriginalTweetID      *int64       `json:"originalTweetId,omitempty"`
	OriginalTweetContent *string      `json:"originalTweetContent,omitempty"`
	OriginalTweetMediaUrl *string     `json:"originalTweetMediaUrl,omitempty"`
	Type                 string       `json:"type"`
	IsRead               bool         `json:"isRead"`
	CreatedAt            time.Time    `json:"createdAt"`
}

func newNotificationResponse(item usecase.NotificationItem) notificationResponse {
	var tweetID *int64
	if item.TweetID.Valid {
		tweetID = &item.TweetID.Int64
	}

	return notificationResponse{
		ID:                    item.ID,
		Actor:                 newUserResponse(item.Actor),
		TweetID:               tweetID,
		TweetContent:          item.TweetContent,
		TweetMediaUrl:         item.TweetMediaUrl,
		OriginalTweetID:       item.OriginalTweetID,
		OriginalTweetContent:  item.OriginalTweetContent,
		OriginalTweetMediaUrl: item.OriginalTweetMediaUrl,
		Type:                  item.Type,
		IsRead:                item.IsRead,
		CreatedAt:             item.CreatedAt,
	}
}
