package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func createTestUserForTemplate(t *testing.T, repo UserRepository) *models.User {
	t.Helper()

	user := &models.User{
		Email: fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func TestTemplateRepository_GetComponentConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUserForTemplate(t, userRepo)
	ctx := context.Background()

	t.Run("returns component config for template with config", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name:        "Test Template",
				Category:    "card",
				Description: "Test description",
			},
			Layout: models.LayoutConfig{
				Mode:            "card",
				CardWidth:       "800px",
				BackgroundColor: "#ffffff",
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

		configJSON, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}
		configStr := string(configJSON)

		template := &models.Template{
			Name:            "Test Template",
			Type:            models.TemplateTypeRSVPPage,
			Description:     "Test",
			HTMLContent:     "<html></html>",
			IsDefault:       false,
			IsActive:        true,
			CreatedBy:       user.ID,
			Category:        models.CategoryCard,
			ComponentConfig: &configStr,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		result, err := repo.GetComponentConfig(ctx, template.ID)
		if err != nil {
			t.Fatalf("GetComponentConfig failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected component config, got nil")
		}

		if result.Version != config.Version {
			t.Errorf("Expected version %s, got %s", config.Version, result.Version)
		}

		if len(result.Components) != len(config.Components) {
			t.Errorf("Expected %d components, got %d", len(config.Components), len(result.Components))
		}
	})

	t.Run("returns nil for template without config", func(t *testing.T) {
		template := &models.Template{
			Name:        "No Config Template",
			Type:        models.TemplateTypeRSVPPage,
			Description: "Test",
			HTMLContent: "<html></html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		result, err := repo.GetComponentConfig(ctx, template.ID)
		if err != nil {
			t.Fatalf("GetComponentConfig failed: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil config, got %+v", result)
		}
	})

	t.Run("returns error for non-existent template", func(t *testing.T) {
		_, err := repo.GetComponentConfig(ctx, 99999)
		if err == nil {
			t.Fatal("Expected error for non-existent template, got nil")
		}

		if _, ok := err.(*models.NotFoundError); !ok {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		invalidJSON := "invalid json"
		template := &models.Template{
			Name:            "Invalid JSON Template",
			Type:            models.TemplateTypeRSVPPage,
			Description:     "Test",
			HTMLContent:     "<html></html>",
			IsDefault:       false,
			IsActive:        true,
			CreatedBy:       user.ID,
			Category:        models.CategoryCard,
			ComponentConfig: &invalidJSON,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		_, err := repo.GetComponentConfig(ctx, template.ID)
		if err == nil {
			t.Fatal("Expected error for invalid JSON, got nil")
		}
	})
}

func TestTemplateRepository_UpdateComponentConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUserForTemplate(t, userRepo)
	ctx := context.Background()

	t.Run("updates component config successfully", func(t *testing.T) {
		template := &models.Template{
			Name:        "Update Test Template",
			Type:        models.TemplateTypeRSVPPage,
			Description: "Test",
			HTMLContent: "<html></html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name:        "Updated Template",
				Category:    "card",
				Description: "Updated description",
			},
			Layout: models.LayoutConfig{
				Mode:            "card",
				CardWidth:       "900px",
				BackgroundColor: "#f0f0f0",
			},
			Components: []models.Component{
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
					ZIndex:  5,
					Visible: true,
				},
			},
		}

		if err := repo.UpdateComponentConfig(ctx, template.ID, config); err != nil {
			t.Fatalf("UpdateComponentConfig failed: %v", err)
		}

		result, err := repo.GetComponentConfig(ctx, template.ID)
		if err != nil {
			t.Fatalf("GetComponentConfig failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected component config, got nil")
		}

		if result.Metadata.Name != config.Metadata.Name {
			t.Errorf("Expected name %s, got %s", config.Metadata.Name, result.Metadata.Name)
		}

		if result.Layout.CardWidth != config.Layout.CardWidth {
			t.Errorf("Expected card width %s, got %s", config.Layout.CardWidth, result.Layout.CardWidth)
		}
	})

	t.Run("returns error for non-existent template", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
		}

		err := repo.UpdateComponentConfig(ctx, 99999, config)
		if err == nil {
			t.Fatal("Expected error for non-existent template, got nil")
		}

		if _, ok := err.(*models.NotFoundError); !ok {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})

	t.Run("clears config when nil is provided", func(t *testing.T) {
		configJSON := `{"version":"1.0","metadata":{"name":"Test"}}`
		template := &models.Template{
			Name:            "Clear Config Template",
			Type:            models.TemplateTypeRSVPPage,
			Description:     "Test",
			HTMLContent:     "<html></html>",
			IsDefault:       false,
			IsActive:        true,
			CreatedBy:       user.ID,
			Category:        models.CategoryCard,
			ComponentConfig: &configJSON,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		if err := repo.UpdateComponentConfig(ctx, template.ID, nil); err != nil {
			t.Fatalf("UpdateComponentConfig failed: %v", err)
		}

		result, err := repo.GetComponentConfig(ctx, template.ID)
		if err != nil {
			t.Fatalf("GetComponentConfig failed: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil config after clearing, got %+v", result)
		}
	})
}

func TestTemplateRepository_ValidateComponentConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	ctx := context.Background()

	t.Run("validates valid config", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name:        "Valid Config",
				Category:    "card",
				Description: "Valid description",
			},
			Layout: models.LayoutConfig{
				Mode: "card",
			},
			Components: []models.Component{
				{
					ID:   "valid-component",
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

		err := repo.ValidateComponentConfig(ctx, config)
		if err != nil {
			t.Errorf("Expected valid config to pass validation, got error: %v", err)
		}
	})

	t.Run("rejects config with empty version", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
		}

		err := repo.ValidateComponentConfig(ctx, config)
		if err == nil {
			t.Error("Expected error for empty version, got nil")
		}
	})

	t.Run("rejects config with invalid component type", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "invalid-component",
					Type: "InvalidType",
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

		err := repo.ValidateComponentConfig(ctx, config)
		if err == nil {
			t.Error("Expected error for invalid component type, got nil")
		}
	})

	t.Run("rejects config with duplicate component IDs", func(t *testing.T) {
		config := &models.ComponentConfiguration{
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
				{
					ID:   "duplicate-id",
					Type: models.ComponentTypeImage,
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

		err := repo.ValidateComponentConfig(ctx, config)
		if err == nil {
			t.Error("Expected error for duplicate component IDs, got nil")
		}
	})

	t.Run("rejects config with empty component ID", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: []models.Component{
				{
					ID:   "",
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

		err := repo.ValidateComponentConfig(ctx, config)
		if err == nil {
			t.Error("Expected error for empty component ID, got nil")
		}
	})

	t.Run("rejects config with too many components", func(t *testing.T) {
		components := make([]models.Component, 51)
		for i := 0; i < 51; i++ {
			components[i] = models.Component{
				ID:   string(rune('a' + i)),
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
				},
				Dimensions: models.Dimensions{
					Width:  "100%",
					Height: "auto",
				},
			}
		}

		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name: "Test",
			},
			Components: components,
		}

		err := repo.ValidateComponentConfig(ctx, config)
		if err == nil {
			t.Error("Expected error for too many components, got nil")
		}
	})

	t.Run("accepts nil config", func(t *testing.T) {
		err := repo.ValidateComponentConfig(ctx, nil)
		if err != nil {
			t.Errorf("Expected nil config to be valid, got error: %v", err)
		}
	})
}
