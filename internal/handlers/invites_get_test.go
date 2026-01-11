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

type mockGetInviteService struct {
	getInviteByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
}

func (m *mockGetInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getInviteByIDFunc != nil {
		return m.getInviteByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

type mockGetInviteEventRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockGetInviteEventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) Create(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) Update(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func TestGetInvite_Success(t *testing.T) {
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

	event := &models.Event{
		ID:        100,
		Title:     "Test Event",
		CreatedBy: 1,
		Status:    models.EventStatusDraft,
	}

	service := &mockGetInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			if id == 1 {
				return invite, nil
			}
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	eventRepo := &mockGetInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			if id == 100 {
				return event, nil
			}
			return nil, &models.NotFoundError{Resource: "event"}
		},
	}

	handler := NewGetInviteHandlers(service, eventRepo)

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
	service := &mockGetInviteService{}
	eventRepo := &mockGetInviteEventRepo{}
	handler := NewGetInviteHandlers(service, eventRepo)

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
	service := &mockGetInviteService{}
	eventRepo := &mockGetInviteEventRepo{}
	handler := NewGetInviteHandlers(service, eventRepo)

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
	service := &mockGetInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	eventRepo := &mockGetInviteEventRepo{}
	handler := NewGetInviteHandlers(service, eventRepo)

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

	event := &models.Event{
		ID:        100,
		Title:     "Test Event",
		CreatedBy: 999,
		Status:    models.EventStatusDraft,
	}

	service := &mockGetInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	eventRepo := &mockGetInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	handler := NewGetInviteHandlers(service, eventRepo)

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

func (m *mockGetInviteEventRepo) CountEvents(ctx context.Context) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *mockGetInviteEventRepo) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGetInviteEventRepo) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	return nil, errors.New("not implemented")
}

