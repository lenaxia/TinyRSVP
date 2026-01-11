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

func TestRSVPHandler_GetRSVPPage_WithCustomColor(t *testing.T) {
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

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`
<!DOCTYPE html>
<html{{if .ThemeCategory}} data-event-theme="{{.ThemeCategory}}"{{end}}>
<head>
{{.ThemeColor}}
</head>
<body>
<h1>{{.Event.Title}}</h1>
</body>
</html>
`))
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
	if !strings.Contains(body, "--theme-primary: #007BFF !important") {
		t.Error("Expected custom color CSS override in response")
	}

	if !strings.Contains(body, "[data-theme=\"dark\"][data-event-theme]") {
		t.Error("Expected dark mode CSS override in response")
	}
}

func TestRSVPHandler_GetRSVPPage_WithoutCustomColor(t *testing.T) {
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
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
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

	handler := NewRSVPHandler(mockInviteService, mockEventRepo, mockRSVPRepo, mockQuestionRepo)

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`
<!DOCTYPE html>
<html{{if .ThemeCategory}} data-event-theme="{{.ThemeCategory}}"{{end}}>
<head>
{{.ThemeColor}}
</head>
<body>
<h1>{{.Event.Title}}</h1>
</body>
</html>
`))
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
	if strings.Contains(body, "--theme-primary:") {
		t.Error("Expected no custom color CSS override when no custom color set")
	}
}

func TestRSVPHandler_GetRSVPPage_WithInvalidCustomColor(t *testing.T) {
	invalidColor := "#FFFF00"
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
				CustomThemeColor: &invalidColor,
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

	tmpl := template.Must(template.New("rsvp_page.html").Parse(`
<!DOCTYPE html>
<html{{if .ThemeCategory}} data-event-theme="{{.ThemeCategory}}"{{end}}>
<head>
{{.ThemeColor}}
</head>
<body>
<h1>{{.Event.Title}}</h1>
</body>
</html>
`))
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
	if strings.Contains(body, "--theme-primary:") {
		t.Error("Expected no custom color CSS override for invalid color (fails contrast)")
	}
}

func TestRSVPHandler_GetRSVPPage_CustomColorFallback(t *testing.T) {
	tests := []struct {
		name           string
		customColor    *string
		expectOverride bool
		description    string
	}{
		{
			name:           "valid custom color",
			customColor:    ptrString("#007BFF"),
			expectOverride: true,
			description:    "should apply custom color override",
		},
		{
			name:           "no custom color",
			customColor:    nil,
			expectOverride: false,
			description:    "should not apply override when no custom color",
		},
		{
			name:           "empty custom color",
			customColor:    ptrString(""),
			expectOverride: false,
			description:    "should not apply override for empty string",
		},
		{
			name:           "invalid format",
			customColor:    ptrString("not-a-color"),
			expectOverride: false,
			description:    "should not apply override for invalid format",
		},
		{
			name:           "fails contrast check",
			customColor:    ptrString("#FFFF00"),
			expectOverride: false,
			description:    "should not apply override for color that fails contrast",
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

			tmpl := template.Must(template.New("rsvp_page.html").Parse(`
<!DOCTYPE html>
<html>
<head>
{{.ThemeColor}}
</head>
<body><h1>{{.Event.Title}}</h1></body>
</html>
`))
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
			hasOverride := strings.Contains(body, "--theme-primary:")

			if hasOverride != tt.expectOverride {
				t.Errorf("%s: hasOverride = %v, want %v", tt.description, hasOverride, tt.expectOverride)
			}
		})
	}
}
