package js

import (
	"os"
	"strings"
	"testing"
)

func getJSPath() string {
	return "form_validation.js"
}

func TestFormValidationJSExists(t *testing.T) {
	jsPath := getJSPath()
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Errorf("form_validation.js does not exist at %s", jsPath)
	}
}

func TestFormValidationJSStructure(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"FormValidator object", "const FormValidator", true},
		{"validateEmail function", "validateEmail", true},
		{"validateRequired function", "validateRequired", true},
		{"validateDateTime function", "validateDateTime", true},
		{"showError function", "showError", true},
		{"clearError function", "clearError", true},
		{"showSuccess function", "showSuccess", true},
		{"clearSuccess function", "clearSuccess", true},
		{"init function", "init", true},
		{"validateField function", "validateField", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in form_validation.js", tt.pattern)
				}
			}
		})
	}
}

func TestFormValidationJSEmailRegex(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "@") {
		t.Error("Email validation regex should contain @ symbol check")
	}
}

func TestFormValidationJSNoConsoleLog(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestFormValidationJSProgressiveEnhancement(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "novalidate") {
		t.Error("JavaScript should handle novalidate attribute for progressive enhancement")
	}
}

func TestFormValidationJSEventListeners(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	requiredListeners := []string{
		"blur",
		"submit",
	}

	for _, listener := range requiredListeners {
		if !strings.Contains(jsContent, listener) {
			t.Errorf("JavaScript should contain '%s' event listener", listener)
		}
	}
}

func TestFormValidationJSErrorClasses(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "error") {
		t.Error("JavaScript should use 'error' class for styling")
	}

	if !strings.Contains(jsContent, "success") {
		t.Error("JavaScript should use 'success' class for styling")
	}
}

func TestFormValidationJSCustomMessages(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "data-error-message") && !strings.Contains(jsContent, "errorMessage") {
		t.Error("JavaScript should support custom error messages")
	}
}

func TestFormValidationJSFileSize(t *testing.T) {
	jsPath := getJSPath()
	info, err := os.Stat(jsPath)
	if err != nil {
		t.Fatalf("Failed to stat form_validation.js: %v", err)
	}

	maxSize := int64(25 * 1024)
	if info.Size() > maxSize {
		t.Errorf("form_validation.js is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestFormValidationJSValidSyntax(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
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
