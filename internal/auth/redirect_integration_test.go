package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

// TestLoginHandler_OpenRedirectAttacks tests if malicious payloads actually cause redirects
func TestLoginHandler_OpenRedirectAttacks(t *testing.T) {
	attackPayloads := []struct {
		name         string
		returnParam  string
		expectSafe   bool
		shouldReject bool
	}{
		{
			name:         "query parameter injection",
			returnParam:  "/events?return=http://evil.com",
			expectSafe:   false,
			shouldReject: true,
		},
		{
			name:         "fragment injection",
			returnParam:  "/events#http://evil.com",
			expectSafe:   false,
			shouldReject: true,
		},
		{
			name:         "semicolon parameter",
			returnParam:  "/;url=http://evil.com",
			expectSafe:   false,
			shouldReject: true,
		},
		{
			name:         "uppercase HTTP",
			returnParam:  "/HTTP://evil.com",
			expectSafe:   false,
			shouldReject: true,
		},
		{
			name:         "mixed case http",
			returnParam:  "/hTTp://evil.com",
			expectSafe:   false,
			shouldReject: true,
		},
		{
			name:         "CRLF encoded",
			returnParam:  "/%0d%0aLocation:%20http://evil.com",
			expectSafe:   false,
			shouldReject: true,
		},
	}

	for _, tt := range attackPayloads {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &MockAuthenticator{
				HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
					return nil
				},
			}

			handler := NewLoginHandler(mockAuth)

			url := "/login?return=" + tt.returnParam
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			location := rec.Header().Get("Location")

			// Check if the location header contains dangerous patterns
			lowerLocation := strings.ToLower(location)
			isDangerous := strings.Contains(lowerLocation, "http:") ||
				strings.Contains(lowerLocation, "https:") ||
				strings.Contains(lowerLocation, "//evil")

			if isDangerous && tt.shouldReject {
				t.Errorf("SECURITY VULNERABILITY: Login handler redirects to dangerous URL.\nPayload: %q\nRedirect Location: %q",
					tt.returnParam, location)
			}

			if tt.shouldReject && location != "/" {
				t.Errorf("Malicious payload should redirect to /, but redirected to %q", location)
			}
		})
	}
}

// TestCallbackHandler_OpenRedirectAttacks tests callback handler
func TestCallbackHandler_OpenRedirectAttacks(t *testing.T) {
	attackPayloads := []struct {
		name         string
		returnParam  string
		shouldReject bool
	}{
		{
			name:         "query parameter injection",
			returnParam:  "/events?return=http://evil.com",
			shouldReject: true,
		},
		{
			name:         "fragment injection",
			returnParam:  "/events#http://evil.com",
			shouldReject: true,
		},
		{
			name:         "mixed case http",
			returnParam:  "/hTTp://evil.com",
			shouldReject: true,
		},
	}

	for _, tt := range attackPayloads {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &MockAuthenticator{
				HandleCallbackFunc: func(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
					return &AuthResult{
						Email: "test@example.com",
						Name:  "Test User",
					}, nil
				},
			}

			mockUserService := &MockUserService{
				GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
					return &models.User{
						ID:    1,
						Email: email,
						Name:  name,
						Role:  models.RoleEventManager,
					}, nil
				},
				UpdateLastLoginFunc: func(ctx context.Context, userID int64) error {
					return nil
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
					return nil
				},
			}

			handler := NewCallbackHandler(mockAuth, mockUserService, mockSessionMgr)

			url := "/auth/callback?return=" + tt.returnParam
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			location := rec.Header().Get("Location")

			// Check if the location header contains dangerous patterns
			lowerLocation := strings.ToLower(location)
			isDangerous := strings.Contains(lowerLocation, "http:") ||
				strings.Contains(lowerLocation, "https:") ||
				strings.Contains(lowerLocation, "//evil")

			if isDangerous && tt.shouldReject {
				t.Errorf("SECURITY VULNERABILITY: Callback handler redirects to dangerous URL.\nPayload: %q\nRedirect Location: %q",
					tt.returnParam, location)
			}

			if tt.shouldReject && location != "/" {
				t.Errorf("Malicious payload should redirect to /, but redirected to %q", location)
			}
		})
	}
}

// TestRealWorldOpenRedirectScenario simulates actual phishing attack
func TestRealWorldOpenRedirectScenario(t *testing.T) {
	// Scenario: Attacker sends phishing email with link:
	// https://trustedapp.com/login?return=/events?next=http://evil.com

	mockAuth := &MockAuthenticator{
		HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	handler := NewLoginHandler(mockAuth)

	attackURL := "/login?return=/events?next=http://phishing-site.evil.com/steal-credentials"
	req := httptest.NewRequest(http.MethodGet, attackURL, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	location := rec.Header().Get("Location")

	// The redirect should NOT contain the malicious URL
	if strings.Contains(location, "phishing-site") ||
		strings.Contains(location, "evil.com") ||
		strings.Contains(location, "http://") {
		t.Errorf("CRITICAL SECURITY VULNERABILITY CONFIRMED!\n"+
			"Real-world phishing attack scenario succeeds.\n"+
			"Attacker's URL: %s\n"+
			"App redirects to: %s\n"+
			"This would redirect authenticated users to a phishing site!",
			attackURL, location)
	}
}
