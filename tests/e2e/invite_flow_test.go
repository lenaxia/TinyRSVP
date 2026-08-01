package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

type inviteTestServer struct {
	router        chi.Router
	database      db.Database
	sessionMgr    auth.SessionManager
	userService   auth.UserService
	authChecker   auth.AuthorizationChecker
	eventService  events.Service
	inviteService invites.IndividualInviteService
	inviteRepo    repositories.InviteRepository
}

func setupInviteTestServer(t *testing.T) *inviteTestServer {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "tinyrsvp-invite-e2e-*.db")
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
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	sessionMgr := auth.NewSessionManager(sessionRepo, false, 0)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, nil, eventValidator, authChecker)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewIndividualInviteService(tokenGenerator, inviteRepo, eventRepo)

	router := chi.NewRouter()
	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	router.Use(func(next http.Handler) http.Handler {
		return requireAuth(next)
	})

	eventHandlers := handlers.NewEventHandlers(eventService)
	eventHandlers.RegisterRoutes(router)

	inviteHandlers := handlers.NewInviteHandlers(inviteService, "http://localhost:8080")
	inviteHandlers.RegisterRoutes(router)

	return &inviteTestServer{
		router:        router,
		database:      database,
		sessionMgr:    sessionMgr,
		userService:   userService,
		authChecker:   authChecker,
		eventService:  eventService,
		inviteService: inviteService,
		inviteRepo:    inviteRepo,
	}
}

func TestInviteEndpointExists(t *testing.T) {
	srv := setupInviteTestServer(t)
	ctx := context.Background()

	admin, err := srv.userService.GetOrCreateUser(ctx, "admin@example.com", "Admin User", nil)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	if err := srv.userService.UpdateUserRole(ctx, admin.ID, models.RoleAdmin); err != nil {
		t.Fatalf("Failed to update admin role: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	adminSession, err := srv.sessionMgr.CreateSession(ctx, admin.ID, req)
	if err != nil {
		t.Fatalf("Failed to create admin session: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	description := "Test Description"
	location := "Test Location"
	event := &models.Event{
		Title:       "Test Event",
		Description: &description,
		Location:    &location,
		StartTime:   startTime,
		EndTime:     &endTime,
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		Status:      models.EventStatusDraft,
		CreatedBy:   admin.ID,
	}

	eventRepo := repositories.NewEventRepository(srv.database)
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	t.Run("POST /api/events/:eventId/invites endpoint exists", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "guest@example.com",
			"name":  "Guest User",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})

		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response["invite"] == nil {
			t.Error("Response should contain invite")
		}
		if response["token"] == nil {
			t.Error("Response should contain token")
		}
		if response["rsvp_url"] == nil {
			t.Error("Response should contain rsvp_url")
		}

		token, ok := response["token"].(string)
		if !ok || token == "" {
			t.Error("Token should be a non-empty string")
		}

		rsvpURL, ok := response["rsvp_url"].(string)
		if !ok || !strings.HasPrefix(rsvpURL, "http://localhost:8080/rsvp/") {
			t.Errorf("RSVP URL should start with base URL, got: %s", rsvpURL)
		}
	})

	t.Run("invite is created in database", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "dbtest@example.com",
			"name":  "DB Test User",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})

		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		inviteData := response["invite"].(map[string]interface{})
		inviteID := int64(inviteData["id"].(float64))

		invite, err := srv.inviteRepo.GetByID(ctx, inviteID)
		if err != nil {
			t.Fatalf("Failed to get invite from database: %v", err)
		}

		if invite.EventID != event.ID {
			t.Errorf("Expected event ID %d, got %d", event.ID, invite.EventID)
		}
		if invite.Email == nil || *invite.Email != "dbtest@example.com" {
			t.Errorf("Expected email dbtest@example.com, got %v", invite.Email)
		}
		if invite.Name == nil || *invite.Name != "DB Test User" {
			t.Errorf("Expected name 'DB Test User', got %v", invite.Name)
		}
		if invite.TokenHash == "" {
			t.Error("Token hash should not be empty")
		}
	})

	t.Run("unauthorized request fails", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "unauthorized@example.com",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusSeeOther && rec.Code != http.StatusFound {
			t.Errorf("Expected status 401, 302, or 303 for unauthorized request, got %d", rec.Code)
		}
	})

	t.Run("invalid event ID fails", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "invalid@example.com",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/events/99999/invites", bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})

		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", rec.Code)
		}
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "duplicate@example.com",
		}
		bodyJSON, _ := json.Marshal(body)

		req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON))
		req1.Header.Set("Content-Type", "application/json")
		req1.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})

		rec1 := httptest.NewRecorder()
		srv.router.ServeHTTP(rec1, req1)

		if rec1.Code != http.StatusCreated {
			t.Fatalf("First invite creation failed: %d", rec1.Code)
		}

		bodyJSON2, _ := json.Marshal(body)
		req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON2))
		req2.Header.Set("Content-Type", "application/json")
		req2.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: adminSession.ID,
		})

		rec2 := httptest.NewRecorder()
		srv.router.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusConflict {
			t.Errorf("Expected status 409 for duplicate, got %d", rec2.Code)
		}
	})
}

func TestInvitePermissions(t *testing.T) {
	srv := setupInviteTestServer(t)
	ctx := context.Background()

	eventManager, err := srv.userService.GetOrCreateUser(ctx, "manager@example.com", "Event Manager", nil)
	if err != nil {
		t.Fatalf("Failed to create event manager: %v", err)
	}
	if err := srv.userService.UpdateUserRole(ctx, eventManager.ID, models.RoleEventManager); err != nil {
		t.Fatalf("Failed to update role: %v", err)
	}

	otherManager, err := srv.userService.GetOrCreateUser(ctx, "other@example.com", "Other Manager", nil)
	if err != nil {
		t.Fatalf("Failed to create other manager: %v", err)
	}
	if err := srv.userService.UpdateUserRole(ctx, otherManager.ID, models.RoleEventManager); err != nil {
		t.Fatalf("Failed to update role: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	managerSession, err := srv.sessionMgr.CreateSession(ctx, eventManager.ID, req)
	if err != nil {
		t.Fatalf("Failed to create manager session: %v", err)
	}

	otherSession, err := srv.sessionMgr.CreateSession(ctx, otherManager.ID, req)
	if err != nil {
		t.Fatalf("Failed to create other session: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	description := "Test Description"
	location := "Test Location"
	event := &models.Event{
		Title:       "Manager Event",
		Description: &description,
		Location:    &location,
		StartTime:   startTime,
		EndTime:     &endTime,
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		Status:      models.EventStatusDraft,
		CreatedBy:   eventManager.ID,
	}

	eventRepo := repositories.NewEventRepository(srv.database)
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	t.Run("event creator can create invite", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "creator@example.com",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: managerSession.ID,
		})

		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("non-creator cannot create invite", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "noncreator@example.com",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/events/%d/invites", event.ID), bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "tinyrsvp_session",
			Value: otherSession.ID,
		})

		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})
}
