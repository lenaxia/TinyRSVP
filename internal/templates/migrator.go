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
				Content: map[string]interface{}{
					"type":  "color",
					"color": "#f8f9fa",
				},
			},
			{
				ID:   "title-text",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "2.5rem",
					"fontWeight": "700",
					"color":      "#2c3e50",
					"lineHeight": "1.2",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "2rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"color":      "#666666",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1rem",
					"color":      "#888888",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"type":  "color",
					"color": "#f8f9fa",
				},
			},
			{
				ID:   "header-image",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: intPtr(1),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  1,
				Visible: true,
				Content: map[string]interface{}{
					"src":            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/wedding-elegance-header.svg{{end}}",
					"alt":            "Wedding invitation design",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"height": "200px",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "Georgia, 'Times New Roman', serif",
					"fontSize":   "2.5rem",
					"fontWeight": "700",
					"color":      "#f4c2c2",
					"lineHeight": "1.2",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "1.5rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"color":      "#666666",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"color":      "#666666",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"type":     "gradient",
					"gradient": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
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
				Content: map[string]interface{}{
					"src":            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/birthday-celebration-header.svg{{end}}",
					"alt":            "Birthday celebration design",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"height": "200px",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "'Comic Sans MS', cursive, sans-serif",
					"fontSize":   "2.5rem",
					"fontWeight": "700",
					"color":      "#ff6b9d",
					"lineHeight": "1.2",
					"textShadow": "2px 2px 4px rgba(0,0,0,0.1)",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "2rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"fontWeight": "600",
					"color":      "#2c3e50",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1rem",
					"color":      "#555555",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"type":  "color",
					"color": "#ffffff",
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
				Content: map[string]interface{}{
					"src":            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/corporate-professional-header.svg{{end}}",
					"alt":            "Corporate event design",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"height": "180px",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "'Helvetica Neue', Helvetica, Arial, sans-serif",
					"fontSize":   "2rem",
					"fontWeight": "600",
					"color":      "#1a365d",
					"lineHeight": "1.3",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "1.5rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "'Helvetica Neue', Helvetica, Arial, sans-serif",
					"fontSize":   "1.125rem",
					"color":      "#4a5568",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "'Helvetica Neue', Helvetica, Arial, sans-serif",
					"fontSize":   "1rem",
					"color":      "#718096",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"type":     "gradient",
					"gradient": "linear-gradient(135deg, #e0f2fe 0%, #dbeafe 100%)",
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
				Content: map[string]interface{}{
					"src":            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/holiday-festive-header.svg{{end}}",
					"alt":            "Holiday festive design",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"height": "200px",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "'Brush Script MT', cursive",
					"fontSize":   "2.5rem",
					"fontWeight": "400",
					"color":      "#dc2626",
					"lineHeight": "1.2",
					"textShadow": "2px 2px 4px rgba(0,0,0,0.1)",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "2rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"fontWeight": "600",
					"color":      "#166534",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1rem",
					"color":      "#065f46",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"type":     "gradient",
					"gradient": "linear-gradient(135deg, #dcfce7 0%, #f0fdf4 100%)",
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
				Content: map[string]interface{}{
					"src":            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/garden-party-header.svg{{end}}",
					"alt":            "Garden party design",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"height": "200px",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "'Palatino Linotype', 'Book Antiqua', Palatino, serif",
					"fontSize":   "2.5rem",
					"fontWeight": "600",
					"color":      "#166534",
					"lineHeight": "1.2",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "2rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"color":      "#15803d",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1rem",
					"color":      "#16a34a",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"type":  "color",
					"color": "#ffffff",
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
				Content: map[string]interface{}{
					"src":            "{{if .Event.CustomThemeImageURL}}{{.Event.CustomThemeImageURL}}{{else}}/static/images/themes/modern-minimalist-header.svg{{end}}",
					"alt":            "Modern minimalist design",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"height": "200px",
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
				Content: map[string]interface{}{
					"text":          "{{.Event.Title}}",
					"textAlign":     "center",
					"fontFamily":    "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":      "2rem",
					"fontWeight":    "300",
					"color":         "#1a202c",
					"lineHeight":    "1.3",
					"letterSpacing": "0.02em",
				},
				Responsive: map[string]interface{}{
					"mobile": map[string]interface{}{
						"fontSize": "1.5rem",
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
				Content: map[string]interface{}{
					"text":       "{{formatDateTime .Event.StartTime}}",
					"textAlign":  "center",
					"fontFamily": "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1.125rem",
					"fontWeight": "400",
					"color":      "#4a5568",
					"lineHeight": "1.5",
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
				Content: map[string]interface{}{
					"text":       "{{.Event.Location}}",
					"textAlign":  "center",
					"fontFamily": "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
					"fontSize":   "1rem",
					"fontWeight": "400",
					"color":      "#718096",
					"lineHeight": "1.5",
				},
			},
		},
	}

	return config, nil
}
