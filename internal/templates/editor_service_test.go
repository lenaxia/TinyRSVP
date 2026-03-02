package templates

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type MockTemplateRepository struct {
	GetByIDFunc                 func(ctx context.Context, id int64) (*models.Template, error)
	GetComponentConfigFunc      func(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error)
	UpdateComponentConfigFunc   func(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error
	ValidateComponentConfigFunc func(ctx context.Context, config *models.ComponentConfiguration) error
}

func (m *MockTemplateRepository) Create(ctx context.Context, template *models.Template) error {
	return nil
}

func (m *MockTemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockTemplateRepository) GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}

func (m *MockTemplateRepository) GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}

func (m *MockTemplateRepository) GetByNameAndType(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}

func (m *MockTemplateRepository) List(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	return nil, nil
}

func (m *MockTemplateRepository) Update(ctx context.Context, template *models.Template) error {
	return nil
}

func (m *MockTemplateRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockTemplateRepository) SetActive(ctx context.Context, id int64, active bool) error {
	return nil
}

func (m *MockTemplateRepository) IsTemplateInUse(ctx context.Context, id int64) (bool, error) {
	return false, nil
}

func (m *MockTemplateRepository) SetDefault(ctx context.Context, id int64) error {
	return nil
}

func (m *MockTemplateRepository) GetTemplatesByCategory(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error) {
	return nil, nil
}

func (m *MockTemplateRepository) ListThemes(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error) {
	return nil, nil
}

func (m *MockTemplateRepository) GetComponentConfig(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error) {
	if m.GetComponentConfigFunc != nil {
		return m.GetComponentConfigFunc(ctx, templateID)
	}
	return nil, nil
}

func (m *MockTemplateRepository) UpdateComponentConfig(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error {
	if m.UpdateComponentConfigFunc != nil {
		return m.UpdateComponentConfigFunc(ctx, templateID, config)
	}
	return nil
}

func (m *MockTemplateRepository) ValidateComponentConfig(ctx context.Context, config *models.ComponentConfiguration) error {
	if m.ValidateComponentConfigFunc != nil {
		return m.ValidateComponentConfigFunc(ctx, config)
	}
	return nil
}

func TestEditorService_GetEditableTemplate(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)
	ctx := context.Background()

	t.Run("returns template with component config", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test Template",
			},
			Components: []models.Component{
				{
					ID:   "test-component",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
					ZIndex:  10,
					Visible: true,
				},
			},
		}

		configJSON, _ := json.Marshal(config)
		configStr := string(configJSON)

		expectedTemplate := &models.Template{
			ID:              1,
			Name:            "Test Template",
			Type:            models.TemplateTypeRSVPPage,
			ComponentConfig: &configStr,
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return expectedTemplate, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return config, nil
		}

		result, err := service.GetEditableTemplate(ctx, 1)
		if err != nil {
			t.Fatalf("GetEditableTemplate failed: %v", err)
		}

		if result.Template == nil {
			t.Fatal("Expected template, got nil")
		}

		if result.ComponentConfig == nil {
			t.Fatal("Expected component config, got nil")
		}

		if len(result.ComponentConfig.Components) != 1 {
			t.Errorf("Expected 1 component, got %d", len(result.ComponentConfig.Components))
		}
	})

	t.Run("returns error for non-existent template", func(t *testing.T) {
		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return nil, &models.NotFoundError{Resource: "Template", ID: id}
		}

		_, err := service.GetEditableTemplate(ctx, 99999)
		if err == nil {
			t.Fatal("Expected error for non-existent template, got nil")
		}

		if _, ok := err.(*models.NotFoundError); !ok {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestEditorService_UpdateComponents(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("updates components successfully", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "old-component",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		mockRepo.UpdateComponentConfigFunc = func(ctx context.Context, id int64, config *models.ComponentConfiguration) error {
			return nil
		}

		newComponents := []models.Component{
			{
				ID:   "new-component",
				Type: models.ComponentTypeImage,
				Position: models.Position{
					Mode: models.PositionModeRelative,
				},
				Dimensions: models.Dimensions{
					Width:  "50%",
					Height: "200px",
				},
			},
		}

		err := service.UpdateComponents(ctx, 1, newComponents)
		if err != nil {
			t.Fatalf("UpdateComponents failed: %v", err)
		}
	})

	t.Run("requires authentication", func(t *testing.T) {
		unauthCtx := context.Background()

		err := service.UpdateComponents(unauthCtx, 1, []models.Component{})
		if err == nil {
			t.Fatal("Expected error for unauthenticated request, got nil")
		}

		if _, ok := err.(*models.UnauthorizedError); !ok {
			t.Errorf("Expected UnauthorizedError, got %T", err)
		}
	})

	t.Run("requires admin role", func(t *testing.T) {
		regularUser := &models.User{
			ID:   2,
			Role: models.RoleEventManager,
		}
		regularCtx := auth.WithUser(context.Background(), regularUser)

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: 1,
			}, nil
		}

		err := service.UpdateComponents(regularCtx, 1, []models.Component{})
		if err == nil {
			t.Fatal("Expected error for non-admin user, got nil")
		}

		if _, ok := err.(*models.ForbiddenError); !ok {
			t.Errorf("Expected ForbiddenError, got %T", err)
		}
	})
}

func TestEditorService_AddComponent(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("adds component successfully", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "existing-component",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		var savedConfig *models.ComponentConfiguration
		mockRepo.UpdateComponentConfigFunc = func(ctx context.Context, id int64, config *models.ComponentConfiguration) error {
			savedConfig = config
			return nil
		}

		newComponent := models.Component{
			ID:   "new-component",
			Type: models.ComponentTypeImage,
			Position: models.Position{
				Mode: models.PositionModeRelative,
			},
			Dimensions: models.Dimensions{
				Width:  "50%",
				Height: "200px",
			},
			ZIndex:  5,
			Visible: true,
		}

		err := service.AddComponent(ctx, 1, newComponent)
		if err != nil {
			t.Fatalf("AddComponent failed: %v", err)
		}

		if savedConfig == nil {
			t.Fatal("Expected config to be saved")
		}

		if len(savedConfig.Components) != 2 {
			t.Errorf("Expected 2 components, got %d", len(savedConfig.Components))
		}
	})

	t.Run("rejects duplicate component ID", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "duplicate-id",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		duplicateComponent := models.Component{
			ID:   "duplicate-id",
			Type: models.ComponentTypeImage,
			Position: models.Position{
				Mode: models.PositionModeRelative,
			},
			Dimensions: models.Dimensions{
				Width:  "50%",
				Height: "200px",
			},
		}

		err := service.AddComponent(ctx, 1, duplicateComponent)
		if err == nil {
			t.Fatal("Expected error for duplicate component ID, got nil")
		}

		if _, ok := err.(*models.ValidationError); !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})
}

func TestEditorService_RemoveComponent(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("removes component successfully", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "component-1",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
				{
					ID:   "component-2",
					Type: models.ComponentTypeImage,
					Position: models.Position{
						Mode: models.PositionModeRelative,
					},
					Dimensions: models.Dimensions{
						Width:  "50%",
						Height: "200px",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		var savedConfig *models.ComponentConfiguration
		mockRepo.UpdateComponentConfigFunc = func(ctx context.Context, id int64, config *models.ComponentConfiguration) error {
			savedConfig = config
			return nil
		}

		err := service.RemoveComponent(ctx, 1, "component-1")
		if err != nil {
			t.Fatalf("RemoveComponent failed: %v", err)
		}

		if savedConfig == nil {
			t.Fatal("Expected config to be saved")
		}

		if len(savedConfig.Components) != 1 {
			t.Errorf("Expected 1 component, got %d", len(savedConfig.Components))
		}

		if savedConfig.Components[0].ID != "component-2" {
			t.Errorf("Expected component-2, got %s", savedConfig.Components[0].ID)
		}
	})

	t.Run("returns error for non-existent component", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "component-1",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		err := service.RemoveComponent(ctx, 1, "non-existent")
		if err == nil {
			t.Fatal("Expected error for non-existent component, got nil")
		}

		if _, ok := err.(*models.NotFoundError); !ok {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestEditorService_UpdateComponentProperty(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("updates component property successfully", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "test-component",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
					ZIndex:  10,
					Visible: true,
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		var savedConfig *models.ComponentConfiguration
		mockRepo.UpdateComponentConfigFunc = func(ctx context.Context, id int64, config *models.ComponentConfiguration) error {
			savedConfig = config
			return nil
		}

		err := service.UpdateComponentProperty(ctx, 1, "test-component", "zIndex", 20)
		if err != nil {
			t.Fatalf("UpdateComponentProperty failed: %v", err)
		}

		if savedConfig == nil {
			t.Fatal("Expected config to be saved")
		}

		if savedConfig.Components[0].ZIndex != 20 {
			t.Errorf("Expected zIndex 20, got %d", savedConfig.Components[0].ZIndex)
		}
	})

	t.Run("returns error for non-existent component", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version:    "1.0",
			Components: []models.Component{},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		err := service.UpdateComponentProperty(ctx, 1, "non-existent", "zIndex", 20)
		if err == nil {
			t.Fatal("Expected error for non-existent component, got nil")
		}

		if _, ok := err.(*models.NotFoundError); !ok {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestEditorService_ReorderComponents(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("reorders components successfully", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:     "component-1",
					Type:   models.ComponentTypeTextBox,
					ZIndex: 10,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
				{
					ID:     "component-2",
					Type:   models.ComponentTypeImage,
					ZIndex: 20,
					Position: models.Position{
						Mode: models.PositionModeRelative,
					},
					Dimensions: models.Dimensions{
						Width:  "50%",
						Height: "200px",
					},
				},
				{
					ID:     "component-3",
					Type:   models.ComponentTypeBackground,
					ZIndex: 5,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "100%",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		var savedConfig *models.ComponentConfiguration
		mockRepo.UpdateComponentConfigFunc = func(ctx context.Context, id int64, config *models.ComponentConfiguration) error {
			savedConfig = config
			return nil
		}

		newOrder := []string{"component-3", "component-1", "component-2"}
		err := service.ReorderComponents(ctx, 1, newOrder)
		if err != nil {
			t.Fatalf("ReorderComponents failed: %v", err)
		}

		if savedConfig == nil {
			t.Fatal("Expected config to be saved")
		}

		if savedConfig.Components[0].ZIndex != 0 {
			t.Errorf("Expected first component zIndex 0, got %d", savedConfig.Components[0].ZIndex)
		}

		if savedConfig.Components[1].ZIndex != 1 {
			t.Errorf("Expected second component zIndex 1, got %d", savedConfig.Components[1].ZIndex)
		}

		if savedConfig.Components[2].ZIndex != 2 {
			t.Errorf("Expected third component zIndex 2, got %d", savedConfig.Components[2].ZIndex)
		}
	})

	t.Run("returns error for missing component in order", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Components: []models.Component{
				{
					ID:   "component-1",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		newOrder := []string{"component-1", "non-existent"}
		err := service.ReorderComponents(ctx, 1, newOrder)
		if err == nil {
			t.Fatal("Expected error for non-existent component in order, got nil")
		}

		if _, ok := err.(*models.ValidationError); !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})
}

func TestEditorService_PreviewChanges(t *testing.T) {
	mockRepo := &MockTemplateRepository{}
	service := NewEditorService(mockRepo)

	adminUser := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("generates preview without saving", func(t *testing.T) {
		existingConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "test-component",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
					ZIndex:  10,
					Visible: true,
				},
			},
		}

		mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:        1,
				Name:      "Test",
				CreatedBy: adminUser.ID,
			}, nil
		}

		mockRepo.GetComponentConfigFunc = func(ctx context.Context, id int64) (*models.ComponentConfiguration, error) {
			return existingConfig, nil
		}

		updateCalled := false
		mockRepo.UpdateComponentConfigFunc = func(ctx context.Context, id int64, config *models.ComponentConfiguration) error {
			updateCalled = true
			return nil
		}

		changes := &ComponentChanges{
			Updates: []ComponentUpdate{
				{
					ComponentID: "test-component",
					Property:    "zIndex",
					Value:       20,
				},
			},
		}

		result, err := service.PreviewChanges(ctx, 1, changes)
		if err != nil {
			t.Fatalf("PreviewChanges failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected preview result, got nil")
		}

		if updateCalled {
			t.Error("Expected UpdateComponentConfig not to be called during preview")
		}

		if result.Components[0].ZIndex != 20 {
			t.Errorf("Expected preview zIndex 20, got %d", result.Components[0].ZIndex)
		}
	})
}
