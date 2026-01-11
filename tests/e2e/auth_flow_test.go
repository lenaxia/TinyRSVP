package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type testServer struct {
	mux         *http.ServeMux
	database    db.Database
	sessionMgr  auth.SessionManager
	userService auth.UserService
	authChecker auth.AuthorizationChecker
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "tinyrsvp-e2e-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: tmpFile.Name(),
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	mux := http.NewServeMux()

	userHandler := handlers.NewUserHandler(userService, authChecker)
	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)

	mux.Handle("/api/users", requireAuth(requireAdmin(http.HandlerFunc(userHandler.ListUsers))))
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(path, "/")
		userID := parts[0]

		switch r.Method {
		case http.MethodGet:
			requireAuth(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.GetUser(w, r, userID)
			}))).ServeHTTP(w, r)
		case http.MethodPatch:
			requireAuth(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.UpdateUserRole(w, r, userID)
			}))).ServeHTTP(w, r)
		case http.MethodDelete:
			requireAuth(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.DeleteUser(w, r, userID)
			}))).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return &testServer{
		mux:         mux,
		database:    database,
		sessionMgr:  sessionMgr,
		userService: userService,
		authChecker: authChecker,
	}
}

func TestForwardAuthFlow(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	t.Run("complete forward auth flow", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "admin@example.com", "Admin User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if err := srv.userService.UpdateUserRole(ctx, user.ID, models.RoleAdmin); err != nil {
			t.Fatalf("Failed to update role: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		session, err := srv.sessionMgr.CreateSession(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: session.ID,
		})

		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("session persistence", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "user@example.com", "Test User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		session, err := srv.sessionMgr.CreateSession(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		retrieved, err := srv.sessionMgr.GetSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve session: %v", err)
		}

		if retrieved.ID != session.ID {
			t.Errorf("Session ID mismatch: got %s, want %s", retrieved.ID, session.ID)
		}
		if retrieved.UserID != user.ID {
			t.Errorf("User ID mismatch: got %d, want %d", retrieved.UserID, user.ID)
		}
	})

	t.Run("bootstrap admin on first login", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "tinyrsvp-bootstrap-*.db")
		if err != nil {
			t.Fatalf("Failed to create temp db: %v", err)
		}
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		database, err := db.NewDatabase(db.Config{
			Type: "sqlite",
			Path: tmpFile.Name(),
		})
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer database.Close()

		migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
		if err != nil {
			t.Fatalf("Failed to create migrator: %v", err)
		}

		if err := migrator.Up(ctx); err != nil {
			t.Fatalf("Failed to run migrations: %v", err)
		}

		userRepo := repositories.NewUserRepository(database)
		userService := auth.NewUserService(userRepo)

		user, err := userService.GetOrCreateUser(ctx, "first@example.com", "First User", nil)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		if user.Role != models.RoleAdmin {
			t.Errorf("First user should be admin, got %s", user.Role)
		}
	})

	t.Run("protected endpoint without auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("Expected status 303 (redirect to login), got %d", rec.Code)
		}

		location := rec.Header().Get("Location")
		if location != "/login?return=%2Fapi%2Fusers" {
			t.Errorf("Expected redirect to /login with return URL, got %s", location)
		}
	})

	t.Run("protected endpoint with invalid session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: "invalid-session-id",
		})
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("Expected status 303 (redirect to login), got %d", rec.Code)
		}

		location := rec.Header().Get("Location")
		if location != "/login?return=%2Fapi%2Fusers" {
			t.Errorf("Expected redirect to /login with return URL, got %s", location)
		}
	})

	t.Run("admin endpoint with non-admin user", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "manager@example.com", "Manager User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if err := srv.userService.UpdateUserRole(ctx, user.ID, models.RoleEventManager); err != nil {
			t.Fatalf("Failed to update role: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		session, err := srv.sessionMgr.CreateSession(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: session.ID,
		})

		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})
}

func TestSessionCleanup(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	t.Run("cleanup expired sessions", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "cleanup@example.com", "Cleanup User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		session, err := srv.sessionMgr.CreateSession(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		_, err = srv.database.DB().ExecContext(ctx,
			"UPDATE sessions SET expires_at = ? WHERE id = ?",
			time.Now().Add(-1*time.Hour), session.ID)
		if err != nil {
			t.Fatalf("Failed to expire session: %v", err)
		}

		count, err := srv.sessionMgr.CleanupExpired(ctx)
		if err != nil {
			t.Fatalf("Failed to cleanup sessions: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 session cleaned up, got %d", count)
		}

		_, err = srv.sessionMgr.GetSession(ctx, session.ID)
		if err == nil {
			t.Error("Expected session to be deleted")
		}
	})

	t.Run("do not cleanup active sessions", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "active@example.com", "Active User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		session, err := srv.sessionMgr.CreateSession(ctx, user.ID, req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		count, err := srv.sessionMgr.CleanupExpired(ctx)
		if err != nil {
			t.Fatalf("Failed to cleanup sessions: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 sessions cleaned up, got %d", count)
		}

		retrieved, err := srv.sessionMgr.GetSession(ctx, session.ID)
		if err != nil {
			t.Fatalf("Session should still exist: %v", err)
		}
		if retrieved.ID != session.ID {
			t.Errorf("Session ID mismatch")
		}
	})
}

func TestLastLoginTracking(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	t.Run("last login updated on login", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "login@example.com", "Login User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.LastLoginAt != nil {
			t.Error("LastLoginAt should be nil for new user")
		}

		time.Sleep(10 * time.Millisecond)

		if err := srv.userService.UpdateLastLogin(ctx, user.ID); err != nil {
			t.Fatalf("Failed to update last login: %v", err)
		}

		updated, err := srv.userService.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if updated.LastLoginAt == nil {
			t.Error("LastLoginAt should be set after update")
		}

		if updated.LastLoginAt.Before(user.CreatedAt) {
			t.Error("LastLoginAt should be after CreatedAt")
		}
	})

	t.Run("last login updated multiple times", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "multilogin@example.com", "Multi Login User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if err := srv.userService.UpdateLastLogin(ctx, user.ID); err != nil {
			t.Fatalf("Failed to update last login: %v", err)
		}

		first, err := srv.userService.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		if err := srv.userService.UpdateLastLogin(ctx, user.ID); err != nil {
			t.Fatalf("Failed to update last login again: %v", err)
		}

		second, err := srv.userService.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if !second.LastLoginAt.After(*first.LastLoginAt) {
			t.Error("Second login should be after first login")
		}
	})
}

func TestUserManagementAPI(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	admin, err := srv.userService.GetOrCreateUser(ctx, "api-admin@example.com", "API Admin", nil)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	if err := srv.userService.UpdateUserRole(ctx, admin.ID, models.RoleAdmin); err != nil {
		t.Fatalf("Failed to update admin role: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	adminSession, err := srv.sessionMgr.CreateSession(ctx, admin.ID, req)
	if err != nil {
		t.Fatalf("Failed to create admin session: %v", err)
	}

	t.Run("list users as admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("get user as admin", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "getuser@example.com", "Get User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d", user.ID), nil)
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("update user role as admin", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "updaterole@example.com", "Update Role User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		body := strings.NewReader(`{"role":"event_manager"}`)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/users/%d", user.ID), body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		updated, err := srv.userService.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}
		if updated.Role != models.RoleEventManager {
			t.Errorf("Expected role event_manager, got %s", updated.Role)
		}
	})

	t.Run("delete user as admin", func(t *testing.T) {
		user, err := srv.userService.GetOrCreateUser(ctx, "deleteuser@example.com", "Delete User", nil)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%d", user.ID), nil)
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", rec.Code)
		}

		_, err = srv.userService.GetUserByID(ctx, user.ID)
		if err == nil {
			t.Error("User should be deleted")
		}
	})
}

func TestConcurrentSessions(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	user, err := srv.userService.GetOrCreateUser(ctx, "concurrent@example.com", "Concurrent User", nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("multiple concurrent sessions for same user", func(t *testing.T) {
		sessions := make([]*models.Session, 5)
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			session, err := srv.sessionMgr.CreateSession(ctx, user.ID, req)
			if err != nil {
				t.Fatalf("Failed to create session %d: %v", i, err)
			}
			sessions[i] = session
		}

		for i, session := range sessions {
			retrieved, err := srv.sessionMgr.GetSession(ctx, session.ID)
			if err != nil {
				t.Errorf("Failed to retrieve session %d: %v", i, err)
			}
			if retrieved.UserID != user.ID {
				t.Errorf("Session %d has wrong user ID", i)
			}
		}
	})
}
