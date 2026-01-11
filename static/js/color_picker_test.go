package js

import (
	"os"
	"strings"
	"testing"
)

func TestColorPickerJavaScript(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	tests := []struct {
		name         string
		wantContains []string
	}{
		{
			name: "has ColorPicker class definition",
			wantContains: []string{
				"class ColorPicker",
				"constructor",
			},
		},
		{
			name: "has initialization method",
			wantContains: []string{
				"init",
				"addEventListener",
			},
		},
		{
			name: "has color input sync functionality",
			wantContains: []string{
				"updateAllInputs",
			},
		},
		{
			name: "has hex validation",
			wantContains: []string{
				"isValidHex",
				"#",
				"test",
			},
		},
		{
			name: "has color preview update",
			wantContains: []string{
				"updatePreview",
				"backgroundColor",
			},
		},
		{
			name: "has reset functionality",
			wantContains: []string{
				"reset",
			},
		},
		{
			name: "has event listeners for color input",
			wantContains: []string{
				"custom-theme-color",
				"input",
				"change",
			},
		},
		{
			name: "has event listeners for hex input",
			wantContains: []string{
				"custom-theme-color-hex",
			},
		},
		{
			name: "has hidden input update",
			wantContains: []string{
				"custom-theme-color-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.wantContains {
				if !strings.Contains(js, want) {
					t.Errorf("Expected JavaScript to contain %q, but it didn't", want)
				}
			}
		})
	}
}

func TestColorPickerValidation(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	validationFeatures := []string{
		"isValidHex",
		"^#[0-9A-Fa-f]{6}$",
		"test(",
	}

	for _, feature := range validationFeatures {
		if !strings.Contains(js, feature) {
			t.Errorf("Expected JavaScript to contain validation feature %q", feature)
		}
	}
}

func TestColorPickerAccessibility(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	accessibilityFeatures := []string{
		"aria-",
		"setAttribute",
	}

	foundAny := false
	for _, feature := range accessibilityFeatures {
		if strings.Contains(js, feature) {
			foundAny = true
			break
		}
	}

	if !foundAny {
		t.Error("Expected JavaScript to contain accessibility features (aria attributes)")
	}
}

func TestColorPickerErrorHandling(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	errorHandlingFeatures := []string{
		"classList.add",
		"classList.remove",
		"error",
	}

	for _, feature := range errorHandlingFeatures {
		if !strings.Contains(js, feature) {
			t.Errorf("Expected JavaScript to contain error handling feature %q", feature)
		}
	}
}

func TestColorPickerExport(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	if !strings.Contains(js, "window.ColorPicker") && !strings.Contains(js, "export") {
		t.Error("Expected JavaScript to export ColorPicker class")
	}
}

func TestColorPickerDOMReady(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	domReadyPatterns := []string{
		"DOMContentLoaded",
		"addEventListener",
	}

	foundDOMReady := false
	for _, pattern := range domReadyPatterns {
		if strings.Contains(js, pattern) {
			foundDOMReady = true
			break
		}
	}

	if !foundDOMReady {
		t.Error("Expected JavaScript to wait for DOM ready before initializing")
	}
}

func TestColorPickerPreviewIntegration(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	previewFeatures := []string{
		"updatePreview",
		"color-preview",
	}

	for _, feature := range previewFeatures {
		if !strings.Contains(js, feature) {
			t.Errorf("Expected JavaScript to contain preview integration feature %q", feature)
		}
	}
}

func TestColorPickerResetButton(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	resetFeatures := []string{
		"reset-color-btn",
		"click",
	}

	for _, feature := range resetFeatures {
		if !strings.Contains(js, feature) {
			t.Errorf("Expected JavaScript to contain reset button feature %q", feature)
		}
	}
}

func TestColorPickerKeyboardSupport(t *testing.T) {
	jsContent, err := os.ReadFile("color_picker.js")
	if err != nil {
		t.Fatalf("Failed to read color_picker.js: %v", err)
	}

	js := string(jsContent)

	keyboardFeatures := []string{
		"keydown",
		"keyup",
		"Enter",
	}

	foundKeyboard := false
	for _, feature := range keyboardFeatures {
		if strings.Contains(js, feature) {
			foundKeyboard = true
			break
		}
	}

	if !foundKeyboard {
		t.Error("Expected JavaScript to contain keyboard support features")
	}
}
