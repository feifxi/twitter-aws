package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	zlog "github.com/rs/zerolog/log"
)

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type apiErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []fieldError `json:"details,omitempty"`
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email format"
	case "min":
		return fmt.Sprintf("must be at least %s units", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s units", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	case "numeric":
		return "must be a valid number"
	case "required_without":
		return fmt.Sprintf("this field is required if %s is not provided", fe.Param())
	}
	return "invalid value"
}

func writeError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		out := make([]fieldError, len(validationErrs))
		for i, fe := range validationErrs {
			out[i] = fieldError{
				Field:   fe.Field(),
				Message: msgForTag(fe),
			}
		}
		ctx.JSON(http.StatusBadRequest, apiErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "invalid request payload",
			Details: out,
		})
		return
	}

	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal server error"

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case "unique_violation":
			status = http.StatusConflict
			code = "CONFLICT"
			message = "resource already exists"
		case "foreign_key_violation":
			status = http.StatusBadRequest
			code = "BAD_REQUEST"
			message = "referenced resource does not exist"
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		status = http.StatusNotFound
		code = "NOT_FOUND"
		message = "resource not found"
	} else if kind, ok := apperr.KindOf(err); ok {
		switch kind {
		case apperr.KindBadRequest:
			status = http.StatusBadRequest
			code = "BAD_REQUEST"
			message = defaultMessage(apperr.MessageOf(err), "invalid request")
		case apperr.KindUnauthorized:
			status = http.StatusUnauthorized
			code = "UNAUTHORIZED"
			message = defaultMessage(apperr.MessageOf(err), "authentication required")
		case apperr.KindForbidden:
			status = http.StatusForbidden
			code = "FORBIDDEN"
			message = defaultMessage(apperr.MessageOf(err), "forbidden")
		case apperr.KindNotFound:
			status = http.StatusNotFound
			code = "NOT_FOUND"
			message = defaultMessage(apperr.MessageOf(err), "resource not found")
		case apperr.KindConflict:
			status = http.StatusConflict
			code = "CONFLICT"
			message = defaultMessage(apperr.MessageOf(err), "conflict")
		case apperr.KindInternal:
			status = http.StatusInternalServerError
			code = "INTERNAL_ERROR"
			message = "internal server error"
		}
	}

	if status >= 500 {
		zlog.Error().Err(err).
			Str("method", ctx.Request.Method).
			Str("path", ctx.FullPath()).
			Msg("Internal server error")
	}

	ctx.JSON(status, apiErrorResponse{Code: code, Message: message})
}

func defaultMessage(in, fallback string) string {
	if in == "" {
		return fallback
	}
	return in
}

func writeValidationError(ctx *gin.Context, field, message string) {
	ctx.JSON(http.StatusBadRequest, apiErrorResponse{
		Code:    "VALIDATION_ERROR",
		Message: "invalid request payload",
		Details: []fieldError{
			{Field: field, Message: message},
		},
	})
}
