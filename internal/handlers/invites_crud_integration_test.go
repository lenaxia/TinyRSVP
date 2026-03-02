package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestInviteCRUDIntegration(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	emailQueueRepo := repositories.NewEmailQueueRepository(database)

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	endTime := time.Now().Add(31 * 24 * time.Hour)
	event := &models.Event{
		Title:       "Test Event",
		Description: testutil.StringPtr("Test Description"),
		Location:    testutil.StringPtr("Test Location"),
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		EndTime:     &endTime,
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		Status:      models.EventStatusDraft,
		CreatedBy:   admin.ID,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tokenGenerator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)

	getHandler := NewGetInviteHandlers(inviteService, eventRepo)
	updateHandler := NewUpdateInviteHandlers(inviteService, eventRepo)
	deleteHandler := NewDeleteInviteHandlers(inviteService, eventRepo)
	sendHandler := NewSendInviteHandlers(inviteService, eventRepo, emailQueueRepo, "https://rsvp.example.com")

	email := "test@example.com"
	invite, plainToken, err := inviteService.CreateInvite(ctx, event.ID, testutil.StringPtr("Test User"), &email, 2, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	t.Run("GET /api/invites/{id}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/invites/1", nil)
		req.Header.Set("Accept", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("inviteId", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(auth.WithUser(req.Context(), admin))

		w := httptest.NewRecorder()
		getHandler.GetInvite(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["email"] != email {
			t.Errorf("Expected email %s, got %v", email, response["email"])
		}
	})

	t.Run("PUT /api/invites/{id}", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":          "Updated Name",
			"max_plus_ones": 3,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader(body))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("inviteId", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(auth.WithUser(req.Context(), admin))

		w := httptest.NewRecorder()
		updateHandler.UpdateInvite(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		updatedInvite, err := inviteService.GetInviteByID(ctx, invite.ID)
		if err != nil {
			t.Fatalf("Failed to get updated invite: %v", err)
		}

		if updatedInvite.Name == nil || *updatedInvite.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %v", updatedInvite.Name)
		}

		if updatedInvite.MaxPlusOnes != 3 {
			t.Errorf("Expected max_plus_ones 3, got %d", updatedInvite.MaxPlusOnes)
		}
	})

	t.Run("POST /api/invites/{id}/send", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/invites/1/send", nil)
		req.Header.Set("Accept", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("inviteId", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(auth.WithUser(req.Context(), admin))

		w := httptest.NewRecorder()
		sendHandler.SendInvite(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		sentInvite, err := inviteService.GetInviteByID(ctx, invite.ID)
		if err != nil {
			t.Fatalf("Failed to get sent invite: %v", err)
		}

		if sentInvite.Status != models.InviteStatusSent {
			t.Errorf("Expected status %s, got %s", models.InviteStatusSent, sentInvite.Status)
		}

		if sentInvite.SentAt == nil {
			t.Error("Expected SentAt to be set")
		}
	})

	t.Run("DELETE /api/invites/{id}", func(t *testing.T) {
		deleteInvite, _, err := inviteService.CreateInvite(ctx, event.ID, testutil.StringPtr("Delete Test"), &email, 2, time.Now().Add(30*24*time.Hour))
		if err != nil {
			t.Fatalf("Failed to create invite for deletion: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/invites/2", nil)
		req.Header.Set("Accept", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("inviteId", "2")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(auth.WithUser(req.Context(), admin))

		w := httptest.NewRecorder()
		deleteHandler.DeleteInvite(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		_, err = inviteService.GetInviteByID(ctx, deleteInvite.ID)
		if err == nil {
			t.Error("Expected error when getting deleted invite, got nil")
		}
	})

	_ = plainToken
}
