package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/storage"
)

func TestAssetHandler_ServeAsset_Success(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	testData := []byte("test image data")
	provider.PutObject(context.Background(), "images/123/test.png", bytes.NewReader(testData), "image/png")

	req := httptest.NewRequest(http.MethodGet, "/assets/images/123/test.png", nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	if !bytes.Equal(w.Body.Bytes(), testData) {
		t.Errorf("Body = %v, want %v", w.Body.Bytes(), testData)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("Content-Type = %s, want image/png", ct)
	}

	cc := w.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control = %s, want to contain max-age", cc)
	}

	xct := w.Header().Get("X-Content-Type-Options")
	if xct != "nosniff" {
		t.Errorf("X-Content-Type-Options = %s, want nosniff", xct)
	}
}

func TestAssetHandler_ServeAsset_NotFound(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	req := httptest.NewRequest(http.MethodGet, "/assets/images/123/missing.png", nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAssetHandler_ServeAsset_PathTraversal(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "double dot",
			path: "/assets/../../../etc/passwd",
		},
		{
			name: "double dot in middle",
			path: "/assets/images/../../../etc/passwd",
		},
		{
			name: "encoded double dot",
			path: "/assets/images%2F..%2F..%2Fetc%2Fpasswd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeAsset(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestAssetHandler_ServeAsset_EmptyPath(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	req := httptest.NewRequest(http.MethodGet, "/assets/", nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAssetHandler_ServeAsset_MethodNotAllowed(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/assets/images/123/test.png", nil)
			w := httptest.NewRecorder()

			handler.ServeAsset(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestAssetHandler_ServeAsset_HeadRequest(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	testData := []byte("test image data")
	provider.PutObject(context.Background(), "images/123/test.png", bytes.NewReader(testData), "image/png")

	req := httptest.NewRequest(http.MethodHead, "/assets/images/123/test.png", nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Body.Len() > 0 {
		t.Error("HEAD request should not return body")
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("Content-Type = %s, want image/png", ct)
	}

	cc := w.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control = %s, want to contain max-age", cc)
	}
}

func TestAssetHandler_ServeAsset_StorageError(t *testing.T) {
	provider := storage.NewMockProvider()
	provider.GetObjectFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return nil, errors.New("storage failure")
	}

	handler := NewAssetHandler(provider)

	req := httptest.NewRequest(http.MethodGet, "/assets/images/123/test.png", nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantType string
	}{
		{
			name:     "JPEG with .jpg",
			path:     "images/123/photo.jpg",
			wantType: "image/jpeg",
		},
		{
			name:     "JPEG with .jpeg",
			path:     "images/123/photo.jpeg",
			wantType: "image/jpeg",
		},
		{
			name:     "PNG",
			path:     "images/123/logo.png",
			wantType: "image/png",
		},
		{
			name:     "GIF",
			path:     "images/123/animation.gif",
			wantType: "image/gif",
		},
		{
			name:     "WebP",
			path:     "images/123/modern.webp",
			wantType: "image/webp",
		},
		{
			name:     "uppercase extension",
			path:     "images/123/PHOTO.JPG",
			wantType: "image/jpeg",
		},
		{
			name:     "mixed case extension",
			path:     "images/123/Photo.PnG",
			wantType: "image/png",
		},
		{
			name:     "unknown extension",
			path:     "images/123/file.txt",
			wantType: "application/octet-stream",
		},
		{
			name:     "no extension",
			path:     "images/123/file",
			wantType: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectContentType(tt.path)
			if got != tt.wantType {
				t.Errorf("detectContentType(%s) = %s, want %s", tt.path, got, tt.wantType)
			}
		})
	}
}

func TestAssetHandler_ServeAsset_CacheHeaders(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)

	testData := []byte("test image data")
	provider.PutObject(context.Background(), "images/123/test.png", bytes.NewReader(testData), "image/png")

	req := httptest.NewRequest(http.MethodGet, "/assets/images/123/test.png", nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	cc := w.Header().Get("Cache-Control")
	if !strings.Contains(cc, "public") {
		t.Errorf("Cache-Control should contain 'public', got: %s", cc)
	}
	if !strings.Contains(cc, "max-age=86400") {
		t.Errorf("Cache-Control should contain 'max-age=86400', got: %s", cc)
	}

	xct := w.Header().Get("X-Content-Type-Options")
	if xct != "nosniff" {
		t.Errorf("X-Content-Type-Options = %s, want nosniff", xct)
	}
}
