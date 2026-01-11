package templates

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestComponentRenderer_Render_LoadsTemplates(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card", "backgroundColor": "#ffffff"},
		"components": [
			{
				"id": "title-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "200px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "{{.Event.Title}}",
					"textAlign": "center",
					"fontSize": "48px",
					"color": "#2c3e50"
				}
			}
		]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event Title",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}
	if !strings.Contains(output, "Test Event Title") {
		t.Error("Output missing event title")
	}
	if !strings.Contains(output, "component-canvas") {
		t.Error("Output missing component-canvas div")
	}
	if !strings.Contains(output, "title-text") {
		t.Error("Output missing title-text component ID")
	}
}

func TestComponentRenderer_Render_AppliesStyles(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [
			{
				"id": "styled-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "100px", "y": "200px"},
				"dimensions": {"width": "500px", "height": "auto"},
				"zIndex": 15,
				"visible": true,
				"content": {
					"text": "Styled Text",
					"color": "#ff0000",
					"fontSize": "24px"
				}
			}
		]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "z-index: 15") {
		t.Error("Output missing z-index style")
	}
	if !strings.Contains(output, "color: #ff0000") {
		t.Error("Output missing color style")
	}
	if !strings.Contains(output, "font-size: 24px") {
		t.Error("Output missing font-size style")
	}
}

func TestComponentRenderer_Render_OrdersByZIndex(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [
			{
				"id": "high-z",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "100px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 20,
				"visible": true,
				"content": {"text": "High Z"}
			},
			{
				"id": "low-z",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "200px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 5,
				"visible": true,
				"content": {"text": "Low Z"}
			}
		]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	lowZPos := strings.Index(output, "low-z")
	highZPos := strings.Index(output, "high-z")

	if lowZPos == -1 || highZPos == -1 {
		t.Fatal("Component IDs not found in output")
	}

	if lowZPos > highZPos {
		t.Error("Components not ordered by zIndex (low-z should appear before high-z in DOM)")
	}
}

func TestComponentRenderer_Render_InterpolatesTemplateVariables(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [
			{
				"id": "event-title",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "100px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "{{.Event.Title}}",
					"color": "#000000"
				}
			}
		]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "My Special Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "My Special Event") {
		t.Errorf("Template variable {{.Event.Title}} was not interpolated. Output:\n%s", output)
	}
	if strings.Contains(output, "{{.Event.Title}}") {
		t.Errorf("Template variable was not processed. Output:\n%s", output)
	}
}

func TestComponentRenderer_Render_NilTemplate(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, nil)

	if err == nil {
		t.Error("Expected error for nil template, got nil")
	}
	if !strings.Contains(err.Error(), "template cannot be nil") {
		t.Errorf("Expected 'template cannot be nil' error, got: %v", err)
	}
}

func TestComponentRenderer_Render_NilEvent(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": []
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &configJSON,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, nil, template)

	if err == nil {
		t.Error("Expected error for nil event, got nil")
	}
	if !strings.Contains(err.Error(), "event cannot be nil") {
		t.Errorf("Expected 'event cannot be nil' error, got: %v", err)
	}
}

func TestComponentRenderer_Render_NoComponentConfig(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: nil,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)

	if err == nil {
		t.Error("Expected error for missing component config, got nil")
	}
	if !strings.Contains(err.Error(), "no component configuration") {
		t.Errorf("Expected 'no component configuration' error, got: %v", err)
	}
}

func TestComponentRenderer_Render_InvalidJSON(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	invalidJSON := "not valid json"

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &invalidJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}
