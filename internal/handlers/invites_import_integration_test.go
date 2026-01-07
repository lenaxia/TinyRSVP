package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestImportInvites_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	ctx := context.Background()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 5,
		Status:      models.EventStatusDraft,
		CreatedBy:   user.ID,
	}

	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	service := invites.NewInviteService(generator, inviteRepo)
	handler := NewImportInviteHandlers(service, "https://rsvp.example.com")

	csvContent := `email,name,max_plus_ones
john@example.com,John Doe,2
jane@example.com,Jane Smith,1
bob@example.com,Bob Johnson,0`

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	if _, err := io.WriteString(part, csvContent); err != nil {
		t.Fatalf("Failed to write CSV content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	authCtx := auth.WithUser(req.Context(), user)
	req = req.WithContext(authCtx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result invites.ImportResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}

	if result.Created != 3 {
		t.Errorf("Expected created 3, got %d", result.Created)
	}

	if result.Failed != 0 {
		t.Errorf("Expected failed 0, got %d", result.Failed)
	}

	if result.Duplicates != 0 {
		t.Errorf("Expected duplicates 0, got %d", result.Duplicates)
	}

	invitesList, err := inviteRepo.ListByEventID(ctx, event.ID, repositories.InviteFilters{})
	if err != nil {
		t.Fatalf("Failed to list invites: %v", err)
	}

	if len(invitesList) != 3 {
		t.Errorf("Expected 3 invites in database, got %d", len(invitesList))
	}

	for _, invite := range invitesList {
		if invite.Status != models.InviteStatusDraft {
			t.Errorf("Expected status 'draft', got '%s'", invite.Status)
		}

		if invite.TokenHash == "" {
			t.Error("Expected non-empty token hash")
		}

		if len(invite.TokenHash) != 43 {
			t.Errorf("Expected token hash length 43, got %d", len(invite.TokenHash))
		}
	}
}

func TestImportInvites_Integration_WithDuplicates(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	ctx := context.Background()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 5,
		Status:      models.EventStatusDraft,
		CreatedBy:   user.ID,
	}

	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	service := invites.NewInviteService(generator, inviteRepo)
	handler := NewImportInviteHandlers(service, "https://rsvp.example.com")

	existingEmail := "existing@example.com"
	existingInvite, _, err := service.CreateInvite(ctx, event.ID, nil, &existingEmail, 2, time.Now().Add(60*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create existing invite: %v", err)
	}

	csvContent := `email,name
john@example.com,John Doe
existing@example.com,Existing User
jane@example.com,Jane Smith
john@example.com,Duplicate John`

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	if _, err := io.WriteString(part, csvContent); err != nil {
		t.Fatalf("Failed to write CSV content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	authCtx := auth.WithUser(req.Context(), user)
	req = req.WithContext(authCtx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result invites.ImportResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Total != 4 {
		t.Errorf("Expected total 4, got %d", result.Total)
	}

	if result.Created != 2 {
		t.Errorf("Expected created 2, got %d", result.Created)
	}

	if result.Duplicates != 2 {
		t.Errorf("Expected duplicates 2, got %d", result.Duplicates)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	invitesList, err := inviteRepo.ListByEventID(ctx, event.ID, repositories.InviteFilters{})
	if err != nil {
		t.Fatalf("Failed to list invites: %v", err)
	}

	if len(invitesList) != 3 {
		t.Errorf("Expected 3 invites in database (1 existing + 2 new), got %d", len(invitesList))
	}

	if existingInvite.ID == 0 {
		t.Error("Existing invite should have been preserved")
	}
}
