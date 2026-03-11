package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) getGlobalFeed(ctx *gin.Context) {
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	viewerID := optionalViewerID(ctx)
	tweets, err := server.feedUC.GetGlobalFeed(ctx, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newTweetResponseList(tweets)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

func (server *Server) getFollowingFeed(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	tweets, err := server.feedUC.GetFollowingFeed(ctx, userID, page, size+1)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newTweetResponseList(tweets)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

func (server *Server) getUserFeed(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}
	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size
	viewerID := optionalViewerID(ctx)
	tweets, err := server.feedUC.GetUserFeed(ctx, req.ID, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	response := newTweetResponseList(tweets)
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}
