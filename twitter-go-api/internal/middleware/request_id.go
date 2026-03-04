package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader     = "X-Request-ID"
	RequestIDContextKey = "request_id"
)

// RequestID ensures every request has a correlation id and returns it to the client.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := strings.TrimSpace(ctx.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx.Set(RequestIDContextKey, requestID)
		ctx.Writer.Header().Set(RequestIDHeader, requestID)
		ctx.Next()
	}
}

func GetRequestID(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}
	if raw, ok := ctx.Get(RequestIDContextKey); ok {
		if v, ok := raw.(string); ok {
			return v
		}
	}
	return ""
}
