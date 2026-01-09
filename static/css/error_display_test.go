package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestErrorDisplayFileExists(t *testing.T) {
	if _, err := os.Stat("error_display.css"); os.IsNotExist(err) {
		t.Fatal("error_display.css file does not exist")
	}
}

func TestErrorDisplayValidCSS(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("error_display.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestAlertBaseClass(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert") {
		t.Error("Missing base .alert class")
	}

	requiredProperties := []string{
		"padding:",
		"border-radius:",
		"margin-bottom:",
		"display:",
		"align-items:",
		"gap:",
	}

	for _, prop := range requiredProperties {
		t.Run("base_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf("Base .alert class should have %s property", prop)
			}
		})
	}
}

func TestAlertErrorVariant(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-error") {
		t.Error("Missing .alert-error variant")
	}

	pattern := regexp.MustCompile(`\.alert-error\s*\{[^}]*background-color:\s*var\(--color-error-light\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-error should use var(--color-error-light) for background")
	}

	pattern = regexp.MustCompile(`\.alert-error\s*\{[^}]*border-left:\s*4px\s+solid\s+var\(--color-error\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-error should have 4px solid left border using var(--color-error)")
	}

	pattern = regexp.MustCompile(`\.alert-error\s*\{[^}]*color:\s*var\(--color-error-dark\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-error should use var(--color-error-dark) for text color")
	}
}

func TestAlertSuccessVariant(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-success") {
		t.Error("Missing .alert-success variant")
	}

	pattern := regexp.MustCompile(`\.alert-success\s*\{[^}]*background-color:\s*var\(--color-success-light\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-success should use var(--color-success-light) for background")
	}

	pattern = regexp.MustCompile(`\.alert-success\s*\{[^}]*border-left:\s*4px\s+solid\s+var\(--color-success\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-success should have 4px solid left border using var(--color-success)")
	}

	pattern = regexp.MustCompile(`\.alert-success\s*\{[^}]*color:\s*var\(--color-success-dark\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-success should use var(--color-success-dark) for text color")
	}
}

func TestAlertWarningVariant(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-warning") {
		t.Error("Missing .alert-warning variant")
	}

	pattern := regexp.MustCompile(`\.alert-warning\s*\{[^}]*background-color:\s*var\(--color-warning-light\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-warning should use var(--color-warning-light) for background")
	}

	pattern = regexp.MustCompile(`\.alert-warning\s*\{[^}]*border-left:\s*4px\s+solid\s+var\(--color-warning\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-warning should have 4px solid left border using var(--color-warning)")
	}

	pattern = regexp.MustCompile(`\.alert-warning\s*\{[^}]*color:\s*var\(--color-warning-darker\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-warning should use var(--color-warning-darker) for text color")
	}
}

func TestAlertInfoVariant(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-info") {
		t.Error("Missing .alert-info variant")
	}

	pattern := regexp.MustCompile(`\.alert-info\s*\{[^}]*background-color:\s*var\(--color-info-light\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-info should use var(--color-info-light) for background")
	}

	pattern = regexp.MustCompile(`\.alert-info\s*\{[^}]*border-left:\s*4px\s+solid\s+var\(--color-info\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-info should have 4px solid left border using var(--color-info)")
	}
}

func TestAlertIcon(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-icon") {
		t.Error("Missing .alert-icon class")
	}

	pattern := regexp.MustCompile(`\.alert-icon\s*\{[^}]*flex-shrink:\s*0`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-icon should have flex-shrink: 0")
	}
}

func TestAlertContent(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-content") {
		t.Error("Missing .alert-content class")
	}

	pattern := regexp.MustCompile(`\.alert-content\s*\{[^}]*flex:\s*1`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-content should have flex: 1")
	}
}

func TestAlertTitle(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-title") {
		t.Error("Missing .alert-title class")
	}

	if !strings.Contains(cssContent, "font-weight:") {
		t.Error(".alert-title should have font-weight property")
	}
}

func TestAlertMessage(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-message") {
		t.Error("Missing .alert-message class")
	}
}

func TestAlertDismissButton(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".alert-dismiss") {
		t.Error("Missing .alert-dismiss class")
	}

	pattern := regexp.MustCompile(`\.alert-dismiss\s*\{[^}]*cursor:\s*pointer`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-dismiss should have cursor: pointer")
	}

	if !strings.Contains(cssContent, ".alert-dismiss:hover") {
		t.Error("Missing .alert-dismiss:hover state")
	}

	if !strings.Contains(cssContent, ".alert-dismiss:focus") {
		t.Error("Missing .alert-dismiss:focus state")
	}
}

func TestFormErrorSummary(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".form-error-summary") {
		t.Error("Missing .form-error-summary class")
	}
}

func TestFormErrorList(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".form-error-list") {
		t.Error("Missing .form-error-list class")
	}

	pattern := regexp.MustCompile(`\.form-error-list\s*\{[^}]*list-style:\s*none`)
	if !pattern.MatchString(cssContent) {
		t.Error(".form-error-list should have list-style: none")
	}
}

func TestFieldError(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".field-error") {
		t.Error("Missing .field-error class")
	}

	pattern := regexp.MustCompile(`\.field-error\s*\{[^}]*color:\s*var\(--color-error\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".field-error should use var(--color-error) for color")
	}
}

func TestErrorDisplayUsesVariables(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"var(--color-error",
		"var(--color-success",
		"var(--color-warning",
		"var(--color-info",
		"var(--spacing-",
	}

	for _, varPrefix := range requiredVars {
		t.Run("uses_"+varPrefix, func(t *testing.T) {
			if !strings.Contains(cssContent, varPrefix) {
				t.Errorf("Error display should use CSS variables with prefix: %s", varPrefix)
			}
		})
	}
}

func TestErrorDisplayNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	hexColorPattern := regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
	if hexColorPattern.MatchString(cssContent) {
		t.Error("Error display should not use hardcoded hex colors, use CSS variables instead")
	}

	rgbPattern := regexp.MustCompile(`rgb\(`)
	if rgbPattern.MatchString(cssContent) {
		t.Error("Error display should not use hardcoded rgb colors, use CSS variables instead")
	}
}

func TestErrorDisplayFocusIndicators(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus") {
		t.Error("Error display should have focus indicators")
	}

	pattern := regexp.MustCompile(`\.alert-dismiss:focus\s*\{[^}]*outline:\s*2px\s+solid`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert-dismiss:focus should have 2px solid outline")
	}
}

func TestErrorDisplayResponsive(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media") {
		t.Error("Error display should include responsive media queries")
	}
}

func TestAlertFlexLayout(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	pattern := regexp.MustCompile(`\.alert\s*\{[^}]*display:\s*flex`)
	if !pattern.MatchString(cssContent) {
		t.Error(".alert should use display: flex")
	}
}

func TestAlertDismissMinimumTouchTarget(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "min-width:") || !strings.Contains(cssContent, "min-height:") {
		t.Error(".alert-dismiss should have minimum touch target size for mobile accessibility")
	}
}
