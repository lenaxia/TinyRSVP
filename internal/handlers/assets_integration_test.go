package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/storage"
)

func TestAssetHandler_Integration_ServeRealImage(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	testData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	path := "images/123/photo.jpg"

	err := provider.PutObject(ctx, path, bytes.NewReader(testData), "image/jpeg")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/"+path, nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	if !bytes.Equal(w.Body.Bytes(), testData) {
		t.Errorf("Body mismatch, got %d bytes, want %d bytes", len(w.Body.Bytes()), len(testData))
	}

	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %s, want image/jpeg", ct)
	}

	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age=86400") {
		t.Errorf("Cache-Control = %s, want to contain max-age=86400", cc)
	}
}

func TestAssetHandler_Integration_ServeMultipleFormats(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	tests := []struct {
		name        string
		path        string
		data        []byte
		contentType string
	}{
		{
			name:        "JPEG",
			path:        "images/123/photo.jpg",
			data:        []byte{0xFF, 0xD8, 0xFF},
			contentType: "image/jpeg",
		},
		{
			name:        "PNG",
			path:        "images/123/logo.png",
			data:        []byte{0x89, 0x50, 0x4E, 0x47},
			contentType: "image/png",
		},
		{
			name:        "GIF",
			path:        "images/123/animation.gif",
			data:        []byte{0x47, 0x49, 0x46, 0x38},
			contentType: "image/gif",
		},
		{
			name:        "WebP",
			path:        "images/123/modern.webp",
			data:        []byte{0x52, 0x49, 0x46, 0x46},
			contentType: "image/webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.PutObject(ctx, tt.path, bytes.NewReader(tt.data), tt.contentType)
			if err != nil {
				t.Fatalf("PutObject() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/assets/"+tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeAsset(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}

			if ct := w.Header().Get("Content-Type"); ct != tt.contentType {
				t.Errorf("Content-Type = %s, want %s", ct, tt.contentType)
			}

			if !bytes.Equal(w.Body.Bytes(), tt.data) {
				t.Error("Body data mismatch")
			}
		})
	}
}

func TestAssetHandler_Integration_ConcurrentRequests(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	testData := []byte("concurrent test data")
	path := "images/123/concurrent.jpg"

	err := provider.PutObject(ctx, path, bytes.NewReader(testData), "image/jpeg")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	numRequests := 20
	results := make(chan int, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/assets/"+path, nil)
			w := httptest.NewRecorder()

			handler.ServeAsset(w, req)

			if w.Code != http.StatusOK {
				errors <- http.ErrNotSupported
				return
			}

			if !bytes.Equal(w.Body.Bytes(), testData) {
				errors <- http.ErrNotSupported
				return
			}

			results <- w.Code
		}()
	}

	for i := 0; i < numRequests; i++ {
		select {
		case code := <-results:
			if code != http.StatusOK {
				t.Errorf("Request %d: Status = %d, want %d", i, code, http.StatusOK)
			}
		case err := <-errors:
			t.Errorf("Request %d: error = %v", i, err)
		}
	}
}

func TestAssetHandler_Integration_LargeFile(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	path := "images/123/large.jpg"
	err := provider.PutObject(ctx, path, bytes.NewReader(largeData), "image/jpeg")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/"+path, nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if len(w.Body.Bytes()) != len(largeData) {
		t.Errorf("Body size = %d, want %d", len(w.Body.Bytes()), len(largeData))
	}

	if !bytes.Equal(w.Body.Bytes(), largeData) {
		t.Error("Large file data mismatch")
	}
}

func TestAssetHandler_Integration_PathSecurity(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	sensitiveData := []byte("sensitive data")
	provider.PutObject(ctx, "sensitive/secret.txt", bytes.NewReader(sensitiveData), "text/plain")

	maliciousPaths := []string{
		"/assets/../sensitive/secret.txt",
		"/assets/images/../../sensitive/secret.txt",
		"/assets/images/../../../etc/passwd",
		"/assets/./../../sensitive/secret.txt",
	}

	for _, path := range maliciousPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.ServeAsset(w, req)

			if w.Code == http.StatusOK {
				t.Errorf("Expected error for malicious path, got status %d", w.Code)
			}

			if bytes.Equal(w.Body.Bytes(), sensitiveData) {
				t.Error("Malicious path returned sensitive data")
			}
		})
	}
}

func TestAssetHandler_Integration_MultipleEvents(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	event1Data := []byte("event 1 image")
	event2Data := []byte("event 2 image")

	provider.PutObject(ctx, "images/100/photo.jpg", bytes.NewReader(event1Data), "image/jpeg")
	provider.PutObject(ctx, "images/200/photo.jpg", bytes.NewReader(event2Data), "image/jpeg")

	req1 := httptest.NewRequest(http.MethodGet, "/assets/images/100/photo.jpg", nil)
	w1 := httptest.NewRecorder()
	handler.ServeAsset(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Event 100 status = %d, want %d", w1.Code, http.StatusOK)
	}
	if !bytes.Equal(w1.Body.Bytes(), event1Data) {
		t.Error("Event 100 data mismatch")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/assets/images/200/photo.jpg", nil)
	w2 := httptest.NewRecorder()
	handler.ServeAsset(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Event 200 status = %d, want %d", w2.Code, http.StatusOK)
	}
	if !bytes.Equal(w2.Body.Bytes(), event2Data) {
		t.Error("Event 200 data mismatch")
	}
}

func TestAssetHandler_Integration_HeadRequestNoBody(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	largeData := make([]byte, 1024*100)
	path := "images/123/large.jpg"

	err := provider.PutObject(ctx, path, bytes.NewReader(largeData), "image/jpeg")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodHead, "/assets/"+path, nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Body.Len() > 0 {
		t.Errorf("HEAD request returned body of %d bytes, want 0", w.Body.Len())
	}

	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %s, want image/jpeg", ct)
	}

	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control = %s, want to contain max-age", cc)
	}
}

func TestAssetHandler_Integration_StreamingBehavior(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	testData := []byte("streaming test data")
	path := "images/123/stream.jpg"

	err := provider.PutObject(ctx, path, bytes.NewReader(testData), "image/jpeg")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/"+path, nil)
	w := httptest.NewRecorder()

	handler.ServeAsset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(body, testData) {
		t.Error("Streamed data mismatch")
	}
}

func TestAssetHandler_Integration_CacheHeadersConsistency(t *testing.T) {
	provider := storage.NewMockProvider()
	handler := NewAssetHandler(provider)
	ctx := context.Background()

	testData := []byte("cache test data")
	path := "images/123/cached.jpg"

	err := provider.PutObject(ctx, path, bytes.NewReader(testData), "image/jpeg")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/assets/"+path, nil)
		w := httptest.NewRecorder()

		handler.ServeAsset(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Status = %d, want %d", i, w.Code, http.StatusOK)
		}

		cc := w.Header().Get("Cache-Control")
		if !strings.Contains(cc, "public") || !strings.Contains(cc, "max-age=86400") {
			t.Errorf("Request %d: Cache-Control = %s, want public, max-age=86400", i, cc)
		}

		xct := w.Header().Get("X-Content-Type-Options")
		if xct != "nosniff" {
			t.Errorf("Request %d: X-Content-Type-Options = %s, want nosniff", i, xct)
		}
	}
}
