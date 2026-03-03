package server

import (
	"mime/multipart"
	"net/http"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

type updateProfileRequest struct {
	DisplayName *string               `form:"displayName" binding:"omitempty,max=30"`
	Bio         *string               `form:"bio" binding:"omitempty,max=160"`
	Avatar      *multipart.FileHeader `form:"avatar"`
}

func (server *Server) updateProfile(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	var req updateProfileRequest
	if err := ctx.ShouldBind(&req); err != nil {
		writeError(ctx, err)
		return
	}

	input := usecase.UpdateProfileInput{
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
	}

	// Check if avatar is provided
	if req.Avatar != nil {
		file, err := req.Avatar.Open()
		if err != nil {
			writeError(ctx, apperr.BadRequest("failed to open avatar file"))
			return
		}
		defer file.Close()

		contentType := req.Avatar.Header.Get("Content-Type")

		input.Avatar = &usecase.AvatarUpload{
			Filename:    req.Avatar.Filename,
			ContentType: contentType,
			Reader:      file,
		}
	}

	updatedUser, err := server.usecase.UpdateProfile(ctx, userID, input)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(usecase.UserItem{User: updatedUser}))
}

func (server *Server) getUser(ctx *gin.Context) {
	var req idURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		writeError(ctx, err)
		return
	}

	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}

	user, err := server.usecase.GetUser(ctx, req.ID, viewerID)
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

	_, err := server.usecase.FollowUser(ctx, followerID, req.ID)
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

	if err := server.usecase.UnfollowUser(ctx, followerID, req.ID); err != nil {
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
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}

	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}

	users, err := server.usecase.ListFollowers(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountFollowers(ctx, req.ID)
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

func (server *Server) listFollowing(ctx *gin.Context) {
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

	users, err := server.usecase.ListFollowing(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.usecase.CountFollowing(ctx, req.ID)
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
