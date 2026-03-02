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
				Content: &models.ComponentContent{
					Background: &models.BackgroundContent{
						Type:     "gradient",
						Gradient: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
					},
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
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "{{.Event.Title}}",
						TextAlign:  "center",
						FontFamily: "Playfair Display, serif",
						FontSize:   "56px",
						FontWeight: "700",
						Color:      "#ffffff",
					},
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
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "Join us for an unforgettable celebration",
						TextAlign:  "center",
						FontFamily: "Lato, sans-serif",
						FontSize:   "24px",
						FontWeight: "300",
						Color:      "#ffffff",
					},
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
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
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
					Padding:        "60px 20px",
				},
				Style: &models.DividerStyle{
					BackgroundColor: "#f8f9fa",
				},
			},
			{
				ID:   "cta-heading",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "Ready to RSVP?",
						TextAlign:  "center",
						FontFamily: "Playfair Display, serif",
						FontSize:   "36px",
						FontWeight: "700",
						Color:      "#2c3e50",
					},
				},
			},
			{
				ID:   "cta-button",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 2; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "auto",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "RSVP Now",
						TextAlign:  "center",
						FontFamily: "Lato, sans-serif",
						FontSize:   "18px",
						FontWeight: "600",
						Color:      "#ffffff",
					},
				},
				Style: &models.DividerStyle{
					BackgroundColor: "#667eea",
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
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  5,
				Visible: true,
				Layout: &models.ContainerLayout{
					Display: "grid",
					Gap:     "20px",
					Padding: "40px 20px",
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
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "/static/images/placeholder-1.jpg",
						Alt:            "Gallery image 1",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
			},
			{
				ID:   "gallery-image-2",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 2; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "/static/images/placeholder-2.jpg",
						Alt:            "Gallery image 2",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
				},
			},
			{
				ID:   "gallery-image-3",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 3; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "300px",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					Image: &models.ImageContent{
						Src:            "/static/images/placeholder-3.jpg",
						Alt:            "Gallery image 3",
						ObjectFit:      "cover",
						ObjectPosition: "center",
					},
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
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
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
					Padding:        "60px 40px",
				},
				Style: &models.DividerStyle{
					BackgroundColor: "#ffffff",
					BorderRadius:    "12px",
				},
			},
			{
				ID:   "testimonial-quote",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 1; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "\"This was an amazing event! Everything was perfect.\"",
						TextAlign:  "center",
						FontFamily: "Georgia, serif",
						FontSize:   "24px",
						FontWeight: "400",
						Color:      "#2c3e50",
					},
				},
			},
			{
				ID:   "testimonial-author",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode:  models.PositionModeRelative,
					Order: func() *int { i := 2; return &i }(),
				},
				Dimensions: models.Dimensions{
					Width:  "auto",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
				Content: &models.ComponentContent{
					TextBox: &models.TextBoxContent{
						Text:       "— Guest Name",
						TextAlign:  "center",
						FontFamily: "Lato, sans-serif",
						FontSize:   "16px",
						FontWeight: "600",
						Color:      "#666666",
					},
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
