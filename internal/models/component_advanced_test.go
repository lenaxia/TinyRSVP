package models

import (
	"encoding/json"
	"testing"
)

func TestAnimationType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		animType AnimationType
		want     bool
	}{
		{"valid fade", AnimationTypeFade, true},
		{"valid slide", AnimationTypeSlide, true},
		{"valid scale", AnimationTypeScale, true},
		{"valid rotate", AnimationTypeRotate, true},
		{"valid bounce", AnimationTypeBounce, true},
		{"invalid empty", AnimationType(""), false},
		{"invalid unknown", AnimationType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.animType.IsValid(); got != tt.want {
				t.Errorf("AnimationType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnimationEasing_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		easing AnimationEasing
		want   bool
	}{
		{"valid linear", AnimationEasingLinear, true},
		{"valid ease-in", AnimationEasingEaseIn, true},
		{"valid ease-out", AnimationEasingEaseOut, true},
		{"valid ease-in-out", AnimationEasingEaseInOut, true},
		{"invalid empty", AnimationEasing(""), false},
		{"invalid unknown", AnimationEasing("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.easing.IsValid(); got != tt.want {
				t.Errorf("AnimationEasing.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnimationIteration_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		iteration AnimationIteration
		want      bool
	}{
		{"valid once", AnimationIterationOnce, true},
		{"valid infinite", AnimationIterationInfinite, true},
		{"valid count", AnimationIterationCount, true},
		{"invalid empty", AnimationIteration(""), false},
		{"invalid unknown", AnimationIteration("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.iteration.IsValid(); got != tt.want {
				t.Errorf("AnimationIteration.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnimationDirection_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		direction AnimationDirection
		want      bool
	}{
		{"valid normal", AnimationDirectionNormal, true},
		{"valid reverse", AnimationDirectionReverse, true},
		{"valid alternate", AnimationDirectionAlternate, true},
		{"invalid empty", AnimationDirection(""), false},
		{"invalid unknown", AnimationDirection("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.direction.IsValid(); got != tt.want {
				t.Errorf("AnimationDirection.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnimationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *AnimationConfig
		wantErr bool
	}{
		{
			name: "valid animation config",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  1000,
				Delay:     0,
				Easing:    AnimationEasingEaseInOut,
				Iteration: AnimationIterationOnce,
				Direction: AnimationDirectionNormal,
			},
			wantErr: false,
		},
		{
			name: "valid with count",
			config: &AnimationConfig{
				Type:           AnimationTypeSlide,
				Duration:       500,
				Delay:          100,
				Easing:         AnimationEasingLinear,
				Iteration:      AnimationIterationCount,
				IterationCount: func() *int { i := 3; return &i }(),
				Direction:      AnimationDirectionAlternate,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			config: &AnimationConfig{
				Type:      AnimationType("invalid"),
				Duration:  1000,
				Easing:    AnimationEasingLinear,
				Iteration: AnimationIterationOnce,
				Direction: AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "negative duration",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  -100,
				Easing:    AnimationEasingLinear,
				Iteration: AnimationIterationOnce,
				Direction: AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "negative delay",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  1000,
				Delay:     -50,
				Easing:    AnimationEasingLinear,
				Iteration: AnimationIterationOnce,
				Direction: AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "invalid easing",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  1000,
				Easing:    AnimationEasing("invalid"),
				Iteration: AnimationIterationOnce,
				Direction: AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "invalid iteration",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  1000,
				Easing:    AnimationEasingLinear,
				Iteration: AnimationIteration("invalid"),
				Direction: AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "invalid direction",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  1000,
				Easing:    AnimationEasingLinear,
				Iteration: AnimationIterationOnce,
				Direction: AnimationDirection("invalid"),
			},
			wantErr: true,
		},
		{
			name: "count iteration without count value",
			config: &AnimationConfig{
				Type:      AnimationTypeFade,
				Duration:  1000,
				Easing:    AnimationEasingLinear,
				Iteration: AnimationIterationCount,
				Direction: AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "count iteration with zero count",
			config: &AnimationConfig{
				Type:           AnimationTypeFade,
				Duration:       1000,
				Easing:         AnimationEasingLinear,
				Iteration:      AnimationIterationCount,
				IterationCount: func() *int { i := 0; return &i }(),
				Direction:      AnimationDirectionNormal,
			},
			wantErr: true,
		},
		{
			name: "count iteration with negative count",
			config: &AnimationConfig{
				Type:           AnimationTypeFade,
				Duration:       1000,
				Easing:         AnimationEasingLinear,
				Iteration:      AnimationIterationCount,
				IterationCount: func() *int { i := -1; return &i }(),
				Direction:      AnimationDirectionNormal,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AnimationConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnimationConfig_JSON(t *testing.T) {
	config := &AnimationConfig{
		Type:      AnimationTypeFade,
		Duration:  1000,
		Delay:     100,
		Easing:    AnimationEasingEaseInOut,
		Iteration: AnimationIterationOnce,
		Direction: AnimationDirectionNormal,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal AnimationConfig: %v", err)
	}

	var decoded AnimationConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AnimationConfig: %v", err)
	}

	if decoded.Type != config.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, config.Type)
	}
	if decoded.Duration != config.Duration {
		t.Errorf("Duration = %v, want %v", decoded.Duration, config.Duration)
	}
	if decoded.Delay != config.Delay {
		t.Errorf("Delay = %v, want %v", decoded.Delay, config.Delay)
	}
	if decoded.Easing != config.Easing {
		t.Errorf("Easing = %v, want %v", decoded.Easing, config.Easing)
	}
	if decoded.Iteration != config.Iteration {
		t.Errorf("Iteration = %v, want %v", decoded.Iteration, config.Iteration)
	}
	if decoded.Direction != config.Direction {
		t.Errorf("Direction = %v, want %v", decoded.Direction, config.Direction)
	}
}

func TestLayoutMode_IsValid(t *testing.T) {
	tests := []struct {
		name string
		mode LayoutMode
		want bool
	}{
		{"valid flexbox", LayoutModeFlexbox, true},
		{"valid grid", LayoutModeGrid, true},
		{"valid absolute", LayoutModeAbsolute, true},
		{"invalid empty", LayoutMode(""), false},
		{"invalid unknown", LayoutMode("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.want {
				t.Errorf("LayoutMode.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGridConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *GridConfig
		wantErr bool
	}{
		{
			name: "valid grid config",
			config: &GridConfig{
				Columns:  "3",
				Rows:     "auto",
				Gap:      "20px",
				AutoFlow: GridAutoFlowRow,
			},
			wantErr: false,
		},
		{
			name: "valid with template",
			config: &GridConfig{
				Columns:  "1fr 2fr 1fr",
				Rows:     "100px auto 100px",
				Gap:      "10px 20px",
				AutoFlow: GridAutoFlowColumn,
			},
			wantErr: false,
		},
		{
			name: "invalid auto flow",
			config: &GridConfig{
				Columns:  "3",
				Rows:     "auto",
				Gap:      "20px",
				AutoFlow: GridAutoFlow("invalid"),
			},
			wantErr: true,
		},
		{
			name: "empty columns",
			config: &GridConfig{
				Columns:  "",
				Rows:     "auto",
				Gap:      "20px",
				AutoFlow: GridAutoFlowRow,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GridConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlexConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *FlexConfig
		wantErr bool
	}{
		{
			name: "valid flex config",
			config: &FlexConfig{
				Direction:      FlexDirectionRow,
				Wrap:           FlexWrapNoWrap,
				JustifyContent: JustifyContentCenter,
				AlignItems:     AlignItemsCenter,
				Gap:            "20px",
			},
			wantErr: false,
		},
		{
			name: "invalid direction",
			config: &FlexConfig{
				Direction:      FlexDirection("invalid"),
				Wrap:           FlexWrapNoWrap,
				JustifyContent: JustifyContentCenter,
				AlignItems:     AlignItemsCenter,
				Gap:            "20px",
			},
			wantErr: true,
		},
		{
			name: "invalid wrap",
			config: &FlexConfig{
				Direction:      FlexDirectionRow,
				Wrap:           FlexWrap("invalid"),
				JustifyContent: JustifyContentCenter,
				AlignItems:     AlignItemsCenter,
				Gap:            "20px",
			},
			wantErr: true,
		},
		{
			name: "invalid justify content",
			config: &FlexConfig{
				Direction:      FlexDirectionRow,
				Wrap:           FlexWrapNoWrap,
				JustifyContent: JustifyContent("invalid"),
				AlignItems:     AlignItemsCenter,
				Gap:            "20px",
			},
			wantErr: true,
		},
		{
			name: "invalid align items",
			config: &FlexConfig{
				Direction:      FlexDirectionRow,
				Wrap:           FlexWrapNoWrap,
				JustifyContent: JustifyContentCenter,
				AlignItems:     AlignItems("invalid"),
				Gap:            "20px",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageEffects_Validate(t *testing.T) {
	tests := []struct {
		name    string
		effects *ImageEffects
		wantErr bool
	}{
		{
			name: "valid image effects",
			effects: &ImageEffects{
				Filter:    func() *string { s := "blur(5px)"; return &s }(),
				Transform: func() *string { s := "rotate(45deg)"; return &s }(),
				BlendMode: func() *string { s := "multiply"; return &s }(),
				Mask:      func() *string { s := "/images/mask.png"; return &s }(),
				ClipPath:  func() *string { s := "circle(50%)"; return &s }(),
			},
			wantErr: false,
		},
		{
			name: "valid minimal",
			effects: &ImageEffects{
				Filter: func() *string { s := "grayscale(100%)"; return &s }(),
			},
			wantErr: false,
		},
		{
			name:    "valid empty",
			effects: &ImageEffects{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.effects.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageEffects.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTextEffects_Validate(t *testing.T) {
	tests := []struct {
		name    string
		effects *TextEffects
		wantErr bool
	}{
		{
			name: "valid text effects",
			effects: &TextEffects{
				Gradient:      func() *string { s := "linear-gradient(90deg, #ff0000, #00ff00)"; return &s }(),
				Stroke:        func() *string { s := "2px #000000"; return &s }(),
				Shadow:        func() *string { s := "2px 2px 4px rgba(0,0,0,0.5)"; return &s }(),
				Transform:     func() *string { s := "uppercase"; return &s }(),
				LetterSpacing: func() *string { s := "0.1em"; return &s }(),
				LineHeight:    func() *string { s := "1.5"; return &s }(),
				WordSpacing:   func() *string { s := "0.2em"; return &s }(),
			},
			wantErr: false,
		},
		{
			name: "valid minimal",
			effects: &TextEffects{
				Transform: func() *string { s := "capitalize"; return &s }(),
			},
			wantErr: false,
		},
		{
			name:    "valid empty",
			effects: &TextEffects{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.effects.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TextEffects.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVisibilityRules_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rules   *VisibilityRules
		wantErr bool
	}{
		{
			name: "valid all visible",
			rules: &VisibilityRules{
				ShowOnMobile:  true,
				ShowOnTablet:  true,
				ShowOnDesktop: true,
			},
			wantErr: false,
		},
		{
			name: "valid mobile only",
			rules: &VisibilityRules{
				ShowOnMobile:  true,
				ShowOnTablet:  false,
				ShowOnDesktop: false,
			},
			wantErr: false,
		},
		{
			name: "valid with expression",
			rules: &VisibilityRules{
				ShowOnMobile:  true,
				ShowOnTablet:  true,
				ShowOnDesktop: true,
				ShowWhen:      func() *string { s := "{{.Event.HasImage}}"; return &s }(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("VisibilityRules.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComponent_WithAnimation(t *testing.T) {
	component := &Component{
		ID:         "test-component",
		Type:       ComponentTypeTextBox,
		Position:   Position{Mode: PositionModeAbsolute},
		Dimensions: Dimensions{Width: "100%", Height: "auto"},
		ZIndex:     10,
		Visible:    true,
	}

	animation := &AnimationConfig{
		Type:      AnimationTypeFade,
		Duration:  1000,
		Easing:    AnimationEasingEaseInOut,
		Iteration: AnimationIterationOnce,
		Direction: AnimationDirectionNormal,
	}

	component.Animation = animation

	if component.Animation == nil {
		t.Error("Animation should not be nil")
	}
	if component.Animation.Type != AnimationTypeFade {
		t.Errorf("Animation type = %v, want %v", component.Animation.Type, AnimationTypeFade)
	}
}
