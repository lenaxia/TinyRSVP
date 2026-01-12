package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestComponentRenderer_RenderAnimation(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	animation := &models.AnimationConfig{
		Type:      models.AnimationTypeFade,
		Duration:  1000,
		Delay:     0,
		Easing:    models.AnimationEasingEaseInOut,
		Iteration: models.AnimationIterationOnce,
		Direction: models.AnimationDirectionNormal,
	}

	component := models.Component{
		ID:   "animated-text",
		Type: models.ComponentTypeTextBox,
		Position: models.Position{
			Mode: models.PositionModeAbsolute,
			X:    func() *string { s := "50%"; return &s }(),
			Y:    func() *string { s := "100px"; return &s }(),
		},
		Dimensions: models.Dimensions{
			Width:  "80%",
			Height: "auto",
		},
		ZIndex:    10,
		Visible:   true,
		Animation: animation,
		Content: &models.ComponentContent{
			TextBox: &models.TextBoxContent{
				Text:      "Animated Text",
				TextAlign: "center",
			},
		},
	}

	css := renderer.GenerateAnimationCSS(&component)
	if css == "" {
		t.Error("Expected animation CSS, got empty string")
	}

	if !strings.Contains(css, "animation-name") {
		t.Error("CSS should contain animation-name")
	}
	if !strings.Contains(css, "animation-duration: 1000ms") {
		t.Error("CSS should contain animation-duration")
	}
	if !strings.Contains(css, "animation-timing-function: ease-in-out") {
		t.Error("CSS should contain animation-timing-function")
	}
}

func TestComponentRenderer_RenderGridLayout(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	layoutMode := models.LayoutModeGrid
	component := models.Component{
		ID:   "grid-container",
		Type: models.ComponentTypeContainer,
		Position: models.Position{
			Mode: models.PositionModeRelative,
		},
		Dimensions: models.Dimensions{
			Width:  "100%",
			Height: "auto",
		},
		ZIndex:     5,
		Visible:    true,
		LayoutMode: &layoutMode,
		GridConfig: &models.GridConfig{
			Columns:  "repeat(3, 1fr)",
			Rows:     "auto",
			Gap:      "20px",
			AutoFlow: models.GridAutoFlowRow,
		},
	}

	css := renderer.GenerateLayoutCSS(&component)
	if css == "" {
		t.Error("Expected layout CSS, got empty string")
	}

	if !strings.Contains(css, "display: grid") {
		t.Error("CSS should contain display: grid")
	}
	if !strings.Contains(css, "grid-template-columns: repeat(3, 1fr)") {
		t.Error("CSS should contain grid-template-columns")
	}
	if !strings.Contains(css, "gap: 20px") {
		t.Error("CSS should contain gap")
	}
}

func TestComponentRenderer_RenderFlexLayout(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	layoutMode := models.LayoutModeFlexbox
	component := models.Component{
		ID:   "flex-container",
		Type: models.ComponentTypeContainer,
		Position: models.Position{
			Mode: models.PositionModeRelative,
		},
		Dimensions: models.Dimensions{
			Width:  "100%",
			Height: "auto",
		},
		ZIndex:     5,
		Visible:    true,
		LayoutMode: &layoutMode,
		FlexConfig: &models.FlexConfig{
			Direction:      models.FlexDirectionRow,
			Wrap:           models.FlexWrapWrap,
			JustifyContent: models.JustifyContentCenter,
			AlignItems:     models.AlignItemsCenter,
			Gap:            "20px",
		},
	}

	css := renderer.GenerateLayoutCSS(&component)
	if css == "" {
		t.Error("Expected layout CSS, got empty string")
	}

	if !strings.Contains(css, "display: flex") {
		t.Error("CSS should contain display: flex")
	}
	if !strings.Contains(css, "flex-direction: row") {
		t.Error("CSS should contain flex-direction")
	}
	if !strings.Contains(css, "justify-content: center") {
		t.Error("CSS should contain justify-content")
	}
}

func TestComponentRenderer_RenderImageEffects(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	component := models.Component{
		ID:   "effect-image",
		Type: models.ComponentTypeImage,
		Position: models.Position{
			Mode: models.PositionModeAbsolute,
			X:    func() *string { s := "0"; return &s }(),
			Y:    func() *string { s := "0"; return &s }(),
		},
		Dimensions: models.Dimensions{
			Width:  "100%",
			Height: "400px",
		},
		ZIndex:  1,
		Visible: true,
		Content: &models.ComponentContent{
			Image: &models.ImageContent{
				Src:       "/images/test.jpg",
				Alt:       "Test image",
				ObjectFit: "cover",
			},
		},
	}

	effects := &models.ImageEffects{
		Filter:    func() *string { s := "blur(5px)"; return &s }(),
		Transform: func() *string { s := "rotate(45deg)"; return &s }(),
		BlendMode: func() *string { s := "multiply"; return &s }(),
	}

	css := renderer.GenerateImageEffectsCSS(&component, effects)
	if css == "" {
		t.Error("Expected image effects CSS, got empty string")
	}

	if !strings.Contains(css, "filter: blur(5px)") {
		t.Error("CSS should contain filter")
	}
	if !strings.Contains(css, "transform: rotate(45deg)") {
		t.Error("CSS should contain transform")
	}
	if !strings.Contains(css, "mix-blend-mode: multiply") {
		t.Error("CSS should contain mix-blend-mode")
	}
}

func TestComponentRenderer_RenderTextEffects(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	component := models.Component{
		ID:   "effect-text",
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
				Text:      "Styled Text",
				TextAlign: "center",
			},
		},
	}

	effects := &models.TextEffects{
		Gradient:      func() *string { s := "linear-gradient(90deg, #ff0000, #00ff00)"; return &s }(),
		Stroke:        func() *string { s := "2px #000000"; return &s }(),
		Shadow:        func() *string { s := "2px 2px 4px rgba(0,0,0,0.5)"; return &s }(),
		LetterSpacing: func() *string { s := "0.1em"; return &s }(),
	}

	css := renderer.GenerateTextEffectsCSS(&component, effects)
	if css == "" {
		t.Error("Expected text effects CSS, got empty string")
	}

	if !strings.Contains(css, "background: linear-gradient") {
		t.Error("CSS should contain gradient background")
	}
	if !strings.Contains(css, "text-shadow: 2px 2px 4px") {
		t.Error("CSS should contain text-shadow")
	}
	if !strings.Contains(css, "letter-spacing: 0.1em") {
		t.Error("CSS should contain letter-spacing")
	}
}

func TestComponentRenderer_RenderVisibilityRules(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	component := models.Component{
		ID:   "responsive-component",
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
		Visibility: &models.VisibilityRules{
			ShowOnMobile:  true,
			ShowOnTablet:  false,
			ShowOnDesktop: true,
		},
		Content: &models.ComponentContent{
			TextBox: &models.TextBoxContent{
				Text: "Responsive Text",
			},
		},
	}

	css := renderer.GenerateVisibilityCSS(&component)
	if css == "" {
		t.Error("Expected visibility CSS, got empty string")
	}

	if !strings.Contains(css, "@media") {
		t.Error("CSS should contain media queries")
	}
}

func TestComponentRenderer_GenerateComponentCSS(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	animation := &models.AnimationConfig{
		Type:      models.AnimationTypeFade,
		Duration:  1000,
		Easing:    models.AnimationEasingEaseInOut,
		Iteration: models.AnimationIterationOnce,
		Direction: models.AnimationDirectionNormal,
	}

	component := models.Component{
		ID:   "full-featured-component",
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
		ZIndex:    10,
		Visible:   true,
		Animation: animation,
		Visibility: &models.VisibilityRules{
			ShowOnMobile:  true,
			ShowOnTablet:  true,
			ShowOnDesktop: true,
		},
		Content: &models.ComponentContent{
			TextBox: &models.TextBoxContent{
				Text:      "Full Featured",
				TextAlign: "center",
			},
		},
	}

	css := renderer.GenerateComponentCSS(&component)
	if css == "" {
		t.Error("Expected component CSS, got empty string")
	}

	if !strings.Contains(css, "#full-featured-component") {
		t.Error("CSS should contain component ID selector")
	}
}

func TestComponentRenderer_RenderWithAdvancedFeatures(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	animation := &models.AnimationConfig{
		Type:      models.AnimationTypeFade,
		Duration:  1000,
		Easing:    models.AnimationEasingEaseInOut,
		Iteration: models.AnimationIterationOnce,
		Direction: models.AnimationDirectionNormal,
	}

	configJSON := `{
		"version": "1.0",
		"metadata": {
			"name": "Test Template",
			"category": "card",
			"description": "Test"
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
				"content": {
					"text": "{{.Event.Title}}",
					"textAlign": "center"
				}
			}
		]
	}`

	template := &models.Template{
		ID:              1,
		Name:            "Test Template",
		Type:            models.TemplateTypeRSVPPage,
		Category:        models.CategoryCard,
		ComponentConfig: &configJSON,
	}

	event := &models.Event{
		ID:    1,
		Title: "Test Event",
	}

	component := models.Component{
		ID:        "test-animated",
		Type:      models.ComponentTypeTextBox,
		Animation: animation,
		Position: models.Position{
			Mode: models.PositionModeAbsolute,
		},
		Dimensions: models.Dimensions{
			Width:  "100%",
			Height: "auto",
		},
		ZIndex:  10,
		Visible: true,
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, event, template)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected rendered output, got empty string")
	}

	if !strings.Contains(output, "Test Event") {
		t.Error("Output should contain event title")
	}

	css := renderer.GenerateAnimationCSS(&component)
	if !strings.Contains(css, "animation") {
		t.Error("Animation CSS should be generated")
	}
}

func TestComponentRenderer_GenerateAnimationKeyframes(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	tests := []struct {
		name          string
		animationType models.AnimationType
		expectKeyword string
	}{
		{"fade animation", models.AnimationTypeFade, "opacity"},
		{"slide animation", models.AnimationTypeSlide, "transform"},
		{"scale animation", models.AnimationTypeScale, "transform"},
		{"rotate animation", models.AnimationTypeRotate, "transform"},
		{"bounce animation", models.AnimationTypeBounce, "transform"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyframes := renderer.GenerateAnimationKeyframes(tt.animationType)
			if keyframes == "" {
				t.Error("Expected keyframes, got empty string")
			}
			if !strings.Contains(keyframes, "@keyframes") {
				t.Error("Keyframes should contain @keyframes")
			}
			if !strings.Contains(keyframes, tt.expectKeyword) {
				t.Errorf("Keyframes should contain %s", tt.expectKeyword)
			}
		})
	}
}

func TestComponentRenderer_GenerateResponsiveVisibilityCSS(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)

	tests := []struct {
		name       string
		visibility *models.VisibilityRules
		expectHide []string
	}{
		{
			name: "mobile only",
			visibility: &models.VisibilityRules{
				ShowOnMobile:  true,
				ShowOnTablet:  false,
				ShowOnDesktop: false,
			},
			expectHide: []string{"tablet", "desktop"},
		},
		{
			name: "desktop only",
			visibility: &models.VisibilityRules{
				ShowOnMobile:  false,
				ShowOnTablet:  false,
				ShowOnDesktop: true,
			},
			expectHide: []string{"mobile", "tablet"},
		},
		{
			name: "all visible",
			visibility: &models.VisibilityRules{
				ShowOnMobile:  true,
				ShowOnTablet:  true,
				ShowOnDesktop: true,
			},
			expectHide: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := models.Component{
				ID:         "test-component",
				Type:       models.ComponentTypeTextBox,
				Visibility: tt.visibility,
				Position: models.Position{
					Mode: models.PositionModeRelative,
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
			}

			css := renderer.GenerateVisibilityCSS(&component)

			if len(tt.expectHide) > 0 {
				if css == "" {
					t.Error("Expected visibility CSS for hidden breakpoints")
				}
				if !strings.Contains(css, "@media") {
					t.Error("CSS should contain media queries")
				}
			}
		})
	}
}
