package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/chanombude/twitter-go-api/internal/config"
	db "github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/service"
	"github.com/chanombude/twitter-go-api/internal/token"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Server struct {
	config      config.Config
	store       db.Store
	tokenMaker  token.Maker
	authUC      usecase.AuthService
	userUC      usecase.UserService
	tweetUC     usecase.TweetService
	feedUC      usecase.FeedService
	searchUC    usecase.SearchService
	discoveryUC usecase.DiscoveryService
	notifyUC    usecase.NotificationService
	messageUC   usecase.MessageService
	router      *gin.Engine
	redis       *redis.Client
	sseClients  map[int64][]*sseClient
	sseMu       sync.RWMutex
	wsClients   map[int64]map[*chatWSClient]struct{}
	wsPublic    map[*chatWSClient]struct{}
	wsMu        sync.RWMutex
}

type idURIRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func NewServer(config config.Config, store db.Store, redisClient *redis.Client) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	storageService, err := service.NewAzureStorageService(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create storage service: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		redis:      redisClient,
		sseClients: make(map[int64][]*sseClient),
		wsClients:  make(map[int64]map[*chatWSClient]struct{}),
		wsPublic:   make(map[*chatWSClient]struct{}),
	}
	services := usecase.NewServices(config, store, tokenMaker, storageService, server.publishNotification)
	server.authUC = services.Auth
	server.userUC = services.User
	server.tweetUC = services.Tweet
	server.feedUC = services.Feed
	server.searchUC = services.Search
	server.discoveryUC = services.Discovery
	server.notifyUC = services.Notification
	server.messageUC = services.Message

	server.setupRouter()

	if redisClient != nil {
		go server.listenRedisNotifications()
	}

	return server, nil
}

func (server *Server) HTTPServer(address string) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           server.router,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		// Keep write timeout disabled to support long-lived SSE streams.
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}
}

type redisNotificationPayload struct {
	RecipientID  int64                `json:"recipientId"`
	Notification notificationResponse `json:"notification"`
}

func (server *Server) publishNotification(notification db.Notification) {
	hydrateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hydrated, err := server.notifyUC.HydrateNotification(hydrateCtx, notification)
	if err != nil {
		log.Error().Err(err).Int64("notification_id", notification.ID).Int64("recipient_id", notification.RecipientID).Msg("Failed to hydrate notification for SSE; event not published")
		return
	}
	server.broadcastToRedis(notification.RecipientID, newNotificationResponse(hydrated))
}

func (server *Server) broadcastToRedis(recipientID int64, notification notificationResponse) {
	if server.redis == nil {
		server.sendNotificationToUser(recipientID, notification)
		return
	}

	payload := redisNotificationPayload{RecipientID: recipientID, Notification: notification}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal redis notification payload")
		return
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.redis.Publish(pubCtx, "notifications", data).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish notification to Redis")
		server.sendNotificationToUser(recipientID, notification)
	}
}
