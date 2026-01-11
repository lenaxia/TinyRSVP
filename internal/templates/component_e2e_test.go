package templates

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEndToEnd_ComponentBasedTemplateFlow(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {
			"name": "Wedding Elegance",
			"category": "card",
			"description": "Elegant wedding invitation"
		},
		"layout": {
			"mode": "card",
			"cardWidth": "800px",
			"backgroundColor": "#ffffff"
		},
		"components": [
			{
				"id": "page-background",
				"type": "Background",
				"position": {"mode": "absolute", "x": "0", "y": "0"},
				"dimensions": {"width": "100%", "height": "100%"},
				"zIndex": 0,
				"visible": true,
				"content": {"type": "color", "color": "#f8f9fa"}
			},
			{
				"id": "header-image",
				"type": "Image",
				"position": {"mode": "absolute", "x": "0", "y": "0"},
				"dimensions": {"width": "100%", "height": "300px"},
				"zIndex": 5,
				"visible": true,
				"content": {
					"src": "/static/images/wedding-header.jpg",
					"alt": "Wedding header",
					"objectFit": "cover"
				}
			},
			{
				"id": "title-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "350px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "{{.Event.Title}}",
					"textAlign": "center",
					"fontFamily": "Playfair Display, serif",
					"fontSize": "48px",
					"fontWeight": "700",
					"color": "#2c3e50"
				}
			},
			{
				"id": "date-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "430px"},
				"dimensions": {"width": "70%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "{{formatDateTime .Event.StartTime}}",
					"textAlign": "center",
					"fontFamily": "Lato, sans-serif",
					"fontSize": "24px",
					"color": "#666666"
				}
			},
			{
				"id": "location-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "480px"},
				"dimensions": {"width": "70%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "{{.Event.Location}}",
					"textAlign": "center",
					"fontFamily": "Lato, sans-serif",
					"fontSize": "20px",
					"color": "#888888"
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
					"position": {"y": "380px"},
					"content": {
						"fontSize": "56px",
						"color": "#8b4789"
					}
				}
			},
			{
				"id": "header-image",
				"updates": {
					"content": {
						"src": "/uploads/events/123/custom-header.jpg"
					}
				}
			}
		],
		"additions": [
			{
				"id": "subtitle-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "450px"},
				"dimensions": {"width": "75%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "Join us for a celebration of love",
					"textAlign": "center",
					"fontFamily": "Lato, sans-serif",
					"fontSize": "24px",
					"fontWeight": "300",
					"color": "#666666",
					"fontStyle": "italic"
				}
			}
		],
		"removals": ["location-text"]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Wedding Elegance",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body>Legacy</body></html>",
		ComponentConfig: &configJSON,
		Category:        models.CategoryCard,
	}

	location := "Grand Ballroom, City Hotel"
	event := &models.Event{
		ID:                 1,
		Title:              "John & Jane's Wedding",
		Location:           &location,
		StartTime:          time.Date(2026, 6, 15, 16, 0, 0, 0, time.UTC),
		Timezone:           "America/Los_Angeles",
		Status:             models.EventStatusPublished,
		CreatedBy:          1,
		ComponentOverrides: &overridesJSON,
	}

	engine := NewEngine()
	templateService := NewService(nil, NewValidator(engine))

	var buf bytes.Buffer
	err := templateService.RenderRSVPPage(&buf, event, template)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}

	if !strings.Contains(output, "John") && !strings.Contains(output, "Jane") {
		t.Error("Output missing event title")
	}

	if !strings.Contains(output, "component-canvas") {
		t.Error("Output missing component-canvas")
	}

	if !strings.Contains(output, "page-background") {
		t.Error("Output missing page-background component")
	}

	if !strings.Contains(output, "header-image") {
		t.Error("Output missing header-image component")
	}

	if !strings.Contains(output, "title-text") {
		t.Error("Output missing title-text component")
	}

	if !strings.Contains(output, "subtitle-text") {
		t.Error("Output missing added subtitle-text component")
	}

	if strings.Contains(output, "location-text") {
		t.Error("Output contains removed location-text component")
	}

	if !strings.Contains(output, "top: 380px") {
		t.Error("Output missing overridden title position")
	}

	if !strings.Contains(output, "font-size: 56px") {
		t.Error("Output missing overridden title font size")
	}

	if !strings.Contains(output, "color: #8b4789") {
		t.Error("Output missing overridden title color")
	}

	if !strings.Contains(output, "/uploads/events/123/custom-header.jpg") {
		t.Error("Output missing overridden header image src")
	}

	if !strings.Contains(output, "Join us for a celebration of love") {
		t.Error("Output missing added subtitle text")
	}

	bgPos := strings.Index(output, "page-background")
	headerPos := strings.Index(output, "header-image")
	titlePos := strings.Index(output, "title-text")

	if bgPos == -1 || headerPos == -1 || titlePos == -1 {
		t.Fatal("Component IDs not found in output")
	}

	if bgPos > headerPos || headerPos > titlePos {
		t.Error("Components not ordered by zIndex (background=0, header=5, title=10)")
	}

	if !strings.Contains(output, "z-index: 0") {
		t.Error("Output missing background z-index")
	}
	if !strings.Contains(output, "z-index: 5") {
		t.Error("Output missing header z-index")
	}
	if !strings.Contains(output, "z-index: 10") {
		t.Error("Output missing title/subtitle z-index")
	}

	if strings.Contains(output, "Legacy") {
		t.Error("Output contains legacy HTML (should use component rendering)")
	}
}

func TestEndToEnd_LegacyTemplateFlow(t *testing.T) {
	template := &models.Template{
		ID:              1,
		Name:            "Legacy Template",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "<html><body><h1>{{.Event.Title}}</h1><p>{{.Event.Location}}</p></body></html>",
		ComponentConfig: nil,
		Category:        models.CategoryPlain,
	}

	location := "Test Location"
	event := &models.Event{
		ID:        1,
		Title:     "Legacy Event",
		Location:  &location,
		StartTime: time.Now().Add(24 * time.Hour),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	engine := NewEngine()
	templateService := NewService(nil, NewValidator(engine))

	var buf bytes.Buffer
	err := templateService.RenderRSVPPage(&buf, event, template)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Legacy Event") {
		t.Error("Output missing event title")
	}

	if !strings.Contains(output, "Test Location") {
		t.Error("Output missing event location")
	}

	if strings.Contains(output, "component-canvas") {
		t.Error("Output uses component rendering (should use legacy HTML)")
	}
}

func TestEndToEnd_ComponentWithTemplateVariables(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [
			{
				"id": "title",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "100px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "Event: {{.Event.Title}}",
					"color": "#000000"
				}
			},
			{
				"id": "date",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "150px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "When: {{formatDateTime .Event.StartTime}}",
					"color": "#666666"
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
		Title:     "Template Variable Test",
		StartTime: time.Date(2026, 6, 15, 16, 0, 0, 0, time.UTC),
		Timezone:  "UTC",
		Status:    models.EventStatusPublished,
		CreatedBy: 1,
	}

	engine := NewEngine()
	templateService := NewService(nil, NewValidator(engine))

	var buf bytes.Buffer
	err := templateService.RenderRSVPPage(&buf, event, template)
	if err != nil {
		t.Fatalf("RenderRSVPPage failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Template Variable Test") {
		t.Error("Event title template variable not interpolated")
	}

	if !strings.Contains(output, "When:") {
		t.Error("Date label not found")
	}

	if strings.Contains(output, "{{.Event.Title}}") {
		t.Error("Template variable not processed")
	}

	if strings.Contains(output, "{{formatDateTime") {
		t.Error("Template function not processed")
	}
}
