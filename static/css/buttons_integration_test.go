package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestButtonsIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	variablesStr := string(variablesContent)
	buttonsStr := string(buttonsContent)

	requiredVars := []string{
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--spacing-8",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-weight-medium",
		"--line-height-normal",
		"--radius-md",
		"--color-primary-50",
		"--color-primary-100",
		"--color-primary-200",
		"--color-primary-600",
		"--color-primary-700",
		"--color-primary-800",
		"--color-gray-200",
		"--color-gray-300",
		"--color-gray-400",
		"--color-error",
		"--color-error-dark",
		"--color-text-primary",
		"--color-border-focus",
		"--transition-base",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(buttonsStr, "var("+varName+")") {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Buttons use %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestButtonsHTTPServing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/buttons.css" {
			content, err := os.ReadFile("buttons.css")
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/css")
			w.Write(content)
		} else {
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/static/css/buttons.css")
	if err != nil {
		t.Fatalf("Failed to fetch buttons.css: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}
}

func TestButtonsResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	breakpoints := []struct {
		name  string
		query string
	}{
		{"tablet", "@media (min-width: 768px)"},
	}

	for _, bp := range breakpoints {
		t.Run(bp.name+"_breakpoint", func(t *testing.T) {
			if !strings.Contains(cssContent, bp.query) {
				t.Errorf("Missing %s breakpoint: %s", bp.name, bp.query)
			}
		})
	}
}

func TestButtonsConsistencyWithExistingCSS(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	buttonsStr := string(buttonsContent)
	variablesStr := string(variablesContent)

	colorVars := []string{
		"--color-primary-600",
		"--color-primary-700",
		"--color-gray-200",
		"--color-error",
		"--color-border-focus",
	}

	for _, colorVar := range colorVars {
		t.Run("color_consistency_"+colorVar, func(t *testing.T) {
			if strings.Contains(buttonsStr, "var("+colorVar+")") {
				if !strings.Contains(variablesStr, colorVar+":") {
					t.Errorf("Buttons use %s but it's not defined in variables.css", colorVar)
				}
			}
		})
	}
}

func TestButtonsFileSize(t *testing.T) {
	info, err := os.Stat("buttons.css")
	if err != nil {
		t.Fatalf("Failed to stat buttons.css: %v", err)
	}

	maxSize := int64(15 * 1024)
	if info.Size() > maxSize {
		t.Errorf("buttons.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}

func TestButtonsNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if strings.Contains(cssContent, ": 8px") || strings.Contains(cssContent, ": 12px") {
		t.Error("Buttons should not use hardcoded pixel values for spacing, use CSS variables instead")
	}

	if strings.Contains(cssContent, ": 1rem") || strings.Contains(cssContent, ": 1.5rem") {
		t.Error("Buttons should not use hardcoded rem values for spacing, use CSS variables instead")
	}
}

func TestButtonsVariantCompleteness(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	variants := []struct {
		name   string
		states []string
	}{
		{
			name:   "primary",
			states: []string{".btn-primary", ".btn-primary:hover", ".btn-primary:active"},
		},
		{
			name:   "secondary",
			states: []string{".btn-secondary", ".btn-secondary:hover", ".btn-secondary:active"},
		},
		{
			name:   "danger",
			states: []string{".btn-danger", ".btn-danger:hover", ".btn-danger:active"},
		},
		{
			name:   "ghost",
			states: []string{".btn-ghost", ".btn-ghost:hover", ".btn-ghost:active"},
		},
	}

	for _, variant := range variants {
		t.Run("variant_"+variant.name, func(t *testing.T) {
			for _, state := range variant.states {
				if !strings.Contains(cssContent, state) {
					t.Errorf("Missing state for %s variant: %s", variant.name, state)
				}
			}
		})
	}
}

func TestButtonsSizeCompleteness(t *testing.T) {
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

			if !strings.Contains(cssContent, "height:") {
				t.Errorf("Button size %s should define height", size)
			}
		})
	}
}

func TestButtonsMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	btnBaseIndex := strings.Index(cssContent, ".btn {")
	if btnBaseIndex == -1 {
		t.Fatal(".btn base styles not found")
	}

	mediaQueryIndex := strings.Index(cssContent, "@media (min-width: 768px)")
	if mediaQueryIndex == -1 {
		t.Fatal("Tablet media query not found")
	}

	if btnBaseIndex > mediaQueryIndex {
		t.Error("Base styles should come before media queries (mobile-first approach)")
	}
}

func TestButtonsAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	accessibilityFeatures := []struct {
		name        string
		requirement string
	}{
		{"focus_indicator", ".btn:focus"},
		{"disabled_state", ".btn:disabled"},
		{"min_touch_target", "height: 40px"},
		{"cursor_feedback", "cursor:"},
	}

	for _, feature := range accessibilityFeatures {
		t.Run("accessibility_"+feature.name, func(t *testing.T) {
			if !strings.Contains(cssContent, feature.requirement) {
				t.Errorf("Missing accessibility feature %s: %s", feature.name, feature.requirement)
			}
		})
	}
}

func TestButtonsIntegrationWithTypography(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	typographyContent, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	buttonsStr := string(buttonsContent)
	typographyStr := string(typographyContent)

	sharedTypographyVars := []string{
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-weight-medium",
		"--line-height-normal",
	}

	for _, varName := range sharedTypographyVars {
		t.Run("shared_typography_"+varName, func(t *testing.T) {
			usedInButtons := strings.Contains(buttonsStr, "var("+varName+")")
			definedInTypography := strings.Contains(typographyStr, varName+":")

			if usedInButtons && !definedInTypography {
				t.Logf("Variable %s is used in buttons but not defined in typography.css", varName)
			}
		})
	}
}

func TestButtonsIntegrationWithSpacing(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	buttonsStr := string(buttonsContent)
	spacingStr := string(spacingContent)

	sharedSpacingVars := []string{
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--spacing-8",
	}

	for _, varName := range sharedSpacingVars {
		t.Run("shared_spacing_"+varName, func(t *testing.T) {
			usedInButtons := strings.Contains(buttonsStr, "var("+varName+")")
			usedInSpacing := strings.Contains(spacingStr, "var("+varName+")")

			if usedInButtons && usedInSpacing {
				t.Logf("Variable %s is shared between buttons and spacing systems", varName)
			}
		})
	}
}

func TestButtonsIntegrationWithColors(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	colorsContent, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	buttonsStr := string(buttonsContent)
	colorsStr := string(colorsContent)

	sharedColorVars := []string{
		"--color-primary-600",
		"--color-primary-700",
		"--color-error",
		"--color-gray-200",
	}

	for _, varName := range sharedColorVars {
		t.Run("shared_color_"+varName, func(t *testing.T) {
			usedInButtons := strings.Contains(buttonsStr, "var("+varName+")")
			usedInColors := strings.Contains(colorsStr, "var("+varName+")")

			if usedInButtons && usedInColors {
				t.Logf("Variable %s is shared between buttons and colors systems", varName)
			}
		})
	}
}

func TestButtonsLoadingAnimation(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@keyframes") {
		t.Error("Loading state should include keyframe animation")
	}

	if !strings.Contains(cssContent, "animation:") {
		t.Error("Loading state should use animation property")
	}

	if !strings.Contains(cssContent, ".btn-loading::after") {
		t.Error("Loading state should use ::after pseudo-element for spinner")
	}
}

func TestButtonsGroupLayout(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	groupClasses := []string{".btn-group", ".btn-group-vertical"}

	for _, class := range groupClasses {
		t.Run("group_"+class, func(t *testing.T) {
			if !strings.Contains(cssContent, class) {
				t.Errorf("Missing button group class: %s", class)
			}

			if !strings.Contains(cssContent, "display: flex") {
				t.Error("Button groups should use flexbox layout")
			}
		})
	}
}

func TestButtonsIconSupport(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".btn-icon") {
		t.Error("Missing icon button support")
	}

	iconSizes := []string{".btn-icon.btn-sm", ".btn-icon.btn-lg"}
	for _, size := range iconSizes {
		t.Run("icon_size_"+size, func(t *testing.T) {
			if !strings.Contains(cssContent, size) {
				t.Errorf("Missing icon button size: %s", size)
			}
		})
	}
}

func TestButtonsTransitionConsistency(t *testing.T) {
	content, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--transition-base)") {
		t.Error("Buttons should use consistent transition variable")
	}

	if strings.Contains(cssContent, "transition: 0.2s") || strings.Contains(cssContent, "transition: 200ms") {
		t.Error("Buttons should not use hardcoded transition values, use CSS variables instead")
	}
}
