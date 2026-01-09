package js

import (
	"os"
	"strings"
	"testing"
)

func getLoadingStatesJSPath() string {
	return "loading_states.js"
}

func TestLoadingStatesJSExists(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Errorf("loading_states.js does not exist at %s", jsPath)
	}
}

func TestLoadingStatesJSStructure(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"LoadingStates object", "const LoadingStates", true},
		{"showButtonLoading function", "showButtonLoading", true},
		{"hideButtonLoading function", "hideButtonLoading", true},
		{"showSpinner function", "showSpinner", true},
		{"hideSpinner function", "hideSpinner", true},
		{"showOverlay function", "showOverlay", true},
		{"hideOverlay function", "hideOverlay", true},
		{"updateProgress function", "updateProgress", true},
		{"setLoadingState function", "setLoadingState", true},
		{"clearLoadingState function", "clearLoadingState", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in loading_states.js", tt.pattern)
				}
			}
		})
	}
}

func TestLoadingStatesJSNoConsoleLog(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestLoadingStatesJSARIASupport(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	ariaAttributes := []string{
		"aria-busy",
		"aria-live",
		"role",
	}

	for _, attr := range ariaAttributes {
		if !strings.Contains(jsContent, attr) {
			t.Errorf("JavaScript should support ARIA attribute: %s", attr)
		}
	}
}

func TestLoadingStatesJSDisableInteractions(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "disabled") {
		t.Error("JavaScript should disable elements during loading")
	}
}

func TestLoadingStatesJSClassManagement(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	classOperations := []string{
		"classList.add",
		"classList.remove",
	}

	for _, operation := range classOperations {
		if !strings.Contains(jsContent, operation) {
			t.Errorf("JavaScript should use %s for class management", operation)
		}
	}
}

func TestLoadingStatesJSTimeoutHandling(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "setTimeout") && !strings.Contains(jsContent, "timeout") {
		t.Error("JavaScript should support timeout handling for loading states")
	}
}

func TestLoadingStatesJSProgressBarSupport(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "progress") {
		t.Error("JavaScript should support progress bar updates")
	}

	if !strings.Contains(jsContent, "width") || !strings.Contains(jsContent, "%") {
		t.Error("JavaScript should update progress bar width as percentage")
	}
}

func TestLoadingStatesJSOverlayManagement(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "loading-overlay") {
		t.Error("JavaScript should manage loading overlay")
	}
}

func TestLoadingStatesJSFileSize(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	info, err := os.Stat(jsPath)
	if err != nil {
		t.Fatalf("Failed to stat loading_states.js: %v", err)
	}

	maxSize := int64(20 * 1024)
	if info.Size() > maxSize {
		t.Errorf("loading_states.js is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestLoadingStatesJSValidSyntax(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
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

func TestLoadingStatesJSErrorHandling(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "if") || !strings.Contains(jsContent, "return") {
		t.Error("JavaScript should include error handling and early returns")
	}
}

func TestLoadingStatesJSButtonStateManagement(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "loading") {
		t.Error("JavaScript should manage 'loading' class for buttons")
	}

	if !strings.Contains(jsContent, "textContent") && !strings.Contains(jsContent, "innerText") {
		t.Error("JavaScript should preserve button text during loading")
	}
}

func TestLoadingStatesJSSpinnerManagement(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "spinner") {
		t.Error("JavaScript should manage spinner elements")
	}

	if !strings.Contains(jsContent, "createElement") || !strings.Contains(jsContent, "remove") {
		t.Error("JavaScript should create and remove spinner elements dynamically")
	}
}

func TestLoadingStatesJSStateTracking(t *testing.T) {
	jsPath := getLoadingStatesJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read loading_states.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "Map") && !strings.Contains(jsContent, "Set") && !strings.Contains(jsContent, "{}") {
		t.Error("JavaScript should track loading states")
	}
}
