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
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
)

type mockRSVPInviteService struct {
	getInviteByTokenFunc     func(ctx context.Context, token string) (*models.Invite, error)
	markViewedFunc           func(ctx context.Context, inviteID int64) error
	unsubscribeFunc          func(ctx context.Context, token string) error
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

func (m *mockRSVPInviteService) UnsubscribeFromReminders(ctx context.Context, token string) error {
	if m.unsubscribeFunc != nil {
		return m.unsubscribeFunc(ctx, token)
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

func (m *mockRSVPEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
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

func (m *mockRSVPRSVPRepository) GetByInviteIDs(ctx context.Context, inviteIDs []int64) ([]*models.RSVP, error) {
	return []*models.RSVP{}, nil
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

type mockRSVPService struct {
	submitRSVPFunc func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
	updateRSVPFunc func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
}

func (m *mockRSVPService) SubmitRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
	if m.submitRSVPFunc != nil {
		return m.submitRSVPFunc(ctx, token, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRSVPService) UpdateRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
	if m.updateRSVPFunc != nil {
		return m.updateRSVPFunc(ctx, token, req)
	}
	return nil, errors.New("not implemented")
}

func TestRSVPHandler_SubmitRSVP_Success(t *testing.T) {
	mockService := &mockRSVPService{
		submitRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
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

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":2,"answers":[]}`
	req := httptest.NewRequest("POST", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestRSVPHandler_SubmitRSVP_InvalidJSON(t *testing.T) {
	mockService := &mockRSVPService{}
	handler := &RSVPHandler{rsvpService: mockService}

	body := `{invalid json`
	req := httptest.NewRequest("POST", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRSVPHandler_SubmitRSVP_ValidationError(t *testing.T) {
	mockService := &mockRSVPService{
		submitRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, &models.ValidationError{
				Field:   "plus_ones",
				Message: "you can bring up to 2 guest(s)",
			}
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":5,"answers":[]}`
	req := httptest.NewRequest("POST", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRSVPHandler_SubmitRSVP_DeadlinePassed(t *testing.T) {
	mockService := &mockRSVPService{
		submitRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, &models.DeadlinePassedError{
				Deadline: time.Now().Add(-1 * time.Hour),
				Message:  "RSVP deadline has passed",
			}
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":0,"answers":[]}`
	req := httptest.NewRequest("POST", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPHandler_SubmitRSVP_DuplicateRSVP(t *testing.T) {
	mockService := &mockRSVPService{
		submitRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, rsvp.ErrDuplicateRSVP
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":0,"answers":[]}`
	req := httptest.NewRequest("POST", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}
}

func TestRSVPHandler_SubmitRSVP_InternalError(t *testing.T) {
	mockService := &mockRSVPService{
		submitRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":0,"answers":[]}`
	req := httptest.NewRequest("POST", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRSVPHandler_SubmitRSVP_FormDataSuccess(t *testing.T) {
	mockService := &mockRSVPService{
		submitRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			if req.Response != "yes" {
				t.Errorf("Expected response 'yes', got %s", req.Response)
			}
			if req.PlusOnes != 2 {
				t.Errorf("Expected plus_ones 2, got %d", req.PlusOnes)
			}
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

	handler := &RSVPHandler{rsvpService: mockService}

	body := "response=yes&plus_ones=2"
	req := httptest.NewRequest("POST", "/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.SubmitRSVP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
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

func TestRSVPHandler_UpdateRSVP_Success(t *testing.T) {
	mockService := &mockRSVPService{
		updateRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return &models.RSVP{
				ID:        1,
				InviteID:  1,
				Response:  models.RSVPResponseMaybe,
				PlusOnes:  1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"maybe","plus_ones":1,"answers":[]}`
	req := httptest.NewRequest("PUT", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.UpdateRSVP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestRSVPHandler_UpdateRSVP_NoExistingRSVP(t *testing.T) {
	mockService := &mockRSVPService{
		updateRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp"}
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":0,"answers":[]}`
	req := httptest.NewRequest("PUT", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.UpdateRSVP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRSVPHandler_UpdateRSVP_DeadlinePassed(t *testing.T) {
	mockService := &mockRSVPService{
		updateRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, &models.DeadlinePassedError{
				Deadline: time.Now().Add(-1 * time.Hour),
				Message:  "RSVP deadline has passed",
			}
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"no","plus_ones":0,"answers":[]}`
	req := httptest.NewRequest("PUT", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.UpdateRSVP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPHandler_UpdateRSVP_ValidationError(t *testing.T) {
	mockService := &mockRSVPService{
		updateRSVPFunc: func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
			return nil, &models.ValidationError{
				Field:   "plus_ones",
				Message: "you can bring up to 2 guest(s)",
			}
		},
	}

	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":5,"answers":[]}`
	req := httptest.NewRequest("PUT", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.UpdateRSVP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRSVPHandler_UpdateRSVP_InvalidJSON(t *testing.T) {
	mockService := &mockRSVPService{}
	handler := &RSVPHandler{rsvpService: mockService}

	body := `{invalid json`
	req := httptest.NewRequest("PUT", "/api/rsvp/validtoken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "validtoken")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.UpdateRSVP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_DeadlinePassed(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	pastDeadline := time.Now().Add(-1 * time.Hour)

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
		markViewedFunc: func(ctx context.Context, inviteID int64) error {
			return nil
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
			return nil, &models.NotFoundError{Resource: "rsvp"}
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`{{.Event.Title}}|DeadlinePassed:{{.DeadlinePassed}}`))
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Event") {
		t.Error("Expected response to contain event title")
	}
	if !strings.Contains(body, "DeadlinePassed:true") {
		t.Errorf("Expected DeadlinePassed to be true, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_DeadlineNotPassed(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	futureDeadline := time.Now().Add(12 * time.Hour)

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
		markViewedFunc: func(ctx context.Context, inviteID int64) error {
			return nil
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
			return nil, &models.NotFoundError{Resource: "rsvp"}
		},
	}

	mockQuestionRepo := &mockRSVPQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`{{.Event.Title}}|DeadlinePassed:{{.DeadlinePassed}}`))
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Event") {
		t.Error("Expected response to contain event title")
	}
	if !strings.Contains(body, "DeadlinePassed:false") {
		t.Errorf("Expected DeadlinePassed to be false, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_NoDeadline(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)

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
		markViewedFunc: func(ctx context.Context, inviteID int64) error {
			return nil
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
				RSVPDeadline: nil,
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

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`{{.Event.Title}}|DeadlinePassed:{{.DeadlinePassed}}`))
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Event") {
		t.Error("Expected response to contain event title")
	}
	if !strings.Contains(body, "DeadlinePassed:false") {
		t.Errorf("Expected DeadlinePassed to be false when no deadline, got: %s", body)
	}
}

func TestRSVPHandler_UpdateRSVP_EmptyToken(t *testing.T) {
	mockService := &mockRSVPService{}
	handler := &RSVPHandler{rsvpService: mockService}

	body := `{"response":"yes","plus_ones":0,"answers":[]}`
	req := httptest.NewRequest("PUT", "/api/rsvp/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.UpdateRSVP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}




func (m *mockRSVPEventRepository) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}
