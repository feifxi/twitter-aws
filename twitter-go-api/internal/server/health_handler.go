package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// healthz godoc
// @Summary		Liveness Probe
// @Description	Simple health check to verify the API server is running and reachable.
// @Tags			Monitoring
// @Produce		json
// @Success		200	{object}	map[string]string	"status ok"
// @Router			/healthz [get]
func (server *Server) healthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyz godoc
// @Summary		Readiness Probe
// @Description	Comprehensive readiness check that verifies active connections to the database and Redis cache.
// @Tags			Monitoring
// @Produce		json
// @Success		200	{object}	map[string]string	"status ready"
// @Failure		503	{object}	map[string]string	"dependency failure"
// @Router			/readyz [get]
func (server *Server) readyz(ctx *gin.Context) {
	checkCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()

	if err := server.store.Ping(checkCtx); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"service": "database",
		})
		return
	}

	if server.redis != nil {
		if err := server.redis.Ping(checkCtx).Err(); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"service": "redis",
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
}
