package middleware

import (
	"github.com/chanombude/twitter-go-api/internal/apiresponse"
	"github.com/gin-gonic/gin"
)

func abortWithError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, apiresponse.Error{
		Code:    code,
		Message: message,
	})
}
