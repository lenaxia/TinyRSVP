package css

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStaticCSSFileServing(t *testing.T) {
	fs := http.FileServer(http.Dir("../../static"))
	handler := http.StripPrefix("/static/", fs)

	t.Run("serves variables.css with correct status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/variables.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("serves variables.css with correct content-type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/variables.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/css") && !strings.Contains(contentType, "text/plain") {
			t.Errorf("Expected Content-Type to contain 'text/css' or 'text/plain', got %s", contentType)
		}
	})

	t.Run("serves variables.css with expected CSS variables", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/variables.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		expectedVariables := []string{
			"--color-primary-600",
			"--color-success",
			"--color-success-dark",
			"--color-warning-dark",
			"--color-warning-darker",
			"--color-text-disabled",
			"--color-text-muted",
			"--color-text-label",
			"--color-surface-disabled",
			"--font-family-sans",
			"--shadow-base",
		}

		for _, variable := range expectedVariables {
			if !strings.Contains(body, variable) {
				t.Errorf("Response body should contain %s", variable)
			}
		}
	})

	t.Run("serves variables.css with :root selector", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/variables.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		if !strings.Contains(body, ":root") {
			t.Error("Response body should contain :root selector")
		}
	})

	t.Run("serves variables.css with dark mode support", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/variables.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		if !strings.Contains(body, "@media (prefers-color-scheme: dark)") {
			t.Error("Response body should contain dark mode media query")
		}
	})

	t.Run("returns 404 for non-existent CSS file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/nonexistent.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 for non-existent file, got %d", w.Code)
		}
	})
}
