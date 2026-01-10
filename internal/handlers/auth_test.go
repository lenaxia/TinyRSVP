package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestShowLogin_ValidReturnURL(t *testing.T) {
	tests := []struct {
		name           string
		returnURL      string
		wantStatusCode int
		wantInBody     string
	}{
		{
			name:           "with valid return URL",
			returnURL:      "/dashboard",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2fdashboard",
		},
		{
			name:           "with root return URL",
			returnURL:      "/",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2f",
		},
		{
			name:           "with empty return URL defaults to root",
			returnURL:      "",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2f",
		},
		{
			name:           "with nested path",
			returnURL:      "/events/123",
			wantStatusCode: http.StatusOK,
			wantInBody:     "return=%2fevents%2f123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandlers{}

			reqURL := "/login"
			if tt.returnURL != "" {
				reqURL += "?return=" + url.QueryEscape(tt.returnURL)
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			h.ShowLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("ShowLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.wantInBody) {
				t.Errorf("ShowLogin() body should contain %q, got %q", tt.wantInBody, body)
			}
		})
	}
}

func TestShowLogin_InvalidReturnURL(t *testing.T) {
	tests := []struct {
		name           string
		returnURL      string
		wantStatusCode int
	}{
		{
			name:           "external URL rejected",
			returnURL:      "https://evil.com/phishing",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "protocol relative URL rejected",
			returnURL:      "//evil.com/phishing",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "javascript protocol rejected",
			returnURL:      "javascript:alert(1)",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "data URL rejected",
			returnURL:      "data:text/html,<script>alert(1)</script>",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandlers{}

			reqURL := "/login?return=" + url.QueryEscape(tt.returnURL)
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			h.ShowLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("ShowLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestOIDCLogin_RedirectsToProvider(t *testing.T) {
	tests := []struct {
		name           string
		returnURL      string
		wantStatusCode int
	}{
		{
			name:           "redirects with return URL",
			returnURL:      "/dashboard",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "redirects without return URL",
			returnURL:      "",
			wantStatusCode: http.StatusFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mockAuthenticator{
				handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
					http.Redirect(w, r, "https://oidc.example.com/authorize", http.StatusFound)
					return nil
				},
			}

			h := &AuthHandlers{
				authenticator: mockAuth,
			}

			reqURL := "/auth/oidc/login"
			if tt.returnURL != "" {
				reqURL += "?return=" + url.QueryEscape(tt.returnURL)
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			h.OIDCLogin(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("OIDCLogin() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			location := w.Header().Get("Location")
			if location == "" {
				t.Error("OIDCLogin() should set Location header")
			}
		})
	}
}

func TestOIDCLogin_InvalidReturnURL(t *testing.T) {
	h := &AuthHandlers{
		authenticator: &mockAuthenticator{},
	}

	reqURL := "/auth/oidc/login?return=" + url.QueryEscape("https://evil.com")
	req := httptest.NewRequest(http.MethodGet, reqURL, nil)
	w := httptest.NewRecorder()

	h.OIDCLogin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("OIDCLogin() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestOIDCLogin_AuthenticatorError(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			return http.ErrAbortHandler
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/login", nil)
	w := httptest.NewRecorder()

	h.OIDCLogin(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("OIDCLogin() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestOIDCCallback_Success(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc123&state=xyz", nil)
	w := httptest.NewRecorder()

	h.OIDCCallback(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("OIDCCallback() status = %v, want %v", w.Code, http.StatusFound)
	}

	location := w.Header().Get("Location")
	if location != "/dashboard" {
		t.Errorf("OIDCCallback() location = %v, want /dashboard", location)
	}
}

func TestOIDCCallback_Error(t *testing.T) {
	tests := []struct {
		name           string
		callbackError  error
		wantStatusCode int
	}{
		{
			name:           "authentication error",
			callbackError:  http.ErrAbortHandler,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mockAuthenticator{
				handleCallbackFunc: func(w http.ResponseWriter, r *http.Request) error {
					return tt.callbackError
				},
			}

			h := &AuthHandlers{
				authenticator: mockAuth,
			}

			req := httptest.NewRequest(http.MethodGet, "/auth/oidc/callback?code=abc123&state=xyz", nil)
			w := httptest.NewRecorder()

			h.OIDCCallback(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("OIDCCallback() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestLogout_Success(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

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
}

func TestLogout_MethodNotAllowed(t *testing.T) {
	h := &AuthHandlers{
		authenticator: &mockAuthenticator{},
	}

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Logout() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestLogout_Error(t *testing.T) {
	mockAuth := &mockAuthenticator{
		handleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return http.ErrAbortHandler
		},
	}

	h := &AuthHandlers{
		authenticator: mockAuth,
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Logout() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestValidateReturnURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid root path",
			url:     "/",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			url:     "/dashboard",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			url:     "/events/123/edit",
			wantErr: false,
		},
		{
			name:    "valid path with query",
			url:     "/dashboard?tab=events",
			wantErr: false,
		},
		{
			name:    "empty string defaults to root",
			url:     "",
			wantErr: false,
		},
		{
			name:    "external URL rejected",
			url:     "https://evil.com",
			wantErr: true,
		},
		{
			name:    "protocol relative URL rejected",
			url:     "//evil.com",
			wantErr: true,
		},
		{
			name:    "javascript protocol rejected",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "data URL rejected",
			url:     "data:text/html,<script>",
			wantErr: true,
		},
		{
			name:    "relative path without leading slash rejected",
			url:     "dashboard",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateReturnURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateReturnURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockAuthenticator struct {
	handleLoginFunc    func(w http.ResponseWriter, r *http.Request) error
	handleCallbackFunc func(w http.ResponseWriter, r *http.Request) error
	handleLogoutFunc   func(w http.ResponseWriter, r *http.Request) error
}

func (m *mockAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	if m.handleLoginFunc != nil {
		return m.handleLoginFunc(w, r)
	}
	return nil
}

func (m *mockAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	if m.handleCallbackFunc != nil {
		return nil, m.handleCallbackFunc(w, r)
	}
	return nil, nil
}

func (m *mockAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	if m.handleLogoutFunc != nil {
		return m.handleLogoutFunc(w, r)
	}
	return nil
}
