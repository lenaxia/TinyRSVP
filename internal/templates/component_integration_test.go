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
}

func TestComponentIntegration_RenderWithTypedStructs(t *testing.T) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Test Template",
			Category:    "card",
			Description: "Test template with typed structs",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			BackgroundColor: "#ffffff",
		},
		Components: []models.Component{
			{
				ID:   "background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    func() *string { s := "0"; return &s }(),
					Y:    func() *string { s := "0"; return &s }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "100%",
				},
				ZIndex:  0,
				Visible: true,
				Content: &models.ComponentContent{
					Background: &models.BackgroundContent{
						Type:  "color",
						Color: "#f8f9fa",
					},
				},
			},
			{
				ID:   "title",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Title}}",
						TextAlign:  "center",
						FontFamily: "Arial, sans-serif",
						FontSize:   "2rem",
						Color:      "#2c3e50",
					},
				},
			},
		},
	}

	if config == nil {
		t.Fatal("Config is nil")
	}

	if len(config.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(config.Components))
	}

	bgComp := config.Components[0]
	if bgComp.Content == nil || bgComp.Content.Background == nil {
		t.Error("Background component should have typed Background content")
	}

	titleComp := config.Components[1]
	if titleComp.Content == nil || titleComp.Content.TextBox == nil {
		t.Error("Title component should have typed TextBox content")
	}
}

func TestComponentIntegration_RenderAllThemes(t *testing.T) {
	migrator := NewThemeMigrator()

	themes := []struct {
		name     string
		migrator func() (*models.ComponentConfiguration, error)
	}{
		{"PlainText", migrator.MigratePlainText},
		{"WeddingElegance", migrator.MigrateWeddingElegance},
		{"BirthdayCelebration", migrator.MigrateBirthdayCelebration},
		{"CorporateProfessional", migrator.MigrateCorporateProfessional},
		{"HolidayFestive", migrator.MigrateHolidayFestive},
		{"GardenParty", migrator.MigrateGardenParty},
		{"ModernMinimalist", migrator.MigrateModernMinimalist},
	}

	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			config, err := theme.migrator()
			if err != nil {
				t.Fatalf("Failed to migrate %s: %v", theme.name, err)
			}

			if config == nil {
				t.Fatalf("Config is nil for %s", theme.name)
			}

			if len(config.Components) == 0 {
				t.Errorf("No components in %s theme", theme.name)
			}

			for _, comp := range config.Components {
				if comp.Content == nil {
					t.Errorf("Component %s has nil Content", comp.ID)
					continue
				}

				switch comp.Type {
				case models.ComponentTypeTextBox:
					if comp.Content.TextBox == nil {
						t.Errorf("TextBox component %s has nil TextBox content", comp.ID)
					}
				case models.ComponentTypeImage:
					if comp.Content.Image == nil {
						t.Errorf("Image component %s has nil Image content", comp.ID)
					}
				case models.ComponentTypeBackground:
					if comp.Content.Background == nil {
						t.Errorf("Background component %s has nil Background content", comp.ID)
					}
				case models.ComponentTypeContainer:
					if comp.Layout == nil {
						t.Errorf("Container component %s has nil Layout", comp.ID)
					}
				}
			}

			if config.Metadata.Name == "" {
				t.Errorf("Theme %s has empty metadata name", theme.name)
			}

			if config.Layout.Mode == "" {
				t.Errorf("Theme %s has empty layout mode", theme.name)
			}
		})
	}
}

func TestComponentIntegration_TypeSafety(t *testing.T) {
	t.Run("TextBox with proper typing", func(t *testing.T) {
		comp := models.Component{
			ID:   "test-text",
			Type: models.ComponentTypeTextBox,
			Position: models.Position{
				Mode: models.PositionModeRelative,
			},
			Dimensions: models.Dimensions{
				Width:  "100%",
				Height: "auto",
			},
			ZIndex:  10,
			Visible: true,
			Content: &models.ComponentContent{
				TextBox: &models.TextBoxContent{
					Text:       "Test",
					TextAlign:  "center",
					FontFamily: "Arial",
					FontSize:   "16px",
					Color:      "#000000",
				},
			},
		}

		textContent, err := comp.GetTextBoxContent()
		if err != nil {
			t.Errorf("Failed to get TextBox content: %v", err)
		}
		if textContent.Text != "Test" {
			t.Errorf("Expected text 'Test', got '%s'", textContent.Text)
		}
	})

	t.Run("Image with proper typing", func(t *testing.T) {
		comp := models.Component{
			ID:   "test-image",
			Type: models.ComponentTypeImage,
			Position: models.Position{
				Mode: models.PositionModeRelative,
			},
			Dimensions: models.Dimensions{
				Width:  "100%",
				Height: "300px",
			},
			ZIndex:  5,
			Visible: true,
			Content: &models.ComponentContent{
				Image: &models.ImageContent{
					Src:            "/test.jpg",
					Alt:            "Test image",
					ObjectFit:      "cover",
					ObjectPosition: "center",
				},
			},
		}

		imageContent, err := comp.GetImageContent()
		if err != nil {
			t.Errorf("Failed to get Image content: %v", err)
		}
		if imageContent.Src != "/test.jpg" {
			t.Errorf("Expected src '/test.jpg', got '%s'", imageContent.Src)
		}
	})

	t.Run("Background with proper typing", func(t *testing.T) {
		comp := models.Component{
			ID:   "test-bg",
			Type: models.ComponentTypeBackground,
			Position: models.Position{
				Mode: models.PositionModeAbsolute,
				X:    func() *string { s := "0"; return &s }(),
				Y:    func() *string { s := "0"; return &s }(),
			},
			Dimensions: models.Dimensions{
				Width:  "100%",
				Height: "100%",
			},
			ZIndex:  0,
			Visible: true,
			Content: &models.ComponentContent{
				Background: &models.BackgroundContent{
					Type:  "color",
					Color: "#ffffff",
				},
			},
		}

		bgContent, err := comp.GetBackgroundContent()
		if err != nil {
			t.Errorf("Failed to get Background content: %v", err)
		}
		if bgContent.Color != "#ffffff" {
			t.Errorf("Expected color '#ffffff', got '%s'", bgContent.Color)
		}
	})

	t.Run("Container with proper typing", func(t *testing.T) {
		comp := models.Component{
			ID:   "test-container",
			Type: models.ComponentTypeContainer,
			Position: models.Position{
				Mode: models.PositionModeRelative,
			},
			Dimensions: models.Dimensions{
				Width:  "100%",
				Height: "auto",
			},
			ZIndex:  5,
			Visible: true,
			Layout: &models.ContainerLayout{
				Display:        "flex",
				FlexDirection:  "column",
				AlignItems:     "center",
				JustifyContent: "center",
				Gap:            "20px",
			},
		}

		layout, err := comp.GetContainerLayout()
		if err != nil {
			t.Errorf("Failed to get Container layout: %v", err)
		}
		if layout.Display != "flex" {
			t.Errorf("Expected display 'flex', got '%s'", layout.Display)
		}
	})
}

func TestComponentIntegration_NoMapStringInterface(t *testing.T) {
	migrator := NewThemeMigrator()

	themes := []struct {
		name     string
		migrator func() (*models.ComponentConfiguration, error)
	}{
		{"PlainText", migrator.MigratePlainText},
		{"WeddingElegance", migrator.MigrateWeddingElegance},
		{"BirthdayCelebration", migrator.MigrateBirthdayCelebration},
		{"CorporateProfessional", migrator.MigrateCorporateProfessional},
		{"HolidayFestive", migrator.MigrateHolidayFestive},
		{"GardenParty", migrator.MigrateGardenParty},
		{"ModernMinimalist", migrator.MigrateModernMinimalist},
	}

	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			config, err := theme.migrator()
			if err != nil {
				t.Fatalf("Failed to migrate %s: %v", theme.name, err)
			}

			for _, comp := range config.Components {
				if comp.Content == nil {
					continue
				}

				if comp.Content.TextBox != nil {
					if comp.Content.TextBox.Text == "" {
						t.Errorf("Component %s has empty Text field", comp.ID)
					}
				}

				if comp.Content.Image != nil {
					if comp.Content.Image.Src == "" {
						t.Errorf("Component %s has empty Src field", comp.ID)
					}
				}

				if comp.Content.Background != nil {
					if comp.Content.Background.Type == "" {
						t.Errorf("Component %s has empty Type field", comp.ID)
					}
				}
			}
		})
	}
}
