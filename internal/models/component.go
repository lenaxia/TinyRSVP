package models

import (
	"encoding/json"
	"fmt"
)

type ComponentType string

const (
	ComponentTypeTextBox    ComponentType = "TextBox"
	ComponentTypeImage      ComponentType = "Image"
	ComponentTypeBackground ComponentType = "Background"
	ComponentTypeOverlay    ComponentType = "Overlay"
	ComponentTypeContainer  ComponentType = "Container"
	ComponentTypeDivider    ComponentType = "Divider"
)

func (ct ComponentType) IsValid() bool {
	switch ct {
	case ComponentTypeTextBox, ComponentTypeImage, ComponentTypeBackground,
		ComponentTypeOverlay, ComponentTypeContainer, ComponentTypeDivider:
		return true
	default:
		return false
	}
}

func (ct ComponentType) String() string {
	return string(ct)
}

type PositionMode string

const (
	PositionModeAbsolute PositionMode = "absolute"
	PositionModeRelative PositionMode = "relative"
	PositionModeFlex     PositionMode = "flex"
)

func (pm PositionMode) IsValid() bool {
	switch pm {
	case PositionModeAbsolute, PositionModeRelative, PositionModeFlex:
		return true
	default:
		return false
	}
}

func (pm PositionMode) String() string {
	return string(pm)
}

type Position struct {
	Mode  PositionMode `json:"mode"`
	X     *string      `json:"x,omitempty"`
	Y     *string      `json:"y,omitempty"`
	Order *int         `json:"order,omitempty"`
}

type Dimensions struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

type TextBoxContent struct {
	Text          string `json:"text"`
	FontFamily    string `json:"fontFamily,omitempty"`
	FontSize      string `json:"fontSize,omitempty"`
	FontWeight    string `json:"fontWeight,omitempty"`
	Color         string `json:"color,omitempty"`
	TextAlign     string `json:"textAlign,omitempty"`
	LineHeight    string `json:"lineHeight,omitempty"`
	LetterSpacing string `json:"letterSpacing,omitempty"`
	TextTransform string `json:"textTransform,omitempty"`
	Padding       string `json:"padding,omitempty"`
	TextShadow    string `json:"textShadow,omitempty"`
	EvaluatedText string `json:"evaluatedText,omitempty"`
	IsEvaluated   bool   `json:"isEvaluated,omitempty"`
}

type ImageContent struct {
	Src            string   `json:"src"`
	Alt            string   `json:"alt,omitempty"`
	ObjectFit      string   `json:"objectFit,omitempty"`
	ObjectPosition string   `json:"objectPosition,omitempty"`
	Opacity        *float64 `json:"opacity,omitempty"`
	Filter         string   `json:"filter,omitempty"`
}

type BackgroundContent struct {
	Type               string `json:"type"`
	Color              string `json:"color,omitempty"`
	Gradient           string `json:"gradient,omitempty"`
	Image              string `json:"image,omitempty"`
	BackgroundSize     string `json:"backgroundSize,omitempty"`
	BackgroundPosition string `json:"backgroundPosition,omitempty"`
}

type OverlayContent struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	BackgroundImage string `json:"backgroundImage,omitempty"`
	BackgroundSize  string `json:"backgroundSize,omitempty"`
	BorderRadius    string `json:"borderRadius,omitempty"`
	Border          string `json:"border,omitempty"`
	BoxShadow       string `json:"boxShadow,omitempty"`
	ClipPath        string `json:"clipPath,omitempty"`
	Placeholder     bool   `json:"placeholder,omitempty"`
}

type ContainerLayout struct {
	Display        string   `json:"display,omitempty"`
	FlexDirection  string   `json:"flexDirection,omitempty"`
	JustifyContent string   `json:"justifyContent,omitempty"`
	AlignItems     string   `json:"alignItems,omitempty"`
	Gap            string   `json:"gap,omitempty"`
	Padding        string   `json:"padding,omitempty"`
	Children       []string `json:"children,omitempty"`
}

type DividerStyle struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Height          string `json:"height,omitempty"`
	Margin          string `json:"margin,omitempty"`
	BorderRadius    string `json:"borderRadius,omitempty"`
}

type ResponsiveConfig struct {
	Mobile  *ResponsiveBreakpoint `json:"mobile,omitempty"`
	Tablet  *ResponsiveBreakpoint `json:"tablet,omitempty"`
	Desktop *ResponsiveBreakpoint `json:"desktop,omitempty"`
}

type ResponsiveBreakpoint struct {
	Width    string `json:"width,omitempty"`
	Height   string `json:"height,omitempty"`
	FontSize string `json:"fontSize,omitempty"`
	Padding  string `json:"padding,omitempty"`
	Margin   string `json:"margin,omitempty"`
	Display  string `json:"display,omitempty"`
	Visible  *bool  `json:"visible,omitempty"`
}

type ComponentContent struct {
	TextBox    *TextBoxContent    `json:"-"`
	Image      *ImageContent      `json:"-"`
	Background *BackgroundContent `json:"-"`
	Overlay    *OverlayContent    `json:"-"`
}

func (cc *ComponentContent) MarshalJSON() ([]byte, error) {
	if cc.TextBox != nil {
		return json.Marshal(cc.TextBox)
	}
	if cc.Image != nil {
		return json.Marshal(cc.Image)
	}
	if cc.Background != nil {
		return json.Marshal(cc.Background)
	}
	if cc.Overlay != nil {
		return json.Marshal(cc.Overlay)
	}
	return json.Marshal(nil)
}

func (cc *ComponentContent) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasText := raw["text"]; hasText {
		cc.TextBox = &TextBoxContent{}
		return json.Unmarshal(data, cc.TextBox)
	}

	if _, hasSrc := raw["src"]; hasSrc {
		cc.Image = &ImageContent{}
		return json.Unmarshal(data, cc.Image)
	}

	if typ, hasType := raw["type"]; hasType {
		if typStr, ok := typ.(string); ok {
			if typStr == "color" || typStr == "gradient" || typStr == "image" {
				cc.Background = &BackgroundContent{}
				return json.Unmarshal(data, cc.Background)
			}
		}
	}

	if _, hasBgColor := raw["backgroundColor"]; hasBgColor {
		cc.Overlay = &OverlayContent{}
		return json.Unmarshal(data, cc.Overlay)
	}

	return nil
}

type Component struct {
	ID         string        `json:"id"`
	Type       ComponentType `json:"type"`
	Position   Position      `json:"position"`
	Dimensions Dimensions    `json:"dimensions"`
	ZIndex     int           `json:"zIndex"`
	Visible    bool          `json:"visible"`
	ClassName  *string       `json:"className,omitempty"`
	Children   []string      `json:"children,omitempty"`

	Content    *ComponentContent `json:"content,omitempty"`
	Layout     *ContainerLayout  `json:"layout,omitempty"`
	Style      *DividerStyle     `json:"style,omitempty"`
	Responsive *ResponsiveConfig `json:"responsive,omitempty"`

	Animation  *AnimationConfig `json:"animation,omitempty"`
	Visibility *VisibilityRules `json:"visibility,omitempty"`
	LayoutMode *LayoutMode      `json:"layoutMode,omitempty"`
	GridConfig *GridConfig      `json:"gridConfig,omitempty"`
	FlexConfig *FlexConfig      `json:"flexConfig,omitempty"`
}

func (c *Component) GetTextBoxContent() (*TextBoxContent, error) {
	if c.Type != ComponentTypeTextBox {
		return nil, fmt.Errorf("component is not a TextBox")
	}
	if c.Content == nil || c.Content.TextBox == nil {
		return nil, fmt.Errorf("TextBox content is nil")
	}
	return c.Content.TextBox, nil
}

func (c *Component) GetImageContent() (*ImageContent, error) {
	if c.Type != ComponentTypeImage {
		return nil, fmt.Errorf("component is not an Image")
	}
	if c.Content == nil || c.Content.Image == nil {
		return nil, fmt.Errorf("Image content is nil")
	}
	return c.Content.Image, nil
}

func (c *Component) GetBackgroundContent() (*BackgroundContent, error) {
	if c.Type != ComponentTypeBackground {
		return nil, fmt.Errorf("component is not a Background")
	}
	if c.Content == nil || c.Content.Background == nil {
		return nil, fmt.Errorf("Background content is nil")
	}
	return c.Content.Background, nil
}

func (c *Component) GetOverlayContent() (*OverlayContent, error) {
	if c.Type != ComponentTypeOverlay {
		return nil, fmt.Errorf("component is not an Overlay")
	}
	if c.Content == nil || c.Content.Overlay == nil {
		return nil, fmt.Errorf("Overlay content is nil")
	}
	return c.Content.Overlay, nil
}

func (c *Component) GetContainerLayout() (*ContainerLayout, error) {
	if c.Type != ComponentTypeContainer {
		return nil, fmt.Errorf("component is not a Container")
	}
	if c.Layout == nil {
		return nil, fmt.Errorf("Container layout is nil")
	}
	return c.Layout, nil
}

func (c *Component) GetDividerStyle() (*DividerStyle, error) {
	if c.Type != ComponentTypeDivider {
		return nil, fmt.Errorf("component is not a Divider")
	}
	if c.Style == nil {
		return nil, fmt.Errorf("Divider style is nil")
	}
	return c.Style, nil
}

type ConfigMetadata struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type LayoutConfig struct {
	Mode            string `json:"mode"`
	CardWidth       string `json:"cardWidth,omitempty"`
	CardMaxWidth    string `json:"cardMaxWidth,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
}

type ComponentConfiguration struct {
	Version    string         `json:"version"`
	Metadata   ConfigMetadata `json:"metadata"`
	Layout     LayoutConfig   `json:"layout"`
	Components []Component    `json:"components"`
}

type ComponentOverride struct {
	ID      string                 `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}

type ComponentOverrides struct {
	Version   string              `json:"version"`
	Overrides []ComponentOverride `json:"overrides"`
	Additions []Component         `json:"additions"`
	Removals  []string            `json:"removals"`
}
