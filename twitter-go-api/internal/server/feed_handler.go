package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) getGlobalFeed(ctx *gin.Context) {
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}
	tweets, err := server.usecase.GetGlobalFeed(ctx, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountGlobalFeed(ctx)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]tweetResponse, 0, len(tweets))
	for _, t := range tweets {
		response = append(response, newTweetResponse(t))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}

func (server *Server) getFollowingFeed(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}
	tweets, err := server.usecase.GetFollowingFeed(ctx, userID, page, size)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountFollowingFeed(ctx, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]tweetResponse, 0, len(tweets))
	for _, t := range tweets {
		response = append(response, newTweetResponse(t))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}

func (server *Server) getUserFeed(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
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
	tweets, err := server.usecase.GetUserFeed(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountUserFeed(ctx, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := make([]tweetResponse, 0, len(tweets))
	for _, t := range tweets {
		response = append(response, newTweetResponse(t))
	}
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}
