package server

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	_ "github.com/chanombude/twitter-go-api/docs"
	"github.com/chanombude/twitter-go-api/internal/logger"
	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (server *Server) setupRouter() {
	configureValidationFieldNames()

	router := gin.New()
	
	// Trusted proxies are critical when behind a load balancer (API Gateway/EICE)
	// We trust private IP ranges typically used in AWS VPCs
	router.SetTrustedProxies([]string{
		"10.0.0.0/8",     // Private IPv4 Class A
		"172.16.0.0/12",  // Private IPv4 Class B
		"192.168.0.0/16", // Private IPv4 Class C
	})

	router.Use(middleware.Metrics())
	router.Use(middleware.RequestID())
	router.Use(logger.GinMiddleware())
	router.Use(gin.Recovery())

	// Handle unmatched routes (silence bot scans and apply strict rate limiting)
	router.NoRoute(
		middleware.RateLimiterWithRedis(server.redis, 1, 5, "rl:noroute"), // Strict: 1 req/sec, burst of 5
		func(c *gin.Context) {
			// Set flag to skip logging this request in logger.GinMiddleware
			c.Set("skip_log", true)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "route not found",
			})
		},
	)

	// Standard security headers.
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Next()
	})

	allowOrigins := parseAllowedOrigins(server.config.FrontendURL)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", server.healthz)
	router.GET("/readyz", server.readyz)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger UI (disabled in production for security)
	if server.config.Environment != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := router.Group("/api/v1")
	if server.config.GatewaySecret != "" {
		api.Use(middleware.GatewayGuard(server.config.GatewaySecret))
	}
	api.Use(middleware.RateLimiterWithRedis(server.redis, 20, 60, "rl:api"))
	server.registerDomainRoutes(api)

	server.router = router
}

func (server *Server) registerDomainRoutes(api *gin.RouterGroup) {
	strictAuthLimiter := middleware.RateLimiterWithRedis(server.redis, 2, 5, "rl:auth")
	strictWriteLimiter := middleware.RateLimiterWithRedis(server.redis, 2, 5, "rl:write")
	optionalAuth := middleware.OptionalAuthMiddleware(server.tokenMaker)
	requiredAuth := middleware.AuthMiddleware(server.tokenMaker)

	server.registerAuthRoutes(api, optionalAuth, requiredAuth, strictAuthLimiter)
	server.registerUserRoutes(api, optionalAuth, requiredAuth, strictWriteLimiter)
	server.registerTweetRoutes(api, optionalAuth, requiredAuth, strictWriteLimiter)
	server.registerFeedRoutes(api, optionalAuth, requiredAuth)
	server.registerSearchRoutes(api, optionalAuth)
	server.registerDiscoveryRoutes(api, optionalAuth)
	server.registerNotificationRoutes(api, requiredAuth)
	server.registerMessageRoutes(api, optionalAuth, requiredAuth, strictWriteLimiter)
	server.registerAssistantRoutes(api, requiredAuth)
}

func (server *Server) registerAuthRoutes(api *gin.RouterGroup, optionalAuth, requiredAuth, strictAuthLimiter gin.HandlerFunc) {
	authPublic := api.Group("/auth")
	authPublic.POST("/google", strictAuthLimiter, server.loginGoogle)
	authPublic.POST("/refresh", strictAuthLimiter, server.refreshToken)
	authPublic.POST("/logout", strictAuthLimiter, optionalAuth, server.logout)

	authPrivate := api.Group("/auth")
	authPrivate.Use(requiredAuth)
	authPrivate.GET("/me", server.getMe)
}

func (server *Server) registerUserRoutes(api *gin.RouterGroup, optionalAuth, requiredAuth, strictWriteLimiter gin.HandlerFunc) {
	usersPublic := api.Group("/users")
	usersPublic.Use(optionalAuth)
	usersPublic.GET("/:id", server.getUser)
	usersPublic.GET("/:id/followers", server.listFollowers)
	usersPublic.GET("/:id/following", server.listFollowing)

	usersPrivate := api.Group("/users")
	usersPrivate.Use(requiredAuth)
	usersPrivate.PUT("/profile", strictWriteLimiter, server.updateProfile)
	usersPrivate.POST("/:id/follow", strictWriteLimiter, server.followUser)
	usersPrivate.DELETE("/:id/follow", server.unfollowUser)
}

func (server *Server) registerTweetRoutes(api *gin.RouterGroup, optionalAuth, requiredAuth, strictWriteLimiter gin.HandlerFunc) {
	tweetsPublic := api.Group("/tweets")
	tweetsPublic.Use(optionalAuth)
	tweetsPublic.GET("/:id", server.getTweet)
	tweetsPublic.GET("/:id/replies", server.getReplies)

	tweetsPrivate := api.Group("/tweets")
	tweetsPrivate.Use(requiredAuth)
	tweetsPrivate.POST("", strictWriteLimiter, server.createTweet)
	tweetsPrivate.DELETE("/:id", server.deleteTweet)
	tweetsPrivate.POST("/:id/like", strictWriteLimiter, server.likeTweet)
	tweetsPrivate.DELETE("/:id/like", server.unlikeTweet)
	tweetsPrivate.POST("/:id/retweet", strictWriteLimiter, server.retweet)
	tweetsPrivate.DELETE("/:id/retweet", server.undoRetweet)

	uploadsPrivate := api.Group("/uploads")
	uploadsPrivate.Use(requiredAuth)
	uploadsPrivate.POST("/presign", strictWriteLimiter, server.presignUpload)
}

func (server *Server) registerFeedRoutes(api *gin.RouterGroup, optionalAuth, requiredAuth gin.HandlerFunc) {
	feedsPublic := api.Group("/feeds")
	feedsPublic.Use(optionalAuth)
	feedsPublic.GET("/global", server.getGlobalFeed)
	feedsPublic.GET("/user/:id", server.getUserFeed)

	feedsPrivate := api.Group("/feeds")
	feedsPrivate.Use(requiredAuth)
	feedsPrivate.GET("/following", server.getFollowingFeed)
}

func (server *Server) registerSearchRoutes(api *gin.RouterGroup, optionalAuth gin.HandlerFunc) {
	searchPublic := api.Group("/search")
	searchPublic.Use(optionalAuth)
	searchPublic.GET("/users", server.searchUsers)
	searchPublic.GET("/tweets", server.searchTweets)
	searchPublic.GET("/hashtags", server.searchHashtags)
}

func (server *Server) registerDiscoveryRoutes(api *gin.RouterGroup, optionalAuth gin.HandlerFunc) {
	discoveryPublic := api.Group("/discovery")
	discoveryPublic.Use(optionalAuth)
	discoveryPublic.GET("/trending", server.getTrendingHashtags)
	discoveryPublic.GET("/users", server.getSuggestedUsers)
}

func (server *Server) registerNotificationRoutes(api *gin.RouterGroup, requiredAuth gin.HandlerFunc) {
	notificationsPrivate := api.Group("/notifications")
	notificationsPrivate.Use(requiredAuth)
	notificationsPrivate.GET("", server.listNotifications)
	notificationsPrivate.GET("/stream", server.streamNotifications)
	notificationsPrivate.GET("/unread-count", server.getUnreadNotificationCount)
	notificationsPrivate.POST("/mark-read", server.markNotificationRead)
}

func (server *Server) registerMessageRoutes(api *gin.RouterGroup, optionalAuth, requiredAuth, strictWriteLimiter gin.HandlerFunc) {
	messagesPublic := api.Group("/messages")
	messagesPublic.GET("/ws", server.streamMessagesWS)

	messagesPrivate := api.Group("/messages")
	messagesPrivate.Use(requiredAuth)
	messagesPrivate.GET("/conversations", server.listConversations)
	messagesPrivate.GET("/conversations/:id/messages", server.listConversationMessages)
	messagesPrivate.POST("/conversations/:id/messages", strictWriteLimiter, server.sendMessageToConversation)
	messagesPrivate.POST("/users/:id/messages", strictWriteLimiter, server.sendMessageToUser)
}
 
func (server *Server) registerAssistantRoutes(api *gin.RouterGroup, requiredAuth gin.HandlerFunc) {
	assistant := api.Group("/assistant")
	assistant.Use(requiredAuth)
	assistant.POST("", server.assistant)
}

func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
