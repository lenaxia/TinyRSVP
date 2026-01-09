package js

import (
	"os"
	"strings"
	"testing"
)

func TestKeyboardNavigationJSExists(t *testing.T) {
	if _, err := os.Stat("keyboard_navigation.js"); os.IsNotExist(err) {
		t.Fatal("keyboard_navigation.js does not exist")
	}
}

func TestKeyboardNavigationJSStructure(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"KeyboardNav object", "const KeyboardNav", true},
		{"handleEscape function", "handleEscape", true},
		{"handleTab function", "handleTab", true},
		{"handleArrowKeys function", "handleArrowKeys", true},
		{"trapFocus function", "trapFocus", true},
		{"init function", "init", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in keyboard_navigation.js", tt.pattern)
				}
			}
		})
	}
}

func TestKeyboardNavigationJSEscapeKey(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "Escape") || !strings.Contains(jsContent, "key") {
		t.Error("JavaScript should handle Escape key")
	}
}

func TestKeyboardNavigationJSTabKey(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "Tab") {
		t.Error("JavaScript should handle Tab key")
	}

	if !strings.Contains(jsContent, "shiftKey") {
		t.Error("JavaScript should handle Shift+Tab for reverse navigation")
	}
}

func TestKeyboardNavigationJSArrowKeys(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	arrowKeys := []string{"ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight"}

	for _, key := range arrowKeys {
		t.Run("handles_"+key, func(t *testing.T) {
			if !strings.Contains(jsContent, key) {
				t.Errorf("JavaScript should handle %s key", key)
			}
		})
	}
}

func TestKeyboardNavigationJSEnterSpace(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "Enter") {
		t.Error("JavaScript should handle Enter key")
	}

	if !strings.Contains(jsContent, "Space") || !strings.Contains(jsContent, " ") {
		t.Error("JavaScript should handle Space key")
	}
}

func TestKeyboardNavigationJSFocusTrap(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "trapFocus") {
		t.Error("JavaScript should have trapFocus function")
	}

	if !strings.Contains(jsContent, "querySelectorAll") {
		t.Error("trapFocus should query for focusable elements")
	}
}

func TestKeyboardNavigationJSFocusableSelector(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	focusableElements := []string{
		"a[href]",
		"button",
		"input",
		"textarea",
		"select",
		"[tabindex]",
	}

	for _, element := range focusableElements {
		t.Run("selector_includes_"+element, func(t *testing.T) {
			if !strings.Contains(jsContent, element) {
				t.Errorf("Focusable selector should include %s", element)
			}
		})
	}
}

func TestKeyboardNavigationJSNoConsoleLog(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestKeyboardNavigationJSFileSize(t *testing.T) {
	info, err := os.Stat("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to stat keyboard_navigation.js: %v", err)
	}

	maxSize := int64(15 * 1024)
	if info.Size() > maxSize {
		t.Errorf("keyboard_navigation.js is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestKeyboardNavigationJSValidSyntax(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	openBraces := strings.Count(jsContent, "{")
	closeBraces := strings.Count(jsContent, "}")

	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}

	openParens := strings.Count(jsContent, "(")
	closeParens := strings.Count(jsContent, ")")

	if openParens != closeParens {
		t.Errorf("Mismatched parentheses: %d open, %d close", openParens, closeParens)
	}
}

func TestKeyboardNavigationJSModuleExport(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "module.exports") {
		t.Error("JavaScript should export module for testing")
	}
}

func TestKeyboardNavigationJSDOMReady(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "DOMContentLoaded") {
		t.Error("JavaScript should wait for DOMContentLoaded before initializing")
	}
}

func TestKeyboardNavigationJSPreventDefault(t *testing.T) {
	content, err := os.ReadFile("keyboard_navigation.js")
	if err != nil {
		t.Fatalf("Failed to read keyboard_navigation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "preventDefault") {
		t.Error("JavaScript should use preventDefault for keyboard event handling")
	}
}
