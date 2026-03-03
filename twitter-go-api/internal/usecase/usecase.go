package usecase

import (
	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/service"
	"github.com/chanombude/twitter-go-api/internal/token"
	"github.com/redis/go-redis/v9"
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
	db.Tweet
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
	db.User
	IsFollowing bool
}

type Usecase struct {
	config              config.Config
	store               *db.Store
	tokenMaker          token.Maker
	storage             service.StorageService
	redis               *redis.Client
	publishNotification func(db.Notification)
}

func New(cfg config.Config, store *db.Store, tokenMaker token.Maker, storage service.StorageService, redisClient *redis.Client, publishNotification func(db.Notification)) *Usecase {
	return &Usecase{
		config:              cfg,
		store:               store,
		tokenMaker:          tokenMaker,
		storage:             storage,
		redis:               redisClient,
		publishNotification: publishNotification,
	}
}

func (u *Usecase) Store() *db.Store {
	return u.store
}
