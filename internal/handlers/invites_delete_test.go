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
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockDeleteInviteService struct {
	getInviteByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
	deleteInviteFunc  func(ctx context.Context, inviteID int64) error
}

func (m *mockDeleteInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getInviteByIDFunc != nil {
		return m.getInviteByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDeleteInviteService) DeleteInvite(ctx context.Context, inviteID int64) error {
	if m.deleteInviteFunc != nil {
		return m.deleteInviteFunc(ctx, inviteID)
	}
	return errors.New("not implemented")
}

type mockDeleteInviteEventRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockDeleteInviteEventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) Create(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) Update(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockDeleteInviteEventRepo) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func TestDeleteInvite_Success(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
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

	service := &mockDeleteInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			if id == 1 {
				return invite, nil
			}
			return nil, &models.NotFoundError{Resource: "invite"}
		},
		deleteInviteFunc: func(ctx context.Context, inviteID int64) error {
			return nil
		},
	}

	eventRepo := &mockDeleteInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			if id == 100 {
				return event, nil
			}
			return nil, &models.NotFoundError{Resource: "event"}
		},
	}

	handler := NewDeleteInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

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
	service := &mockDeleteInviteService{}
	eventRepo := &mockDeleteInviteEventRepo{}
	handler := NewDeleteInviteHandlers(service, eventRepo)

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
	service := &mockDeleteInviteService{}
	eventRepo := &mockDeleteInviteEventRepo{}
	handler := NewDeleteInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

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
	service := &mockDeleteInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	eventRepo := &mockDeleteInviteEventRepo{}
	handler := NewDeleteInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

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
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
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
		CreatedBy: 999,
		Status:    models.EventStatusDraft,
	}

	service := &mockDeleteInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	eventRepo := &mockDeleteInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	handler := NewDeleteInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  models.RoleEventManager,
	}

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
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       stringPtr("test@example.com"),
		Name:        stringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusResponded,
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

	service := &mockDeleteInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		deleteInviteFunc: func(ctx context.Context, inviteID int64) error {
			return errors.New("cannot delete responded invite")
		},
	}

	eventRepo := &mockDeleteInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	handler := NewDeleteInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

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

func (m *mockDeleteInviteEventRepo) CountEvents(ctx context.Context) (int, error) {
	return 0, errors.New("not implemented")
}
