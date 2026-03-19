package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// A simple in-memory client manager for SSE.
type sseClient struct {
	channel chan NotificationResponse
}

func (server *Server) sendNotificationToUser(userID int64, notification NotificationResponse) {
	server.sseMu.RLock()
	userClients, ok := server.sseClients[userID]
	snapshot := append([]*sseClient(nil), userClients...)
	server.sseMu.RUnlock()
	if !ok {
		return
	}

	for _, client := range snapshot {
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

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-server.done
		cancel()
	}()

	pubsub := server.redis.Subscribe(ctx, "notifications")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var payload redisNotificationPayload
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal notification from Redis")
			continue
		}

		server.sendNotificationToUser(payload.RecipientID, payload.Notification)
	}
}

// streamNotifications godoc
// @Summary		Stream Real-time Notifications (SSE)
// @Description	Establish a Server-Sent Events (SSE) connection to receive near real-time notifications for likes, replies, retweets, and new followers.
// @Tags			Notifications
// @Produce			text/event-stream
// @Success		200	{string}	string	"text/event-stream"
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/notifications/stream [get]
func (server *Server) streamNotifications(ctx *gin.Context) {
	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		writeError(ctx, apperr.Internal("streaming unsupported", nil))
		return
	}

	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	// Flush headers immediately to establish the connection with intermediaries (API Gateway)
	flusher.Flush()

	client := &sseClient{channel: make(chan NotificationResponse, 32)}

	server.sseMu.Lock()
	server.sseClients[userID] = append(server.sseClients[userID], client)
	connectionCount := len(server.sseClients[userID])
	server.sseMu.Unlock()
	log.Info().Int64("user_id", userID).Int("connections", connectionCount).Msg("SSE client connected")

	// Suggest reconnect delay to client.
	fmt.Fprint(ctx.Writer, "retry: 3000\n\n")
	fmt.Fprintf(ctx.Writer, "event: connected\ndata: {\"status\": \"ok\"}\n\n")
	flusher.Flush()

	defer func() {
		server.sseMu.Lock()
		defer server.sseMu.Unlock()

		userClients := server.sseClients[userID]
		for i, c := range userClients {
			if c == client {
				server.sseClients[userID] = append(userClients[:i], userClients[i+1:]...)
				break
			}
		}

		if len(server.sseClients[userID]) == 0 {
			delete(server.sseClients, userID)
		}
		log.Info().Int64("user_id", userID).Int("connections", len(server.sseClients[userID])).Msg("SSE client disconnected")
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case <-ticker.C:
			// Send comment line as heartbeat to keep intermediaries from timing out idle connection.
			fmt.Fprint(ctx.Writer, ": ping\n\n")
			flusher.Flush()
		case notification := <-client.channel:
			data, err := json.Marshal(notification)
			if err != nil {
				log.Error().Err(err).Int64("notification_id", notification.ID).Msg("Failed to marshal SSE notification payload")
				continue
			}
			fmt.Fprintf(ctx.Writer, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// listNotifications godoc
// @Summary		List Notifications
// @Description	Get a paginated list of social notifications (likes, follows, etc.) for the current user.
// @Tags			Notifications
// @Produce		json
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[NotificationResponse]
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/notifications [get]
func (server *Server) listNotifications(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	items, err := server.notifyUC.ListNotifications(ctx, userID, page, size+1)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newNotificationResponseList(items)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// getUnreadNotificationCount godoc
// @Summary		Get Unread Count
// @Description	Calculate how many unread notifications the current user has.
// @Tags			Notifications
// @Produce		json
// @Success		200	{object}	map[string]int64	"unread count"
// @Security		BearerAuth
// @Failure		401	{object}	ErrorResponse
// @Router			/notifications/unread-count [get]
func (server *Server) getUnreadNotificationCount(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	count, err := server.notifyUC.CountUnreadNotifications(ctx, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// markNotificationRead godoc
// @Summary		Mark All Read
// @Description	Clear all unread notification badges for the current user.
// @Tags			Notifications
// @Success		200	{object}	SuccessResponse
// @Security		BearerAuth
// @Failure		401	{object}	ErrorResponse
// @Router			/notifications/mark-read [post]
func (server *Server) markNotificationRead(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	if err := server.notifyUC.MarkAllNotificationsRead(ctx, userID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
}
