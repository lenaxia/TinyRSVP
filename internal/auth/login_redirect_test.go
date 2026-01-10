package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginHandler_ReturnURL(t *testing.T) {
	tests := []struct {
		name           string
		returnParam    string
		wantLocation   string
		wantStatusCode int
	}{
		{
			name:           "redirects to root when no return param",
			returnParam:    "",
			wantLocation:   "/",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "redirects to return URL when provided",
			returnParam:    "/events",
			wantLocation:   "/events",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "redirects to return URL with query params",
			returnParam:    "/events?page=2",
			wantLocation:   "/events?page=2",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "redirects to nested path",
			returnParam:    "/events/123/edit",
			wantLocation:   "/events/123/edit",
			wantStatusCode: http.StatusFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &MockAuthenticator{
				HandleLoginFunc: func(w http.ResponseWriter, r *http.Request) error {
					return nil
				},
			}

			handler := NewLoginHandler(mockAuth)

			url := "/login"
			if tt.returnParam != "" {
				url += "?return=" + tt.returnParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			location := rec.Header().Get("Location")
			if location != tt.wantLocation {
				t.Errorf("Expected Location %q, got %q", tt.wantLocation, location)
			}
		})
	}
}

func TestCallbackHandler_ReturnURL(t *testing.T) {
	tests := []struct {
		name           string
		returnParam    string
		wantLocation   string
		wantStatusCode int
	}{
		{
			name:           "redirects to root when no return param",
			returnParam:    "",
			wantLocation:   "/",
			wantStatusCode: http.StatusFound,
		},
		{
			name:           "redirects to return URL when provided",
			returnParam:    "/events",
			wantLocation:   "/events",
			wantStatusCode: http.StatusFound,
		},
	}

	for _, tt := range tests {
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
				GetOrCreateUserFunc: func(email, name string, oidcSubject *string) (*models.User, error) {
					return &models.User{
						ID:    1,
						Email: email,
						Name:  name,
						Role:  models.RoleEventManager,
					}, nil
				},
				UpdateLastLoginFunc: func(userID int64) error {
					return nil
				},
			}

			mockSessionMgr := &MockSessionManager{
				CreateSessionFunc: func(userID int64, r *http.Request) (*models.Session, error) {
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

			url := "/auth/callback"
			if tt.returnParam != "" {
				url += "?return=" + tt.returnParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			location := rec.Header().Get("Location")
			if location != tt.wantLocation {
				t.Errorf("Expected Location %q, got %q", tt.wantLocation, location)
			}
		})
	}
}
