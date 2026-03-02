package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestHandleThemePreview_MissingThemeID(t *testing.T) {
	handler := &TemplateHandlers{
		service: &mockTemplateService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes/preview", nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleThemePreview_InvalidThemeID(t *testing.T) {
	handler := &TemplateHandlers{
		service: &mockTemplateService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes/preview?theme_id=invalid", nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleThemePreview_ThemeNotFound(t *testing.T) {
	handler := &TemplateHandlers{
		service: &mockTemplateService{
			GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
				return nil, &models.NotFoundError{Resource: "Template", ID: id}
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes/preview?theme_id=999", nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleThemePreview_Success(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          1,
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		Description: "Test theme description",
		HTMLContent: "<div>{{.Event.Title}}</div>",
		Category:    "modern",
		IsActive:    true,
		IsDefault:   false,
		Version:     1,
		CreatedBy:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	handler := &TemplateHandlers{
		service: &mockTemplateService{
			GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
				if id == 1 {
					return mockTemplate, nil
				}
				return nil, &models.NotFoundError{Resource: "Template", ID: id}
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes/preview?theme_id=1&title=Test+Event&location=Test+Location", nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}
}

func TestHandleThemePreview_WithDefaultData(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          1,
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<div>{{.Event.Title}}</div>",
		Category:    "modern",
	}

	handler := &TemplateHandlers{
		service: &mockTemplateService{
			GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
				return mockTemplate, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes/preview?theme_id=1", nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleThemePreview_WithThemeMode(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          1,
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<div>{{.Event.Title}}</div>",
		Category:    "modern",
	}

	handler := &TemplateHandlers{
		service: &mockTemplateService{
			GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
				return mockTemplate, nil
			},
		},
	}

	tests := []struct {
		name      string
		themeMode string
	}{
		{"light mode", "light"},
		{"dark mode", "dark"},
		{"no mode specified", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/themes/preview?theme_id=1"
			if tt.themeMode != "" {
				url += "&theme_mode=" + tt.themeMode
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.HandleThemePreview(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

func TestHandleThemePreview_WithEventData(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          1,
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<div>{{.Event.Title}}</div>",
		Category:    "modern",
	}

	handler := &TemplateHandlers{
		service: &mockTemplateService{
			GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
				return mockTemplate, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=1&title=My+Event&location=My+Location&description=My+Description&start_time=2026-01-15T10:00:00Z",
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}
