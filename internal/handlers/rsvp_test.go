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
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
	mockrepos "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
	mocksvcs "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestRSVPHandler_GetRSVPPage_ValidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	rsvpDeadline := startTime.Add(-24 * time.Hour)

	email := "test@example.com"
	name := "Test User"
	desc := "Test event description"
	loc := "Test Location"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:          1,
		EventID:     1,
		Email:       &email,
		Name:        &name,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusSent,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}, nil)
	mockInviteSvc.EXPECT().MarkInviteViewed(gomock.Any(), gomock.Any()).Return(nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:           1,
		Title:        "Test Event",
		Description:  &desc,
		StartTime:    startTime,
		EndTime:      &endTime,
		Timezone:     "America/Los_Angeles",
		Location:     &loc,
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "rsvp"})

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetTemplates(testTemplate(t, "rsvp_page.html"))

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "invite"})

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(nil, &models.ForbiddenError{Message: "this invite has expired"})

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/expiredtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_RevokedInvite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(nil, &models.ForbiddenError{Message: "this invite has been revoked"})

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

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
		Title:  "Cancelled Event",
		Status: models.EventStatusCancelled,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_ArchivedEvent(t *testing.T) {
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
		Title:  "Archived Event",
		Status: models.EventStatusArchived,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

	handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_WithExistingRSVP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Status:    models.InviteStatusResponded,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)
	mockInviteSvc.EXPECT().MarkInviteViewed(gomock.Any(), gomock.Any()).Return(nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:       1,
		InviteID: 1,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}, nil)

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   999,
		Email:     &email,
		Status:    models.InviteStatusSent,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "event"})

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

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

func TestRSVPHandler_SubmitRSVP_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().SubmitRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:        1,
		InviteID:  1,
		Response:  models.RSVPResponseYes,
		PlusOnes:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().SubmitRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &models.ValidationError{
		Field:   "plus_ones",
		Message: "you can bring up to 2 guest(s)",
	})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().SubmitRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &models.DeadlinePassedError{
		Deadline: time.Now().Add(-1 * time.Hour),
		Message:  "RSVP deadline has passed",
	})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().SubmitRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rsvp.ErrDuplicateRSVP)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().SubmitRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("database connection failed"))

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().SubmitRSVP(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error) {
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
		})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().UpdateRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.RSVP{
		ID:        1,
		InviteID:  1,
		Response:  models.RSVPResponseMaybe,
		PlusOnes:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().UpdateRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "rsvp"})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().UpdateRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &models.DeadlinePassedError{
		Deadline: time.Now().Add(-1 * time.Hour),
		Message:  "RSVP deadline has passed",
	})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
	mockService.EXPECT().UpdateRSVP(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &models.ValidationError{
		Field:   "plus_ones",
		Message: "you can bring up to 2 guest(s)",
	})

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
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
		Status:    models.InviteStatusSent,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)
	mockInviteSvc.EXPECT().MarkInviteViewed(gomock.Any(), gomock.Any()).Return(nil)

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
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "rsvp"})

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

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
		Status:    models.InviteStatusSent,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)
	mockInviteSvc.EXPECT().MarkInviteViewed(gomock.Any(), gomock.Any()).Return(nil)

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
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "rsvp"})

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"

	mockInviteSvc := mocksvcs.NewMockInviteService(ctrl)
	mockInviteSvc.EXPECT().GetInviteByToken(gomock.Any(), gomock.Any()).Return(&models.Invite{
		ID:        1,
		EventID:   1,
		Email:     &email,
		Status:    models.InviteStatusSent,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}, nil)
	mockInviteSvc.EXPECT().MarkInviteViewed(gomock.Any(), gomock.Any()).Return(nil)

	mockEventRepo := mockrepos.NewMockEventRepository(ctrl)
	mockEventRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&models.Event{
		ID:           1,
		Title:        "Test Event",
		StartTime:    startTime,
		Timezone:     "America/Los_Angeles",
		Status:       models.EventStatusPublished,
		RSVPDeadline: nil,
	}, nil)

	mockRSVPRepo := mockrepos.NewMockRSVPRepository(ctrl)
	mockRSVPRepo.EXPECT().GetByInviteID(gomock.Any(), gomock.Any()).Return(nil, &models.NotFoundError{Resource: "rsvp"})

	mockQuestionRepo := mockrepos.NewMockQuestionRepository(ctrl)
	mockQuestionRepo.EXPECT().GetByEventID(gomock.Any(), gomock.Any()).Return([]*models.PreferenceQuestion{}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocksvcs.NewMockRSVPService(ctrl)
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
