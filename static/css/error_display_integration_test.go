package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestErrorDisplayIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	errorDisplayContent, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	variablesStr := string(variablesContent)
	errorDisplayStr := string(errorDisplayContent)

	requiredVars := []string{
		"--color-error",
		"--color-error-light",
		"--color-error-dark",
		"--color-success",
		"--color-success-light",
		"--color-success-dark",
		"--color-warning",
		"--color-warning-light",
		"--color-warning-darker",
		"--color-info",
		"--color-info-light",
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--radius-md",
		"--radius-base",
		"--font-size-sm",
		"--font-size-base",
		"--font-weight-semibold",
		"--line-height-normal",
		"--transition-fast",
		"--color-border-focus",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(errorDisplayStr, "var("+varName+")") {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Error display uses %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestErrorDisplayAccessibilityIntegration(t *testing.T) {
	errorDisplayContent, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	errorDisplayStr := string(errorDisplayContent)

	t.Run("dismiss_button_touch_target", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, "min-width: 44px") {
			t.Error("Dismiss button should have min-width: 44px for touch accessibility")
		}
		if !strings.Contains(errorDisplayStr, "min-height: 44px") {
			t.Error("Dismiss button should have min-height: 44px for touch accessibility")
		}
	})

	t.Run("focus_indicators_visible", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, ".alert-dismiss:focus") {
			t.Error("Dismiss button should have visible focus indicator")
		}
		pattern := regexp.MustCompile(`outline:\s*2px\s+solid`)
		if !pattern.MatchString(errorDisplayStr) {
			t.Error("Focus indicator should be at least 2px solid")
		}
	})

	t.Run("color_contrast_sufficient", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, "var(--color-error-dark)") {
			t.Error("Error alerts should use dark color for text to ensure contrast")
		}
		if !strings.Contains(errorDisplayStr, "var(--color-success-dark)") {
			t.Error("Success alerts should use dark color for text to ensure contrast")
		}
		if !strings.Contains(errorDisplayStr, "var(--color-warning-darker)") {
			t.Error("Warning alerts should use darker color for text to ensure contrast")
		}
	})
}

func TestErrorDisplayResponsiveIntegration(t *testing.T) {
	errorDisplayContent, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	errorDisplayStr := string(errorDisplayContent)

	t.Run("mobile_styles_present", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, "@media (max-width: 767px)") {
			t.Error("Should have mobile-specific styles for screens up to 767px")
		}
	})

	t.Run("tablet_desktop_styles_present", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, "@media (min-width: 768px)") {
			t.Error("Should have tablet/desktop styles for screens 768px and up")
		}
	})

	t.Run("touch_targets_maintained_mobile", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, "min-width: 44px") && !strings.Contains(errorDisplayStr, "min-height: 44px") {
			t.Error("Touch targets should be maintained at 44px minimum on mobile")
		}
	})
}

func TestErrorDisplayWithForms(t *testing.T) {
	errorDisplayContent, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Skip("forms.css not found, skipping form integration tests")
	}

	errorDisplayStr := string(errorDisplayContent)
	formsStr := string(formsContent)

	t.Run("field_error_styles_compatible", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, ".field-error") {
			t.Error("Should have .field-error class for inline field errors")
		}

		if strings.Contains(formsStr, ".form-group") && !strings.Contains(errorDisplayStr, ".field-error") {
			t.Error("Field errors should be styled to work with form groups")
		}
	})

	t.Run("form_error_summary_present", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, ".form-error-summary") {
			t.Error("Should have .form-error-summary class for form-level errors")
		}
	})

	t.Run("error_list_styles_present", func(t *testing.T) {
		if !strings.Contains(errorDisplayStr, ".form-error-list") {
			t.Error("Should have .form-error-list class for listing multiple errors")
		}
	})
}

func TestErrorDisplayFileSize(t *testing.T) {
	info, err := os.Stat("error_display.css")
	if err != nil {
		t.Fatalf("Failed to stat error_display.css: %v", err)
	}

	maxSize := int64(10 * 1024)
	if info.Size() > maxSize {
		t.Errorf("error_display.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}

func TestErrorDisplayNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if strings.Contains(cssContent, ": 8px") || strings.Contains(cssContent, ": 12px") {
		t.Error("Error display should not use hardcoded pixel values for spacing, use CSS variables instead")
	}

	if strings.Contains(cssContent, ": 1rem") || strings.Contains(cssContent, ": 1.5rem") {
		t.Error("Error display should not use hardcoded rem values for spacing, use CSS variables instead")
	}
}

func TestErrorDisplayVariantCompleteness(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	variants := []struct {
		name   string
		states []string
	}{
		{
			name:   "error",
			states: []string{".alert-error"},
		},
		{
			name:   "success",
			states: []string{".alert-success"},
		},
		{
			name:   "warning",
			states: []string{".alert-warning"},
		},
		{
			name:   "info",
			states: []string{".alert-info"},
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

func TestErrorDisplayMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	alertBaseIndex := strings.Index(cssContent, ".alert {")
	if alertBaseIndex == -1 {
		t.Fatal(".alert base styles not found")
	}

	mediaQueryIndex := strings.Index(cssContent, "@media (min-width: 768px)")
	if mediaQueryIndex == -1 {
		t.Fatal("Tablet media query not found")
	}

	if alertBaseIndex > mediaQueryIndex {
		t.Error("Base styles should come before media queries (mobile-first approach)")
	}
}

func TestErrorDisplayAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	accessibilityFeatures := []struct {
		name        string
		requirement string
	}{
		{"focus_indicator", ".alert-dismiss:focus"},
		{"hover_state", ".alert-dismiss:hover"},
		{"min_touch_target", "min-height: 44px"},
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

func TestErrorDisplayFlexboxLayout(t *testing.T) {
	content, err := os.ReadFile("error_display.css")
	if err != nil {
		t.Fatalf("Failed to read error_display.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "display: flex") {
		t.Error("Alert should use flexbox layout")
	}

	if !strings.Contains(cssContent, "flex: 1") {
		t.Error("Alert content should flex to fill available space")
	}

	if !strings.Contains(cssContent, "flex-shrink: 0") {
		t.Error("Alert icon should not shrink")
	}
}
