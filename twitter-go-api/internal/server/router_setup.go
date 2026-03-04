package server

import (
	"reflect"
	"strings"
	"time"

	"github.com/chanombude/twitter-go-api/internal/logger"
	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	// "github.com/rs/zerolog/log"
)

func (server *Server) setupRouter() {
	configureValidationFieldNames()

	router := gin.New()
	// if err := router.SetTrustedProxies(parseTrustedProxies(server.config.TrustedProxies)); err != nil {
	// 	log.Warn().Err(err).Msg("Failed to set trusted proxies, falling back to default proxy behavior")
	// }
	if server.config.MaxMultipartMemoryBytes > 0 {
		router.MaxMultipartMemory = server.config.MaxMultipartMemoryBytes
	}
	router.Use(middleware.RequestID())
	router.Use(logger.GinMiddleware())
	router.Use(gin.Recovery())

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

	api := router.Group("/api/v1")
	api.Use(middleware.RateLimiterWithRedis(server.redis, 20, 60, "rl:default"))
	server.registerDomainRoutes(api)

	server.router = router
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

func parseTrustedProxies(raw string) []string {
	return parseAllowedOrigins(raw)
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
