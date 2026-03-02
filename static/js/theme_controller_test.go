package js

import (
	"os"
	"strings"
	"testing"
)

func TestThemeControllerStructure(t *testing.T) {
	content, err := os.ReadFile("theme_controller.js")
	if err != nil {
		t.Fatalf("Failed to read theme_controller.js: %v", err)
	}

	jsContent := string(content)

	t.Run("has ThemeController class", func(t *testing.T) {
		if !strings.Contains(jsContent, "class ThemeController") {
			t.Error("JavaScript must contain ThemeController class")
		}
	})

	t.Run("has STORAGE_KEY constant", func(t *testing.T) {
		if !strings.Contains(jsContent, "STORAGE_KEY") {
			t.Error("ThemeController must have STORAGE_KEY constant")
		}

		if !strings.Contains(jsContent, "tinyrsvp-theme") {
			t.Error("STORAGE_KEY should be 'tinyrsvp-theme'")
		}
	})

	t.Run("has THEMES constant with light and dark", func(t *testing.T) {
		if !strings.Contains(jsContent, "THEMES") {
			t.Error("ThemeController must have THEMES constant")
		}

		if !strings.Contains(jsContent, "LIGHT") || !strings.Contains(jsContent, "light") {
			t.Error("THEMES must include LIGHT theme")
		}

		if !strings.Contains(jsContent, "DARK") || !strings.Contains(jsContent, "dark") {
			t.Error("THEMES must include DARK theme")
		}
	})

	t.Run("has init method", func(t *testing.T) {
		if !strings.Contains(jsContent, "init()") {
			t.Error("ThemeController must have init() method")
		}
	})

	t.Run("has getSavedTheme method", func(t *testing.T) {
		if !strings.Contains(jsContent, "getSavedTheme()") {
			t.Error("ThemeController must have getSavedTheme() method")
		}

		if !strings.Contains(jsContent, "localStorage.getItem") {
			t.Error("getSavedTheme should use localStorage.getItem")
		}
	})

	t.Run("has getSystemTheme method", func(t *testing.T) {
		if !strings.Contains(jsContent, "getSystemTheme()") {
			t.Error("ThemeController must have getSystemTheme() method")
		}

		if !strings.Contains(jsContent, "prefers-color-scheme") {
			t.Error("getSystemTheme should check prefers-color-scheme media query")
		}

		if !strings.Contains(jsContent, "window.matchMedia") {
			t.Error("getSystemTheme should use window.matchMedia")
		}
	})

	t.Run("has setTheme method", func(t *testing.T) {
		if !strings.Contains(jsContent, "setTheme(") {
			t.Error("ThemeController must have setTheme() method")
		}

		if !strings.Contains(jsContent, "document.documentElement.setAttribute") {
			t.Error("setTheme should set data-theme attribute on documentElement")
		}

		if !strings.Contains(jsContent, "data-theme") {
			t.Error("setTheme should use 'data-theme' attribute")
		}

		if !strings.Contains(jsContent, "localStorage.setItem") {
			t.Error("setTheme should persist theme to localStorage")
		}
	})

	t.Run("has toggleTheme method", func(t *testing.T) {
		if !strings.Contains(jsContent, "toggleTheme()") {
			t.Error("ThemeController must have toggleTheme() method")
		}

		if !strings.Contains(jsContent, "document.documentElement.getAttribute") {
			t.Error("toggleTheme should get current theme from data-theme attribute")
		}
	})

	t.Run("has updateToggleButton method", func(t *testing.T) {
		if !strings.Contains(jsContent, "updateToggleButton(") {
			t.Error("ThemeController must have updateToggleButton() method")
		}

		if !strings.Contains(jsContent, "theme-toggle") {
			t.Error("updateToggleButton should reference theme-toggle button ID")
		}

		if !strings.Contains(jsContent, "aria-label") {
			t.Error("updateToggleButton should update aria-label for accessibility")
		}
	})

	t.Run("has attachEventListeners method", func(t *testing.T) {
		if !strings.Contains(jsContent, "attachEventListeners()") {
			t.Error("ThemeController must have attachEventListeners() method")
		}

		if !strings.Contains(jsContent, "addEventListener") {
			t.Error("attachEventListeners should add event listeners")
		}

		if !strings.Contains(jsContent, "click") {
			t.Error("attachEventListeners should listen for click events")
		}
	})
}

func TestThemeControllerInitialization(t *testing.T) {
	content, err := os.ReadFile("theme_controller.js")
	if err != nil {
		t.Fatalf("Failed to read theme_controller.js: %v", err)
	}

	jsContent := string(content)

	t.Run("initializes on DOM ready", func(t *testing.T) {
		if !strings.Contains(jsContent, "DOMContentLoaded") {
			t.Error("ThemeController should initialize on DOMContentLoaded")
		}

		if !strings.Contains(jsContent, "new ThemeController()") {
			t.Error("Should instantiate ThemeController")
		}
	})

	t.Run("handles already loaded DOM", func(t *testing.T) {
		if !strings.Contains(jsContent, "document.readyState") {
			t.Error("Should check document.readyState for already loaded DOM")
		}

		if !strings.Contains(jsContent, "loading") {
			t.Error("Should check if document is still loading")
		}
	})

	t.Run("has theme preference priority", func(t *testing.T) {
		initSection := extractInitMethod(jsContent)

		savedThemeIdx := strings.Index(initSection, "getSavedTheme")
		systemThemeIdx := strings.Index(initSection, "getSystemTheme")

		if savedThemeIdx == -1 || systemThemeIdx == -1 {
			t.Error("init should call both getSavedTheme and getSystemTheme")
		}

		if savedThemeIdx > systemThemeIdx {
			t.Error("init should check saved theme before system theme (priority: saved > system > default)")
		}
	})
}

func TestThemeControllerAccessibility(t *testing.T) {
	content, err := os.ReadFile("theme_controller.js")
	if err != nil {
		t.Fatalf("Failed to read theme_controller.js: %v", err)
	}

	jsContent := string(content)

	t.Run("updates button icon for visual feedback", func(t *testing.T) {
		if !strings.Contains(jsContent, "☀️") || !strings.Contains(jsContent, "🌙") {
			t.Error("updateToggleButton should use sun and moon icons")
		}
	})

	t.Run("updates screen reader text", func(t *testing.T) {
		if !strings.Contains(jsContent, "sr-only") || !strings.Contains(jsContent, ".sr-only") {
			t.Error("updateToggleButton should update screen reader only text")
		}

		if !strings.Contains(jsContent, "Switch to") {
			t.Error("Button labels should indicate what will happen (Switch to...)")
		}
	})

	t.Run("handles missing button gracefully", func(t *testing.T) {
		updateButtonSection := extractUpdateToggleButtonMethod(jsContent)

		if !strings.Contains(updateButtonSection, "if (!button)") && !strings.Contains(updateButtonSection, "if(!button)") {
			t.Error("updateToggleButton should check if button exists before updating")
		}

		if !strings.Contains(updateButtonSection, "return") {
			t.Error("updateToggleButton should return early if button doesn't exist")
		}
	})
}

func extractInitMethod(js string) string {
	startIdx := strings.Index(js, "init()")
	if startIdx == -1 {
		return ""
	}

	braceCount := 0
	inMethod := false
	var result strings.Builder

	for i := startIdx; i < len(js); i++ {
		char := js[i]
		result.WriteByte(char)

		if char == '{' {
			braceCount++
			inMethod = true
		} else if char == '}' {
			braceCount--
			if inMethod && braceCount == 0 {
				break
			}
		}
	}

	return result.String()
}

func extractUpdateToggleButtonMethod(js string) string {
	startIdx := strings.Index(js, "updateToggleButton(")
	if startIdx == -1 {
		return ""
	}

	braceCount := 0
	inMethod := false
	var result strings.Builder

	for i := startIdx; i < len(js); i++ {
		char := js[i]
		result.WriteByte(char)

		if char == '{' {
			braceCount++
			inMethod = true
		} else if char == '}' {
			braceCount--
			if inMethod && braceCount == 0 {
				break
			}
		}
	}

	return result.String()
}
