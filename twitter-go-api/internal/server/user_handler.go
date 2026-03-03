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
		if server.config.MaxAvatarBytes > 0 && req.Avatar.Size > server.config.MaxAvatarBytes {
			writeValidationError(ctx, "avatar", "file size exceeds limit")
			return
		}

		if !hasAllowedExtension(req.Avatar.Filename, avatarAllowedExts) {
			writeValidationError(ctx, "avatar", "unsupported file extension")
			return
		}

		file, reader, detectedContentType, err := openAndDetectUpload(req.Avatar)
		if err != nil {
			writeError(ctx, apperr.BadRequest("failed to inspect avatar file"))
			return
		}
		defer file.Close()

		if !isAllowedType(detectedContentType, avatarAllowedMIMEs) {
			writeValidationError(ctx, "avatar", "unsupported avatar type")
			return
		}

		input.Avatar = &usecase.AvatarUpload{
			Filename:    req.Avatar.Filename,
			ContentType: detectedContentType,
			Reader:      reader,
		}
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

	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}

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
	page, size, ok := parsePageAndSize(ctx)
	if !ok {
		return
	}

	var viewerID *int64
	if id, ok := getCurrentUserID(ctx); ok {
		viewerID = &id
	}

	users, err := server.userUC.ListFollowers(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.userUC.CountFollowers(ctx, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newUserResponseList(users)
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

	users, err := server.userUC.ListFollowing(ctx, req.ID, page, size, viewerID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	total, err := server.userUC.CountFollowing(ctx, req.ID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response := newUserResponseList(users)
	ctx.JSON(http.StatusOK, buildPageResponse(response, page, size, total))
}
