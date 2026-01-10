package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRequireAuth_RedirectToLogin(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		query          string
		wantLocation   string
		wantStatusCode int
	}{
		{
			name:           "redirects to login with return URL for root path",
			path:           "/",
			query:          "",
			wantLocation:   "/login?return=%2F",
			wantStatusCode: http.StatusSeeOther,
		},
		{
			name:           "redirects to login with return URL for events path",
			path:           "/events",
			query:          "",
			wantLocation:   "/login?return=%2Fevents",
			wantStatusCode: http.StatusSeeOther,
		},
		{
			name:           "redirects to login with return URL including query params",
			path:           "/events",
			query:          "page=2&status=published",
			wantLocation:   "/login?return=%2Fevents%3Fpage%3D2%26status%3Dpublished",
			wantStatusCode: http.StatusSeeOther,
		},
		{
			name:           "redirects to login with return URL for nested path",
			path:           "/events/123/edit",
			query:          "",
			wantLocation:   "/login?return=%2Fevents%2F123%2Fedit",
			wantStatusCode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionMgr := &mockSessionManager{
				getSessionFromRequestFunc: func(r *http.Request) (string, error) {
					return "", errors.New("session not found")
				},
			}
			mockUserService := &mockUserService{}

			middleware := RequireAuth(mockSessionMgr, mockUserService)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called when not authenticated")
			}))

			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, rec.Code)
			}

			location := rec.Header().Get("Location")
			if location != tt.wantLocation {
				t.Errorf("Expected Location header %q, got %q", tt.wantLocation, location)
			}
		})
	}
}

func TestRequireAuth_WithValidSession(t *testing.T) {
	t.Run("allows request with valid session", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
			Role:  models.RoleEventManager,
		}

		session := &models.Session{
			ID:     "session-123",
			UserID: user.ID,
		}

		mockSessionMgr := &mockSessionManager{
			getSessionFromRequestFunc: func(r *http.Request) (string, error) {
				return session.ID, nil
			},
			getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
				return session, nil
			},
			refreshSessionFunc: func(ctx context.Context, sessionID string) error {
				return nil
			},
		}

		mockUserService := &mockUserService{
			getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
				return user, nil
			},
		}

		handlerCalled := false
		middleware := RequireAuth(mockSessionMgr, mockUserService)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			
			ctxUser, ok := auth.UserFromContext(r.Context())
			if !ok {
				t.Error("Expected user in context")
				return
			}
			
			if ctxUser.ID != user.ID {
				t.Errorf("Expected user ID %d, got %d", user.ID, ctxUser.ID)
			}
			
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !handlerCalled {
			t.Error("Handler was not called with valid session")
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})
}

