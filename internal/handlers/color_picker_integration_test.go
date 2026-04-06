package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestColorPickerIntegration(t *testing.T) {
	tests := []struct {
		name           string
		themeID        string
		customColor    string
		customImage    string
		wantStatus     int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:        "preview with custom color and custom image",
			themeID:     "1",
			customColor: "#FF5733",
			customImage: "https://example.com/image.jpg",
			wantStatus:  http.StatusOK,
			wantContains: []string{
				"--primary-color: #FF5733",
				`<img class="theme-header-image" src="https://example.com/image.jpg"`,
			},
		},
		{
			name:        "preview with custom color only",
			themeID:     "1",
			customColor: "#00FF00",
			customImage: "",
			wantStatus:  http.StatusOK,
			wantContains: []string{
				"--primary-color: #00FF00",
			},
			wantNotContain: []string{
				`<div class="event-header-image">`,
			},
		},
		{
			name:        "preview with custom image only",
			themeID:     "1",
			customColor: "",
			customImage: "https://example.com/image.jpg",
			wantStatus:  http.StatusOK,
			wantContains: []string{
				`<img class="theme-header-image" src="https://example.com/image.jpg"`,
			},
			wantNotContain: []string{
				"--primary-color:",
			},
		},
		{
			name:        "preview with neither custom color nor image",
			themeID:     "1",
			customColor: "",
			customImage: "",
			wantStatus:  http.StatusOK,
			wantNotContain: []string{
				"--primary-color:",
				`<div class="event-header-image">`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					return &models.Template{
						ID:          1,
						Name:        "Test Theme",
						Category:    "modern",
						Description: "Test theme",
					}, nil
				},
			}

			handlers := &TemplateHandlers{
				service: mockService,
			}

			url := "/api/themes/preview?theme_id=" + tt.themeID
			if tt.customColor != "" {
				url += "&custom_color=" + tt.customColor
			}
			if tt.customImage != "" {
				url += "&custom_image_url=" + tt.customImage
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handlers.HandleThemePreview(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			body := w.Body.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("Expected response to contain %q, but it didn't", want)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(body, notWant) {
					t.Errorf("Expected response NOT to contain %q, but it did", notWant)
				}
			}
		})
	}
}

func TestColorPickerIntegrationWithInvalidColors(t *testing.T) {
	tests := []struct {
		name        string
		customColor string
		wantStatus  int
	}{
		{
			name:        "invalid color is ignored - too short",
			customColor: "#FFF",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "invalid color is ignored - no hash",
			customColor: "FF5733",
			wantStatus:  http.StatusOK,
		},
		{
			name:        "invalid color is ignored - invalid chars",
			customColor: "#GGGGGG",
			wantStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					return &models.Template{
						ID:          1,
						Name:        "Test Theme",
						Category:    "modern",
						Description: "Test theme",
					}, nil
				},
			}

			handlers := &TemplateHandlers{
				service: mockService,
			}

			url := "/api/themes/preview?theme_id=1&custom_color=" + tt.customColor
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handlers.HandleThemePreview(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			body := w.Body.String()
			if strings.Contains(body, "--primary-color:") {
				t.Error("Expected invalid color to be ignored, but custom color CSS was included")
			}
		})
	}
}

func TestColorPickerCSSVariableOverride(t *testing.T) {
	mockService := &mockTemplateService{
		GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:          1,
				Name:        "Test Theme",
				Category:    "modern",
				Description: "Test theme",
			}, nil
		},
	}

	handlers := &TemplateHandlers{
		service: mockService,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/themes/preview?theme_id=1&custom_color=%23FF5733", nil)
	w := httptest.NewRecorder()

	handlers.HandleThemePreview(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	requiredVariables := []string{
		"--primary-color: #FF5733",
		"--primary-color-hover: #FF5733",
		"--primary-color-alpha: #FF573333",
	}

	for _, variable := range requiredVariables {
		if !strings.Contains(body, variable) {
			t.Errorf("Expected CSS to contain variable %q", variable)
		}
	}

	if !strings.Contains(body, ":root {") {
		t.Error("Expected CSS variables to be in :root selector")
	}
}
