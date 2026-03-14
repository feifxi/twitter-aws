package usecase

import (
	"context"
	"testing"

	"github.com/chanombude/twitter-go-api/internal/db"
)

type mockStore struct {
	db.Querier
	getUserFn            func(ctx context.Context, arg db.GetUserParams) (db.GetUserRow, error)
	updateUserProfileFn func(ctx context.Context, arg db.UpdateUserProfileParams) (db.User, error)
}

func (m *mockStore) GetUser(ctx context.Context, arg db.GetUserParams) (db.GetUserRow, error) {
	return m.getUserFn(ctx, arg)
}

func (m *mockStore) UpdateUserProfile(ctx context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
	return m.updateUserProfileFn(ctx, arg)
}

func (m *mockStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	return nil
}

func (m *mockStore) ExecTxAfterCommit(ctx context.Context, fn func(*db.Queries) error, afterCommit func()) error {
	return nil
}

func (m *mockStore) Ping(ctx context.Context) error {
	return nil
}

type mockStorage struct {
	publicURLFn  func(key string) string
	deleteFileFn func(ctx context.Context, key string) error
}

func (m *mockStorage) PublicURL(key string) string {
	return m.publicURLFn(key)
}

func (m *mockStorage) DeleteFile(ctx context.Context, key string) error {
	return m.deleteFileFn(ctx, key)
}

func (m *mockStorage) GeneratePresignedURL(ctx context.Context, filename, contentType, folder string) (string, string, error) {
	return "", "", nil
}

func TestUpdateProfile_ClearingFields(t *testing.T) {
	ctx := context.Background()
	
	oldAvatar := "https://cdn.com/old_avatar.png"
	oldBanner := "https://cdn.com/old_banner.png"
	
	store := &mockStore{
		getUserFn: func(ctx context.Context, arg db.GetUserParams) (db.GetUserRow, error) {
			return db.GetUserRow{
				User: db.User{
					ID:        arg.ID,
					AvatarUrl: &oldAvatar,
					BannerUrl: &oldBanner,
				},
			}, nil
		},
		updateUserProfileFn: func(ctx context.Context, arg db.UpdateUserProfileParams) (db.User, error) {
			// Verify that URLs are nil when keys are empty strings
			if arg.AvatarUrl != nil {
				t.Errorf("expected AvatarUrl to be nil, got %v", *arg.AvatarUrl)
			}
			if arg.BannerUrl != nil {
				t.Errorf("expected BannerUrl to be nil, got %v", *arg.BannerUrl)
			}
			return db.User{ID: arg.ID}, nil
		},
	}

	storage := &mockStorage{
		deleteFileFn: func(ctx context.Context, key string) error {
			return nil
		},
	}

	uc := &UserUsecase{
		store:   store,
		storage: storage,
	}

	empty := ""
	input := UpdateProfileInput{
		AvatarKey: &empty,
		BannerKey: &empty,
	}

	_, err := uc.UpdateProfile(ctx, 1, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
