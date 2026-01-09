package defaults

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultRSVPPageUsesExternalCSSVariables(t *testing.T) {
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
			"#007bff",
			"#dc3545",
			"#0056b3",
			"#004494",
		}

		for _, color := range hardcodedColors {
			if strings.Contains(html, color) {
				t.Errorf("Template contains hardcoded color %s instead of using CSS variables", color)
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

func TestDefaultRSVPPageNoHardcodedStyles(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	t.Run("does not use hardcoded hex colors in styles", func(t *testing.T) {
		lines := strings.Split(html, "\n")
		styleSection := false
		for i, line := range lines {
			if strings.Contains(line, "<style>") {
				styleSection = true
				continue
			}
			if strings.Contains(line, "</style>") {
				styleSection = false
				continue
			}

			if styleSection {
				if strings.Contains(line, "#") && (strings.Contains(line, "color:") || strings.Contains(line, "background:") || strings.Contains(line, "border:")) {
					hexPattern := strings.Index(line, "#")
					if hexPattern != -1 {
						snippet := line[hexPattern : min(hexPattern+7, len(line))]
						if len(snippet) == 7 && isHexColor(snippet) {
							t.Errorf("Line %d contains hardcoded hex color: %s", i+1, strings.TrimSpace(line))
						}
					}
				}
			}
		}
	})
}

func isHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
