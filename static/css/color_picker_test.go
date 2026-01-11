package css

import (
	"os"
	"strings"
	"testing"
)

func TestColorPickerCSS(t *testing.T) {
	cssContent, err := os.ReadFile("color_picker.css")
	if err != nil {
		t.Fatalf("Failed to read color_picker.css: %v", err)
	}

	css := string(cssContent)

	tests := []struct {
		name         string
		wantContains []string
	}{
		{
			name: "has color-picker-section styles",
			wantContains: []string{
				".color-picker-section",
			},
		},
		{
			name: "has color-picker-container styles",
			wantContains: []string{
				".color-picker-container",
			},
		},
		{
			name: "has color-input styles",
			wantContains: []string{
				".color-input",
				"cursor: pointer",
			},
		},
		{
			name: "has color-hex-input styles",
			wantContains: []string{
				".color-hex-input",
				"font-family: monospace",
			},
		},
		{
			name: "has color-preview styles",
			wantContains: []string{
				".color-preview",
				"border-radius",
				"border:",
			},
		},
		{
			name: "has color-controls layout",
			wantContains: []string{
				".color-controls",
				"display:",
				"gap:",
			},
		},
		{
			name: "has reset button styles",
			wantContains: []string{
				".btn-reset-color",
			},
		},
		{
			name: "has responsive design",
			wantContains: []string{
				"@media",
				"max-width:",
			},
		},
		{
			name: "has focus styles for accessibility",
			wantContains: []string{
				":focus",
				"outline:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.wantContains {
				if !strings.Contains(css, want) {
					t.Errorf("Expected CSS to contain %q, but it didn't", want)
				}
			}
		})
	}
}

func TestColorPickerCSSStructure(t *testing.T) {
	cssContent, err := os.ReadFile("color_picker.css")
	if err != nil {
		t.Fatalf("Failed to read color_picker.css: %v", err)
	}

	css := string(cssContent)

	requiredSelectors := []string{
		".color-picker-section",
		".color-picker-title",
		".color-picker-container",
		".color-input-group",
		".color-controls",
		".color-input",
		".color-hex-input",
		".color-preview",
		".color-actions",
		".btn-reset-color",
	}

	for _, selector := range requiredSelectors {
		if !strings.Contains(css, selector) {
			t.Errorf("Expected CSS to contain selector %q", selector)
		}
	}
}

func TestColorPickerCSSAccessibility(t *testing.T) {
	cssContent, err := os.ReadFile("color_picker.css")
	if err != nil {
		t.Fatalf("Failed to read color_picker.css: %v", err)
	}

	css := string(cssContent)

	accessibilityFeatures := []string{
		":focus",
		":hover",
		"outline:",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(css, feature) {
			t.Errorf("Expected CSS to contain accessibility feature %q", feature)
		}
	}
}

func TestColorPickerCSSResponsive(t *testing.T) {
	cssContent, err := os.ReadFile("color_picker.css")
	if err != nil {
		t.Fatalf("Failed to read color_picker.css: %v", err)
	}

	css := string(cssContent)

	if !strings.Contains(css, "@media") {
		t.Error("Expected CSS to contain media queries for responsive design")
	}

	if !strings.Contains(css, "max-width:") && !strings.Contains(css, "min-width:") {
		t.Error("Expected CSS to contain breakpoint definitions")
	}
}
