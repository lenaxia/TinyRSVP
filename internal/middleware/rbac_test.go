package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockSessionManager struct {
	getSessionFromRequestFunc func(r *http.Request) (string, error)
	getSessionFunc            func(ctx context.Context, sessionID string) (*models.Session, error)
	refreshSessionFunc        func(ctx context.Context, sessionID string) error
}

func (m *mockSessionManager) CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
	return nil, nil
}

func (m *mockSessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	if m.getSessionFunc != nil {
		return m.getSessionFunc(ctx, sessionID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockSessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	if m.refreshSessionFunc != nil {
		return m.refreshSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *mockSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockSessionManager) DeleteUserSessions(ctx context.Context, userID int64) error {
	return nil
}

func (m *mockSessionManager) CleanupExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockSessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	return nil
}

func (m *mockSessionManager) ClearSessionCookie(w http.ResponseWriter) error {
	return nil
}

func (m *mockSessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
	if m.getSessionFromRequestFunc != nil {
		return m.getSessionFromRequestFunc(r)
	}
	return "", fmt.Errorf("not implemented")
}

type mockUserService struct {
	getUserByIDFunc func(ctx context.Context, id int64) (*models.User, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
	return nil
}

func (m *mockUserService) UpdateLastLogin(ctx context.Context, userID int64) error {
	return nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int64) error {
	return nil
}

func (m *mockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return nil, nil
}

func (m *mockUserService) CountUsers(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockUserService) CountAdmins(ctx context.Context) (int, error) {
	return 0, nil
}

func TestRequireAuth_ValidSession(t *testing.T) {
	mockSession := &models.Session{
		ID:        "session123",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mockUser := &models.User{
		ID:    1,
		Email: "user@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}

	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "session123", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			return mockSession, nil
		},
		refreshSessionFunc: func(ctx context.Context, sessionID string) error {
			return nil
		},
	}

	userService := &mockUserService{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return mockUser, nil
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
		}
		if user.ID != mockUser.ID {
			t.Errorf("Expected user ID %d, got %d", mockUser.ID, user.ID)
		}

		session, ok := auth.SessionFromContext(r.Context())
		if !ok {
			t.Error("Expected session in context")
		}
		if session.ID != mockSession.ID {
			t.Errorf("Expected session ID %q, got %q", mockSession.ID, session.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireAuth_MissingSessionCookie(t *testing.T) {
	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "", http.ErrNoCookie
		},
	}

	userService := &mockUserService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	expectedLocation := "/login?return=%2Fprotected"
	if location != expectedLocation {
		t.Errorf("Expected Location header %q, got %q", expectedLocation, location)
	}
}

func TestRequireAuth_InvalidSession(t *testing.T) {
	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "invalid", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			return nil, fmt.Errorf("session not found")
		},
	}

	userService := &mockUserService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	expectedLocation := "/login?return=%2Fprotected"
	if location != expectedLocation {
		t.Errorf("Expected Location header %q, got %q", expectedLocation, location)
	}
}

func TestRequireAuth_ExpiredSession(t *testing.T) {
	expiredSession := &models.Session{
		ID:        "expired",
		UserID:    1,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "expired", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			if expiredSession.IsExpired() {
				return nil, fmt.Errorf("session expired")
			}
			return expiredSession, nil
		},
	}

	userService := &mockUserService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	expectedLocation := "/login?return=%2Fprotected"
	if location != expectedLocation {
		t.Errorf("Expected Location header %q, got %q", expectedLocation, location)
	}
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	mockSession := &models.Session{
		ID:        "session123",
		UserID:    999,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "session123", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			return mockSession, nil
		},
	}

	userService := &mockUserService{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	expectedLocation := "/login?return=%2Fprotected"
	if location != expectedLocation {
		t.Errorf("Expected Location header %q, got %q", expectedLocation, location)
	}
}

func TestRequireAdmin_AdminAllowed(t *testing.T) {
	user := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	middleware := RequireAdmin(authChecker)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin", nil)

	ctx := auth.WithUser(r.Context(), user)
	r = r.WithContext(ctx)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireAdmin_EventManagerDenied(t *testing.T) {
	user := &models.User{
		ID:   2,
		Role: models.RoleEventManager,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	middleware := RequireAdmin(authChecker)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin", nil)

	ctx := auth.WithUser(r.Context(), user)
	r = r.WithContext(ctx)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireAdmin_NoUserInContext(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	middleware := RequireAdmin(authChecker)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireEventManager_AdminAllowed(t *testing.T) {
	user := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	middleware := RequireEventManager(authChecker)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/events", nil)

	ctx := auth.WithUser(r.Context(), user)
	r = r.WithContext(ctx)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireEventManager_EventManagerAllowed(t *testing.T) {
	user := &models.User{
		ID:   2,
		Role: models.RoleEventManager,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	middleware := RequireEventManager(authChecker)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/events", nil)

	ctx := auth.WithUser(r.Context(), user)
	r = r.WithContext(ctx)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireEventManager_NoUserDenied(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	middleware := RequireEventManager(authChecker)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/events", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestMiddlewareChaining(t *testing.T) {
	mockSession := &models.Session{
		ID:        "session123",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mockUser := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin",
		Role:  models.RoleAdmin,
	}

	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "session123", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			return mockSession, nil
		},
		refreshSessionFunc: func(ctx context.Context, sessionID string) error {
			return nil
		},
	}

	userService := &mockUserService{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return mockUser, nil
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
		}
		if user.Role != models.RoleAdmin {
			t.Error("Expected admin user")
		}
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	authMiddleware := RequireAuth(sessionMgr, userService)
	adminMiddleware := RequireAdmin(authChecker)

	chainedHandler := authMiddleware(adminMiddleware(handler))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/endpoint", nil)

	chainedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestMiddlewareChaining_AuthFailsBeforeRoleCheck(t *testing.T) {
	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "", http.ErrNoCookie
		},
	}

	userService := &mockUserService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	authMiddleware := RequireAuth(sessionMgr, userService)
	adminMiddleware := RequireAdmin(authChecker)

	chainedHandler := authMiddleware(adminMiddleware(handler))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/endpoint", nil)

	chainedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	expectedLocation := "/login?return=%2Fadmin%2Fendpoint"
	if location != expectedLocation {
		t.Errorf("Expected Location header %q, got %q", expectedLocation, location)
	}
}

func TestMiddlewareChaining_AuthSucceedsRoleFails(t *testing.T) {
	mockSession := &models.Session{
		ID:        "session123",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mockUser := &models.User{
		ID:    1,
		Email: "user@example.com",
		Name:  "Regular User",
		Role:  models.RoleEventManager,
	}

	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "session123", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			return mockSession, nil
		},
		refreshSessionFunc: func(ctx context.Context, sessionID string) error {
			return nil
		},
	}

	userService := &mockUserService{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return mockUser, nil
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	authMiddleware := RequireAuth(sessionMgr, userService)
	adminMiddleware := RequireAdmin(authChecker)

	chainedHandler := authMiddleware(adminMiddleware(handler))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/endpoint", nil)

	chainedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireAuth_RefreshSessionError(t *testing.T) {
	mockSession := &models.Session{
		ID:        "session123",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mockUser := &models.User{
		ID:    1,
		Email: "user@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}

	sessionMgr := &mockSessionManager{
		getSessionFromRequestFunc: func(r *http.Request) (string, error) {
			return "session123", nil
		},
		getSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
			return mockSession, nil
		},
		refreshSessionFunc: func(ctx context.Context, sessionID string) error {
			return fmt.Errorf("database error")
		},
	}

	userService := &mockUserService{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return mockUser, nil
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
		}
		if user.ID != mockUser.ID {
			t.Errorf("Expected user ID %d, got %d", mockUser.ID, user.ID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d (refresh failure should not block request)", http.StatusOK, w.Code)
	}
}
