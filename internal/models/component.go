package models

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

type Component struct {
	ID         string                 `json:"id"`
	Type       ComponentType          `json:"type"`
	Position   Position               `json:"position"`
	Dimensions Dimensions             `json:"dimensions"`
	ZIndex     int                    `json:"zIndex"`
	Visible    bool                   `json:"visible"`
	ClassName  *string                `json:"className,omitempty"`
	Content    map[string]interface{} `json:"content,omitempty"`
	Layout     map[string]interface{} `json:"layout,omitempty"`
	Style      map[string]interface{} `json:"style,omitempty"`
	Children   []string               `json:"children,omitempty"`
	Responsive map[string]interface{} `json:"responsive,omitempty"`
	Animation  *AnimationConfig       `json:"animation,omitempty"`
	Visibility *VisibilityRules       `json:"visibility,omitempty"`
	LayoutMode *LayoutMode            `json:"layoutMode,omitempty"`
	GridConfig *GridConfig            `json:"gridConfig,omitempty"`
	FlexConfig *FlexConfig            `json:"flexConfig,omitempty"`
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
