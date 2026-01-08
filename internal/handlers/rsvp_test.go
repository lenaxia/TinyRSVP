package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockRSVPInviteService struct {
	getInviteByTokenFunc func(ctx context.Context, token string) (*models.Invite, error)
	markViewedFunc       func(ctx context.Context, inviteID int64) error
}

func (m *mockRSVPInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	if m.getInviteByTokenFunc != nil {
		return m.getInviteByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRSVPInviteService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	if m.markViewedFunc != nil {
		return m.markViewedFunc(ctx, inviteID)
	}
	return nil
}

type mockRSVPEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockRSVPEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRSVPEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRSVPEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRSVPEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockRSVPEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockRSVPEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRSVPEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, nil
}

type mockRSVPRSVPRepository struct {
	getByInviteIDFunc func(ctx context.Context, inviteID int64) (*models.RSVP, error)
}

func (m *mockRSVPRSVPRepository) Create(ctx context.Context, rsvp *models.RSVP) error {
	return nil
}

func (m *mockRSVPRSVPRepository) GetByID(ctx context.Context, id int64) (*models.RSVP, error) {
	return nil, nil
}

func (m *mockRSVPRSVPRepository) GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error) {
	if m.getByInviteIDFunc != nil {
		return m.getByInviteIDFunc(ctx, inviteID)
	}
	return nil, &models.NotFoundError{Resource: "rsvp"}
}

func (m *mockRSVPRSVPRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error) {
	return nil, nil
}

func (m *mockRSVPRSVPRepository) Update(ctx context.Context, rsvp *models.RSVP) error {
	return nil
}

func (m *mockRSVPRSVPRepository) GetStats(ctx context.Context, eventID int64) (*repositories.RSVPStats, error) {
	return nil, nil
}

type mockRSVPQuestionRepository struct {
	getByEventIDFunc func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
}

func (m *mockRSVPQuestionRepository) Create(ctx context.Context, question *models.PreferenceQuestion) error {
	return nil
}

func (m *mockRSVPQuestionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	return nil, nil
}

func (m *mockRSVPQuestionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	if m.getByEventIDFunc != nil {
		return m.getByEventIDFunc(ctx, eventID)
	}
	return []*models.PreferenceQuestion{}, nil
}

func (m *mockRSVPQuestionRepository) Update(ctx context.Context, question *models.PreferenceQuestion) error {
	return nil
}

func (m *mockRSVPQuestionRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRSVPQuestionRepository) UpdateDisplayOrder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return nil
}

func (m *mockRSVPQuestionRepository) Reorder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return nil
}

func TestRSVPHandler_GetRSVPPage_ValidToken(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	rsvpDeadline := startTime.Add(-24 * time.Hour)

	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			name := "Test User"
			return &models.Invite{
				ID:          1,
				EventID:     1,
				Email:       &email,
				Name:        &name,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusSent,
				ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
		markViewedFunc: func(ctx context.Context, inviteID int64) error {
			return nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			desc := "Test event description"
			loc := "Test Location"
			return &models.Event{
				ID:           1,
				Title:        "Test Event",
				Description:  &desc,
				StartTime:    startTime,
				EndTime:      &endTime,
				Timezone:     "America/Los_Angeles",
				Location:     &loc,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &rsvpDeadline,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp"}
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}
}

func TestRSVPHandler_GetRSVPPage_InvalidToken(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invite not found")
		},
	}

	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/invalidtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_ExpiredToken(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invite has expired")
		},
	}

	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/expiredtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_RevokedInvite(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invite has been revoked")
		},
	}

	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/revokedtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_CancelledEvent(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			return &models.Invite{
				ID:        1,
				EventID:   1,
				Email:     &email,
				Status:    models.InviteStatusSent,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:     1,
				Title:  "Cancelled Event",
				Status: models.EventStatusCancelled,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_ArchivedEvent(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			return &models.Invite{
				ID:        1,
				EventID:   1,
				Email:     &email,
				Status:    models.InviteStatusSent,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:     1,
				Title:  "Archived Event",
				Status: models.EventStatusArchived,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_WithExistingRSVP(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)

	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			return &models.Invite{
				ID:        1,
				EventID:   1,
				Email:     &email,
				Status:    models.InviteStatusResponded,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
		markViewedFunc: func(ctx context.Context, inviteID int64) error {
			return nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusPublished,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return &models.RSVP{
				ID:       1,
				InviteID: 1,
				Response: models.RSVPResponseYes,
				PlusOnes: 2,
			}, nil
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_EventNotFound(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			return &models.Invite{
				ID:        1,
				EventID:   999,
				Email:     &email,
				Status:    models.InviteStatusSent,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return nil, &models.NotFoundError{Resource: "event"}
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_EmptyToken(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{}
	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

