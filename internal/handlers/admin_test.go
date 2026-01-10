package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockAdminDashboardService struct {
	stats *AdminDashboardStats
	err   error
}

func (m *mockAdminDashboardService) GetAdminStats(ctx context.Context) (*AdminDashboardStats, error) {
	return m.stats, m.err
}

type mockUserListService struct {
	users []*models.User
	total int
	err   error
}

func (m *mockUserListService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return m.users, m.err
}

func (m *mockUserListService) CountUsers(ctx context.Context) (int, error) {
	return m.total, m.err
}

func TestAdminDashboardHandler_Success(t *testing.T) {
	service := &mockAdminDashboardService{
		stats: &AdminDashboardStats{
			TotalUsers:   10,
			TotalEvents:  5,
			TotalInvites: 50,
		},
	}

	handler := NewAdminDashboardHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.AdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAdminDashboardHandler_Unauthorized(t *testing.T) {
	service := &mockAdminDashboardService{
		stats: &AdminDashboardStats{},
	}

	handler := NewAdminDashboardHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	w.Header().Set("Accept", "application/json")

	handler.AdminDashboard(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestAdminDashboardHandler_ServiceError(t *testing.T) {
	service := &mockAdminDashboardService{
		err: errors.New("database error"),
	}

	handler := NewAdminDashboardHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.AdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (renders error state), got %d", w.Code)
	}
}

func TestUserManagementHandler_Success(t *testing.T) {
	service := &mockUserListService{
		users: []*models.User{
			{
				ID:    1,
				Email: "user1@example.com",
				Name:  "User One",
				Role:  models.RoleEventManager,
			},
		},
		total: 1,
	}

	handler := NewUserManagementHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.UserManagementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUserManagementHandler_Unauthorized(t *testing.T) {
	service := &mockUserListService{
		users: []*models.User{},
		total: 0,
	}

	handler := NewUserManagementHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()
	w.Header().Set("Accept", "application/json")

	handler.UserManagementPage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestUserManagementHandler_ServiceError(t *testing.T) {
	service := &mockUserListService{
		err: errors.New("database error"),
	}

	handler := NewUserManagementHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.UserManagementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (renders error state), got %d", w.Code)
	}
}

func TestUserManagementHandler_WithPagination(t *testing.T) {
	service := &mockUserListService{
		users: []*models.User{
			{
				ID:    1,
				Email: "user1@example.com",
				Name:  "User One",
				Role:  models.RoleEventManager,
			},
		},
		total: 100,
	}

	handler := NewUserManagementHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/admin/users?limit=10&offset=20", nil)
	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.UserManagementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
