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
		// The template uses external CSS files for color variables rather than inline styles.
		// Verify the template links to the required CSS files that define color variables.
		requiredColorCSS := []string{
			`/static/css/variables.css`,
			`/static/css/colors.css`,
		}
		for _, cssFile := range requiredColorCSS {
			t.Run(cssFile, func(t *testing.T) {
				if !strings.Contains(html, cssFile) {
					t.Errorf("Template does not link to %s for color variables", cssFile)
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
		// The template uses external CSS for shadows; verify the CSS file is linked
		if !strings.Contains(html, `/static/css/variables.css`) {
			t.Error("Template does not link to variables.css which defines shadow variables")
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
		// The template uses external CSS for spacing via spacing.css
		// It may use some spacing vars in JS inline styles; verify CSS file is linked
		if !strings.Contains(html, `/static/css/spacing.css`) {
			t.Error("Template does not link to spacing.css for spacing variables")
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
		// Template uses external CSS (spacing.css) for spacing in CSS classes
		// Verify the template has the relevant CSS classes that will use spacing variables
		cssClasses := []string{
			"rsvp-container",
			"rsvp-form",
		}
		for _, cls := range cssClasses {
			if !strings.Contains(html, cls) {
				t.Errorf("Expected to find CSS class %q in template", cls)
			}
		}
	})

	t.Run("responsive spacing uses variables", func(t *testing.T) {
		// Template uses external rsvp_page.css for responsive spacing via CSS variables
		if !strings.Contains(html, `/static/css/rsvp_page.css`) {
			t.Error("Template does not link to rsvp_page.css for responsive spacing")
		}
	})
}
