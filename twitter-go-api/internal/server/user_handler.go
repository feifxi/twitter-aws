package server

import (
	"net/http"

	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type updateProfileRequest struct {
	DisplayName *string `json:"displayName" binding:"omitempty,max=30"`
	Bio         *string `json:"bio" binding:"omitempty,max=160"`
	AvatarKey   *string `json:"avatarKey" binding:"omitempty"`
	BannerKey   *string `json:"bannerKey" binding:"omitempty"`
}

// updateProfile godoc
// @Summary		Update User Profile
// @Description	Modify the display name, bio, or profile images of the authenticated user.
// @Tags			Users
// @Accept			json
// @Produce		json
// @Param			request	body			updateProfileRequest	true	"Profile Update Details"
// @Success		200		{object}	UserResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		401		{object}	ErrorResponse
// @Security		BearerAuth
// @Router			/users/profile [put]
func (server *Server) updateProfile(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var req updateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	input := usecase.UpdateProfileInput{
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		AvatarKey:   req.AvatarKey,
		BannerKey:   req.BannerKey,
	}

	updatedUser, err := server.userUC.UpdateProfile(ctx, userID, input)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(updatedUser))
}

// getUser godoc
// @Summary		Get User Profile
// @Description	Get public profile information for a user by their ID.
// @Tags			Users
// @Produce		json
// @Param			id	path		int64	true	"User ID"
// @Success		200	{object}	UserResponse
// @Failure		404	{object}	ErrorResponse
// @Router			/users/{id} [get]
func (server *Server) getUser(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}

	viewerID := optionalViewerID(ctx)

	user, err := server.userUC.GetUser(ctx, req.ID, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}

// followUser godoc
// @Summary		Follow User
// @Description	Establish a following relationship with another user.
// @Tags			Users
// @Produce		json
// @Param			id	path		int64	true	"User ID to follow"
// @Success		200	{object}	SuccessResponse
// @Failure		401	{object}	ErrorResponse
// @Security		BearerAuth
// @Router			/users/{id}/follow [post]
func (server *Server) followUser(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}

	followerID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	_, err := server.userUC.FollowUser(ctx, followerID, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
}

// unfollowUser godoc
// @Summary		Unfollow User
// @Description	Remove an existing following relationship.
// @Tags			Users
// @Produce		json
// @Param			id	path		int64	true	"User ID to unfollow"
// @Success		200	{object}	SuccessResponse
// @Failure		401	{object}	ErrorResponse
// @Security		BearerAuth
// @Router			/users/{id}/follow [delete]
func (server *Server) unfollowUser(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}

	followerID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	if err := server.userUC.UnfollowUser(ctx, followerID, req.ID); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
}

// listFollowers godoc
// @Summary		List Followers
// @Description	Get a paginated list of users following the specified user.
// @Tags			Users
// @Produce		json
// @Param			id		path		int		true	"User ID"
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[UserResponse]
// @Router			/users/{id}/followers [get]
func (server *Server) listFollowers(ctx *gin.Context) {
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
	users, err := server.userUC.ListFollowers(ctx, req.ID, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newUserResponseList(users)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}

// listFollowing godoc
// @Summary		List Following
// @Description	Get a paginated list of users followed by the specified user.
// @Tags			Users
// @Produce		json
// @Param			id		path		int		true	"User ID"
// @Param			cursor	query		string	false	"Pagination cursor"
// @Param			size	query		int		false	"Number of items per page"
// @Success		200		{object}	PageResponse[UserResponse]
// @Router			/users/{id}/following [get]
func (server *Server) listFollowing(ctx *gin.Context) {
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
	users, err := server.userUC.ListFollowing(ctx, req.ID, page, size+1, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newUserResponseList(users)
	ctx.JSON(http.StatusOK, BuildPageResponse(response, size, offset))
}
