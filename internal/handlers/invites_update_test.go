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
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockUpdateInviteService struct {
	getInviteByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
	updateInviteFunc  func(ctx context.Context, req *invites.UpdateInviteRequest) error
}

func (m *mockUpdateInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getInviteByIDFunc != nil {
		return m.getInviteByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUpdateInviteService) UpdateInvite(ctx context.Context, req *invites.UpdateInviteRequest) error {
	if m.updateInviteFunc != nil {
		return m.updateInviteFunc(ctx, req)
	}
	return errors.New("not implemented")
}

type mockUpdateInviteEventRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockUpdateInviteEventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) Create(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) Update(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func TestUpdateInvite_Success(t *testing.T) {
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

	service := &mockUpdateInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			if id == 1 {
				return invite, nil
			}
			return nil, &models.NotFoundError{Resource: "invite"}
		},
		updateInviteFunc: func(ctx context.Context, req *invites.UpdateInviteRequest) error {
			return nil
		},
	}

	eventRepo := &mockUpdateInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			if id == 100 {
				return event, nil
			}
			return nil, &models.NotFoundError{Resource: "event"}
		},
	}

	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

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
	service := &mockUpdateInviteService{}
	eventRepo := &mockUpdateInviteEventRepo{}
	handler := NewUpdateInviteHandlers(service, eventRepo)

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
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
	service := &mockUpdateInviteService{}
	eventRepo := &mockUpdateInviteEventRepo{}
	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
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
	service := &mockUpdateInviteService{}
	eventRepo := &mockUpdateInviteEventRepo{}
	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

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
	service := &mockUpdateInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	eventRepo := &mockUpdateInviteEventRepo{}
	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
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

	service := &mockUpdateInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	eventRepo := &mockUpdateInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  models.RoleEventManager,
	}

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
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

	event := &models.Event{
		ID:        100,
		Title:     "Test Event",
		CreatedBy: 1,
		Status:    models.EventStatusDraft,
	}

	service := &mockUpdateInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		updateInviteFunc: func(ctx context.Context, req *invites.UpdateInviteRequest) error {
			return errors.New("cannot update responded invite")
		},
	}

	eventRepo := &mockUpdateInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
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

	event := &models.Event{
		ID:        100,
		Title:     "Test Event",
		CreatedBy: 1,
		Status:    models.EventStatusDraft,
	}

	service := &mockUpdateInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		updateInviteFunc: func(ctx context.Context, req *invites.UpdateInviteRequest) error {
			return errors.New("cannot update revoked invite")
		},
	}

	eventRepo := &mockUpdateInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	handler := NewUpdateInviteHandlers(service, eventRepo)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
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


func (m *mockUpdateInviteEventRepo) CountEvents(ctx context.Context) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *mockUpdateInviteEventRepo) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUpdateInviteEventRepo) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	return nil, errors.New("not implemented")
}

