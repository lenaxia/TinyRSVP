package css

import (
	"os"
	"strings"
	"testing"
)

func TestThemePreviewModalCSSExists(t *testing.T) {
	if _, err := os.Stat("theme_preview_modal.css"); os.IsNotExist(err) {
		t.Fatal("theme_preview_modal.css file does not exist")
	}
}

func TestThemePreviewModalCSSStructure(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		selector string
		category string
	}{
		{"has theme preview modal", "#theme-preview-modal", "structure"},
		{"has modal backdrop", ".modal-backdrop", "structure"},
		{"has modal container", ".modal-container", "structure"},
		{"has modal header", ".modal-header", "structure"},
		{"has modal body", ".modal-body", "structure"},
		{"has modal footer", ".modal-footer", "structure"},
		{"has preview frame", "#theme-preview-frame", "content"},
		{"has theme toggle", "#preview-theme-toggle", "interaction"},
		{"has close button", ".modal-close", "interaction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.selector) {
				t.Errorf("theme_preview_modal.css missing %s (selector: %s)", tt.name, tt.selector)
			}
		})
	}
}

func TestThemePreviewModalCSSPositioning(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	positionTests := []struct {
		name    string
		pattern string
	}{
		{"modal is fixed position", "position: fixed"},
		{"modal covers viewport", "width: 100%"},
		{"modal covers viewport height", "height: 100%"},
		{"modal has high z-index", "z-index"},
		{"backdrop is absolute", "position: absolute"},
	}

	for _, tt := range positionTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.pattern) {
				t.Errorf("theme_preview_modal.css missing positioning: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalCSSHiddenState(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "[hidden]") {
		t.Error("theme_preview_modal.css should handle hidden attribute")
	}

	if !strings.Contains(cssContent, "display: none") {
		t.Error("theme_preview_modal.css should hide modal when hidden")
	}
}

func TestThemePreviewModalCSSBackdrop(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	backdropTests := []struct {
		name    string
		pattern string
	}{
		{"has backdrop blur", "backdrop-filter"},
		{"has backdrop opacity", "rgba"},
		{"covers full area", "width: 100%"},
		{"covers full height", "height: 100%"},
	}

	for _, tt := range backdropTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.pattern) {
				t.Errorf("theme_preview_modal.css missing backdrop style: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalCSSIframe(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	iframeTests := []struct {
		name    string
		pattern string
	}{
		{"iframe full width", "width: 100%"},
		{"iframe has height", "height:"},
		{"iframe has border", "border:"},
		{"iframe has border radius", "border-radius"},
	}

	for _, tt := range iframeTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.pattern) {
				t.Errorf("theme_preview_modal.css missing iframe style: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalCSSResponsive(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media") {
		t.Error("theme_preview_modal.css should have responsive media queries")
	}

	if !strings.Contains(cssContent, "max-width: 767px") {
		t.Error("theme_preview_modal.css should have mobile breakpoint")
	}
}

func TestThemePreviewModalCSSMobileOptimizations(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	mobileTests := []struct {
		name    string
		pattern string
	}{
		{"full screen on mobile", "100vw"},
		{"full height on mobile", "100vh"},
		{"removes border radius on mobile", "border-radius: 0"},
		{"footer column layout", "flex-direction: column"},
	}

	for _, tt := range mobileTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.pattern) {
				t.Errorf("theme_preview_modal.css missing mobile optimization: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalCSSAccessibility(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	a11yTests := []struct {
		name    string
		pattern string
	}{
		{"has focus styles", ":focus"},
		{"has hover states", ":hover"},
		{"uses CSS variables", "var(--"},
		{"has transitions", "transition"},
	}

	for _, tt := range a11yTests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(cssContent, tt.pattern) {
				t.Errorf("theme_preview_modal.css missing accessibility feature: %s", tt.name)
			}
		})
	}
}

func TestThemePreviewModalCSSFlexLayout(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "display: flex") {
		t.Error("theme_preview_modal.css should use flexbox for layout")
	}

	if !strings.Contains(cssContent, "flex-direction: column") {
		t.Error("theme_preview_modal.css should use column flex direction for modal container")
	}
}

func TestThemePreviewModalCSSHeaderActions(t *testing.T) {
	content, err := os.ReadFile("theme_preview_modal.css")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".modal-header-actions") {
		t.Error("theme_preview_modal.css should style modal header actions")
	}

	if !strings.Contains(cssContent, "gap:") {
		t.Error("theme_preview_modal.css should use gap for spacing")
	}
}
