package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

// TestRequireAuth_NoTestBypass verifies that the production RequireAuth
// middleware does NOT honor the X-Test-User-ID header. This is a security
// regression test — if someone accidentally re-adds the bypass to the
// production middleware, this test will fail.
func TestRequireAuth_NoTestBypass(t *testing.T) {
	mockSessionMgr := &mockSessionMgrForBypassTest{}
	mockUserService := &mockUserServiceForBypassTest{}

	handler := RequireAuth(mockSessionMgr, mockUserService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("production RequireAuth should NOT call next handler when X-Test-User-ID is set but no session exists")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Test-User-ID", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect to login (303), got %d — the X-Test-User-ID bypass may be present in production RequireAuth", rec.Code)
	}
}

// TestTestRequireAuth_BypassWorks verifies that the test-only
// TestRequireAuth middleware DOES honor the X-Test-User-ID header.
func TestTestRequireAuth_BypassWorks(t *testing.T) {
	mockSessionMgr := &mockSessionMgrForBypassTest{}
	mockUserService := &mockUserServiceForBypassTest{
		users: map[int64]*models.User{
			1: {ID: 1, Email: "admin@test.com", Name: "Admin", Role: models.RoleAdmin},
		},
	}

	called := false
	handler := TestRequireAuth(mockSessionMgr, mockUserService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("expected user in context")
		}
		if user.ID != 1 {
			t.Errorf("user ID = %d, want 1", user.ID)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Test-User-ID", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected handler to be called via test bypass")
	}
}

// TestTestRequireAuth_FallsThroughWithoutHeader verifies that
// TestRequireAuth falls through to normal session auth when the
// X-Test-User-ID header is absent.
func TestTestRequireAuth_FallsThroughWithoutHeader(t *testing.T) {
	mockSessionMgr := &mockSessionMgrForBypassTest{}
	mockUserService := &mockUserServiceForBypassTest{}

	handler := TestRequireAuth(mockSessionMgr, mockUserService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler without session or test header")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect to login (303) when no session and no test header, got %d", rec.Code)
	}
}

// TestTestRequireAuth_InvalidUserIDFallsThrough verifies that an
// invalid X-Test-User-ID value falls through to normal session auth.
func TestTestRequireAuth_InvalidUserIDFallsThrough(t *testing.T) {
	mockSessionMgr := &mockSessionMgrForBypassTest{}
	mockUserService := &mockUserServiceForBypassTest{}

	handler := TestRequireAuth(mockSessionMgr, mockUserService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler with invalid test user ID")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Test-User-ID", "not-a-number")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect to login (303) for invalid test user ID, got %d", rec.Code)
	}
}

// TestTestRequireAuth_NonExistentUserFallsThrough verifies that a valid
// but non-existent user ID falls through to normal session auth.
func TestTestRequireAuth_NonExistentUserFallsThrough(t *testing.T) {
	mockSessionMgr := &mockSessionMgrForBypassTest{}
	mockUserService := &mockUserServiceForBypassTest{
		users: map[int64]*models.User{}, // empty — user 999 doesn't exist
	}

	handler := TestRequireAuth(mockSessionMgr, mockUserService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler for non-existent test user")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Test-User-ID", "999")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect to login (303) for non-existent test user, got %d", rec.Code)
	}
}

// --- Mocks ---

type mockSessionMgrForBypassTest struct{}

func (m *mockSessionMgrForBypassTest) CreateSession(_ context.Context, _ int64, _ *http.Request) (*models.Session, error) {
	return nil, nil
}
func (m *mockSessionMgrForBypassTest) GetSession(_ context.Context, _ string) (*models.Session, error) {
	return nil, fmt.Errorf("session not found")
}
func (m *mockSessionMgrForBypassTest) DeleteSession(_ context.Context, _ string) error { return nil }
func (m *mockSessionMgrForBypassTest) DeleteUserSessions(_ context.Context, _ int64) error { return nil }
func (m *mockSessionMgrForBypassTest) RefreshSession(_ context.Context, _ string) error { return nil }
func (m *mockSessionMgrForBypassTest) CleanupExpired(_ context.Context) (int64, error) { return 0, nil }
func (m *mockSessionMgrForBypassTest) SetSessionCookie(_ http.ResponseWriter, _ string) error { return nil }
func (m *mockSessionMgrForBypassTest) ClearSessionCookie(_ http.ResponseWriter) error    { return nil }
func (m *mockSessionMgrForBypassTest) GetSessionFromRequest(_ *http.Request) (string, error) {
	return "", fmt.Errorf("session not found")
}

type mockUserServiceForBypassTest struct {
	users map[int64]*models.User
}

func (m *mockUserServiceForBypassTest) CreateUser(_ context.Context, _, _ string, _ *string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserServiceForBypassTest) GetOrCreateUser(_ context.Context, _, _ string, _ *string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserServiceForBypassTest) GetUserByID(_ context.Context, id int64) (*models.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("user not found")
}
func (m *mockUserServiceForBypassTest) GetUserByEmail(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserServiceForBypassTest) UpdateUser(_ context.Context, _ *models.User) error { return nil }
func (m *mockUserServiceForBypassTest) UpdateUserRole(_ context.Context, _ int64, _ models.UserRole) error { return nil }
func (m *mockUserServiceForBypassTest) UpdateLastLogin(_ context.Context, _ int64) error                   { return nil }
func (m *mockUserServiceForBypassTest) DeleteUser(_ context.Context, _ int64) error                        { return nil }
func (m *mockUserServiceForBypassTest) ListUsers(_ context.Context, _, _ int) ([]*models.User, error)     { return nil, nil }
func (m *mockUserServiceForBypassTest) CountUsers(_ context.Context) (int, error)                         { return 0, nil }
func (m *mockUserServiceForBypassTest) CountAdmins(_ context.Context) (int, error)                        { return 0, nil }
