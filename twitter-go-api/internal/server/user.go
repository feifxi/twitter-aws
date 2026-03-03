package server

import (
	"github.com/chanombude/twitter-go-api/internal/usecase"
)



func newUserResponse(user usecase.UserItem) userResponse {
	var displayName, bio, avatarUrl *string
	if user.DisplayName.Valid {
		displayName = &user.DisplayName.String
	}
	if user.Bio.Valid {
		bio = &user.Bio.String
	}
	if user.AvatarUrl.Valid {
		avatarUrl = &user.AvatarUrl.String
	}

	return userResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		DisplayName:    displayName,
		Bio:            bio,
		AvatarUrl:      avatarUrl,
		IsFollowing:    user.IsFollowing,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
	}
}
