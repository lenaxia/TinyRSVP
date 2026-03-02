package models

import "fmt"

type AnimationType string

const (
	AnimationTypeFade   AnimationType = "fade"
	AnimationTypeSlide  AnimationType = "slide"
	AnimationTypeScale  AnimationType = "scale"
	AnimationTypeRotate AnimationType = "rotate"
	AnimationTypeBounce AnimationType = "bounce"
)

func (at AnimationType) IsValid() bool {
	switch at {
	case AnimationTypeFade, AnimationTypeSlide, AnimationTypeScale,
		AnimationTypeRotate, AnimationTypeBounce:
		return true
	default:
		return false
	}
}

func (at AnimationType) String() string {
	return string(at)
}

type AnimationEasing string

const (
	AnimationEasingLinear    AnimationEasing = "linear"
	AnimationEasingEaseIn    AnimationEasing = "ease-in"
	AnimationEasingEaseOut   AnimationEasing = "ease-out"
	AnimationEasingEaseInOut AnimationEasing = "ease-in-out"
)

func (ae AnimationEasing) IsValid() bool {
	switch ae {
	case AnimationEasingLinear, AnimationEasingEaseIn,
		AnimationEasingEaseOut, AnimationEasingEaseInOut:
		return true
	default:
		return false
	}
}

func (ae AnimationEasing) String() string {
	return string(ae)
}

type AnimationIteration string

const (
	AnimationIterationOnce     AnimationIteration = "once"
	AnimationIterationInfinite AnimationIteration = "infinite"
	AnimationIterationCount    AnimationIteration = "count"
)

func (ai AnimationIteration) IsValid() bool {
	switch ai {
	case AnimationIterationOnce, AnimationIterationInfinite, AnimationIterationCount:
		return true
	default:
		return false
	}
}

func (ai AnimationIteration) String() string {
	return string(ai)
}

type AnimationDirection string

const (
	AnimationDirectionNormal    AnimationDirection = "normal"
	AnimationDirectionReverse   AnimationDirection = "reverse"
	AnimationDirectionAlternate AnimationDirection = "alternate"
)

func (ad AnimationDirection) IsValid() bool {
	switch ad {
	case AnimationDirectionNormal, AnimationDirectionReverse, AnimationDirectionAlternate:
		return true
	default:
		return false
	}
}

func (ad AnimationDirection) String() string {
	return string(ad)
}

type AnimationConfig struct {
	Type           AnimationType      `json:"type"`
	Duration       int                `json:"duration"`
	Delay          int                `json:"delay"`
	Easing         AnimationEasing    `json:"easing"`
	Iteration      AnimationIteration `json:"iteration"`
	IterationCount *int               `json:"iterationCount,omitempty"`
	Direction      AnimationDirection `json:"direction"`
}

func (ac *AnimationConfig) Validate() error {
	if !ac.Type.IsValid() {
		return fmt.Errorf("invalid animation type: %s", ac.Type)
	}
	if ac.Duration < 0 {
		return fmt.Errorf("duration must be non-negative, got %d", ac.Duration)
	}
	if ac.Delay < 0 {
		return fmt.Errorf("delay must be non-negative, got %d", ac.Delay)
	}
	if !ac.Easing.IsValid() {
		return fmt.Errorf("invalid animation easing: %s", ac.Easing)
	}
	if !ac.Iteration.IsValid() {
		return fmt.Errorf("invalid animation iteration: %s", ac.Iteration)
	}
	if ac.Iteration == AnimationIterationCount {
		if ac.IterationCount == nil {
			return fmt.Errorf("iteration count required when iteration is 'count'")
		}
		if *ac.IterationCount <= 0 {
			return fmt.Errorf("iteration count must be positive, got %d", *ac.IterationCount)
		}
	}
	if !ac.Direction.IsValid() {
		return fmt.Errorf("invalid animation direction: %s", ac.Direction)
	}
	return nil
}

type LayoutMode string

const (
	LayoutModeFlexbox  LayoutMode = "flexbox"
	LayoutModeGrid     LayoutMode = "grid"
	LayoutModeAbsolute LayoutMode = "absolute"
)

func (lm LayoutMode) IsValid() bool {
	switch lm {
	case LayoutModeFlexbox, LayoutModeGrid, LayoutModeAbsolute:
		return true
	default:
		return false
	}
}

func (lm LayoutMode) String() string {
	return string(lm)
}

type GridAutoFlow string

const (
	GridAutoFlowRow    GridAutoFlow = "row"
	GridAutoFlowColumn GridAutoFlow = "column"
	GridAutoFlowDense  GridAutoFlow = "dense"
)

func (gaf GridAutoFlow) IsValid() bool {
	switch gaf {
	case GridAutoFlowRow, GridAutoFlowColumn, GridAutoFlowDense:
		return true
	default:
		return false
	}
}

func (gaf GridAutoFlow) String() string {
	return string(gaf)
}

type GridConfig struct {
	Columns  string       `json:"columns"`
	Rows     string       `json:"rows,omitempty"`
	Gap      string       `json:"gap,omitempty"`
	AutoFlow GridAutoFlow `json:"autoFlow"`
}

func (gc *GridConfig) Validate() error {
	if gc.Columns == "" {
		return fmt.Errorf("columns is required")
	}
	if !gc.AutoFlow.IsValid() {
		return fmt.Errorf("invalid grid auto flow: %s", gc.AutoFlow)
	}
	return nil
}

type FlexDirection string

const (
	FlexDirectionRow           FlexDirection = "row"
	FlexDirectionColumn        FlexDirection = "column"
	FlexDirectionRowReverse    FlexDirection = "row-reverse"
	FlexDirectionColumnReverse FlexDirection = "column-reverse"
)

func (fd FlexDirection) IsValid() bool {
	switch fd {
	case FlexDirectionRow, FlexDirectionColumn,
		FlexDirectionRowReverse, FlexDirectionColumnReverse:
		return true
	default:
		return false
	}
}

func (fd FlexDirection) String() string {
	return string(fd)
}

type FlexWrap string

const (
	FlexWrapNoWrap      FlexWrap = "nowrap"
	FlexWrapWrap        FlexWrap = "wrap"
	FlexWrapWrapReverse FlexWrap = "wrap-reverse"
)

func (fw FlexWrap) IsValid() bool {
	switch fw {
	case FlexWrapNoWrap, FlexWrapWrap, FlexWrapWrapReverse:
		return true
	default:
		return false
	}
}

func (fw FlexWrap) String() string {
	return string(fw)
}

type JustifyContent string

const (
	JustifyContentFlexStart    JustifyContent = "flex-start"
	JustifyContentCenter       JustifyContent = "center"
	JustifyContentFlexEnd      JustifyContent = "flex-end"
	JustifyContentSpaceBetween JustifyContent = "space-between"
	JustifyContentSpaceAround  JustifyContent = "space-around"
	JustifyContentSpaceEvenly  JustifyContent = "space-evenly"
)

func (jc JustifyContent) IsValid() bool {
	switch jc {
	case JustifyContentFlexStart, JustifyContentCenter, JustifyContentFlexEnd,
		JustifyContentSpaceBetween, JustifyContentSpaceAround, JustifyContentSpaceEvenly:
		return true
	default:
		return false
	}
}

func (jc JustifyContent) String() string {
	return string(jc)
}

type AlignItems string

const (
	AlignItemsFlexStart AlignItems = "flex-start"
	AlignItemsCenter    AlignItems = "center"
	AlignItemsFlexEnd   AlignItems = "flex-end"
	AlignItemsStretch   AlignItems = "stretch"
	AlignItemsBaseline  AlignItems = "baseline"
)

func (ai AlignItems) IsValid() bool {
	switch ai {
	case AlignItemsFlexStart, AlignItemsCenter, AlignItemsFlexEnd,
		AlignItemsStretch, AlignItemsBaseline:
		return true
	default:
		return false
	}
}

func (ai AlignItems) String() string {
	return string(ai)
}

type FlexConfig struct {
	Direction      FlexDirection  `json:"direction"`
	Wrap           FlexWrap       `json:"wrap"`
	JustifyContent JustifyContent `json:"justifyContent"`
	AlignItems     AlignItems     `json:"alignItems"`
	Gap            string         `json:"gap,omitempty"`
}

func (fc *FlexConfig) Validate() error {
	if !fc.Direction.IsValid() {
		return fmt.Errorf("invalid flex direction: %s", fc.Direction)
	}
	if !fc.Wrap.IsValid() {
		return fmt.Errorf("invalid flex wrap: %s", fc.Wrap)
	}
	if !fc.JustifyContent.IsValid() {
		return fmt.Errorf("invalid justify content: %s", fc.JustifyContent)
	}
	if !fc.AlignItems.IsValid() {
		return fmt.Errorf("invalid align items: %s", fc.AlignItems)
	}
	return nil
}

type ImageEffects struct {
	Filter    *string `json:"filter,omitempty"`
	Transform *string `json:"transform,omitempty"`
	BlendMode *string `json:"blendMode,omitempty"`
	Mask      *string `json:"mask,omitempty"`
	ClipPath  *string `json:"clipPath,omitempty"`
}

func (ie *ImageEffects) Validate() error {
	return nil
}

type TextEffects struct {
	Gradient      *string `json:"gradient,omitempty"`
	Stroke        *string `json:"stroke,omitempty"`
	Shadow        *string `json:"shadow,omitempty"`
	Transform     *string `json:"transform,omitempty"`
	LetterSpacing *string `json:"letterSpacing,omitempty"`
	LineHeight    *string `json:"lineHeight,omitempty"`
	WordSpacing   *string `json:"wordSpacing,omitempty"`
}

func (te *TextEffects) Validate() error {
	return nil
}

type VisibilityRules struct {
	ShowOnMobile  bool    `json:"showOnMobile"`
	ShowOnTablet  bool    `json:"showOnTablet"`
	ShowOnDesktop bool    `json:"showOnDesktop"`
	ShowWhen      *string `json:"showWhen,omitempty"`
}

func (vr *VisibilityRules) Validate() error {
	return nil
}
