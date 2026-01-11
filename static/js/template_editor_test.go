package js

import (
	"os"
	"strings"
	"testing"
)

func TestTemplateEditorJavaScript(t *testing.T) {
	jsContent, err := os.ReadFile("template_editor.js")
	if err != nil {
		t.Fatalf("Failed to read template_editor.js: %v", err)
	}

	body := string(jsContent)

	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "has TemplateEditor class definition",
			expected: []string{
				"class TemplateEditor",
				"constructor",
			},
		},
		{
			name: "has initialization methods",
			expected: []string{
				"init",
				"initializeElements",
				"initializeModules",
				"attachEventListeners",
				"attachGlobalListeners",
			},
		},
		{
			name: "has module integration",
			expected: []string{
				"VisualCanvas",
				"ComponentPalette",
				"PropertiesPanel",
			},
		},
		{
			name: "has API integration",
			expected: []string{
				"loadTemplate",
				"save",
				"preview",
				"fetch",
				"/components",
			},
		},
		{
			name: "has toolbar actions",
			expected: []string{
				"btn-save",
				"btn-preview",
				"btn-undo",
				"btn-redo",
				"btn-zoom-in",
				"btn-zoom-out",
				"btn-zoom-reset",
			},
		},
		{
			name: "has view mode controls",
			expected: []string{
				"btn-mode-desktop",
				"btn-mode-tablet",
				"btn-mode-mobile",
				"setMode",
			},
		},
		{
			name: "has grid controls",
			expected: []string{
				"btn-toggle-grid",
				"btn-toggle-snap",
				"toggleGrid",
				"toggleSnap",
			},
		},
		{
			name: "has zoom controls",
			expected: []string{
				"zoomIn",
				"zoomOut",
				"zoomReset",
			},
		},
		{
			name: "has undo/redo",
			expected: []string{
				"undo",
				"redo",
			},
		},
		{
			name: "has event handling",
			expected: []string{
				"component-selected",
				"component-deselected",
				"component-added",
				"component-removed",
				"component-moved",
				"property-changed",
			},
		},
		{
			name: "has dirty state tracking",
			expected: []string{
				"isDirty",
				"markDirty",
				"beforeunload",
				"unsaved changes",
			},
		},
		{
			name: "has status updates",
			expected: []string{
				"updateStatus",
				"editor-status",
			},
		},
		{
			name: "has error handling",
			expected: []string{
				"showError",
				"catch",
				"editor-error",
			},
		},
		{
			name: "has keyboard shortcuts",
			expected: []string{
				"ctrlKey",
				"metaKey",
			},
		},
		{
			name: "has initialization",
			expected: []string{
				"DOMContentLoaded",
				"document.readyState",
				"templateId",
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
