package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getGlobalFeed godoc
// @Summary		Global Feed
// @Description	Get a paginated list of all tweets.
// @Tags			Feeds
// @Produce		json
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[TweetResponse]
// @Router			/feeds/global [get]
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
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// getFollowingFeed godoc
// @Summary		Following Feed
// @Description	Get a paginated list of tweets from users the current user follows.
// @Tags			Feeds
// @Produce		json
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[TweetResponse]
// @Security		BearerAuth
// @Failure		401		{object}	ErrorResponse
// @Router			/feeds/following [get]
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
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// getUserFeed godoc
// @Summary		User Profile Feed
// @Description	Get a paginated list of tweets for a specific user profile.
// @Tags			Feeds
// @Produce		json
// @Param			id		path		int		true	"User ID"
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[TweetResponse]
// @Router			/feeds/user/{id} [get]
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
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}
