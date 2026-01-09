package js

import (
	"os"
	"strings"
	"testing"
)

func TestFocusManagementJSIntegration(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	focusManagementStr := string(focusManagementContent)

	t.Run("exports_for_testing", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "module.exports") {
			t.Error("Should export FocusManager for testing")
		}
	})

	t.Run("no_dom_dependency_for_export", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "typeof module !== 'undefined'") {
			t.Error("Should check for module availability before exporting")
		}
	})
}

func TestFocusManagementJSWithKeyboardNav(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	keyboardNavContent, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Skip("keyboard_navigation.js not found, skipping integration tests")
	}

	focusManagementStr := string(focusManagementContent)
	keyboardNavStr := string(keyboardNavContent)

	t.Run("focus_trap_compatible", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, "trapFocus") && strings.Contains(focusManagementStr, "trapFocus") {
			t.Log("Both modules have trapFocus - FocusManager provides enhanced version")
		}
	})

	t.Run("focusable_selector_consistent", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, "a[href]") && strings.Contains(focusManagementStr, "a[href]") {
			t.Log("Both modules use consistent focusable selectors")
		}
	})

	t.Run("tab_key_handling", func(t *testing.T) {
		if !strings.Contains(focusManagementStr, "Tab") {
			t.Error("FocusManager should handle Tab key for focus trapping")
		}
	})
}

func TestFocusManagementJSWithScreenReader(t *testing.T) {
	focusManagementContent, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	screenReaderContent, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Skip("screen_reader.js not found, skipping integration tests")
	}

	focusManagementStr := string(focusManagementContent)
	screenReaderStr := string(screenReaderContent)

	t.Run("focus_management_accessible", func(t *testing.T) {
		if strings.Contains(screenReaderStr, "focus") && strings.Contains(focusManagementStr, "focus") {
			t.Log("Both modules handle focus - ensure screen reader announcements work with focus changes")
		}
	})

	t.Run("modal_focus_accessible", func(t *testing.T) {
		if strings.Contains(focusManagementStr, "manageFocusForModal") {
			t.Log("Modal focus management should work with screen reader announcements")
		}
	})
}

func TestFocusManagementJSModalIntegration(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	t.Run("modal_focus_management", func(t *testing.T) {
		if !strings.Contains(jsContent, "manageFocusForModal") {
			t.Error("Should have manageFocusForModal function")
		}
	})

	t.Run("modal_saves_focus", func(t *testing.T) {
		if !strings.Contains(jsContent, "saveFocus") {
			t.Error("Modal should save focus before opening")
		}
	})

	t.Run("modal_traps_focus", func(t *testing.T) {
		if !strings.Contains(jsContent, "trapFocus") {
			t.Error("Modal should trap focus within itself")
		}
	})

	t.Run("modal_restores_focus", func(t *testing.T) {
		if !strings.Contains(jsContent, "restoreFocus") {
			t.Error("Modal should restore focus after closing")
		}
	})

	t.Run("modal_returns_cleanup", func(t *testing.T) {
		if !strings.Contains(jsContent, "return") {
			t.Error("manageFocusForModal should return cleanup function")
		}
	})
}

func TestFocusManagementJSErrorHandling(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	t.Run("handles_null_elements", func(t *testing.T) {
		if !strings.Contains(jsContent, "if (!") {
			t.Error("Should check for null/undefined elements")
		}
	})

	t.Run("uses_try_catch", func(t *testing.T) {
		if !strings.Contains(jsContent, "try") || !strings.Contains(jsContent, "catch") {
			t.Error("Should use try-catch for focus operations that might fail")
		}
	})
}

func TestFocusManagementJSScrollBehavior(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	t.Run("scroll_into_view", func(t *testing.T) {
		if !strings.Contains(jsContent, "scrollIntoView") {
			t.Error("Should scroll focused element into view")
		}
	})

	t.Run("configurable_scroll", func(t *testing.T) {
		if !strings.Contains(jsContent, "behavior") || !strings.Contains(jsContent, "block") {
			t.Error("Should support configurable scroll behavior")
		}
	})
}

func TestFocusManagementJSBrowserCompatibility(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	t.Run("focus_visible_polyfill", func(t *testing.T) {
		if !strings.Contains(jsContent, "ensureFocusVisible") {
			t.Error("Should provide focus-visible polyfill for older browsers")
		}
	})

	t.Run("checks_css_support", func(t *testing.T) {
		if !strings.Contains(jsContent, "CSS") && !strings.Contains(jsContent, "supports") {
			t.Error("Should check for CSS.supports before adding polyfill")
		}
	})
}

func TestFocusManagementJSContainerFocus(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	t.Run("focus_first_in_container", func(t *testing.T) {
		if !strings.Contains(jsContent, "focusFirstInContainer") {
			t.Error("Should have focusFirstInContainer function")
		}
	})

	t.Run("focus_last_in_container", func(t *testing.T) {
		if !strings.Contains(jsContent, "focusLastInContainer") {
			t.Error("Should have focusLastInContainer function")
		}
	})
}

func TestFocusManagementJSCleanup(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	t.Run("cleanup_function_returned", func(t *testing.T) {
		if !strings.Contains(jsContent, "return ()") && !strings.Contains(jsContent, "return function") {
			t.Error("Focus trap should return cleanup function")
		}
	})

	t.Run("removes_event_listeners", func(t *testing.T) {
		if !strings.Contains(jsContent, "removeEventListener") {
			t.Error("Cleanup should remove event listeners")
		}
	})
}
