package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_Returns200(t *testing.T) {
	handler := NewHealthHandler("0.1.0")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	handler := NewHealthHandler("0.1.0")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("Expected status healthy, got %s", response.Status)
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if time.Since(response.Timestamp) > 5*time.Second {
		t.Error("Timestamp should be recent")
	}

	if response.Checks == nil {
		t.Error("Expected checks map to be initialized")
	}
}

func TestHealthHandler_IncludesVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"version 0.1.0", "0.1.0"},
		{"version 1.0.0", "1.0.0"},
		{"version dev", "dev"},
		{"empty version", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler(tt.version)

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			var response HealthResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Version != tt.version {
				t.Errorf("Expected version %s, got %s", tt.version, response.Version)
			}
		})
	}
}

func TestHealthHandler_MultipleRequests(t *testing.T) {
	handler := NewHealthHandler("0.1.0")

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i, w.Code)
		}

		var response HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Request %d: Failed to decode response: %v", i, err)
		}

		if response.Status != StatusHealthy {
			t.Errorf("Request %d: Expected status healthy, got %s", i, response.Status)
		}
	}
}

func TestHealthHandler_InvalidMethod(t *testing.T) {
	handler := NewHealthHandler("0.1.0")

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s (handler doesn't restrict methods), got %d", method, w.Code)
			}
		})
	}
}
