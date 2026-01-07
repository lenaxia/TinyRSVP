package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/tinyrsvp/internal/auth"
	"github.com/yourusername/tinyrsvp/internal/db"
	"github.com/yourusername/tinyrsvp/internal/db/repositories"
	"github.com/yourusername/tinyrsvp/internal/models"
)

func setupIntegrationTest(t *testing.T) (auth.SessionManager, auth.UserService, func()) {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		database.Close()
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)

	cleanup := func() {
		database.Close()
	}

	return sessionMgr, userService, cleanup
}

func TestIntegration_RequireAuth_WithRealDatabase(t *testing.T) {
	sessionMgr, userService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	user, err := userService.CreateUser(ctx, "test@example.com", "Test User", nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	r := httptest.NewRequest("GET", "/protected", nil)
	session, err := sessionMgr.CreateSession(ctx, user.ID, r)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUser, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
		}
		if ctxUser.ID != user.ID {
			t.Errorf("Expected user ID %d, got %d", user.ID, ctxUser.ID)
		}

		ctxSession, ok := auth.SessionFromContext(r.Context())
		if !ok {
			t.Error("Expected session in context")
		}
		if ctxSession.ID != session.ID {
			t.Errorf("Expected session ID %q, got %q", session.ID, ctxSession.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/protected", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: session.ID,
	})

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIntegration_RequireAuth_Unauthorized_WithRealDatabase(t *testing.T) {
	sessionMgr, userService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != "Cookie" {
		t.Errorf("Expected WWW-Authenticate header 'Cookie', got %q", wwwAuth)
	}
}

func TestIntegration_MiddlewareChain_AdminOnly_WithRealDatabase(t *testing.T) {
	sessionMgr, userService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	adminUser, err := userService.CreateUser(ctx, "admin@example.com", "Admin User", nil)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	if err := userService.UpdateUserRole(ctx, adminUser.ID, models.RoleAdmin); err != nil {
		t.Fatalf("Failed to update user role: %v", err)
	}

	r := httptest.NewRequest("GET", "/admin", nil)
	session, err := sessionMgr.CreateSession(ctx, adminUser.ID, r)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
		}
		if user.Role != models.RoleAdmin {
			t.Errorf("Expected admin role, got %s", user.Role)
		}
		w.WriteHeader(http.StatusOK)
	})

	authChecker := auth.NewAuthorizationChecker()
	authMiddleware := RequireAuth(sessionMgr, userService)
	adminMiddleware := RequireAdmin(authChecker)

	chainedHandler := authMiddleware(adminMiddleware(handler))

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/admin", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: session.ID,
	})

	chainedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIntegration_MiddlewareChain_NonAdminDenied_WithRealDatabase(t *testing.T) {
	sessionMgr, userService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	regularUser, err := userService.CreateUser(ctx, "user@example.com", "Regular User", nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if err := userService.UpdateUserRole(ctx, regularUser.ID, models.RoleEventManager); err != nil {
		t.Fatalf("Failed to update user role: %v", err)
	}

	r := httptest.NewRequest("GET", "/admin", nil)
	session, err := sessionMgr.CreateSession(ctx, regularUser.ID, r)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
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
	r = httptest.NewRequest("GET", "/admin", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: session.ID,
	})

	chainedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestIntegration_RequireEventManager_BothRolesAllowed_WithRealDatabase(t *testing.T) {
	sessionMgr, userService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name  string
		email string
		role  models.UserRole
	}{
		{"admin allowed", "admin@example.com", models.RoleAdmin},
		{"event manager allowed", "manager@example.com", models.RoleEventManager},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.CreateUser(ctx, tt.email, tt.name, nil)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			if err := userService.UpdateUserRole(ctx, user.ID, tt.role); err != nil {
				t.Fatalf("Failed to update user role: %v", err)
			}

			r := httptest.NewRequest("GET", "/events", nil)
			session, err := sessionMgr.CreateSession(ctx, user.ID, r)
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			authChecker := auth.NewAuthorizationChecker()
			authMiddleware := RequireAuth(sessionMgr, userService)
			managerMiddleware := RequireEventManager(authChecker)

			chainedHandler := authMiddleware(managerMiddleware(handler))

			w := httptest.NewRecorder()
			r = httptest.NewRequest("GET", "/events", nil)
			r.AddCookie(&http.Cookie{
				Name:  "tinyrsvp_session",
				Value: session.ID,
			})

			chainedHandler.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

func TestIntegration_SessionRefresh_UpdatesLastAccess(t *testing.T) {
	sessionMgr, userService, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	user, err := userService.CreateUser(ctx, "test@example.com", "Test User", nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	r := httptest.NewRequest("GET", "/protected", nil)
	session, err := sessionMgr.CreateSession(ctx, user.ID, r)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	originalLastAccess := session.LastAccessedAt

	time.Sleep(100 * time.Millisecond)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth(sessionMgr, userService)

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/protected", nil)
	r.AddCookie(&http.Cookie{
		Name:  "tinyrsvp_session",
		Value: session.ID,
	})

	middleware(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	refreshedSession, err := sessionMgr.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get refreshed session: %v", err)
	}

	if !refreshedSession.LastAccessedAt.After(originalLastAccess) {
		t.Error("Expected LastAccessedAt to be updated after request")
	}
}
