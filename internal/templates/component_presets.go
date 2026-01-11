package templates

import (
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type PresetMetadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	ThumbnailURL string   `json:"thumbnailUrl"`
}

type ComponentPreset struct {
	Metadata   PresetMetadata     `json:"metadata"`
	Components []models.Component `json:"components"`
}

var presets = map[string]*ComponentPreset{
	"hero-section": {
		Metadata: PresetMetadata{
			Name:         "hero-section",
			Description:  "Hero section with background image and title",
			Category:     "header",
			Tags:         []string{"hero", "header", "banner"},
			ThumbnailURL: "/static/images/presets/hero-section.png",
		},
		Components: []models.Component{
			{
				ID:   "hero-background",
				Type: models.ComponentTypeBackground,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    func() *string { s := "0"; return &s }(),
					Y:    func() *string { s := "0"; return &s }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "500px",
				},
				ZIndex:  0,
				Visible: true,
				Content: map[string]interface{}{
					"type":  "gradient",
					"gradient": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
				},
			},
			{
				ID:   "hero-title",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    func() *string { s := "50%"; return &s }(),
					Y:    func() *string { s := "200px"; return &s }(),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "{{.Event.Title}}",
					"textAlign":  "center",
					"fontFamily": "Playfair Display, serif",
					"fontSize":   "56px",
					"fontWeight": "700",
					"color":      "#ffffff",
				},
			},
			{
				ID:   "hero-subtitle",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    func() *string { s := "50%"; return &s }(),
					Y:    func() *string { s := "280px"; return &s }(),
				},
				Dimensions: models.Dimensions{
					Width:  "70%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "Join us for an unforgettable celebration",
					"textAlign":  "center",
					"fontFamily": "Lato, sans-serif",
					"fontSize":   "24px",
					"fontWeight": "300",
					"color":      "#ffffff",
				},
			},
		},
	},
	"call-to-action": {
		Metadata: PresetMetadata{
			Name:         "call-to-action",
			Description:  "Call-to-action section with button",
			Category:     "action",
			Tags:         []string{"cta", "button", "action"},
			ThumbnailURL: "/static/images/presets/call-to-action.png",
		},
		Components: []models.Component{
			{
				ID:   "cta-container",
				Type: models.ComponentTypeContainer,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  5,
				Visible: true,
				Layout: map[string]interface{}{
					"display":        "flex",
					"flexDirection":  "column",
					"alignItems":     "center",
					"justifyContent": "center",
					"gap":            "20px",
					"padding":        "60px 20px",
				},
				Style: map[string]interface{}{
					"backgroundColor": "#f8f9fa",
				},
			},
			{
				ID:   "cta-heading",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "Ready to RSVP?",
					"textAlign":  "center",
					"fontFamily": "Playfair Display, serif",
					"fontSize":   "36px",
					"fontWeight": "700",
					"color":      "#2c3e50",
				},
			},
			{
				ID:   "cta-button",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 2; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "auto",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "RSVP Now",
					"textAlign":  "center",
					"fontFamily": "Lato, sans-serif",
					"fontSize":   "18px",
					"fontWeight": "600",
					"color":      "#ffffff",
				},
				Style: map[string]interface{}{
					"backgroundColor": "#667eea",
					"padding":         "15px 40px",
					"borderRadius":    "8px",
					"cursor":          "pointer",
				},
			},
		},
	},
	"image-gallery": {
		Metadata: PresetMetadata{
			Name:         "image-gallery",
			Description:  "Grid-based image gallery",
			Category:     "media",
			Tags:         []string{"gallery", "images", "grid"},
			ThumbnailURL: "/static/images/presets/image-gallery.png",
		},
		Components: []models.Component{
			{
				ID:   "gallery-container",
				Type: models.ComponentTypeContainer,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  5,
				Visible: true,
				Layout: map[string]interface{}{
					"display":             "grid",
					"gridTemplateColumns": "repeat(3, 1fr)",
					"gap":                 "20px",
					"padding":             "40px 20px",
				},
				LayoutMode: func() *models.LayoutMode { m := models.LayoutModeGrid; return &m }(),
				GridConfig: &models.GridConfig{
					Columns:  "repeat(3, 1fr)",
					Gap:      "20px",
					AutoFlow: models.GridAutoFlowRow,
				},
			},
			{
				ID:   "gallery-image-1",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"src":            "/static/images/placeholder-1.jpg",
					"alt":            "Gallery image 1",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
			},
			{
				ID:   "gallery-image-2",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 2; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"src":            "/static/images/placeholder-2.jpg",
					"alt":            "Gallery image 2",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
			},
			{
				ID:   "gallery-image-3",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 3; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"src":            "/static/images/placeholder-3.jpg",
					"alt":            "Gallery image 3",
					"objectFit":      "cover",
					"objectPosition": "center",
				},
			},
		},
	},
	"testimonial": {
		Metadata: PresetMetadata{
			Name:         "testimonial",
			Description:  "Testimonial section with quote and author",
			Category:     "content",
			Tags:         []string{"testimonial", "quote", "review"},
			ThumbnailURL: "/static/images/presets/testimonial.png",
		},
		Components: []models.Component{
			{
				ID:   "testimonial-container",
				Type: models.ComponentTypeContainer,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  5,
				Visible: true,
				Layout: map[string]interface{}{
					"display":        "flex",
					"flexDirection":  "column",
					"alignItems":     "center",
					"justifyContent": "center",
					"gap":            "20px",
					"padding":        "60px 40px",
				},
				Style: map[string]interface{}{
					"backgroundColor": "#ffffff",
					"borderRadius":    "12px",
					"boxShadow":       "0 4px 12px rgba(0,0,0,0.1)",
				},
			},
			{
				ID:   "testimonial-quote",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "\"This was an amazing event! Everything was perfect.\"",
					"textAlign":  "center",
					"fontFamily": "Georgia, serif",
					"fontSize":   "24px",
					"fontWeight": "400",
					"fontStyle":  "italic",
					"color":      "#2c3e50",
				},
			},
			{
				ID:   "testimonial-author",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeRelative,
					Order: func() *int { i := 2; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "auto",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: map[string]interface{}{
					"text":       "— Guest Name",
					"textAlign":  "center",
					"fontFamily": "Lato, sans-serif",
					"fontSize":   "16px",
					"fontWeight": "600",
					"color":      "#666666",
				},
			},
		},
	},
}

func GetPreset(name string) (*ComponentPreset, error) {
	if name == "" {
		return nil, fmt.Errorf("preset name cannot be empty")
	}

	preset, exists := presets[name]
	if !exists {
		return nil, fmt.Errorf("preset not found: %s", name)
	}

	return preset, nil
}

func ListPresets() []PresetMetadata {
	result := make([]PresetMetadata, 0, len(presets))
	for _, preset := range presets {
		result = append(result, preset.Metadata)
	}
	return result
}

