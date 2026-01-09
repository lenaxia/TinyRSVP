package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoadingStatesIntegrationServing(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/css/loading_states.css", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}
}

func TestLoadingStatesIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	loadingContent, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	combinedCSS := string(variablesContent) + "\n" + string(loadingContent)

	requiredVariables := []string{
		"--color-primary-600",
		"--color-gray-200",
		"--color-gray-300",
		"--color-gray-600",
		"--color-gray-700",
		"--color-success",
		"--color-warning",
		"--color-error",
		"--spacing-2",
		"--spacing-4",
		"--radius-base",
		"--radius-md",
		"--radius-lg",
		"--radius-full",
		"--transition-base",
		"--z-index-modal",
	}

	for _, variable := range requiredVariables {
		t.Run("variable_"+variable, func(t *testing.T) {
			if !strings.Contains(combinedCSS, variable) {
				t.Errorf("Combined CSS should contain variable: %s", variable)
			}
		})
	}
}

func TestLoadingStatesIntegrationWithButtons(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	loadingContent, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	buttonsCSS := string(buttonsContent)
	loadingCSS := string(loadingContent)

	if !strings.Contains(buttonsCSS, ".btn") {
		t.Error("buttons.css should define .btn class")
	}

	if !strings.Contains(loadingCSS, ".btn.loading") && !strings.Contains(loadingCSS, ".btn-loading") {
		t.Error("loading_states.css should define button loading states")
	}

	buttonVariants := []string{".btn-primary", ".btn-secondary", ".btn-danger", ".btn-ghost"}
	for _, variant := range buttonVariants {
		t.Run("button_variant_"+variant, func(t *testing.T) {
			if !strings.Contains(buttonsCSS, variant) {
				t.Errorf("buttons.css should define %s", variant)
			}
			if !strings.Contains(loadingCSS, variant+".loading") && !strings.Contains(loadingCSS, variant+".btn-loading") {
				t.Errorf("loading_states.css should define loading state for %s", variant)
			}
		})
	}
}

func TestLoadingStatesIntegrationAnimationPerformance(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "transform") {
		t.Error("Animations should use transform for better performance")
	}

	if strings.Contains(cssContent, "@keyframes") && strings.Contains(cssContent, "left:") {
		lines := strings.Split(cssContent, "\n")
		inKeyframes := false
		for _, line := range lines {
			if strings.Contains(line, "@keyframes") {
				inKeyframes = true
			}
			if inKeyframes && strings.Contains(line, "}") && !strings.Contains(line, "{") {
				inKeyframes = false
			}
			if inKeyframes && strings.Contains(line, "left:") && !strings.Contains(line, "/*") {
				t.Error("Avoid animating 'left' property in keyframes, use 'transform' instead for better performance")
				break
			}
		}
	}
}

func TestLoadingStatesIntegrationAccessibility(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	accessibilityFeatures := []struct {
		name    string
		pattern string
	}{
		{"ARIA busy support", "[aria-busy"},
		{"Reduced motion support", "@media (prefers-reduced-motion: reduce)"},
		{"Pointer events disabled during loading", "pointer-events: none"},
		{"Visual loading indicator", "animation:"},
	}

	for _, feature := range accessibilityFeatures {
		t.Run("accessibility_"+feature.name, func(t *testing.T) {
			if !strings.Contains(cssContent, feature.pattern) {
				t.Errorf("Missing accessibility feature: %s (pattern: %s)", feature.name, feature.pattern)
			}
		})
	}
}

func TestLoadingStatesIntegrationResponsive(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media") {
		t.Error("Loading states should include responsive styles")
	}

	if !strings.Contains(cssContent, "max-width") && !strings.Contains(cssContent, "min-width") {
		t.Error("Loading states should include breakpoint-based styles")
	}
}

func TestLoadingStatesIntegrationDarkMode(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "prefers-color-scheme: dark") {
		t.Error("Loading states should include dark mode support")
	}

	if !strings.Contains(cssContent, ".skeleton") {
		t.Error("Skeleton screens should be defined")
	}
}

func TestLoadingStatesIntegrationNoConflicts(t *testing.T) {
	files := []string{
		"variables.css",
		"buttons.css",
		"forms.css",
		"loading_states.css",
	}

	var combinedCSS strings.Builder
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Logf("Skipping %s: %v", file, err)
			continue
		}
		combinedCSS.Write(content)
		combinedCSS.WriteString("\n")
	}

	combined := combinedCSS.String()

	if strings.Count(combined, "@keyframes spin") > 1 {
		t.Error("Duplicate @keyframes spin definition found")
	}

	if strings.Count(combined, "@keyframes loading") > 1 {
		t.Error("Duplicate @keyframes loading definition found")
	}
}

func TestLoadingStatesIntegrationSpinnerInButton(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "::after") {
		t.Error("Button loading spinner should use ::after pseudo-element")
	}

	if !strings.Contains(cssContent, "position: absolute") {
		t.Error("Button loading spinner should be absolutely positioned")
	}

	if !strings.Contains(cssContent, "top: 50%") || !strings.Contains(cssContent, "left: 50%") {
		t.Error("Button loading spinner should be centered")
	}
}

func TestLoadingStatesIntegrationSkeletonScreens(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	skeletonTypes := []string{
		".skeleton-text",
		".skeleton-heading",
		".skeleton-avatar",
		".skeleton-button",
	}

	for _, skeletonType := range skeletonTypes {
		t.Run("skeleton_type_"+skeletonType, func(t *testing.T) {
			if !strings.Contains(cssContent, skeletonType) {
				t.Errorf("Missing skeleton type: %s", skeletonType)
			}
		})
	}

	if !strings.Contains(cssContent, "linear-gradient") {
		t.Error("Skeleton screens should use linear-gradient for shimmer effect")
	}
}

func TestLoadingStatesIntegrationProgressBar(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".progress") {
		t.Error("Missing progress bar container")
	}

	if !strings.Contains(cssContent, ".progress-bar") {
		t.Error("Missing progress bar fill")
	}

	progressVariants := []string{
		".progress-bar-success",
		".progress-bar-warning",
		".progress-bar-error",
	}

	for _, variant := range progressVariants {
		t.Run("progress_variant_"+variant, func(t *testing.T) {
			if !strings.Contains(cssContent, variant) {
				t.Errorf("Missing progress bar variant: %s", variant)
			}
		})
	}
}

func TestLoadingStatesIntegrationOverlay(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".loading-overlay") {
		t.Error("Missing loading overlay class")
	}

	requiredOverlayProperties := []string{
		"position: fixed",
		"z-index:",
		"display: flex",
		"align-items: center",
		"justify-content: center",
	}

	for _, prop := range requiredOverlayProperties {
		if !strings.Contains(cssContent, prop) {
			t.Errorf("Loading overlay should have property: %s", prop)
		}
	}
}
