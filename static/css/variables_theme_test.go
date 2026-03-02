package css

import (
	"os"
	"strings"
	"testing"
)

func TestVariablesThemeStructure(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	cssContent := string(content)

	t.Run("uses data-theme attribute instead of media query", func(t *testing.T) {
		if strings.Contains(cssContent, "@media (prefers-color-scheme: dark)") {
			t.Error("CSS should not use @media (prefers-color-scheme: dark), should use [data-theme=\"dark\"] instead")
		}

		if !strings.Contains(cssContent, "[data-theme=\"dark\"]") {
			t.Error("CSS must contain [data-theme=\"dark\"] selector")
		}
	})

	t.Run("has complete dark mode color palette", func(t *testing.T) {
		requiredDarkVars := []string{
			"--color-background",
			"--color-surface",
			"--color-surface-disabled",
			"--color-text-primary",
			"--color-text-secondary",
			"--color-text-tertiary",
			"--color-text-muted",
			"--color-text-label",
			"--color-text-disabled",
			"--color-border",
			"--color-border-focus",
		}

		darkThemeSection := extractDarkThemeSection(cssContent)
		if darkThemeSection == "" {
			t.Fatal("Could not find [data-theme=\"dark\"] section")
		}

		for _, varName := range requiredDarkVars {
			if !strings.Contains(darkThemeSection, varName) {
				t.Errorf("Dark theme missing required variable: %s", varName)
			}
		}
	})

	t.Run("has dark mode primary color scale", func(t *testing.T) {
		darkThemeSection := extractDarkThemeSection(cssContent)

		primaryShades := []string{
			"--color-primary-50",
			"--color-primary-100",
			"--color-primary-200",
			"--color-primary-300",
			"--color-primary-400",
			"--color-primary-500",
			"--color-primary-600",
			"--color-primary-700",
			"--color-primary-800",
			"--color-primary-900",
		}

		for _, shade := range primaryShades {
			if !strings.Contains(darkThemeSection, shade) {
				t.Errorf("Dark theme missing primary color shade: %s", shade)
			}
		}
	})

	t.Run("has dark mode state colors", func(t *testing.T) {
		darkThemeSection := extractDarkThemeSection(cssContent)

		stateColors := []string{
			"--color-success",
			"--color-success-dark",
			"--color-success-light",
			"--color-warning",
			"--color-warning-dark",
			"--color-warning-light",
			"--color-error",
			"--color-error-dark",
			"--color-error-light",
			"--color-info",
		}

		for _, color := range stateColors {
			if !strings.Contains(darkThemeSection, color) {
				t.Errorf("Dark theme missing state color: %s", color)
			}
		}
	})

	t.Run("has dark mode gray scale", func(t *testing.T) {
		darkThemeSection := extractDarkThemeSection(cssContent)

		grayShades := []string{
			"--color-gray-50",
			"--color-gray-100",
			"--color-gray-200",
			"--color-gray-300",
			"--color-gray-400",
			"--color-gray-500",
			"--color-gray-600",
			"--color-gray-700",
			"--color-gray-800",
			"--color-gray-900",
		}

		for _, shade := range grayShades {
			if !strings.Contains(darkThemeSection, shade) {
				t.Errorf("Dark theme missing gray shade: %s", shade)
			}
		}
	})

	t.Run("has dark mode shadows", func(t *testing.T) {
		darkThemeSection := extractDarkThemeSection(cssContent)

		shadows := []string{
			"--shadow-sm",
			"--shadow-base",
			"--shadow-md",
			"--shadow-lg",
			"--shadow-xl",
			"--shadow-2xl",
		}

		for _, shadow := range shadows {
			if !strings.Contains(darkThemeSection, shadow) {
				t.Errorf("Dark theme missing shadow: %s", shadow)
			}
		}
	})

	t.Run("light theme remains default in :root", func(t *testing.T) {
		if !strings.Contains(cssContent, ":root {") {
			t.Error("CSS must have :root selector for light theme defaults")
		}

		rootSection := extractRootSection(cssContent)
		if !strings.Contains(rootSection, "--color-background: #ffffff") {
			t.Error("Light theme background should be white (#ffffff) in :root")
		}
	})
}

func TestVariablesThemeTransitions(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	cssContent := string(content)

	t.Run("has transition variables defined", func(t *testing.T) {
		transitions := []string{
			"--transition-fast",
			"--transition-base",
			"--transition-slow",
		}

		for _, transition := range transitions {
			if !strings.Contains(cssContent, transition) {
				t.Errorf("Missing transition variable: %s", transition)
			}
		}
	})
}

func extractDarkThemeSection(css string) string {
	startIdx := strings.Index(css, "[data-theme=\"dark\"]")
	if startIdx == -1 {
		return ""
	}

	braceCount := 0
	inSection := false
	var result strings.Builder

	for i := startIdx; i < len(css); i++ {
		char := css[i]
		result.WriteByte(char)

		if char == '{' {
			braceCount++
			inSection = true
		} else if char == '}' {
			braceCount--
			if inSection && braceCount == 0 {
				break
			}
		}
	}

	return result.String()
}

func extractRootSection(css string) string {
	startIdx := strings.Index(css, ":root {")
	if startIdx == -1 {
		return ""
	}

	braceCount := 0
	inSection := false
	var result strings.Builder

	for i := startIdx; i < len(css); i++ {
		char := css[i]
		result.WriteByte(char)

		if char == '{' {
			braceCount++
			inSection = true
		} else if char == '}' {
			braceCount--
			if inSection && braceCount == 0 {
				break
			}
		}
	}

	return result.String()
}
