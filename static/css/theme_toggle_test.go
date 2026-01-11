package css

import (
	"os"
	"strings"
	"testing"
)

func TestThemeToggleCSS(t *testing.T) {
	content, err := os.ReadFile("theme_toggle.css")
	if err != nil {
		t.Fatalf("Failed to read theme_toggle.css: %v", err)
	}
	
	cssContent := string(content)
	
	t.Run("has theme-toggle class", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-toggle") {
			t.Error("CSS must contain .theme-toggle class")
		}
	})
	
	t.Run("has proper button sizing", func(t *testing.T) {
		requiredProps := []string{
			"width:",
			"height:",
			"padding:",
		}
		
		for _, prop := range requiredProps {
			if !strings.Contains(cssContent, prop) {
				t.Errorf("theme-toggle should have %s property", prop)
			}
		}
	})
	
	t.Run("has accessibility focus styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-toggle:focus") {
			t.Error("theme-toggle must have :focus styles for accessibility")
		}
		
		if !strings.Contains(cssContent, "outline:") {
			t.Error("theme-toggle:focus should have outline for keyboard navigation")
		}
	})
	
	t.Run("has hover styles", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-toggle:hover") {
			t.Error("theme-toggle should have :hover styles for visual feedback")
		}
	})
	
	t.Run("has theme-icon class", func(t *testing.T) {
		if !strings.Contains(cssContent, ".theme-icon") {
			t.Error("CSS must contain .theme-icon class for icon display")
		}
	})
	
	t.Run("uses CSS variables for theming", func(t *testing.T) {
		cssVars := []string{
			"var(--color-border)",
			"var(--color-surface)",
			"var(--color-primary-600)",
			"var(--color-border-focus)",
			"var(--radius-base)",
			"var(--transition-fast)",
		}
		
		for _, cssVar := range cssVars {
			if !strings.Contains(cssContent, cssVar) {
				t.Errorf("theme-toggle should use CSS variable: %s", cssVar)
			}
		}
	})
	
	t.Run("has responsive styles", func(t *testing.T) {
		if !strings.Contains(cssContent, "@media") {
			t.Error("theme-toggle should have responsive styles for different screen sizes")
		}
		
		if !strings.Contains(cssContent, "min-width:") {
			t.Error("responsive styles should use min-width media queries")
		}
	})
	
	t.Run("has proper cursor pointer", func(t *testing.T) {
		if !strings.Contains(cssContent, "cursor: pointer") {
			t.Error("theme-toggle should have cursor: pointer for clickability indication")
		}
	})
}

func TestThemeToggleAccessibility(t *testing.T) {
	content, err := os.ReadFile("theme_toggle.css")
	if err != nil {
		t.Fatalf("Failed to read theme_toggle.css: %v", err)
	}
	
	cssContent := string(content)
	
	t.Run("has minimum touch target size", func(t *testing.T) {
		if !strings.Contains(cssContent, "44px") {
			t.Error("theme-toggle should have 44px minimum size for touch accessibility")
		}
	})
	
	t.Run("has focus outline offset", func(t *testing.T) {
		if !strings.Contains(cssContent, "outline-offset:") {
			t.Error("theme-toggle:focus should have outline-offset for better visibility")
		}
	})
}
