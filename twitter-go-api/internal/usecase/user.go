package usecase

import (
	"context"
	"database/sql"
	"strings"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/db"
)

type AvatarUpload struct {
	Filename    string
	ContentType string
	Reader      interface {
		Read(p []byte) (n int, err error)
	}
}

type UpdateProfileInput struct {
	Bio         *string
	DisplayName *string
	Avatar      *AvatarUpload
}

func (u *Usecase) GetUser(ctx context.Context, targetUserID int64, viewerID *int64) (UserItem, error) {
	user, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: nullViewerID(viewerID)})
	if err != nil {
		return UserItem{}, err
	}

	return UserItem{User: user.User, IsFollowing: user.IsFollowing}, nil
}

func (u *Usecase) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (db.User, error) {
	existingUser, err := u.store.GetUser(ctx, db.GetUserParams{ID: userID})
	if err != nil {
		return db.User{}, err
	}

	newAvatar := existingUser.User.AvatarUrl
	uploadedAvatarURL := ""
	if input.Avatar != nil {
		contentType := strings.ToLower(input.Avatar.ContentType)
		if !strings.HasPrefix(contentType, "image/") {
			return db.User{}, apperr.BadRequest("avatar must be an image")
		}

		uploadedAvatarURL, err = u.storage.UploadFile(ctx, input.Avatar.Reader, input.Avatar.Filename, input.Avatar.ContentType)
		if err != nil {
			return db.User{}, err
		}
		newAvatar = sql.NullString{String: uploadedAvatarURL, Valid: true}
	}

	bio := existingUser.User.Bio
	if input.Bio != nil {
		bio = nullStringFromPtr(input.Bio)
	}

	displayName := existingUser.User.DisplayName
	if input.DisplayName != nil {
		displayName = nullStringFromPtr(input.DisplayName)
	}

	updatedUser, err := u.store.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:          userID,
		Bio:         bio,
		DisplayName: displayName,
		AvatarUrl:   newAvatar,
	})
	if err != nil {
		if uploadedAvatarURL != "" {
			_ = u.storage.DeleteFile(ctx, uploadedAvatarURL)
		}
		return db.User{}, err
	}

	if uploadedAvatarURL != "" && existingUser.User.AvatarUrl.Valid {
		_ = u.storage.DeleteFile(ctx, existingUser.User.AvatarUrl.String)
	}

	return updatedUser, nil
}

func (u *Usecase) FollowUser(ctx context.Context, followerID, targetUserID int64) (bool, error) {
	targetUser, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: sql.NullInt64{Valid: false}})
	if err != nil {
		return false, err
	}

	var inserted bool
	var pendingNotification db.Notification
	err = u.store.ExecTxAfterCommit(ctx, func(q *db.Queries) error {
		var err error
		inserted, err = q.FollowUser(ctx, db.FollowUserParams{FollowerID: followerID, FollowingID: targetUserID})
		if err != nil {
			return err
		}

		if inserted {
			pendingNotification, _ = u.createNotification(ctx, q, targetUser.User.ID, followerID, nil, NotifTypeFollow)
		}
		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			u.dispatchNotification(pendingNotification)
		}
	})
	if err != nil {
		return false, err
	}

	return inserted, nil
}

func (u *Usecase) UnfollowUser(ctx context.Context, followerID, targetUserID int64) error {
	if _, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: sql.NullInt64{Valid: false}}); err != nil {
		return err
	}

	_, err := u.store.UnfollowUser(ctx, db.UnfollowUserParams{FollowerID: followerID, FollowingID: targetUserID})
	return err
}

func (u *Usecase) ListFollowers(ctx context.Context, targetUserID int64, page, size int32, viewerID *int64) ([]UserItem, error) {
	users, err := u.store.ListFollowersUsers(ctx, db.ListFollowersUsersParams{
		FollowingID: targetUserID,
		Limit:       size,
		Offset:      page * size,
		ViewerID:    nullViewerID(viewerID),
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(users))
	for _, r := range users {
		items = append(items, UserItem{User: r.User, IsFollowing: r.IsFollowing})
	}
	return items, nil
}

func (u *Usecase) ListFollowing(ctx context.Context, targetUserID int64, page, size int32, viewerID *int64) ([]UserItem, error) {
	users, err := u.store.ListFollowingUsers(ctx, db.ListFollowingUsersParams{
		FollowerID: targetUserID,
		Limit:      size,
		Offset:     page * size,
		ViewerID:   nullViewerID(viewerID),
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(users))
	for _, r := range users {
		items = append(items, UserItem{User: r.User, IsFollowing: r.IsFollowing})
	}
	return items, nil
}

func (u *Usecase) CountFollowers(ctx context.Context, targetUserID int64) (int64, error) {
	return u.store.CountFollowersUsers(ctx, targetUserID)
}

func (u *Usecase) CountFollowing(ctx context.Context, targetUserID int64) (int64, error) {
	return u.store.CountFollowingUsers(ctx, targetUserID)
}
