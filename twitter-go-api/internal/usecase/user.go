package usecase

import (
	"context"
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

func (u *UserUsecase) GetUser(ctx context.Context, targetUserID int64, viewerID *int64) (UserItem, error) {
	user, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: viewerID})
	if err != nil {
		return UserItem{}, err
	}

	return newUserItemFromDB(user.User, user.IsFollowing), nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (UserItem, error) {
	existingUser, err := u.store.GetUser(ctx, db.GetUserParams{ID: userID})
	if err != nil {
		return UserItem{}, err
	}

	newAvatar := existingUser.User.AvatarUrl
	uploadedAvatarURL := ""
	if input.Avatar != nil {
		contentType := strings.ToLower(input.Avatar.ContentType)
		if !strings.HasPrefix(contentType, "image/") {
			return UserItem{}, apperr.BadRequest("avatar must be an image")
		}

		uploadedAvatarURL, err = u.storage.UploadFile(ctx, input.Avatar.Reader, input.Avatar.Filename, input.Avatar.ContentType)
		if err != nil {
			return UserItem{}, err
		}
		newAvatar = &uploadedAvatarURL
	}

	bio := existingUser.User.Bio
	if input.Bio != nil {
		bio = input.Bio
	}

	displayName := existingUser.User.DisplayName
	if input.DisplayName != nil {
		displayName = input.DisplayName
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
		return UserItem{}, err
	}

	if uploadedAvatarURL != "" && existingUser.User.AvatarUrl != nil {
		_ = u.storage.DeleteFile(ctx, *existingUser.User.AvatarUrl)
	}

	return newUserItemFromDB(updatedUser, false), nil
}

func (u *UserUsecase) FollowUser(ctx context.Context, followerID, targetUserID int64) (bool, error) {
	targetUser, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: nil})
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
			pendingNotification, _ = createNotification(ctx, q, targetUser.User.ID, followerID, nil, NotifTypeFollow)
		}
		return nil
	}, func() {
		if pendingNotification.ID != 0 {
			dispatchNotification(u.publishNotification, pendingNotification)
		}
	})
	if err != nil {
		return false, err
	}

	return inserted, nil
}

func (u *UserUsecase) UnfollowUser(ctx context.Context, followerID, targetUserID int64) error {
	if _, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: nil}); err != nil {
		return err
	}

	_, err := u.store.UnfollowUser(ctx, db.UnfollowUserParams{FollowerID: followerID, FollowingID: targetUserID})
	return err
}

func (u *UserUsecase) ListFollowers(ctx context.Context, targetUserID int64, page, size int32, viewerID *int64) ([]UserItem, error) {
	vID := viewerID
	if _, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: vID}); err != nil {
		return nil, err
	}

	users, err := u.store.ListFollowersUsers(ctx, db.ListFollowersUsersParams{
		FollowingID: targetUserID,
		Limit:       size,
		Offset:      page * size,
		ViewerID:    vID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(users))
	for _, r := range users {
		items = append(items, newUserItemFromDB(r.User, r.IsFollowing))
	}
	return items, nil
}

func (u *UserUsecase) ListFollowing(ctx context.Context, targetUserID int64, page, size int32, viewerID *int64) ([]UserItem, error) {
	vID := viewerID
	if _, err := u.store.GetUser(ctx, db.GetUserParams{ID: targetUserID, ViewerID: vID}); err != nil {
		return nil, err
	}

	users, err := u.store.ListFollowingUsers(ctx, db.ListFollowingUsersParams{
		FollowerID: targetUserID,
		Limit:      size,
		Offset:     page * size,
		ViewerID:   vID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(users))
	for _, r := range users {
		items = append(items, newUserItemFromDB(r.User, r.IsFollowing))
	}
	return items, nil
}

func (u *UserUsecase) CountFollowers(ctx context.Context, targetUserID int64) (int64, error) {
	return u.store.CountFollowersUsers(ctx, targetUserID)
}

func (u *UserUsecase) CountFollowing(ctx context.Context, targetUserID int64) (int64, error) {
	return u.store.CountFollowingUsers(ctx, targetUserID)
}
