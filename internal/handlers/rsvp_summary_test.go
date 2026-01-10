package handlers

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockRSVPSummaryEventRepository struct {
	event *models.Event
	err   error
}

func (m *mockRSVPSummaryEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.event, nil
}

func (m *mockRSVPSummaryEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRSVPSummaryEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRSVPSummaryEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRSVPSummaryEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPSummaryEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockRSVPSummaryEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockRSVPSummaryEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPSummaryEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockRSVPSummaryEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, nil
}

type mockRSVPSummaryRSVPRepository struct {
	stats *repositories.RSVPStats
	rsvps []*models.RSVP
	err   error
}

func (m *mockRSVPSummaryRSVPRepository) Create(ctx context.Context, rsvp *models.RSVP) error {
	return nil
}

func (m *mockRSVPSummaryRSVPRepository) GetByID(ctx context.Context, id int64) (*models.RSVP, error) {
	return nil, nil
}

func (m *mockRSVPSummaryRSVPRepository) GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error) {
	return nil, nil
}

func (m *mockRSVPSummaryRSVPRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rsvps, nil
}

func (m *mockRSVPSummaryRSVPRepository) Update(ctx context.Context, rsvp *models.RSVP) error {
	return nil
}

func (m *mockRSVPSummaryRSVPRepository) GetStats(ctx context.Context, eventID int64) (*repositories.RSVPStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func (m *mockRSVPSummaryRSVPRepository) GetByInviteIDs(ctx context.Context, inviteIDs []int64) ([]*models.RSVP, error) {
	return []*models.RSVP{}, nil
}

type mockRSVPSummaryQuestionRepository struct {
	questions []*models.PreferenceQuestion
	err       error
}

func (m *mockRSVPSummaryQuestionRepository) Create(ctx context.Context, question *models.PreferenceQuestion) error {
	return nil
}

func (m *mockRSVPSummaryQuestionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	return nil, nil
}

func (m *mockRSVPSummaryQuestionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.questions, nil
}

func (m *mockRSVPSummaryQuestionRepository) Update(ctx context.Context, question *models.PreferenceQuestion) error {
	return nil
}

func (m *mockRSVPSummaryQuestionRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRSVPSummaryQuestionRepository) List(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	return m.questions, m.err
}

func (m *mockRSVPSummaryQuestionRepository) Reorder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return nil
}

type mockRSVPSummaryAnswerRepository struct {
	answers []*models.RSVPAnswer
	err     error
}

func (m *mockRSVPSummaryAnswerRepository) Create(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockRSVPSummaryAnswerRepository) GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.answers, nil
}

func (m *mockRSVPSummaryAnswerRepository) Update(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockRSVPSummaryAnswerRepository) DeleteByRSVPID(ctx context.Context, rsvpID int64) error {
	return nil
}

func (m *mockRSVPSummaryAnswerRepository) GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error) {
	return nil, nil
}

func TestRSVPSummaryHandler_GetRSVPSummary_Success(t *testing.T) {
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

	mockEventRepo := &mockRSVPSummaryEventRepository{event: event}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{stats: stats}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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
	mockEventRepo := &mockRSVPSummaryEventRepository{}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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
	mockEventRepo := &mockRSVPSummaryEventRepository{}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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
	mockEventRepo := &mockRSVPSummaryEventRepository{
		err: &models.NotFoundError{Resource: "event", ID: 1},
	}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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
	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 2,
	}

	mockEventRepo := &mockRSVPSummaryEventRepository{event: event}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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

	mockEventRepo := &mockRSVPSummaryEventRepository{event: event}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{stats: stats}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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

	mockEventRepo := &mockRSVPSummaryEventRepository{event: event}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{stats: stats, rsvps: rsvps}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{questions: questions}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{answers: answers}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

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
	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	mockEventRepo := &mockRSVPSummaryEventRepository{event: event}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{err: errors.New("database error")}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

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
	mockEventRepo := &mockRSVPSummaryEventRepository{}
	mockRSVPRepo := &mockRSVPSummaryRSVPRepository{}
	mockQuestionRepo := &mockRSVPSummaryQuestionRepository{}
	mockAnswerRepo := &mockRSVPSummaryAnswerRepository{}

	handler := NewRSVPSummaryHandler(mockEventRepo, mockRSVPRepo, mockQuestionRepo, mockAnswerRepo)

	tmpl := template.New("test")
	handler.SetTemplates(tmpl)

	if handler.templates == nil {
		t.Error("Expected templates to be set")
	}
}

