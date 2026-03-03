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

	authData, err := server.usecase.LoginWithGoogle(ctx, req.IdToken)
	if err != nil {
		writeError(ctx, err)
		return
	}

	server.setSessionCookies(ctx, authData.AccessToken, authData.RefreshToken)
	ctx.JSON(http.StatusOK, newAuthResponse(authData.AccessToken, authData.User))
}

func (server *Server) refreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		writeError(ctx, apperr.Unauthorized("missing refresh token"))
		return
	}

	authData, err := server.usecase.RefreshSession(ctx, refreshToken)
	if err != nil {
		writeError(ctx, err)
		return
	}

	server.setSessionCookies(ctx, authData.AccessToken, authData.RefreshToken)
	ctx.JSON(http.StatusOK, newAuthResponse(authData.AccessToken, authData.User))
}

func (server *Server) logout(ctx *gin.Context) {
	userID, ok := getCurrentUserID(ctx)
	if ok {
		server.usecase.Logout(ctx, &userID, nil)
	} else if refreshToken, err := ctx.Cookie("refresh_token"); err == nil {
		server.usecase.Logout(ctx, nil, &refreshToken)
	}

	server.clearSessionCookies(ctx)
	ctx.JSON(http.StatusOK, successResponse())
}

func (server *Server) setSessionCookies(ctx *gin.Context, accessToken, refreshToken string) {
	ctx.SetSameSite(server.cookieSameSite())
	ctx.SetCookie(
		"access_token",
		accessToken,
		server.config.TokenDurationMinutes*60,
		"/",
		server.config.CookieDomain,
		server.config.CookieSecure,
		true,
	)

	ctx.SetSameSite(server.cookieSameSite())
	ctx.SetCookie(
		"refresh_token",
		refreshToken,
		server.config.RefreshTokenDurationDays*24*60*60,
		"/api/v1/auth/refresh",
		server.config.CookieDomain,
		server.config.CookieSecure,
		true,
	)
}

func (server *Server) getMe(ctx *gin.Context) {
	userID, ok := mustCurrentUserID(ctx)
	if !ok {
		return
	}

	user, err := server.usecase.GetMe(ctx, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}

func (server *Server) clearSessionCookies(ctx *gin.Context) {
	ctx.SetSameSite(server.cookieSameSite())
	ctx.SetCookie("access_token", "", -1, "/", server.config.CookieDomain, server.config.CookieSecure, true)

	ctx.SetSameSite(server.cookieSameSite())
	ctx.SetCookie("refresh_token", "", -1, "/api/v1/auth/refresh", server.config.CookieDomain, server.config.CookieSecure, true)
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
