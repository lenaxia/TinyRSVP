package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
)

func TestAuthFlow_Integration_LoginToCallback(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://oidc.example.com/authorize?state=test123", http.StatusFound)
			return nil
		},
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
			return &auth.AuthResult{Email: "user@example.com", Name: "Test User", OIDCSubject: strPtr("sub-1")}, nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   &mockAuthUserService{},
		sessionMgr:    &mockAuthSessionManager{},
	}

	t.Run("complete login flow with return URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/oidc/login?return=/events/123", nil)
		w := httptest.NewRecorder()

		h.OIDCLogin(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("OIDCLogin() status = %v, want %v", w.Code, http.StatusFound)
		}

		location := w.Header().Get("Location")
		if !strings.Contains(location, "oidc.example.com") {
			t.Errorf("OIDCLogin() should redirect to OIDC provider, got %v", location)
		}

		callbackReq := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc123&state=test123", nil)
		callbackW := httptest.NewRecorder()

		h.OIDCCallback(callbackW, callbackReq)

		if callbackW.Code != http.StatusFound {
			t.Errorf("OIDCCallback() status = %v, want %v", callbackW.Code, http.StatusFound)
		}

		callbackLocation := callbackW.Header().Get("Location")
		if callbackLocation != "/" {
			t.Errorf("OIDCCallback() location = %v, want / (default when no return URL)", callbackLocation)
		}
	})
}

func TestAuthFlow_Integration_LoginPageToOIDC(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://oidc.example.com/authorize", http.StatusFound)
			return nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

	t.Run("login page displays with return URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login?return=/dashboard", nil)
		w := httptest.NewRecorder()

		h.ShowLogin(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ShowLogin() status = %v, want %v", w.Code, http.StatusOK)
		}

		body := w.Body.String()
		if !strings.Contains(body, "return=%2fdashboard") {
			t.Error("ShowLogin() should contain return URL in OIDC login link")
		}

		if !strings.Contains(body, "/auth/oidc/login") {
			t.Error("ShowLogin() should contain OIDC login link")
		}
	})
}

func TestAuthFlow_Integration_LogoutFlow(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
		userService:   &mockAuthUserService{},
		sessionMgr:    &mockAuthSessionManager{},
	}

	t.Run("logout redirects to login page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()

		h.Logout(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("Logout() status = %v, want %v", w.Code, http.StatusFound)
		}

		location := w.Header().Get("Location")
		if location != "/login" {
			t.Errorf("Logout() location = %v, want /login", location)
		}
	})
}

func TestAuthFlow_Integration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		setupAuth      func() *mockAuthenticator
		handler        func(*AuthHandlers, http.ResponseWriter, *http.Request)
		requestURL     string
		method         string
		wantStatusCode int
		wantInBody     string
	}{
		{
			name: "login with invalid return URL shows error",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{}
			},
			handler: func(h *AuthHandlers, w http.ResponseWriter, r *http.Request) {
				h.ShowLogin(w, r)
			},
			requestURL:     "/login?return=https://evil.com",
			method:         http.MethodGet,
			wantStatusCode: http.StatusBadRequest,
			wantInBody:     "Invalid return URL",
		},
		{
			name: "OIDC login with invalid return URL shows error",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{}
			},
			handler: func(h *AuthHandlers, w http.ResponseWriter, r *http.Request) {
				h.OIDCLogin(w, r)
			},
			requestURL:     "/auth/oidc/login?return=javascript:alert(1)",
			method:         http.MethodGet,
			wantStatusCode: http.StatusBadRequest,
			wantInBody:     "Invalid return URL",
		},
		{
			name: "callback with auth error shows unauthorized",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{
					handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error) {
						return nil, http.ErrAbortHandler
					},
				}
			},
			handler: func(h *AuthHandlers, w http.ResponseWriter, r *http.Request) {
				h.OIDCCallback(w, r)
			},
			requestURL:     "/auth/oidc/callback?code=invalid",
			method:         http.MethodGet,
			wantStatusCode: http.StatusUnauthorized,
			wantInBody:     "Authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := tt.setupAuth()
			h := &AuthHandlers{
				authenticator: mockAuth,
				userService:   &mockAuthUserService{},
				sessionMgr:    &mockAuthSessionManager{},
			}

			req := httptest.NewRequest(tt.method, tt.requestURL, nil)
			req.Header.Set("Accept", "text/html")
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			tt.handler(h, w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantInBody) {
				t.Errorf("Handler body should contain %q, got %q", tt.wantInBody, body)
			}
		})
	}
}

func TestAuthFlow_Integration_ContentNegotiation(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
		returnURL    string
		wantJSON     bool
	}{
		{
			name:         "invalid return URL with JSON accept returns JSON",
			acceptHeader: "application/json",
			returnURL:    "https://evil.com",
			wantJSON:     true,
		},
		{
			name:         "invalid return URL with HTML accept returns HTML",
			acceptHeader: "text/html",
			returnURL:    "https://evil.com",
			wantJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandlers{
				authenticator: &mockAuthenticator{},
			}

			reqURL := "/login?return=" + url.QueryEscape(tt.returnURL)
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			req.Header.Set("Accept", tt.acceptHeader)
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			h.ShowLogin(w, req)

			contentType := w.Header().Get("Content-Type")
			isJSON := strings.Contains(contentType, "application/json")

			if isJSON != tt.wantJSON {
				t.Errorf("Got JSON response = %v, want %v (Content-Type: %s)", isJSON, tt.wantJSON, contentType)
			}
		})
	}
}

func TestAuthFlow_Integration_RequestIDPropagation(t *testing.T) {
	requestID := "test-request-id-12345"

	h := &AuthHandlers{
		authenticator: &mockAuthenticator{},
	}

	req := httptest.NewRequest(http.MethodGet, "/login?return=https://evil.com", nil)
	req.Header.Set("Accept", "text/html")
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, requestID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	h.ShowLogin(w, req)

	body := w.Body.String()
	if !strings.Contains(body, requestID) {
		t.Errorf("Error page should contain request ID %q", requestID)
	}
}
