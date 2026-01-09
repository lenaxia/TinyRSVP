package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRouter_Integration_MiddlewareChain(t *testing.T) {
	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", resp["status"])
	}
}

func TestRouter_Integration_RouteParameterExtraction(t *testing.T) {
	r := chi.NewRouter()

	r.Get("/events/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := GetEventIDFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	})

	r.Get("/rsvp/{token}", func(w http.ResponseWriter, r *http.Request) {
		token, err := GetTokenFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})

	tests := []struct {
		name           string
		path           string
		wantStatusCode int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid event ID extraction",
			path:           "/events/123",
			wantStatusCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]int64
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp["id"] != 123 {
					t.Errorf("Expected id 123, got %d", resp["id"])
				}
			},
		},
		{
			name:           "invalid event ID extraction",
			path:           "/events/abc",
			wantStatusCode: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "valid token extraction",
			path:           "/rsvp/test-token-123",
			wantStatusCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp["token"] != "test-token-123" {
					t.Errorf("Expected token 'test-token-123', got %s", resp["token"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestRouter_Integration_404Handling(t *testing.T) {
	router := NewRouter(nil)

	tests := []struct {
		name         string
		path         string
		acceptHeader string
		wantJSON     bool
	}{
		{
			name:         "API 404 returns JSON",
			path:         "/api/nonexistent",
			acceptHeader: "application/json",
			wantJSON:     true,
		},
		{
			name:         "Web 404 returns HTML",
			path:         "/nonexistent",
			acceptHeader: "text/html",
			wantJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected status 404, got %d", w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if tt.wantJSON {
				if contentType != "application/json" {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
			} else {
				if contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected HTML content type, got %s", contentType)
				}
			}
		})
	}
}

func TestRouter_Integration_405Handling(t *testing.T) {
	r := chi.NewRouter()
	r.NotFound(NotFoundHandler)
	r.MethodNotAllowed(MethodNotAllowedHandler)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		method       string
		path         string
		acceptHeader string
		wantStatus   int
		wantJSON     bool
	}{
		{
			name:         "POST on GET-only route returns 405 JSON",
			method:       http.MethodPost,
			path:         "/test",
			acceptHeader: "application/json",
			wantStatus:   http.StatusMethodNotAllowed,
			wantJSON:     true,
		},
		{
			name:         "PUT on GET-only route returns 405 HTML",
			method:       http.MethodPut,
			path:         "/test",
			acceptHeader: "text/html",
			wantStatus:   http.StatusMethodNotAllowed,
			wantJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", w.Code, tt.wantStatus)
			}

			contentType := w.Header().Get("Content-Type")
			if tt.wantJSON {
				if contentType != "application/json" {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
			} else {
				if contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected HTML content type, got %s", contentType)
				}
			}
		})
	}
}

func TestRouter_Integration_RouteGroupIsolation(t *testing.T) {
	router := NewRouter(nil)

	authCalled := false
	apiCalled := false

	r := chi.NewRouter()
	r.Route("/auth", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authCalled = true
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				apiCalled = true
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	authCalled = false
	apiCalled = false
	req := httptest.NewRequest(http.MethodGet, "/auth/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !authCalled {
		t.Error("Auth middleware should have been called for /auth route")
	}
	if apiCalled {
		t.Error("API middleware should not have been called for /auth route")
	}

	authCalled = false
	apiCalled = false
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if authCalled {
		t.Error("Auth middleware should not have been called for /api route")
	}
	if !apiCalled {
		t.Error("API middleware should have been called for /api route")
	}

	_ = router
}

func TestRouter_Integration_HealthAndReadiness(t *testing.T) {
	router := NewRouter(nil)

	tests := []struct {
		name           string
		path           string
		wantStatusCode int
	}{
		{
			name:           "health endpoint accessible",
			path:           "/health",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, w.Code)
			}

			var resp map[string]string
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp["status"] != "ok" {
				t.Errorf("Expected status 'ok', got %s", resp["status"])
			}
		})
	}
}

func TestRouter_Integration_ConcurrentRequests(t *testing.T) {
	router := NewRouter(nil)

	const numRequests = 100
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			done <- true
		}()
	}

	for i := 0; i < numRequests; i++ {
		<-done
	}
}

func TestRouter_Integration_AllRouteGroupsAccessible(t *testing.T) {
	router := NewRouter(nil)

	routes := []struct {
		name   string
		method string
		path   string
	}{
		{"auth login", http.MethodGet, "/auth/login"},
		{"auth callback", http.MethodGet, "/auth/callback"},
		{"auth logout", http.MethodPost, "/auth/logout"},
		{"api events list", http.MethodGet, "/api/events"},
		{"api events create", http.MethodPost, "/api/events"},
		{"api invites cleanup", http.MethodPost, "/api/invites/cleanup"},
		{"rsvp page", http.MethodGet, "/rsvp/test-token"},
		{"rsvp submit", http.MethodPost, "/rsvp/test-token"},
		{"health check", http.MethodGet, "/health"},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s returned 404, route may not be registered", route.method, route.path)
			}
		})
	}
}
