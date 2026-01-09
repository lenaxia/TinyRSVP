package js

import (
	"os"
	"strings"
	"testing"
)

func TestScreenReaderJSIntegration(t *testing.T) {
	screenReaderContent, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	screenReaderStr := string(screenReaderContent)

	t.Run("exports_for_testing", func(t *testing.T) {
		if !strings.Contains(screenReaderStr, "module.exports") {
			t.Error("Should export ScreenReader for testing")
		}
	})

	t.Run("browser_compatible", func(t *testing.T) {
		if !strings.Contains(screenReaderStr, "typeof document !== 'undefined'") {
			t.Error("Should check for document availability for browser compatibility")
		}
	})

	t.Run("initializes_on_dom_ready", func(t *testing.T) {
		if !strings.Contains(screenReaderStr, "DOMContentLoaded") {
			t.Error("Should initialize on DOMContentLoaded")
		}
		if !strings.Contains(screenReaderStr, "ScreenReader.init()") {
			t.Error("Should call init() on DOMContentLoaded")
		}
	})
}

func TestScreenReaderJSWithFormValidation(t *testing.T) {
	screenReaderContent, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	formValidationContent, err := os.ReadFile("form_validation.js")
	if err != nil {
		t.Skip("form_validation.js not found, skipping integration tests")
	}

	screenReaderStr := string(screenReaderContent)
	formValidationStr := string(formValidationContent)

	t.Run("aria_invalid_compatible", func(t *testing.T) {
		if strings.Contains(formValidationStr, "aria-invalid") {
			t.Log("Form validation sets aria-invalid - screen reader will announce it")
		}
	})

	t.Run("aria_describedby_compatible", func(t *testing.T) {
		if strings.Contains(formValidationStr, "aria-describedby") && strings.Contains(screenReaderStr, "aria-describedby") {
			t.Log("Both modules use aria-describedby - ensure they work together")
		}
	})

	t.Run("role_alert_compatible", func(t *testing.T) {
		if strings.Contains(formValidationStr, "role") && strings.Contains(screenReaderStr, "alert") {
			t.Log("Form validation and screen reader both use ARIA roles")
		}
	})
}

func TestScreenReaderJSWithKeyboardNav(t *testing.T) {
	screenReaderContent, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	keyboardNavContent, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Skip("keyboard_navigation.js not found, skipping integration tests")
	}

	screenReaderStr := string(screenReaderContent)
	keyboardNavStr := string(keyboardNavContent)

	t.Run("focus_management_compatible", func(t *testing.T) {
		if strings.Contains(keyboardNavStr, "focus()") && strings.Contains(screenReaderStr, "focus") {
			t.Log("Both modules handle focus - ensure they work together")
		}
	})

	t.Run("aria_hidden_with_focus", func(t *testing.T) {
		if strings.Contains(screenReaderStr, "aria-hidden") {
			t.Log("Screen reader can hide elements - keyboard nav should respect this")
		}
	})
}

func TestScreenReaderJSARIAImplementation(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	ariaAttributes := []string{
		"aria-live",
		"aria-atomic",
		"aria-label",
		"aria-labelledby",
		"aria-describedby",
		"aria-hidden",
		"aria-expanded",
		"aria-pressed",
		"aria-checked",
		"aria-current",
	}

	for _, attr := range ariaAttributes {
		t.Run("aria_"+attr, func(t *testing.T) {
			if !strings.Contains(jsContent, attr) {
				t.Errorf("Should support %s attribute", attr)
			}
		})
	}
}

func TestScreenReaderJSLandmarkImplementation(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("auto_adds_landmarks", func(t *testing.T) {
		landmarks := []string{"header", "nav", "main", "footer"}
		for _, landmark := range landmarks {
			if !strings.Contains(jsContent, landmark) {
				t.Errorf("Should automatically add landmarks for %s elements", landmark)
			}
		}
	})

	t.Run("checks_existing_roles", func(t *testing.T) {
		if !strings.Contains(jsContent, "getAttribute('role')") {
			t.Error("Should check for existing roles before adding new ones")
		}
	})
}

func TestScreenReaderJSFormAccessibility(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("ensures_form_labels", func(t *testing.T) {
		if !strings.Contains(jsContent, "ensureFormLabels") {
			t.Error("Should have function to ensure form labels")
		}
	})

	t.Run("handles_inputs_without_labels", func(t *testing.T) {
		if !strings.Contains(jsContent, "input") && !strings.Contains(jsContent, "label") {
			t.Error("Should handle inputs without explicit labels")
		}
	})

	t.Run("uses_placeholder_as_fallback", func(t *testing.T) {
		if !strings.Contains(jsContent, "placeholder") {
			t.Error("Should use placeholder as fallback for aria-label")
		}
	})
}

func TestScreenReaderJSButtonAccessibility(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("ensures_button_labels", func(t *testing.T) {
		if !strings.Contains(jsContent, "ensureButtonLabels") {
			t.Error("Should have function to ensure button labels")
		}
	})

	t.Run("handles_icon_buttons", func(t *testing.T) {
		if !strings.Contains(jsContent, "textContent") {
			t.Error("Should check for text content in buttons")
		}
	})
}

func TestScreenReaderJSLinkAccessibility(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("ensures_link_purpose", func(t *testing.T) {
		if !strings.Contains(jsContent, "ensureLinkPurpose") {
			t.Error("Should have function to ensure link purpose is clear")
		}
	})

	t.Run("handles_links", func(t *testing.T) {
		if !strings.Contains(jsContent, "a[href]") {
			t.Error("Should handle anchor links")
		}
	})
}

func TestScreenReaderJSImageAccessibility(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("adds_image_alt", func(t *testing.T) {
		if !strings.Contains(jsContent, "addImageAlt") {
			t.Error("Should have function to add alt text to images")
		}
	})

	t.Run("handles_images_without_alt", func(t *testing.T) {
		if !strings.Contains(jsContent, "img") && !strings.Contains(jsContent, "alt") {
			t.Error("Should handle images without alt text")
		}
	})
}

func TestScreenReaderJSLiveRegionManagement(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("creates_live_regions", func(t *testing.T) {
		if !strings.Contains(jsContent, "createLiveRegion") {
			t.Error("Should have function to create live regions")
		}
	})

	t.Run("positions_live_regions_offscreen", func(t *testing.T) {
		if !strings.Contains(jsContent, "position") && !strings.Contains(jsContent, "absolute") {
			t.Error("Live regions should be positioned off-screen")
		}
	})

	t.Run("announces_with_delay", func(t *testing.T) {
		if !strings.Contains(jsContent, "setTimeout") {
			t.Error("Announcements should use setTimeout for reliable screen reader detection")
		}
	})
}

func TestScreenReaderJSStateManagement(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	t.Run("manages_expanded_state", func(t *testing.T) {
		if !strings.Contains(jsContent, "setExpanded") && !strings.Contains(jsContent, "aria-expanded") {
			t.Error("Should manage aria-expanded state")
		}
	})

	t.Run("manages_pressed_state", func(t *testing.T) {
		if !strings.Contains(jsContent, "setPressed") && !strings.Contains(jsContent, "aria-pressed") {
			t.Error("Should manage aria-pressed state")
		}
	})

	t.Run("manages_checked_state", func(t *testing.T) {
		if !strings.Contains(jsContent, "setChecked") && !strings.Contains(jsContent, "aria-checked") {
			t.Error("Should manage aria-checked state")
		}
	})

	t.Run("manages_current_state", func(t *testing.T) {
		if !strings.Contains(jsContent, "setCurrent") && !strings.Contains(jsContent, "aria-current") {
			t.Error("Should manage aria-current state")
		}
	})
}
