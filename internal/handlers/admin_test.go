package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestAdminDashboardHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := services.NewMockAdminDashboardService(ctrl)
	mockAdminService.EXPECT().GetAdminStats(gomock.Any()).Return(&AdminDashboardStats{
		TotalUsers:   10,
		TotalEvents:  5,
		TotalInvites: 50,
	}, nil)

	handler := NewAdminDashboardHandler(mockAdminService)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := services.NewMockAdminDashboardService(ctrl)
	handler := NewAdminDashboardHandler(mockAdminService)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	w.Header().Set("Accept", "application/json")

	handler.AdminDashboard(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestAdminDashboardHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := services.NewMockAdminDashboardService(ctrl)
	mockAdminService.EXPECT().GetAdminStats(gomock.Any()).Return(nil, errors.New("database error"))
	handler := NewAdminDashboardHandler(mockAdminService)

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

	if w.Code == http.StatusOK {
		t.Error("Expected non-200 status for primary stats failure (should use HandleError)")
	}
}

func TestUserManagementHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := services.NewMockUserService(ctrl)
	mockUserService.EXPECT().ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*models.User{
		{ID: 1, Email: "user1@example.com", Name: "User One", Role: models.RoleEventManager},
	}, nil)
	mockUserService.EXPECT().CountUsers(gomock.Any()).Return(1, nil)

	handler := NewUserManagementHandler(mockUserService)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := services.NewMockUserService(ctrl)
	handler := NewUserManagementHandler(mockUserService)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()
	w.Header().Set("Accept", "application/json")

	handler.UserManagementPage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestUserManagementHandler_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := services.NewMockUserService(ctrl)
	mockUserService.EXPECT().ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
	mockUserService.EXPECT().CountUsers(gomock.Any()).Return(0, errors.New("database error")).AnyTimes()

	handler := NewUserManagementHandler(mockUserService)

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

	if w.Code == http.StatusOK {
		t.Error("Expected non-200 status for primary list failure (should use HandleError)")
	}
}

func TestUserManagementHandler_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := services.NewMockUserService(ctrl)
	mockUserService.EXPECT().ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*models.User{
		{ID: 1, Email: "user1@example.com", Name: "User One", Role: models.RoleEventManager},
	}, nil)
	mockUserService.EXPECT().CountUsers(gomock.Any()).Return(100, nil)

	handler := NewUserManagementHandler(mockUserService)

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
