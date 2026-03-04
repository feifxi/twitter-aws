package logger

import (
	"os"
	"time"

	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger configures the global zerolog instance.
func InitLogger(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		// Pretty console output for unhandled development logging
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		// Strict JSON for production
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
}

// GinMiddleware replaces the default Gin logger with Zerolog.
func GinMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery
		route := ctx.FullPath()
		if route == "" {
			route = path
		}
		requestID := middleware.GetRequestID(ctx)

		ctx.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		duration := time.Since(start)
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()
		method := ctx.Request.Method
		userAgent := ctx.Request.UserAgent()
		errorMessage := ctx.Errors.ByType(gin.ErrorTypePrivate).String()
		bodySize := ctx.Writer.Size()

		event := log.Info()
		if statusCode >= 400 {
			event = log.Error()
		}

		event.
			Int("status", statusCode).
			Str("method", method).
			Str("path", path).
			Str("route", route).
			Str("request_id", requestID).
			Str("ip", clientIP).
			Str("user_agent", userAgent).
			Dur("duration", duration).
			Int("body_size", bodySize).
			Str("error", errorMessage).
			Msg("HTTP request")
	}
}
