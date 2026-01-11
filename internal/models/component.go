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

func (p *Position) Validate() error {
	if !p.Mode.IsValid() {
		return &ValidationError{Field: "mode", Message: "Invalid position mode"}
	}

	switch p.Mode {
	case PositionModeAbsolute:
		if p.X == nil {
			return &ValidationError{Field: "x", Message: "X coordinate required for absolute positioning"}
		}
		if p.Y == nil {
			return &ValidationError{Field: "y", Message: "Y coordinate required for absolute positioning"}
		}
	case PositionModeFlex, PositionModeRelative:
		if p.Order == nil {
			return &ValidationError{Field: "order", Message: "Order required for flex/relative positioning"}
		}
	}

	return nil
}

type Dimensions struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

func (d *Dimensions) Validate() error {
	if d.Width == "" {
		return &ValidationError{Field: "width", Message: "Width is required"}
	}
	if d.Height == "" {
		return &ValidationError{Field: "height", Message: "Height is required"}
	}
	return nil
}

type TextBoxContent struct {
	Text            string  `json:"text"`
	TextAlign       string  `json:"textAlign,omitempty"`
	FontFamily      string  `json:"fontFamily,omitempty"`
	FontSize        string  `json:"fontSize,omitempty"`
	FontWeight      string  `json:"fontWeight,omitempty"`
	Color           string  `json:"color,omitempty"`
	LineHeight      string  `json:"lineHeight,omitempty"`
	LetterSpacing   string  `json:"letterSpacing,omitempty"`
	TextTransform   string  `json:"textTransform,omitempty"`
	TextShadow      string  `json:"textShadow,omitempty"`
	BackgroundColor string  `json:"backgroundColor,omitempty"`
	Padding         string  `json:"padding,omitempty"`
	BorderRadius    string  `json:"borderRadius,omitempty"`
	MaxWidth        string  `json:"maxWidth,omitempty"`
	FontStyle       string  `json:"fontStyle,omitempty"`
}

func (tbc *TextBoxContent) Validate() error {
	if tbc.Text == "" {
		return &ValidationError{Field: "text", Message: "Text content is required"}
	}
	if len(tbc.Text) > 10000 {
		return &ValidationError{Field: "text", Message: "Text content cannot exceed 10000 characters"}
	}
	return nil
}

type ImageContent struct {
	Src            string   `json:"src"`
	Alt            string   `json:"alt,omitempty"`
	ObjectFit      string   `json:"objectFit,omitempty"`
	ObjectPosition string   `json:"objectPosition,omitempty"`
	Opacity        *float64 `json:"opacity,omitempty"`
	Filter         string   `json:"filter,omitempty"`
}

func (ic *ImageContent) Validate() error {
	if ic.Src == "" {
		return &ValidationError{Field: "src", Message: "Image source is required"}
	}
	if ic.Opacity != nil {
		if *ic.Opacity < 0 || *ic.Opacity > 1 {
			return &ValidationError{Field: "opacity", Message: "Opacity must be between 0 and 1"}
		}
	}
	return nil
}

type BackgroundType string

const (
	BackgroundTypeColor    BackgroundType = "color"
	BackgroundTypeGradient BackgroundType = "gradient"
	BackgroundTypeImage    BackgroundType = "image"
)

type BackgroundImageConfig struct {
	Src      string `json:"src"`
	Repeat   string `json:"repeat,omitempty"`
	Size     string `json:"size,omitempty"`
	Position string `json:"position,omitempty"`
}

type BackgroundContent struct {
	Type     BackgroundType         `json:"type"`
	Color    string                 `json:"color,omitempty"`
	Gradient string                 `json:"gradient,omitempty"`
	Image    *BackgroundImageConfig `json:"image,omitempty"`
}

func (bc *BackgroundContent) Validate() error {
	switch bc.Type {
	case BackgroundTypeColor:
		if bc.Color == "" {
			return &ValidationError{Field: "color", Message: "Color required for color background"}
		}
	case BackgroundTypeGradient:
		if bc.Gradient == "" {
			return &ValidationError{Field: "gradient", Message: "Gradient required for gradient background"}
		}
	case BackgroundTypeImage:
		if bc.Image == nil || bc.Image.Src == "" {
			return &ValidationError{Field: "image", Message: "Image configuration required for image background"}
		}
	default:
		return &ValidationError{Field: "type", Message: "Invalid background type"}
	}
	return nil
}

type OverlayPlaceholder struct {
	Show bool   `json:"show"`
	Text string `json:"text,omitempty"`
	Icon string `json:"icon,omitempty"`
}

type OverlayContent struct {
	BackgroundColor  string              `json:"backgroundColor,omitempty"`
	BackgroundImage  string              `json:"backgroundImage,omitempty"`
	BackgroundSize   string              `json:"backgroundSize,omitempty"`
	BackgroundPosition string            `json:"backgroundPosition,omitempty"`
	BorderRadius     string              `json:"borderRadius,omitempty"`
	Border           string              `json:"border,omitempty"`
	BoxShadow        string              `json:"boxShadow,omitempty"`
	ClipPath         string              `json:"clipPath,omitempty"`
	Placeholder      *OverlayPlaceholder `json:"placeholder,omitempty"`
}

func (oc *OverlayContent) Validate() error {
	return nil
}

type LayoutConfig struct {
	Display        string `json:"display,omitempty"`
	FlexDirection  string `json:"flexDirection,omitempty"`
	AlignItems     string `json:"alignItems,omitempty"`
	JustifyContent string `json:"justifyContent,omitempty"`
	Gap            string `json:"gap,omitempty"`
	Padding        string `json:"padding,omitempty"`
}

type ContainerStyle struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	BorderRadius    string `json:"borderRadius,omitempty"`
	BoxShadow       string `json:"boxShadow,omitempty"`
	Border          string `json:"border,omitempty"`
	Margin          string `json:"margin,omitempty"`
}

type ContainerContent struct {
	Layout   LayoutConfig   `json:"layout"`
	Style    ContainerStyle `json:"style,omitempty"`
	Children []string       `json:"children,omitempty"`
}

func (cc *ContainerContent) Validate() error {
	return nil
}

type DividerStyle struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Margin          string `json:"margin,omitempty"`
	BorderRadius    string `json:"borderRadius,omitempty"`
	Border          string `json:"border,omitempty"`
}

type DividerContent struct {
	Style DividerStyle `json:"style"`
}

func (dc *DividerContent) Validate() error {
	return nil
}

type ResponsiveConfig struct {
	Mobile  map[string]interface{} `json:"mobile,omitempty"`
	Tablet  map[string]interface{} `json:"tablet,omitempty"`
	Desktop map[string]interface{} `json:"desktop,omitempty"`
}

type Component struct {
	ID         string                 `json:"id"`
	Type       ComponentType          `json:"type"`
	Position   Position               `json:"position"`
	Dimensions Dimensions             `json:"dimensions"`
	ZIndex     int                    `json:"zIndex"`
	Visible    bool                   `json:"visible"`
	ClassName  string                 `json:"className,omitempty"`
	Content    map[string]interface{} `json:"content"`
	Responsive *ResponsiveConfig      `json:"responsive,omitempty"`
}

func (c *Component) Validate() error {
	if c.ID == "" {
		return &ValidationError{Field: "id", Message: "Component ID is required"}
	}

	if !c.Type.IsValid() {
		return &ValidationError{Field: "type", Message: "Invalid component type"}
	}

	if err := c.Position.Validate(); err != nil {
		return err
	}

	if err := c.Dimensions.Validate(); err != nil {
		return err
	}

	if c.ZIndex < 0 {
		return &ValidationError{Field: "zIndex", Message: "zIndex must be >= 0"}
	}

	return nil
}

func (c *Component) GetTextBoxContent() (*TextBoxContent, error) {
	if c.Type != ComponentTypeTextBox {
		return nil, fmt.Errorf("component is not a TextBox")
	}

	jsonData, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var content TextBoxContent
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

func (c *Component) GetImageContent() (*ImageContent, error) {
	if c.Type != ComponentTypeImage {
		return nil, fmt.Errorf("component is not an Image")
	}

	jsonData, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var content ImageContent
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

func (c *Component) GetBackgroundContent() (*BackgroundContent, error) {
	if c.Type != ComponentTypeBackground {
		return nil, fmt.Errorf("component is not a Background")
	}

	jsonData, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var content BackgroundContent
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

func (c *Component) GetOverlayContent() (*OverlayContent, error) {
	if c.Type != ComponentTypeOverlay {
		return nil, fmt.Errorf("component is not an Overlay")
	}

	jsonData, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var content OverlayContent
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

func (c *Component) GetContainerContent() (*ContainerContent, error) {
	if c.Type != ComponentTypeContainer {
		return nil, fmt.Errorf("component is not a Container")
	}

	jsonData, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var content ContainerContent
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

func (c *Component) GetDividerContent() (*DividerContent, error) {
	if c.Type != ComponentTypeDivider {
		return nil, fmt.Errorf("component is not a Divider")
	}

	jsonData, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var content DividerContent
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return nil, err
	}

	return &content, nil
}

type ConfigMetadata struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
}

type PageLayoutConfig struct {
	Mode           string `json:"mode"`
	CardWidth      string `json:"cardWidth,omitempty"`
	CardMaxWidth   string `json:"cardMaxWidth,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	MinHeight      string `json:"minHeight,omitempty"`
	Padding        string `json:"padding,omitempty"`
}

type ComponentConfiguration struct {
	Version    string           `json:"version"`
	Metadata   ConfigMetadata   `json:"metadata"`
	Layout     PageLayoutConfig `json:"layout"`
	Components []Component      `json:"components"`
}

func (cc *ComponentConfiguration) Validate() error {
	if cc.Version == "" {
		return &ValidationError{Field: "version", Message: "Version is required"}
	}

	if cc.Metadata.Name == "" {
		return &ValidationError{Field: "metadata.name", Message: "Metadata name is required"}
	}

	componentIDs := make(map[string]bool)
	for i, comp := range cc.Components {
		if err := comp.Validate(); err != nil {
			return fmt.Errorf("component[%d]: %w", i, err)
		}

		if componentIDs[comp.ID] {
			return &ValidationError{
				Field:   "components",
				Message: fmt.Sprintf("Duplicate component ID: %s", comp.ID),
			}
		}
		componentIDs[comp.ID] = true
	}

	return nil
}

type ComponentOverride struct {
	ID      string                 `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}

type ComponentOverrides struct {
	Version   string              `json:"version"`
	Overrides []ComponentOverride `json:"overrides,omitempty"`
	Additions []Component         `json:"additions,omitempty"`
	Removals  []string            `json:"removals,omitempty"`
}

func (co *ComponentOverrides) Validate() error {
	if co.Version == "" {
		return &ValidationError{Field: "version", Message: "Version is required"}
	}

	overrideIDs := make(map[string]bool)
	for i, override := range co.Overrides {
		if override.ID == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("overrides[%d].id", i),
				Message: "Override ID is required",
			}
		}

		if overrideIDs[override.ID] {
			return &ValidationError{
				Field:   "overrides",
				Message: fmt.Sprintf("Duplicate override ID: %s", override.ID),
			}
		}
		overrideIDs[override.ID] = true
	}

	for i, addition := range co.Additions {
		if err := addition.Validate(); err != nil {
			return fmt.Errorf("additions[%d]: %w", i, err)
		}
	}

	return nil
}
