package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupIntegrationTestDB(t *testing.T) db.Database {
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

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestDeleteUser_CascadesSessions(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	user := &models.User{
		Email: "user@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	session1 := &models.Session{
		ID:        "session1",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := sessionRepo.Create(ctx, session1); err != nil {
		t.Fatalf("Failed to create session1: %v", err)
	}

	session2 := &models.Session{
		ID:        "session2",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := sessionRepo.Create(ctx, session2); err != nil {
		t.Fatalf("Failed to create session2: %v", err)
	}

	sessions, err := sessionRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions before deletion, got %d", len(sessions))
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/2", nil)
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req, "2")

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	sessions, err = sessionRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get sessions after deletion: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions after user deletion (cascade), got %d", len(sessions))
	}

	_, err = userRepo.GetByID(ctx, user.ID)
	var notFoundErr *models.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Error("Expected user to be deleted")
	}
}

func TestDeleteUser_LastAdminProtection(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req, "1")

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 (Conflict), got %d", w.Code)
		t.Logf("Response: %s", w.Body.String())
	}

	_, err := userRepo.GetByID(ctx, admin.ID)
	if err != nil {
		t.Error("Expected admin to still exist after failed deletion")
	}
}

func TestUpdateUserRole_LastAdminProtection(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	reqBody := `{"role": "event_manager"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/users/1/role", bytes.NewReader([]byte(reqBody)))
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UpdateUserRole(w, req, "1")

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 (Conflict), got %d", w.Code)
		t.Logf("Response: %s", w.Body.String())
	}

	user, err := userRepo.GetByID(ctx, admin.ID)
	if err != nil {
		t.Fatalf("Failed to get admin: %v", err)
	}

	if user.Role != models.RoleAdmin {
		t.Errorf("Expected admin role to remain unchanged, got %s", user.Role)
	}
}

func TestListUsers_PermissionCheck_AdminAllowed(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d", w.Code)
	}
}

func TestListUsers_PermissionCheck_NonAdminDenied(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	eventManager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, eventManager); err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	ctx = auth.WithUser(ctx, eventManager)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ListUsers(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-admin, got %d", w.Code)
	}
}

func TestGetUser_PermissionCheck_AdminAllowed(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	targetUser := &models.User{
		Email: "target@example.com",
		Name:  "Target User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/2", nil)
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetUser(w, req, "2")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d", w.Code)
	}
}

func TestGetUser_PermissionCheck_NonAdminDenied(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	eventManager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, eventManager); err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}

	targetUser := &models.User{
		Email: "target@example.com",
		Name:  "Target User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users/2", nil)
	ctx = auth.WithUser(ctx, eventManager)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetUser(w, req, "2")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-admin, got %d", w.Code)
	}
}

func TestUpdateUserRole_PermissionCheck_AdminAllowed(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	targetUser := &models.User{
		Email: "target@example.com",
		Name:  "Target User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	reqBody := `{"role": "admin"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/users/2/role", bytes.NewReader([]byte(reqBody)))
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UpdateUserRole(w, req, "2")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d", w.Code)
	}
}

func TestUpdateUserRole_PermissionCheck_NonAdminDenied(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	eventManager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, eventManager); err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}

	targetUser := &models.User{
		Email: "target@example.com",
		Name:  "Target User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	reqBody := `{"role": "admin"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/users/2/role", bytes.NewReader([]byte(reqBody)))
	ctx = auth.WithUser(ctx, eventManager)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UpdateUserRole(w, req, "2")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-admin, got %d", w.Code)
	}
}

func TestDeleteUser_PermissionCheck_AdminAllowed(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	targetUser := &models.User{
		Email: "target@example.com",
		Name:  "Target User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/2", nil)
	ctx = auth.WithUser(ctx, admin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req, "2")

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for admin, got %d", w.Code)
	}
}

func TestDeleteUser_PermissionCheck_NonAdminDenied(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()
	handler := NewUserHandler(userService, authChecker)

	ctx := context.Background()

	eventManager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, eventManager); err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}

	targetUser := &models.User{
		Email: "target@example.com",
		Name:  "Target User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, targetUser); err != nil {
		t.Fatalf("Failed to create target user: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/2", nil)
	ctx = auth.WithUser(ctx, eventManager)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req, "2")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-admin, got %d", w.Code)
	}
}
