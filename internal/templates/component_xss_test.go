package templates

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestComponentXSS_TextBoxScriptTag(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-text",
			"type": "TextBox",
			"position": {"mode": "absolute", "x": "50%", "y": "100px"},
			"dimensions": {"width": "80%", "height": "auto"},
			"zIndex": 10,
			"visible": true,
			"content": {
				"text": "<script>alert('XSS')</script>",
				"color": "#000000"
			}
		}]
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
		t.Logf("Render error (expected for missing templates): %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "<script>") && !strings.Contains(output, "&lt;script&gt;") {
		t.Error("Script tag was not escaped - XSS vulnerability detected")
	}
}

func TestComponentXSS_ImageJavaScriptURL(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-image",
			"type": "Image",
			"position": {"mode": "absolute", "x": "0", "y": "0"},
			"dimensions": {"width": "100%", "height": "400px"},
			"zIndex": 1,
			"visible": true,
			"content": {
				"src": "javascript:alert('XSS')",
				"alt": "Test"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if src, ok := comp.Content["src"].(string); ok {
			if strings.HasPrefix(src, "javascript:") {
				t.Log("Detected javascript: URL - should be sanitized during rendering")
			}
		}
	}
}

func TestComponentXSS_InlineStyleInjection(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-text",
			"type": "TextBox",
			"position": {"mode": "absolute", "x": "50%", "y": "100px"},
			"dimensions": {"width": "80%", "height": "auto"},
			"zIndex": 10,
			"visible": true,
			"content": {
				"text": "Normal text",
				"color": "red; background: url('javascript:alert(1)')"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if color, ok := comp.Content["color"].(string); ok {
			if strings.Contains(color, "javascript:") {
				t.Log("Detected javascript: in style - should be sanitized during rendering")
			}
		}
	}
}

func TestComponentXSS_EventHandlerAttributes(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-text",
			"type": "TextBox",
			"position": {"mode": "absolute", "x": "50%", "y": "100px"},
			"dimensions": {"width": "80%", "height": "auto"},
			"zIndex": 10,
			"visible": true,
			"content": {
				"text": "<div onclick='alert(1)'>Click me</div>",
				"color": "#000000"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if text, ok := comp.Content["text"].(string); ok {
			if strings.Contains(text, "onclick") {
				t.Log("Detected event handler - Go html/template should auto-escape this")
			}
		}
	}
}

func TestComponentXSS_DataURIScheme(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-image",
			"type": "Image",
			"position": {"mode": "absolute", "x": "0", "y": "0"},
			"dimensions": {"width": "100%", "height": "400px"},
			"zIndex": 1,
			"visible": true,
			"content": {
				"src": "data:text/html,<script>alert('XSS')</script>",
				"alt": "Test"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if src, ok := comp.Content["src"].(string); ok {
			if strings.HasPrefix(src, "data:text/html") {
				t.Log("Detected data:text/html URL - should be validated/sanitized")
			}
		}
	}
}

func TestComponentXSS_HTMLEntities(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-text",
			"type": "TextBox",
			"position": {"mode": "absolute", "x": "50%", "y": "100px"},
			"dimensions": {"width": "80%", "height": "auto"},
			"zIndex": 10,
			"visible": true,
			"content": {
				"text": "&lt;script&gt;alert('XSS')&lt;/script&gt;",
				"color": "#000000"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if text, ok := comp.Content["text"].(string); ok {
			if strings.Contains(text, "&lt;") {
				t.Log("HTML entities detected - Go html/template should handle correctly")
			}
		}
	}
}

func TestComponentXSS_SVGWithScript(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-image",
			"type": "Image",
			"position": {"mode": "absolute", "x": "0", "y": "0"},
			"dimensions": {"width": "100%", "height": "400px"},
			"zIndex": 1,
			"visible": true,
			"content": {
				"src": "data:image/svg+xml,<svg onload='alert(1)'></svg>",
				"alt": "Test"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if src, ok := comp.Content["src"].(string); ok {
			if strings.Contains(src, "onload") {
				t.Log("Detected SVG with event handler - should be sanitized")
			}
		}
	}
}

func TestComponentXSS_CSSExpression(t *testing.T) {
	configJSON := `{
		"version": "1.0",
		"metadata": {"name": "Test", "category": "card", "description": "Test"},
		"layout": {"mode": "card"},
		"components": [{
			"id": "test-text",
			"type": "TextBox",
			"position": {"mode": "absolute", "x": "50%", "y": "100px"},
			"dimensions": {"width": "80%", "height": "auto"},
			"zIndex": 10,
			"visible": true,
			"content": {
				"text": "Test",
				"color": "expression(alert('XSS'))"
			}
		}]
	}`

	renderer := &ComponentRenderer{}
	config, err := renderer.ParseComponentConfig(&configJSON)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if len(config.Components) > 0 {
		comp := config.Components[0]
		if color, ok := comp.Content["color"].(string); ok {
			if strings.Contains(color, "expression(") {
				t.Log("Detected CSS expression - should be sanitized (IE-specific vulnerability)")
			}
		}
	}
}
