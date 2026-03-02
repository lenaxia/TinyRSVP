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
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestInviteHandlers_Integration_CreateInvite(t *testing.T) {
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
		Status:      models.EventStatusDraft,
		CreatedBy:   user.ID,
		MaxPlusOnes: 5,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	service := invites.NewIndividualInviteService(generator, inviteRepo, eventRepo)
	handler := NewInviteHandlers(service, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email":         "guest@example.com",
		"name":          "John Doe",
		"max_plus_ones": 2,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	authCtx := auth.WithUser(req.Context(), user)
	req = req.WithContext(authCtx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response CreateInviteResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}

	if response.RSVPURL == "" {
		t.Error("Expected rsvp_url in response")
	}

	if response.Invite == nil {
		t.Fatal("Expected invite in response")
	}

	if response.Invite.Email == nil || *response.Invite.Email != "guest@example.com" {
		t.Errorf("Expected email 'guest@example.com', got %v", response.Invite.Email)
	}

	if response.Invite.Status != models.InviteStatusDraft {
		t.Errorf("Expected status 'draft', got %s", response.Invite.Status)
	}

	if response.Invite.MaxPlusOnes != 2 {
		t.Errorf("Expected max_plus_ones 2, got %d", response.Invite.MaxPlusOnes)
	}

	retrievedInvite, err := inviteRepo.GetByID(ctx, response.Invite.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve invite from database: %v", err)
	}

	if retrievedInvite.Email == nil || *retrievedInvite.Email != "guest@example.com" {
		t.Errorf("Database email mismatch: got %v", retrievedInvite.Email)
	}

	tokenHash, _ := generator.Hash(response.Token)
	if retrievedInvite.TokenHash != tokenHash {
		t.Error("Token hash mismatch in database")
	}
}

func TestInviteHandlers_Integration_DuplicateEmail(t *testing.T) {
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
		Status:      models.EventStatusDraft,
		CreatedBy:   user.ID,
		MaxPlusOnes: 5,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	service := invites.NewIndividualInviteService(generator, inviteRepo, eventRepo)

	firstReq := &invites.CreateIndividualInviteRequest{
		EventID: event.ID,
		Email:   "duplicate@example.com",
	}
	_, err := service.CreateIndividualInvite(ctx, user, firstReq)
	if err != nil {
		t.Fatalf("Failed to create first invite: %v", err)
	}

	handler := NewInviteHandlers(service, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "duplicate@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	authCtx := auth.WithUser(req.Context(), user)
	req = req.WithContext(authCtx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestInviteHandlers_Integration_PermissionCheck(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	ctx := context.Background()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)

	creator := &models.User{
		Email: "creator@example.com",
		Name:  "Event Creator",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, creator); err != nil {
		t.Fatalf("Failed to create creator: %v", err)
	}

	otherUser := &models.User{
		Email: "other@example.com",
		Name:  "Other User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, otherUser); err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   creator.ID,
		MaxPlusOnes: 5,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	secret := []byte("test-secret-key-32-bytes-long!!")
	generator := token.NewGenerator(secret)
	service := invites.NewIndividualInviteService(generator, inviteRepo, eventRepo)
	handler := NewInviteHandlers(service, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	authCtx := auth.WithUser(req.Context(), otherUser)
	req = req.WithContext(authCtx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d. Body: %s", w.Code, w.Body.String())
	}
}
