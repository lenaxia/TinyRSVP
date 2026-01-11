package css

import (
	"os"
	"strings"
	"testing"
)

func TestTemplateEditorCSS(t *testing.T) {
	content, err := os.ReadFile("template_editor.css")
	if err != nil {
		t.Fatalf("Failed to read template_editor.css: %v", err)
	}

	cssContent := string(content)

	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "has main editor layout",
			expected: []string{
				".template-editor",
				"display: grid",
				"grid-template-columns",
				"grid-template-areas",
			},
		},
		{
			name: "has toolbar styles",
			expected: []string{
				".editor-toolbar",
				".editor-toolbar-section",
				".editor-toolbar-divider",
			},
		},
		{
			name: "has palette styles",
			expected: []string{
				"#component-palette",
				".palette-header",
				".palette-title",
				".palette-search",
				".palette-list",
				".palette-item",
			},
		},
		{
			name: "has canvas styles",
			expected: []string{
				"#template-canvas",
				".visual-canvas",
				".canvas-component",
				".canvas-component.selected",
			},
		},
		{
			name: "has resize handles",
			expected: []string{
				".resize-handle",
				".resize-handle-nw",
				".resize-handle-n",
				".resize-handle-ne",
				".resize-handle-e",
				".resize-handle-se",
				".resize-handle-s",
				".resize-handle-sw",
				".resize-handle-w",
			},
		},
		{
			name: "has properties panel styles",
			expected: []string{
				"#properties-panel",
				".properties-empty",
				".properties-header",
				".properties-form",
				".properties-section",
				".properties-field",
			},
		},
		{
			name: "has form input styles",
			expected: []string{
				".properties-input",
				".properties-select",
				".properties-textarea",
				".properties-checkbox",
				".properties-range",
			},
		},
		{
			name: "has status bar styles",
			expected: []string{
				".editor-status-bar",
				".editor-status",
				".editor-status-info",
				".editor-status-success",
				".editor-status-error",
			},
		},
		{
			name: "has error display styles",
			expected: []string{
				".editor-error",
				".editor-error-content",
				".editor-error-message",
				"@keyframes slideIn",
			},
		},
		{
			name: "has responsive modes",
			expected: []string{
				".visual-canvas.mode-desktop",
				".visual-canvas.mode-tablet",
				".visual-canvas.mode-mobile",
			},
		},
		{
			name: "has accessibility styles",
			expected: []string{
				".sr-only",
				":focus",
			},
		},
		{
			name: "uses CSS variables",
			expected: []string{
				"var(--color-",
				"var(--spacing-",
				"var(--font-size-",
				"var(--radius-",
				"var(--shadow-",
			},
		},
		{
			name: "has responsive design",
			expected: []string{
				"@media",
				"max-width",
			},
		},
		{
			name: "has transitions",
			expected: []string{
				"transition:",
				"var(--transition-",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, exp := range tt.expected {
				if !strings.Contains(cssContent, exp) {
					t.Errorf("Expected CSS to contain %q", exp)
				}
			}
		})
	}
}
