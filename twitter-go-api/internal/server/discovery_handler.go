package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getTrendingHashtags godoc
// @Summary		Trending Hashtags
// @Description	Get a list of currently trending hashtags based on recent usage.
// @Tags			Discovery
// @Produce		json
// @Param			limit	query		int		false	"Max number of results"
// @Success		200		{array}		HashtagResponse
// @Router			/discovery/trending [get]
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

// getSuggestedUsers godoc
// @Summary		Suggested Users
// @Description	Get a paginated list of suggested users to follow.
// @Tags			Discovery
// @Produce		json
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[UserResponse]
// @Router			/discovery/users [get]
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
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}
