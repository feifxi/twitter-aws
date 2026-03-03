package usecase

import (
	"time"

	"github.com/chanombude/twitter-go-api/internal/db"
)

// Notification type constants.
const (
	NotifTypeReply   = "REPLY"
	NotifTypeLike    = "LIKE"
	NotifTypeRetweet = "RETWEET"
	NotifTypeFollow  = "FOLLOW"
)

// MediaType constants.
const (
	MediaTypeNone  = "NONE"
	MediaTypeImage = "IMAGE"
	MediaTypeVideo = "VIDEO"
)

// Role constants.
const (
	RoleUser = "USER"
)

type TweetItem struct {
	ID             int64
	UserID         int64
	Content        *string
	MediaType      *string
	MediaUrl       *string
	ParentID       *int64
	RetweetID      *int64
	ReplyCount     int32
	RetweetCount   int32
	LikeCount      int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Author         UserItem
	IsLiked        bool
	IsRetweeted    bool
	IsFollowing    bool
	ParentUsername *string
	OriginalTweet  *TweetItem
}

type TweetHydrationInput struct {
	Tweet       db.Tweet
	IsLiked     bool
	IsRetweeted bool
	IsFollowing bool
}

type UserItem struct {
	ID             int64
	Username       string
	Email          string
	DisplayName    *string
	Bio            *string
	AvatarUrl      *string
	Role           string
	Provider       string
	FollowersCount int32
	FollowingCount int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IsFollowing    bool
}
