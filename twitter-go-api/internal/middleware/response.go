package middleware

import (
	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/gin-gonic/gin"
)

func abortWithError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, apperr.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
