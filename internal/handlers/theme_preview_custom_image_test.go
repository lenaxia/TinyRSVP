package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestHandleThemePreview_WithCustomImage(t *testing.T) {
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

	customImageURL := "https://example.com/images/123/custom.jpg"
	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=1&custom_image_url="+customImageURL,
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, customImageURL) {
		t.Errorf("Expected response to contain custom image URL %s, but it was not found", customImageURL)
	}
}

func TestHandleThemePreview_CustomImageInHeader(t *testing.T) {
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

	customImageURL := "https://example.com/images/456/header.png"
	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=1&custom_image_url="+customImageURL,
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "theme-header-image") {
		t.Error("Expected response to contain theme-header-image class")
	}
	if !strings.Contains(body, customImageURL) {
		t.Errorf("Expected header image to use custom URL %s", customImageURL)
	}
}

func TestHandleThemePreview_NoCustomImage_UsesThemeDefault(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          1,
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<div>{{.Event.Title}}</div>",
		Category:    "elegant",
	}

	handler := &TemplateHandlers{
		service: &mockTemplateService{
			GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
				return mockTemplate, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=1",
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "https://example.com") {
		t.Error("Should not contain custom image URL when none provided")
	}
}

func TestHandleThemePreview_EmptyCustomImageURL(t *testing.T) {
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
		"/api/themes/preview?theme_id=1&custom_image_url=",
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleThemePreview_CustomImageWithAllEventData(t *testing.T) {
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

	customImageURL := "https://example.com/images/789/wedding.jpg"
	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=1&custom_image_url="+customImageURL+
			"&title=My+Wedding&location=Beach&description=Join+us",
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, customImageURL) {
		t.Errorf("Expected custom image URL %s in response", customImageURL)
	}
	if !strings.Contains(body, "My Wedding") {
		t.Error("Expected event title in response")
	}
	if !strings.Contains(body, "Beach") {
		t.Error("Expected location in response")
	}
}

func TestHandleThemePreview_CustomImageWithLightAndDarkMode(t *testing.T) {
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

	customImageURL := "https://example.com/images/999/custom.jpg"

	tests := []struct {
		name      string
		themeMode string
	}{
		{"light mode with custom image", "light"},
		{"dark mode with custom image", "dark"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet,
				"/api/themes/preview?theme_id=1&custom_image_url="+customImageURL+
					"&theme_mode="+tt.themeMode,
				nil)
			w := httptest.NewRecorder()

			handler.HandleThemePreview(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			body := w.Body.String()
			if !strings.Contains(body, customImageURL) {
				t.Errorf("Expected custom image URL in %s", tt.themeMode)
			}
			if !strings.Contains(body, `data-theme="`+tt.themeMode+`"`) {
				t.Errorf("Expected data-theme=%s in response", tt.themeMode)
			}
		})
	}
}

func TestHandleThemePreview_CustomImageURLEncoding(t *testing.T) {
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

	customImageURL := "https://example.com/images/123/my%20image.jpg"
	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=1&custom_image_url="+customImageURL,
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "my%20image.jpg") && !strings.Contains(body, "my image.jpg") {
		t.Error("Expected URL-encoded or decoded image filename in response")
	}
}
