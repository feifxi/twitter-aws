package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = req
	return c
}

func TestResolveAccessTokenFromCookie(t *testing.T) {
	t.Parallel()

	ctx := newTestContext()
	ctx.Request.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	ctx.Request.Header.Set("Authorization", "Bearer header-token")

	got, err := resolveAccessToken(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "cookie-token" {
		t.Fatalf("expected cookie token, got %q", got)
	}
}

func TestResolveAccessTokenFromHeader(t *testing.T) {
	t.Parallel()

	ctx := newTestContext()
	ctx.Request.Header.Set("Authorization", "Bearer header-token")

	got, err := resolveAccessToken(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "header-token" {
		t.Fatalf("expected header token, got %q", got)
	}
}

func TestResolveAccessTokenInvalidHeader(t *testing.T) {
	t.Parallel()

	ctx := newTestContext()
	ctx.Request.Header.Set("Authorization", "Token invalid")

	if _, err := resolveAccessToken(ctx); err == nil {
		t.Fatal("expected invalid header error")
	}
}
