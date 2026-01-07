package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockIndividualInviteService struct {
	createIndividualInviteFunc func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error)
}

func (m *mockIndividualInviteService) CreateIndividualInvite(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
	if m.createIndividualInviteFunc != nil {
		return m.createIndividualInviteFunc(ctx, user, req)
	}
	return nil, nil
}

func TestInviteHandlers_CreateInvite_Success(t *testing.T) {
	email := "guest@example.com"
	name := "John Doe"
	maxPlusOnes := 2

	mockService := &mockIndividualInviteService{
		createIndividualInviteFunc: func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
			return &invites.CreateIndividualInviteResponse{
				Invite: &models.Invite{
					ID:          123,
					EventID:     1,
					Email:       &email,
					Name:        &name,
					MaxPlusOnes: maxPlusOnes,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   time.Now().Add(60 * 24 * time.Hour),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Token: "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p",
			}, nil
		},
	}

	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email":         "guest@example.com",
		"name":          "John Doe",
		"max_plus_ones": 2,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{
		ID:   100,
		Role: models.RoleEventManager,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["token"] != "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p" {
		t.Errorf("Expected token in response, got %v", response["token"])
	}

	if response["rsvp_url"] == nil {
		t.Error("Expected rsvp_url in response")
	}

	invite, ok := response["invite"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected invite object in response")
	}

	if invite["email"] != "guest@example.com" {
		t.Errorf("Expected email 'guest@example.com', got %v", invite["email"])
	}
}

func TestInviteHandlers_CreateInvite_InvalidJSON(t *testing.T) {
	mockService := &mockIndividualInviteService{}
	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_MissingEmail(t *testing.T) {
	mockService := &mockIndividualInviteService{}
	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"name": "John Doe",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_Unauthorized(t *testing.T) {
	mockService := &mockIndividualInviteService{}
	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_InvalidEventID(t *testing.T) {
	mockService := &mockIndividualInviteService{}
	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/invalid/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_ServiceError_NotFound(t *testing.T) {
	mockService := &mockIndividualInviteService{
		createIndividualInviteFunc: func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
			return nil, &models.NotFoundError{Resource: "Event", ID: 999}
		},
	}

	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/999/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_ServiceError_PermissionDenied(t *testing.T) {
	mockService := &mockIndividualInviteService{
		createIndividualInviteFunc: func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
			return nil, &models.PermissionDeniedError{Action: "create invite", Resource: "Event"}
		},
	}

	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_ServiceError_Conflict(t *testing.T) {
	mockService := &mockIndividualInviteService{
		createIndividualInviteFunc: func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
			return nil, &models.ConflictError{Resource: "Invite", Field: "email", Value: "guest@example.com"}
		},
	}

	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_ServiceError_Validation(t *testing.T) {
	mockService := &mockIndividualInviteService{
		createIndividualInviteFunc: func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
			return nil, &models.ValidationError{Field: "email", Message: "invalid email format"}
		},
	}

	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "notanemail",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteHandlers_CreateInvite_ServiceError_Internal(t *testing.T) {
	mockService := &mockIndividualInviteService{
		createIndividualInviteFunc: func(ctx context.Context, user *models.User, req *invites.CreateIndividualInviteRequest) (*invites.CreateIndividualInviteResponse, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewInviteHandlers(mockService, "https://rsvp.example.com")

	reqBody := map[string]interface{}{
		"email": "guest@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateInvite(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
