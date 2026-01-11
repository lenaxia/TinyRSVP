package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestColorOverrideSystem_EndToEnd_Integration(t *testing.T) {
	tests := []struct {
		name              string
		customColor       *string
		expectCSS         bool
		expectedColorInCSS string
		description       string
	}{
		{
			name:              "valid color generates CSS override",
			customColor:       ptrString("#007BFF"),
			expectCSS:         true,
			expectedColorInCSS: "#007BFF",
			description:       "Bootstrap blue passes contrast on both backgrounds",
		},
		{
			name:              "valid green generates CSS override",
			customColor:       ptrString("#16A34A"),
			expectCSS:         true,
			expectedColorInCSS: "#16A34A",
			description:       "Green-600 passes contrast on both backgrounds",
		},
		{
			name:              "invalid light color rejected",
			customColor:       ptrString("#FFFF00"),
			expectCSS:         false,
			expectedColorInCSS: "",
			description:       "Yellow fails contrast on dark background",
		},
		{
			name:              "invalid dark color rejected",
			customColor:       ptrString("#000080"),
			expectCSS:         false,
			expectedColorInCSS: "",
			description:       "Navy fails contrast on dark background",
		},
		{
			name:              "no custom color",
			customColor:       nil,
			expectCSS:         false,
			expectedColorInCSS: "",
			description:       "No custom color means no CSS override",
		},
		{
			name:              "empty custom color",
			customColor:       ptrString(""),
			expectCSS:         false,
			expectedColorInCSS: "",
			description:       "Empty string means no CSS override",
		},
		{
			name:              "invalid format rejected",
			customColor:       ptrString("not-a-color"),
			expectCSS:         false,
			expectedColorInCSS: "",
			description:       "Invalid format is rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now().Add(24 * time.Hour)

			mockInviteService := &mockRSVPInviteService{
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
						Title:            "Test Event",
						StartTime:        startTime,
						Timezone:         "UTC",
						Status:           models.EventStatusPublished,
						CustomThemeColor: tt.customColor,
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

			handler := NewRSVPHandler(mockInviteService, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

			tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}
			handler.SetTemplates(tmpl)

			r := chi.NewRouter()
			r.Get("/rsvp/{token}", handler.GetRSVPPage)

			req := httptest.NewRequest(http.MethodGet, "/rsvp/test-token", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			body := w.Body.String()

			if tt.expectCSS {
				if !strings.Contains(body, "<style>") {
					t.Errorf("%s: Expected <style> tag in response", tt.description)
				}
				if !strings.Contains(body, "[data-event-theme]") {
					t.Errorf("%s: Expected [data-event-theme] selector", tt.description)
				}
				if !strings.Contains(body, "--theme-primary: "+tt.expectedColorInCSS) {
					t.Errorf("%s: Expected --theme-primary: %s in CSS", tt.description, tt.expectedColorInCSS)
				}
				if !strings.Contains(body, "[data-theme=\"dark\"][data-event-theme]") {
					t.Errorf("%s: Expected dark mode selector", tt.description)
				}
			} else {
				if strings.Contains(body, "--theme-primary:") {
					t.Errorf("%s: Expected no CSS override, but found one", tt.description)
				}
			}
		})
	}
}

func TestColorOverrideSystem_LightAndDarkMode_Integration(t *testing.T) {
	customColor := "#007BFF"
	startTime := time.Now().Add(24 * time.Hour)

	mockInviteService := &mockRSVPInviteService{
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
				Title:            "Test Event",
				StartTime:        startTime,
				Timezone:         "UTC",
				Status:           models.EventStatusPublished,
				CustomThemeColor: &customColor,
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

	handler := NewRSVPHandler(mockInviteService, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest(http.MethodGet, "/rsvp/test-token", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "[data-event-theme]") {
		t.Error("Expected light mode CSS selector")
	}

	if !strings.Contains(body, "[data-theme=\"dark\"][data-event-theme]") {
		t.Error("Expected dark mode CSS selector")
	}

	if !strings.Contains(body, "--theme-primary: #007BFF") {
		t.Error("Expected color to be applied in both modes")
	}

	cssCount := strings.Count(body, "--theme-primary: #007BFF")
	if cssCount != 2 {
		t.Errorf("Expected color to appear twice (light and dark mode), got %d occurrences", cssCount)
	}
}

func TestColorOverrideSystem_FallbackToThemeDefault_Integration(t *testing.T) {
	templateID := int64(5)
	invalidColor := "#CCCCCC"
	startTime := time.Now().Add(24 * time.Hour)

	mockInviteService := &mockRSVPInviteService{
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
				Title:            "Test Event",
				StartTime:        startTime,
				Timezone:         "UTC",
				Status:           models.EventStatusPublished,
				TemplateID:       &templateID,
				CustomThemeColor: &invalidColor,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          5,
				Name:        "Modern Minimalist",
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

	handler := NewRSVPHandler(mockInviteService, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest(http.MethodGet, "/rsvp/test-token", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	if strings.Contains(body, "--theme-primary: #CCCCCC") {
		t.Error("Invalid color should not generate CSS override")
	}

	if !strings.Contains(body, `/static/css/themes/card.css`) {
		t.Error("Should still load theme CSS even when custom color is invalid")
	}

	if !strings.Contains(body, `data-event-theme="card"`) {
		t.Error("Should still apply theme category even when custom color is invalid")
	}
}

func TestColorOverrideSystem_ContrastValidation_Integration(t *testing.T) {
	tests := []struct {
		name        string
		color       string
		shouldPass  bool
		description string
	}{
		{
			name:        "#007BFF passes",
			color:       "#007BFF",
			shouldPass:  true,
			description: "Bootstrap blue",
		},
		{
			name:        "#16A34A passes",
			color:       "#16A34A",
			shouldPass:  true,
			description: "Green-600",
		},
		{
			name:        "#2563EB passes",
			color:       "#2563EB",
			shouldPass:  true,
			description: "Blue-600",
		},
		{
			name:        "#FFFF00 fails",
			color:       "#FFFF00",
			shouldPass:  false,
			description: "Yellow fails on dark",
		},
		{
			name:        "#FFB6C1 fails",
			color:       "#FFB6C1",
			shouldPass:  false,
			description: "Light pink fails on dark",
		},
		{
			name:        "#000080 fails",
			color:       "#000080",
			shouldPass:  false,
			description: "Navy fails on dark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason := validateCustomColorContrast(tt.color)
			if valid != tt.shouldPass {
				t.Errorf("%s: valid = %v, want %v (reason: %s)", tt.description, valid, tt.shouldPass, reason)
			}

			if tt.shouldPass && reason != "" {
				t.Errorf("%s: expected no error reason, got: %s", tt.description, reason)
			}

			if !tt.shouldPass && reason == "" {
				t.Errorf("%s: expected error reason, got empty string", tt.description)
			}
		})
	}
}

func TestColorOverrideSystem_CSSGeneration_Integration(t *testing.T) {
	validColor := "#007BFF"
	css := generateColorOverrideCSS(validColor)

	if css == "" {
		t.Fatal("Expected CSS to be generated for valid color")
	}

	if !strings.Contains(css, "<style>") {
		t.Error("Expected <style> tag")
	}

	if !strings.Contains(css, "</style>") {
		t.Error("Expected closing </style> tag")
	}

	if !strings.Contains(css, "[data-event-theme]") {
		t.Error("Expected light mode selector")
	}

	if !strings.Contains(css, "[data-theme=\"dark\"][data-event-theme]") {
		t.Error("Expected dark mode selector")
	}

	if !strings.Contains(css, "--theme-primary: #007BFF !important") {
		t.Error("Expected CSS variable with !important")
	}

	occurrences := strings.Count(css, "--theme-primary: #007BFF !important")
	if occurrences != 2 {
		t.Errorf("Expected color to appear twice (light and dark), got %d", occurrences)
	}
}

func TestColorOverrideSystem_TemplateIntegration(t *testing.T) {
	customColor := "#16A34A"
	startTime := time.Now().Add(24 * time.Hour)
	templateID := int64(3)

	mockInviteService := &mockRSVPInviteService{
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
				Title:            "Garden Party",
				StartTime:        startTime,
				Timezone:         "UTC",
				Status:           models.EventStatusPublished,
				TemplateID:       &templateID,
				CustomThemeColor: &customColor,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          3,
				Name:        "Garden Party",
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

	handler := NewRSVPHandler(mockInviteService, mockEventRepo, mockRSVPRepo, mockQuestionRepo)
	handler.SetTemplateRepository(mockTemplateRepo)

	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	handler.SetTemplates(tmpl)

	r := chi.NewRouter()
	r.Get("/rsvp/{token}", handler.GetRSVPPage)

	req := httptest.NewRequest(http.MethodGet, "/rsvp/test-token", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, `data-event-theme="card"`) {
		t.Error("Expected theme category to be applied")
	}

	if !strings.Contains(body, `/static/css/themes/card.css`) {
		t.Error("Expected theme CSS to be loaded")
	}

	if !strings.Contains(body, "--theme-primary: #16A34A") {
		t.Error("Expected custom color to override theme default")
	}

	if !strings.Contains(body, "Garden Party") {
		t.Error("Expected event title to be rendered")
	}
}
