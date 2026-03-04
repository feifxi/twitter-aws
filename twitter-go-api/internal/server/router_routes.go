package server

import (
	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

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
	messagesPublic.GET("/public/:room/messages", optionalAuth, server.listPublicRoomMessages)

	messagesPrivate := api.Group("/messages")
	messagesPrivate.Use(requiredAuth)
	messagesPrivate.GET("/conversations", server.listConversations)
	messagesPrivate.GET("/conversations/:id/messages", server.listConversationMessages)
	messagesPrivate.POST("/conversations/:id/messages", strictWriteLimiter, server.sendMessageToConversation)
	messagesPrivate.POST("/users/:id/messages", strictWriteLimiter, server.sendMessageToUser)
	messagesPrivate.POST("/public/:room/messages", strictWriteLimiter, server.sendPublicRoomMessage)
}
