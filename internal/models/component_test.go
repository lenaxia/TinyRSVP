package models

import (
	"encoding/json"
	"testing"
)

func TestComponentType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		compType ComponentType
		want     bool
	}{
		{"TextBox is valid", ComponentTypeTextBox, true},
		{"Image is valid", ComponentTypeImage, true},
		{"Background is valid", ComponentTypeBackground, true},
		{"Overlay is valid", ComponentTypeOverlay, true},
		{"Container is valid", ComponentTypeContainer, true},
		{"Divider is valid", ComponentTypeDivider, true},
		{"Invalid type", ComponentType("invalid"), false},
		{"Empty type", ComponentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.compType.IsValid(); got != tt.want {
				t.Errorf("ComponentType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionMode_IsValid(t *testing.T) {
	tests := []struct {
		name string
		mode PositionMode
		want bool
	}{
		{"Absolute is valid", PositionModeAbsolute, true},
		{"Relative is valid", PositionModeRelative, true},
		{"Flex is valid", PositionModeFlex, true},
		{"Invalid mode", PositionMode("invalid"), false},
		{"Empty mode", PositionMode(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.want {
				t.Errorf("PositionMode.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pos     Position
		wantErr bool
	}{
		{
			name: "Valid absolute position",
			pos: Position{
				Mode: PositionModeAbsolute,
				X:    strPtr("50%"),
				Y:    strPtr("100px"),
			},
			wantErr: false,
		},
		{
			name: "Valid relative position with order",
			pos: Position{
				Mode:  PositionModeRelative,
				Order: intPtr(1),
			},
			wantErr: false,
		},
		{
			name: "Valid flex position with order",
			pos: Position{
				Mode:  PositionModeFlex,
				Order: intPtr(2),
			},
			wantErr: false,
		},
		{
			name: "Invalid mode",
			pos: Position{
				Mode: PositionMode("invalid"),
			},
			wantErr: true,
		},
		{
			name: "Absolute without X",
			pos: Position{
				Mode: PositionModeAbsolute,
				Y:    strPtr("100px"),
			},
			wantErr: true,
		},
		{
			name: "Absolute without Y",
			pos: Position{
				Mode: PositionModeAbsolute,
				X:    strPtr("50%"),
			},
			wantErr: true,
		},
		{
			name: "Flex without order",
			pos: Position{
				Mode: PositionModeFlex,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pos.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Position.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDimensions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dims    Dimensions
		wantErr bool
	}{
		{
			name: "Valid dimensions",
			dims: Dimensions{
				Width:  "100%",
				Height: "auto",
			},
			wantErr: false,
		},
		{
			name: "Valid pixel dimensions",
			dims: Dimensions{
				Width:  "800px",
				Height: "600px",
			},
			wantErr: false,
		},
		{
			name: "Empty width",
			dims: Dimensions{
				Width:  "",
				Height: "100px",
			},
			wantErr: true,
		},
		{
			name: "Empty height",
			dims: Dimensions{
				Width:  "100px",
				Height: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dims.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Dimensions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTextBoxContent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		content TextBoxContent
		wantErr bool
	}{
		{
			name: "Valid text box content",
			content: TextBoxContent{
				Text:       "{{.Event.Title}}",
				TextAlign:  "center",
				FontFamily: "Arial, sans-serif",
				FontSize:   "16px",
				Color:      "#000000",
			},
			wantErr: false,
		},
		{
			name: "Empty text",
			content: TextBoxContent{
				Text: "",
			},
			wantErr: true,
		},
		{
			name: "Text too long",
			content: TextBoxContent{
				Text: string(make([]byte, 10001)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.content.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TextBoxContent.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageContent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		content ImageContent
		wantErr bool
	}{
		{
			name: "Valid image content",
			content: ImageContent{
				Src:       "/static/images/header.jpg",
				Alt:       "Header image",
				ObjectFit: "cover",
			},
			wantErr: false,
		},
		{
			name: "Empty src",
			content: ImageContent{
				Src: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid opacity",
			content: ImageContent{
				Src:     "/static/images/header.jpg",
				Opacity: floatPtr(1.5),
			},
			wantErr: true,
		},
		{
			name: "Negative opacity",
			content: ImageContent{
				Src:     "/static/images/header.jpg",
				Opacity: floatPtr(-0.1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.content.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageContent.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComponent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		comp    Component
		wantErr bool
	}{
		{
			name: "Valid TextBox component",
			comp: Component{
				ID:   "title-text",
				Type: ComponentTypeTextBox,
				Position: Position{
					Mode: PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("100px"),
				},
				Dimensions: Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":      "Test Title",
					"textAlign": "center",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty ID",
			comp: Component{
				ID:   "",
				Type: ComponentTypeTextBox,
			},
			wantErr: true,
		},
		{
			name: "Invalid type",
			comp: Component{
				ID:   "test",
				Type: ComponentType("invalid"),
			},
			wantErr: true,
		},
		{
			name: "Negative zIndex",
			comp: Component{
				ID:     "test",
				Type:   ComponentTypeTextBox,
				ZIndex: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Component.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComponent_JSONSerialization(t *testing.T) {
	comp := Component{
		ID:   "test-component",
		Type: ComponentTypeTextBox,
		Position: Position{
			Mode: PositionModeAbsolute,
			X:    strPtr("50%"),
			Y:    strPtr("100px"),
		},
		Dimensions: Dimensions{
			Width:  "80%",
			Height: "auto",
		},
		ZIndex:  10,
		Visible: true,
		Content: map[string]interface{}{
			"text":      "Test",
			"textAlign": "center",
		},
	}

	jsonData, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("Failed to marshal component: %v", err)
	}

	var decoded Component
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal component: %v", err)
	}

	if decoded.ID != comp.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, comp.ID)
	}
	if decoded.Type != comp.Type {
		t.Errorf("Type mismatch: got %v, want %v", decoded.Type, comp.Type)
	}
	if decoded.ZIndex != comp.ZIndex {
		t.Errorf("ZIndex mismatch: got %v, want %v", decoded.ZIndex, comp.ZIndex)
	}
}

func TestComponentConfiguration_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ComponentConfiguration
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: ComponentConfiguration{
				Version: "1.0",
				Metadata: ConfigMetadata{
					Name:        "Test Template",
					Category:    "card",
					Description: "Test description",
				},
				Layout: PageLayoutConfig{
					Mode:      "card",
					CardWidth: "800px",
				},
				Components: []Component{
					{
						ID:   "test-1",
						Type: ComponentTypeTextBox,
						Position: Position{
							Mode: PositionModeAbsolute,
							X:    strPtr("0"),
							Y:    strPtr("0"),
						},
						Dimensions: Dimensions{
							Width:  "100%",
							Height: "auto",
						},
						Content: map[string]interface{}{
							"text": "Test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty version",
			config: ComponentConfiguration{
				Version: "",
			},
			wantErr: true,
		},
		{
			name: "Duplicate component IDs",
			config: ComponentConfiguration{
				Version: "1.0",
				Metadata: ConfigMetadata{
					Name: "Test",
				},
				Components: []Component{
					{
						ID:   "duplicate",
						Type: ComponentTypeTextBox,
						Position: Position{
							Mode: PositionModeAbsolute,
							X:    strPtr("0"),
							Y:    strPtr("0"),
						},
						Dimensions: Dimensions{
							Width:  "100%",
							Height: "auto",
						},
					},
					{
						ID:   "duplicate",
						Type: ComponentTypeImage,
						Position: Position{
							Mode: PositionModeAbsolute,
							X:    strPtr("0"),
							Y:    strPtr("0"),
						},
						Dimensions: Dimensions{
							Width:  "100%",
							Height: "auto",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ComponentConfiguration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComponentOverrides_Validate(t *testing.T) {
	tests := []struct {
		name      string
		overrides ComponentOverrides
		wantErr   bool
	}{
		{
			name: "Valid overrides",
			overrides: ComponentOverrides{
				Version: "1.0",
				Overrides: []ComponentOverride{
					{
						ID: "title-text",
						Updates: map[string]interface{}{
							"position": map[string]interface{}{
								"y": "200px",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty version",
			overrides: ComponentOverrides{
				Version: "",
			},
			wantErr: true,
		},
		{
			name: "Duplicate override IDs",
			overrides: ComponentOverrides{
				Version: "1.0",
				Overrides: []ComponentOverride{
					{
						ID: "duplicate",
						Updates: map[string]interface{}{
							"zIndex": 20,
						},
					},
					{
						ID: "duplicate",
						Updates: map[string]interface{}{
							"visible": false,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.overrides.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ComponentOverrides.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
