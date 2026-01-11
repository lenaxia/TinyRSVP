package handlers

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

func TestRSVPHandler_ComponentRenderingIntegration(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
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
					"color": "#000000"
				}
			}
		]
	}`

	templateObj := &models.Template{
		ID:              1,
		Name:            "Component Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body>Legacy</body></html>",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:         1,
		Title:      "Component Test Event",
		StartTime:  time.Now().Add(24 * time.Hour),
		Timezone:   "UTC",
		Status:     models.EventStatusPublished,
		CreatedBy:  1,
		TemplateID: func() *int64 { id := int64(1); return &id }(),
	}

	engine := templates.NewEngine()
	templateService := templates.NewService(nil, templates.NewValidator(engine))

	var buf bytes.Buffer
	err := templateService.RenderRSVPPage(&buf, event, templateObj)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Component Test Event") {
		t.Error("Component rendering did not interpolate event title")
	}
	if !strings.Contains(output, "component-canvas") {
		t.Error("Component rendering did not use component template")
	}
	if strings.Contains(output, "Legacy") {
		t.Error("Component rendering used legacy HTML instead of component rendering")
	}
}

func TestRSVPHandler_LegacyRenderingIntegration(t *testing.T) {
	templateObj := &models.Template{
		ID:              1,
		Name:            "Legacy Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body><h1>{{.Event.Title}}</h1></body></html>",
		ComponentConfig: nil,
	}

	event := &models.Event{
		ID:         1,
		Title:      "Legacy Test Event",
		StartTime:  time.Now().Add(24 * time.Hour),
		Timezone:   "UTC",
		Status:     models.EventStatusPublished,
		CreatedBy:  1,
		TemplateID: func() *int64 { id := int64(1); return &id }(),
	}

	engine := templates.NewEngine()
	templateService := templates.NewService(nil, templates.NewValidator(engine))

	var buf bytes.Buffer
	err := templateService.RenderRSVPPage(&buf, event, templateObj)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Legacy Test Event") {
		t.Error("Legacy rendering did not interpolate event title")
	}
	if strings.Contains(output, "component-canvas") {
		t.Error("Legacy rendering used component template instead of legacy HTML")
	}
}

func TestTemplateService_RenderRSVPPage_WithOverrides(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
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
					"color": "#000000",
					"fontSize": "48px"
				}
			}
		]
	}`

	overridesJSON := `{
		"version": "1.0",
		"overrides": [
			{
				"id": "title-text",
				"updates": {
					"content": {"color": "#ff0000"}
				}
			}
		]
	}`

	templateObj := &models.Template{
		ID:              1,
		Name:            "Component Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body>Legacy</body></html>",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:                 1,
		Title:              "Override Test Event",
		StartTime:          time.Now().Add(24 * time.Hour),
		Timezone:           "UTC",
		Status:             models.EventStatusPublished,
		CreatedBy:          1,
		ComponentOverrides: &overridesJSON,
	}

	engine := templates.NewEngine()
	templateService := templates.NewService(nil, templates.NewValidator(engine))

	var buf bytes.Buffer
	err := templateService.RenderRSVPPage(&buf, event, templateObj)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Override Test Event") {
		t.Error("Component rendering did not interpolate event title")
	}
	if !strings.Contains(output, "color: #ff0000") {
		t.Error("Component rendering did not apply color override")
	}
	if !strings.Contains(output, "font-size: 48px") {
		t.Error("Component rendering did not preserve fontSize")
	}
}
