package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestDeleteInvite_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
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

	mockSvc := services.NewMockInviteService(ctrl)
	mockEventRepo := repositories.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)
	mockSvc.EXPECT().DeleteInvite(gomock.Any(), int64(1)).Return(nil)

	handler := NewDeleteInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.DeleteInvite(w, req)

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

func TestDeleteInvite_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := services.NewMockInviteService(ctrl)
	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	handler := NewDeleteInviteHandlers(mockSvc, mockEventRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DeleteInvite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteInvite_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := services.NewMockInviteService(ctrl)
	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	handler := NewDeleteInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/invalid", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.DeleteInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteInvite_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := services.NewMockInviteService(ctrl)
	mockEventRepo := repositories.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(999)).Return(nil, &models.NotFoundError{Resource: "invite"})

	handler := NewDeleteInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/999", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.DeleteInvite(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteInvite_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{ID: 100, Title: "Test Event", CreatedBy: 999, Status: models.EventStatusDraft}

	mockSvc := services.NewMockInviteService(ctrl)
	mockEventRepo := repositories.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)

	handler := NewDeleteInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "user@example.com", Role: models.RoleEventManager}

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.DeleteInvite(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteInvite_CannotDeleteRespondedInvite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusResponded,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	event := &models.Event{ID: 100, Title: "Test Event", CreatedBy: 1, Status: models.EventStatusDraft}

	mockSvc := services.NewMockInviteService(ctrl)
	mockEventRepo := repositories.NewMockEventRepository(ctrl)

	mockSvc.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), int64(100)).Return(event, nil)
	mockSvc.EXPECT().DeleteInvite(gomock.Any(), int64(1)).Return(errors.New("cannot delete responded invite"))

	handler := NewDeleteInviteHandlers(mockSvc, mockEventRepo)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/1", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.DeleteInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
