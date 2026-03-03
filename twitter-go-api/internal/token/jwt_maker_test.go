package token

import (
	"strings"
	"testing"
	"time"
)

func TestNewJWTMakerRejectsShortKey(t *testing.T) {
	t.Parallel()

	if _, err := NewJWTMaker("short-key"); err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestJWTMakerCreateAndVerifyToken(t *testing.T) {
	t.Parallel()

	secret := strings.Repeat("a", 32)
	maker, err := NewJWTMaker(secret)
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}

	token, err := maker.CreateToken(42, time.Minute)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	payload, err := maker.VerifyToken(token)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	if payload.UserID != 42 {
		t.Fatalf("unexpected user id: got %d", payload.UserID)
	}
}
