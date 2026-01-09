package js

import (
	"os"
	"strings"
	"testing"
)

func TestScreenReaderJSExists(t *testing.T) {
	if _, err := os.Stat("screen_reader.js"); os.IsNotExist(err) {
		t.Fatal("screen_reader.js does not exist")
	}
}

func TestScreenReaderJSStructure(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"ScreenReader object", "const ScreenReader", true},
		{"announce function", "announce", true},
		{"setLiveRegion function", "setLiveRegion", true},
		{"addLandmark function", "addLandmark", true},
		{"addLabel function", "addLabel", true},
		{"addDescription function", "addDescription", true},
		{"init function", "init", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in screen_reader.js", tt.pattern)
				}
			}
		})
	}
}

func TestScreenReaderJSARIALive(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	ariaLiveValues := []string{"polite", "assertive", "off"}

	for _, value := range ariaLiveValues {
		t.Run("aria_live_"+value, func(t *testing.T) {
			if !strings.Contains(jsContent, value) {
				t.Errorf("Should support aria-live=\"%s\"", value)
			}
		})
	}
}

func TestScreenReaderJSARIAAtomic(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "aria-atomic") {
		t.Error("Should support aria-atomic attribute")
	}
}

func TestScreenReaderJSARIALabel(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "aria-label") {
		t.Error("Should support aria-label attribute")
	}
}

func TestScreenReaderJSARIALabelledBy(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "aria-labelledby") {
		t.Error("Should support aria-labelledby attribute")
	}
}

func TestScreenReaderJSARIADescribedBy(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "aria-describedby") {
		t.Error("Should support aria-describedby attribute")
	}
}

func TestScreenReaderJSARIAHidden(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "aria-hidden") {
		t.Error("Should support aria-hidden attribute")
	}
}

func TestScreenReaderJSRoles(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	roles := []string{"status", "alert", "region", "navigation", "main", "banner", "contentinfo"}

	for _, role := range roles {
		t.Run("role_"+role, func(t *testing.T) {
			if !strings.Contains(jsContent, role) {
				t.Errorf("Should support role=\"%s\"", role)
			}
		})
	}
}

func TestScreenReaderJSAnnounce(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "announce") {
		t.Error("Should have announce function for screen reader announcements")
	}

	if !strings.Contains(jsContent, "createElement") {
		t.Error("announce function should create elements for live regions")
	}
}

func TestScreenReaderJSLiveRegion(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	hasStatus := strings.Contains(jsContent, "'status'") || strings.Contains(jsContent, "\"status\"")
	hasAlert := strings.Contains(jsContent, "'alert'") || strings.Contains(jsContent, "\"alert\"")
	
	if !hasStatus || !hasAlert {
		t.Error("Should create live regions with appropriate roles (status and alert)")
	}
}

func TestScreenReaderJSNoConsoleLog(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestScreenReaderJSFileSize(t *testing.T) {
	info, err := os.Stat("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to stat screen_reader.js: %v", err)
	}

	maxSize := int64(15 * 1024)
	if info.Size() > maxSize {
		t.Errorf("screen_reader.js is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestScreenReaderJSValidSyntax(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
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

func TestScreenReaderJSModuleExport(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "module.exports") {
		t.Error("JavaScript should export module for testing")
	}
}

func TestScreenReaderJSDOMReady(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "DOMContentLoaded") {
		t.Error("JavaScript should wait for DOMContentLoaded before initializing")
	}
}

func TestScreenReaderJSLandmarks(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	landmarks := []string{"banner", "navigation", "main", "contentinfo"}

	for _, landmark := range landmarks {
		t.Run("landmark_"+landmark, func(t *testing.T) {
			if !strings.Contains(jsContent, landmark) {
				t.Errorf("Should support %s landmark", landmark)
			}
		})
	}
}

func TestScreenReaderJSHeadingHierarchy(t *testing.T) {
	content, err := os.ReadFile("screen_reader.js")
	if err != nil {
		t.Fatalf("Failed to read screen_reader.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "h1") || strings.Contains(jsContent, "h2") || strings.Contains(jsContent, "heading") {
		t.Log("Handles heading hierarchy")
	}
}
