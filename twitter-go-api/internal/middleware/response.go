package middleware

import "github.com/gin-gonic/gin"

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type apiErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []fieldError `json:"details,omitempty"`
}

func abortWithError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, apiErrorResponse{
		Code:    code,
		Message: message,
	})
}
