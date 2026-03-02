package assets

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"path/filepath"
	"strings"
)

type ImageValidator struct {
	maxSize      int64
	maxWidth     int
	maxHeight    int
	allowedTypes map[string]bool
}

type ValidationResult struct {
	ContentType string
	Format      string
	Width       int
	Height      int
	Size        int64
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewImageValidator() *ImageValidator {
	return &ImageValidator{
		maxSize:   5 * 1024 * 1024,
		maxWidth:  4096,
		maxHeight: 4096,
		allowedTypes: map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		},
	}
}

func (v *ImageValidator) Validate(data []byte) (*ValidationResult, error) {
	if int64(len(data)) > v.maxSize {
		return nil, &ValidationError{
			Field:   "file",
			Message: fmt.Sprintf("Image size exceeds %d bytes", v.maxSize),
		}
	}

	contentType := detectContentType(data)
	if !v.allowedTypes[contentType] {
		return nil, &ValidationError{
			Field:   "file",
			Message: "Only JPEG, PNG, GIF, and WebP images are allowed",
		}
	}

	if err := detectMaliciousPatterns(data); err != nil {
		return nil, err
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, &ValidationError{
			Field:   "file",
			Message: "File is not a valid image",
		}
	}

	if config.Width > v.maxWidth || config.Height > v.maxHeight {
		return nil, &ValidationError{
			Field:   "file",
			Message: fmt.Sprintf("Image dimensions exceed %dx%d pixels", v.maxWidth, v.maxHeight),
		}
	}

	return &ValidationResult{
		ContentType: contentType,
		Format:      format,
		Width:       config.Width,
		Height:      config.Height,
		Size:        int64(len(data)),
	}, nil
}

func detectContentType(data []byte) string {
	if len(data) < 12 {
		return ""
	}

	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg"
	}

	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png"
	}

	if bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")) {
		return "image/gif"
	}

	if bytes.HasPrefix(data, []byte("RIFF")) && len(data) >= 12 && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp"
	}

	return ""
}

func detectMaliciousPatterns(data []byte) error {
	dataLower := bytes.ToLower(data)

	maliciousPatterns := []struct {
		pattern []byte
		message string
	}{
		{[]byte("<script"), "File contains script tags"},
		{[]byte("javascript:"), "File contains javascript protocol"},
		{[]byte("data:text/html"), "File contains data URI with HTML"},
		{[]byte("onerror="), "File contains event handler"},
		{[]byte("onload="), "File contains event handler"},
		{[]byte("onclick="), "File contains event handler"},
		{[]byte("onmouseover="), "File contains event handler"},
		{[]byte("<?php"), "File contains PHP code"},
		{[]byte("#!/"), "File contains shell script"},
	}

	for _, mp := range maliciousPatterns {
		if bytes.Contains(dataLower, mp.pattern) {
			return &ValidationError{
				Field:   "file",
				Message: mp.message,
			}
		}
	}

	return nil
}

func stripEXIF(data []byte, format string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	default:
		return data, nil
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generateUniqueFilename(original string) string {
	ext := filepath.Ext(original)
	base := strings.TrimSuffix(filepath.Base(original), ext)

	base = sanitizeFilename(base)

	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	suffix := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%s_%s%s", base, suffix, ext)
}

func sanitizeFilename(name string) string {
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)

	if len(name) > 50 {
		name = name[:50]
	}

	return name
}
