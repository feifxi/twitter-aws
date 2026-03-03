package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (server *Server) getTrendingHashtags(ctx *gin.Context) {
	limit := int32(10)
	if rawLimit := strings.TrimSpace(ctx.Query("limit")); rawLimit != "" {
		if parsed, err := strconv.ParseInt(rawLimit, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}
	if limit > maxSize {
		limit = maxSize
	}
	hashtags, err := server.usecase.GetTrendingHashtags(ctx, limit)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]hashtagResponse, 0, len(hashtags))
	for _, h := range hashtags {
		response = append(response, newHashtagResponse(h))
	}
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) getSuggestedUsers(ctx *gin.Context) {
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	users, err := server.usecase.GetSuggestedUsers(ctx, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountSuggestedUsers(ctx, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]userResponse, 0, len(users))
	for _, user := range users {
		response = append(response, newUserResponse(user))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}
