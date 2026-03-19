package server

import (
	"net/http"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/gin-gonic/gin"
)

// searchUsers godoc
// @Summary		Search Users
// @Description	Search for users by username or display name with pagination.
// @Tags			Search
// @Produce		json
// @Param			q		query		string	true	"Search query"
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[UserResponse]
// @Router			/search/users [get]
func (server *Server) searchUsers(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	if query == "" {
		writeError(ctx, apperr.BadRequest("search query is required"))
		return
	}

	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size

	viewerID := optionalViewerID(ctx)
	users, err := server.searchUC.SearchUsers(ctx, query, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newUserResponseList(users)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// searchTweets godoc
// @Summary		Search Tweets
// @Description	Search for tweets by content (using full-text search) with pagination.
// @Tags			Search
// @Produce		json
// @Param			q		query		string	true	"Search query"
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[TweetResponse]
// @Router			/search/tweets [get]
func (server *Server) searchTweets(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	if query == "" {
		writeError(ctx, apperr.BadRequest("search query is required"))
		return
	}

	offset, size, ok := parseOffsetAndSize(ctx)
	if !ok {
		return
	}
	page := offset / size

	viewerID := optionalViewerID(ctx)
	tweets, err := server.searchUC.SearchTweets(ctx, query, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newTweetResponseList(tweets)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// searchHashtags godoc
// @Summary		Search Hashtags
// @Description	Find trending or specific hashtags matching a prefix.
// @Tags			Search
// @Produce		json
// @Param			q		query		string	true	"Hashtag prefix"
// @Param			limit	query		int		false	"Max number of results"
// @Success		200		{array}		HashtagResponse
// @Router			/search/hashtags [get]
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
