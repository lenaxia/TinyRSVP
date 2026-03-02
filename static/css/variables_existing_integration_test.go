package css

import (
	"os"
	"strings"
	"testing"
)

func TestCSSVariablesIntegrationWithExistingRSVPTemplate(t *testing.T) {
	cssContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	templatePath := "../../templates/web/rsvp_page.html"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("RSVP template not found at %s, skipping integration test", templatePath)
	}

	template := string(templateContent)
	css := string(cssContent)

	oldToNewMapping := map[string]string{
		"--primary-color":  "--color-primary-600",
		"--primary-hover":  "--color-primary-700",
		"--success-color":  "--color-success",
		"--warning-color":  "--color-warning",
		"--error-color":    "--color-error",
		"--text-primary":   "--color-text-primary",
		"--text-secondary": "--color-text-secondary",
		"--bg-primary":     "--color-background",
		"--bg-secondary":   "--color-surface",
		"--border-color":   "--color-border",
	}

	for oldVar, newVar := range oldToNewMapping {
		if strings.Contains(template, oldVar) {
			if !strings.Contains(css, newVar) {
				t.Errorf("Template uses %s which should map to %s, but %s is not defined in variables.css", oldVar, newVar, newVar)
			}
		}
	}
}

func TestCSSVariablesCanReplaceInlineStyles(t *testing.T) {
	cssContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	css := string(cssContent)

	inlineStylesToReplace := map[string]string{
		"#2563eb":                        "--color-primary-600",
		"#1d4ed8":                        "--color-primary-700",
		"#16a34a":                        "--color-success",
		"#ea580c":                        "--color-warning",
		"#dc2626":                        "--color-error",
		"#1f2937":                        "--color-gray-800",
		"#6b7280":                        "--color-text-secondary",
		"#ffffff":                        "--color-background",
		"#f9fafb":                        "--color-surface",
		"#e5e7eb":                        "--color-border",
		"rgba(37, 99, 235, 0.1)":         "derived from --color-primary-600",
		"0 1px 3px 0 rgba(0, 0, 0, 0.1)": "--shadow-base",
	}

	for inlineValue, expectedVar := range inlineStylesToReplace {
		if strings.Contains(inlineValue, "rgba") || strings.Contains(inlineValue, "0 1px") {
			if !strings.Contains(css, expectedVar) && !strings.Contains(expectedVar, "derived") {
				t.Errorf("Inline style '%s' should be replaced with %s, but it's not defined", inlineValue, expectedVar)
			}
		} else {
			if !strings.Contains(strings.ToLower(css), strings.ToLower(inlineValue)) {
				t.Logf("Warning: Inline color %s should be replaced with %s", inlineValue, expectedVar)
			}
		}
	}
}

func TestCSSVariablesProvideAllRequiredTokens(t *testing.T) {
	cssContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	css := string(cssContent)

	requiredForRSVPPage := []string{
		"--color-primary-600",
		"--color-primary-700",
		"--color-success",
		"--color-warning",
		"--color-error",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-border-focus",
		"--spacing-2",
		"--spacing-4",
		"--spacing-6",
		"--font-size-base",
		"--font-size-sm",
		"--font-weight-normal",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--radius-base",
		"--radius-lg",
		"--shadow-base",
		"--transition-base",
	}

	for _, variable := range requiredForRSVPPage {
		if !strings.Contains(css, variable) {
			t.Errorf("CSS variables missing %s which is required for RSVP page", variable)
		}
	}
}

func TestCSSVariablesBreakpointCompatibility(t *testing.T) {
	cssContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	templatePath := "../../templates/web/rsvp_page.html"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("RSVP template not found at %s, skipping integration test", templatePath)
	}

	template := string(templateContent)
	css := string(cssContent)

	if strings.Contains(template, "@media (min-width: 768px)") {
		if !strings.Contains(css, "--breakpoint-md: 768px") {
			t.Error("Template uses 768px breakpoint but variables.css doesn't define --breakpoint-md")
		}
	}

	if strings.Contains(template, "@media (min-width: 1024px)") {
		if !strings.Contains(css, "--breakpoint-lg: 1024px") {
			t.Error("Template uses 1024px breakpoint but variables.css doesn't define --breakpoint-lg")
		}
	}
}

func TestCSSVariablesEmailTemplateCompatibility(t *testing.T) {
	cssContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	css := string(cssContent)

	emailRequiredVars := []string{
		"--color-primary-600",
		"--color-success",
		"--color-text-primary",
		"--color-background",
		"--spacing-4",
		"--font-size-base",
	}

	for _, variable := range emailRequiredVars {
		if !strings.Contains(css, variable) {
			t.Errorf("CSS variables missing %s which may be needed for email templates", variable)
		}
	}
}
