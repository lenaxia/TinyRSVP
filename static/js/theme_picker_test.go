package js

import (
	"os"
	"strings"
	"testing"
)

func TestThemePickerStructure(t *testing.T) {
	content, err := os.ReadFile("theme_picker.js")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	jsContent := string(content)

	t.Run("has ThemePicker class", func(t *testing.T) {
		if !strings.Contains(jsContent, "class ThemePicker") {
			t.Error("JavaScript must contain ThemePicker class")
		}
	})

	t.Run("has constructor", func(t *testing.T) {
		if !strings.Contains(jsContent, "constructor()") {
			t.Error("ThemePicker must have constructor")
		}

		if !strings.Contains(jsContent, "this.gallery") {
			t.Error("Constructor should initialize gallery property")
		}

		if !strings.Contains(jsContent, "this.filterSelect") {
			t.Error("Constructor should initialize filterSelect property")
		}

		if !strings.Contains(jsContent, "this.hiddenInput") {
			t.Error("Constructor should initialize hiddenInput property")
		}
	})

	t.Run("has init method", func(t *testing.T) {
		if !strings.Contains(jsContent, "init()") {
			t.Error("ThemePicker must have init() method")
		}

		if !strings.Contains(jsContent, "attachEventListeners") {
			t.Error("init should call attachEventListeners")
		}

		if !strings.Contains(jsContent, "initializeKeyboardNavigation") {
			t.Error("init should call initializeKeyboardNavigation")
		}
	})

	t.Run("has attachEventListeners method", func(t *testing.T) {
		if !strings.Contains(jsContent, "attachEventListeners()") {
			t.Error("ThemePicker must have attachEventListeners() method")
		}

		if !strings.Contains(jsContent, "addEventListener") {
			t.Error("attachEventListeners should add event listeners")
		}
	})

	t.Run("has selectTheme method", func(t *testing.T) {
		if !strings.Contains(jsContent, "selectTheme(") {
			t.Error("ThemePicker must have selectTheme() method")
		}

		if !strings.Contains(jsContent, "classList.remove('selected')") {
			t.Error("selectTheme should remove 'selected' class from previous selection")
		}

		if !strings.Contains(jsContent, "classList.add('selected')") {
			t.Error("selectTheme should add 'selected' class to new selection")
		}

		if !strings.Contains(jsContent, "setAttribute('aria-checked'") {
			t.Error("selectTheme should update aria-checked attribute")
		}

		if !strings.Contains(jsContent, "setAttribute('tabindex'") {
			t.Error("selectTheme should update tabindex for keyboard navigation")
		}
	})

	t.Run("has filterThemes method", func(t *testing.T) {
		if !strings.Contains(jsContent, "filterThemes(") {
			t.Error("ThemePicker must have filterThemes() method")
		}

		if !strings.Contains(jsContent, "data-category") {
			t.Error("filterThemes should check data-category attribute")
		}

		if !strings.Contains(jsContent, "style.display") {
			t.Error("filterThemes should toggle display style")
		}
	})

	t.Run("has previewTheme method", func(t *testing.T) {
		if !strings.Contains(jsContent, "previewTheme(") {
			t.Error("ThemePicker must have previewTheme() method")
		}

		if !strings.Contains(jsContent, "CustomEvent") {
			t.Error("previewTheme should dispatch CustomEvent")
		}

		if !strings.Contains(jsContent, "theme-preview-requested") {
			t.Error("previewTheme should dispatch 'theme-preview-requested' event")
		}
	})

	t.Run("has initializeKeyboardNavigation method", func(t *testing.T) {
		if !strings.Contains(jsContent, "initializeKeyboardNavigation()") {
			t.Error("ThemePicker must have initializeKeyboardNavigation() method")
		}

		if !strings.Contains(jsContent, "keydown") {
			t.Error("initializeKeyboardNavigation should listen for keydown events")
		}
	})

	t.Run("has announceSelection method", func(t *testing.T) {
		if !strings.Contains(jsContent, "announceSelection(") {
			t.Error("ThemePicker must have announceSelection() method")
		}

		if !strings.Contains(jsContent, "role") && !strings.Contains(jsContent, "setAttribute('role'") {
			t.Error("announceSelection should set role attribute for screen readers")
		}

		if !strings.Contains(jsContent, "aria-live") {
			t.Error("announceSelection should use aria-live for screen reader announcements")
		}

		if !strings.Contains(jsContent, "sr-only") {
			t.Error("announceSelection should use sr-only class for screen reader only content")
		}
	})
}

func TestThemePickerEventHandling(t *testing.T) {
	content, err := os.ReadFile("theme_picker.js")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	jsContent := string(content)

	t.Run("handles filter change events", func(t *testing.T) {
		if !strings.Contains(jsContent, "filterSelect") {
			t.Error("Should reference filterSelect element")
		}

		if !strings.Contains(jsContent, "change") {
			t.Error("Should listen for change events on filter select")
		}

		if !strings.Contains(jsContent, "filterThemes") {
			t.Error("Should call filterThemes on filter change")
		}
	})

	t.Run("handles theme selection clicks", func(t *testing.T) {
		if !strings.Contains(jsContent, "btn-select") {
			t.Error("Should handle clicks on .btn-select buttons")
		}

		if !strings.Contains(jsContent, "data-theme-id") {
			t.Error("Should read data-theme-id attribute")
		}

		if !strings.Contains(jsContent, "closest") {
			t.Error("Should use closest() for event delegation")
		}
	})

	t.Run("handles preview button clicks", func(t *testing.T) {
		if !strings.Contains(jsContent, "btn-preview") {
			t.Error("Should handle clicks on .btn-preview buttons")
		}

		if !strings.Contains(jsContent, "previewTheme") {
			t.Error("Should call previewTheme on preview button click")
		}
	})

	t.Run("handles card clicks for selection", func(t *testing.T) {
		if !strings.Contains(jsContent, "theme-card") {
			t.Error("Should handle clicks on .theme-card elements")
		}
	})
}

func TestThemePickerKeyboardNavigation(t *testing.T) {
	content, err := os.ReadFile("theme_picker.js")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	jsContent := string(content)

	t.Run("handles Enter key", func(t *testing.T) {
		if !strings.Contains(jsContent, "Enter") {
			t.Error("Should handle Enter key")
		}

		if !strings.Contains(jsContent, "preventDefault") {
			t.Error("Should prevent default behavior for keyboard events")
		}
	})

	t.Run("handles Space key", func(t *testing.T) {
		if !strings.Contains(jsContent, " ") || !strings.Contains(jsContent, "' '") {
			t.Error("Should handle Space key")
		}
	})

	t.Run("handles arrow keys", func(t *testing.T) {
		if !strings.Contains(jsContent, "ArrowRight") {
			t.Error("Should handle ArrowRight key")
		}

		if !strings.Contains(jsContent, "ArrowLeft") {
			t.Error("Should handle ArrowLeft key")
		}

		if !strings.Contains(jsContent, "ArrowDown") {
			t.Error("Should handle ArrowDown key")
		}

		if !strings.Contains(jsContent, "ArrowUp") {
			t.Error("Should handle ArrowUp key")
		}
	})

	t.Run("handles Home key", func(t *testing.T) {
		if !strings.Contains(jsContent, "Home") {
			t.Error("Should handle Home key to jump to first theme")
		}
	})

	t.Run("handles End key", func(t *testing.T) {
		if !strings.Contains(jsContent, "End") {
			t.Error("Should handle End key to jump to last theme")
		}
	})

	t.Run("navigates to next visible card", func(t *testing.T) {
		if !strings.Contains(jsContent, "nextElementSibling") {
			t.Error("Should use nextElementSibling for forward navigation")
		}

		if !strings.Contains(jsContent, "previousElementSibling") {
			t.Error("Should use previousElementSibling for backward navigation")
		}

		if !strings.Contains(jsContent, "display") && !strings.Contains(jsContent, "style.display") {
			t.Error("Should check display style to skip hidden cards")
		}
	})

	t.Run("focuses navigated card", func(t *testing.T) {
		if !strings.Contains(jsContent, "focus()") {
			t.Error("Should call focus() on navigated card")
		}
	})
}

func TestThemePickerAccessibility(t *testing.T) {
	content, err := os.ReadFile("theme_picker.js")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	jsContent := string(content)

	t.Run("updates aria-checked attribute", func(t *testing.T) {
		if !strings.Contains(jsContent, "aria-checked") {
			t.Error("Should update aria-checked attribute for radio group semantics")
		}

		if !strings.Contains(jsContent, "true") && !strings.Contains(jsContent, "'true'") {
			t.Error("Should set aria-checked to 'true' for selected theme")
		}

		if !strings.Contains(jsContent, "false") && !strings.Contains(jsContent, "'false'") {
			t.Error("Should set aria-checked to 'false' for unselected themes")
		}
	})

	t.Run("manages tabindex for roving tabindex pattern", func(t *testing.T) {
		if !strings.Contains(jsContent, "tabindex") {
			t.Error("Should manage tabindex attribute")
		}

		if !strings.Contains(jsContent, "0") || !strings.Contains(jsContent, "'0'") {
			t.Error("Should set tabindex='0' for focusable selected theme")
		}

		if !strings.Contains(jsContent, "-1") || !strings.Contains(jsContent, "'-1'") {
			t.Error("Should set tabindex='-1' for non-focusable unselected themes")
		}
	})

	t.Run("announces selection to screen readers", func(t *testing.T) {
		if !strings.Contains(jsContent, "announceSelection") {
			t.Error("Should call announceSelection when theme is selected")
		}

		if !strings.Contains(jsContent, "theme selected") || !strings.Contains(jsContent, "selected") {
			t.Error("Announcement should indicate theme was selected")
		}

		if !strings.Contains(jsContent, "setTimeout") {
			t.Error("Should use setTimeout to remove announcement element after delay")
		}

		if !strings.Contains(jsContent, "remove()") {
			t.Error("Should remove announcement element after timeout")
		}
	})

	t.Run("creates live region for announcements", func(t *testing.T) {
		if !strings.Contains(jsContent, "createElement") {
			t.Error("Should create element for screen reader announcement")
		}

		if !strings.Contains(jsContent, "status") || !strings.Contains(jsContent, "'status'") {
			t.Error("Should use role='status' for polite announcements")
		}

		if !strings.Contains(jsContent, "polite") || !strings.Contains(jsContent, "'polite'") {
			t.Error("Should use aria-live='polite' for non-intrusive announcements")
		}
	})
}

func TestThemePickerInitialization(t *testing.T) {
	content, err := os.ReadFile("theme_picker.js")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	jsContent := string(content)

	t.Run("initializes on DOM ready", func(t *testing.T) {
		if !strings.Contains(jsContent, "DOMContentLoaded") {
			t.Error("ThemePicker should initialize on DOMContentLoaded")
		}

		if !strings.Contains(jsContent, "new ThemePicker()") {
			t.Error("Should instantiate ThemePicker")
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

	t.Run("checks for gallery existence", func(t *testing.T) {
		initSection := extractInitMethod(jsContent)

		if !strings.Contains(initSection, "if (!this.gallery)") && !strings.Contains(initSection, "if(!this.gallery)") {
			t.Error("init should check if gallery exists before proceeding")
		}

		if !strings.Contains(initSection, "return") {
			t.Error("init should return early if gallery doesn't exist")
		}
	})
}

func TestThemePickerHiddenInputUpdate(t *testing.T) {
	content, err := os.ReadFile("theme_picker.js")
	if err != nil {
		t.Fatalf("Failed to read theme_picker.js: %v", err)
	}

	jsContent := string(content)

	t.Run("updates hidden input value", func(t *testing.T) {
		selectThemeSection := extractSelectThemeMethod(jsContent)

		if !strings.Contains(selectThemeSection, "this.hiddenInput") {
			t.Error("selectTheme should reference hiddenInput")
		}

		if !strings.Contains(selectThemeSection, ".value") {
			t.Error("selectTheme should update hiddenInput value")
		}

		if !strings.Contains(selectThemeSection, "themeId") {
			t.Error("selectTheme should set value to themeId")
		}
	})

	t.Run("checks hidden input exists", func(t *testing.T) {
		selectThemeSection := extractSelectThemeMethod(jsContent)

		if !strings.Contains(selectThemeSection, "if (this.hiddenInput)") && !strings.Contains(selectThemeSection, "if(this.hiddenInput)") {
			t.Error("selectTheme should check if hiddenInput exists before updating")
		}
	})
}

func extractSelectThemeMethod(js string) string {
	// Look for the method definition (not a call site).
	// The definition is "selectTheme(themeId) {" (with a { on the same line or next).
	// We search for the pattern where selectTheme( is followed by a { before a ;
	idx := 0
	for {
		startIdx := strings.Index(js[idx:], "selectTheme(")
		if startIdx == -1 {
			return ""
		}
		startIdx += idx

		// Check if this is a method definition by looking ahead for '{' before ';' or next line
		rest := js[startIdx:]
		bracePos := strings.Index(rest, "{")
		semiPos := strings.Index(rest, ";")
		// If { comes before ;, this is likely the method definition
		if bracePos != -1 && (semiPos == -1 || bracePos < semiPos) {
			// Extract the method body
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
		idx = startIdx + 1
	}
}
