package css

import (
	"os"
	"strings"
	"testing"
)

func TestEpic11TwoLayerThemeCompatibility(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	cssContent := string(content)

	t.Run("uses data-theme attribute for system theme", func(t *testing.T) {
		if !strings.Contains(cssContent, "[data-theme=\"dark\"]") {
			t.Error("System theme must use [data-theme] attribute for Epic 11 compatibility")
		}

		if strings.Contains(cssContent, "@media (prefers-color-scheme: dark)") {
			t.Error("Should not use media query - data attribute allows two-layer system")
		}
	})

	t.Run("system theme variables don't conflict with event theme namespace", func(t *testing.T) {
		if strings.Contains(cssContent, "--theme-primary") {
			t.Error("System theme should not use --theme-* namespace (reserved for event themes)")
		}

		if strings.Contains(cssContent, "--event-") {
			t.Error("System theme should not use --event-* namespace (reserved for event themes)")
		}
	})

	t.Run("all color variables use semantic naming", func(t *testing.T) {
		requiredVars := []string{
			"--color-background",
			"--color-surface",
			"--color-text-primary",
			"--color-border",
		}

		for _, varName := range requiredVars {
			if !strings.Contains(cssContent, varName) {
				t.Errorf("Missing semantic variable for two-layer system: %s", varName)
			}
		}
	})

	t.Run("supports programmatic theme switching", func(t *testing.T) {
		jsContent, err := os.ReadFile("../js/theme_controller.js")
		if err != nil {
			t.Fatalf("Failed to read theme_controller.js: %v", err)
		}

		js := string(jsContent)

		if !strings.Contains(js, "setAttribute('data-theme'") {
			t.Error("Theme controller must set data-theme attribute for Epic 11 compatibility")
		}

		if !strings.Contains(js, "localStorage") {
			t.Error("Theme persistence required for consistent user experience across event themes")
		}
	})

	t.Run("complete palette allows event theme layering", func(t *testing.T) {
		darkSection := extractDarkThemeSection(cssContent)

		criticalVars := []string{
			"--color-primary-500",
			"--color-primary-600",
			"--color-success",
			"--color-error",
			"--color-warning",
		}

		for _, varName := range criticalVars {
			if !strings.Contains(darkSection, varName) {
				t.Errorf("Dark mode missing %s - event themes need complete palette", varName)
			}
		}
	})
}

func TestEpic11EventThemeLayeringExample(t *testing.T) {
	t.Run("demonstrates two-layer theme system", func(t *testing.T) {
		exampleCSS := `
		/* Layer 1: System Theme (Story 10.12) */
		:root {
			--color-background: #ffffff;
			--color-surface: #f9fafb;
		}
		
		[data-theme="dark"] {
			--color-background: #0f172a;
			--color-surface: #1e293b;
		}
		
		/* Layer 2: Event Theme (Epic 11) */
		[data-event-theme="wedding"] {
			--theme-primary: #f4c2c2;
		}
		
		/* Component using both layers */
		.component {
			background: var(--color-surface);     /* System theme */
			border-color: var(--theme-primary);   /* Event theme */
		}
		`

		if !strings.Contains(exampleCSS, "data-theme") {
			t.Error("Example should show data-theme attribute usage")
		}

		if !strings.Contains(exampleCSS, "data-event-theme") {
			t.Error("Example should show data-event-theme attribute usage")
		}

		if !strings.Contains(exampleCSS, "var(--color-") && !strings.Contains(exampleCSS, "var(--theme-") {
			t.Error("Example should show both system and event theme variables")
		}
	})
}
