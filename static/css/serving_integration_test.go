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

		if !strings.Contains(body, "[data-theme=\"dark\"]") {
			t.Error("Response body should contain [data-theme=\"dark\"] selector for dark mode")
		}
	})

	t.Run("serves spacing.css with correct status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/spacing.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("serves spacing.css with expected utilities", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/spacing.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		expectedUtilities := []string{
			".m-4",
			".p-4",
			".gap-4",
			".mt-4",
			".mb-4",
			".px-4",
			".py-4",
			".gap-x-4",
			".gap-y-4",
			".-m-4",
			".mx-auto",
		}

		for _, utility := range expectedUtilities {
			if !strings.Contains(body, utility) {
				t.Errorf("Response body should contain spacing utility %s", utility)
			}
		}
	})

	t.Run("serves forms.css with correct status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/forms.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("serves forms.css with expected form classes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/forms.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		expectedClasses := []string{
			".form-group",
			".form-label",
			".form-input",
			".form-textarea",
			".form-select",
			".form-checkbox",
			".form-radio",
			".form-error",
			".form-success",
			".form-help-text",
		}

		for _, class := range expectedClasses {
			if !strings.Contains(body, class) {
				t.Errorf("Response body should contain form class %s", class)
			}
		}
	})

	t.Run("serves buttons.css with correct status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/buttons.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("serves buttons.css with expected button classes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/buttons.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		expectedClasses := []string{
			".btn",
			".btn-primary",
			".btn-secondary",
			".btn-danger",
			".btn-ghost",
			".btn-sm",
			".btn-md",
			".btn-lg",
			".btn-loading",
			".btn-icon",
			".btn-group",
			".btn-block",
		}

		for _, class := range expectedClasses {
			if !strings.Contains(body, class) {
				t.Errorf("Response body should contain button class %s", class)
			}
		}
	})

	t.Run("serves loading_states.css with correct status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/loading_states.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("serves loading_states.css with expected classes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/loading_states.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		body := w.Body.String()

		expectedClasses := []string{
			".spinner",
			".skeleton",
			".loading-overlay",
			".progress",
			".progress-bar",
			".btn.loading",
			"@keyframes spin",
			"@keyframes loading",
		}

		for _, class := range expectedClasses {
			if !strings.Contains(body, class) {
				t.Errorf("Response body should contain loading state class %s", class)
			}
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
