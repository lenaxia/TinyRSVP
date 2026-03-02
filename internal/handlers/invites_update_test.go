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
	mockrepos "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
	mocksvcs "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestUpdateInvite_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{ID: 100, Title: "Test Event", CreatedBy: 1, Status: models.EventStatusDraft}

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)
	mockSvc.EXPECT().UpdateInvite(gomock.Any(), gomock.Any()).Return(nil)

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	reqBody := map[string]interface{}{"name": "Updated Name", "max_plus_ones": 3}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] == nil {
		t.Error("Expected success message in response")
	}
}

func TestUpdateInvite_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	reqBody := map[string]interface{}{"name": "Updated Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateInvite_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	reqBody := map[string]interface{}{"name": "Updated Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/invalid", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateInvite_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateInvite_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(999)).Return(nil, &models.NotFoundError{Resource: "invite"})

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	reqBody := map[string]interface{}{"name": "Updated Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/999", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateInvite_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	event := &models.Event{ID: 100, Title: "Test Event", CreatedBy: 999, Status: models.EventStatusDraft}

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "user@example.com", Role: models.RoleEventManager}

	reqBody := map[string]interface{}{"name": "Updated Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUpdateInvite_CannotUpdateRespondedInvite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusResponded,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	event := &models.Event{ID: 100, Title: "Test Event", CreatedBy: 1, Status: models.EventStatusDraft}

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)
	mockSvc.EXPECT().UpdateInvite(gomock.Any(), gomock.Any()).Return(errors.New("cannot update responded invite"))

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	reqBody := map[string]interface{}{"name": "Updated Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateInvite_CannotUpdateRevokedInvite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusRevoked,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	event := &models.Event{ID: 100, Title: "Test Event", CreatedBy: 1, Status: models.EventStatusDraft}

	mockSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)
	mockSvc.EXPECT().UpdateInvite(gomock.Any(), gomock.Any()).Return(errors.New("cannot update revoked invite"))

	handler := NewUpdateInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	reqBody := map[string]interface{}{"name": "Updated Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/invites/1", bytes.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.UpdateInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// unused import guard
var _ *invites.UpdateInviteRequest
