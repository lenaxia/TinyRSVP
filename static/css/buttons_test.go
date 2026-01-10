package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestButtonsFileExists(t *testing.T) {
	if _, err := os.Stat("buttons.css"); os.IsNotExist(err) {
		t.Fatal("buttons.css file does not exist")
	}
}

func TestButtonsValidCSS(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("buttons.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestButtonsBaseClass(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn") {
		t.Error("Missing base .btn class")
	}

	requiredProperties := []string{
		"display:",
		"align-items:",
		"justify-content:",
		"padding:",
		"font-size:",
		"font-weight:",
		"border-radius:",
		"border:",
		"cursor:",
		"transition:",
		"min-height:",
	}

	for _, prop := range requiredProperties {
		t.Run("base_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf("Base .btn class should have %s property", prop)
			}
		})
	}
}

func TestButtonsMinimumTouchTarget(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	pattern := regexp.MustCompile(`\.btn\s*\{[^}]*min-height:\s*40px`)
	if !pattern.MatchString(cssContent) {
		t.Error("Base .btn class should have min-height: 40px")
	}
}

func TestButtonsPrimaryVariant(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-primary") {
		t.Error("Missing .btn-primary variant")
	}

	pattern := regexp.MustCompile(`\.btn-primary\s*\{[^}]*background-color:\s*var\(--color-primary-600\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-primary should use var(--color-primary-600) for background")
	}

	if !strings.Contains(cssContent, ".btn-primary:hover") {
		t.Error("Missing .btn-primary:hover state")
	}
}

func TestButtonsSecondaryVariant(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-secondary") {
		t.Error("Missing .btn-secondary variant")
	}

	pattern := regexp.MustCompile(`\.btn-secondary\s*\{[^}]*background-color:\s*var\(--color-gray-300\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-secondary should use var(--color-gray-300) for background")
	}

	if !strings.Contains(cssContent, ".btn-secondary:hover") {
		t.Error("Missing .btn-secondary:hover state")
	}
}

func TestButtonsDangerVariant(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-danger") {
		t.Error("Missing .btn-danger variant")
	}

	pattern := regexp.MustCompile(`\.btn-danger\s*\{[^}]*background-color:\s*var\(--color-error\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-danger should use var(--color-error) for background")
	}

	if !strings.Contains(cssContent, ".btn-danger:hover") {
		t.Error("Missing .btn-danger:hover state")
	}
}

func TestButtonsGhostVariant(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-ghost") {
		t.Error("Missing .btn-ghost variant")
	}

	pattern := regexp.MustCompile(`\.btn-ghost\s*\{[^}]*background-color:\s*transparent`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-ghost should have transparent background")
	}

	if !strings.Contains(cssContent, ".btn-ghost:hover") {
		t.Error("Missing .btn-ghost:hover state")
	}
}

func TestButtonsSizes(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	sizes := []string{".btn-sm", ".btn-md", ".btn-lg"}

	for _, size := range sizes {
		t.Run("size_"+size, func(t *testing.T) {
			if !strings.Contains(cssContent, size) {
				t.Errorf("Missing button size: %s", size)
			}
		})
	}
}

func TestButtonsDisabledState(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	disabledSelectors := []string{
		".btn:disabled",
		".btn[disabled]",
	}

	foundDisabled := false
	for _, selector := range disabledSelectors {
		if strings.Contains(cssContent, selector) {
			foundDisabled = true
			break
		}
	}

	if !foundDisabled {
		t.Error("Missing disabled state for buttons")
	}

	if !strings.Contains(cssContent, "opacity: 0.5") {
		t.Error("Disabled buttons should have opacity: 0.5")
	}

	if !strings.Contains(cssContent, "cursor: not-allowed") {
		t.Error("Disabled buttons should have cursor: not-allowed")
	}
}

func TestButtonsLoadingState(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-loading") {
		t.Error("Missing .btn-loading state")
	}

	pattern := regexp.MustCompile(`\.btn-loading\s*\{[^}]*cursor:\s*wait`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-loading should have cursor: wait")
	}
}

func TestButtonsFocusIndicator(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn:focus") {
		t.Error("Missing .btn:focus state")
	}

	pattern := regexp.MustCompile(`\.btn:focus\s*\{[^}]*outline:\s*2px\s+solid\s+var\(--color-border-focus\)`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn:focus should have 2px solid outline using var(--color-border-focus)")
	}

	pattern = regexp.MustCompile(`\.btn:focus\s*\{[^}]*outline-offset:\s*2px`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn:focus should have outline-offset: 2px")
	}
}

func TestButtonsIconButton(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-icon") {
		t.Error("Missing .btn-icon class")
	}

	pattern := regexp.MustCompile(`\.btn-icon\s*\{[^}]*padding:[^}]*\}`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-icon should have square padding")
	}
}

func TestButtonsGroup(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-group") {
		t.Error("Missing .btn-group class")
	}

	pattern := regexp.MustCompile(`\.btn-group\s*\{[^}]*display:\s*flex`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-group should use display: flex")
	}
}

func TestButtonsUsesVariables(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"var(--spacing-",
		"var(--font-size-",
		"var(--font-weight-",
		"var(--radius-",
		"var(--color-",
		"var(--transition-",
	}

	for _, varPrefix := range requiredVars {
		t.Run("uses_"+varPrefix, func(t *testing.T) {
			if !strings.Contains(cssContent, varPrefix) {
				t.Errorf("Buttons should use CSS variables with prefix: %s", varPrefix)
			}
		})
	}
}

func TestButtonsNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	hexColorPattern := regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
	if hexColorPattern.MatchString(cssContent) {
		t.Error("Buttons should not use hardcoded hex colors, use CSS variables instead")
	}

	rgbPattern := regexp.MustCompile(`rgb\(`)
	if rgbPattern.MatchString(cssContent) {
		t.Error("Buttons should not use hardcoded rgb colors, use CSS variables instead")
	}
}

func TestButtonsTransitions(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	pattern := regexp.MustCompile(`\.btn\s*\{[^}]*transition:[^}]*var\(--transition-`)
	if !pattern.MatchString(cssContent) {
		t.Error("Base .btn class should use CSS variable for transition")
	}
}

func TestButtonsHoverStates(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	variants := []string{"primary", "secondary", "danger", "ghost"}

	for _, variant := range variants {
		t.Run("hover_"+variant, func(t *testing.T) {
			hoverClass := ".btn-" + variant + ":hover"
			if !strings.Contains(cssContent, hoverClass) {
				t.Errorf("Missing hover state for %s", variant)
			}
		})
	}
}

func TestButtonsActiveStates(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	variants := []string{"primary", "secondary", "danger"}

	for _, variant := range variants {
		t.Run("active_"+variant, func(t *testing.T) {
			activeClass := ".btn-" + variant + ":active"
			if !strings.Contains(cssContent, activeClass) {
				t.Errorf("Missing active state for %s", variant)
			}
		})
	}
}

func TestButtonsFullWidth(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-block") {
		t.Error("Missing .btn-block class for full-width buttons")
	}

	pattern := regexp.MustCompile(`\.btn-block\s*\{[^}]*width:\s*100%`)
	if !pattern.MatchString(cssContent) {
		t.Error(".btn-block should have width: 100%")
	}
}
