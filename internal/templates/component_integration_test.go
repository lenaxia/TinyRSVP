package templates

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestComponentIntegration_RenderWithoutOverrides(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

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
				"id": "title-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "200px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {
					"text": "{{.Event.Title}}",
					"textAlign": "center",
					"fontFamily": "Arial, sans-serif",
					"fontSize": "48px",
					"color": "#2c3e50"
				}
			}
		]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Wedding Elegance",
		Type:            models.TemplateTypeRSVPPage,
		HTMLContent:     "test",
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:        1,
		Title:     "John & Jane's Wedding",
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

	if !strings.Contains(output, "John") && !strings.Contains(output, "Jane") && !strings.Contains(output, "Wedding") {
		t.Errorf("Output missing event title. Output:\n%s", output)
	}
	if !strings.Contains(output, "page-background") {
		t.Error("Output missing page-background component ID")
	}
	if !strings.Contains(output, "title-text") {
		t.Error("Output missing title-text component ID")
	}
	if !strings.Contains(output, "z-index: 0") {
		t.Error("Output missing background z-index")
	}
	if !strings.Contains(output, "z-index: 10") {
		t.Error("Output missing title z-index")
	}

	bgPos := strings.Index(output, "page-background")
	titlePos := strings.Index(output, "title-text")
	if bgPos == -1 || titlePos == -1 {
		t.Fatal("Component IDs not found")
	}
	if bgPos > titlePos {
		t.Error("Components not ordered by zIndex (background should appear before title in DOM)")
	}
}

func TestComponentIntegration_RenderWithOverrides(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

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
					"position": {"y": "250px"},
					"content": {"color": "#ff0000"}
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
		ID:                 1,
		Title:              "Test Event",
		StartTime:          time.Now(),
		Timezone:           "UTC",
		Status:             models.EventStatusPublished,
		CreatedBy:          1,
		ComponentOverrides: &overridesJSON,
	}

	config, err := renderer.ParseComponentConfig(template.ComponentConfig)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	eventOverrides, err := renderer.ParseComponentOverrides(event.ComponentOverrides)
	if err != nil {
		t.Fatalf("Failed to parse overrides: %v", err)
	}

	merged, err := renderer.MergeConfigurations(config, eventOverrides)
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	if len(merged.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(merged.Components))
	}

	comp := merged.Components[0]
	if comp.Position.Y == nil || *comp.Position.Y != "250px" {
		t.Errorf("Position.Y = %v, want 250px", comp.Position.Y)
	}
	if comp.Content["color"] != "#ff0000" {
		t.Errorf("Content color = %v, want #ff0000", comp.Content["color"])
	}
	if comp.Content["fontSize"] != "48px" {
		t.Errorf("Content fontSize = %v, want 48px (should be preserved)", comp.Content["fontSize"])
	}
}

func TestComponentIntegration_AddComponents(t *testing.T) {
	renderer := &ComponentRenderer{}

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
				"content": {"text": "Title"}
			}
		]
	}`

	overridesJSON := `{
		"version": "1.0",
		"additions": [
			{
				"id": "subtitle-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "270px"},
				"dimensions": {"width": "70%", "height": "auto"},
				"zIndex": 11,
				"visible": true,
				"content": {"text": "Subtitle"}
			}
		]
	}`

	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	eventOverrides, err := renderer.ParseComponentOverrides(&overridesJSON)
	if err != nil {
		t.Fatalf("Failed to parse overrides: %v", err)
	}

	merged, err := renderer.MergeConfigurations(config, eventOverrides)
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	if len(merged.Components) != 2 {
		t.Fatalf("Expected 2 components, got %d", len(merged.Components))
	}

	foundSubtitle := false
	for _, comp := range merged.Components {
		if comp.ID == "subtitle-text" {
			foundSubtitle = true
			break
		}
	}
	if !foundSubtitle {
		t.Error("Added subtitle component not found")
	}
}

func TestComponentIntegration_RemoveComponents(t *testing.T) {
	renderer := &ComponentRenderer{}

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
				"content": {"text": "Title"}
			},
			{
				"id": "location-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "270px"},
				"dimensions": {"width": "70%", "height": "auto"},
				"zIndex": 11,
				"visible": true,
				"content": {"text": "Location"}
			}
		]
	}`

	overridesJSON := `{
		"version": "1.0",
		"removals": ["location-text"]
	}`

	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	eventOverrides, err := renderer.ParseComponentOverrides(&overridesJSON)
	if err != nil {
		t.Fatalf("Failed to parse overrides: %v", err)
	}

	merged, err := renderer.MergeConfigurations(config, eventOverrides)
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	if len(merged.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(merged.Components))
	}

	if merged.Components[0].ID != "title-text" {
		t.Errorf("Remaining component ID = %v, want title-text", merged.Components[0].ID)
	}
}

func TestComponentIntegration_UpdateComponentProperties(t *testing.T) {
	renderer := &ComponentRenderer{}

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
					"fontSize": "48px",
					"fontFamily": "Arial, sans-serif"
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
					"position": {"y": "250px"},
					"content": {
						"color": "#8b4789",
						"fontSize": "56px"
					}
				}
			}
		]
	}`

	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	eventOverrides, err := renderer.ParseComponentOverrides(&overridesJSON)
	if err != nil {
		t.Fatalf("Failed to parse overrides: %v", err)
	}

	merged, err := renderer.MergeConfigurations(config, eventOverrides)
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	comp := merged.Components[0]

	if comp.Position.Y == nil || *comp.Position.Y != "250px" {
		t.Errorf("Position.Y = %v, want 250px", comp.Position.Y)
	}
	if comp.Position.X == nil || *comp.Position.X != "50%" {
		t.Errorf("Position.X = %v, want 50%% (should be preserved)", comp.Position.X)
	}
	if comp.Content["color"] != "#8b4789" {
		t.Errorf("Content color = %v, want #8b4789", comp.Content["color"])
	}
	if comp.Content["fontSize"] != "56px" {
		t.Errorf("Content fontSize = %v, want 56px", comp.Content["fontSize"])
	}
	if comp.Content["fontFamily"] != "Arial, sans-serif" {
		t.Errorf("Content fontFamily = %v, want Arial, sans-serif (should be preserved)", comp.Content["fontFamily"])
	}
}

func TestComponentIntegration_ComponentOrdering(t *testing.T) {
	renderer := &ComponentRenderer{}

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [
			{
				"id": "component-3",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "300px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 30,
				"visible": true,
				"content": {"text": "Third"}
			},
			{
				"id": "component-1",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "100px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {"text": "First"}
			},
			{
				"id": "component-2",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "200px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 20,
				"visible": true,
				"content": {"text": "Second"}
			}
		]
	}`

	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	merged, err := renderer.MergeConfigurations(config, nil)
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	if len(merged.Components) != 3 {
		t.Fatalf("Expected 3 components, got %d", len(merged.Components))
	}

	expectedOrder := []string{"component-1", "component-2", "component-3"}
	for i, expectedID := range expectedOrder {
		if merged.Components[i].ID != expectedID {
			t.Errorf("Component[%d].ID = %v, want %v", i, merged.Components[i].ID, expectedID)
		}
	}
}

func TestComponentIntegration_ComplexMerge(t *testing.T) {
	renderer := &ComponentRenderer{}

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [
			{
				"id": "background",
				"type": "Background",
				"position": {"mode": "absolute", "x": "0", "y": "0"},
				"dimensions": {"width": "100%", "height": "100%"},
				"zIndex": 0,
				"visible": true,
				"content": {"type": "color", "color": "#ffffff"}
			},
			{
				"id": "title",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "100px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {"text": "Title", "color": "#000000"}
			},
			{
				"id": "location",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "200px"},
				"dimensions": {"width": "70%", "height": "auto"},
				"zIndex": 11,
				"visible": true,
				"content": {"text": "Location", "color": "#666666"}
			}
		]
	}`

	overridesJSON := `{
		"version": "1.0",
		"overrides": [
			{
				"id": "title",
				"updates": {
					"content": {"color": "#ff0000", "fontSize": "56px"}
				}
			}
		],
		"additions": [
			{
				"id": "subtitle",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "170px"},
				"dimensions": {"width": "75%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {"text": "Subtitle", "color": "#333333"}
			}
		],
		"removals": ["location"]
	}`

	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	eventOverrides, err := renderer.ParseComponentOverrides(&overridesJSON)
	if err != nil {
		t.Fatalf("Failed to parse overrides: %v", err)
	}

	merged, err := renderer.MergeConfigurations(config, eventOverrides)
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	if len(merged.Components) != 3 {
		t.Fatalf("Expected 3 components (background, title, subtitle), got %d", len(merged.Components))
	}

	foundBackground := false
	foundTitle := false
	foundSubtitle := false
	foundLocation := false

	for _, comp := range merged.Components {
		switch comp.ID {
		case "background":
			foundBackground = true
		case "title":
			foundTitle = true
			if comp.Content["color"] != "#ff0000" {
				t.Errorf("Title color = %v, want #ff0000", comp.Content["color"])
			}
			if comp.Content["fontSize"] != "56px" {
				t.Errorf("Title fontSize = %v, want 56px", comp.Content["fontSize"])
			}
		case "subtitle":
			foundSubtitle = true
		case "location":
			foundLocation = true
		}
	}

	if !foundBackground {
		t.Error("Background component not found")
	}
	if !foundTitle {
		t.Error("Title component not found")
	}
	if !foundSubtitle {
		t.Error("Subtitle component not found (should be added)")
	}
	if foundLocation {
		t.Error("Location component found (should be removed)")
	}
}

func TestComponentIntegration_HTMLOutputStructure(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card", "backgroundColor": "#f0f0f0"},
		"components": [
			{
				"id": "test-text",
				"type": "TextBox",
				"position": {"mode": "absolute", "x": "50%", "y": "100px"},
				"dimensions": {"width": "80%", "height": "auto"},
				"zIndex": 10,
				"visible": true,
				"content": {"text": "Test Content"}
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

	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE")
	}
	if !strings.Contains(output, "component-canvas") {
		t.Error("Output missing component-canvas div")
	}
	if !strings.Contains(output, "test-text") {
		t.Error("Output missing test-text component ID")
	}
	if !strings.Contains(output, "Test Content") {
		t.Error("Output missing component content")
	}
	if !strings.Contains(output, "z-index: 10") {
		t.Error("Output missing z-index style")
	}
	if !strings.Contains(output, "background-color: #f0f0f0") {
		t.Error("Output missing background color from layout")
	}
}
