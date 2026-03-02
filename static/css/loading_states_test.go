package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestLoadingStatesFileExists(t *testing.T) {
	if _, err := os.Stat("loading_states.css"); os.IsNotExist(err) {
		t.Fatal("loading_states.css file does not exist")
	}
}

func TestLoadingStatesValidCSS(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("loading_states.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestLoadingStatesSpinnerAnimation(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@keyframes spin") {
		t.Error("Missing @keyframes spin animation")
	}

	pattern := regexp.MustCompile(`@keyframes\s+spin\s*\{[^}]*transform:\s*rotate\(360deg\)`)
	if !pattern.MatchString(cssContent) {
		t.Error("Spin animation should rotate 360deg")
	}
}

func TestLoadingStatesButtonLoading(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn.loading") || !strings.Contains(cssContent, ".btn-loading") {
		t.Error("Missing button loading state class")
	}

	pattern := regexp.MustCompile(`\.(btn\.loading|btn-loading)\s*\{[^}]*position:\s*relative`)
	if !pattern.MatchString(cssContent) {
		t.Error("Button loading state should have position: relative")
	}

	pattern = regexp.MustCompile(`\.(btn\.loading|btn-loading)\s*\{[^}]*color:\s*transparent`)
	if !pattern.MatchString(cssContent) {
		t.Error("Button loading state should have color: transparent to hide text")
	}
}

func TestLoadingStatesButtonLoadingSpinner(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn.loading::after") && !strings.Contains(cssContent, ".btn-loading::after") {
		t.Error("Missing button loading spinner pseudo-element")
	}

	requiredProperties := []string{
		"content:",
		"position: absolute",
		"border:",
		"border-radius:",
		"animation:",
	}

	for _, prop := range requiredProperties {
		if !strings.Contains(cssContent, prop) {
			t.Errorf("Button loading spinner should have %s property", prop)
		}
	}
}

func TestLoadingStatesButtonLoadingDisabled(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	pattern := regexp.MustCompile(`\.(btn\.loading|btn-loading)\s*\{[^}]*pointer-events:\s*none`)
	if !pattern.MatchString(cssContent) {
		t.Error("Button loading state should have pointer-events: none to disable interactions")
	}
}

func TestLoadingStatesSkeletonAnimation(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@keyframes loading") || !strings.Contains(cssContent, "@keyframes skeleton") {
		t.Error("Missing skeleton loading animation keyframes")
	}

	if !strings.Contains(cssContent, "background-position") {
		t.Error("Skeleton animation should use background-position")
	}
}

func TestLoadingStatesSkeletonClass(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".skeleton") {
		t.Error("Missing .skeleton class")
	}

	requiredProperties := []string{
		"background:",
		"background-size:",
		"animation:",
	}

	for _, prop := range requiredProperties {
		if !strings.Contains(cssContent, prop) {
			t.Errorf("Skeleton class should have %s property", prop)
		}
	}
}

func TestLoadingStatesSkeletonVariants(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	variants := []string{
		".skeleton-text",
		".skeleton-heading",
		".skeleton-avatar",
		".skeleton-button",
	}

	for _, variant := range variants {
		t.Run("variant_"+variant, func(t *testing.T) {
			if !strings.Contains(cssContent, variant) {
				t.Errorf("Missing skeleton variant: %s", variant)
			}
		})
	}
}

func TestLoadingStatesSpinnerClass(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".spinner") {
		t.Error("Missing .spinner class for standalone spinner")
	}

	requiredProperties := []string{
		"border:",
		"border-radius:",
		"animation:",
	}

	for _, prop := range requiredProperties {
		if !strings.Contains(cssContent, prop) {
			t.Errorf("Spinner class should have %s property", prop)
		}
	}
}

func TestLoadingStatesSpinnerSizes(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	sizes := []string{".spinner-sm", ".spinner-md", ".spinner-lg"}

	for _, size := range sizes {
		t.Run("size_"+size, func(t *testing.T) {
			if !strings.Contains(cssContent, size) {
				t.Errorf("Missing spinner size: %s", size)
			}
		})
	}
}

func TestLoadingStatesProgressBar(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".progress") {
		t.Error("Missing .progress class for progress bar container")
	}

	if !strings.Contains(cssContent, ".progress-bar") {
		t.Error("Missing .progress-bar class for progress bar fill")
	}

	pattern := regexp.MustCompile(`\.progress\s*\{[^}]*overflow:\s*hidden`)
	if !pattern.MatchString(cssContent) {
		t.Error("Progress container should have overflow: hidden")
	}
}

func TestLoadingStatesProgressBarTransition(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	pattern := regexp.MustCompile(`\.progress-bar\s*\{[^}]*transition:`)
	if !pattern.MatchString(cssContent) {
		t.Error("Progress bar should have smooth transition")
	}
}

func TestLoadingStatesARIASupport(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "[aria-busy") {
		t.Error("Should include styles for aria-busy attribute")
	}
}

func TestLoadingStatesOverlay(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".loading-overlay") {
		t.Error("Missing .loading-overlay class for full-page loading")
	}

	requiredProperties := []string{
		"position: fixed",
		"z-index:",
		"background:",
	}

	for _, prop := range requiredProperties {
		if !strings.Contains(cssContent, prop) {
			t.Errorf("Loading overlay should have %s property", prop)
		}
	}
}

func TestLoadingStatesUsesVariables(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"var(--color-",
		"var(--spacing-",
	}

	for _, varPrefix := range requiredVars {
		t.Run("uses_"+varPrefix, func(t *testing.T) {
			if !strings.Contains(cssContent, varPrefix) {
				t.Errorf("Loading states should use CSS variables with prefix: %s", varPrefix)
			}
		})
	}
}

func TestLoadingStatesNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	hexColorPattern := regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
	matches := hexColorPattern.FindAllString(cssContent, -1)

	allowedColors := map[string]bool{
		"#f0f0f0": true,
		"#e0e0e0": true,
	}

	for _, match := range matches {
		if !allowedColors[strings.ToLower(match)] {
			t.Errorf("Loading states should minimize hardcoded hex colors, found: %s (use CSS variables instead)", match)
		}
	}
}

func TestLoadingStatesReducedMotion(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media (prefers-reduced-motion: reduce)") {
		t.Error("Missing reduced motion media query for accessibility")
	}

	pattern := regexp.MustCompile(`@media\s*\(prefers-reduced-motion:\s*reduce\)\s*\{[^}]*animation:`)
	if !pattern.MatchString(cssContent) {
		t.Error("Reduced motion media query should disable or reduce animations")
	}
}

func TestLoadingStatesInlineSpinner(t *testing.T) {
	content, err := os.ReadFile("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to read loading_states.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".spinner-inline") {
		t.Error("Missing .spinner-inline class for inline loading indicators")
	}

	pattern := regexp.MustCompile(`\.spinner-inline\s*\{[^}]*display:\s*inline-block`)
	if !pattern.MatchString(cssContent) {
		t.Error("Inline spinner should have display: inline-block")
	}
}

func TestLoadingStatesFileSize(t *testing.T) {
	info, err := os.Stat("loading_states.css")
	if err != nil {
		t.Fatalf("Failed to stat loading_states.css: %v", err)
	}

	maxSize := int64(15 * 1024)
	if info.Size() > maxSize {
		t.Errorf("loading_states.css is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}
