package templates

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestService_RenderRSVPPage_ComponentMode(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	svc := &service{
		componentRenderer: renderer,
	}

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

	template := &models.Template{
		ID:              1,
		Name:            "Component Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body>Legacy</body></html>",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Component Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := svc.RenderRSVPPage(&buf, event, template)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Component Event") {
		t.Error("Component rendering did not interpolate event title")
	}
	if !strings.Contains(output, "component-canvas") {
		t.Error("Component rendering did not use component template")
	}
	if strings.Contains(output, "Legacy") {
		t.Error("Component rendering used legacy HTML instead of component rendering")
	}
}

func TestService_RenderRSVPPage_LegacyMode(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	svc := &service{
		componentRenderer: renderer,
	}

	template := &models.Template{
		ID:              1,
		Name:            "Legacy Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body><h1>{{.Event.Title}}</h1></body></html>",
		ComponentConfig: nil,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Legacy Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := svc.RenderRSVPPage(&buf, event, template)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Legacy Event") {
		t.Error("Legacy rendering did not interpolate event title")
	}
	if strings.Contains(output, "component-canvas") {
		t.Error("Legacy rendering used component template instead of legacy HTML")
	}
}

func TestService_RenderRSVPPage_EmptyComponentConfig(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	svc := &service{
		componentRenderer: renderer,
	}

	emptyConfig := ""
	template := &models.Template{
		ID:              1,
		Name:            "Empty Config Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body><h1>{{.Event.Title}}</h1></body></html>",
		ComponentConfig: &emptyConfig,
	}

	event := &models.Event{
		ID:        1,
		Title:     "Empty Config Event",
		StartTime: time.Now(),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	var buf bytes.Buffer
	err := svc.RenderRSVPPage(&buf, event, template)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Empty Config Event") {
		t.Error("Empty config rendering did not interpolate event title")
	}
	if strings.Contains(output, "component-canvas") {
		t.Error("Empty config rendering used component template instead of legacy HTML")
	}
}

func TestService_RenderRSVPPage_NilTemplate(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	svc := &service{
		componentRenderer: renderer,
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
	err := svc.RenderRSVPPage(&buf, event, nil)

	if err == nil {
		t.Error("Expected error for nil template, got nil")
	}
}

func TestService_RenderRSVPPage_NilEvent(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	svc := &service{
		componentRenderer: renderer,
	}

	template := &models.Template{
		ID:          1,
		Name:        "Test Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html><body>Test</body></html>",
	}

	var buf bytes.Buffer
	err := svc.RenderRSVPPage(&buf, nil, template)

	if err == nil {
		t.Error("Expected error for nil event, got nil")
	}
}
