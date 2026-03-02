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
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEventHandlers_FullCRUDFlow_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, validator, authChecker)
	handlers := NewEventHandlers(eventService)

	ctx := context.Background()

	manager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager); err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	createReq := `{
		"title": "Integration Test Event",
		"description": "Testing full CRUD flow",
		"start_time": "2026-06-15T14:00:00-07:00",
		"timezone": "America/Los_Angeles",
		"max_plus_ones": 2
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader([]byte(createReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	ctx = auth.WithUser(ctx, manager)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var createResp EventResponse
	if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	if createResp.ID == 0 {
		t.Fatal("Expected non-zero event ID")
	}
	if createResp.Status != models.EventStatusDraft {
		t.Errorf("Expected status draft, got %s", createResp.Status)
	}
	if createResp.CreatedBy != manager.ID {
		t.Errorf("Expected created_by %d, got %d", manager.ID, createResp.CreatedBy)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/events/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.GetEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var getResp EventResponse
	if err := json.NewDecoder(w.Body).Decode(&getResp); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	if getResp.Title != "Integration Test Event" {
		t.Errorf("Expected title 'Integration Test Event', got %s", getResp.Title)
	}

	updateReq := `{
		"title": "Updated Event Title",
		"timezone": "America/Los_Angeles",
		"version": 1
	}`

	req = httptest.NewRequest(http.MethodPut, "/api/events/1", bytes.NewReader([]byte(updateReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var updateResp EventResponse
	if err := json.NewDecoder(w.Body).Decode(&updateResp); err != nil {
		t.Fatalf("Failed to decode update response: %v", err)
	}

	if updateResp.Title != "Updated Event Title" {
		t.Errorf("Expected updated title, got %s", updateResp.Title)
	}
	if updateResp.Version != 2 {
		t.Errorf("Expected version 2, got %d", updateResp.Version)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/events/1", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.DeleteEvent(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/events/1", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.GetEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 after soft delete, got %d", w.Code)
	}

	var deletedResp EventResponse
	if err := json.NewDecoder(w.Body).Decode(&deletedResp); err != nil {
		t.Fatalf("Failed to decode deleted event response: %v", err)
	}

	if deletedResp.Status != models.EventStatusArchived {
		t.Errorf("Expected status archived after delete, got %s", deletedResp.Status)
	}
}

func TestEventHandlers_LifecycleTransitions_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, validator, authChecker)
	handlers := NewEventHandlers(eventService)

	ctx := context.Background()

	manager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager); err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	event := &models.Event{
		Title:       "Lifecycle Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		CreatedBy:   manager.ID,
		Status:      models.EventStatusDraft,
		Version:     1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/publish", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.PublishEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for publish, got %d: %s", w.Code, w.Body.String())
	}

	updatedEvent, err := eventRepo.GetByID(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("Failed to get updated event: %v", err)
	}
	if updatedEvent.Status != models.EventStatusPublished {
		t.Errorf("Expected status published, got %s", updatedEvent.Status)
	}

	cancelReq := `{"reason": "Testing cancellation flow with valid reason"}`
	req = httptest.NewRequest(http.MethodPost, "/api/events/1/cancel", bytes.NewReader([]byte(cancelReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.CancelEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for cancel, got %d: %s", w.Code, w.Body.String())
	}

	updatedEvent, err = eventRepo.GetByID(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("Failed to get cancelled event: %v", err)
	}
	if updatedEvent.Status != models.EventStatusCancelled {
		t.Errorf("Expected status cancelled, got %s", updatedEvent.Status)
	}
}

func TestEventHandlers_PermissionEnforcement_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, validator, authChecker)
	handlers := NewEventHandlers(eventService)

	ctx := context.Background()

	manager1 := &models.User{
		Email: "manager1@example.com",
		Name:  "Event Manager 1",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager1); err != nil {
		t.Fatalf("Failed to create manager1: %v", err)
	}

	manager2 := &models.User{
		Email: "manager2@example.com",
		Name:  "Event Manager 2",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager2); err != nil {
		t.Fatalf("Failed to create manager2: %v", err)
	}

	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	event := &models.Event{
		Title:       "Permission Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		CreatedBy:   manager1.ID,
		Status:      models.EventStatusDraft,
		Version:     1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	updateReq := `{"title": "Unauthorized Update", "timezone": "America/Los_Angeles", "version": 1}`
	req := httptest.NewRequest(http.MethodPut, "/api/events/1", bytes.NewReader([]byte(updateReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), manager2)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.UpdateEvent(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for non-owner manager, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/events/1", bytes.NewReader([]byte(updateReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	ctx = auth.WithUser(req.Context(), admin)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	handlers.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEventHandlers_ConcurrentUpdates_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	authChecker := auth.NewAuthorizationChecker()
	validator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, validator, authChecker)
	handlers := NewEventHandlers(eventService)

	ctx := context.Background()

	manager := &models.User{
		Email: "manager@example.com",
		Name:  "Event Manager",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, manager); err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	event := &models.Event{
		Title:       "Concurrent Update Test",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 2,
		CreatedBy:   manager.ID,
		Status:      models.EventStatusDraft,
		Version:     1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	updateReq1 := `{"title": "First Update", "timezone": "America/Los_Angeles", "version": 1}`
	req1 := httptest.NewRequest(http.MethodPut, "/api/events/1", bytes.NewReader([]byte(updateReq1)))
	req1.Header.Set("Content-Type", "application/json")
	rctx1 := chi.NewRouteContext()
	rctx1.URLParams.Add("id", "1")
	req1 = req1.WithContext(context.WithValue(req1.Context(), chi.RouteCtxKey, rctx1))
	ctx1 := auth.WithUser(req1.Context(), manager)
	req1 = req1.WithContext(ctx1)
	w1 := httptest.NewRecorder()

	handlers.UpdateEvent(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for first update, got %d: %s", w1.Code, w1.Body.String())
	}

	updateReq2 := `{"title": "Second Update", "timezone": "America/Los_Angeles", "version": 1}`
	req2 := httptest.NewRequest(http.MethodPut, "/api/events/1", bytes.NewReader([]byte(updateReq2)))
	req2.Header.Set("Content-Type", "application/json")
	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("id", "1")
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))
	ctx2 := auth.WithUser(req2.Context(), manager)
	req2 = req2.WithContext(ctx2)
	w2 := httptest.NewRecorder()

	handlers.UpdateEvent(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("Expected status 409 for version conflict, got %d: %s", w2.Code, w2.Body.String())
	}
}
