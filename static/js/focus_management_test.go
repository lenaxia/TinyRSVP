package js

import (
	"os"
	"strings"
	"testing"
)

func TestFocusManagementJSExists(t *testing.T) {
	if _, err := os.Stat("focus_management.js"); os.IsNotExist(err) {
		t.Fatal("focus_management.js does not exist")
	}
}

func TestFocusManagementJSStructure(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"FocusManager object", "const FocusManager", true},
		{"saveFocus function", "saveFocus", true},
		{"restoreFocus function", "restoreFocus", true},
		{"moveFocusTo function", "moveFocusTo", true},
		{"trapFocus function", "trapFocus", true},
		{"releaseFocusTrap function", "releaseFocusTrap", true},
		{"getFocusableElements function", "getFocusableElements", true},
		{"getFirstFocusable function", "getFirstFocusable", true},
		{"getLastFocusable function", "getLastFocusable", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in focus_management.js", tt.pattern)
				}
			}
		})
	}
}

func TestFocusManagementJSSaveFocus(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "document.activeElement") {
		t.Error("saveFocus should save document.activeElement")
	}
}

func TestFocusManagementJSRestoreFocus(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "focus()") {
		t.Error("restoreFocus should call focus() on saved element")
	}
}

func TestFocusManagementJSMoveFocusTo(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "scrollIntoView") {
		t.Error("moveFocusTo should scroll element into view")
	}
}

func TestFocusManagementJSTrapFocus(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "trapFocus") {
		t.Error("Should have trapFocus function")
	}

	if !strings.Contains(jsContent, "Tab") {
		t.Error("trapFocus should handle Tab key")
	}

	if !strings.Contains(jsContent, "shiftKey") {
		t.Error("trapFocus should handle Shift+Tab")
	}
}

func TestFocusManagementJSReleaseTrap(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "releaseFocusTrap") {
		t.Error("Should have releaseFocusTrap function to clean up event listeners")
	}

	if !strings.Contains(jsContent, "removeEventListener") {
		t.Error("releaseFocusTrap should remove event listeners")
	}
}

func TestFocusManagementJSFocusableSelector(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
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

func TestFocusManagementJSExcludesDisabled(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, ":not([disabled])") {
		t.Error("Focusable selector should exclude disabled elements")
	}
}

func TestFocusManagementJSExcludesNegativeTabindex(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "[tabindex=\"-1\"]") || !strings.Contains(jsContent, "not") {
		t.Error("Focusable selector should exclude elements with tabindex=\"-1\"")
	}
}

func TestFocusManagementJSNoConsoleLog(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestFocusManagementJSFileSize(t *testing.T) {
	info, err := os.Stat("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to stat focus_management.js: %v", err)
	}

	maxSize := int64(10 * 1024)
	if info.Size() > maxSize {
		t.Errorf("focus_management.js is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestFocusManagementJSValidSyntax(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
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

func TestFocusManagementJSModuleExport(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "module.exports") {
		t.Error("JavaScript should export module for testing")
	}
}

func TestFocusManagementJSPreventDefault(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "preventDefault") {
		t.Error("Focus trap should use preventDefault for Tab key handling")
	}
}

func TestFocusManagementJSCircularNavigation(t *testing.T) {
	content, err := os.ReadFile("focus_management.js")
	if err != nil {
		t.Fatalf("Failed to read focus_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "firstFocusable") || !strings.Contains(jsContent, "lastFocusable") {
		t.Error("Focus trap should handle circular navigation between first and last focusable elements")
	}
}
