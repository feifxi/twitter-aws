package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/chanombude/twitter-go-api/internal/middleware"
	"github.com/chanombude/twitter-go-api/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newHandlerTestContext(method, path string, body *bytes.Buffer, contentType string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	reqBody := bytes.NewReader(nil)
	if body != nil {
		reqBody = bytes.NewReader(body.Bytes())
	}
	req := httptest.NewRequest(method, path, reqBody)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	ctx.Request = req
	return ctx, w
}

func setAuthorizedUser(ctx *gin.Context, userID int64) {
	ctx.Set(middleware.AuthorizationPayloadKey, &token.Payload{
		ID:        uuid.New(),
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(time.Hour),
	})
}

func TestCreateTweetRejectsUnsupportedMediaExtension(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("content", "hello")

	part, err := writer.CreateFormFile("media", "payload.exe")
	if err != nil {
		t.Fatalf("failed to create media part: %v", err)
	}
	if _, err := part.Write([]byte("MZ fake")); err != nil {
		t.Fatalf("failed to write media part: %v", err)
	}
	_ = writer.Close()

	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/tweets", &body, writer.FormDataContentType())
	setAuthorizedUser(ctx, 1)

	s := &Server{config: config.Config{MaxMediaBytes: 10 << 20}}
	s.createTweet(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"VALIDATION_ERROR"`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestCreateTweetRejectsInvalidDetectedMediaType(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("content", "hello")

	part, err := writer.CreateFormFile("media", "payload.png")
	if err != nil {
		t.Fatalf("failed to create media part: %v", err)
	}
	// Not real PNG bytes, should be detected as text/plain.
	if _, err := part.Write([]byte("not-an-image")); err != nil {
		t.Fatalf("failed to write media part: %v", err)
	}
	_ = writer.Close()

	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/tweets", &body, writer.FormDataContentType())
	setAuthorizedUser(ctx, 1)

	s := &Server{config: config.Config{MaxMediaBytes: 10 << 20}}
	s.createTweet(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unsupported media type") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestUpdateProfileRejectsUnsupportedAvatarExtension(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("displayName", "tester")

	part, err := writer.CreateFormFile("avatar", "avatar.svg")
	if err != nil {
		t.Fatalf("failed to create avatar part: %v", err)
	}
	if _, err := part.Write([]byte("<svg></svg>")); err != nil {
		t.Fatalf("failed to write avatar part: %v", err)
	}
	_ = writer.Close()

	ctx, rec := newHandlerTestContext(http.MethodPut, "/api/v1/users/profile", &body, writer.FormDataContentType())
	setAuthorizedUser(ctx, 1)

	s := &Server{config: config.Config{MaxAvatarBytes: 5 << 20}}
	s.updateProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unsupported file extension") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestRefreshTokenMissingCookie(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/auth/refresh", nil, "")

	s := &Server{}
	s.refreshToken(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if got["code"] != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED code, got %v", got["code"])
	}
}

func TestLogoutWithoutSessionStillSucceeds(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/auth/logout", nil, "")

	s := &Server{}
	s.logout(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}
