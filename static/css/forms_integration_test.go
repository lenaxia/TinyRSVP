package css

import (
	"os"
	"strings"
	"testing"
)

func TestFormsIntegrationWithVariables(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	forms := string(formsContent)
	variables := string(variablesContent)

	t.Run("AllSpacingVariablesExist", func(t *testing.T) {
		spacingVars := []string{
			"--spacing-1",
			"--spacing-2",
			"--spacing-3",
			"--spacing-4",
			"--spacing-10",
		}

		for _, varName := range spacingVars {
			if strings.Contains(forms, "var("+varName+")") {
				if !strings.Contains(variables, varName+":") {
					t.Errorf("forms.css uses %s but it's not defined in variables.css", varName)
				}
			}
		}
	})

	t.Run("AllColorVariablesExist", func(t *testing.T) {
		colorVars := []string{
			"--color-border",
			"--color-border-focus",
			"--color-error",
			"--color-success",
			"--color-text-primary",
			"--color-text-secondary",
			"--color-text-label",
			"--color-text-disabled",
			"--color-background",
			"--color-surface-disabled",
			"--color-primary-600",
			"--color-gray-400",
		}

		for _, varName := range colorVars {
			if strings.Contains(forms, "var("+varName+")") {
				if !strings.Contains(variables, varName+":") {
					t.Errorf("forms.css uses %s but it's not defined in variables.css", varName)
				}
			}
		}
	})

	t.Run("AllTypographyVariablesExist", func(t *testing.T) {
		typographyVars := []string{
			"--font-size-base",
			"--font-size-sm",
			"--font-weight-medium",
			"--font-family-sans",
			"--line-height-normal",
		}

		for _, varName := range typographyVars {
			if strings.Contains(forms, "var("+varName+")") {
				if !strings.Contains(variables, varName+":") {
					t.Errorf("forms.css uses %s but it's not defined in variables.css", varName)
				}
			}
		}
	})

	t.Run("AllBorderRadiusVariablesExist", func(t *testing.T) {
		radiusVars := []string{
			"--radius-sm",
			"--radius-md",
			"--radius-full",
		}

		for _, varName := range radiusVars {
			if strings.Contains(forms, "var("+varName+")") {
				if !strings.Contains(variables, varName+":") {
					t.Errorf("forms.css uses %s but it's not defined in variables.css", varName)
				}
			}
		}
	})

	t.Run("AllTransitionVariablesExist", func(t *testing.T) {
		transitionVars := []string{
			"--transition-fast",
		}

		for _, varName := range transitionVars {
			if strings.Contains(forms, "var("+varName+")") {
				if !strings.Contains(variables, varName+":") {
					t.Errorf("forms.css uses %s but it's not defined in variables.css", varName)
				}
			}
		}
	})
}

func TestFormsIntegrationWithTypography(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	_, err = os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("ConsistentFontSizeUsage", func(t *testing.T) {
		if !strings.Contains(forms, "var(--font-size-base)") {
			t.Error("forms.css should use --font-size-base for consistency with typography system")
		}
		if !strings.Contains(forms, "var(--font-size-sm)") {
			t.Error("forms.css should use --font-size-sm for help text and error messages")
		}
	})

	t.Run("ConsistentFontFamilyUsage", func(t *testing.T) {
		if !strings.Contains(forms, "var(--font-family-sans)") {
			t.Error("forms.css should use --font-family-sans for consistency")
		}
	})

	t.Run("ConsistentLineHeightUsage", func(t *testing.T) {
		if !strings.Contains(forms, "line-height:") {
			t.Error("forms.css should define line-height for readability")
		}
	})

	t.Run("NoConflictingTypographyClasses", func(t *testing.T) {
		typographyClasses := []string{
			".text-primary",
			".text-secondary",
			".text-bold",
			".text-large",
		}

		for _, class := range typographyClasses {
			if strings.Contains(forms, class+" {") {
				t.Errorf("forms.css should not redefine typography class %s", class)
			}
		}
	})
}

func TestFormsIntegrationWithColors(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	_, err = os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("NoConflictingColorClasses", func(t *testing.T) {
		colorClasses := []string{
			".bg-primary",
			".bg-error",
			".bg-success",
			".text-error",
			".text-success",
		}

		for _, class := range colorClasses {
			if strings.Contains(forms, class+" {") {
				t.Errorf("forms.css should not redefine color class %s", class)
			}
		}
	})

	t.Run("UsesSemanticColorVariables", func(t *testing.T) {
		semanticColors := []string{
			"var(--color-error)",
			"var(--color-success)",
		}

		for _, color := range semanticColors {
			if !strings.Contains(forms, color) {
				t.Errorf("forms.css should use semantic color %s", color)
			}
		}
	})
}

func TestFormsIntegrationWithSpacing(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	_, err = os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("NoConflictingSpacingClasses", func(t *testing.T) {
		spacingClasses := []string{
			".m-4",
			".p-3",
			".mb-2",
			".gap-2",
		}

		for _, class := range spacingClasses {
			if strings.Contains(forms, class+" {") {
				t.Errorf("forms.css should not redefine spacing class %s", class)
			}
		}
	})

	t.Run("ConsistentSpacingScale", func(t *testing.T) {
		spacingVars := []string{
			"var(--spacing-1)",
			"var(--spacing-2)",
			"var(--spacing-3)",
			"var(--spacing-4)",
		}

		for _, spacingVar := range spacingVars {
			if !strings.Contains(forms, spacingVar) {
				t.Errorf("forms.css should use spacing variable %s for consistency", spacingVar)
			}
		}
	})
}

func TestFormsIntegrationWithGrid(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	_, err = os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("NoConflictingFlexClasses", func(t *testing.T) {
		flexClasses := []string{
			".flex {",
			".items-center {",
			".gap-2 {",
		}

		for _, class := range flexClasses {
			if strings.Contains(forms, class) {
				t.Errorf("forms.css should not redefine flex class %s", class)
			}
		}
	})

	t.Run("UsesFlexForCheckWrappers", func(t *testing.T) {
		if !strings.Contains(forms, ".form-check-wrapper") {
			t.Error("forms.css should define .form-check-wrapper")
		}
		if !strings.Contains(forms, "display: flex") {
			t.Error("forms.css should use flexbox for check wrappers")
		}
	})
}

func TestFormsResponsiveIntegration(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	forms := string(formsContent)
	variables := string(variablesContent)

	t.Run("ResponsiveBreakpointsMatch", func(t *testing.T) {
		if strings.Contains(forms, "@media") {
			if strings.Contains(forms, "min-width: 768px") {
				if !strings.Contains(variables, "--breakpoint-md: 768px") {
					t.Error("forms.css breakpoint should match --breakpoint-md in variables.css")
				}
			}
		}
	})

	t.Run("MobileFirstMediaQueries", func(t *testing.T) {
		if strings.Contains(forms, "@media") {
			if !strings.Contains(forms, "min-width") {
				t.Error("forms.css should use min-width for mobile-first approach")
			}
			if strings.Contains(forms, "max-width") {
				t.Error("forms.css should not use max-width (use mobile-first approach)")
			}
		}
	})
}

func TestFormsAccessibilityIntegration(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("TouchTargetSize", func(t *testing.T) {
		if !strings.Contains(forms, "width: 20px") || !strings.Contains(forms, "height: 20px") {
			t.Error("Checkbox and radio inputs should be at least 20px for touch targets")
		}
	})

	t.Run("FocusVisibility", func(t *testing.T) {
		focusStates := []string{
			".form-input:focus",
			".form-select:focus",
			".form-textarea:focus",
			".form-checkbox:focus",
			".form-radio:focus",
		}

		for _, state := range focusStates {
			if !strings.Contains(forms, state) {
				t.Errorf("Missing focus state for %s", state)
			}
		}

		if !strings.Contains(forms, "outline:") {
			t.Error("Focus states should include visible outline")
		}
		if !strings.Contains(forms, "outline-offset:") {
			t.Error("Focus states should include outline-offset for better visibility")
		}
	})

	t.Run("DisabledStateIndicators", func(t *testing.T) {
		if !strings.Contains(forms, "cursor: not-allowed") {
			t.Error("Disabled states should use cursor: not-allowed")
		}
		if !strings.Contains(forms, "opacity:") {
			t.Error("Disabled states should use opacity for visual indication")
		}
	})

	t.Run("ColorContrastCompliance", func(t *testing.T) {
		if !strings.Contains(forms, "var(--color-text-primary)") {
			t.Error("Form inputs should use --color-text-primary for WCAG compliance")
		}
		if !strings.Contains(forms, "var(--color-text-disabled)") {
			t.Error("Disabled inputs should use --color-text-disabled for proper contrast")
		}
	})
}

func TestFormsHTMLIntegration(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("SupportsStandardHTMLElements", func(t *testing.T) {
		elements := []string{
			".form-input",
			".form-select",
			".form-textarea",
			".form-checkbox",
			".form-radio",
		}

		for _, element := range elements {
			if !strings.Contains(forms, element) {
				t.Errorf("Missing styles for %s", element)
			}
		}
	})

	t.Run("SupportsHTMLStates", func(t *testing.T) {
		states := []string{
			":focus",
			":disabled",
			":checked",
			":hover",
		}

		for _, state := range states {
			if !strings.Contains(forms, state) {
				t.Errorf("Missing support for %s state", state)
			}
		}
	})

	t.Run("SupportsPlaceholders", func(t *testing.T) {
		if !strings.Contains(forms, "::placeholder") {
			t.Error("Missing placeholder styling")
		}
	})
}

func TestFormsLoadOrder(t *testing.T) {
	t.Run("DependencyOrder", func(t *testing.T) {
		dependencies := []string{
			"variables.css",
			"typography.css",
			"colors.css",
			"spacing.css",
			"grid.css",
			"forms.css",
		}

		for i, dep := range dependencies {
			if _, err := os.Stat(dep); os.IsNotExist(err) {
				t.Errorf("Dependency %s does not exist (load order position %d)", dep, i+1)
			}
		}
	})
}

func TestFormsPerformance(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	forms := string(formsContent)

	t.Run("FileSize", func(t *testing.T) {
		size := len(formsContent)
		maxSize := 15 * 1024

		if size > maxSize {
			t.Errorf("forms.css is %d bytes, should be under %d bytes for performance", size, maxSize)
		}
	})

	t.Run("UsesTransitions", func(t *testing.T) {
		if !strings.Contains(forms, "transition:") {
			t.Error("Forms should use transitions for smooth interactions")
		}
		if !strings.Contains(forms, "var(--transition-fast)") {
			t.Error("Forms should use --transition-fast variable for consistency")
		}
	})

	t.Run("NoExpensiveSelectors", func(t *testing.T) {
		expensivePatterns := []string{
			"* {",
			"[class*=",
			"[id*=",
		}

		for _, pattern := range expensivePatterns {
			if strings.Contains(forms, pattern) {
				t.Errorf("forms.css contains expensive selector pattern: %s", pattern)
			}
		}
	})
}
