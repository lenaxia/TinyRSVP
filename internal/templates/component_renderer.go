package templates

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type ComponentRenderer struct {
	engine *Engine
}

func NewComponentRenderer(engine *Engine) *ComponentRenderer {
	return &ComponentRenderer{
		engine: engine,
	}
}

func (r *ComponentRenderer) ParseComponentConfig(jsonStr *string) (*models.ComponentConfiguration, error) {
	if jsonStr == nil || *jsonStr == "" {
		return nil, nil
	}

	var config models.ComponentConfiguration
	if err := json.Unmarshal([]byte(*jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse component configuration: %w", err)
	}

	return &config, nil
}

func (r *ComponentRenderer) ParseComponentOverrides(jsonStr *string) (*models.ComponentOverrides, error) {
	if jsonStr == nil || *jsonStr == "" {
		return nil, nil
	}

	var overrides models.ComponentOverrides
	if err := json.Unmarshal([]byte(*jsonStr), &overrides); err != nil {
		return nil, fmt.Errorf("failed to parse component overrides: %w", err)
	}

	return &overrides, nil
}

func (r *ComponentRenderer) MergeConfigurations(base *models.ComponentConfiguration, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error) {
	if base == nil {
		return nil, fmt.Errorf("base configuration cannot be nil")
	}

	result := &models.ComponentConfiguration{
		Version:    base.Version,
		Metadata:   base.Metadata,
		Layout:     base.Layout,
		Components: make([]models.Component, len(base.Components)),
	}

	copy(result.Components, base.Components)

	if overrides == nil {
		sort.Slice(result.Components, func(i, j int) bool {
			return result.Components[i].ZIndex < result.Components[j].ZIndex
		})
		return result, nil
	}

	componentMap := make(map[string]*models.Component)
	for i := range result.Components {
		componentMap[result.Components[i].ID] = &result.Components[i]
	}

	for _, override := range overrides.Overrides {
		if comp, exists := componentMap[override.ID]; exists {
			if err := r.applyOverride(comp, override.Updates); err != nil {
				return nil, fmt.Errorf("failed to apply override for component %s: %w", override.ID, err)
			}
		}
	}

	for _, addition := range overrides.Additions {
		result.Components = append(result.Components, addition)
	}

	removalSet := make(map[string]bool)
	for _, id := range overrides.Removals {
		removalSet[id] = true
	}

	filtered := make([]models.Component, 0, len(result.Components))
	for _, comp := range result.Components {
		if !removalSet[comp.ID] {
			filtered = append(filtered, comp)
		}
	}
	result.Components = filtered

	sort.Slice(result.Components, func(i, j int) bool {
		return result.Components[i].ZIndex < result.Components[j].ZIndex
	})

	return result, nil
}

func (r *ComponentRenderer) applyOverride(comp *models.Component, updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "position":
			if posMap, ok := value.(map[string]interface{}); ok {
				r.deepMergePosition(&comp.Position, posMap)
			}
		case "dimensions":
			if dimMap, ok := value.(map[string]interface{}); ok {
				r.deepMergeDimensions(&comp.Dimensions, dimMap)
			}
		case "content":
			if contentMap, ok := value.(map[string]interface{}); ok {
				if comp.Content == nil {
					comp.Content = make(map[string]interface{})
				}
				r.deepMergeMap(comp.Content, contentMap)
			}
		case "layout":
			if layoutMap, ok := value.(map[string]interface{}); ok {
				if comp.Layout == nil {
					comp.Layout = make(map[string]interface{})
				}
				r.deepMergeMap(comp.Layout, layoutMap)
			}
		case "style":
			if styleMap, ok := value.(map[string]interface{}); ok {
				if comp.Style == nil {
					comp.Style = make(map[string]interface{})
				}
				r.deepMergeMap(comp.Style, styleMap)
			}
		case "responsive":
			if respMap, ok := value.(map[string]interface{}); ok {
				if comp.Responsive == nil {
					comp.Responsive = make(map[string]interface{})
				}
				r.deepMergeMap(comp.Responsive, respMap)
			}
		case "zIndex":
			if zIndex, ok := value.(float64); ok {
				comp.ZIndex = int(zIndex)
			}
		case "visible":
			if visible, ok := value.(bool); ok {
				comp.Visible = visible
			}
		case "className":
			if className, ok := value.(string); ok {
				comp.ClassName = &className
			}
		case "children":
			if children, ok := value.([]interface{}); ok {
				comp.Children = make([]string, 0, len(children))
				for _, child := range children {
					if childStr, ok := child.(string); ok {
						comp.Children = append(comp.Children, childStr)
					}
				}
			}
		}
	}
	return nil
}

func (r *ComponentRenderer) deepMergePosition(pos *models.Position, updates map[string]interface{}) {
	if mode, ok := updates["mode"].(string); ok {
		pos.Mode = models.PositionMode(mode)
	}
	if x, ok := updates["x"].(string); ok {
		pos.X = &x
	}
	if y, ok := updates["y"].(string); ok {
		pos.Y = &y
	}
	if order, ok := updates["order"].(float64); ok {
		orderInt := int(order)
		pos.Order = &orderInt
	}
}

func (r *ComponentRenderer) deepMergeDimensions(dim *models.Dimensions, updates map[string]interface{}) {
	if width, ok := updates["width"].(string); ok {
		dim.Width = width
	}
	if height, ok := updates["height"].(string); ok {
		dim.Height = height
	}
}

func (r *ComponentRenderer) deepMergeMap(target map[string]interface{}, source map[string]interface{}) {
	for key, value := range source {
		if existingValue, exists := target[key]; exists {
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if sourceMap, ok := value.(map[string]interface{}); ok {
					r.deepMergeMap(existingMap, sourceMap)
					continue
				}
			}
		}
		target[key] = value
	}
}

func (r *ComponentRenderer) Render(w io.Writer, event *models.Event, template *models.Template) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	config, err := r.ParseComponentConfig(template.ComponentConfig)
	if err != nil {
		return fmt.Errorf("failed to parse template component config: %w", err)
	}

	if config == nil {
		return fmt.Errorf("template has no component configuration")
	}

	overrides, err := r.ParseComponentOverrides(event.ComponentOverrides)
	if err != nil {
		return fmt.Errorf("failed to parse event component overrides: %w", err)
	}

	finalConfig, err := r.MergeConfigurations(config, overrides)
	if err != nil {
		return fmt.Errorf("failed to merge configurations: %w", err)
	}

	data := map[string]interface{}{
		"Event":         event,
		"Template":      template,
		"Configuration": finalConfig,
		"Components":    finalConfig.Components,
	}

	if r.engine == nil {
		return fmt.Errorf("template engine not initialized")
	}

	tmpl, err := r.engine.Parse("{{/* Component rendering will be implemented with HTML templates */}}")
	if err != nil {
		return fmt.Errorf("failed to parse component template: %w", err)
	}

	return r.engine.Execute(w, tmpl, data)
}
