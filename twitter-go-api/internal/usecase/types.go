package usecase

import (
	"io"
	"time"

	"github.com/chanombude/twitter-go-api/internal/db"
)

// FileUpload represents an uploaded file (used for avatars, tweet media, etc.).
type FileUpload struct {
	Filename    string
	ContentType string
	Reader      io.Reader
}

// Notification type constants.
const (
	NotifTypeReply   = "REPLY"
	NotifTypeLike    = "LIKE"
	NotifTypeRetweet = "RETWEET"
	NotifTypeFollow  = "FOLLOW"
)

// MediaType constants.
const (
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

type MessageItem struct {
	ID             int64
	ConversationID int64
	Sender         UserItem
	Content        string
	CreatedAt      time.Time
}

type PublicRoomMessageItem struct {
	ID        int64
	RoomKey   string
	Sender    UserItem
	Content   string
	CreatedAt time.Time
}

type ConversationItem struct {
	ID          int64
	Peer        UserItem
	LastMessage MessageItem
	UpdatedAt   time.Time
}
