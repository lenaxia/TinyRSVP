package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/admin"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupAdminIntegrationTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestAdminDashboard_Integration_Success(t *testing.T) {
	database := setupAdminIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	adminUser := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	userService := auth.NewUserService(userRepo)
	adminService := admin.NewAdminService(userService, eventRepo, inviteRepo)
	handler := NewAdminDashboardHandler(adminService)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := auth.WithUser(req.Context(), adminUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.AdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestAdminDashboard_Integration_NonAdminDenied(t *testing.T) {
	database := setupAdminIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	regularUser := &models.User{
		Email: "user@example.com",
		Name:  "Regular User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(context.Background(), regularUser); err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	userService := auth.NewUserService(userRepo)
	adminService := admin.NewAdminService(userService, eventRepo, inviteRepo)
	handler := NewAdminDashboardHandler(adminService)
	authChecker := auth.NewAuthorizationChecker()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := auth.WithUser(req.Context(), regularUser)
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")

	requireAdmin := middleware.RequireAdmin(authChecker)
	protectedHandler := requireAdmin(http.HandlerFunc(handler.AdminDashboard))

	w := httptest.NewRecorder()
	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestUserManagement_Integration_Success(t *testing.T) {
	database := setupAdminIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)

	adminUser := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	regularUser := &models.User{
		Email: "user@example.com",
		Name:  "Regular User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(context.Background(), regularUser); err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	userService := auth.NewUserService(userRepo)
	handler := NewUserManagementHandler(userService)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	ctx := auth.WithUser(req.Context(), adminUser)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.UserManagementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestUserManagement_Integration_NonAdminDenied(t *testing.T) {
	database := setupAdminIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)

	regularUser := &models.User{
		Email: "user@example.com",
		Name:  "Regular User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(context.Background(), regularUser); err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	userService := auth.NewUserService(userRepo)
	handler := NewUserManagementHandler(userService)
	authChecker := auth.NewAuthorizationChecker()

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	ctx := auth.WithUser(req.Context(), regularUser)
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")

	requireAdmin := middleware.RequireAdmin(authChecker)
	protectedHandler := requireAdmin(http.HandlerFunc(handler.UserManagementPage))

	w := httptest.NewRecorder()
	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestAdminDashboard_Integration_WithStats(t *testing.T) {
	database := setupAdminIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	adminUser := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   adminUser.ID,
		MaxPlusOnes: 0,
	}
	if err := eventRepo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	userService := auth.NewUserService(userRepo)
	adminService := admin.NewAdminService(userService, eventRepo, inviteRepo)
	handler := NewAdminDashboardHandler(adminService)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := auth.WithUser(req.Context(), adminUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.AdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestUserManagement_Integration_WithPagination(t *testing.T) {
	database := setupAdminIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)

	adminUser := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	for i := 1; i <= 15; i++ {
		user := &models.User{
			Email: fmt.Sprintf("user%d@example.com", i),
			Name:  fmt.Sprintf("User %d", i),
			Role:  models.RoleEventManager,
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	userService := auth.NewUserService(userRepo)
	handler := NewUserManagementHandler(userService)

	req := httptest.NewRequest(http.MethodGet, "/admin/users?limit=10&offset=0", nil)
	ctx := auth.WithUser(req.Context(), adminUser)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.UserManagementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
