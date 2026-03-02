package js

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestThemePreviewModal_CustomImageURLParameter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customImageURL := r.URL.Query().Get("custom_image_url")
		if customImageURL == "" {
			t.Error("Expected custom_image_url parameter to be present")
		}
		if !strings.Contains(customImageURL, "custom.jpg") {
			t.Errorf("Expected custom_image_url to contain 'custom.jpg', got: %s", customImageURL)
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, server.URL+"/api/themes/preview?theme_id=1&custom_image_url=https://example.com/custom.jpg", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestThemePreviewModal_NoCustomImageURL(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customImageURL := r.URL.Query().Get("custom_image_url")
		if customImageURL != "" {
			t.Errorf("Expected no custom_image_url parameter, got: %s", customImageURL)
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, server.URL+"/api/themes/preview?theme_id=1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestThemePreviewModal_CustomImageURLWithEventData(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customImageURL := r.URL.Query().Get("custom_image_url")
		title := r.URL.Query().Get("title")
		location := r.URL.Query().Get("location")

		if customImageURL == "" {
			t.Error("Expected custom_image_url parameter")
		}
		if title == "" {
			t.Error("Expected title parameter")
		}
		if location == "" {
			t.Error("Expected location parameter")
		}

		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	req := httptest.NewRequest(http.MethodGet,
		server.URL+"/api/themes/preview?theme_id=1&custom_image_url=https://example.com/image.jpg&title=Test&location=Here",
		nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
