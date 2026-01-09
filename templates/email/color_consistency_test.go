package email

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestEmailTemplateColorConsistency(t *testing.T) {
	cssVarsPath := filepath.Join("..", "..", "static", "css", "variables.css")
	cssContent, err := os.ReadFile(cssVarsPath)
	if err != nil {
		t.Fatalf("Failed to read CSS variables file: %v", err)
	}

	cssVars := parseCSSVariables(string(cssContent))

	emailTemplates := []string{
		"rsvp_confirmation.html",
	}

	for _, templateFile := range emailTemplates {
		t.Run(templateFile, func(t *testing.T) {
			content, err := os.ReadFile(templateFile)
			if err != nil {
				t.Fatalf("Failed to read template %s: %v", templateFile, err)
			}

			validateTemplateColors(t, string(content), cssVars, templateFile)
		})
	}

	internalTemplates := []struct {
		path string
		name string
	}{
		{"../../internal/templates/defaults/invite_email.html", "invite_email.html"},
		{"../../internal/templates/defaults/confirmation_page.html", "confirmation_page.html"},
	}

	for _, tmpl := range internalTemplates {
		t.Run(tmpl.name, func(t *testing.T) {
			content, err := os.ReadFile(tmpl.path)
			if err != nil {
				t.Fatalf("Failed to read template %s: %v", tmpl.name, err)
			}

			validateTemplateColors(t, string(content), cssVars, tmpl.name)
		})
	}
}

func parseCSSVariables(content string) map[string]string {
	vars := make(map[string]string)
	re := regexp.MustCompile(`--([a-z0-9-]+):\s*(#[0-9a-fA-F]{3,6}|rgba?\([^)]+\));`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			varName := match[1]
			varValue := strings.ToLower(strings.TrimSpace(match[2]))
			vars[varName] = varValue
		}
	}

	return vars
}

func validateTemplateColors(t *testing.T, content string, cssVars map[string]string, templateName string) {
	colorMap := map[string][]string{
		"#f5f5f5": {"color-surface-disabled"},
		"#111827": {"color-text-primary", "color-gray-900"},
		"#666":    {"color-text-muted"},
		"#555":    {"color-text-label"},
		"#16a34a": {"color-success"},
		"#166534": {"color-success-dark"},
		"#2563eb": {"color-primary-600"},
		"#1d4ed8": {"color-primary-700"},
		"#92400e": {"color-warning-dark"},
		"#f59e0b": {"color-warning"},
		"#fef3c7": {"color-warning-light"},
		"#e5e7eb": {"color-border", "color-gray-200"},
		"#f9fafb": {"color-surface", "color-gray-50"},
		"#4b5563": {"color-gray-600"},
		"#374151": {"color-gray-700"},
		"#dcfce7": {"color-success-light"},
		"#fee2e2": {"color-error-light"},
		"#991b1b": {"color-error-dark"},
	}

	hexColorRe := regexp.MustCompile(`(?i)(#[0-9a-fA-F]{3,6})`)
	matches := hexColorRe.FindAllString(content, -1)

	foundColors := make(map[string]bool)
	for _, match := range matches {
		color := strings.ToLower(match)
		foundColors[color] = true
	}

	for color := range foundColors {
		expectedVars, exists := colorMap[color]
		if !exists {
			t.Errorf("Template %s uses color %s which is not mapped to CSS variables", templateName, color)
			continue
		}

		found := false
		for _, varName := range expectedVars {
			if cssValue, ok := cssVars[varName]; ok {
				if normalizeColor(cssValue) == normalizeColor(color) {
					found = true
					break
				}
			}
		}

		if !found {
			t.Errorf("Template %s uses color %s but it doesn't match any of the expected CSS variables: %v", 
				templateName, color, expectedVars)
		}
	}

	commentRe := regexp.MustCompile(`(?s)<!--.*?Color Mappings:.*?-->`)
	if !commentRe.MatchString(content) {
		t.Errorf("Template %s is missing color mapping documentation comment", templateName)
	}
}

func normalizeColor(color string) string {
	color = strings.ToLower(strings.TrimSpace(color))
	
	if len(color) == 4 && color[0] == '#' {
		r := string(color[1])
		g := string(color[2])
		b := string(color[3])
		color = "#" + r + r + g + g + b + b
	}
	
	return color
}

func TestColorNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#fff", "#ffffff"},
		{"#FFF", "#ffffff"},
		{"#abc", "#aabbcc"},
		{"#123456", "#123456"},
		{"#ABCDEF", "#abcdef"},
		{"  #fff  ", "#ffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeColor(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeColor(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseCSSVariables(t *testing.T) {
	cssContent := `
:root {
    --color-primary-600: #2563eb;
    --color-text-primary: #111827;
    --color-surface: #f9fafb;
}
`
	vars := parseCSSVariables(cssContent)

	expectedVars := map[string]string{
		"color-primary-600":  "#2563eb",
		"color-text-primary": "#111827",
		"color-surface":      "#f9fafb",
	}

	for key, expectedValue := range expectedVars {
		if value, ok := vars[key]; !ok {
			t.Errorf("Expected CSS variable %s not found", key)
		} else if value != expectedValue {
			t.Errorf("CSS variable %s = %s, want %s", key, value, expectedValue)
		}
	}
}
