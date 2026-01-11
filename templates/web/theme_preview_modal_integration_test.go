package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThemePreviewModalIntegration(t *testing.T) {
	t.Run("HTML template exists", func(t *testing.T) {
		path := filepath.Join("partials", "theme_preview_modal.html")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("theme_preview_modal.html does not exist at %s", path)
		}
	})

	t.Run("CSS file exists", func(t *testing.T) {
		path := filepath.Join("..", "..", "static", "css", "theme_preview_modal.css")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("theme_preview_modal.css does not exist at %s", path)
		}
	})

	t.Run("JavaScript file exists", func(t *testing.T) {
		path := filepath.Join("..", "..", "static", "js", "theme_preview_modal.js")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("theme_preview_modal.js does not exist at %s", path)
		}
	})
}

func TestThemePreviewModalHTMLCSSIntegration(t *testing.T) {
	htmlPath := filepath.Join("partials", "theme_preview_modal.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	cssPath := filepath.Join("..", "..", "static", "css", "theme_preview_modal.css")
	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS: %v", err)
	}

	html := string(htmlContent)
	css := string(cssContent)

	t.Run("HTML IDs have corresponding CSS", func(t *testing.T) {
		ids := []string{
			"theme-preview-modal",
			"theme-preview-frame",
			"preview-theme-toggle",
		}

		for _, id := range ids {
			if !strings.Contains(html, `id="`+id+`"`) {
				t.Errorf("HTML missing id: %s", id)
			}
			if !strings.Contains(css, "#"+id) {
				t.Errorf("CSS missing selector for id: %s", id)
			}
		}
	})

	t.Run("HTML classes have corresponding CSS", func(t *testing.T) {
		classes := []string{
			"modal-backdrop",
			"modal-container",
			"modal-header",
			"modal-body",
			"modal-footer",
			"modal-close",
			"modal-header-actions",
			"theme-icon",
		}

		for _, class := range classes {
			if !strings.Contains(html, `class="`+class) && !strings.Contains(html, class+`"`) {
				t.Errorf("HTML missing class: %s", class)
			}
			if !strings.Contains(css, "."+class) {
				t.Errorf("CSS missing selector for class: %s", class)
			}
		}
	})
}

func TestThemePreviewModalHTMLJSIntegration(t *testing.T) {
	htmlPath := filepath.Join("partials", "theme_preview_modal.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	jsPath := filepath.Join("..", "..", "static", "js", "theme_preview_modal.js")
	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read JS: %v", err)
	}

	html := string(htmlContent)
	js := string(jsContent)

	t.Run("JS queries HTML elements", func(t *testing.T) {
		elements := map[string]string{
			"theme-preview-modal":    "getElementById",
			"theme-preview-frame":    "getElementById",
			"preview-theme-toggle":   "getElementById",
			"select-previewed-theme": "getElementById",
		}

		for id, method := range elements {
			if !strings.Contains(html, `id="`+id+`"`) {
				t.Errorf("HTML missing element with id: %s", id)
			}
			if !strings.Contains(js, id) {
				t.Errorf("JS doesn't reference element: %s", id)
			}
			if !strings.Contains(js, method) {
				t.Errorf("JS doesn't use %s method", method)
			}
		}
	})

	t.Run("JS handles HTML events", func(t *testing.T) {
		eventHandlers := []string{
			"modal-close",
			"modal-backdrop",
		}

		for _, handler := range eventHandlers {
			if !strings.Contains(html, handler) {
				t.Errorf("HTML missing element: %s", handler)
			}
			if !strings.Contains(js, handler) {
				t.Errorf("JS doesn't reference: %s", handler)
			}
		}
	})
}

func TestThemePreviewModalAccessibilityIntegration(t *testing.T) {
	htmlPath := filepath.Join("partials", "theme_preview_modal.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	jsPath := filepath.Join("..", "..", "static", "js", "theme_preview_modal.js")
	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read JS: %v", err)
	}

	html := string(htmlContent)
	js := string(jsContent)

	t.Run("ARIA attributes in HTML match JS behavior", func(t *testing.T) {
		if !strings.Contains(html, `role="dialog"`) {
			t.Error("HTML missing role=dialog")
		}

		if !strings.Contains(html, `aria-modal="true"`) {
			t.Error("HTML missing aria-modal")
		}

		if !strings.Contains(html, `aria-labelledby=`) {
			t.Error("HTML missing aria-labelledby")
		}

		if !strings.Contains(js, "aria-live") {
			t.Error("JS should create aria-live announcements")
		}
	})

	t.Run("Focus management implemented", func(t *testing.T) {
		if !strings.Contains(js, "lastFocusedElement") {
			t.Error("JS should store last focused element")
		}

		if !strings.Contains(js, "focus()") {
			t.Error("JS should manage focus")
		}

		if !strings.Contains(js, "setupFocusTrap") {
			t.Error("JS should implement focus trap")
		}
	})

	t.Run("Keyboard navigation implemented", func(t *testing.T) {
		if !strings.Contains(js, "Escape") {
			t.Error("JS should handle Escape key")
		}

		if !strings.Contains(js, "Tab") {
			t.Error("JS should handle Tab key for focus trap")
		}
	})
}

func TestThemePreviewModalThemePickerIntegration(t *testing.T) {
	pickerPath := filepath.Join("..", "..", "static", "js", "theme_picker.js")
	pickerContent, err := os.ReadFile(pickerPath)
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	modalPath := filepath.Join("..", "..", "static", "js", "theme_preview_modal.js")
	modalContent, err := os.ReadFile(modalPath)
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	picker := string(pickerContent)
	modal := string(modalContent)

	t.Run("Theme picker dispatches preview event", func(t *testing.T) {
		if !strings.Contains(picker, "theme-preview-requested") {
			t.Error("Theme picker should dispatch theme-preview-requested event")
		}

		if !strings.Contains(picker, "CustomEvent") {
			t.Error("Theme picker should use CustomEvent")
		}
	})

	t.Run("Modal listens for preview event", func(t *testing.T) {
		if !strings.Contains(modal, "theme-preview-requested") {
			t.Error("Modal should listen for theme-preview-requested event")
		}

		if !strings.Contains(modal, "addEventListener") {
			t.Error("Modal should add event listener")
		}
	})

	t.Run("Modal dispatches selection event", func(t *testing.T) {
		if !strings.Contains(modal, "theme-selected") {
			t.Error("Modal should dispatch theme-selected event")
		}
	})

	t.Run("Theme picker handles selection event", func(t *testing.T) {
		if !strings.Contains(modal, "theme-selected") {
			t.Error("Modal integration should handle theme-selected event")
		}

		if !strings.Contains(modal, "themePicker") {
			t.Error("Modal should reference themePicker")
		}
	})
}

func TestThemePreviewModalResponsiveIntegration(t *testing.T) {
	cssPath := filepath.Join("..", "..", "static", "css", "theme_preview_modal.css")
	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read CSS: %v", err)
	}

	css := string(cssContent)

	t.Run("Mobile breakpoint defined", func(t *testing.T) {
		if !strings.Contains(css, "@media (max-width: 767px)") {
			t.Error("CSS should have mobile breakpoint")
		}
	})

	t.Run("Mobile optimizations present", func(t *testing.T) {
		mobileSection := css[strings.Index(css, "@media (max-width: 767px)"):]

		optimizations := []string{
			"100vw",
			"100vh",
			"border-radius: 0",
			"flex-direction: column",
		}

		for _, opt := range optimizations {
			if !strings.Contains(mobileSection, opt) {
				t.Errorf("Mobile section missing optimization: %s", opt)
			}
		}
	})
}
