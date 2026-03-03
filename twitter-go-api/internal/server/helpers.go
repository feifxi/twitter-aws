package server

import (
	"math"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/chanombude/twitter-go-api/internal/token"
	"github.com/gin-gonic/gin"
)

const (
	defaultPage = int32(0)
	defaultSize = int32(20)
	maxSize     = int32(50)
)

type pageResponse[T any] struct {
	Content       []T   `json:"content"`
	Page          int32 `json:"page"`
	Size          int32 `json:"size"`
	TotalElements int64 `json:"totalElements"`
	TotalPages    int32 `json:"totalPages"`
	First         bool  `json:"first"`
	Last          bool  `json:"last"`
}

func buildPageResponse[T any](content []T, page, size int32, total int64) pageResponse[T] {
	totalPages := int32(0)
	if size > 0 && total > 0 {
		totalPages = int32(math.Ceil(float64(total) / float64(size)))
	}

	last := true
	if total > 0 {
		last = int64((page+1)*size) >= total
	}

	return pageResponse[T]{
		Content:       content,
		Page:          page,
		Size:          size,
		TotalElements: total,
		TotalPages:    totalPages,
		First:         page == 0,
		Last:          last,
	}
}

func getCurrentUserID(ctx *gin.Context) (int64, bool) {
	payload, ok := ctx.Get(middleware.AuthorizationPayloadKey)
	if !ok {
		return 0, false
	}
	authPayload, ok := payload.(*token.Payload)
	if !ok {
		return 0, false
	}
	return authPayload.UserID, true
}

func mustCurrentUserID(ctx *gin.Context) (int64, bool) {
	userID, ok := getCurrentUserID(ctx)
	if !ok {
		writeError(ctx, apperr.Unauthorized("authentication required"))
		return 0, false
	}
	return userID, true
}

func parsePageAndSize(ctx *gin.Context) (int32, int32, bool) {
	type paginationQuery struct {
		Page *int32 `form:"page" binding:"omitempty,min=0"`
		Size *int32 `form:"size" binding:"omitempty,min=1"`
	}

	var req paginationQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		writeError(ctx, err)
		return 0, 0, false
	}

	page := defaultPage
	if req.Page != nil {
		page = *req.Page
	}

	size := defaultSize
	if req.Size != nil {
		size = min(*req.Size, maxSize)
	}

	return page, size, true
}
