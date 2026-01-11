package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRSVPHandler_CSRFTokenIntegration(t *testing.T) {
	t.Run("RSVP page includes CSRF token in template data", func(t *testing.T) {
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
				return &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(48 * time.Hour),
					Timezone:  "UTC",
					Status:    models.EventStatusPublished,
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

		tmpl := template.Must(template.New("rsvp_page.html").Parse(`CSRFToken={{.CSRFToken}}|Token={{.Token}}`))
		handler.SetTemplates(tmpl)

		csrfToken := "test-csrf-token-12345"
		req := httptest.NewRequest(http.MethodGet, "/rsvp/valid-token", nil)
		ctx := context.WithValue(req.Context(), middleware.CSRFTokenKey, csrfToken)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.GetRSVPPage(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "CSRFToken="+csrfToken) {
			t.Errorf("Expected CSRF token in response body, got: %s", body)
		}
	})

	t.Run("RSVP page renders with empty CSRF token when not in context", func(t *testing.T) {
		mockInviteSvc := &mockRSVPInviteService{
			getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
				email := "test@example.com"
				return &models.Invite{
					ID:          1,
					EventID:     1,
					Email:       &email,
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
				return &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(48 * time.Hour),
					Timezone:  "UTC",
					Status:    models.EventStatusPublished,
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

		tmpl := template.Must(template.New("rsvp_page.html").Parse(`CSRFToken={{.CSRFToken}}|Token={{.Token}}`))
		handler.SetTemplates(tmpl)

		req := httptest.NewRequest(http.MethodGet, "/rsvp/valid-token", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		handler.GetRSVPPage(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "CSRFToken=") {
			t.Errorf("Expected CSRFToken field in response, got: %s", body)
		}
		if strings.Contains(body, "CSRFToken=test-csrf-token") {
			t.Errorf("Expected empty CSRF token, but got a value: %s", body)
		}
	})

	t.Run("confirmation page includes CSRF token in template data", func(t *testing.T) {
		mockInviteSvc := &mockRSVPInviteService{
			getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
				email := "test@example.com"
				return &models.Invite{
					ID:          1,
					EventID:     1,
					Email:       &email,
					MaxPlusOnes: 2,
					Status:      models.InviteStatusSent,
					ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
				}, nil
			},
		}

		mockEventRepo := &mockRSVPEventRepository{
			getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
				return &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(48 * time.Hour),
					Timezone:  "UTC",
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

		handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
		handler.SetAnswerRepository(&mockAnswerRepository{})

		csrfToken := "test-csrf-token-67890"
		tmpl := template.Must(template.New("confirmation.html").Parse(`CSRFToken={{.CSRFToken}}|Response={{.RSVP.Response}}`))
		handler.SetConfirmationTemplates(tmpl)

		req := httptest.NewRequest(http.MethodGet, "/rsvp/valid-token/confirmation", nil)
		ctx := context.WithValue(req.Context(), middleware.CSRFTokenKey, csrfToken)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.GetConfirmationPage(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "CSRFToken="+csrfToken) {
			t.Errorf("Expected CSRF token in confirmation page, got: %s", body)
		}
	})
}

func TestRSVPHandler_FormSubmissionWithCSRF(t *testing.T) {
	t.Run("form submission succeeds with valid CSRF token", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("response", "yes")
		formData.Set("plus_ones", "1")
		formData.Set("csrf_token", "valid-csrf-token")

		req := httptest.NewRequest(http.MethodPost, "/rsvp/valid-token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		csrfToken := "valid-csrf-token"
		cookie := &http.Cookie{
			Name:  middleware.CSRFCookieName,
			Value: csrfToken,
		}
		req.AddCookie(cookie)
		
		ctx := context.WithValue(req.Context(), middleware.CSRFTokenKey, csrfToken)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		
		csrfMiddleware := middleware.CSRF(32)
		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200 with valid CSRF token, got %d", rec.Code)
		}
	})

	t.Run("form submission fails without CSRF token", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("response", "yes")
		formData.Set("plus_ones", "1")

		req := httptest.NewRequest(http.MethodPost, "/rsvp/valid-token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rec := httptest.NewRecorder()
		
		csrfMiddleware := middleware.CSRF(32)
		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 without CSRF token, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Invalid or missing CSRF token") {
			t.Errorf("Expected CSRF error message, got: %s", body)
		}
	})

	t.Run("form submission fails with mismatched CSRF token", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("response", "yes")
		formData.Set("plus_ones", "1")
		formData.Set("csrf_token", "wrong-token")

		req := httptest.NewRequest(http.MethodPost, "/rsvp/valid-token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		cookie := &http.Cookie{
			Name:  middleware.CSRFCookieName,
			Value: "correct-token",
		}
		req.AddCookie(cookie)

		rec := httptest.NewRecorder()
		
		csrfMiddleware := middleware.CSRF(32)
		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 with mismatched CSRF token, got %d", rec.Code)
		}
	})
}

func TestRSVPHandler_TemplateCSRFRendering(t *testing.T) {
	t.Run("actual template renders CSRF token in hidden field", func(t *testing.T) {
		mockInviteSvc := &mockRSVPInviteService{
			getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
				email := "test@example.com"
				return &models.Invite{
					ID:          1,
					EventID:     1,
					Email:       &email,
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
				return &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(48 * time.Hour),
					Timezone:  "UTC",
					Status:    models.EventStatusPublished,
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

		tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}
		handler.SetTemplates(tmpl)

		csrfToken := "integration-test-csrf-token"
		req := httptest.NewRequest(http.MethodGet, "/rsvp/valid-token", nil)
		ctx := context.WithValue(req.Context(), middleware.CSRFTokenKey, csrfToken)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.GetRSVPPage(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		
		expectedHiddenField := `<input type="hidden" name="csrf_token" value="` + csrfToken + `">`
		if !strings.Contains(body, expectedHiddenField) {
			t.Errorf("Expected CSRF hidden field in rendered template")
			t.Logf("Looking for: %s", expectedHiddenField)
			if len(body) > 500 {
				t.Logf("Body snippet: %s", body[:500])
			} else {
				t.Logf("Body: %s", body)
			}
		}

		if !strings.Contains(body, `name="csrf_token"`) {
			t.Error("Expected csrf_token field name in form")
		}

		if !strings.Contains(body, csrfToken) {
			t.Error("Expected CSRF token value in rendered HTML")
		}
	})

	t.Run("confirmation page renders with CSRF token", func(t *testing.T) {
		mockInviteSvc := &mockRSVPInviteService{
			getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
				email := "test@example.com"
				return &models.Invite{
					ID:          1,
					EventID:     1,
					Email:       &email,
					MaxPlusOnes: 2,
					Status:      models.InviteStatusSent,
					ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
				}, nil
			},
		}

		mockEventRepo := &mockRSVPEventRepository{
			getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
				return &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(48 * time.Hour),
					Timezone:  "UTC",
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

		handler := NewRSVPHandler(mockInviteSvc, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
		handler.SetAnswerRepository(&mockConfirmationAnswerRepository{})

		csrfToken := "confirmation-csrf-token"
		tmpl := template.Must(template.New("confirmation.html").Parse(`CSRFToken={{.CSRFToken}}|Response={{.RSVP.Response}}`))
		handler.SetConfirmationTemplates(tmpl)

		req := httptest.NewRequest(http.MethodGet, "/rsvp/valid-token/confirmation", nil)
		ctx := context.WithValue(req.Context(), middleware.CSRFTokenKey, csrfToken)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.GetConfirmationPage(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "CSRFToken="+csrfToken) {
			t.Errorf("Expected CSRF token in confirmation page, got: %s", body)
		}
	})
}

func TestRSVPHandler_CSRFTokenInContext(t *testing.T) {
	t.Run("GetCSRFToken extracts token from context", func(t *testing.T) {
		expectedToken := "context-test-token"
		ctx := context.WithValue(context.Background(), middleware.CSRFTokenKey, expectedToken)

		actualToken := middleware.GetCSRFToken(ctx)

		if actualToken != expectedToken {
			t.Errorf("Expected token %q, got %q", expectedToken, actualToken)
		}
	})

	t.Run("GetCSRFToken returns empty string when not in context", func(t *testing.T) {
		ctx := context.Background()

		actualToken := middleware.GetCSRFToken(ctx)

		if actualToken != "" {
			t.Errorf("Expected empty token, got %q", actualToken)
		}
	})

	t.Run("GetCSRFToken returns empty string for wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.CSRFTokenKey, 12345)

		actualToken := middleware.GetCSRFToken(ctx)

		if actualToken != "" {
			t.Errorf("Expected empty token for wrong type, got %q", actualToken)
		}
	})
}

func TestRSVPHandler_ActualTemplateWithCSRF(t *testing.T) {
	t.Run("rsvp_page.html template includes csrf.js script", func(t *testing.T) {
		tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		mockInviteSvc := &mockRSVPInviteService{
			getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
				email := "test@example.com"
				return &models.Invite{
					ID:          1,
					EventID:     1,
					Email:       &email,
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
				return &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(48 * time.Hour),
					Timezone:  "UTC",
					Status:    models.EventStatusPublished,
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
		handler.SetTemplates(tmpl)

		csrfToken := "script-test-token"
		req := httptest.NewRequest(http.MethodGet, "/rsvp/valid-token", nil)
		ctx := context.WithValue(req.Context(), middleware.CSRFTokenKey, csrfToken)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.GetRSVPPage(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		
		if !strings.Contains(body, `<script src="/static/js/csrf.js"></script>`) {
			t.Error("Expected csrf.js script tag in rendered template")
		}
	})
}

type mockConfirmationAnswerRepository struct{}

func (m *mockConfirmationAnswerRepository) Create(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockConfirmationAnswerRepository) GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
	return []*models.RSVPAnswer{}, nil
}

func (m *mockConfirmationAnswerRepository) GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error) {
	return []*models.RSVPAnswer{}, nil
}

func (m *mockConfirmationAnswerRepository) Update(ctx context.Context, answer *models.RSVPAnswer) error {
	return nil
}

func (m *mockConfirmationAnswerRepository) DeleteByRSVPID(ctx context.Context, rsvpID int64) error {
	return nil
}
