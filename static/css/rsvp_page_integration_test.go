package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRSVPPageCSSIntegration(t *testing.T) {
	cssContent, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(cssContent)

	t.Run("serves correctly via HTTP", func(t *testing.T) {
		handler := http.FileServer(http.Dir("."))
		req := httptest.NewRequest("GET", "/rsvp_page.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/css") && !strings.Contains(contentType, "text/plain") {
			t.Errorf("Expected CSS content type, got %s", contentType)
		}
	})

	t.Run("integrates with variables.css", func(t *testing.T) {
		variablesUsed := []string{
			"var(--color-",
			"var(--spacing-",
			"var(--font-",
			"var(--radius-",
			"var(--shadow-",
		}

		for _, varPrefix := range variablesUsed {
			if !strings.Contains(css, varPrefix) {
				t.Errorf("CSS should use variables from variables.css (missing %s)", varPrefix)
			}
		}
	})

	t.Run("mobile-first responsive design", func(t *testing.T) {
		if !strings.Contains(css, "@media (min-width: 768px)") {
			t.Error("Missing tablet breakpoint")
		}

		if !strings.Contains(css, "@media (min-width: 1024px)") {
			t.Error("Missing desktop breakpoint")
		}

		mediaIndex := strings.Index(css, "@media")
		if mediaIndex == -1 {
			t.Fatal("No media queries found")
		}

		beforeMedia := css[:mediaIndex]
		mobileStyles := []string{
			".rsvp-page",
			".event-details",
			".rsvp-form",
		}

		for _, style := range mobileStyles {
			if !strings.Contains(beforeMedia, style) {
				t.Errorf("Mobile-first: %s should be defined before media queries", style)
			}
		}
	})

	t.Run("touch-friendly controls", func(t *testing.T) {
		if !strings.Contains(css, "min-height: 44px") && !strings.Contains(css, "min-height:44px") {
			t.Error("Missing 44px minimum height for touch targets")
		}
	})

	t.Run("loading states defined", func(t *testing.T) {
		if !strings.Contains(css, ".btn-loading") {
			t.Error("Missing .btn-loading class")
		}

		if !strings.Contains(css, "animation") {
			t.Error("Loading state should include animation")
		}
	})

	t.Run("error states defined", func(t *testing.T) {
		errorClasses := []string{
			".alert-error",
			".error-message",
		}

		for _, class := range errorClasses {
			if !strings.Contains(css, class) {
				t.Errorf("Missing error state class: %s", class)
			}
		}
	})

	t.Run("response options styled", func(t *testing.T) {
		if !strings.Contains(css, ".response-option") {
			t.Error("Missing .response-option class")
		}

		if !strings.Contains(css, "input[type=\"radio\"]:checked") {
			t.Error("Missing checked state for radio buttons")
		}
	})

	t.Run("plus-ones controls styled", func(t *testing.T) {
		plusOnesClasses := []string{
			".plus-ones-selector",
			".plus-ones-controls",
			".plus-ones-btn",
			".plus-ones-value",
		}

		for _, class := range plusOnesClasses {
			if !strings.Contains(css, class) {
				t.Errorf("Missing plus-ones class: %s", class)
			}
		}
	})

	t.Run("preference questions styled", func(t *testing.T) {
		questionClasses := []string{
			".preference-questions",
			".question-item",
		}

		for _, class := range questionClasses {
			if !strings.Contains(css, class) {
				t.Errorf("Missing question class: %s", class)
			}
		}
	})

	t.Run("no hardcoded colors", func(t *testing.T) {
		hardcodedColors := []string{
			"#2563eb",
			"#1d4ed8",
			"#dc3545",
			"#28a745",
			"rgb(37, 99, 235)",
		}

		for _, color := range hardcodedColors {
			if strings.Contains(css, color) {
				t.Errorf("Found hardcoded color %s, should use CSS variables", color)
			}
		}
	})

	t.Run("transitions defined", func(t *testing.T) {
		if !strings.Contains(css, "transition") {
			t.Error("CSS should include transitions for smooth interactions")
		}
	})

	t.Run("accessibility focus states", func(t *testing.T) {
		if !strings.Contains(css, ":focus") {
			t.Error("Missing focus states for accessibility")
		}

		if !strings.Contains(css, "outline") {
			t.Error("Focus states should include outline for accessibility")
		}
	})
}
