package handlers

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockAnswerRepository struct {
	getByRSVPIDFunc func(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error)
}

func (m *mockAnswerRepository) Create(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockAnswerRepository) GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
	if m.getByRSVPIDFunc != nil {
		return m.getByRSVPIDFunc(ctx, rsvpID)
	}
	return []*models.RSVPAnswer{}, nil
}

func (m *mockAnswerRepository) GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error) {
	return nil, nil
}

func (m *mockAnswerRepository) Update(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockAnswerRepository) DeleteByRSVPID(ctx context.Context, rsvpID int64) error {
	return nil
}

func TestRSVPHandler_GetConfirmationPage_Success(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	rsvpDeadline := startTime.Add(-1 * time.Hour)

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
				Status:      models.InviteStatusResponded,
				ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			}, nil
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
				Timezone:     "America/Los_Angeles",
				Location:     &loc,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &rsvpDeadline,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return &models.RSVP{
				ID:        1,
				InviteID:  1,
				Response:  models.RSVPResponseYes,
				PlusOnes:  2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{
				{
					ID:           1,
					EventID:      1,
					QuestionText: "Dietary restrictions?",
					QuestionType: models.QuestionTypeText,
					Required:     true,
				},
			}, nil
		},
	}

	textAnswer := "Vegetarian"
	mockAnswerRepo := &mockAnswerRepository{
		getByRSVPIDFunc: func(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
			return []*models.RSVPAnswer{
				{
					ID:         1,
					RSVPID:     1,
					QuestionID: 1,
					AnswerText: &textAnswer,
				},
			}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestRSVPHandler_GetConfirmationPage_NoRSVP(t *testing.T) {
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
				Title:  "Test Event",
				Status: models.EventStatusPublished,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp"}
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{}
	mockAnswerRepo := &mockAnswerRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRSVPHandler_GetConfirmationPage_InvalidToken(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invite not found")
		},
	}

	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}
	mockAnswerRepo := &mockAnswerRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRSVPHandler_GetConfirmationPage_WithTemplate(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)

	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			name := "Test User"
			return &models.Invite{
				ID:        1,
				EventID:   1,
				Email:     &email,
				Name:      &name,
				Status:    models.InviteStatusResponded,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Birthday Party",
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
				PlusOnes: 1,
			}, nil
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	mockAnswerRepo := &mockAnswerRepository{
		getByRSVPIDFunc: func(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
			return []*models.RSVPAnswer{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	tmpl := template.Must(template.New("confirmation.html").Parse(`Event:{{.Event.Title}}|Response:{{.RSVP.Response}}|PlusOnes:{{.RSVP.PlusOnes}}|CanUpdate:{{.CanUpdate}}`))
	handler.SetConfirmationTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Event:Birthday Party") {
		t.Error("Expected response to contain event title")
	}
	if !strings.Contains(body, "Response:yes") {
		t.Error("Expected response to contain RSVP response")
	}
	if !strings.Contains(body, "PlusOnes:1") {
		t.Error("Expected response to contain plus ones count")
	}
}

func TestRSVPHandler_GetConfirmationPage_CancelledEvent(t *testing.T) {
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
	mockAnswerRepo := &mockAnswerRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}
}

func TestRSVPHandler_GetConfirmationPage_CanUpdateTrue(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	futureDeadline := time.Now().Add(12 * time.Hour)

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
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Title:        "Test Event",
				StartTime:    startTime,
				Timezone:     "America/Los_Angeles",
				Status:       models.EventStatusPublished,
				RSVPDeadline: &futureDeadline,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return &models.RSVP{
				ID:       1,
				InviteID: 1,
				Response: models.RSVPResponseYes,
				PlusOnes: 0,
			}, nil
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	mockAnswerRepo := &mockAnswerRepository{
		getByRSVPIDFunc: func(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
			return []*models.RSVPAnswer{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	tmpl := template.Must(template.New("confirmation.html").Parse(`CanUpdate:{{.CanUpdate}}`))
	handler.SetConfirmationTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "CanUpdate:true") {
		t.Errorf("Expected CanUpdate to be true, got: %s", body)
	}
}

func TestRSVPHandler_GetConfirmationPage_CanUpdateFalse_DeadlinePassed(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	pastDeadline := time.Now().Add(-1 * time.Hour)

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
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Title:        "Test Event",
				StartTime:    startTime,
				Timezone:     "America/Los_Angeles",
				Status:       models.EventStatusPublished,
				RSVPDeadline: &pastDeadline,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return &models.RSVP{
				ID:       1,
				InviteID: 1,
				Response: models.RSVPResponseYes,
				PlusOnes: 0,
			}, nil
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	mockAnswerRepo := &mockAnswerRepository{
		getByRSVPIDFunc: func(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
			return []*models.RSVPAnswer{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	tmpl := template.Must(template.New("confirmation.html").Parse(`CanUpdate:{{.CanUpdate}}`))
	handler.SetConfirmationTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "CanUpdate:false") {
		t.Errorf("Expected CanUpdate to be false, got: %s", body)
	}
}
