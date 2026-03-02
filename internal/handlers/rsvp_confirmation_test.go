package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/models"
	mockrepos "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
	mocksvcs "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestRSVPHandler_GetConfirmationPage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	rsvpDeadline := startTime.Add(-1 * time.Hour)
	email := "test@example.com"
	name := "Test User"
	desc := "Test event description"
	loc := "Test Location"
	textAnswer := "Vegetarian"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:          1,
		EventID:     1,
		Email:       &email,
		Name:        &name,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusResponded,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:           1,
		Title:        "Test Event",
		Description:  &desc,
		StartTime:    startTime,
		Timezone:     "America/Los_Angeles",
		Location:     &loc,
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:        1,
		InviteID:  1,
		Response:  models.RSVPResponseYes,
		PlusOnes:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{
		{
			ID:           1,
			EventID:      1,
			QuestionText: "Dietary restrictions?",
			QuestionType: models.QuestionTypeText,
			Required:     true,
		},
	}, nil)

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)
	mockAnswerRepo.EXPECT().GetByRSVPID(gomock.Any(), gomock.Any()).Return([]*models.RSVPAnswer{
		{
			ID:         1,
			RSVPID:     1,
			QuestionID: 1,
			AnswerText: &textAnswer,
		},
	}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Status:    models.InviteStatusSent,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:     1,
		Title:  "Test Event",
		Status: models.EventStatusPublished,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "rsvp"})

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "invite"})

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"
	name := "Test User"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Name:      &name,
		Status:    models.InviteStatusResponded,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:        1,
		Title:     "Birthday Party",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)
	mockAnswerRepo.EXPECT().GetByRSVPID(gomock.Any(), gomock.Any()).Return([]*models.RSVPAnswer{}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Status:    models.InviteStatusResponded,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:     1,
		Title:  "Cancelled Event",
		Status: models.EventStatusCancelled,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	futureDeadline := time.Now().Add(12 * time.Hour)
	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Status:    models.InviteStatusResponded,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:           1,
		Title:        "Test Event",
		StartTime:    startTime,
		Timezone:     "America/Los_Angeles",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &futureDeadline,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 0,
	}, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)
	mockAnswerRepo.EXPECT().GetByRSVPID(gomock.Any(), gomock.Any()).Return([]*models.RSVPAnswer{}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	pastDeadline := time.Now().Add(-1 * time.Hour)
	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Status:    models.InviteStatusResponded,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:           1,
		Title:        "Test Event",
		StartTime:    startTime,
		Timezone:     "America/Los_Angeles",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &pastDeadline,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 0,
	}, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)
	mockAnswerRepo.EXPECT().GetByRSVPID(gomock.Any(), gomock.Any()).Return([]*models.RSVPAnswer{}, nil)

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
