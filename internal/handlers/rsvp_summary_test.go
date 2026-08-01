package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	mockrepos "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
	"go.uber.org/mock/gomock"
	"html/template"
)

func TestRSVPSummaryHandler_GetRSVPSummary_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 100,
		YesCount:     60,
		NoCount:      20,
		MaybeCount:   10,
		NoResponse:   10,
		TotalGuests:  75,
	}

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(event, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetStats(gomock.Any(), gomock.Any()).Return(stats, nil)
	mockRSVPRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.RSVP{}, nil).AnyTimes()

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil).AnyTimes()

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_InvalidEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/invalid/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_EventNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "event", ID: 1})

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 2,
	}

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(event, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_AdminCanAccessAnyEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 2,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 50,
		YesCount:     30,
		NoCount:      10,
		MaybeCount:   5,
		NoResponse:   5,
		TotalGuests:  40,
	}

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(event, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetStats(gomock.Any(), gomock.Any()).Return(stats, nil)
	mockRSVPRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.RSVP{}, nil).AnyTimes()

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil).AnyTimes()

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_WithQuestions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 50,
		YesCount:     30,
		NoCount:      10,
		MaybeCount:   5,
		NoResponse:   5,
		TotalGuests:  40,
	}

	options := "Vegetarian,Vegan,None"
	questions := []*models.PreferenceQuestion{
		{
			ID:           1,
			EventID:      1,
			QuestionText: "Dietary restrictions?",
			QuestionType: models.QuestionTypeSingleChoice,
			Options:      &options,
		},
	}

	rsvps := []*models.RSVP{
		{ID: 1, InviteID: 1, Response: models.RSVPResponseYes, PlusOnes: 1},
		{ID: 2, InviteID: 2, Response: models.RSVPResponseYes, PlusOnes: 0},
		{ID: 3, InviteID: 3, Response: models.RSVPResponseNo, PlusOnes: 0},
	}

	vegetarian := "Vegetarian"
	vegan := "Vegan"
	answers := []*models.RSVPAnswer{
		{ID: 1, RSVPID: 1, QuestionID: 1, AnswerText: &vegetarian},
		{ID: 2, RSVPID: 2, QuestionID: 1, AnswerText: &vegan},
	}

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(event, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetStats(gomock.Any(), gomock.Any()).Return(stats, nil)
	mockRSVPRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return(rsvps, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return(questions, nil)

	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)
	mockAnswerRepo.EXPECT().GetByRSVPID(gomock.Any(), gomock.Any()).Return(answers, nil).AnyTimes()

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_summary.html"))

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_GetRSVPSummary_StatsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(event, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetStats(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "stats"})

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

	r := chi.NewRouter()
	r.Get("/events/{id}/rsvp-summary", handler.GetRSVPSummary)

	req := httptest.NewRequest("GET", "/events/1/rsvp-summary", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRSVPSummaryHandler_SetTemplates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockAnswerRepo := mockrepos.NewMockAnswerRepository(ctrl)

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

	tmpl := template.New("test")
	handler.SetTemplates(tmpl)

	if handler.templates == nil {
		t.Error("Expected templates to be set")
	}
}
