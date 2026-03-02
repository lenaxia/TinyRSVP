package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestThemePreviewIntegration_CustomImageFlow(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          1,
		Name:        "Wedding Theme",
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

	tests := []struct {
		name              string
		customImageURL    string
		expectImageInHTML bool
		expectImageClass  bool
	}{
		{
			name:              "with custom image",
			customImageURL:    "https://storage.example.com/events/123/wedding.jpg",
			expectImageInHTML: true,
			expectImageClass:  true,
		},
		{
			name:              "without custom image",
			customImageURL:    "",
			expectImageInHTML: false,
			expectImageClass:  false,
		},
		{
			name:              "with custom image from CDN",
			customImageURL:    "https://cdn.example.com/images/event-header.png",
			expectImageInHTML: true,
			expectImageClass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/themes/preview?theme_id=1&title=Our+Wedding&location=Beach"
			if tt.customImageURL != "" {
				url += "&custom_image_url=" + tt.customImageURL
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.HandleThemePreview(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			body := w.Body.String()

			if tt.expectImageInHTML {
				if !strings.Contains(body, tt.customImageURL) {
					t.Errorf("Expected HTML to contain custom image URL: %s", tt.customImageURL)
				}
			} else {
				if strings.Contains(body, `<div class="event-header-image">`) {
					t.Error("Expected no header image div when no custom image provided")
				}
			}

			if tt.expectImageClass {
				if !strings.Contains(body, "event-header-image") {
					t.Error("Expected HTML to contain event-header-image class")
				}
				if !strings.Contains(body, "<img") {
					t.Error("Expected HTML to contain img tag")
				}
			}

			if !strings.Contains(body, "Our Wedding") {
				t.Error("Expected HTML to contain event title")
			}
			if !strings.Contains(body, "Beach") {
				t.Error("Expected HTML to contain location")
			}
		})
	}
}

func TestThemePreviewIntegration_FallbackBehavior(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          2,
		Name:        "Modern Theme",
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

	t.Run("custom image takes precedence", func(t *testing.T) {
		customImageURL := "https://example.com/custom.jpg"
		req := httptest.NewRequest(http.MethodGet,
			"/api/themes/preview?theme_id=2&custom_image_url="+customImageURL,
			nil)
		w := httptest.NewRecorder()

		handler.HandleThemePreview(w, req)

		body := w.Body.String()
		if !strings.Contains(body, customImageURL) {
			t.Error("Custom image should be displayed when provided")
		}
	})

	t.Run("no image when custom not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/api/themes/preview?theme_id=2",
			nil)
		w := httptest.NewRecorder()

		handler.HandleThemePreview(w, req)

		body := w.Body.String()
		if strings.Contains(body, `<div class="event-header-image">`) {
			t.Error("Should not display header image div when no custom image")
		}
	})
}

func TestThemePreviewIntegration_ResponsiveImageDisplay(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          3,
		Name:        "Responsive Theme",
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

	customImageURL := "https://example.com/responsive.jpg"
	req := httptest.NewRequest(http.MethodGet,
		"/api/themes/preview?theme_id=3&custom_image_url="+customImageURL,
		nil)
	w := httptest.NewRecorder()

	handler.HandleThemePreview(w, req)

	body := w.Body.String()

	if !strings.Contains(body, "width: 100%") {
		t.Error("Expected responsive width styling")
	}
	if !strings.Contains(body, "object-fit: cover") {
		t.Error("Expected object-fit: cover for proper image display")
	}
	if !strings.Contains(body, "max-height: 400px") {
		t.Error("Expected max-height constraint")
	}
}

func TestThemePreviewIntegration_LightDarkModeWithCustomImage(t *testing.T) {
	mockTemplate := &models.Template{
		ID:          4,
		Name:        "Dual Mode Theme",
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

	customImageURL := "https://example.com/image.jpg"

	modes := []string{"light", "dark"}
	for _, mode := range modes {
		t.Run("mode_"+mode, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet,
				"/api/themes/preview?theme_id=4&custom_image_url="+customImageURL+"&theme_mode="+mode,
				nil)
			w := httptest.NewRecorder()

			handler.HandleThemePreview(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			body := w.Body.String()
			if !strings.Contains(body, customImageURL) {
				t.Errorf("Custom image should display in %s mode", mode)
			}
			if !strings.Contains(body, `data-theme="`+mode+`"`) {
				t.Errorf("Expected data-theme=%s attribute", mode)
			}
		})
	}
}
