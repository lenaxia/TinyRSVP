package js

import (
	"os"
	"strings"
	"testing"
)

func TestVisualCanvasJavaScript(t *testing.T) {
	jsContent, err := os.ReadFile("visual_canvas.js")
	if err != nil {
		t.Fatalf("Failed to read visual_canvas.js: %v", err)
	}

	body := string(jsContent)

	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "has VisualCanvas class definition",
			expected: []string{
				"class VisualCanvas",
				"constructor",
			},
		},
		{
			name: "has core methods",
			expected: []string{
				"init",
				"attachEventListeners",
				"render",
				"addComponent",
				"removeComponent",
				"selectComponent",
				"deselectComponent",
			},
		},
		{
			name: "has drag and drop support",
			expected: []string{
				"handleMouseDown",
				"handleMouseMove",
				"handleMouseUp",
				"startDrag",
				"updateDrag",
				"endDrag",
				"dragState",
			},
		},
		{
			name: "has resize support",
			expected: []string{
				"startResize",
				"updateResize",
				"endResize",
				"resizeState",
				"resize-handle",
			},
		},
		{
			name: "has component rendering",
			expected: []string{
				"renderComponent",
				"renderTextBox",
				"renderImage",
				"renderBackground",
				"renderOverlay",
				"renderContainer",
				"renderDivider",
			},
		},
		{
			name: "has undo/redo support",
			expected: []string{
				"history",
				"historyIndex",
				"saveState",
				"undo",
				"redo",
			},
		},
		{
			name: "has keyboard shortcuts",
			expected: []string{
				"ArrowUp",
				"ArrowDown",
				"ArrowLeft",
				"ArrowRight",
				"Delete",
			},
		},
		{
			name: "has zoom controls",
			expected: []string{
				"setZoom",
				"zoom",
			},
		},
		{
			name: "has responsive modes",
			expected: []string{
				"setMode",
				"desktop",
				"tablet",
				"mobile",
			},
		},
		{
			name: "has grid support",
			expected: []string{
				"toggleGrid",
				"snapToGrid",
				"gridSize",
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
			name: "has accessibility features",
			expected: []string{
				"aria-label",
				"aria-selected",
				"role",
				"tabindex",
			},
		},
		{
			name: "has event dispatching",
			expected: []string{
				"dispatchEvent",
				"CustomEvent",
				"component-selected",
				"component-moved",
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
