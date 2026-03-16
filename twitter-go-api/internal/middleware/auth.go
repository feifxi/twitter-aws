package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := resolveAccessToken(ctx)
		if err != nil {
			abortWithError(ctx, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			abortWithError(ctx, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired access token")
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

// OptionalAuthMiddleware decodes access tokens when present, but never blocks the request.
func OptionalAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := resolveAccessToken(ctx)
		if err != nil {
			ctx.Next()
			return
		}

		payload, err := tokenMaker.VerifyToken(accessToken)
		if err == nil {
			ctx.Set(AuthorizationPayloadKey, payload)
		}

		ctx.Next()
	}
}

func resolveAccessToken(ctx *gin.Context) (string, error) {
	authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
	if len(authorizationHeader) == 0 {
		return "", errors.New("authorization header is not provided")
	}

	fields := strings.Fields(authorizationHeader)
	if len(fields) >= 2 && strings.ToLower(fields[0]) == authorizationTypeBearer {
		return fields[1], nil
	}

	return "", errors.New("invalid authorization header format")
}
