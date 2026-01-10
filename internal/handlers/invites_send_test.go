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
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockSendInviteService struct {
	getInviteByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
	sendInviteFunc    func(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error
}

func (m *mockSendInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getInviteByIDFunc != nil {
		return m.getInviteByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSendInviteService) SendInvite(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
	if m.sendInviteFunc != nil {
		return m.sendInviteFunc(ctx, req, emailRepo)
	}
	return errors.New("not implemented")
}

type mockSendInviteEventRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockSendInviteEventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) Create(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) Update(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSendInviteEventRepo) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

type mockSendInviteEmailRepo struct {
	createFunc func(ctx context.Context, email *models.EmailQueue) error
}

func (m *mockSendInviteEmailRepo) Create(ctx context.Context, email *models.EmailQueue) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, email)
	}
	return nil
}

func (m *mockSendInviteEmailRepo) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockSendInviteEmailRepo) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockSendInviteEmailRepo) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockSendInviteEmailRepo) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockSendInviteEmailRepo) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	return nil
}

func (m *mockSendInviteEmailRepo) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockSendInviteEmailRepo) MarkSending(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSendInviteEmailRepo) MarkSent(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSendInviteEmailRepo) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockSendInviteEmailRepo) MarkCancelled(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSendInviteEmailRepo) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	return nil
}

func (m *mockSendInviteEmailRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSendInviteEmailRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *mockSendInviteEmailRepo) GetStats(ctx context.Context) (*repositories.EmailQueueStats, error) {
	return &repositories.EmailQueueStats{}, nil
}

func TestSendInvite_Success(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"
	
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
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

	service := &mockSendInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			if id == 1 {
				return invite, nil
			}
			return nil, &models.NotFoundError{Resource: "invite"}
		},
		sendInviteFunc: func(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
			return nil
		},
	}

	eventRepo := &mockSendInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			if id == 100 {
				return event, nil
			}
			return nil, &models.NotFoundError{Resource: "event"}
		},
	}

	emailRepo := &mockSendInviteEmailRepo{}

	handler := NewSendInviteHandlers(service, eventRepo, emailRepo, "https://rsvp.example.com")

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/invites/1/send", nil)
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.SendInvite(w, req)

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

func TestSendInvite_Unauthorized(t *testing.T) {
	service := &mockSendInviteService{}
	eventRepo := &mockSendInviteEventRepo{}
	emailRepo := &mockSendInviteEmailRepo{}
	handler := NewSendInviteHandlers(service, eventRepo, emailRepo, "https://rsvp.example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/invites/1/send", nil)
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.SendInvite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSendInvite_InvalidID(t *testing.T) {
	service := &mockSendInviteService{}
	eventRepo := &mockSendInviteEventRepo{}
	emailRepo := &mockSendInviteEmailRepo{}
	handler := NewSendInviteHandlers(service, eventRepo, emailRepo, "https://rsvp.example.com")

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/invites/invalid/send", nil)
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.SendInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSendInvite_NotFound(t *testing.T) {
	service := &mockSendInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	eventRepo := &mockSendInviteEventRepo{}
	emailRepo := &mockSendInviteEmailRepo{}
	handler := NewSendInviteHandlers(service, eventRepo, emailRepo, "https://rsvp.example.com")

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/invites/999/send", nil)
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.SendInvite(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSendInvite_PermissionDenied(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"
	
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
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

	service := &mockSendInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	eventRepo := &mockSendInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	emailRepo := &mockSendInviteEmailRepo{}
	handler := NewSendInviteHandlers(service, eventRepo, emailRepo, "https://rsvp.example.com")

	user := &models.User{
		ID:    1,
		Email: "user@example.com",
		Role:  models.RoleEventManager,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/invites/1/send", nil)
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.SendInvite(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestSendInvite_NoEmail(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	
	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       nil,
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

	service := &mockSendInviteService{
		getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		sendInviteFunc: func(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
			return errors.New("invite has no email address")
		},
	}

	eventRepo := &mockSendInviteEventRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	emailRepo := &mockSendInviteEmailRepo{}
	handler := NewSendInviteHandlers(service, eventRepo, emailRepo, "https://rsvp.example.com")

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/invites/1/send", nil)
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inviteId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.SendInvite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}


func (m *mockSendInviteEventRepo) CountEvents(ctx context.Context) (int, error) {
	return 0, errors.New("not implemented")
}
