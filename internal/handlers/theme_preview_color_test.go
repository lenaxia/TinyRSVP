package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestHandleThemePreviewCustomColor(t *testing.T) {
	tests := []struct {
		name           string
		themeID        string
		customColor    string
		wantStatus     int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:        "preview with custom color",
			themeID:     "1",
			customColor: "#FF5733",
			wantStatus:  http.StatusOK,
			wantContains: []string{
				"--primary-color: #FF5733",
				"<style>",
				":root",
			},
		},
		{
			name:        "preview without custom color uses default",
			themeID:     "1",
			customColor: "",
			wantStatus:  http.StatusOK,
			wantNotContain: []string{
				"--primary-color: #",
			},
		},
		{
			name:        "preview with lowercase hex color",
			themeID:     "1",
			customColor: "#ff5733",
			wantStatus:  http.StatusOK,
			wantContains: []string{
				"--primary-color: #ff5733",
			},
		},
		{
			name:        "preview with uppercase hex color",
			themeID:     "1",
			customColor: "#FF5733",
			wantStatus:  http.StatusOK,
			wantContains: []string{
				"--primary-color: #FF5733",
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

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handlers.HandleThemePreview(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			body := w.Body.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("Expected response to contain %q, but it didn't.\nBody:\n%s", want, body)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(body, notWant) {
					t.Errorf("Expected response NOT to contain %q, but it did.\nBody:\n%s", notWant, body)
				}
			}
		})
	}
}

func TestHandleThemePreviewCustomColorValidation(t *testing.T) {
	tests := []struct {
		name        string
		customColor string
		wantStatus  int
		wantError   bool
	}{
		{
			name:        "valid 6-digit hex color",
			customColor: "#FF5733",
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "valid lowercase hex color",
			customColor: "#ff5733",
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "invalid hex color - too short",
			customColor: "#FFF",
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "invalid hex color - no hash",
			customColor: "FF5733",
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "invalid hex color - invalid characters",
			customColor: "#GGGGGG",
			wantStatus:  http.StatusOK,
			wantError:   false,
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
		})
	}
}

func TestHandleThemePreviewCustomColorCSS(t *testing.T) {
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

	requiredCSS := []string{
		"<style>",
		":root",
		"--primary-color:",
		"#FF5733",
		"</style>",
	}

	for _, css := range requiredCSS {
		if !strings.Contains(body, css) {
			t.Errorf("Expected response to contain CSS %q", css)
		}
	}

	if !strings.Contains(body, "<head>") {
		t.Error("Expected custom color CSS to be in <head> section")
	}
}
