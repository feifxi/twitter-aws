package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chanombude/twitter-go-api/internal/apperr"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type AuthResult struct {
	User         UserItem
	AccessToken  string
	RefreshToken string
}

func (u *Usecase) LoginWithGoogle(ctx context.Context, idToken string) (AuthResult, error) {
	payload, err := idtoken.Validate(context.Background(), idToken, u.config.GoogleClientID)
	if err != nil {
		return AuthResult{}, apperr.Unauthorized("invalid google token")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || strings.TrimSpace(email) == "" {
		return AuthResult{}, apperr.Unauthorized("invalid google token")
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	user, err := u.store.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return AuthResult{}, err
		}

		username := buildUniqueUsername(email)
		user, err = u.store.CreateUser(ctx, db.CreateUserParams{
			Username:    username,
			Email:       email,
			DisplayName: nullStringFromPtr(&name),
			AvatarUrl:   nullStringFromPtr(&picture),
			Role:        "USER",
			Provider:    "GOOGLE",
		})
		if err != nil {
			return AuthResult{}, err
		}
	}

	accessToken, refreshToken, err := u.issueSession(ctx, user.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: UserItem{User: user}, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (u *Usecase) RefreshSession(ctx context.Context, refreshToken string) (AuthResult, error) {
	session, err := u.store.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return AuthResult{}, apperr.Unauthorized("refresh token not found or revoked")
	}

	if time.Now().After(session.ExpiryDate) {
		_ = u.store.DeleteRefreshToken(ctx, refreshToken)
		return AuthResult{}, apperr.Unauthorized("refresh token expired")
	}

	accessToken, newRefreshToken, err := u.issueSession(ctx, session.UserID)
	if err != nil {
		return AuthResult{}, err
	}

	user, err := u.store.GetUser(ctx, db.GetUserParams{ID: session.UserID, ViewerID: sql.NullInt64{Valid: false}})
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: UserItem{
			User:        user.User,
			IsFollowing: user.IsFollowing,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *Usecase) Logout(ctx context.Context, userID *int64, refreshToken *string) {
	if userID != nil {
		_ = u.store.DeleteRefreshTokensByUser(ctx, *userID)
		return
	}
	if refreshToken != nil && strings.TrimSpace(*refreshToken) != "" {
		_ = u.store.DeleteRefreshToken(ctx, *refreshToken)
	}
}

func (u *Usecase) GetMe(ctx context.Context, userID int64) (UserItem, error) {
	user, err := u.store.GetUser(ctx, db.GetUserParams{ID: userID, ViewerID: sql.NullInt64{Valid: false}})
	if err != nil {
		return UserItem{}, err
	}
	return UserItem{
		User:        user.User,
		IsFollowing: user.IsFollowing,
	}, nil
}

func (u *Usecase) issueSession(ctx context.Context, userID int64) (accessToken string, refreshToken string, err error) {
	accessToken, err = u.tokenMaker.CreateToken(userID, time.Duration(u.config.TokenDurationMinutes)*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken = uuid.NewString()
	expiresAt := time.Now().Add(time.Duration(u.config.RefreshTokenDurationDays) * 24 * time.Hour)

	if err := u.store.DeleteRefreshTokensByUser(ctx, userID); err != nil {
		return "", "", err
	}

	_, err = u.store.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:     userID,
		Token:      refreshToken,
		ExpiryDate: expiresAt,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func buildUniqueUsername(email string) string {
	base := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")[0]
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, base)
	if clean == "" {
		clean = "user"
	}
	return fmt.Sprintf("%s_%s", clean, strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
}
