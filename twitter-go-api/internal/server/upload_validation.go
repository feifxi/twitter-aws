package server

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const sniffLen = 512

var avatarAllowedMIMEs = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/gif":  {},
	"image/webp": {},
}

var avatarAllowedExts = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
}

var mediaAllowedMIMEs = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/gif":       {},
	"image/webp":      {},
	"video/mp4":       {},
	"video/webm":      {},
	"video/quicktime": {},
}

var mediaAllowedExts = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
	".mp4":  {},
	".webm": {},
	".mov":  {},
}

func openAndDetectUpload(fileHeader *multipart.FileHeader) (multipart.File, io.Reader, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, "", err
	}

	contentType, replay, err := detectContentType(file)
	if err != nil {
		_ = file.Close()
		return nil, nil, "", err
	}

	return file, replay, contentType, nil
}

func detectContentType(r io.Reader) (string, io.Reader, error) {
	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", nil, err
	}
	if n == 0 {
		return "", nil, io.EOF
	}

	head := buf[:n]
	contentType := normalizeContentType(http.DetectContentType(head))
	return contentType, io.MultiReader(bytes.NewReader(head), r), nil
}

func normalizeContentType(raw string) string {
	return strings.ToLower(strings.TrimSpace(strings.SplitN(raw, ";", 2)[0]))
}

func isAllowedType(contentType string, allowed map[string]struct{}) bool {
	_, ok := allowed[normalizeContentType(contentType)]
	return ok
}

func hasAllowedExtension(filename string, allowed map[string]struct{}) bool {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))
	_, ok := allowed[ext]
	return ok
}
