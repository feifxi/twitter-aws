package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// A simple in-memory client manager for SSE.
type sseClient struct {
	channel chan notificationResponse
}

var (
	clients = make(map[int64][]*sseClient)
	mu      sync.RWMutex
)

func sendNotificationToUser(userID int64, notification notificationResponse) {
	mu.RLock()
	userClients, ok := clients[userID]
	mu.RUnlock()
	if !ok {
		return
	}

	for _, client := range userClients {
		select {
		case client.channel <- notification:
		default:
			log.Warn().Int64("user_id", userID).Int64("notification_id", notification.ID).Msg("Dropped SSE notification due to full client buffer")
		}
	}
}

// listenRedisNotifications subscribes to the Redis channel and forwards messages to local SSE clients.
func (server *Server) listenRedisNotifications() {
	if server.redis == nil {
		log.Warn().Msg("Redis client is nil, SSE will only work for single-instance deployments")
		return
	}

	ctx := context.Background()
	pubsub := server.redis.Subscribe(ctx, "notifications")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var payload redisNotificationPayload
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal notification from Redis")
			continue
		}

		sendNotificationToUser(payload.RecipientID, payload.Notification)
	}
}

func (server *Server) streamNotifications(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")

	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	client := &sseClient{channel: make(chan notificationResponse, 10)}

	mu.Lock()
	clients[userID] = append(clients[userID], client)
	connectionCount := len(clients[userID])
	mu.Unlock()
	log.Info().Int64("user_id", userID).Int("connections", connectionCount).Msg("SSE client connected")

	fmt.Fprintf(ctx.Writer, "event: connected\ndata: {\"status\": \"ok\"}\n\n")
	ctx.Writer.Flush()

	defer func() {
		mu.Lock()
		userClients := clients[userID]
		for i, c := range userClients {
			if c == client {
				clients[userID] = append(userClients[:i], userClients[i+1:]...)
				break
			}
		}
		remaining := len(clients[userID])
		mu.Unlock()
		close(client.channel)
		log.Info().Int64("user_id", userID).Int("connections", remaining).Msg("SSE client disconnected")
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(ctx.Writer, "event: ping\ndata: {}\n\n")
			ctx.Writer.Flush()
		case notification := <-client.channel:
			data, _ := json.Marshal(notification)
			fmt.Fprintf(ctx.Writer, "event: notification\ndata: %s\n\n", data)
			ctx.Writer.Flush()
		}
	}
}

func (server *Server) listNotifications(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	notifications, err := server.usecase.ListNotifications(ctx, userID, page, size)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountNotifications(ctx, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	content := make([]notificationResponse, 0, len(notifications))
	for _, n := range notifications {
		content = append(content, newNotificationResponse(n))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(content, page, size, total))
}

func (server *Server) getUnreadNotificationCount(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	count, err := server.usecase.CountUnreadNotifications(ctx, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, count)
}

func (server *Server) markNotificationRead(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	if err := server.usecase.MarkAllNotificationsRead(ctx, userID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
