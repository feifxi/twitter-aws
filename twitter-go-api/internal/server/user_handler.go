package server

import (
	"net/http"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type updateProfileRequest struct {
	DisplayName *string `json:"displayName" binding:"omitempty,max=30"`
	Bio         *string `json:"bio" binding:"omitempty,max=160"`
	AvatarKey   *string `json:"avatarKey" binding:"omitempty"`
	BannerKey   *string `json:"bannerKey" binding:"omitempty"`
}

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
	if followerID == req.ID {
		writeError(ctx, apperr.BadRequest("cannot follow yourself"))
		return
	}

	_, err := server.userUC.FollowUser(ctx, followerID, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, successResponse())
}

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
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}

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
	ctx.JSON(http.StatusOK, buildPageResponse(response, size, offset))
}
