package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

func setupEditorTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestTemplateEditorIntegration(t *testing.T) {
	database := setupEditorTestDB(t)
	defer database.Close()

	templateRepo := repositories.NewTemplateRepository(database)
	userRepo := repositories.NewUserRepository(database)
	editorService := templates.NewEditorService(templateRepo)
	handlers := NewTemplateEditorHandlers(editorService)

	adminUser := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	ctx := auth.WithUser(context.Background(), adminUser)

	t.Run("full CRUD workflow", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
			Metadata: models.ConfigMetadata{
				Name:        "Integration Test Template",
				Category:    "card",
				Description: "Test template for integration testing",
			},
			Layout: models.LayoutConfig{
				Mode:            "card",
				CardWidth:       "800px",
				BackgroundColor: "#ffffff",
			},
			Components: []models.Component{
				{
					ID:   "title-text",
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
					Content: &models.ComponentContent{
						TextBox: &models.TextBoxContent{
							Text:      "{{.Event.Title}}",
							TextAlign: "center",
							FontSize:  "48px",
						},
					},
				},
			},
		}

		configJSON, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}
		configStr := string(configJSON)

		template := &models.Template{
			Name:            "Integration Test Template",
			Type:            models.TemplateTypeRSVPPage,
			Description:     "Test template",
			HTMLContent:     "<html></html>",
			IsDefault:       false,
			IsActive:        true,
			CreatedBy:       adminUser.ID,
			Category:        models.CategoryCard,
			ComponentConfig: &configStr,
		}

		if err := templateRepo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/templates/"+strconv.FormatInt(template.ID, 10)+"/components", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", strconv.FormatInt(template.ID, 10))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.GetComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var getResponse ComponentConfigResponse
		if err := json.NewDecoder(w.Body).Decode(&getResponse); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if getResponse.ComponentConfig == nil {
			t.Fatal("Expected component config, got nil")
		}

		if len(getResponse.ComponentConfig.Components) != 1 {
			t.Errorf("Expected 1 component, got %d", len(getResponse.ComponentConfig.Components))
		}

		updateReq := UpdateComponentsRequest{
			Components: []models.Component{
				{
					ID:   "title-text",
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
					Content: &models.ComponentContent{
						TextBox: &models.TextBoxContent{
							Text:      "{{.Event.Title}}",
							TextAlign: "center",
							FontSize:  "56px",
						},
					},
				},
				{
					ID:   "subtitle-text",
					Type: models.ComponentTypeTextBox,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "80%",
						Height: "auto",
					},
					ZIndex:  9,
					Visible: true,
					Content: &models.ComponentContent{
						TextBox: &models.TextBoxContent{
							Text:      "Join us for a celebration",
							TextAlign: "center",
							FontSize:  "24px",
						},
					},
				},
			},
		}

		body, _ := json.Marshal(updateReq)
		updateHTTPReq := httptest.NewRequest(http.MethodPut, "/api/templates/"+strconv.FormatInt(template.ID, 10)+"/components", bytes.NewReader(body))
		updateHTTPReq.Header.Set("Content-Type", "application/json")
		updateHTTPReq.Header.Set("Accept", "application/json")
		updateHTTPReq = updateHTTPReq.WithContext(ctx)

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("id", strconv.FormatInt(template.ID, 10))
		updateHTTPReq = updateHTTPReq.WithContext(context.WithValue(updateHTTPReq.Context(), chi.RouteCtxKey, rctx2))

		w2 := httptest.NewRecorder()
		handlers.UpdateComponents(w2, updateHTTPReq)

		if w2.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w2.Code, w2.Body.String())
		}

		updatedConfig, err := templateRepo.GetComponentConfig(ctx, template.ID)
		if err != nil {
			t.Fatalf("Failed to get updated config: %v", err)
		}

		if len(updatedConfig.Components) != 2 {
			t.Errorf("Expected 2 components after update, got %d", len(updatedConfig.Components))
		}

		textBoxContent, err := updatedConfig.Components[0].GetTextBoxContent()
		if err != nil {
			t.Fatalf("Failed to get TextBox content: %v", err)
		}
		if textBoxContent.FontSize != "56px" {
			t.Errorf("Expected fontSize 56px, got %v", textBoxContent.FontSize)
		}
	})

	t.Run("preview without saving", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
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

		template := &models.Template{
			Name:            "Preview Test Template",
			Type:            models.TemplateTypeRSVPPage,
			Description:     "Test",
			HTMLContent:     "<html></html>",
			IsDefault:       false,
			IsActive:        true,
			CreatedBy:       adminUser.ID,
			Category:        models.CategoryCard,
			ComponentConfig: &configStr,
		}

		if err := templateRepo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		previewReq := PreviewComponentsRequest{
			Updates: []templates.ComponentUpdate{
				{
					ComponentID: "test-component",
					Property:    "zIndex",
					Value:       float64(20),
				},
			},
		}

		body, _ := json.Marshal(previewReq)
		req := httptest.NewRequest(http.MethodPost, "/api/templates/"+strconv.FormatInt(template.ID, 10)+"/components/preview", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", strconv.FormatInt(template.ID, 10))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.PreviewComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var previewResponse PreviewResponse
		if err := json.NewDecoder(w.Body).Decode(&previewResponse); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if previewResponse.Preview.Components[0].ZIndex != 20 {
			t.Errorf("Expected preview zIndex 20, got %d", previewResponse.Preview.Components[0].ZIndex)
		}

		actualConfig, err := templateRepo.GetComponentConfig(ctx, template.ID)
		if err != nil {
			t.Fatalf("Failed to get actual config: %v", err)
		}

		if actualConfig.Components[0].ZIndex != 10 {
			t.Errorf("Expected actual zIndex to remain 10, got %d", actualConfig.Components[0].ZIndex)
		}
	})

	t.Run("validation endpoint", func(t *testing.T) {
		validConfig := &models.ComponentConfiguration{
			Version: "1.0",
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

		configJSON, _ := json.Marshal(validConfig)
		configStr := string(configJSON)

		template := &models.Template{
			Name:            "Validation Test Template",
			Type:            models.TemplateTypeRSVPPage,
			Description:     "Test",
			HTMLContent:     "<html></html>",
			IsDefault:       false,
			IsActive:        true,
			CreatedBy:       adminUser.ID,
			Category:        models.CategoryCard,
			ComponentConfig: &configStr,
		}

		if err := templateRepo.Create(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/templates/"+strconv.FormatInt(template.ID, 10)+"/components/validate", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", strconv.FormatInt(template.ID, 10))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.ValidateComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var validationResponse ValidationResponse
		if err := json.NewDecoder(w.Body).Decode(&validationResponse); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !validationResponse.Valid {
			t.Errorf("Expected valid to be true, got false with errors: %v", validationResponse.Errors)
		}
	})
}
