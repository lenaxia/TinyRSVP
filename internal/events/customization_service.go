package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type CustomizationService interface {
	GetEventCustomization(ctx context.Context, eventID int64) (*EventCustomizationData, error)
	UpdateEventCustomization(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error
	PreviewEventCustomization(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error)
	ResetEventCustomization(ctx context.Context, eventID int64) error
	ValidateEventCustomization(overrides *models.ComponentOverrides) error
}

type EventCustomizationData struct {
	Event          *models.Event                  `json:"event"`
	Template       *models.Template               `json:"template"`
	TemplateConfig *models.ComponentConfiguration `json:"templateConfig"`
	EventOverrides *models.ComponentOverrides     `json:"eventOverrides,omitempty"`
	MergedConfig   *models.ComponentConfiguration `json:"mergedConfig"`
}

type customizationService struct {
	eventRepo    repositories.EventRepository
	templateRepo templateRepository
	authz        auth.AuthorizationChecker
}

type templateRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Template, error)
}

func NewCustomizationService(
	eventRepo repositories.EventRepository,
	templateRepo templateRepository,
	authz auth.AuthorizationChecker,
) CustomizationService {
	return &customizationService{
		eventRepo:    eventRepo,
		templateRepo: templateRepo,
		authz:        authz,
	}
}

func (s *customizationService) GetEventCustomization(ctx context.Context, eventID int64) (*EventCustomizationData, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "get event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return nil, &models.PermissionDeniedError{
			Action:   "get event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	if event.TemplateID == nil {
		return nil, fmt.Errorf("event has no template assigned")
	}

	template, err := s.templateRepo.GetByID(ctx, *event.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	var templateConfig *models.ComponentConfiguration
	if template.ComponentConfig != nil && *template.ComponentConfig != "" {
		templateConfig = &models.ComponentConfiguration{}
		if err := json.Unmarshal([]byte(*template.ComponentConfig), templateConfig); err != nil {
			return nil, fmt.Errorf("failed to parse template config: %w", err)
		}
	}

	var eventOverrides *models.ComponentOverrides
	if event.ComponentOverrides != nil && *event.ComponentOverrides != "" {
		eventOverrides = &models.ComponentOverrides{}
		if err := json.Unmarshal([]byte(*event.ComponentOverrides), eventOverrides); err != nil {
			return nil, fmt.Errorf("failed to parse event overrides: %w", err)
		}
	}

	mergedConfig := s.mergeConfigurations(templateConfig, eventOverrides)

	return &EventCustomizationData{
		Event:          event,
		Template:       template,
		TemplateConfig: templateConfig,
		EventOverrides: eventOverrides,
		MergedConfig:   mergedConfig,
	}, nil
}

func (s *customizationService) UpdateEventCustomization(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "update event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "update event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	if err := s.ValidateEventCustomization(overrides); err != nil {
		return err
	}

	if err := s.eventRepo.UpdateComponentOverrides(ctx, eventID, overrides); err != nil {
		return fmt.Errorf("failed to update component overrides: %w", err)
	}

	return nil
}

func (s *customizationService) PreviewEventCustomization(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "preview event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if !s.authz.CanViewEvent(ctx, user, event) {
		return nil, &models.PermissionDeniedError{
			Action:   "preview event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	if event.TemplateID == nil {
		return nil, fmt.Errorf("event has no template assigned")
	}

	template, err := s.templateRepo.GetByID(ctx, *event.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	var templateConfig *models.ComponentConfiguration
	if template.ComponentConfig != nil && *template.ComponentConfig != "" {
		templateConfig = &models.ComponentConfiguration{}
		if err := json.Unmarshal([]byte(*template.ComponentConfig), templateConfig); err != nil {
			return nil, fmt.Errorf("failed to parse template config: %w", err)
		}
	}

	if err := s.ValidateEventCustomization(overrides); err != nil {
		return nil, err
	}

	mergedConfig := s.mergeConfigurations(templateConfig, overrides)

	return mergedConfig, nil
}

func (s *customizationService) ResetEventCustomization(ctx context.Context, eventID int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "reset event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "reset event customization",
			Resource: "Event",
			ID:       eventID,
		}
	}

	if err := s.eventRepo.DeleteComponentOverrides(ctx, eventID); err != nil {
		return fmt.Errorf("failed to delete component overrides: %w", err)
	}

	return nil
}

func (s *customizationService) ValidateEventCustomization(overrides *models.ComponentOverrides) error {
	if overrides == nil {
		return fmt.Errorf("overrides cannot be nil")
	}

	if overrides.Version == "" {
		return &models.ValidationError{
			Field:   "version",
			Message: "version is required",
		}
	}

	for i, override := range overrides.Overrides {
		if override.ID == "" {
			return &models.ValidationError{
				Field:   fmt.Sprintf("overrides[%d].id", i),
				Message: "component ID is required",
			}
		}
	}

	for i, removal := range overrides.Removals {
		if removal == "" {
			return &models.ValidationError{
				Field:   fmt.Sprintf("removals[%d]", i),
				Message: "component ID is required",
			}
		}
	}

	for i, addition := range overrides.Additions {
		if addition.ID == "" {
			return &models.ValidationError{
				Field:   fmt.Sprintf("additions[%d].id", i),
				Message: "component ID is required",
			}
		}
		if !addition.Type.IsValid() {
			return &models.ValidationError{
				Field:   fmt.Sprintf("additions[%d].type", i),
				Message: "invalid component type",
			}
		}
	}

	return nil
}

func (s *customizationService) mergeConfigurations(templateConfig *models.ComponentConfiguration, overrides *models.ComponentOverrides) *models.ComponentConfiguration {
	if templateConfig == nil {
		templateConfig = &models.ComponentConfiguration{
			Version:    "1.0",
			Components: []models.Component{},
		}
	}

	if overrides == nil {
		return templateConfig
	}

	merged := &models.ComponentConfiguration{
		Version:    templateConfig.Version,
		Metadata:   templateConfig.Metadata,
		Layout:     templateConfig.Layout,
		Components: make([]models.Component, 0, len(templateConfig.Components)),
	}

	removalSet := make(map[string]bool)
	for _, id := range overrides.Removals {
		removalSet[id] = true
	}

	overrideMap := make(map[string]models.ComponentOverride)
	for _, override := range overrides.Overrides {
		overrideMap[override.ID] = override
	}

	for _, component := range templateConfig.Components {
		if removalSet[component.ID] {
			continue
		}

		if override, ok := overrideMap[component.ID]; ok {
			component = s.applyOverride(component, override)
		}

		merged.Components = append(merged.Components, component)
	}

	merged.Components = append(merged.Components, overrides.Additions...)

	return merged
}

func (s *customizationService) applyOverride(component models.Component, override models.ComponentOverride) models.Component {
	updatesJSON, err := json.Marshal(override.Updates)
	if err != nil {
		return component
	}

	componentJSON, err := json.Marshal(component)
	if err != nil {
		return component
	}

	var componentMap map[string]interface{}
	if err := json.Unmarshal(componentJSON, &componentMap); err != nil {
		return component
	}

	var updatesMap map[string]interface{}
	if err := json.Unmarshal(updatesJSON, &updatesMap); err != nil {
		return component
	}

	s.deepMerge(componentMap, updatesMap)

	mergedJSON, err := json.Marshal(componentMap)
	if err != nil {
		return component
	}

	var mergedComponent models.Component
	if err := json.Unmarshal(mergedJSON, &mergedComponent); err != nil {
		return component
	}

	return mergedComponent
}

func (s *customizationService) deepMerge(dst, src map[string]interface{}) {
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			if dstMap, dstIsMap := dstVal.(map[string]interface{}); dstIsMap {
				if srcMap, srcIsMap := srcVal.(map[string]interface{}); srcIsMap {
					s.deepMerge(dstMap, srcMap)
					continue
				}
			}
		}
		dst[key] = srcVal
	}
}
