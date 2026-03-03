package server

import (
	"bytes"
	"io"
	"testing"
)

func TestNormalizeContentType(t *testing.T) {
	t.Parallel()

	got := normalizeContentType("Image/JPEG; charset=binary")
	if got != "image/jpeg" {
		t.Fatalf("expected image/jpeg, got %q", got)
	}
}

func TestHasAllowedExtension(t *testing.T) {
	t.Parallel()

	if !hasAllowedExtension("avatar.JPG", avatarAllowedExts) {
		t.Fatal("expected JPG extension to be allowed")
	}
	if hasAllowedExtension("avatar.svg", avatarAllowedExts) {
		t.Fatal("expected SVG extension to be rejected")
	}
}

func TestDetectContentType(t *testing.T) {
	t.Parallel()

	// PNG signature followed by padding bytes.
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	input := append(png, bytes.Repeat([]byte{0x00}, 32)...)

	contentType, replay, err := detectContentType(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("detectContentType returned error: %v", err)
	}
	if contentType != "image/png" {
		t.Fatalf("expected image/png, got %q", contentType)
	}

	replayed, err := io.ReadAll(replay)
	if err != nil {
		t.Fatalf("failed to read replayed content: %v", err)
	}
	if !bytes.Equal(replayed, input) {
		t.Fatal("replayed content does not match original input")
	}
}
