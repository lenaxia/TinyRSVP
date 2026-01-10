package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestFormsCSS(t *testing.T) {
	content, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	css := string(content)

	t.Run("FormGroupClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-group") {
			t.Error("Missing .form-group class")
		}
		if !strings.Contains(css, "margin-bottom: var(--spacing-4)") {
			t.Error(".form-group should use --spacing-4 for margin-bottom")
		}
	})

	t.Run("FormLabelClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-label") {
			t.Error("Missing .form-label class")
		}
		if !strings.Contains(css, "display: block") {
			t.Error(".form-label should have display: block")
		}
		if !strings.Contains(css, "margin-bottom: var(--spacing-2)") {
			t.Error(".form-label should use --spacing-2 for margin-bottom")
		}
		if !strings.Contains(css, "font-weight: var(--font-weight-medium)") {
			t.Error(".form-label should use --font-weight-medium")
		}
	})

	t.Run("FormInputClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-input") {
			t.Error("Missing .form-input class")
		}
		if !strings.Contains(css, "width: 100%") {
			t.Error(".form-input should have width: 100%")
		}
		if !strings.Contains(css, "padding: var(--spacing-2)") {
			t.Error(".form-input should use --spacing-2 for padding")
		}
		if !strings.Contains(css, "border: 1px solid var(--color-border)") {
			t.Error(".form-input should use --color-border")
		}
		if !strings.Contains(css, "border-radius: var(--radius-md)") {
			t.Error(".form-input should use --radius-md")
		}
		if !strings.Contains(css, "font-size: var(--font-size-base)") {
			t.Error(".form-input should use --font-size-base")
		}
	})

	t.Run("FormInputFocusState", func(t *testing.T) {
		if !strings.Contains(css, ".form-input:focus") {
			t.Error("Missing .form-input:focus state")
		}
		if !strings.Contains(css, "outline: 2px solid var(--color-border-focus)") {
			t.Error(".form-input:focus should have 2px outline with --color-border-focus")
		}
		if !strings.Contains(css, "outline-offset: 2px") {
			t.Error(".form-input:focus should have outline-offset: 2px")
		}
	})

	t.Run("FormInputErrorState", func(t *testing.T) {
		if !strings.Contains(css, ".form-input.error") {
			t.Error("Missing .form-input.error state")
		}
		if !strings.Contains(css, "border-color: var(--color-error)") {
			t.Error(".form-input.error should use --color-error for border")
		}
	})

	t.Run("FormInputDisabledState", func(t *testing.T) {
		if !strings.Contains(css, ".form-input:disabled") {
			t.Error("Missing .form-input:disabled state")
		}
		if !strings.Contains(css, "background-color: var(--color-surface-disabled)") {
			t.Error(".form-input:disabled should use --color-surface-disabled")
		}
		if !strings.Contains(css, "cursor: not-allowed") {
			t.Error(".form-input:disabled should have cursor: not-allowed")
		}
	})

	t.Run("FormErrorClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-error") {
			t.Error("Missing .form-error class")
		}
		if !strings.Contains(css, "color: var(--color-error)") {
			t.Error(".form-error should use --color-error")
		}
		if !strings.Contains(css, "font-size: var(--font-size-sm)") {
			t.Error(".form-error should use --font-size-sm")
		}
		if !strings.Contains(css, "margin-top: var(--spacing-1)") {
			t.Error(".form-error should use --spacing-1 for margin-top")
		}
	})

	t.Run("FormTextareaClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-textarea") {
			t.Error("Missing .form-textarea class")
		}
		if !strings.Contains(css, "min-height: 120px") {
			t.Error(".form-textarea should have min-height: 120px")
		}
		if !strings.Contains(css, "resize: vertical") {
			t.Error(".form-textarea should have resize: vertical")
		}
	})

	t.Run("FormSelectClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-select") {
			t.Error("Missing .form-select class")
		}
		if !strings.Contains(css, "appearance: none") {
			t.Error(".form-select should have appearance: none")
		}
		if !strings.Contains(css, "background-image:") {
			t.Error(".form-select should have background-image for dropdown arrow")
		}
		if !strings.Contains(css, "background-repeat: no-repeat") {
			t.Error(".form-select should have background-repeat: no-repeat")
		}
		if !strings.Contains(css, "background-position: right") {
			t.Error(".form-select should have background-position: right")
		}
	})

	t.Run("FormCheckboxClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-checkbox") {
			t.Error("Missing .form-checkbox class")
		}
		if !strings.Contains(css, "width: 20px") {
			t.Error(".form-checkbox should have width: 20px")
		}
		if !strings.Contains(css, "height: 20px") {
			t.Error(".form-checkbox should have height: 20px")
		}
		if !strings.Contains(css, "cursor: pointer") {
			t.Error(".form-checkbox should have cursor: pointer")
		}
	})

	t.Run("FormRadioClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-radio") {
			t.Error("Missing .form-radio class")
		}
		if !strings.Contains(css, "width: 20px") {
			t.Error(".form-radio should have width: 20px")
		}
		if !strings.Contains(css, "height: 20px") {
			t.Error(".form-radio should have height: 20px")
		}
		if !strings.Contains(css, "cursor: pointer") {
			t.Error(".form-radio should have cursor: pointer")
		}
	})

	t.Run("FormCheckWrapperClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-check-wrapper") {
			t.Error("Missing .form-check-wrapper class")
		}
		if !strings.Contains(css, "display: flex") {
			t.Error(".form-check-wrapper should have display: flex")
		}
		if !strings.Contains(css, "align-items: center") {
			t.Error(".form-check-wrapper should have align-items: center")
		}
		if !strings.Contains(css, "gap: var(--spacing-2)") {
			t.Error(".form-check-wrapper should use --spacing-2 for gap")
		}
	})

	t.Run("FormCheckLabelClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-check-label") {
			t.Error("Missing .form-check-label class")
		}
		if !strings.Contains(css, "cursor: pointer") {
			t.Error(".form-check-label should have cursor: pointer")
		}
		if !strings.Contains(css, "user-select: none") {
			t.Error(".form-check-label should have user-select: none")
		}
	})

	t.Run("FormHelpTextClass", func(t *testing.T) {
		if !strings.Contains(css, ".form-help-text") {
			t.Error("Missing .form-help-text class")
		}
		if !strings.Contains(css, "font-size: var(--font-size-sm)") {
			t.Error(".form-help-text should use --font-size-sm")
		}
		if !strings.Contains(css, "color: var(--color-text-secondary)") {
			t.Error(".form-help-text should use --color-text-secondary")
		}
		if !strings.Contains(css, "margin-top: var(--spacing-1)") {
			t.Error(".form-help-text should use --spacing-1 for margin-top")
		}
	})

	t.Run("TouchFriendlyMinimumSize", func(t *testing.T) {
		checkboxMatch := regexp.MustCompile(`\.form-checkbox[^}]*height:\s*20px`).FindString(css)
		radioMatch := regexp.MustCompile(`\.form-radio[^}]*height:\s*20px`).FindString(css)
		
		if checkboxMatch == "" {
			t.Error(".form-checkbox should have minimum 20px height for touch targets")
		}
		if radioMatch == "" {
			t.Error(".form-radio should have minimum 20px height for touch targets")
		}
	})

	t.Run("FormInputTransition", func(t *testing.T) {
		if !strings.Contains(css, "transition:") {
			t.Error("Form inputs should have transition for smooth state changes")
		}
	})

	t.Run("FormRequiredIndicator", func(t *testing.T) {
		if !strings.Contains(css, ".form-required") {
			t.Error("Missing .form-required class for required field indicator")
		}
		if !strings.Contains(css, "color: var(--color-error)") {
			t.Error(".form-required should use --color-error")
		}
	})
}

func TestFormsAccessibility(t *testing.T) {
	content, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	css := string(content)

	t.Run("FocusIndicators", func(t *testing.T) {
		focusSelectors := []string{
			".form-input:focus",
			".form-select:focus",
			".form-textarea:focus",
			".form-checkbox:focus",
			".form-radio:focus",
		}

		for _, selector := range focusSelectors {
			if !strings.Contains(css, selector) {
				t.Errorf("Missing focus state for %s", selector)
			}
		}
	})

	t.Run("DisabledStates", func(t *testing.T) {
		disabledSelectors := []string{
			".form-input:disabled",
			".form-select:disabled",
			".form-textarea:disabled",
			".form-checkbox:disabled",
			".form-radio:disabled",
		}

		for _, selector := range disabledSelectors {
			if !strings.Contains(css, selector) {
				t.Errorf("Missing disabled state for %s", selector)
			}
		}
	})

	t.Run("VisibleOutlines", func(t *testing.T) {
		if !strings.Contains(css, "outline:") {
			t.Error("Form elements should have visible focus outlines")
		}
		if !strings.Contains(css, "outline-offset:") {
			t.Error("Form elements should have outline-offset for better visibility")
		}
	})
}

func TestFormsResponsive(t *testing.T) {
	content, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	css := string(content)

	t.Run("MobileFirstApproach", func(t *testing.T) {
		if !strings.Contains(css, ".form-input") {
			t.Error("Base form styles should be defined for mobile-first approach")
		}
	})

	t.Run("ResponsiveBreakpoints", func(t *testing.T) {
		if strings.Contains(css, "@media") {
			if !strings.Contains(css, "min-width") {
				t.Error("Media queries should use min-width for mobile-first approach")
			}
		}
	})
}

func TestFormsValidation(t *testing.T) {
	content, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	css := string(content)

	t.Run("ValidCSSSyntax", func(t *testing.T) {
		openBraces := strings.Count(css, "{")
		closeBraces := strings.Count(css, "}")
		if openBraces != closeBraces {
			t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
		}
	})

	t.Run("NoHardcodedColors", func(t *testing.T) {
		hexColorPattern := regexp.MustCompile(`(?i)#[0-9a-f]{3,6}`)
		rgbPattern := regexp.MustCompile(`(?i)rgb\(`)
		
		hexMatches := hexColorPattern.FindAllString(css, -1)
		rgbMatches := rgbPattern.FindAllString(css, -1)
		
		if len(hexMatches) > 0 {
			t.Errorf("Found hardcoded hex colors: %v. Use CSS variables instead", hexMatches)
		}
		if len(rgbMatches) > 0 {
			t.Errorf("Found hardcoded RGB colors: %v. Use CSS variables instead", rgbMatches)
		}
	})

	t.Run("UsesDesignTokens", func(t *testing.T) {
		requiredVars := []string{
			"var(--spacing-",
			"var(--color-",
			"var(--font-size-",
			"var(--radius-",
		}

		for _, varPrefix := range requiredVars {
			if !strings.Contains(css, varPrefix) {
				t.Errorf("Should use design tokens starting with %s", varPrefix)
			}
		}
	})
}

func TestFormsErrorStates(t *testing.T) {
	content, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	css := string(content)

	t.Run("ErrorInputStyling", func(t *testing.T) {
		if !strings.Contains(css, ".form-input.error") {
			t.Error("Missing error state styling for inputs")
		}
	})

	t.Run("ErrorMessageStyling", func(t *testing.T) {
		if !strings.Contains(css, ".form-error") {
			t.Error("Missing error message styling")
		}
	})

	t.Run("ErrorColorUsage", func(t *testing.T) {
		if !strings.Contains(css, "var(--color-error)") {
			t.Error("Error states should use --color-error variable")
		}
	})
}

func TestFormsSuccessStates(t *testing.T) {
	content, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	css := string(content)

	t.Run("SuccessInputStyling", func(t *testing.T) {
		if !strings.Contains(css, ".form-input.success") {
			t.Error("Missing success state styling for inputs")
		}
	})

	t.Run("SuccessMessageStyling", func(t *testing.T) {
		if !strings.Contains(css, ".form-success") {
			t.Error("Missing success message styling")
		}
	})

	t.Run("SuccessColorUsage", func(t *testing.T) {
		if !strings.Contains(css, "var(--color-success)") {
			t.Error("Success states should use --color-success variable")
		}
	})
}
