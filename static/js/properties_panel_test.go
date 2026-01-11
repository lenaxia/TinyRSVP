package js

import (
	"os"
	"strings"
	"testing"
)

func TestPropertiesPanelJavaScript(t *testing.T) {
	jsContent, err := os.ReadFile("properties_panel.js")
	if err != nil {
		t.Fatalf("Failed to read properties_panel.js: %v", err)
	}

	body := string(jsContent)

	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "has PropertiesPanel class definition",
			expected: []string{
				"class PropertiesPanel",
				"constructor",
			},
		},
		{
			name: "has core methods",
			expected: []string{
				"init",
				"showEmptyState",
				"showComponent",
				"createSection",
				"createField",
			},
		},
		{
			name: "has property sections",
			expected: []string{
				"addBasicProperties",
				"addPositionProperties",
				"addDimensionProperties",
				"addStyleProperties",
				"addContentProperties",
			},
		},
		{
			name: "has field types",
			expected: []string{
				"text",
				"select",
				"textarea",
				"checkbox",
				"range",
				"color",
			},
		},
		{
			name: "has component type support",
			expected: []string{
				"TextBox",
				"Image",
				"Background",
				"Container",
			},
		},
		{
			name: "has property change handling",
			expected: []string{
				"handlePropertyChange",
				"parsePropertyPath",
				"property-changed",
			},
		},
		{
			name: "has delete functionality",
			expected: []string{
				"btn-delete-component",
				"component-delete-requested",
			},
		},
		{
			name: "has empty state",
			expected: []string{
				"properties-empty",
				"No Component Selected",
			},
		},
		{
			name: "has accessibility features",
			expected: []string{
				"aria-label",
				"role",
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
