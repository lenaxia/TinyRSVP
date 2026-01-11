package js

import (
	"os"
	"strings"
	"testing"
)

func TestThemePreviewModalColorIntegration(t *testing.T) {
	jsContent, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	js := string(jsContent)

	tests := []struct {
		name         string
		wantContains []string
	}{
		{
			name: "extracts custom color from form",
			wantContains: []string{
				"custom-theme-color-value",
				"custom_color",
			},
		},
		{
			name: "includes custom color in preview URL",
			wantContains: []string{
				"custom_color",
			},
		},
		{
			name: "listens for color change events",
			wantContains: []string{
				"colorChanged",
				"addEventListener",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.wantContains {
				if !strings.Contains(js, want) {
					t.Errorf("Expected theme_preview_modal.js to contain %q for color integration, but it didn't", want)
				}
			}
		})
	}
}

func TestThemePreviewModalColorRefresh(t *testing.T) {
	jsContent, err := os.ReadFile("theme_preview_modal.js")
	if err != nil {
		t.Fatalf("Failed to read theme_preview_modal.js: %v", err)
	}

	js := string(jsContent)

	refreshFeatures := []string{
		"colorChanged",
		"loadPreview",
	}

	for _, feature := range refreshFeatures {
		if !strings.Contains(js, feature) {
			t.Errorf("Expected theme_preview_modal.js to contain refresh feature %q", feature)
		}
	}
}
