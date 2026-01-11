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
)

type mockTemplateRepository struct {
	getByIDFunc          func(ctx context.Context, id int64) (*models.Template, error)
	getDefaultByTypeFunc func(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
}

func (m *mockTemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTemplateRepository) GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	if m.getDefaultByTypeFunc != nil {
		return m.getDefaultByTypeFunc(ctx, templateType)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTemplateRepository) Create(ctx context.Context, template *models.Template) error {
	return nil
}

func (m *mockTemplateRepository) GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) List(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) Update(ctx context.Context, template *models.Template) error {
	return nil
}

func (m *mockTemplateRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockTemplateRepository) SetActive(ctx context.Context, id int64, active bool) error {
	return nil
}

func (m *mockTemplateRepository) IsTemplateInUse(ctx context.Context, id int64) (bool, error) {
	return false, nil
}

func (m *mockTemplateRepository) SetDefault(ctx context.Context, id int64) error {
	return nil
}

func (m *mockTemplateRepository) GetTemplatesByCategory(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) ListThemes(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) GetByNameAndType(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) GetComponentConfig(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error) {
	return nil, nil
}

func (m *mockTemplateRepository) UpdateComponentConfig(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error {
	return nil
}

func (m *mockTemplateRepository) ValidateComponentConfig(ctx context.Context, config *models.ComponentConfiguration) error {
	return nil
}

func TestRSVPHandler_GetRSVPPage_WithEventTheme(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(5)
	imageURL := "/static/images/themes/wedding-elegance-header.svg"

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
				ID:         1,
				Title:      "Wedding Reception",
				StartTime:  startTime,
				Timezone:   "America/Los_Angeles",
				Status:     models.EventStatusPublished,
				TemplateID: &templateID,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			if id != 5 {
				t.Errorf("Expected template ID 5, got %d", id)
			}
			return &models.Template{
				ID:          5,
				Name:        "Wedding Elegance",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryCard,
				ImageURL:    &imageURL,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeCategory:{{.ThemeCategory}}|ThemeImageURL:{{.ThemeImageURL}}`))
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
	if !strings.Contains(body, "ThemeCategory:card") {
		t.Errorf("Expected theme category 'card', got: %s", body)
	}
	if !strings.Contains(body, "ThemeImageURL:/static/images/themes/wedding-elegance-header.svg") {
		t.Errorf("Expected theme image URL, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_WithDefaultTheme(t *testing.T) {
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
				ID:         1,
				Title:      "Simple Event",
				StartTime:  startTime,
				Timezone:   "America/Los_Angeles",
				Status:     models.EventStatusPublished,
				TemplateID: nil,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getDefaultByTypeFunc: func(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
			if templateType != models.TemplateTypeRSVPPage {
				t.Errorf("Expected template type rsvp_page, got %s", templateType)
			}
			return &models.Template{
				ID:          1,
				Name:        "Plain Text",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryPlain,
				IsDefault:   true,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeCategory:{{.ThemeCategory}}`))
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
	if !strings.Contains(body, "ThemeCategory:plain") {
		t.Errorf("Expected default theme category 'plain', got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_ThemeLoadError_FallbackToDefault(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	invalidTemplateID := int64(999)

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
				ID:         1,
				Title:      "Event With Invalid Theme",
				StartTime:  startTime,
				Timezone:   "America/Los_Angeles",
				Status:     models.EventStatusPublished,
				TemplateID: &invalidTemplateID,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return nil, &models.NotFoundError{Resource: "Template", ID: id}
		},
		getDefaultByTypeFunc: func(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
			return &models.Template{
				ID:          1,
				Name:        "Plain Text",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryPlain,
				IsDefault:   true,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeCategory:{{.ThemeCategory}}`))
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
	if !strings.Contains(body, "ThemeCategory:plain") {
		t.Errorf("Expected fallback to default theme 'plain', got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_WithCustomThemeImage(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(5)
	customImageURL := "/uploads/custom-header.jpg"
	defaultImageURL := "/static/images/themes/wedding-elegance-header.svg"

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
				ID:                  1,
				Title:               "Wedding with Custom Image",
				StartTime:           startTime,
				Timezone:            "America/Los_Angeles",
				Status:              models.EventStatusPublished,
				TemplateID:          &templateID,
				CustomThemeImageURL: &customImageURL,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          5,
				Name:        "Wedding Elegance",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryCard,
				ImageURL:    &defaultImageURL,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeImageURL:{{.ThemeImageURL}}`))
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
	if !strings.Contains(body, "ThemeImageURL:/uploads/custom-header.jpg") {
		t.Errorf("Expected custom image URL to override default, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_WithCustomThemeColor(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(5)
	customColor := "#007BFF"

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
				ID:               1,
				Title:            "Birthday Party",
				StartTime:        startTime,
				Timezone:         "America/Los_Angeles",
				Status:           models.EventStatusPublished,
				TemplateID:       &templateID,
				CustomThemeColor: &customColor,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          5,
				Name:        "Birthday Celebration",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryCard,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeColor:{{.ThemeColor}}`))
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
	if !strings.Contains(body, "--theme-primary: #007BFF") {
		t.Errorf("Expected custom theme color CSS, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_NoTemplateRepository(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(5)

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
				ID:         1,
				Title:      "Event",
				StartTime:  startTime,
				Timezone:   "America/Los_Angeles",
				Status:     models.EventStatusPublished,
				TemplateID: &templateID,
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

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeCategory:{{.ThemeCategory}}`))
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
	if !strings.Contains(body, "ThemeCategory:") {
		t.Errorf("Expected empty theme category when no template repo, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_WithBothCustomOverrides(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(5)
	customImageURL := "/uploads/custom-header.jpg"
	customColor := "#16A34A"
	defaultImageURL := "/static/images/themes/default-header.svg"

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
				ID:                  1,
				Title:               "Fully Customized Event",
				StartTime:           startTime,
				Timezone:            "America/Los_Angeles",
				Status:              models.EventStatusPublished,
				TemplateID:          &templateID,
				CustomThemeImageURL: &customImageURL,
				CustomThemeColor:    &customColor,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          5,
				Name:        "Corporate Professional",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryCard,
				ImageURL:    &defaultImageURL,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeImageURL:{{.ThemeImageURL}}|ThemeColor:{{.ThemeColor}}`))
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
	if !strings.Contains(body, "ThemeImageURL:/uploads/custom-header.jpg") {
		t.Errorf("Expected custom image URL, got: %s", body)
	}
	if !strings.Contains(body, "--theme-primary: #16A34A") {
		t.Errorf("Expected custom color CSS, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_EmptyCustomOverrides(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(5)
	emptyString := ""
	defaultImageURL := "/static/images/themes/wedding-elegance-header.svg"

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
				ID:                  1,
				Title:               "Event With Empty Overrides",
				StartTime:           startTime,
				Timezone:            "America/Los_Angeles",
				Status:              models.EventStatusPublished,
				TemplateID:          &templateID,
				CustomThemeImageURL: &emptyString,
				CustomThemeColor:    &emptyString,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          5,
				Name:        "Wedding Elegance",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryCard,
				ImageURL:    &defaultImageURL,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeImageURL:{{.ThemeImageURL}}|ThemeColor:{{.ThemeColor}}`))
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
	if !strings.Contains(body, "ThemeImageURL:/static/images/themes/wedding-elegance-header.svg") {
		t.Errorf("Expected default image URL when custom is empty, got: %s", body)
	}
	if strings.Contains(body, "--theme-primary:") {
		t.Errorf("Expected no theme color CSS when custom is empty, got: %s", body)
	}
}

func TestRSVPHandler_GetRSVPPage_DefaultThemeLoadError(t *testing.T) {
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
				ID:         1,
				Title:      "Event",
				StartTime:  startTime,
				Timezone:   "America/Los_Angeles",
				Status:     models.EventStatusPublished,
				TemplateID: nil,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getDefaultByTypeFunc: func(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
			return nil, errors.New("database error")
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
	handler.SetTemplateRepository(mockTemplateRepo)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest("GET", "/rsvp/validtoken", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when default theme fails to load, got %d", w.Code)
	}
}

func TestRSVPHandler_GetRSVPPage_ThemeWithNoImage(t *testing.T) {
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(1)

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
				ID:         1,
				Title:      "Plain Event",
				StartTime:  startTime,
				Timezone:   "America/Los_Angeles",
				Status:     models.EventStatusPublished,
				TemplateID: &templateID,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          1,
				Name:        "Plain Text",
				Type:        models.TemplateTypeRSVPPage,
				Category:    models.CategoryPlain,
				ImageURL:    nil,
				HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
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
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`ThemeCategory:{{.ThemeCategory}}|ThemeImageURL:{{.ThemeImageURL}}`))
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
	if !strings.Contains(body, "ThemeCategory:plain") {
		t.Errorf("Expected theme category 'plain', got: %s", body)
	}
	if strings.Contains(body, "ThemeImageURL:/") {
		t.Errorf("Expected empty image URL for plain theme, got: %s", body)
	}
}