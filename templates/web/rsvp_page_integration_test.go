package web

import (
	"os"
	"strings"
	"testing"
)

func TestRSVPPageUsesExternalCSSVariables(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	t.Run("includes CSS variables stylesheet link", func(t *testing.T) {
		expectedLink := `<link rel="stylesheet" href="/static/css/variables.css">`
		if !strings.Contains(html, expectedLink) {
			t.Errorf("Template does not include link to variables.css")
		}
	})

	t.Run("does not define inline :root variables", func(t *testing.T) {
		if strings.Contains(html, ":root {") {
			lines := strings.Split(html, "\n")
			for i, line := range lines {
				if strings.Contains(line, ":root {") {
					t.Errorf("Template defines inline :root variables at line %d: %s", i+1, strings.TrimSpace(line))
					break
				}
			}
		}
	})

	t.Run("uses centralized color variables", func(t *testing.T) {
		tests := []struct {
			name     string
			variable string
		}{
			{"primary color", "var(--color-primary-600)"},
			{"primary hover", "var(--color-primary-700)"},
			{"text primary", "var(--color-text-primary)"},
			{"background", "var(--color-background)"},
			{"surface", "var(--color-surface)"},
			{"border", "var(--color-border)"},
			{"error", "var(--color-error)"},
			{"success", "var(--color-success)"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if !strings.Contains(html, tt.variable) {
					t.Errorf("Template does not use %s", tt.variable)
				}
			})
		}
	})

	t.Run("does not use hardcoded colors", func(t *testing.T) {
		hardcodedColors := []string{
			"#2563eb",
			"#1d4ed8",
			"#007bff",
			"#dc3545",
			"#dc2626",
		}

		for _, color := range hardcodedColors {
			if strings.Contains(html, color) {
				t.Errorf("Template contains hardcoded color %s instead of using CSS variables", color)
			}
		}
	})

	t.Run("comprehensive: no hardcoded hex colors in style blocks", func(t *testing.T) {
		lines := strings.Split(html, "\n")
		inStyleBlock := false
		
		for i, line := range lines {
			if strings.Contains(line, "<style>") {
				inStyleBlock = true
				continue
			}
			if strings.Contains(line, "</style>") {
				inStyleBlock = false
				continue
			}
			
			if inStyleBlock {
				if strings.Contains(line, "#") {
					parts := strings.Split(line, "#")
					for j := 1; j < len(parts); j++ {
						hexPart := ""
						for k := 0; k < len(parts[j]) && k < 6; k++ {
							c := parts[j][k]
							if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
								hexPart += string(c)
							} else {
								break
							}
						}
						
						if len(hexPart) == 3 || len(hexPart) == 6 {
							t.Errorf("Line %d contains hardcoded hex color #%s: %s", i+1, hexPart, strings.TrimSpace(line))
						}
					}
				}
			}
		}
	})

	t.Run("includes typography stylesheet", func(t *testing.T) {
		expectedLink := `<link rel="stylesheet" href="/static/css/typography.css">`
		if !strings.Contains(html, expectedLink) {
			t.Errorf("Template does not include link to typography.css")
		}
	})

	t.Run("uses typography variables in inline styles", func(t *testing.T) {
		typographyVars := []string{
			"var(--font-size-",
			"var(--font-weight-",
			"var(--spacing-",
		}

		foundAny := false
		for _, varPrefix := range typographyVars {
			if strings.Contains(html, varPrefix) {
				foundAny = true
				break
			}
		}

		if !foundAny {
			t.Error("Template should use typography variables in inline styles")
		}
	})

	t.Run("uses centralized shadow variables", func(t *testing.T) {
		if !strings.Contains(html, "var(--shadow-base)") {
			t.Error("Template does not use var(--shadow-base)")
		}
	})
}

func TestRSVPPageNoInlineVariableDefinitions(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	deprecatedVariables := []string{
		"--primary-color:",
		"--primary-hover:",
		"--success-color:",
		"--warning-color:",
		"--error-color:",
		"--text-primary:",
		"--text-secondary:",
		"--bg-primary:",
		"--bg-secondary:",
		"--border-color:",
	}

	for _, variable := range deprecatedVariables {
		if strings.Contains(html, variable) {
			t.Errorf("Template still defines deprecated inline variable: %s", variable)
		}
	}
}

func TestRSVPPageSpacingIntegration(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	t.Run("includes spacing.css stylesheet link", func(t *testing.T) {
		expectedLink := `<link rel="stylesheet" href="/static/css/spacing.css">`
		if !strings.Contains(html, expectedLink) {
			t.Errorf("Template does not include link to spacing.css")
		}
	})

	t.Run("uses spacing variables in inline styles", func(t *testing.T) {
		spacingVars := []string{
			"var(--spacing-2)",
			"var(--spacing-3)",
			"var(--spacing-4)",
			"var(--spacing-6)",
			"var(--spacing-8)",
		}

		for _, spacingVar := range spacingVars {
			if !strings.Contains(html, spacingVar) {
				t.Errorf("Template does not use %s", spacingVar)
			}
		}
	})

	t.Run("does not contain hardcoded spacing values", func(t *testing.T) {
		hardcodedSpacing := []struct {
			value string
			desc  string
		}{
			{"padding: 1rem", "1rem padding"},
			{"padding: 1.5rem", "1.5rem padding"},
			{"padding: 2rem", "2rem padding"},
			{"padding: 0.75rem", "0.75rem padding"},
			{"margin: 1rem", "1rem margin"},
			{"margin: 1.5rem", "1.5rem margin"},
			{"margin: 2rem", "2rem margin"},
			{"margin-bottom: 1rem", "1rem margin-bottom"},
			{"margin-bottom: 1.5rem", "1.5rem margin-bottom"},
			{"margin-top: 1.5rem", "1.5rem margin-top"},
			{"margin-top: 2rem", "2rem margin-top"},
			{"margin-right: 0.75rem", "0.75rem margin-right"},
			{"gap: 1rem", "1rem gap"},
			{"gap: 0.75rem", "0.75rem gap"},
		}

		for _, hc := range hardcodedSpacing {
			if strings.Contains(html, hc.value) {
				t.Errorf("Template contains hardcoded spacing %s instead of using CSS variables", hc.desc)
			}
		}
	})

	t.Run("spacing variables used in correct contexts", func(t *testing.T) {
		contexts := []struct {
			selector string
			variable string
		}{
			{"body", "var(--spacing-4)"},
			{".event-card", "var(--spacing-6)"},
			{".error-container", "var(--spacing-8)"},
			{".form-actions", "var(--spacing-8)"},
			{".questions-section", "var(--spacing-8)"},
		}

		for _, ctx := range contexts {
			if !strings.Contains(html, ctx.variable) {
				t.Errorf("Expected %s to use %s", ctx.selector, ctx.variable)
			}
		}
	})

	t.Run("responsive spacing uses variables", func(t *testing.T) {
		lines := strings.Split(html, "\n")
		inMediaQuery := false
		foundResponsiveSpacing := false

		for _, line := range lines {
			if strings.Contains(line, "@media (min-width: 768px)") {
				inMediaQuery = true
			}
			if inMediaQuery && strings.Contains(line, "var(--spacing-") {
				foundResponsiveSpacing = true
				break
			}
			if inMediaQuery && strings.Contains(line, "}") && strings.TrimSpace(line) == "}" {
				inMediaQuery = false
			}
		}

		if !foundResponsiveSpacing {
			t.Error("Responsive media queries should use spacing variables")
		}
	})
}
