package templates

import (
	"context"
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EditorService interface {
	GetEditableTemplate(ctx context.Context, templateID int64) (*EditableTemplate, error)
	UpdateComponents(ctx context.Context, templateID int64, components []models.Component) error
	AddComponent(ctx context.Context, templateID int64, component models.Component) error
	RemoveComponent(ctx context.Context, templateID int64, componentID string) error
	UpdateComponentProperty(ctx context.Context, templateID int64, componentID string, property string, value interface{}) error
	ReorderComponents(ctx context.Context, templateID int64, componentIDs []string) error
	PreviewChanges(ctx context.Context, templateID int64, changes *ComponentChanges) (*models.ComponentConfiguration, error)
}

type EditableTemplate struct {
	Template        *models.Template
	ComponentConfig *models.ComponentConfiguration
}

type ComponentUpdate struct {
	ComponentID string      `json:"component_id"`
	Property    string      `json:"property"`
	Value       interface{} `json:"value"`
}

type ComponentChanges struct {
	Updates   []ComponentUpdate `json:"updates"`
	Additions []models.Component `json:"additions"`
	Removals  []string          `json:"removals"`
}

type editorService struct {
	repo repositories.TemplateRepository
}

func NewEditorService(repo repositories.TemplateRepository) EditorService {
	return &editorService{
		repo: repo,
	}
}

func (s *editorService) GetEditableTemplate(ctx context.Context, templateID int64) (*EditableTemplate, error) {
	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component config: %w", err)
	}

	return &EditableTemplate{
		Template:        template,
		ComponentConfig: config,
	}, nil
}

func (s *editorService) UpdateComponents(ctx context.Context, templateID int64, components []models.Component) error {
	if err := s.checkPermission(ctx, templateID); err != nil {
		return err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return err
	}

	if config == nil {
		config = &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Updated Template",
			},
		}
	}

	config.Components = components

	if err := s.repo.ValidateComponentConfig(ctx, config); err != nil {
		return err
	}

	return s.repo.UpdateComponentConfig(ctx, templateID, config)
}

func (s *editorService) AddComponent(ctx context.Context, templateID int64, component models.Component) error {
	if err := s.checkPermission(ctx, templateID); err != nil {
		return err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return err
	}

	if config == nil {
		config = &models.ComponentConfiguration{
			Version:    "1.0",
			Components: []models.Component{},
		}
	}

	for _, existing := range config.Components {
		if existing.ID == component.ID {
			return &models.ValidationError{
				Field:   "component.id",
				Message: fmt.Sprintf("component with ID %s already exists", component.ID),
			}
		}
	}

	config.Components = append(config.Components, component)

	if err := s.repo.ValidateComponentConfig(ctx, config); err != nil {
		return err
	}

	return s.repo.UpdateComponentConfig(ctx, templateID, config)
}

func (s *editorService) RemoveComponent(ctx context.Context, templateID int64, componentID string) error {
	if err := s.checkPermission(ctx, templateID); err != nil {
		return err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return err
	}

	if config == nil {
		return &models.NotFoundError{
			Resource: "Component",
			ID:       0,
		}
	}

	found := false
	newComponents := make([]models.Component, 0, len(config.Components))
	for _, comp := range config.Components {
		if comp.ID == componentID {
			found = true
			continue
		}
		newComponents = append(newComponents, comp)
	}

	if !found {
		return &models.NotFoundError{
			Resource: "Component",
			ID:       0,
		}
	}

	config.Components = newComponents

	return s.repo.UpdateComponentConfig(ctx, templateID, config)
}

func (s *editorService) UpdateComponentProperty(ctx context.Context, templateID int64, componentID string, property string, value interface{}) error {
	if err := s.checkPermission(ctx, templateID); err != nil {
		return err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return err
	}

	if config == nil {
		return &models.NotFoundError{
			Resource: "Component",
			ID:       0,
		}
	}

	found := false
	for i := range config.Components {
		if config.Components[i].ID == componentID {
			found = true
			if err := s.applyPropertyUpdate(&config.Components[i], property, value); err != nil {
				return err
			}
			break
		}
	}

	if !found {
		return &models.NotFoundError{
			Resource: "Component",
			ID:       0,
		}
	}

	if err := s.repo.ValidateComponentConfig(ctx, config); err != nil {
		return err
	}

	return s.repo.UpdateComponentConfig(ctx, templateID, config)
}

func (s *editorService) ReorderComponents(ctx context.Context, templateID int64, componentIDs []string) error {
	if err := s.checkPermission(ctx, templateID); err != nil {
		return err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return err
	}

	if config == nil {
		return &models.NotFoundError{
			Resource: "Template",
			ID:       templateID,
		}
	}

	componentMap := make(map[string]*models.Component)
	for i := range config.Components {
		componentMap[config.Components[i].ID] = &config.Components[i]
	}

	newComponents := make([]models.Component, 0, len(componentIDs))
	for i, id := range componentIDs {
		comp, exists := componentMap[id]
		if !exists {
			return &models.ValidationError{
				Field:   "component_ids",
				Message: fmt.Sprintf("component with ID %s not found", id),
			}
		}
		comp.ZIndex = i
		newComponents = append(newComponents, *comp)
	}

	config.Components = newComponents

	return s.repo.UpdateComponentConfig(ctx, templateID, config)
}

func (s *editorService) PreviewChanges(ctx context.Context, templateID int64, changes *ComponentChanges) (*models.ComponentConfiguration, error) {
	if err := s.checkPermission(ctx, templateID); err != nil {
		return nil, err
	}

	config, err := s.repo.GetComponentConfig(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = &models.ComponentConfiguration{
			Version:    "1.0",
			Components: []models.Component{},
		}
	}

	previewConfig := s.deepCopyConfig(config)

	for _, update := range changes.Updates {
		for i := range previewConfig.Components {
			if previewConfig.Components[i].ID == update.ComponentID {
				if err := s.applyPropertyUpdate(&previewConfig.Components[i], update.Property, update.Value); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	for _, addition := range changes.Additions {
		previewConfig.Components = append(previewConfig.Components, addition)
	}

	removalSet := make(map[string]bool)
	for _, id := range changes.Removals {
		removalSet[id] = true
	}

	filtered := make([]models.Component, 0, len(previewConfig.Components))
	for _, comp := range previewConfig.Components {
		if !removalSet[comp.ID] {
			filtered = append(filtered, comp)
		}
	}
	previewConfig.Components = filtered

	if err := s.repo.ValidateComponentConfig(ctx, previewConfig); err != nil {
		return nil, err
	}

	return previewConfig, nil
}

func (s *editorService) checkPermission(ctx context.Context, templateID int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok || user.ID == 0 {
		return &models.UnauthorizedError{Message: "authentication required"}
	}

	if user.Role != models.RoleAdmin {
		template, err := s.repo.GetByID(ctx, templateID)
		if err != nil {
			return err
		}

		if template.CreatedBy != user.ID {
			return &models.ForbiddenError{Message: "only admins can edit templates"}
		}
	}

	return nil
}

func (s *editorService) applyPropertyUpdate(comp *models.Component, property string, value interface{}) error {
	switch property {
	case "zIndex":
		if v, ok := value.(int); ok {
			comp.ZIndex = v
		} else if v, ok := value.(float64); ok {
			comp.ZIndex = int(v)
		} else {
			return &models.ValidationError{
				Field:   property,
				Message: "zIndex must be a number",
			}
		}
	case "visible":
		if v, ok := value.(bool); ok {
			comp.Visible = v
		} else {
			return &models.ValidationError{
				Field:   property,
				Message: "visible must be a boolean",
			}
		}
	case "className":
		if v, ok := value.(string); ok {
			comp.ClassName = &v
		} else {
			return &models.ValidationError{
				Field:   property,
				Message: "className must be a string",
			}
		}
	case "position", "dimensions", "content", "layout", "style", "responsive":
		if v, ok := value.(map[string]interface{}); ok {
			switch property {
			case "position":
				if mode, ok := v["mode"].(string); ok {
					comp.Position.Mode = models.PositionMode(mode)
				}
				if x, ok := v["x"].(string); ok {
					comp.Position.X = &x
				}
				if y, ok := v["y"].(string); ok {
					comp.Position.Y = &y
				}
				if order, ok := v["order"].(float64); ok {
					orderInt := int(order)
					comp.Position.Order = &orderInt
				}
			case "dimensions":
				if width, ok := v["width"].(string); ok {
					comp.Dimensions.Width = width
				}
				if height, ok := v["height"].(string); ok {
					comp.Dimensions.Height = height
				}
			case "content":
				if comp.Content == nil {
					comp.Content = make(map[string]interface{})
				}
				for k, val := range v {
					comp.Content[k] = val
				}
			case "layout":
				if comp.Layout == nil {
					comp.Layout = make(map[string]interface{})
				}
				for k, val := range v {
					comp.Layout[k] = val
				}
			case "style":
				if comp.Style == nil {
					comp.Style = make(map[string]interface{})
				}
				for k, val := range v {
					comp.Style[k] = val
				}
			case "responsive":
				if comp.Responsive == nil {
					comp.Responsive = make(map[string]interface{})
				}
				for k, val := range v {
					comp.Responsive[k] = val
				}
			}
		} else {
			return &models.ValidationError{
				Field:   property,
				Message: fmt.Sprintf("%s must be an object", property),
			}
		}
	default:
		return &models.ValidationError{
			Field:   property,
			Message: fmt.Sprintf("unknown property: %s", property),
		}
	}

	return nil
}

func (s *editorService) deepCopyConfig(config *models.ComponentConfiguration) *models.ComponentConfiguration {
	if config == nil {
		return nil
	}

	copy := &models.ComponentConfiguration{
		Version:    config.Version,
		Metadata:   config.Metadata,
		Layout:     config.Layout,
		Components: make([]models.Component, len(config.Components)),
	}

	for i, comp := range config.Components {
		copy.Components[i] = s.deepCopyComponent(comp)
	}

	return copy
}

func (s *editorService) deepCopyComponent(comp models.Component) models.Component {
	copied := models.Component{
		ID:         comp.ID,
		Type:       comp.Type,
		Position:   comp.Position,
		Dimensions: comp.Dimensions,
		ZIndex:     comp.ZIndex,
		Visible:    comp.Visible,
		ClassName:  comp.ClassName,
		Children:   make([]string, len(comp.Children)),
	}

	copy(copied.Children, comp.Children)

	if comp.Content != nil {
		copied.Content = s.deepCopyMap(comp.Content)
	}

	if comp.Layout != nil {
		copied.Layout = s.deepCopyMap(comp.Layout)
	}

	if comp.Style != nil {
		copied.Style = s.deepCopyMap(comp.Style)
	}

	if comp.Responsive != nil {
		copied.Responsive = s.deepCopyMap(comp.Responsive)
	}

	return copied
}

func (s *editorService) deepCopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	copied := make(map[string]interface{}, len(m))
	for k, v := range m {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			copied[k] = s.deepCopyMap(nestedMap)
		} else {
			copied[k] = v
		}
	}

	return copied
}
