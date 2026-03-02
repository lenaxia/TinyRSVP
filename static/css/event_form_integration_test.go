package css

import (
	"os"
	"strings"
	"testing"
)

func TestEventFormCSSIntegration(t *testing.T) {
	content, err := os.ReadFile("event_form.css")
	if err != nil {
		t.Fatalf("Failed to read event_form.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "contains all required classes",
			expected: []string{
				".event-form",
				".form-section",
				".form-section-title",
				".form-actions",
				".question-item",
				".alert",
				".alert-error",
			},
		},
		{
			name: "uses design system variables",
			expected: []string{
				"var(--color-surface)",
				"var(--color-border)",
				"var(--spacing-",
				"var(--radius-",
				"var(--font-size-",
				"var(--font-weight-",
			},
		},
		{
			name: "contains responsive breakpoints",
			expected: []string{
				"@media (max-width: 767px)",
				"@media (min-width: 768px)",
				"@media (min-width: 1024px)",
			},
		},
		{
			name: "contains mobile-first styles",
			expected: []string{
				"flex-direction: column",
				"width: 100%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(cssContent, exp) {
					t.Errorf("Expected CSS to contain %q", exp)
				}
			}
		})
	}
}

func TestEventFormCSSValidation(t *testing.T) {
	content, err := os.ReadFile("event_form.css")
	if err != nil {
		t.Fatalf("Failed to read event_form.css: %v", err)
	}

	cssContent := string(content)

	t.Run("no hardcoded colors", func(t *testing.T) {
		hardcodedColors := []string{"#fff", "#000", "rgb(", "rgba("}
		for _, color := range hardcodedColors {
			if strings.Contains(cssContent, color) {
				t.Errorf("CSS should not contain hardcoded color %q, use CSS variables instead", color)
			}
		}
	})

	t.Run("no hardcoded spacing", func(t *testing.T) {
		lines := strings.Split(cssContent, "\n")
		for i, line := range lines {
			if strings.Contains(line, "px") && !strings.Contains(line, "@media") && !strings.Contains(line, "max-width") && !strings.Contains(line, "min-width") && !strings.Contains(line, "border:") {
				if !strings.Contains(line, "var(--") {
					// Allow small pixel values (1-3px) used for borders, outlines, or badge padding
					trimmed := strings.TrimSpace(line)
					if strings.Contains(trimmed, "1px") || strings.Contains(trimmed, "2px") || strings.Contains(trimmed, "3px") {
						continue
					}
					t.Errorf("Line %d contains hardcoded px value without using CSS variable: %s", i+1, strings.TrimSpace(line))
				}
			}
		}
	})
}
