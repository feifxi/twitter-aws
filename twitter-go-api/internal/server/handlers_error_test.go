package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	db "github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// --- Additional mocks for error-path testing ---

type mockSearchUC struct {
	searchUsersFn    func(ctx context.Context, q string, page, size int32, viewerID *int64) ([]usecase.UserItem, error)
	searchTweetsFn   func(ctx context.Context, q string, page, size int32, viewerID *int64) ([]usecase.TweetItem, error)
	searchHashtagsFn func(ctx context.Context, q string, limit int32) ([]db.Hashtag, error)
}

func (m *mockSearchUC) SearchUsers(ctx context.Context, q string, page, size int32, viewerID *int64) ([]usecase.UserItem, error) {
	if m.searchUsersFn != nil {
		return m.searchUsersFn(ctx, q, page, size, viewerID)
	}
	return nil, nil
}
func (m *mockSearchUC) SearchTweets(ctx context.Context, q string, page, size int32, viewerID *int64) ([]usecase.TweetItem, error) {
	if m.searchTweetsFn != nil {
		return m.searchTweetsFn(ctx, q, page, size, viewerID)
	}
	return nil, nil
}
func (m *mockSearchUC) SearchHashtags(ctx context.Context, q string, limit int32) ([]db.Hashtag, error) {
	if m.searchHashtagsFn != nil {
		return m.searchHashtagsFn(ctx, q, limit)
	}
	return nil, nil
}

// --- getTweet error paths ---

// mockTweetUCNotFound returns pgx.ErrNoRows for GetTweet.
type mockTweetUCNotFound struct{ mockTweetUC }

func (m *mockTweetUCNotFound) GetTweet(_ context.Context, _ int64, _ *int64) (usecase.TweetItem, error) {
	return usecase.TweetItem{}, pgx.ErrNoRows
}

func TestGetTweet_NotFoundReturns404(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodGet, "/api/v1/tweets/99", nil, "")
	ctx.Params = gin.Params{{Key: "id", Value: "99"}}

	s := &Server{tweetUC: &mockTweetUCNotFound{}}
	s.getTweet(ctx)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing tweet, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["code"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND code, got %v", out["code"])
	}
}

// --- createTweet error paths ---

func TestCreateTweet_UsecaseErrorReturnsInternalError(t *testing.T) {
	t.Parallel()

	mock := &mockTweetUC{
		createTweetFn: func(_ context.Context, _ usecase.CreateTweetInput) (usecase.TweetItem, error) {
			return usecase.TweetItem{}, apperr.Internal("db failure", errors.New("connection reset"))
		},
	}

	req := mustBuildMultipartForm(http.MethodPost, "/api/v1/tweets", map[string]string{"content": "hello"})
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	setAuthorizedUser(ctx, 1)

	s := &Server{tweetUC: mock}
	s.createTweet(ctx)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for usecase error, got %d body=%s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["code"] != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR code, got %v", out["code"])
	}
}

func TestCreateTweetWithoutMedia_SuccessPath(t *testing.T) {
	t.Parallel()

	mock := &mockTweetUC{
		createTweetFn: func(_ context.Context, input usecase.CreateTweetInput) (usecase.TweetItem, error) {
			content := "hello"
			return usecase.TweetItem{
				ID:        77,
				Content:   &content,
				CreatedAt: time.Now(),
				Author:    usecase.UserItem{ID: input.UserID, Username: "tester", Email: "t@t.com"},
			}, nil
		},
	}

	req := mustBuildMultipartForm(http.MethodPost, "/api/v1/tweets", map[string]string{"content": "hello"})
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	setAuthorizedUser(ctx, 3)

	s := &Server{tweetUC: mock}
	s.createTweet(ctx)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["id"] != float64(77) {
		t.Fatalf("expected id=77, got %v", out["id"])
	}
}

// --- followUser guard ---

func TestFollowUser_CannotFollowSelf(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/users/5/follow", nil, "")
	ctx.Params = gin.Params{{Key: "id", Value: "5"}}
	setAuthorizedUser(ctx, 5)

	s := &Server{userUC: &mockUserUC{}}
	s.followUser(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for self-follow, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "cannot follow yourself") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

// --- unauthenticated protected routes ---

func TestDeleteTweet_MissingAuth(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodDelete, "/api/v1/tweets/1", nil, "")
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}

	s := &Server{tweetUC: &mockTweetUC{}}
	s.deleteTweet(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated delete, got %d", rec.Code)
	}
}

func TestGetUnreadNotificationCount_MissingAuth(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodGet, "/api/v1/notifications/unread-count", nil, "")

	s := &Server{sseClients: make(map[int64][]*sseClient)}
	s.getUnreadNotificationCount(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMarkNotificationRead_MissingAuth(t *testing.T) {
	t.Parallel()

	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/notifications/mark-read", nil, "")

	s := &Server{sseClients: make(map[int64][]*sseClient)}
	s.markNotificationRead(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// --- feed error propagation ---

func TestGetGlobalFeed_UsecaseErrorPropagates(t *testing.T) {
	t.Parallel()

	mock := &mockFeedUC{
		getGlobalFeedFn: func(_ context.Context, _, _ int32, _ *int64) ([]usecase.TweetItem, error) {
			return nil, apperr.Internal("feed failure", errors.New("db error"))
		},
	}

	ctx, rec := newHandlerTestContext(http.MethodGet, "/api/v1/feeds/global", nil, "")
	s := &Server{feedUC: mock}
	s.getGlobalFeed(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for feed error, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- search hashtags ---

func TestSearchHashtags_ReturnsResults(t *testing.T) {
	t.Parallel()

	mock := &mockSearchUC{
		searchHashtagsFn: func(_ context.Context, _ string, _ int32) ([]db.Hashtag, error) {
			return []db.Hashtag{{Text: "golang"}}, nil
		},
	}

	ctx, rec := newHandlerTestContext(http.MethodGet, "/api/v1/search/hashtags?q=go", nil, "")
	s := &Server{searchUC: mock}
	s.searchHashtags(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v body=%s", err, rec.Body.String())
	}
	if len(out) != 1 || out[0]["text"] != "golang" {
		t.Fatalf("unexpected response: %v", out)
	}
}

// mustBuildMultipartForm builds a multipart form POST request with given string fields.
func mustBuildMultipartForm(method, path string, fields map[string]string) *http.Request {
	boundary := "testboundary"
	var body strings.Builder
	for k, v := range fields {
		body.WriteString("--" + boundary + "\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"" + k + "\"\r\n\r\n")
		body.WriteString(v + "\r\n")
	}
	body.WriteString("--" + boundary + "--\r\n")

	req, err := http.NewRequest(method, path, strings.NewReader(body.String()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	return req
}
