package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetInvite_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{
		ID:        100,
		Title:     "Test Event",
		CreatedBy: 1,
		Status:    models.EventStatusDraft,
	}

	// Create mocks
	mockService := mocks.NewMockInviteService(ctrl)
	mockEventRepo := mocks.NewMockEventRepository(ctrl)

	// Set expectations
	mockService.EXPECT().
		GetInviteByID(gomock.Any(), int64(1)).
		Return(invite, nil)

	mockEventRepo.EXPECT().
		GetByID(gomock.Any(), int64(100)).
		Return(event, nil)

	handler := NewGetInviteHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetInvite(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["id"] != float64(1) {
		t.Errorf("Expected invite ID 1, got %v", response["id"])
	}

	if response["event_id"] != float64(100) {
		t.Errorf("Expected event ID 100, got %v", response["event_id"])
	}
}

func TestGetInvite_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockInviteService(ctrl)
	mockEventRepo := mocks.NewMockEventRepository(ctrl)
	handler := NewGetInviteHandlers(mockService, mockEventRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetInvite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetInvite_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockInviteService(ctrl)
	mockEventRepo := mocks.NewMockEventRepository(ctrl)
	handler := NewGetInviteHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/invites/invalid", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetInvite_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockInviteService(ctrl)
	mockEventRepo := mocks.NewMockEventRepository(ctrl)

	// Set expectation
	mockService.EXPECT().
		GetInviteByID(gomock.Any(), int64(999)).
		Return(nil, &models.NotFoundError{Resource: "invite"})

	handler := NewGetInviteHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/invites/999", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetInvite(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetInvite_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{
		ID:        100,
		Title:     "Test Event",
		CreatedBy: 999,
		Status:    models.EventStatusDraft,
	}

	mockService := mocks.NewMockInviteService(ctrl)
	mockEventRepo := mocks.NewMockEventRepository(ctrl)

	// Set expectations
	mockService.EXPECT().
		GetInviteByID(gomock.Any(), int64(1)).
		Return(invite, nil)

	mockEventRepo.EXPECT().
		GetByID(gomock.Any(), int64(100)).
		Return(event, nil)

	handler := NewGetInviteHandlers(mockService, mockEventRepo)

	user := &models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  models.RoleEventManager,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetInvite(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}
