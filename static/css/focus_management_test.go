package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestFocusManagementFileExists(t *testing.T) {
	if _, err := os.Stat("focus_management.css"); os.IsNotExist(err) {
		t.Fatal("focus_management.css file does not exist")
	}
}

func TestFocusManagementValidCSS(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("focus_management.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestFocusManagementIndicators(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus") {
		t.Error("Missing :focus pseudo-class")
	}

	pattern := regexp.MustCompile(`:focus\s*\{[^}]*outline:\s*2px\s+solid`)
	if !pattern.MatchString(cssContent) {
		t.Error(":focus should have 2px solid outline")
	}

	pattern = regexp.MustCompile(`:focus\s*\{[^}]*outline-offset:\s*2px`)
	if !pattern.MatchString(cssContent) {
		t.Error(":focus should have outline-offset: 2px")
	}
}

func TestFocusManagementButtonStyles(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn:focus") {
		t.Error("Missing .btn:focus styles")
	}

	pattern := regexp.MustCompile(`\.btn:focus\s*\{[^}]*box-shadow:`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn:focus should have box-shadow for enhanced visibility")
	}
}

func TestFocusManagementFocusWithin(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus-within") {
		t.Error("Missing :focus-within pseudo-class for container focus")
	}
}

func TestFocusManagementFormGroupFocusWithin(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".form-group:focus-within") {
		t.Error("Missing .form-group:focus-within for form field containers")
	}
}

func TestFocusManagementContrast(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--color-border-focus)") {
		t.Error("Focus indicators should use --color-border-focus for consistent contrast")
	}
}

func TestFocusManagementCustomStyles(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".custom-focus:focus") {
		t.Error("Missing .custom-focus class for custom focus styles")
	}
}

func TestFocusManagementUsesVariables(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"var(--color-border-focus)",
		"var(--color-primary-",
		"var(--spacing-",
	}

	for _, varPrefix := range requiredVars {
		t.Run("uses_"+varPrefix, func(t *testing.T) {
			if !strings.Contains(cssContent, varPrefix) {
				t.Errorf("Focus management should use CSS variables with prefix: %s", varPrefix)
			}
		})
	}
}

func TestFocusManagementNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	hexColorPattern := regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
	matches := hexColorPattern.FindAllString(cssContent, -1)

	for _, match := range matches {
		if match != "#fff" && match != "#ffffff" && match != "#000" && match != "#000000" {
			t.Errorf("Focus management should not use hardcoded hex colors except pure black/white, found: %s", match)
		}
	}
}

func TestFocusManagementVisibleSupport(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus-visible") {
		t.Error("Should support :focus-visible for better mouse/keyboard UX")
	}
}

func TestFocusManagementNotFocusVisible(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ":focus:not(:focus-visible)") {
		t.Error("Should hide focus outline for mouse users with :focus:not(:focus-visible)")
	}
}

func TestFocusManagementTrapStyles(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if strings.Contains(cssContent, ".focus-trap") {
		t.Log("Includes focus trap container styles")
	}
}

func TestFocusManagementVisuallyHidden(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".sr-only") && !strings.Contains(cssContent, ".visually-hidden") {
		t.Error("Should have visually-hidden class for screen reader only content")
	}
}

func TestFocusManagementFileSize(t *testing.T) {
	info, err := os.Stat("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to stat focus_management.css: %v", err)
	}

	maxSize := int64(8 * 1024)
	if info.Size() > maxSize {
		t.Errorf("focus_management.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}
