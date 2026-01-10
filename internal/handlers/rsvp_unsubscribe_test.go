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

func TestUnsubscribeHandler_Success(t *testing.T) {
	unsubscribeCalled := false
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			name := "Test User"
			return &models.Invite{
				ID:           1,
				EventID:      1,
				Email:        &email,
				Name:         &name,
				Status:       models.InviteStatusSent,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
		unsubscribeFunc: func(ctx context.Context, token string) error {
			unsubscribeCalled = true
			return nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusPublished,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/validtoken123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !unsubscribeCalled {
		t.Error("Expected UnsubscribeFromReminders to be called")
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}
}

func TestUnsubscribeHandler_InvalidToken(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, &models.NotFoundError{
				Resource: "invite",
				ID:       "token",
			}
		},
	}

	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/invalidtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_ExpiredToken(t *testing.T) {
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
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/expiredtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_RevokedInvite(t *testing.T) {
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
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/revokedtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_AlreadyUnsubscribed(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			return &models.Invite{
				ID:           1,
				EventID:      1,
				Email:        &email,
				Status:       models.InviteStatusSent,
				Unsubscribed: true,
				ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
		unsubscribeFunc: func(ctx context.Context, token string) error {
			return nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusPublished,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_UnsubscribeError(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			return &models.Invite{
				ID:           1,
				EventID:      1,
				Email:        &email,
				Status:       models.InviteStatusSent,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
		unsubscribeFunc: func(ctx context.Context, token string) error {
			return errors.New("database error")
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusPublished,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_EmptyToken(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{}
	mockEventRepo := &mockRSVPEventRepository{}
	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_EventNotFound(t *testing.T) {
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
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_WithTemplate(t *testing.T) {
	mockInviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			email := "test@example.com"
			name := "Test User"
			return &models.Invite{
				ID:           1,
				EventID:      1,
				Email:        &email,
				Name:         &name,
				Status:       models.InviteStatusSent,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
			}, nil
		},
		unsubscribeFunc: func(ctx context.Context, token string) error {
			return nil
		},
	}

	mockEventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusPublished,
			}, nil
		},
	}

	mockRSVPRepo := &mockRSVPRSVPRepository{}
	mockQuestionRepo := &mockRSVPQuestionRepository{}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	tmpl := template.Must(template.New("unsubscribe.html").Parse(`Unsubscribed from {{.Event.Title}}`))
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/unsubscribe/{token}", handler.Unsubscribe)

	req := httptest.NewRequest("GET", "/unsubscribe/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Event") {
		t.Errorf("Expected response to contain event title, got: %s", body)
	}
}
