package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/chanombude/twitter-go-api/internal/config"
	db "github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/logger"
	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/chanombude/twitter-go-api/internal/service"
	"github.com/chanombude/twitter-go-api/internal/token"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Server struct {
	config     config.Config
	store      db.Store
	tokenMaker token.Maker
	storage    service.StorageService
	usecase    *usecase.Usecase
	router     *gin.Engine
	redis      *redis.Client
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
		storage:    storageService,
		redis:      redisClient,
	}
	server.usecase = usecase.New(config, store, tokenMaker, storageService, redisClient, server.publishNotification)

	server.setupRouter()

	if redisClient != nil {
		go server.listenRedisNotifications()
	}

	return server, nil
}

func (server *Server) setupRouter() {
	configureValidationFieldNames()

	router := gin.New()
	if server.config.MaxMultipartMemoryBytes > 0 {
		router.MaxMultipartMemory = server.config.MaxMultipartMemoryBytes
	}
	router.Use(logger.GinMiddleware())
	router.Use(gin.Recovery())

	// Standard security headers.
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	})

	var allowOrigins []string
	if strings.TrimSpace(server.config.FrontendURL) != "" {
		allowOrigins = strings.Split(server.config.FrontendURL, ",")
		for i := range allowOrigins {
			allowOrigins[i] = strings.TrimSpace(allowOrigins[i])
		}
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Default rate limiter.
	router.Use(middleware.RateLimiter(20, 60))

	// Stricter auth limits.
	strictAuthLimiter := middleware.RateLimiter(2, 5)
	router.POST("/api/v1/auth/google", strictAuthLimiter, server.loginGoogle)
	router.POST("/api/v1/auth/refresh", strictAuthLimiter, server.refreshToken)
	router.POST("/api/v1/auth/logout", strictAuthLimiter, middleware.OptionalAuthMiddleware(server.tokenMaker), server.logout)

	api := router.Group("/api/v1")

	publicRoutes := api.Group("")
	publicRoutes.Use(middleware.OptionalAuthMiddleware(server.tokenMaker))
	publicRoutes.GET("/users/:id", server.getUser)
	publicRoutes.GET("/users/:id/followers", server.listFollowers)
	publicRoutes.GET("/users/:id/following", server.listFollowing)
	publicRoutes.GET("/tweets/:id", server.getTweet)
	publicRoutes.GET("/tweets/:id/replies", server.getReplies)
	publicRoutes.GET("/feeds/global", server.getGlobalFeed)
	publicRoutes.GET("/feeds/user/:id", server.getUserFeed)
	publicRoutes.GET("/search/users", server.searchUsers)
	publicRoutes.GET("/search/tweets", server.searchTweets)
	publicRoutes.GET("/search/hashtags", server.searchHashtags)
	publicRoutes.GET("/discovery/trending", server.getTrendingHashtags)
	publicRoutes.GET("/discovery/users", server.getSuggestedUsers)

	strictWriteLimiter := middleware.RateLimiter(2, 5)
	authRoutes := api.Group("")
	authRoutes.Use(middleware.AuthMiddleware(server.tokenMaker))
	authRoutes.GET("/auth/me", server.getMe)
	authRoutes.PUT("/users/profile", strictWriteLimiter, server.updateProfile)
	authRoutes.POST("/users/:id/follow", strictWriteLimiter, server.followUser)
	authRoutes.DELETE("/users/:id/follow", server.unfollowUser)
	authRoutes.POST("/tweets", strictWriteLimiter, server.createTweet)
	authRoutes.DELETE("/tweets/:id", server.deleteTweet)
	authRoutes.POST("/tweets/:id/like", strictWriteLimiter, server.likeTweet)
	authRoutes.DELETE("/tweets/:id/like", server.unlikeTweet)
	authRoutes.POST("/tweets/:id/retweet", strictWriteLimiter, server.retweet)
	authRoutes.DELETE("/tweets/:id/retweet", server.undoRetweet)
	authRoutes.GET("/feeds/following", server.getFollowingFeed)
	authRoutes.GET("/notifications", server.listNotifications)
	authRoutes.GET("/notifications/stream", server.streamNotifications)
	authRoutes.GET("/notifications/unread-count", server.getUnreadNotificationCount)
	authRoutes.POST("/notifications/mark-read", server.markNotificationRead)

	server.router = router
}

func configureValidationFieldNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		for _, tag := range []string{"json", "form", "uri"} {
			name := parseValidationTagName(field.Tag.Get(tag))
			if name != "" {
				return name
			}
		}
		return field.Name
	})
}

func parseValidationTagName(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return ""
	}
	name := strings.TrimSpace(parts[0])
	if name == "" || name == "-" {
		return ""
	}
	return name
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

	hydrated, err := server.usecase.HydrateNotification(hydrateCtx, notification)
	if err != nil {
		log.Error().Err(err).Int64("notification_id", notification.ID).Int64("recipient_id", notification.RecipientID).Msg("Failed to hydrate notification for SSE; event not published")
		return
	}
	server.broadcastToRedis(notification.RecipientID, newNotificationResponse(hydrated))
}

func (server *Server) broadcastToRedis(recipientID int64, notification notificationResponse) {
	if server.redis == nil {
		sendNotificationToUser(recipientID, notification)
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
		sendNotificationToUser(recipientID, notification)
	}
}
