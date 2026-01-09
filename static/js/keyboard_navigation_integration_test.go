package js

import (
	"os"
	"strings"
	"testing"
)

func TestKeyboardNavigationJSIntegration(t *testing.T) {
	keyboardNavContent, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	keyboardNavStr := string(keyboardNavContent)

	t.Run("exports_for_testing", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, "module.exports") {
			t.Error("Should export KeyboardNav for testing")
		}
	})

	t.Run("browser_compatible", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, "typeof document !== 'undefined'") {
			t.Error("Should check for document availability for browser compatibility")
		}
	})

	t.Run("initializes_on_dom_ready", func(t *testing.T) {
		if !strings.Contains(keyboardNavStr, "DOMContentLoaded") {
			t.Error("Should initialize on DOMContentLoaded")
		}
		if !strings.Contains(keyboardNavStr, "KeyboardNav.init()") {
			t.Error("Should call init() on DOMContentLoaded")
		}
	})
}

func TestKeyboardNavigationJSWithFormValidation(t *testing.T) {
	keyboardNavContent, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	formValidationContent, err := os.ReadFile("form_validation.js")
	if err != nil {
		t.Skip("form_validation.js not found, skipping integration tests")
	}

	keyboardNavStr := string(keyboardNavContent)
	formValidationStr := string(formValidationContent)

	t.Run("focus_management_compatible", func(t *testing.T) {
		if strings.Contains(formValidationStr, ".focus()") && strings.Contains(keyboardNavStr, "focus()") {
			t.Log("Both modules handle focus - ensure they work together")
		}
	})

	t.Run("error_focus_compatible", func(t *testing.T) {
		if strings.Contains(formValidationStr, "firstError.focus()") {
			t.Log("Form validation focuses first error - keyboard nav should support this")
		}
	})
}

func TestKeyboardNavigationJSWithLoadingStates(t *testing.T) {
	keyboardNavContent, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	loadingStatesContent, err := os.ReadFile("loading_states.js")
	if err != nil {
		t.Skip("loading_states.js not found, skipping integration tests")
	}

	keyboardNavStr := string(keyboardNavContent)
	loadingStatesStr := string(loadingStatesContent)

	t.Run("disabled_elements_excluded", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, ":not([disabled])") {
			t.Log("Keyboard nav correctly excludes disabled elements")
		} else {
			t.Error("Should exclude disabled elements from keyboard navigation")
		}
	})

	t.Run("loading_state_compatible", func(t *testing.T) {
		if strings.Contains(loadingStatesStr, "disabled") && strings.Contains(keyboardNavStr, ":not([disabled])") {
			t.Log("Loading states and keyboard nav handle disabled elements consistently")
		}
	})
}

func TestKeyboardNavigationJSModalHandling(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	t.Run("modal_escape_closes", func(t *testing.T) {
		if !strings.Contains(jsContent, "Escape") {
			t.Error("Should handle Escape key to close modals")
		}
	})

	t.Run("modal_focus_trap", func(t *testing.T) {
		if !strings.Contains(jsContent, "trapFocus") {
			t.Error("Should trap focus within modals")
		}
	})

	t.Run("modal_focus_restore", func(t *testing.T) {
		if !strings.Contains(jsContent, "saveFocus") && !strings.Contains(jsContent, "restoreFocus") {
			t.Error("Should save and restore focus when modal closes")
		}
	})

	t.Run("modal_role_support", func(t *testing.T) {
		if !strings.Contains(jsContent, "role=\"dialog\"") && !strings.Contains(jsContent, "[role=\"dialog\"]") {
			t.Error("Should support ARIA role=\"dialog\" for modals")
		}
	})
}

func TestKeyboardNavigationJSDropdownHandling(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	t.Run("dropdown_arrow_navigation", func(t *testing.T) {
		if !strings.Contains(jsContent, "ArrowDown") || !strings.Contains(jsContent, "ArrowUp") {
			t.Error("Should handle arrow keys for dropdown navigation")
		}
	})

	t.Run("dropdown_escape_closes", func(t *testing.T) {
		if !strings.Contains(jsContent, "Escape") {
			t.Error("Should handle Escape key to close dropdowns")
		}
	})

	t.Run("dropdown_enter_activates", func(t *testing.T) {
		if !strings.Contains(jsContent, "Enter") {
			t.Error("Should handle Enter key to activate dropdown items")
		}
	})
}

func TestKeyboardNavigationJSCustomButtons(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	t.Run("role_button_support", func(t *testing.T) {
		if !strings.Contains(jsContent, "role=\"button\"") && !strings.Contains(jsContent, "[role=\"button\"]") {
			t.Error("Should support elements with role=\"button\"")
		}
	})

	t.Run("enter_space_activation", func(t *testing.T) {
		if !strings.Contains(jsContent, "Enter") || !strings.Contains(jsContent, "Space") {
			t.Error("Should activate custom buttons with Enter and Space keys")
		}
	})
}

func TestKeyboardNavigationJSFocusManagement(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	t.Run("move_focus_function", func(t *testing.T) {
		if !strings.Contains(jsContent, "moveFocusTo") {
			t.Error("Should have moveFocusTo function")
		}
	})

	t.Run("save_restore_focus", func(t *testing.T) {
		if !strings.Contains(jsContent, "saveFocus") || !strings.Contains(jsContent, "restoreFocus") {
			t.Error("Should have saveFocus and restoreFocus functions")
		}
	})

	t.Run("scroll_into_view", func(t *testing.T) {
		if !strings.Contains(jsContent, "scrollIntoView") {
			t.Error("Should scroll focused element into view")
		}
	})
}

func TestKeyboardNavigationJSNoGlobalState(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	lines := strings.Split(jsContent, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "var ") && !strings.Contains(line, "function") {
			t.Errorf("Line %d: Should use const or let instead of var: %s", i+1, trimmed)
		}
	}
}
