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

	t.Run("uses centralized typography variables", func(t *testing.T) {
		tests := []struct {
			name     string
			variable string
		}{
			{"font family", "var(--font-family-sans)"},
			{"line height", "var(--line-height-normal)"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if !strings.Contains(html, tt.variable) {
					t.Errorf("Template does not use %s", tt.variable)
				}
			})
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
