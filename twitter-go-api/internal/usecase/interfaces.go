package usecase

import (
	"context"

	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/service"
	"github.com/chanombude/twitter-go-api/internal/token"
)

type AuthService interface {
	LoginWithGoogle(ctx context.Context, idToken string) (AuthResult, error)
	RefreshSession(ctx context.Context, refreshToken string) (AuthResult, error)
	Logout(ctx context.Context, userID *int64, refreshToken *string)
	GetMe(ctx context.Context, userID int64) (UserItem, error)
}

type UserService interface {
	GetUser(ctx context.Context, targetUserID int64, viewerID *int64) (UserItem, error)
	UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (UserItem, error)
	FollowUser(ctx context.Context, followerID, targetUserID int64) (bool, error)
	UnfollowUser(ctx context.Context, followerID, targetUserID int64) error
	ListFollowers(ctx context.Context, targetUserID int64, page, size int32, viewerID *int64) ([]UserItem, error)
	ListFollowing(ctx context.Context, targetUserID int64, page, size int32, viewerID *int64) ([]UserItem, error)
}

type TweetService interface {
	CreateTweet(ctx context.Context, input CreateTweetInput) (TweetItem, error)
	DeleteTweet(ctx context.Context, userID, tweetID int64) error
	GetTweet(ctx context.Context, tweetID int64, viewerID *int64) (TweetItem, error)
	ListReplies(ctx context.Context, tweetID int64, page, size int32, viewerID *int64) ([]TweetItem, error)
	LikeTweet(ctx context.Context, userID, tweetID int64) error
	UnlikeTweet(ctx context.Context, userID, tweetID int64) error
	Retweet(ctx context.Context, userID, tweetID int64) (TweetItem, error)
	UndoRetweet(ctx context.Context, userID, tweetID int64) error
}

type FeedService interface {
	GetGlobalFeed(ctx context.Context, page, size int32, viewerID *int64) ([]TweetItem, error)
	GetFollowingFeed(ctx context.Context, userID int64, page, size int32) ([]TweetItem, error)
	GetUserFeed(ctx context.Context, userID int64, page, size int32, viewerID *int64) ([]TweetItem, error)
}

type SearchService interface {
	SearchUsers(ctx context.Context, query string, page, size int32, viewerID *int64) ([]UserItem, error)
	SearchTweets(ctx context.Context, query string, page, size int32, viewerID *int64) ([]TweetItem, error)
	SearchHashtags(ctx context.Context, query string, limit int32) ([]db.Hashtag, error)
}

type DiscoveryService interface {
	GetTrendingHashtags(ctx context.Context, limit int32) ([]db.Hashtag, error)
	GetSuggestedUsers(ctx context.Context, page, size int32, viewerID *int64) ([]UserItem, error)
}

type NotificationService interface {
	ListNotifications(ctx context.Context, userID int64, page, size int32) ([]NotificationItem, error)
	CountUnreadNotifications(ctx context.Context, userID int64) (int64, error)
	MarkAllNotificationsRead(ctx context.Context, userID int64) error
	HydrateNotification(ctx context.Context, notification db.Notification) (NotificationItem, error)
}

type MessageService interface {
	ListConversations(ctx context.Context, userID int64, page, size int32) ([]ConversationItem, error)
	ListMessages(ctx context.Context, userID, conversationID int64, page, size int32) ([]MessageItem, error)
	SendMessageToUser(ctx context.Context, senderID, recipientID int64, content string) (MessageItem, []int64, error)
	SendMessageToConversation(ctx context.Context, senderID, conversationID int64, content string) (MessageItem, []int64, error)
	ListPublicRoomMessages(ctx context.Context, roomKey string, page, size int32, viewerID *int64) ([]PublicRoomMessageItem, error)
	SendPublicRoomMessage(ctx context.Context, senderID int64, roomKey, content string) (PublicRoomMessageItem, error)
}

type AuthUsecase struct {
	config     config.Config
	store      db.Store
	tokenMaker token.Maker
}

type UserUsecase struct {
	store               db.Store
	storage             service.StorageService
	publishNotification func(db.Notification)
}

type TweetUsecase struct {
	store               db.Store
	storage             service.StorageService
	publishNotification func(db.Notification)
}

type FeedUsecase struct {
	store db.Store
}

type SearchUsecase struct {
	store db.Store
}

type DiscoveryUsecase struct {
	store db.Store
}

type NotificationUsecase struct {
	store db.Store
}

type MessageUsecase struct {
	store db.Store
}

type ServiceSet struct {
	Auth         AuthService
	User         UserService
	Tweet        TweetService
	Feed         FeedService
	Search       SearchService
	Discovery    DiscoveryService
	Notification NotificationService
	Message      MessageService
}

func NewServices(cfg config.Config, store db.Store, tokenMaker token.Maker, storage service.StorageService, publishNotification func(db.Notification)) ServiceSet {
	return ServiceSet{
		Auth: &AuthUsecase{
			config:     cfg,
			store:      store,
			tokenMaker: tokenMaker,
		},
		User: &UserUsecase{
			store:               store,
			storage:             storage,
			publishNotification: publishNotification,
		},
		Tweet: &TweetUsecase{
			store:               store,
			storage:             storage,
			publishNotification: publishNotification,
		},
		Feed:         &FeedUsecase{store: store},
		Search:       &SearchUsecase{store: store},
		Discovery:    &DiscoveryUsecase{store: store},
		Notification: &NotificationUsecase{store: store},
		Message:      &MessageUsecase{store: store},
	}
}

var (
	_ AuthService         = (*AuthUsecase)(nil)
	_ UserService         = (*UserUsecase)(nil)
	_ TweetService        = (*TweetUsecase)(nil)
	_ FeedService         = (*FeedUsecase)(nil)
	_ SearchService       = (*SearchUsecase)(nil)
	_ DiscoveryService    = (*DiscoveryUsecase)(nil)
	_ NotificationService = (*NotificationUsecase)(nil)
	_ MessageService      = (*MessageUsecase)(nil)
)
