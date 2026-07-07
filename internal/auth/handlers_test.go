package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestLoginHandler_Success(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://provider.com/authorize?state=abc", http.StatusFound)
			return nil
		},
	}

	handler := NewLoginHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}
}

func TestLoginHandler_Error(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			return fmt.Errorf("provider unavailable")
		},
	}

	handler := NewLoginHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCallbackHandler_Success(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{
				Email:       "user@example.com",
				Name:        "Test User",
				OIDCSubject: &subject,
			}, nil
		},
	}

	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{
				ID:          1,
				Email:       email,
				Name:        name,
				Role:        models.RoleAdmin,
				OIDCSubject: oidcSubject,
			}, nil
		},
	}

	mockSessionMgr := &MockSessionManager{
		CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{
				ID:     "session-123",
				UserID: userID,
			}, nil
		},
		SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			http.SetCookie(w, &http.Cookie{
				Name:  SessionCookieName,
				Value: sessionID,
			})
			return nil
		},
	}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Expected session cookie, got none")
	}

	if sessionCookie.Value != "session-123" {
		t.Errorf("Expected session ID session-123, got %s", sessionCookie.Value)
	}
}

func TestCallbackHandler_AuthError(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return nil, fmt.Errorf("state mismatch")
		},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCallbackHandler_UserCreationError(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{
				Email:       "user@example.com",
				Name:        "Test User",
				OIDCSubject: &subject,
			}, nil
		},
	}

	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	mockSessionMgr := &MockSessionManager{}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLogoutHandler_Success(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	handler := NewLogoutHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/logout", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}
}

func TestLogoutHandler_Error(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLogoutFunc: func(w http.ResponseWriter, r *http.Request) error {
			return fmt.Errorf("session deletion failed")
		},
	}

	handler := NewLogoutHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/logout", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLogoutHandler_MethodNotAllowed(t *testing.T) {
	mockAuth := &MockAuthenticator{}
	handler := NewLogoutHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/logout", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

type MockAuthenticator struct {
	HandleLoginFunc    func(w http.ResponseWriter, r *http.Request) error
	HandleCallbackFunc func(w http.ResponseWriter, r *http.Request) (*AuthResult, error)
	HandleLogoutFunc   func(w http.ResponseWriter, r *http.Request) error
}

func (m *MockAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	if m.HandleLoginFunc != nil {
		return m.HandleLoginFunc(w, r)
	}
	return nil
}

func (m *MockAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	if m.HandleCallbackFunc != nil {
		return m.HandleCallbackFunc(w, r)
	}
	return &AuthResult{Email: "test@example.com", Name: "Test User"}, nil
}

func (m *MockAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	if m.HandleLogoutFunc != nil {
		return m.HandleLogoutFunc(w, r)
	}
	return nil
}

func TestLoginHandler_OpenRedirectPrevention(t *testing.T) {
	tests := []struct {
		name         string
		returnParam  string
		wantLocation string
	}{
		{"no param defaults to root", "", "/"},
		{"relative path allowed", "/events", "/events"},
		{"absolute URL blocked", "http://evil.com", "/"},
		{"protocol relative blocked", "//evil.com", "/"},
		{"javascript blocked", "javascript:alert(1)", "/"},
		{"backslash blocked", "\\evil.com", "/"},
		{"mixed slashes blocked", "/\\evil.com", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &MockAuthenticator{
				HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
					return nil
				},
			}

			handler := NewLoginHandler(mockAuth)

			w := httptest.NewRecorder()
			url := "/login"
			if tt.returnParam != "" {
				url += "?return=" + tt.returnParam
			}
			r := httptest.NewRequest("GET", url, nil)

			handler.ServeHTTP(w, r)

			if w.Code != http.StatusFound {
				t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
			}

			location := w.Header().Get("Location")
			if location != tt.wantLocation {
				t.Errorf("Expected redirect to %s, got %s", tt.wantLocation, location)
			}
		})
	}
}

func TestCallbackHandler_OpenRedirectPrevention(t *testing.T) {
	tests := []struct {
		name         string
		returnParam  string
		wantLocation string
	}{
		{"no param defaults to root", "", "/"},
		{"relative path allowed", "/events", "/events"},
		{"absolute URL blocked", "http://evil.com", "/"},
		{"protocol relative blocked", "//evil.com", "/"},
		{"javascript blocked", "javascript:alert(1)", "/"},
		{"backslash blocked", "\\evil.com", "/"},
		{"mixed slashes blocked", "/\\evil.com", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := "user123"
			mockAuth := &MockAuthenticator{
				HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
					return &AuthResult{
						Email:       "user@example.com",
						Name:        "Test User",
						OIDCSubject: &subject,
					}, nil
				},
			}

			mockUserService := &MockUserService{
				GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
					return &models.User{
						ID:          1,
						Email:       email,
						Name:        name,
						Role:        models.RoleAdmin,
						OIDCSubject: oidcSubject,
					}, nil
				},
			}

			mockSessionMgr := &MockSessionManager{
				CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
					return &models.Session{
						ID:     "session-123",
						UserID: userID,
					}, nil
				},
				SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
					http.SetCookie(w, &http.Cookie{
						Name:  SessionCookieName,
						Value: sessionID,
					})
					return nil
				},
			}

			handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

			w := httptest.NewRecorder()
			url := "/auth/callback?code=test&state=abc"
			if tt.returnParam != "" {
				url += "&return=" + tt.returnParam
			}
			r := httptest.NewRequest("GET", url, nil)

			handler.ServeHTTP(w, r)

			if w.Code != http.StatusFound {
				t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
			}

			location := w.Header().Get("Location")
			if location != tt.wantLocation {
				t.Errorf("Expected redirect to %s, got %s", tt.wantLocation, location)
			}
		})
	}
}

// --- Return URL cookie preservation (Epic 10 Story 06) ---

// getCookieValue finds a cookie by name in the recorder's cookies and returns
// its value, or "" if not found.
func getCookieValue(w *httptest.ResponseRecorder, name string) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func TestLoginHandler_StoresReturnURLInCookie(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://provider.example/authorize?state=abc", http.StatusFound)
			return nil
		},
	}

	handler := NewLoginHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login?return=/events/42/edit", nil)

	handler.ServeHTTP(w, r)

	cookieVal := getCookieValue(w, ReturnURLCookieName)
	if cookieVal != "/events/42/edit" {
		t.Errorf("Expected return URL cookie %q, got %q", "/events/42/edit", cookieVal)
	}

	location := w.Header().Get("Location")
	if location != "https://provider.example/authorize?state=abc" {
		t.Errorf("Expected redirect to OIDC provider, got %s", location)
	}
}

func TestLoginHandler_DirectRedirectWhenAuthDoesNotRedirect(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	handler := NewLoginHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login?return=/dashboard", nil)

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/dashboard" {
		t.Errorf("Expected redirect to /dashboard, got %s", location)
	}

	cookieVal := getCookieValue(w, ReturnURLCookieName)
	if cookieVal != "/dashboard" {
		t.Errorf("Expected return URL cookie %q, got %q", "/dashboard", cookieVal)
	}
}

func TestLoginHandler_StoresValidatedReturnURL(t *testing.T) {
	mockAuth := &MockAuthenticator{
		HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			http.Redirect(w, r, "https://provider.example/authorize", http.StatusFound)
			return nil
		},
	}

	handler := NewLoginHandler(mockAuth)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login?return=https://evil.com", nil)

	handler.ServeHTTP(w, r)

	cookieVal := getCookieValue(w, ReturnURLCookieName)
	if cookieVal != "/" {
		t.Errorf("Expected validated return URL cookie \"/\", got %q", cookieVal)
	}
}

func TestCallbackHandler_RetrievesReturnURLFromCookie(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{Email: "user@example.com", Name: "Test User", OIDCSubject: &subject}, nil
		},
	}

	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin, OIDCSubject: oidcSubject}, nil
		},
	}

	mockSessionMgr := &MockSessionManager{
		CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{ID: "session-123", UserID: userID}, nil
		},
		SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: sessionID})
			return nil
		},
	}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)
	r.AddCookie(&http.Cookie{Name: ReturnURLCookieName, Value: "/events/42/edit"})

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/events/42/edit" {
		t.Errorf("Expected redirect to /events/42/edit (from cookie), got %s", location)
	}
}

func TestCallbackHandler_QueryReturnTakesPrecedenceOverCookie(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{Email: "u@example.com", Name: "U", OIDCSubject: &subject}, nil
		},
	}
	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin}, nil
		},
	}
	mockSessionMgr := &MockSessionManager{
		CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{ID: "s1", UserID: userID}, nil
		},
		SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: sessionID})
			return nil
		},
	}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc&return=/from-query", nil)
	r.AddCookie(&http.Cookie{Name: ReturnURLCookieName, Value: "/from-cookie"})

	handler.ServeHTTP(w, r)

	location := w.Header().Get("Location")
	if location != "/from-query" {
		t.Errorf("Expected query param to win (/from-query), got %s", location)
	}
}

func TestCallbackHandler_ClearsReturnURLCookie(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{Email: "u@example.com", Name: "U", OIDCSubject: &subject}, nil
		},
	}
	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin}, nil
		},
	}
	mockSessionMgr := &MockSessionManager{
		CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{ID: "s1", UserID: userID}, nil
		},
		SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: sessionID})
			return nil
		},
	}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)
	r.AddCookie(&http.Cookie{Name: ReturnURLCookieName, Value: "/events/42/edit"})

	handler.ServeHTTP(w, r)

	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == ReturnURLCookieName {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatal("Expected return URL cookie to be set (for deletion), but it was absent")
	}
	if cookie.MaxAge >= 0 {
		t.Errorf("Expected return URL cookie MaxAge < 0 (delete), got %d", cookie.MaxAge)
	}
}

func TestCallbackHandler_FallbackToRootWhenNoReturnURL(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{Email: "u@example.com", Name: "U", OIDCSubject: &subject}, nil
		},
	}
	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin}, nil
		},
	}
	mockSessionMgr := &MockSessionManager{
		CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{ID: "s1", UserID: userID}, nil
		},
		SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: sessionID})
			return nil
		},
	}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)

	handler.ServeHTTP(w, r)

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected fallback redirect to /, got %s", location)
	}
}

func TestCallbackHandler_InvalidReturnURLInCookieFallsBackToRoot(t *testing.T) {
	subject := "user123"
	mockAuth := &MockAuthenticator{
		HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
			return &AuthResult{Email: "u@example.com", Name: "U", OIDCSubject: &subject}, nil
		},
	}
	mockUserService := &MockUserService{
		GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin}, nil
		},
	}
	mockSessionMgr := &MockSessionManager{
		CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
			return &models.Session{ID: "s1", UserID: userID}, nil
		},
		SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
			http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: sessionID})
			return nil
		},
	}

	handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/callback?code=test&state=abc", nil)
	r.AddCookie(&http.Cookie{Name: ReturnURLCookieName, Value: "https://evil.com"})

	handler.ServeHTTP(w, r)

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected open redirect to be blocked and fall back to /, got %s", location)
	}
}
