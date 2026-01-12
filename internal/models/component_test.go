package models

import (
	"encoding/json"
	"testing"
)

func TestComponentType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		ct   ComponentType
		want bool
	}{
		{"TextBox valid", ComponentTypeTextBox, true},
		{"Image valid", ComponentTypeImage, true},
		{"Background valid", ComponentTypeBackground, true},
		{"Overlay valid", ComponentTypeOverlay, true},
		{"Container valid", ComponentTypeContainer, true},
		{"Divider valid", ComponentTypeDivider, true},
		{"Invalid type", ComponentType("invalid"), false},
		{"Empty type", ComponentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsValid(); got != tt.want {
				t.Errorf("ComponentType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionMode_IsValid(t *testing.T) {
	tests := []struct {
		name string
		pm   PositionMode
		want bool
	}{
		{"Absolute valid", PositionModeAbsolute, true},
		{"Relative valid", PositionModeRelative, true},
		{"Flex valid", PositionModeFlex, true},
		{"Invalid mode", PositionMode("invalid"), false},
		{"Empty mode", PositionMode(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pm.IsValid(); got != tt.want {
				t.Errorf("PositionMode.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponent_JSONSerialization(t *testing.T) {
	tests := []struct {
		name      string
		component Component
		wantErr   bool
	}{
		{
			name: "TextBox component",
			component: Component{
				ID:      "title-text",
				Type:    ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: Position{
					Mode: PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("200px"),
				},
				Dimensions: Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				Content: &ComponentContent{
					TextBox: &TextBoxContent{
						Text:       "{{.Event.Title}}",
						TextAlign:  "center",
						FontFamily: "Arial, sans-serif",
						FontSize:   "48px",
						Color:      "#000000",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Image component",
			component: Component{
				ID:      "header-image",
				Type:    ComponentTypeImage,
				ZIndex:  1,
				Visible: true,
				Position: Position{
					Mode: PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
				},
				Dimensions: Dimensions{
					Width:  "100%",
					Height: "400px",
				},
				Content: &ComponentContent{
					Image: &ImageContent{
						Src:       "/static/images/header.jpg",
						Alt:       "Header image",
						ObjectFit: "cover",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			var decoded Component
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
				return
			}

			if decoded.ID != tt.component.ID {
				t.Errorf("ID = %v, want %v", decoded.ID, tt.component.ID)
			}
			if decoded.Type != tt.component.Type {
				t.Errorf("Type = %v, want %v", decoded.Type, tt.component.Type)
			}
			if decoded.ZIndex != tt.component.ZIndex {
				t.Errorf("ZIndex = %v, want %v", decoded.ZIndex, tt.component.ZIndex)
			}
		})
	}
}

func TestComponentConfiguration_JSONSerialization(t *testing.T) {
	config := ComponentConfiguration{
		Version: "1.0",
		Metadata: ConfigMetadata{
			Name:        "Test Template",
			Category:    "card",
			Description: "Test description",
		},
		Layout: LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#ffffff",
		},
		Components: []Component{
			{
				ID:      "test-component",
				Type:    ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: Position{
					Mode: PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("100px"),
				},
				Dimensions: Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				Content: &ComponentContent{
					TextBox: &TextBoxContent{
						Text: "Test",
					},
				},
			},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ComponentConfiguration
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Version != config.Version {
		t.Errorf("Version = %v, want %v", decoded.Version, config.Version)
	}
	if decoded.Metadata.Name != config.Metadata.Name {
		t.Errorf("Metadata.Name = %v, want %v", decoded.Metadata.Name, config.Metadata.Name)
	}
	if len(decoded.Components) != len(config.Components) {
		t.Errorf("len(Components) = %v, want %v", len(decoded.Components), len(config.Components))
	}
}

func TestComponentOverrides_JSONSerialization(t *testing.T) {
	overrides := ComponentOverrides{
		Version: "1.0",
		Overrides: []ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"position": map[string]interface{}{
						"y": "250px",
					},
					"content": map[string]interface{}{
						"color": "#ff0000",
					},
				},
			},
		},
		Additions: []Component{
			{
				ID:      "new-component",
				Type:    ComponentTypeTextBox,
				ZIndex:  15,
				Visible: true,
				Position: Position{
					Mode: PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("300px"),
				},
				Dimensions: Dimensions{
					Width:  "70%",
					Height: "auto",
				},
				Content: &ComponentContent{
					TextBox: &TextBoxContent{
						Text: "New text",
					},
				},
			},
		},
		Removals: []string{"old-component"},
	}

	data, err := json.Marshal(overrides)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ComponentOverrides
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Version != overrides.Version {
		t.Errorf("Version = %v, want %v", decoded.Version, overrides.Version)
	}
	if len(decoded.Overrides) != len(overrides.Overrides) {
		t.Errorf("len(Overrides) = %v, want %v", len(decoded.Overrides), len(overrides.Overrides))
	}
	if len(decoded.Additions) != len(overrides.Additions) {
		t.Errorf("len(Additions) = %v, want %v", len(decoded.Additions), len(overrides.Additions))
	}
	if len(decoded.Removals) != len(overrides.Removals) {
		t.Errorf("len(Removals) = %v, want %v", len(decoded.Removals), len(overrides.Removals))
	}
}

func TestComponent_GetterMethods(t *testing.T) {
	t.Run("GetTextBoxContent success", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeTextBox,
			Content: &ComponentContent{
				TextBox: &TextBoxContent{
					Text: "Test",
				},
			},
		}

		content, err := comp.GetTextBoxContent()
		if err != nil {
			t.Errorf("GetTextBoxContent() error = %v", err)
		}
		if content.Text != "Test" {
			t.Errorf("Text = %v, want %v", content.Text, "Test")
		}
	})

	t.Run("GetTextBoxContent wrong type", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeImage,
		}

		_, err := comp.GetTextBoxContent()
		if err == nil {
			t.Error("GetTextBoxContent() expected error for wrong type")
		}
	})

	t.Run("GetImageContent success", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeImage,
			Content: &ComponentContent{
				Image: &ImageContent{
					Src: "/test.jpg",
				},
			},
		}

		content, err := comp.GetImageContent()
		if err != nil {
			t.Errorf("GetImageContent() error = %v", err)
		}
		if content.Src != "/test.jpg" {
			t.Errorf("Src = %v, want %v", content.Src, "/test.jpg")
		}
	})

	t.Run("GetBackgroundContent success", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeBackground,
			Content: &ComponentContent{
				Background: &BackgroundContent{
					Type:  "color",
					Color: "#ffffff",
				},
			},
		}

		content, err := comp.GetBackgroundContent()
		if err != nil {
			t.Errorf("GetBackgroundContent() error = %v", err)
		}
		if content.Color != "#ffffff" {
			t.Errorf("Color = %v, want %v", content.Color, "#ffffff")
		}
	})

	t.Run("GetOverlayContent success", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeOverlay,
			Content: &ComponentContent{
				Overlay: &OverlayContent{
					BackgroundColor: "rgba(0,0,0,0.5)",
				},
			},
		}

		content, err := comp.GetOverlayContent()
		if err != nil {
			t.Errorf("GetOverlayContent() error = %v", err)
		}
		if content.BackgroundColor != "rgba(0,0,0,0.5)" {
			t.Errorf("BackgroundColor = %v, want %v", content.BackgroundColor, "rgba(0,0,0,0.5)")
		}
	})

	t.Run("GetContainerLayout success", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeContainer,
			Layout: &ContainerLayout{
				Display: "flex",
			},
		}

		layout, err := comp.GetContainerLayout()
		if err != nil {
			t.Errorf("GetContainerLayout() error = %v", err)
		}
		if layout.Display != "flex" {
			t.Errorf("Display = %v, want %v", layout.Display, "flex")
		}
	})

	t.Run("GetDividerStyle success", func(t *testing.T) {
		comp := Component{
			Type: ComponentTypeDivider,
			Style: &DividerStyle{
				Height: "2px",
			},
		}

		style, err := comp.GetDividerStyle()
		if err != nil {
			t.Errorf("GetDividerStyle() error = %v", err)
		}
		if style.Height != "2px" {
			t.Errorf("Height = %v, want %v", style.Height, "2px")
		}
	})
}

func strPtr(s string) *string {
	return &s
}
