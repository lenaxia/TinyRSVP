package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
				if err := r.deepMergeContent(comp, contentMap); err != nil {
					return fmt.Errorf("failed to merge content: %w", err)
				}
			}
		case "layout":
			if layoutMap, ok := value.(map[string]interface{}); ok {
				if err := r.deepMergeLayout(comp, layoutMap); err != nil {
					return fmt.Errorf("failed to merge layout: %w", err)
				}
			}
		case "style":
			if styleMap, ok := value.(map[string]interface{}); ok {
				if err := r.deepMergeStyle(comp, styleMap); err != nil {
					return fmt.Errorf("failed to merge style: %w", err)
				}
			}
		case "responsive":
			if respMap, ok := value.(map[string]interface{}); ok {
				if err := r.deepMergeResponsive(comp, respMap); err != nil {
					return fmt.Errorf("failed to merge responsive: %w", err)
				}
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

func (r *ComponentRenderer) deepMergeContent(comp *models.Component, updates map[string]interface{}) error {
	if comp.Content == nil {
		comp.Content = &models.ComponentContent{}
	}

	switch comp.Type {
	case models.ComponentTypeTextBox:
		if comp.Content.TextBox == nil {
			comp.Content.TextBox = &models.TextBoxContent{}
		}
		return r.mergeTextBoxContent(comp.Content.TextBox, updates)
	case models.ComponentTypeImage:
		if comp.Content.Image == nil {
			comp.Content.Image = &models.ImageContent{}
		}
		return r.mergeImageContent(comp.Content.Image, updates)
	case models.ComponentTypeBackground:
		if comp.Content.Background == nil {
			comp.Content.Background = &models.BackgroundContent{}
		}
		return r.mergeBackgroundContent(comp.Content.Background, updates)
	case models.ComponentTypeOverlay:
		if comp.Content.Overlay == nil {
			comp.Content.Overlay = &models.OverlayContent{}
		}
		return r.mergeOverlayContent(comp.Content.Overlay, updates)
	}
	return nil
}

func (r *ComponentRenderer) mergeTextBoxContent(content *models.TextBoxContent, updates map[string]interface{}) error {
	if text, ok := updates["text"].(string); ok {
		content.Text = text
	}
	if fontFamily, ok := updates["fontFamily"].(string); ok {
		content.FontFamily = fontFamily
	}
	if fontSize, ok := updates["fontSize"].(string); ok {
		content.FontSize = fontSize
	}
	if fontWeight, ok := updates["fontWeight"].(string); ok {
		content.FontWeight = fontWeight
	}
	if color, ok := updates["color"].(string); ok {
		content.Color = color
	}
	if textAlign, ok := updates["textAlign"].(string); ok {
		content.TextAlign = textAlign
	}
	if lineHeight, ok := updates["lineHeight"].(string); ok {
		content.LineHeight = lineHeight
	}
	if letterSpacing, ok := updates["letterSpacing"].(string); ok {
		content.LetterSpacing = letterSpacing
	}
	if textTransform, ok := updates["textTransform"].(string); ok {
		content.TextTransform = textTransform
	}
	if padding, ok := updates["padding"].(string); ok {
		content.Padding = padding
	}
	if textShadow, ok := updates["textShadow"].(string); ok {
		content.TextShadow = textShadow
	}
	return nil
}

func (r *ComponentRenderer) mergeImageContent(content *models.ImageContent, updates map[string]interface{}) error {
	if src, ok := updates["src"].(string); ok {
		content.Src = src
	}
	if alt, ok := updates["alt"].(string); ok {
		content.Alt = alt
	}
	if objectFit, ok := updates["objectFit"].(string); ok {
		content.ObjectFit = objectFit
	}
	if objectPosition, ok := updates["objectPosition"].(string); ok {
		content.ObjectPosition = objectPosition
	}
	if opacity, ok := updates["opacity"].(float64); ok {
		content.Opacity = &opacity
	}
	if filter, ok := updates["filter"].(string); ok {
		content.Filter = filter
	}
	return nil
}

func (r *ComponentRenderer) mergeBackgroundContent(content *models.BackgroundContent, updates map[string]interface{}) error {
	if typ, ok := updates["type"].(string); ok {
		content.Type = typ
	}
	if color, ok := updates["color"].(string); ok {
		content.Color = color
	}
	if gradient, ok := updates["gradient"].(string); ok {
		content.Gradient = gradient
	}
	if image, ok := updates["image"].(string); ok {
		content.Image = image
	}
	if backgroundSize, ok := updates["backgroundSize"].(string); ok {
		content.BackgroundSize = backgroundSize
	}
	if backgroundPosition, ok := updates["backgroundPosition"].(string); ok {
		content.BackgroundPosition = backgroundPosition
	}
	return nil
}

func (r *ComponentRenderer) mergeOverlayContent(content *models.OverlayContent, updates map[string]interface{}) error {
	if backgroundColor, ok := updates["backgroundColor"].(string); ok {
		content.BackgroundColor = backgroundColor
	}
	if backgroundImage, ok := updates["backgroundImage"].(string); ok {
		content.BackgroundImage = backgroundImage
	}
	if backgroundSize, ok := updates["backgroundSize"].(string); ok {
		content.BackgroundSize = backgroundSize
	}
	if borderRadius, ok := updates["borderRadius"].(string); ok {
		content.BorderRadius = borderRadius
	}
	if border, ok := updates["border"].(string); ok {
		content.Border = border
	}
	if boxShadow, ok := updates["boxShadow"].(string); ok {
		content.BoxShadow = boxShadow
	}
	if clipPath, ok := updates["clipPath"].(string); ok {
		content.ClipPath = clipPath
	}
	if placeholder, ok := updates["placeholder"].(bool); ok {
		content.Placeholder = placeholder
	}
	return nil
}

func (r *ComponentRenderer) deepMergeLayout(comp *models.Component, updates map[string]interface{}) error {
	if comp.Type != models.ComponentTypeContainer {
		return nil
	}

	if comp.Layout == nil {
		comp.Layout = &models.ContainerLayout{}
	}

	if display, ok := updates["display"].(string); ok {
		comp.Layout.Display = display
	}
	if flexDirection, ok := updates["flexDirection"].(string); ok {
		comp.Layout.FlexDirection = flexDirection
	}
	if justifyContent, ok := updates["justifyContent"].(string); ok {
		comp.Layout.JustifyContent = justifyContent
	}
	if alignItems, ok := updates["alignItems"].(string); ok {
		comp.Layout.AlignItems = alignItems
	}
	if gap, ok := updates["gap"].(string); ok {
		comp.Layout.Gap = gap
	}
	if padding, ok := updates["padding"].(string); ok {
		comp.Layout.Padding = padding
	}
	if children, ok := updates["children"].([]interface{}); ok {
		comp.Layout.Children = make([]string, 0, len(children))
		for _, child := range children {
			if childStr, ok := child.(string); ok {
				comp.Layout.Children = append(comp.Layout.Children, childStr)
			}
		}
	}
	return nil
}

func (r *ComponentRenderer) deepMergeStyle(comp *models.Component, updates map[string]interface{}) error {
	if comp.Type != models.ComponentTypeDivider {
		return nil
	}

	if comp.Style == nil {
		comp.Style = &models.DividerStyle{}
	}

	if backgroundColor, ok := updates["backgroundColor"].(string); ok {
		comp.Style.BackgroundColor = backgroundColor
	}
	if height, ok := updates["height"].(string); ok {
		comp.Style.Height = height
	}
	if margin, ok := updates["margin"].(string); ok {
		comp.Style.Margin = margin
	}
	if borderRadius, ok := updates["borderRadius"].(string); ok {
		comp.Style.BorderRadius = borderRadius
	}
	return nil
}

func (r *ComponentRenderer) deepMergeResponsive(comp *models.Component, updates map[string]interface{}) error {
	if comp.Responsive == nil {
		comp.Responsive = &models.ResponsiveConfig{}
	}

	if mobile, ok := updates["mobile"].(map[string]interface{}); ok {
		if comp.Responsive.Mobile == nil {
			comp.Responsive.Mobile = &models.ResponsiveBreakpoint{}
		}
		r.mergeResponsiveBreakpoint(comp.Responsive.Mobile, mobile)
	}
	if tablet, ok := updates["tablet"].(map[string]interface{}); ok {
		if comp.Responsive.Tablet == nil {
			comp.Responsive.Tablet = &models.ResponsiveBreakpoint{}
		}
		r.mergeResponsiveBreakpoint(comp.Responsive.Tablet, tablet)
	}
	if desktop, ok := updates["desktop"].(map[string]interface{}); ok {
		if comp.Responsive.Desktop == nil {
			comp.Responsive.Desktop = &models.ResponsiveBreakpoint{}
		}
		r.mergeResponsiveBreakpoint(comp.Responsive.Desktop, desktop)
	}
	return nil
}

func (r *ComponentRenderer) mergeResponsiveBreakpoint(bp *models.ResponsiveBreakpoint, updates map[string]interface{}) {
	if width, ok := updates["width"].(string); ok {
		bp.Width = width
	}
	if height, ok := updates["height"].(string); ok {
		bp.Height = height
	}
	if fontSize, ok := updates["fontSize"].(string); ok {
		bp.FontSize = fontSize
	}
	if padding, ok := updates["padding"].(string); ok {
		bp.Padding = padding
	}
	if margin, ok := updates["margin"].(string); ok {
		bp.Margin = margin
	}
	if display, ok := updates["display"].(string); ok {
		bp.Display = display
	}
	if visible, ok := updates["visible"].(bool); ok {
		bp.Visible = &visible
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

	if err := r.evaluateComponentTemplates(finalConfig, event, template); err != nil {
		return fmt.Errorf("failed to evaluate component templates: %w", err)
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

	tmpl, err := r.loadComponentTemplates()
	if err != nil {
		return fmt.Errorf("failed to load component templates: %w", err)
	}

	return r.engine.Execute(w, tmpl, data)
}

func (r *ComponentRenderer) evaluateComponentTemplates(config *models.ComponentConfiguration, event *models.Event, tmpl *models.Template) error {
	templateData := map[string]interface{}{
		"Event":    event,
		"Template": tmpl,
	}

	for i := range config.Components {
		comp := &config.Components[i]
		if comp.Content == nil {
			continue
		}

		if comp.Type == models.ComponentTypeTextBox && comp.Content.TextBox != nil {
			textVal := comp.Content.TextBox.Text
			if strings.Contains(textVal, "{{") {
				parsedTmpl, err := r.engine.Parse(textVal)
				if err != nil {
					return fmt.Errorf("failed to parse template in component %s: %w", comp.ID, err)
				}

				evaluated, err := r.engine.ExecuteToString(parsedTmpl, templateData)
				if err != nil {
					return fmt.Errorf("failed to evaluate template in component %s: %w", comp.ID, err)
				}

				comp.Content.TextBox.EvaluatedText = evaluated
				comp.Content.TextBox.IsEvaluated = true
			}
		}
	}

	return nil
}

func (r *ComponentRenderer) loadComponentTemplates() (*template.Template, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tmpl := template.New("base_component.html").Funcs(r.engine.funcMap)

	baseTemplate := filepath.Join(projectRoot, "templates/web/rsvp_themes/base_component.html")
	tmpl, err = tmpl.ParseFiles(baseTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base component template: %w", err)
	}

	partials := []string{
		"templates/web/partials/component_textbox.html",
		"templates/web/partials/component_image.html",
		"templates/web/partials/component_background.html",
		"templates/web/partials/component_overlay.html",
		"templates/web/partials/component_container.html",
		"templates/web/partials/component_divider.html",
	}

	for _, partial := range partials {
		fullPath := filepath.Join(projectRoot, partial)
		tmpl, err = tmpl.ParseFiles(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse partial %s: %w", partial, err)
		}
	}

	return tmpl, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}
