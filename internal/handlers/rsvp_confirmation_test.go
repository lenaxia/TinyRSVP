package handlers

import (
	"fmt"
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
	handler.SetConfirmationTemplates(testTemplate(t, "confirmation.html"))

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
	handler.SetConfirmationTemplates(testTemplate(t, "confirmation.html"))

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
	handler.SetConfirmationTemplates(testTemplate(t, "confirmation.html"))

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
	handler.SetConfirmationTemplates(testTemplate(t, "confirmation.html"))

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

// TestRSVPHandler_ConfirmationPage_ErrorPath_NilEvent verifies that when the
// confirmation page is rendered with an error message (Event/RSVP are nil),
// the template does not panic and returns the correct error status code.
// This is a regression test for the bug where {{.Event.Title}} in the title
// block caused a nil-pointer template execution failure, which (due to headers
// already being written) was silently swallowed and produced a 500.
func TestRSVPHandler_ConfirmationPage_ErrorPath_NilEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "invite"})

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_page.html"))

	// Use a template that mimics the real one: accesses .Event.Title in title
	// block and branches on .ErrorMessage in content — should not panic when
	// Event is nil.
	tmpl := template.Must(template.New("confirmation.html").Parse(
		`{{if .Event}}Title:{{.Event.Title}}{{else}}Error:{{.ErrorMessage}}{{end}}`,
	))
	handler.SetConfirmationTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/badtoken/confirmation", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should be a 4xx (not found), definitely not 500.
	if w.Code == http.StatusInternalServerError {
		t.Errorf("renderConfirmationPage returned 500 on error path — nil-event bug not fixed. Body: %s", w.Body.String())
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for bad token, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Invite not found") {
		t.Errorf("Expected error message in body, got: %s", w.Body.String())
	}
}

// TestRSVPHandler_ConfirmationPage_TemplateExecError_Returns500 verifies that
// when template execution fails (e.g. calls an undefined function), the handler
// returns a proper 500 with an error message — not a partial response with the
// wrong status. This is a regression test for the buffered-write fix.
func TestRSVPHandler_ConfirmationPage_TemplateExecError_Returns500(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	email := "test@example.com"
	name := "Test"
	startTime := time.Now().Add(24 * time.Hour)

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID: 1, EventID: 1, Email: &email, Name: &name,
		Status: models.InviteStatusResponded, ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID: 1, Title: "Test Event", StartTime: startTime,
		Timezone: "UTC", Status: models.EventStatusPublished,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID: 1, InviteID: 1, Response: models.RSVPResponseYes,
	}, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)
	mockAnswerRepo.EXPECT().GetByRSVPID(gomock.Any(), gomock.Any()).Return([]*models.RSVPAnswer{}, nil)

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetAnswerRepository(mockAnswerRepo)

	// Template that calls an undefined function — will fail at Execute time.
	tmpl := template.Must(template.New("confirmation.html").Funcs(template.FuncMap{
		"willFail": func() (string, error) { return "", fmt.Errorf("intentional template failure") },
	}).Parse(`{{willFail}}`))
	handler.SetConfirmationTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}/confirmation", handler.GetConfirmationPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken/confirmation", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// With the buffered-write fix, we must get a clean 500 — not 200 with a
	// partial body.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 on template execution failure, got %d. Body: %s", w.Code, w.Body.String())
	}
}
