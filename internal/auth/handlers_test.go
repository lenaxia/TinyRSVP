package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/tinyrsvp/internal/models"
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
	if location != "/dashboard" {
		t.Errorf("Expected redirect to /dashboard, got %s", location)
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
