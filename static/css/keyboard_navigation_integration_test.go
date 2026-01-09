package css

import (
	"os"
	"strings"
	"testing"
)

func TestKeyboardNavigationIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	keyboardNavContent, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	variablesStr := string(variablesContent)
	keyboardNavStr := string(keyboardNavContent)

	requiredVars := []string{
		"--color-border-focus",
		"--color-primary-600",
		"--spacing-2",
		"--spacing-4",
		"--radius-base",
		"--radius-sm",
		"--font-weight-medium",
		"--transition-fast",
		"--z-index-tooltip",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(keyboardNavStr, "var("+varName+")") {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Keyboard navigation uses %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestKeyboardNavigationAccessibilityIntegration(t *testing.T) {
	keyboardNavContent, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	keyboardNavStr := string(keyboardNavContent)

	t.Run("skip_link_accessible", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, ".skip-link") {
			t.Error("Should have skip link for keyboard navigation")
		}
		if !strings.Contains(keyboardNavStr, ".skip-link:focus") {
			t.Error("Skip link should be visible on focus")
		}
	})

	t.Run("focus_indicators_visible", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, "outline: 2px solid") {
			t.Error("Focus indicators should be at least 2px solid for visibility")
		}
		if !strings.Contains(keyboardNavStr, "outline-offset: 2px") {
			t.Error("Focus indicators should have outline-offset for clarity")
		}
	})

	t.Run("focus_visible_support", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, ":focus-visible") {
			t.Error("Should support :focus-visible for better mouse/keyboard UX")
		}
	})

	t.Run("programmatic_focus_handling", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, "[tabindex=\"-1\"]:focus") {
			t.Error("Should handle programmatic focus with tabindex=\"-1\"")
		}
	})
}

func TestKeyboardNavigationWithButtons(t *testing.T) {
	keyboardNavContent, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Skip("buttons.css not found, skipping button integration tests")
	}

	keyboardNavStr := string(keyboardNavContent)
	buttonsStr := string(buttonsContent)

	t.Run("button_focus_consistent", func(t *testing.T) {
		if strings.Contains(buttonsStr, ".btn:focus") && strings.Contains(keyboardNavStr, "button:focus") {
			t.Log("Button focus styles are defined in both files - ensure consistency")
		}
	})

	t.Run("focus_border_color_consistent", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, "var(--color-border-focus)") {
			t.Error("Keyboard navigation should use consistent focus border color")
		}
	})
}

func TestKeyboardNavigationWithForms(t *testing.T) {
	keyboardNavContent, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Skip("forms.css not found, skipping form integration tests")
	}

	keyboardNavStr := string(keyboardNavContent)
	formsStr := string(formsContent)

	t.Run("form_input_focus_consistent", func(t *testing.T) {
		if strings.Contains(formsStr, "input:focus") && strings.Contains(keyboardNavStr, "input:focus") {
			t.Log("Input focus styles defined in both files - ensure consistency")
		}
	})

	t.Run("focus_within_support", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, ":focus-within") {
			t.Log("Supports :focus-within for container focus indication")
		}
	})
}

func TestKeyboardNavigationFileSize(t *testing.T) {
	info, err := os.Stat("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to stat keyboard_navigation.css: %v", err)
	}

	maxSize := int64(5 * 1024)
	if info.Size() > maxSize {
		t.Errorf("keyboard_navigation.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}

func TestKeyboardNavigationReducedMotion(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@media (prefers-reduced-motion: reduce)") {
		t.Error("Should respect prefers-reduced-motion for accessibility")
	}
}

func TestKeyboardNavigationFocusOrder(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.css: %v", err)
	}

	cssContent := string(content)

	skipLinkIndex := strings.Index(cssContent, ".skip-link")
	if skipLinkIndex == -1 {
		t.Fatal(".skip-link not found")
	}

	t.Run("skip_link_appears_first", func(t *testing.T) {
		if skipLinkIndex > 100 {
			t.Error("Skip link should be defined early in the file for logical tab order")
		}
	})
}
