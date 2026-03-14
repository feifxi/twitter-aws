package server

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/chanombude/twitter-go-api/internal/config"
)

type mockStorageService struct {
	generatePresignedURLFn func(ctx context.Context, filename, contentType, folder string) (string, string, error)
}

func (m *mockStorageService) GeneratePresignedURL(ctx context.Context, filename, contentType, folder string) (string, string, error) {
	return m.generatePresignedURLFn(ctx, filename, contentType, folder)
}
func (m *mockStorageService) DeleteFile(ctx context.Context, objectKey string) error { return nil }
func (m *mockStorageService) PublicURL(objectKey string) string                      { return "" }

func TestPresignUpload_BannersFolderAllowed(t *testing.T) {
	t.Parallel()

	mock := &mockStorageService{
		generatePresignedURLFn: func(_ context.Context, filename, contentType, folder string) (string, string, error) {
			if folder != "banners" {
				t.Fatalf("expected folder 'banners', got %q", folder)
			}
			return "http://presigned.url", "banners/uuid_file.png", nil
		},
	}

	reqBody := `{"filename":"banner.png","contentType":"image/png","folder":"banners"}`
	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/uploads/presign", bytes.NewBufferString(reqBody), "application/json")
	setAuthorizedUser(ctx, 1)

	s := &Server{
		storage: mock,
		config: config.Config{
			MaxBannerBytes: 10 << 20, // 10MB
		},
	}
	s.presignUpload(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestPresignUpload_BannersSizeLimit(t *testing.T) {
	t.Parallel()

	mock := &mockStorageService{
		generatePresignedURLFn: func(_ context.Context, filename, contentType, folder string) (string, string, error) {
			return "http://presigned.url", "banners/uuid_file.png", nil
		},
	}

	// 11 MiB > 10 MiB limit
	reqBody := `{"filename":"big.png","contentType":"image/png","folder":"banners","contentLength":11534336}`
	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/uploads/presign", bytes.NewBufferString(reqBody), "application/json")
	setAuthorizedUser(ctx, 1)

	s := &Server{
		storage: mock,
		config: config.Config{
			MaxBannerBytes: 10 << 20, // 10MB
		},
	}
	s.presignUpload(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized file, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestPresignUpload_InvalidFolderRejected(t *testing.T) {
	t.Parallel()

	reqBody := `{"filename":"file.png","contentType":"image/png","folder":"malicious"}`
	ctx, rec := newHandlerTestContext(http.MethodPost, "/api/v1/uploads/presign", bytes.NewBufferString(reqBody), "application/json")
	setAuthorizedUser(ctx, 1)

	s := &Server{}
	s.presignUpload(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported folder, got %d, body=%s", rec.Code, rec.Body.String())
	}
}
