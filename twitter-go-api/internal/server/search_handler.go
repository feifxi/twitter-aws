package server

import (
	"net/http"
	"strconv"
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
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	users, err := server.usecase.SearchUsers(ctx, req.Query, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountSearchUsers(ctx, req.Query)
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

func (server *Server) searchTweets(ctx *gin.Context) {
	var req searchQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		writeError(ctx, err)
		return
	}
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	tweets, err := server.usecase.SearchTweets(ctx, req.Query, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountSearchTweets(ctx, req.Query)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]tweetResponse, 0, len(tweets))
	for _, tweet := range tweets {
		response = append(response, newTweetResponse(tweet))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}

func (server *Server) searchHashtags(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	limit := int32(5)
	if rawLimit := strings.TrimSpace(ctx.Query("limit")); rawLimit != "" {
		if parsed, err := strconv.ParseInt(rawLimit, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}
	if limit > maxSize {
		limit = maxSize
	}
	hashtags, err := server.usecase.SearchHashtags(ctx, query, limit)
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
