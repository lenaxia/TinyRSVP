package js

import (
	"os"
	"strings"
	"testing"
)

func TestComponentPaletteJavaScript(t *testing.T) {
	jsContent, err := os.ReadFile("component_palette.js")
	if err != nil {
		t.Fatalf("Failed to read component_palette.js: %v", err)
	}

	body := string(jsContent)

	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "has ComponentPalette class definition",
			expected: []string{
				"class ComponentPalette",
				"constructor",
			},
		},
		{
			name: "has core methods",
			expected: []string{
				"init",
				"render",
				"createPaletteItem",
				"attachEventListeners",
				"filterComponents",
			},
		},
		{
			name: "has drag and drop support",
			expected: []string{
				"handleDragStart",
				"handleDragEnd",
				"draggable",
				"dragstart",
				"dragend",
				"dataTransfer",
			},
		},
		{
			name: "has component types",
			expected: []string{
				"TextBox",
				"Image",
				"Background",
				"Overlay",
				"Container",
				"Divider",
			},
		},
		{
			name: "has search functionality",
			expected: []string{
				"searchTerm",
				"component-search",
				"filterComponents",
			},
		},
		{
			name: "has keyboard navigation",
			expected: []string{
				"keydown",
				"Enter",
			},
		},
		{
			name: "has accessibility features",
			expected: []string{
				"aria-label",
				"role",
				"tabindex",
			},
		},
		{
			name: "has empty state",
			expected: []string{
				"palette-empty",
				"No components found",
			},
		},
		{
			name: "has event dispatching",
			expected: []string{
				"dispatchEvent",
				"CustomEvent",
			},
		},
		{
			name: "has module export",
			expected: []string{
				"module.exports",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(body, exp) {
					t.Errorf("Expected content to contain %q", exp)
				}
			}
		})
	}
}
