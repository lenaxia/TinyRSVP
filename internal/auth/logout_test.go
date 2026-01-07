package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLogout_Success(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{
		DeleteSessionFunc: func(ctx context.Context, sessionID string) error {
			if sessionID != "test-session-123" {
				return fmt.Errorf("unexpected session ID: %s", sessionID)
			}
			return nil
		},
		GetSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "test-session-123", nil
		},
		ClearSessionCookieFunc: func(w http.ResponseWriter) error {
			http.SetCookie(w, &http.Cookie{
				Name:     "tinyrsvp_session",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			return nil
		},
	}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/logout", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: "test-session-123",
	})

	err = auth.HandleLogout(w, r)
	if err != nil {
		t.Fatalf("HandleLogout failed: %v", err)
	}

	cookies := w.Result().Cookies()
	var clearCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "tinyrsvp_session" {
			clearCookie = c
			break
		}
	}

	if clearCookie == nil {
		t.Fatal("Expected session cookie to be set for clearing")
	}

	if clearCookie.MaxAge != -1 {
		t.Errorf("Expected MaxAge -1 for cookie deletion, got %d", clearCookie.MaxAge)
	}

	if clearCookie.Value != "" {
		t.Errorf("Expected empty cookie value, got %s", clearCookie.Value)
	}
}

func TestHandleLogout_NoSessionCookie(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{
		GetSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "", http.ErrNoCookie
		},
	}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/logout", nil)

	err = auth.HandleLogout(w, r)
	if err != nil {
		t.Errorf("Expected no error for missing session cookie, got %v", err)
	}
}

func TestHandleLogout_DeleteSessionError(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{
		DeleteSessionFunc: func(ctx context.Context, sessionID string) error {
			return fmt.Errorf("database error")
		},
		GetSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "test-session-123", nil
		},
	}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/logout", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: "test-session-123",
	})

	err = auth.HandleLogout(w, r)
	if err == nil {
		t.Fatal("Expected error from DeleteSession, got nil")
	}
}

func TestHandleLogout_CookieClearing(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}

	clearCookieCalled := false
	mockSessionMgr := &MockSessionManager{
		DeleteSessionFunc: func(ctx context.Context, sessionID string) error {
			return nil
		},
		GetSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "test-session-123", nil
		},
		ClearSessionCookieFunc: func(w http.ResponseWriter) error {
			clearCookieCalled = true
			http.SetCookie(w, &http.Cookie{
				Name:     "tinyrsvp_session",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			return nil
		},
	}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/logout", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: "test-session-123",
	})

	err = auth.HandleLogout(w, r)
	if err != nil {
		t.Fatalf("HandleLogout failed: %v", err)
	}

	if !clearCookieCalled {
		t.Error("Expected ClearSessionCookie to be called")
	}
}
