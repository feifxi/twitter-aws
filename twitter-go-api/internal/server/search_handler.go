package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type searchQueryRequest struct {
	Query string `form:"q" binding:"required"`
}

func (server *Server) searchUsers(ctx *gin.Context) {
	var req searchQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		writeError(ctx, err)
		return
	}
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	viewerID := optionalViewerID(ctx)
	users, err := server.searchUC.SearchUsers(ctx, req.Query, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newUserResponseList(users)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

func (server *Server) searchTweets(ctx *gin.Context) {
	var req searchQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		writeError(ctx, err)
		return
	}
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	viewerID := optionalViewerID(ctx)
	tweets, err := server.searchUC.SearchTweets(ctx, req.Query, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newTweetResponseList(tweets)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

func (server *Server) searchHashtags(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	limit := parseLimit(ctx, 5, maxSize)
	hashtags, err := server.searchUC.SearchHashtags(ctx, query, limit)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newHashtagResponseList(hashtags)
	ctx.JSON(http.StatusOK, response)
}
