package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) getTrendingHashtags(ctx *gin.Context) {
	limit := parseLimit(ctx, 10, maxSize)
	hashtags, err := server.discoveryUC.GetTrendingHashtags(ctx, limit)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newHashtagResponseList(hashtags)
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) getSuggestedUsers(ctx *gin.Context) {
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	viewerID := optionalViewerID(ctx)
	users, err := server.discoveryUC.GetSuggestedUsers(ctx, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newUserResponseList(users)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}
