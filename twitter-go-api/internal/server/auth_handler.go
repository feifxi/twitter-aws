package server

import (
	"net/http"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/gin-gonic/gin"
)

type googleLoginRequest struct {
	IdToken string `json:"idToken" binding:"required"`
}

func (server *Server) loginGoogle(ctx *gin.Context) {
	var req googleLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, err)
		return
	}

	authData, err := server.authUC.LoginWithGoogle(ctx, req.IdToken)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newAuthResponse(authData.AccessToken, authData.RefreshToken, authData.User))
}

func (server *Server) refreshToken(ctx *gin.Context) {
	refreshToken := resolveRefreshToken(ctx)
	if refreshToken == "" {
		writeError(ctx, apperr.Unauthorized("missing refresh token"))
		return
	}

	authData, err := server.authUC.RefreshSession(ctx, refreshToken)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newAuthResponse(authData.AccessToken, authData.RefreshToken, authData.User))
}

func (server *Server) logout(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if ok {
		server.authUC.Logout(ctx, &userID, nil)
	} else if rt := resolveRefreshToken(ctx); rt != "" {
		server.authUC.Logout(ctx, nil, &rt)
	}

	ctx.JSON(http.StatusOK, successResponse())
}

func (server *Server) getMe(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	user, err := server.authUC.GetMe(ctx, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// resolveRefreshToken reads the refresh token from the JSON body field.
func resolveRefreshToken(ctx *gin.Context) string {
	var body refreshTokenRequest
	if ctx.ShouldBindJSON(&body) == nil && strings.TrimSpace(body.RefreshToken) != "" {
		return body.RefreshToken
	}
	return ""
}

func (server *Server) cookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(server.config.CookieSameSite)) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}