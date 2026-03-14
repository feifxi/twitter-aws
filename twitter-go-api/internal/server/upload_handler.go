package server

import (
	"net/http"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/gin-gonic/gin"
)

var (
	presignAllowedMIMEs = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"video/mp4":  true,
		"video/webm": true,
	}

	presignAllowedFolders = map[string]bool{
		"tweets":  true,
		"avatars": true,
		"banners": true,
	}
)

type presignRequest struct {
	Filename      string `json:"filename" binding:"required"`
	ContentType   string `json:"contentType" binding:"required"`
	Folder        string `json:"folder" binding:"required"`
	ContentLength *int64 `json:"contentLength" binding:"omitempty,min=1"`
}

func (server *Server) presignUpload(ctx *gin.Context) {
	if _, ok := mustCurrentUserID(ctx); !ok {
		return
	}

	var req presignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	contentType := strings.ToLower(strings.TrimSpace(req.ContentType))
	if !presignAllowedMIMEs[contentType] {
		writeValidationError(ctx, "contentType", "unsupported content type")
		return
	}

	folder := strings.ToLower(strings.TrimSpace(req.Folder))
	if !presignAllowedFolders[folder] {
		writeValidationError(ctx, "folder", "unsupported folder")
		return
	}

	if req.ContentLength != nil {
		maxSize := server.maxUploadSize(folder)
		if maxSize > 0 && *req.ContentLength > maxSize {
			writeValidationError(ctx, "contentLength", "file too large")
			return
		}
	}

	presignedURL, objectKey, err := server.storage.GeneratePresignedURL(ctx, req.Filename, contentType, folder)
	if err != nil {
		writeError(ctx, apperr.Internal("failed to generate upload URL", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"presignedUrl": presignedURL,
		"objectKey":    objectKey,
	})
}

func (server *Server) maxUploadSize(folder string) int64 {
	switch folder {
	case "avatars":
		return server.config.MaxAvatarBytes
	case "banners":
		return server.config.MaxBannerBytes
	case "tweets":
		return server.config.MaxMediaBytes
	default:
		return 0
	}
}
