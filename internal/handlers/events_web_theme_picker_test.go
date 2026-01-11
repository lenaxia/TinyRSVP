package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestParseEventFormDataWithTemplateID(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test Event")
	form.Set("start_time", "2026-06-15T14:00")
	form.Set("timezone", "America/Los_Angeles")
	form.Set("template_id", "3")

	event, err := parseEventFormData(form)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if event.TemplateID == nil {
		t.Fatal("Expected TemplateID to be set")
	}

	if *event.TemplateID != 3 {
		t.Errorf("Expected TemplateID to be 3, got %d", *event.TemplateID)
	}
}

func TestParseEventFormDataWithoutTemplateID(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test Event")
	form.Set("start_time", "2026-06-15T14:00")
	form.Set("timezone", "America/Los_Angeles")

	event, err := parseEventFormData(form)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if event.TemplateID != nil {
		t.Errorf("Expected TemplateID to be nil when not provided, got %d", *event.TemplateID)
	}
}

func TestParseEventFormDataWithInvalidTemplateID(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test Event")
	form.Set("start_time", "2026-06-15T14:00")
	form.Set("timezone", "America/Los_Angeles")
	form.Set("template_id", "invalid")

	event, err := parseEventFormData(form)
	if err == nil {
		t.Error("Expected error for invalid template_id")
	}

	if event != nil {
		t.Error("Expected nil event when template_id is invalid")
	}

	if !strings.Contains(err.Error(), "template_id") {
		t.Errorf("Error should mention template_id, got: %v", err)
	}
}

func TestParseEventFormDataWithZeroTemplateID(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test Event")
	form.Set("start_time", "2026-06-15T14:00")
	form.Set("timezone", "America/Los_Angeles")
	form.Set("template_id", "0")

	event, err := parseEventFormData(form)
	if err == nil {
		t.Error("Expected error for zero template_id")
	}

	if event != nil {
		t.Error("Expected nil event when template_id is zero")
	}
}

func TestParseEventFormDataWithNegativeTemplateID(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test Event")
	form.Set("start_time", "2026-06-15T14:00")
	form.Set("timezone", "America/Los_Angeles")
	form.Set("template_id", "-5")

	event, err := parseEventFormData(form)
	if err == nil {
		t.Error("Expected error for negative template_id")
	}

	if event != nil {
		t.Error("Expected nil event when template_id is negative")
	}
}

func TestNewEventFormWithTemplateService(t *testing.T) {
	mockEventService := &mockEventService{}
	mockTemplateService := &mockTemplateService{
		ListTemplatesFunc: func(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
			return []*models.Template{
				{
					ID:           1,
					Name:         "Plain Theme",
					Type:         models.TemplateTypeRSVPPage,
					Category:     models.CategoryPlain,
					Description:  "Simple plain text theme",
					ThumbnailURL: ptrString("/static/images/themes/plain-thumbnail.svg"),
					IsDefault:    true,
					IsActive:     true,
					SortOrder:    1,
				},
				{
					ID:           2,
					Name:         "Card Theme",
					Type:         models.TemplateTypeRSVPPage,
					Category:     models.CategoryCard,
					Description:  "Elegant card design",
					ThumbnailURL: ptrString("/static/images/themes/card-thumbnail.svg"),
					IsDefault:    false,
					IsActive:     true,
					SortOrder:    2,
				},
			}, nil
		},
		GetDefaultTemplateFunc: func(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Plain Theme",
				Type:      models.TemplateTypeRSVPPage,
				IsDefault: true,
			}, nil
		},
	}

	handler := &EventWebHandlers{
		service:         mockEventService,
		templateService: mockTemplateService,
	}

	req := httptest.NewRequest(http.MethodGet, "/events/new", nil)
	user := &models.User{
		ID:   1,
		Role: models.RoleEventManager,
	}
	ctx := auth.WithUser(req.Context(), user)
	ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.NewEventForm(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func ptrString(s string) *string {
	return &s
}

func ptrInt64(i int64) *int64 {
	return &i
}
