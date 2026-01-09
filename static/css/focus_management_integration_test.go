package css

import (
	"os"
	"strings"
	"testing"
)

func TestFocusManagementIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	focusManagementContent, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	variablesStr := string(variablesContent)
	focusManagementStr := string(focusManagementContent)

	requiredVars := []string{
		"--color-border-focus",
		"--color-primary-200",
		"--spacing-0",
		"--radius-sm",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(focusManagementStr, "var("+varName+")") {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Focus management uses %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestFocusManagementIntegrationWithKeyboardNav(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	keyboardNavContent, err := os.ReadFile("keyboard_navigation.css")
	if err != nil {
		t.Skip("keyboard_navigation.css not found, skipping integration tests")
	}

	focusManagementStr := string(focusManagementContent)
	keyboardNavStr := string(keyboardNavContent)

	t.Run("focus_indicators_consistent", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, ":focus") && strings.Contains(focusManagementStr, ":focus") {
			t.Log("Both files define focus styles - ensure consistency")
		}
	})

	t.Run("focus_visible_consistent", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, ":focus-visible") && strings.Contains(focusManagementStr, ":focus-visible") {
			t.Log("Both files use :focus-visible - ensure consistency")
		}
	})

	t.Run("focus_border_color_consistent", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "var(--color-border-focus)") {
			t.Error("Focus management should use consistent focus border color")
		}
	})
}

func TestFocusManagementIntegrationWithButtons(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Skip("buttons.css not found, skipping integration tests")
	}

	focusManagementStr := string(focusManagementContent)
	buttonsStr := string(buttonsContent)

	t.Run("button_focus_enhanced", func(t *testing.T) {
		if strings.Contains(buttonsStr, ".btn:focus") && strings.Contains(focusManagementStr, ".btn:focus") {
			t.Log("Both files define button focus - focus_management enhances buttons")
		}
	})

	t.Run("button_focus_box_shadow", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "box-shadow") {
			t.Error("Focus management should add box-shadow to button focus for enhanced visibility")
		}
	})
}

func TestFocusManagementIntegrationWithForms(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Skip("forms.css not found, skipping integration tests")
	}

	focusManagementStr := string(focusManagementContent)
	formsStr := string(formsContent)

	t.Run("form_group_focus_within", func(t *testing.T) {
		if strings.Contains(formsStr, ".form-group") && !strings.Contains(focusManagementStr, ".form-group:focus-within") {
			t.Error("Should have .form-group:focus-within to highlight focused form fields")
		}
	})

	t.Run("input_focus_consistent", func(t *testing.T) {
		if strings.Contains(formsStr, "input:focus") && strings.Contains(focusManagementStr, "input:focus") {
			t.Log("Both files define input focus - ensure consistency")
		}
	})
}

func TestFocusManagementAccessibility(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	focusManagementStr := string(focusManagementContent)

	t.Run("focus_indicators_visible", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "outline: 2px solid") {
			t.Error("Focus indicators should be at least 2px solid for visibility")
		}
	})

	t.Run("focus_indicators_contrast", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "var(--color-border-focus)") {
			t.Error("Focus indicators should use --color-border-focus for 3:1 contrast ratio")
		}
	})

	t.Run("screen_reader_only_class", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, ".sr-only") && !strings.Contains(focusManagementStr, ".visually-hidden") {
			t.Error("Should have screen reader only class for accessible hidden content")
		}
	})

	t.Run("focusable_when_focused", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "focusable:focus") || !strings.Contains(focusManagementStr, "-focusable:focus") {
			t.Error("Should allow screen reader only content to become visible when focused")
		}
	})
}

func TestFocusManagementMouseKeyboardUX(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	focusManagementStr := string(focusManagementContent)

	t.Run("hides_focus_for_mouse", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, ":focus:not(:focus-visible)") {
			t.Error("Should hide focus outline for mouse clicks with :focus:not(:focus-visible)")
		}
	})

	t.Run("shows_focus_for_keyboard", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, ":focus-visible") {
			t.Error("Should show focus outline for keyboard navigation with :focus-visible")
		}
	})
}

func TestFocusManagementNoOutlineNone(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	outlineNoneCount := strings.Count(cssContent, "outline: none")
	focusVisibleCount := strings.Count(cssContent, ":focus-visible")
	notFocusVisibleCount := strings.Count(cssContent, ":focus:not(:focus-visible)")

	if outlineNoneCount > 0 && (focusVisibleCount == 0 && notFocusVisibleCount == 0) {
		t.Error("Should not use outline: none without providing :focus-visible alternative")
	}
}

func TestFocusManagementCustomFocusRings(t *testing.T) {
	content, err := os.ReadFile("focus_management.css")
	if err != nil {
		t.Fatalf("Failed to read focus_management.css: %v", err)
	}

	cssContent := string(content)

	t.Run("focus_ring_utility", func(t *testing.T) {
		if !strings.Contains(cssContent, ".focus-ring") {
			t.Error("Should have .focus-ring utility class")
		}
	})

	t.Run("focus_ring_inset_utility", func(t *testing.T) {
		if !strings.Contains(cssContent, ".focus-ring-inset") {
			t.Error("Should have .focus-ring-inset utility class for inset focus rings")
		}
	})
}
