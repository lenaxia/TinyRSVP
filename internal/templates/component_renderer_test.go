package templates

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestParseComponentConfig(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr *string
		wantErr bool
		check   func(*testing.T, *models.ComponentConfiguration)
	}{
		{
			name:    "nil input",
			jsonStr: nil,
			wantErr: false,
			check: func(t *testing.T, config *models.ComponentConfiguration) {
				if config != nil {
					t.Error("Expected nil config for nil input")
				}
			},
		},
		{
			name:    "empty string",
			jsonStr: strPtr(""),
			wantErr: false,
			check: func(t *testing.T, config *models.ComponentConfiguration) {
				if config != nil {
					t.Error("Expected nil config for empty string")
				}
			},
		},
		{
			name:    "invalid JSON",
			jsonStr: strPtr("{invalid json}"),
			wantErr: true,
		},
		{
			name: "valid configuration",
			jsonStr: strPtr(`{
				"version": "1.0",
				"metadata": {
					"name": "Test Template",
					"category": "card",
					"description": "Test description"
				},
				"layout": {
					"mode": "card",
					"cardWidth": "800px"
				},
				"components": [
					{
						"id": "test-component",
						"type": "TextBox",
						"position": {"mode": "absolute", "x": "50%", "y": "100px"},
						"dimensions": {"width": "80%", "height": "auto"},
						"zIndex": 10,
						"visible": true,
						"content": {"text": "Test"}
					}
				]
			}`),
			wantErr: false,
			check: func(t *testing.T, config *models.ComponentConfiguration) {
				if config == nil {
					t.Fatal("Expected non-nil config")
				}
				if config.Version != "1.0" {
					t.Errorf("Version = %v, want 1.0", config.Version)
				}
				if config.Metadata.Name != "Test Template" {
					t.Errorf("Metadata.Name = %v, want Test Template", config.Metadata.Name)
				}
				if len(config.Components) != 1 {
					t.Errorf("len(Components) = %v, want 1", len(config.Components))
				}
			},
		},
	}

	renderer := &ComponentRenderer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := renderer.ParseComponentConfig(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseComponentConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, config)
			}
		})
	}
}

func TestParseComponentOverrides(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr *string
		wantErr bool
		check   func(*testing.T, *models.ComponentOverrides)
	}{
		{
			name:    "nil input",
			jsonStr: nil,
			wantErr: false,
			check: func(t *testing.T, overrides *models.ComponentOverrides) {
				if overrides != nil {
					t.Error("Expected nil overrides for nil input")
				}
			},
		},
		{
			name:    "empty string",
			jsonStr: strPtr(""),
			wantErr: false,
			check: func(t *testing.T, overrides *models.ComponentOverrides) {
				if overrides != nil {
					t.Error("Expected nil overrides for empty string")
				}
			},
		},
		{
			name:    "invalid JSON",
			jsonStr: strPtr("{invalid json}"),
			wantErr: true,
		},
		{
			name: "valid overrides",
			jsonStr: strPtr(`{
				"version": "1.0",
				"overrides": [
					{
						"id": "title-text",
						"updates": {
							"position": {"y": "250px"},
							"content": {"color": "#ff0000"}
						}
					}
				],
				"additions": [
					{
						"id": "new-component",
						"type": "TextBox",
						"position": {"mode": "absolute", "x": "50%", "y": "300px"},
						"dimensions": {"width": "70%", "height": "auto"},
						"zIndex": 15,
						"visible": true,
						"content": {"text": "New text"}
					}
				],
				"removals": ["old-component"]
			}`),
			wantErr: false,
			check: func(t *testing.T, overrides *models.ComponentOverrides) {
				if overrides == nil {
					t.Fatal("Expected non-nil overrides")
				}
				if overrides.Version != "1.0" {
					t.Errorf("Version = %v, want 1.0", overrides.Version)
				}
				if len(overrides.Overrides) != 1 {
					t.Errorf("len(Overrides) = %v, want 1", len(overrides.Overrides))
				}
				if len(overrides.Additions) != 1 {
					t.Errorf("len(Additions) = %v, want 1", len(overrides.Additions))
				}
				if len(overrides.Removals) != 1 {
					t.Errorf("len(Removals) = %v, want 1", len(overrides.Removals))
				}
			},
		},
	}

	renderer := &ComponentRenderer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides, err := renderer.ParseComponentOverrides(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseComponentOverrides() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, overrides)
			}
		})
	}
}

func TestMergeConfigurations_NoOverrides(t *testing.T) {
	renderer := &ComponentRenderer{}

	base := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "component-1",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("100px"),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:  "Original",
						Color: "#000000",
					},
				},
			},
		},
	}

	result, err := renderer.MergeConfigurations(base, nil)
	if err != nil {
		t.Fatalf("MergeConfigurations() error = %v", err)
	}

	if len(result.Components) != 1 {
		t.Errorf("len(Components) = %v, want 1", len(result.Components))
	}

	comp := result.Components[0]
	if comp.ID != "component-1" {
		t.Errorf("Component ID = %v, want component-1", comp.ID)
	}
}

func TestMergeConfigurations_SimpleOverride(t *testing.T) {
	renderer := &ComponentRenderer{}

	base := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "title-text",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("100px"),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:     "Original Title",
						Color:    "#000000",
						FontSize: "48px",
					},
				},
			},
		},
	}

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#ff0000",
					},
				},
			},
		},
	}

	result, err := renderer.MergeConfigurations(base, overrides)
	if err != nil {
		t.Fatalf("MergeConfigurations() error = %v", err)
	}

	if len(result.Components) != 1 {
		t.Fatalf("len(Components) = %v, want 1", len(result.Components))
	}

	comp := result.Components[0]
	if comp.Content == nil || comp.Content.TextBox == nil {
		t.Fatal("Component Content or TextBox is nil")
	}
	if comp.Content.TextBox.Color != "#ff0000" {
		t.Errorf("Content color = %v, want #ff0000", comp.Content.TextBox.Color)
	}
	if comp.Content.TextBox.Text != "Original Title" {
		t.Errorf("Content text = %v, want Original Title", comp.Content.TextBox.Text)
	}
	if comp.Content.TextBox.FontSize != "48px" {
		t.Errorf("Content fontSize = %v, want 48px", comp.Content.TextBox.FontSize)
	}
}

func TestMergeConfigurations_NestedOverride(t *testing.T) {
	renderer := &ComponentRenderer{}

	base := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "title-text",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("100px"),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:  "Original",
						Color: "#000000",
					},
				},
			},
		},
	}

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"position": map[string]interface{}{
						"y": "250px",
					},
				},
			},
		},
	}

	result, err := renderer.MergeConfigurations(base, overrides)
	if err != nil {
		t.Fatalf("MergeConfigurations() error = %v", err)
	}

	comp := result.Components[0]
	if comp.Position.Y == nil || *comp.Position.Y != "250px" {
		t.Errorf("Position.Y = %v, want 250px", comp.Position.Y)
	}
	if comp.Position.X == nil || *comp.Position.X != "50%" {
		t.Errorf("Position.X = %v, want 50%%", comp.Position.X)
	}
}

func TestMergeConfigurations_Additions(t *testing.T) {
	renderer := &ComponentRenderer{}

	base := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "component-1",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("100px"),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
		},
	}

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Additions: []models.Component{
			{
				ID:      "new-component",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  15,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("300px"),
				},
				Dimensions: models.Dimensions{
					Width:  "70%",
					Height: "auto",
				},
			},
		},
	}

	result, err := renderer.MergeConfigurations(base, overrides)
	if err != nil {
		t.Fatalf("MergeConfigurations() error = %v", err)
	}

	if len(result.Components) != 2 {
		t.Fatalf("len(Components) = %v, want 2", len(result.Components))
	}

	found := false
	for _, comp := range result.Components {
		if comp.ID == "new-component" {
			found = true
			break
		}
	}
	if !found {
		t.Error("New component not found in result")
	}
}

func TestMergeConfigurations_Removals(t *testing.T) {
	renderer := &ComponentRenderer{}

	base := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "component-1",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
			{
				ID:      "component-2",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  20,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
		},
	}

	overrides := &models.ComponentOverrides{
		Version:  "1.0",
		Removals: []string{"component-1"},
	}

	result, err := renderer.MergeConfigurations(base, overrides)
	if err != nil {
		t.Fatalf("MergeConfigurations() error = %v", err)
	}

	if len(result.Components) != 1 {
		t.Fatalf("len(Components) = %v, want 1", len(result.Components))
	}

	if result.Components[0].ID != "component-2" {
		t.Errorf("Remaining component ID = %v, want component-2", result.Components[0].ID)
	}
}

func TestMergeConfigurations_ZIndexSorting(t *testing.T) {
	renderer := &ComponentRenderer{}

	base := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "component-3",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  30,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
			{
				ID:      "component-1",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
			{
				ID:      "component-2",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  20,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
		},
	}

	result, err := renderer.MergeConfigurations(base, nil)
	if err != nil {
		t.Fatalf("MergeConfigurations() error = %v", err)
	}

	if len(result.Components) != 3 {
		t.Fatalf("len(Components) = %v, want 3", len(result.Components))
	}

	if result.Components[0].ID != "component-1" {
		t.Errorf("First component ID = %v, want component-1", result.Components[0].ID)
	}
	if result.Components[1].ID != "component-2" {
		t.Errorf("Second component ID = %v, want component-2", result.Components[1].ID)
	}
	if result.Components[2].ID != "component-3" {
		t.Errorf("Third component ID = %v, want component-3", result.Components[2].ID)
	}
}

func TestMergeConfigurations_NilBase(t *testing.T) {
	renderer := &ComponentRenderer{}

	overrides := &models.ComponentOverrides{
		Version: "1.0",
	}

	_, err := renderer.MergeConfigurations(nil, overrides)
	if err == nil {
		t.Error("Expected error for nil base configuration")
	}
}

func TestRender_NilTemplate(t *testing.T) {
	renderer := &ComponentRenderer{}
	var buf bytes.Buffer

	event := &models.Event{
		Title: "Test Event",
	}

	err := renderer.Render(&buf, event, nil)
	if err == nil {
		t.Error("Expected error for nil template")
	}
}

func mustMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}
