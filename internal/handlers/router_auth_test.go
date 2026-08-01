package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
)

func TestRouter_WithAuthHandlers(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://oidc.example.com/authorize", http.StatusFound)
			return nil
		},
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "user@example.com"}, nil
		},
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	authHandlers := NewAuthHandlers(mockAuth, &mockAuthUserService{}, &mockAuthSessionManager{})

	router := NewRouter(&RouterHandlers{
		AuthHandlers: authHandlers,
	})

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantLocation   string
	}{
		{
			name:           "GET /login returns login page",
			method:         http.MethodGet,
			path:           "/login",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "GET /auth/oidc/login redirects to OIDC",
			method:         http.MethodGet,
			path:           "/auth/oidc/login",
			wantStatusCode: http.StatusFound,
			wantLocation:   "oidc.example.com",
		},
		{
			name:           "GET /auth/oidc/callback redirects to dashboard",
			method:         http.MethodGet,
			path:           "/auth/oidc/callback?code=test&state=test&return=/dashboard",
			wantStatusCode: http.StatusFound,
			wantLocation:   "/dashboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.wantLocation != "" {
				location := w.Header().Get("Location")
				if !strings.Contains(location, tt.wantLocation) {
					t.Errorf("Location = %v, should contain %v", location, tt.wantLocation)
				}
			}
		})
	}
}

func TestRouter_WithAuthHandlers_ListRoutes(t *testing.T) {
	mockAuth := &mockAuthenticator{}
	authHandlers := NewAuthHandlers(mockAuth, &mockAuthUserService{}, &mockAuthSessionManager{})

	router := NewRouter(&RouterHandlers{
		AuthHandlers: authHandlers,
	})

	routes := router.ListRoutes()

	expectedRoutes := []string{
		"/login",
		"/auth/oidc/login",
		"/auth/oidc/callback",
		"/logout",
	}

	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Pattern == expectedRoute {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Available routes:")
			for _, route := range routes {
				t.Logf("  %s %s", route.Method, route.Pattern)
			}
			t.Errorf("ListRoutes() missing expected route: %s", expectedRoute)
		}
	}
}

func TestRouter_WithAuthHandlers_ReturnURLFlow(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://oidc.example.com/authorize", http.StatusFound)
			return nil
		},
	}

	authHandlers := NewAuthHandlers(mockAuth, &mockAuthUserService{}, &mockAuthSessionManager{})
	router := NewRouter(&RouterHandlers{
		AuthHandlers: authHandlers,
	})

	req := httptest.NewRequest(http.MethodGet, "/login?return=/events/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "return=%2fevents%2f123") {
		t.Error("Login page should preserve return URL in OIDC login link")
	}
}

func TestRouter_WithAuthHandlers_InvalidReturnURL(t *testing.T) {
	mockAuth := &mockAuthenticator{}
	authHandlers := NewAuthHandlers(mockAuth, &mockAuthUserService{}, &mockAuthSessionManager{})

	router := NewRouter(&RouterHandlers{
		AuthHandlers: authHandlers,
	})

	req := httptest.NewRequest(http.MethodGet, "/login?return=https://evil.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestRouter_WithAuthHandlers_LogoutMethodRestriction(t *testing.T) {
	mockAuth := &mockAuthenticator{}
	authHandlers := NewAuthHandlers(mockAuth, &mockAuthUserService{}, &mockAuthSessionManager{})

	router := NewRouter(&RouterHandlers{
		AuthHandlers: authHandlers,
	})

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %v, want %v (GET should not be allowed on /logout)", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestRouter_WithAuthHandlers_LogoutWithCSRF(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}
	authHandlers := NewAuthHandlers(mockAuth, &mockAuthUserService{}, &mockAuthSessionManager{})

	router := NewRouter(&RouterHandlers{
		AuthHandlers: authHandlers,
	})

	req1 := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	cookies := rec1.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == middleware.CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("Expected CSRF cookie from GET request")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req2.Header.Set(middleware.CSRFHeaderName, csrfCookie.Value)
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusFound {
		t.Errorf("Status = %v, want %v", rec2.Code, http.StatusFound)
	}

	location := rec2.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Location = %v, want /login", location)
	}
}
