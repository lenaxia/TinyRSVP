package web

import (
	"os"
	"strings"
	"testing"
)

func TestTemplateEditorHTMLExists(t *testing.T) {
	_, err := os.Stat("template_editor.html")
	if err != nil {
		t.Fatalf("template_editor.html should exist: %v", err)
	}
}

func TestTemplateEditorHTMLStructure(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	requiredElements := []string{
		`{{template "base" .}}`,
		`{{define "title"}}`,
		`{{define "content"}}`,
		`{{define "css-extra"}}`,
		`{{define "js-extra"}}`,
	}

	for _, element := range requiredElements {
		if !strings.Contains(html, element) {
			t.Errorf("Expected template to contain %q", element)
		}
	}
}

func TestTemplateEditorHTMLEditorLayout(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	layoutElements := []string{
		`class="template-editor"`,
		`class="editor-toolbar"`,
		`id="component-palette"`,
		`id="template-canvas"`,
		`id="properties-panel"`,
		`class="editor-status-bar"`,
	}

	for _, element := range layoutElements {
		if !strings.Contains(html, element) {
			t.Errorf("Expected layout element %q", element)
		}
	}
}

func TestTemplateEditorHTMLToolbarButtons(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	buttons := []string{
		`id="btn-save"`,
		`id="btn-preview"`,
		`id="btn-undo"`,
		`id="btn-redo"`,
		`id="btn-zoom-in"`,
		`id="btn-zoom-out"`,
		`id="btn-zoom-reset"`,
		`id="btn-mode-desktop"`,
		`id="btn-mode-tablet"`,
		`id="btn-mode-mobile"`,
		`id="btn-toggle-grid"`,
		`id="btn-toggle-snap"`,
	}

	for _, button := range buttons {
		if !strings.Contains(html, button) {
			t.Errorf("Expected toolbar button %q", button)
		}
	}
}

func TestTemplateEditorHTMLAccessibility(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	ariaAttributes := []string{
		`aria-label`,
		`role="region"`,
	}

	for _, attr := range ariaAttributes {
		if !strings.Contains(html, attr) {
			t.Errorf("Expected ARIA attribute %q for accessibility", attr)
		}
	}
}

func TestTemplateEditorHTMLDataAttributes(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `data-template-id`) {
		t.Error("Expected data-template-id attribute")
	}

	if !strings.Contains(html, `{{.Template.ID}}`) {
		t.Error("Expected template ID template variable")
	}
}

func TestTemplateEditorHTMLScriptIncludes(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	scripts := []string{
		`/static/js/visual_canvas.js`,
		`/static/js/component_palette.js`,
		`/static/js/properties_panel.js`,
		`/static/js/template_editor.js`,
	}

	for _, script := range scripts {
		if !strings.Contains(html, script) {
			t.Errorf("Expected script include %q", script)
		}
	}
}

func TestTemplateEditorHTMLStyleIncludes(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `/static/css/template_editor.css`) {
		t.Error("Expected template_editor.css stylesheet include")
	}
}

func TestTemplateEditorHTMLTemplateVariables(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	variables := []string{
		`{{.Template.Name}}`,
		`{{.Template.ID}}`,
	}

	for _, variable := range variables {
		if !strings.Contains(html, variable) {
			t.Errorf("Expected template variable %q", variable)
		}
	}
}

func TestTemplateEditorHTMLStatusDisplay(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `id="editor-status"`) {
		t.Error("Expected editor status element")
	}

	if !strings.Contains(html, `editor-status`) {
		t.Error("Expected editor status class")
	}
}

func TestTemplateEditorHTMLButtonTypes(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `type="button"`) {
		t.Error("Expected button type attributes")
	}
}

func TestTemplateEditorHTMLTooltips(t *testing.T) {
	content, err := os.ReadFile("template_editor.html")
	if err != nil {
		t.Fatalf("Failed to read template_editor.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `title=`) {
		t.Error("Expected title attributes for tooltips")
	}
}
