package templates

import (
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type ThemeMigrator struct{}

func NewThemeMigrator() *ThemeMigrator {
	return &ThemeMigrator{}
}

func (m *ThemeMigrator) MigratePlainText() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Plain Text",
			Category:    "simple",
			Description: "Simple, clean text-based invitation",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#f8f9fa",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
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
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
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
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "2.5rem",
						FontWeight: "700",
						Color:      "#2c3e50",
						LineHeight: "1.2",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "2rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						Color:      "#666666",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1rem",
						Color:      "#888888",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}

func (m *ThemeMigrator) MigrateWeddingElegance() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Wedding Elegance",
			Category:    "card",
			Description: "Elegant wedding invitation with floral design",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#f8f9fa",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
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
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  1,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/wedding-elegance-header.svg{{end}}",
						Alt:            "Wedding invitation design",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						Height: "200px",
					},
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
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
						FontFamily: "Georgia, 'Times New Roman', serif",
						FontSize:   "2.5rem",
						FontWeight: "700",
						Color:      "#f4c2c2",
						LineHeight: "1.2",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "1.5rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						Color:      "#666666",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(4),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						Color:      "#666666",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}

func (m *ThemeMigrator) MigrateBirthdayCelebration() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Birthday Celebration",
			Category:    "card",
			Description: "Fun birthday invitation with balloons and confetti",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#fff5f7",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "100%",
				},
				ZIndex:  0,
				Visible: true,
				Content: &models.ComponentContent{
					Background: &models.BackgroundContent{
						Type:     "gradient",
						Gradient: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
					},
				},
			},
			{
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  1,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/birthday-celebration-header.svg{{end}}",
						Alt:            "Birthday celebration design",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						Height: "200px",
					},
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
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
						FontFamily: "'Comic Sans MS', cursive, sans-serif",
						FontSize:   "2.5rem",
						FontWeight: "700",
						Color:      "#ff6b9d",
						LineHeight: "1.2",
						TextShadow: "2px 2px 4px rgba(0,0,0,0.1)",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "2rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						FontWeight: "600",
						Color:      "#2c3e50",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(4),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1rem",
						Color:      "#555555",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}

func (m *ThemeMigrator) MigrateCorporateProfessional() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Corporate Professional",
			Category:    "card",
			Description: "Professional business event invitation",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#ffffff",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
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
			},
			{
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "250px",
				},
				ZIndex:  1,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/corporate-professional-header.svg{{end}}",
						Alt:            "Corporate event design",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						Height: "180px",
					},
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
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
						FontFamily: "'Helvetica Neue', Helvetica, Arial, sans-serif",
						FontSize:   "2rem",
						FontWeight: "600",
						Color:      "#1a365d",
						LineHeight: "1.3",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "1.5rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "'Helvetica Neue', Helvetica, Arial, sans-serif",
						FontSize:   "1.125rem",
						Color:      "#4a5568",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(4),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "'Helvetica Neue', Helvetica, Arial, sans-serif",
						FontSize:   "1rem",
						Color:      "#718096",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}

func (m *ThemeMigrator) MigrateHolidayFestive() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Holiday Festive",
			Category:    "card",
			Description: "Festive holiday celebration invitation",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#f0f9ff",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "100%",
				},
				ZIndex:  0,
				Visible: true,
				Content: &models.ComponentContent{
					Background: &models.BackgroundContent{
						Type:     "gradient",
						Gradient: "linear-gradient(135deg, #e0f2fe 0%, #dbeafe 100%)",
					},
				},
			},
			{
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  1,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/holiday-festive-header.svg{{end}}",
						Alt:            "Holiday festive design",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						Height: "200px",
					},
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
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
						FontFamily: "'Brush Script MT', cursive",
						FontSize:   "2.5rem",
						FontWeight: "400",
						Color:      "#dc2626",
						LineHeight: "1.2",
						TextShadow: "2px 2px 4px rgba(0,0,0,0.1)",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "2rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						FontWeight: "600",
						Color:      "#166534",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(4),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1rem",
						Color:      "#065f46",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}

func (m *ThemeMigrator) MigrateGardenParty() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Garden Party",
			Category:    "card",
			Description: "Fresh garden party invitation with floral elements",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#f0fdf4",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "100%",
				},
				ZIndex:  0,
				Visible: true,
				Content: &models.ComponentContent{
					Background: &models.BackgroundContent{
						Type:     "gradient",
						Gradient: "linear-gradient(135deg, #dcfce7 0%, #f0fdf4 100%)",
					},
				},
			},
			{
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  1,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/garden-party-header.svg{{end}}",
						Alt:            "Garden party design",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						Height: "200px",
					},
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
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
						FontFamily: "'Palatino Linotype', 'Book Antiqua', Palatino, serif",
						FontSize:   "2.5rem",
						FontWeight: "600",
						Color:      "#166534",
						LineHeight: "1.2",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "2rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						Color:      "#15803d",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(4),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1rem",
						Color:      "#16a34a",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}

func (m *ThemeMigrator) MigrateModernMinimalist() (*models.ComponentConfiguration, error) {
	config := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Modern Minimalist",
			Category:    "card",
			Description: "Clean, modern minimalist invitation design",
		},
		Layout: models.LayoutConfig{
			Mode:            "card",
			CardWidth:       "800px",
			CardMaxWidth:    "90vw",
			BackgroundColor: "#ffffff",
		},
		Components: []models.Component{
			{
				ID:   "page-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("0"),
					Y:    strPtr("0"),
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
			},
			{
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "280px",
				},
				ZIndex:  1,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/modern-minimalist-header.svg{{end}}",
						Alt:            "Modern minimalist design",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						Height: "200px",
					},
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(2),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:          "{{.Event.Title}}",
						TextAlign:     "center",
						FontFamily:    "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:      "2rem",
						FontWeight:    "300",
						Color:         "#1a202c",
						LineHeight:    "1.3",
						LetterSpacing: "0.02em",
					},
				},
				Responsive: &models.ResponsiveConfig{
					Mobile: &models.ResponsiveBreakpoint{
						FontSize: "1.5rem",
					},
				},
			},
			{
				ID:   "date-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(3),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{formatDateTime .Event.StartTime}}",
						TextAlign:  "center",
						FontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1.125rem",
						FontWeight: "400",
						Color:      "#4a5568",
						LineHeight: "1.5",
					},
				},
			},
			{
				ID:   "location-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: intPtr(4),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Location}}",
						TextAlign:  "center",
						FontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
						FontSize:   "1rem",
						FontWeight: "400",
						Color:      "#718096",
						LineHeight: "1.5",
					},
				},
			},
		},
	}

	return config, nil
}
