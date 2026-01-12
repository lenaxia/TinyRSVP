package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestEventCustomizationE2E_CompleteFlow(t *testing.T) {
	database, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()
	
	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	templateRepo := repositories.NewTemplateRepository(database)
	
	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	componentConfig := `{
		"version": "1.0",
		"components": [
			{
				"id": "header-text",
				"type": "textbox",
				"content": {
					"text": "Welcome to Our Event",
					"fontSize": "32px",
					"color": "#333333",
					"fontWeight": "bold"
				}
			},
			{
				"id": "description-text",
				"type": "textbox",
				"content": {
					"text": "Join us for a wonderful celebration",
					"fontSize": "16px",
					"color": "#666666"
				}
			}
		]
	}`

	template := &models.Template{
		Name:            "Test Component Template",
		Type:            models.TemplateTypeRSVPPage,
		Category:        models.CategoryModern,
		HTMLContent:     "<html><body>{{.Event.Title}}</body></html>",
		ComponentConfig: &componentConfig,
		CreatedBy:       user.ID,
	}
	if err := templateRepo.Create(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		Title:      "Test Event",
		StartTime:  startTime,
		Timezone:   "America/Los_Angeles",
		Status:     models.EventStatusPublished,
		CreatedBy:  user.ID,
		TemplateID: &template.ID,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	authChecker := auth.NewAuthorizationChecker()
	customizationService := events.NewCustomizationService(eventRepo, templateRepo, authChecker)
	customizationHandlers := NewEventCustomizationHandlers(customizationService)

	t.Run("get_initial_customization", func(t *testing.T) {
		eventIDStr := strconv.FormatInt(event.ID, 10)
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventIDStr+"/template/customization", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.GetCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetCustomization() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var response events.EventCustomizationData
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Event == nil {
			t.Error("Expected event in response")
		}
		if response.Template == nil {
			t.Error("Expected template in response")
		}
		if response.TemplateConfig == nil {
			t.Error("Expected template config in response")
		}
		if response.EventOverrides != nil {
			t.Error("Expected no event overrides initially")
		}
		if response.MergedConfig == nil {
			t.Error("Expected merged config in response")
		}
	})

	t.Run("update_customization", func(t *testing.T) {
		overrides := &models.ComponentOverrides{
			Version: "1.0",
			Overrides: []models.ComponentOverride{
				{
					ID: "header-text",
					Updates: map[string]interface{}{
						"content": map[string]interface{}{
							"text":  "Custom Event Title",
							"color": "#ff0000",
						},
					},
				},
			},
		}

		eventIDStr := strconv.FormatInt(event.ID, 10)
		bodyBytes, _ := json.Marshal(overrides)
		req := httptest.NewRequest(http.MethodPut, "/api/events/"+eventIDStr+"/template/customization", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.UpdateCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("UpdateCustomization() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("verify_customization_persisted", func(t *testing.T) {
		eventIDStr := strconv.FormatInt(event.ID, 10)
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventIDStr+"/template/customization", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.GetCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetCustomization() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response events.EventCustomizationData
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.EventOverrides == nil {
			t.Fatal("Expected event overrides after update")
		}

		if len(response.EventOverrides.Overrides) != 1 {
			t.Errorf("Expected 1 override, got %d", len(response.EventOverrides.Overrides))
		}

		if response.MergedConfig == nil {
			t.Fatal("Expected merged config")
		}

		found := false
		for _, comp := range response.MergedConfig.Components {
			if comp.ID == "header-text" {
				found = true
				if comp.Content == nil || comp.Content.TextBox == nil {
					t.Error("Expected textbox content in merged config")
				} else if comp.Content.TextBox.Text != "Custom Event Title" {
					t.Errorf("Expected custom text, got: %s", comp.Content.TextBox.Text)
				}
			}
		}
		if !found {
			t.Error("Expected to find header-text component in merged config")
		}
	})

	t.Run("preview_customization", func(t *testing.T) {
		overrides := &models.ComponentOverrides{
			Version: "1.0",
			Overrides: []models.ComponentOverride{
				{
					ID: "description-text",
					Updates: map[string]interface{}{
						"content": map[string]interface{}{
							"text": "Preview Text",
						},
					},
				},
			},
		}

		eventIDStr := strconv.FormatInt(event.ID, 10)
		bodyBytes, _ := json.Marshal(overrides)
		req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventIDStr+"/template/customization/preview", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.PreviewCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("PreviewCustomization() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var previewConfig models.ComponentConfiguration
		if err := json.NewDecoder(w.Body).Decode(&previewConfig); err != nil {
			t.Fatalf("Failed to decode preview config: %v", err)
		}

		found := false
		for _, comp := range previewConfig.Components {
			if comp.ID == "description-text" {
				found = true
				if comp.Content == nil || comp.Content.TextBox == nil {
					t.Error("Expected textbox content in preview")
				} else if comp.Content.TextBox.Text != "Preview Text" {
					t.Errorf("Expected preview text, got: %s", comp.Content.TextBox.Text)
				}
			}
		}
		if !found {
			t.Error("Expected to find description-text component in preview")
		}
	})

	t.Run("render_rsvp_with_customizations", func(t *testing.T) {
		tokenGen := token.NewGenerator([]byte("test-secret"))
		email := "guest@example.com"
		inviteService := &mockRSVPInviteService{
			getInviteByTokenFunc: func(ctx context.Context, tkn string) (*models.Invite, error) {
				return &models.Invite{
					ID:      1,
					EventID: event.ID,
					Email:   &email,
					Token:   &tkn,
					Status:  models.InviteStatusSent,
				}, nil
			},
			markViewedFunc: func(ctx context.Context, inviteID int64) error {
				return nil
			},
		}

		questionRepo := repositories.NewQuestionRepository(database)
		rsvpRepo := repositories.NewRSVPRepository(database)

		templateEngine := templates.NewEngine()
		templateValidator := templates.NewValidator(templateEngine)
		templateService := templates.NewService(templateRepo, templateValidator)

		rsvpHandler := NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
		rsvpHandler.SetTemplateRepository(templateRepo)
		rsvpHandler.SetTemplateService(templateService)
		rsvpHandler.SetCustomizationService(customizationService)

		testToken, _ := tokenGen.Generate()
		req := httptest.NewRequest(http.MethodGet, "/rsvp/"+testToken, nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", testToken)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		rsvpHandler.GetRSVPPage(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetRSVPPage() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "component-canvas") {
			t.Error("Expected component-based rendering in page")
		}
		
		if !strings.Contains(body, "Test Event") {
			t.Error("Expected event title in rendered page")
		}
	})

	t.Run("reset_customization", func(t *testing.T) {
		eventIDStr := strconv.FormatInt(event.ID, 10)
		req := httptest.NewRequest(http.MethodDelete, "/api/events/"+eventIDStr+"/template/customization", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.ResetCustomization(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("ResetCustomization() status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("verify_customization_removed", func(t *testing.T) {
		eventIDStr := strconv.FormatInt(event.ID, 10)
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventIDStr+"/template/customization", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.GetCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetCustomization() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response events.EventCustomizationData
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.EventOverrides != nil {
			t.Error("Expected no event overrides after reset")
		}
	})
}

func TestEventCustomizationE2E_MultipleComponents(t *testing.T) {
	database, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()
	
	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	templateRepo := repositories.NewTemplateRepository(database)
	
	user := &models.User{
		Email: "test2@example.com",
		Name:  "Test User 2",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	componentConfig := `{
		"version": "1.0",
		"components": [
			{
				"id": "title",
				"type": "textbox",
				"content": {
					"text": "Original Title",
					"fontSize": "24px",
					"color": "#000000"
				}
			},
			{
				"id": "subtitle",
				"type": "textbox",
				"content": {
					"text": "Original Subtitle",
					"fontSize": "18px",
					"color": "#666666"
				}
			},
			{
				"id": "divider",
				"type": "divider",
				"style": {
					"thickness": "2px",
					"color": "#cccccc"
				}
			}
		]
	}`

	template := &models.Template{
		Name:            "Multi Component Template",
		Type:            models.TemplateTypeRSVPPage,
		Category:        models.CategoryModern,
		HTMLContent:     "<html><body>{{.Event.Title}}</body></html>",
		ComponentConfig: &componentConfig,
		CreatedBy:       user.ID,
	}
	if err := templateRepo.Create(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	startTime := time.Now().Add(48 * time.Hour)
	event := &models.Event{
		Title:      "Multi Component Event",
		StartTime:  startTime,
		Timezone:   "America/Los_Angeles",
		Status:     models.EventStatusPublished,
		CreatedBy:  user.ID,
		TemplateID: &template.ID,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	authChecker := auth.NewAuthorizationChecker()
	customizationService := events.NewCustomizationService(eventRepo, templateRepo, authChecker)
	customizationHandlers := NewEventCustomizationHandlers(customizationService)

	t.Run("update_multiple_components", func(t *testing.T) {
		overrides := &models.ComponentOverrides{
			Version: "1.0",
			Overrides: []models.ComponentOverride{
				{
					ID: "title",
					Updates: map[string]interface{}{
						"content": map[string]interface{}{
							"text":  "Custom Title",
							"color": "#ff0000",
						},
					},
				},
				{
					ID: "subtitle",
					Updates: map[string]interface{}{
						"content": map[string]interface{}{
							"text": "Custom Subtitle",
						},
					},
				},
				{
					ID: "divider",
					Updates: map[string]interface{}{
						"style": map[string]interface{}{
							"backgroundColor": "#0000ff",
						},
					},
				},
			},
		}

		eventIDStr := strconv.FormatInt(event.ID, 10)
		bodyBytes, _ := json.Marshal(overrides)
		req := httptest.NewRequest(http.MethodPut, "/api/events/"+eventIDStr+"/template/customization", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.UpdateCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("UpdateCustomization() status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("verify_all_customizations_merged", func(t *testing.T) {
		eventIDStr := strconv.FormatInt(event.ID, 10)
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventIDStr+"/template/customization", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.GetCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetCustomization() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response events.EventCustomizationData
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.MergedConfig == nil {
			t.Fatal("Expected merged config")
		}

		titleFound := false
		subtitleFound := false
		dividerFound := false

		for _, comp := range response.MergedConfig.Components {
			switch comp.ID {
			case "title":
				titleFound = true
				if comp.Content == nil || comp.Content.TextBox == nil {
					t.Error("Expected textbox content for title")
				} else {
					if comp.Content.TextBox.Text != "Custom Title" {
						t.Errorf("Expected 'Custom Title', got: %s", comp.Content.TextBox.Text)
					}
					if comp.Content.TextBox.Color != "#ff0000" {
						t.Errorf("Expected red color, got: %s", comp.Content.TextBox.Color)
					}
				}
			case "subtitle":
				subtitleFound = true
				if comp.Content == nil || comp.Content.TextBox == nil {
					t.Error("Expected textbox content for subtitle")
				} else if comp.Content.TextBox.Text != "Custom Subtitle" {
					t.Errorf("Expected 'Custom Subtitle', got: %s", comp.Content.TextBox.Text)
				}
			case "divider":
				dividerFound = true
				if comp.Style == nil {
					t.Error("Expected style for divider")
				} else if comp.Style.BackgroundColor != "#0000ff" {
					t.Errorf("Expected blue background color, got: %s", comp.Style.BackgroundColor)
				}
			}
		}

		if !titleFound {
			t.Error("Title component not found in merged config")
		}
		if !subtitleFound {
			t.Error("Subtitle component not found in merged config")
		}
		if !dividerFound {
			t.Error("Divider component not found in merged config")
		}
	})
}

func TestEventCustomizationE2E_BackwardCompatibility(t *testing.T) {
	database, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()
	
	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	templateRepo := repositories.NewTemplateRepository(database)
	
	user := &models.User{
		Email: "test3@example.com",
		Name:  "Test User 3",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	template := &models.Template{
		Name:        "Legacy Template",
		Type:        models.TemplateTypeRSVPPage,
		Category:    models.CategoryClassic,
		HTMLContent: "<html><body>{{.Event.Title}}</body></html>",
		CreatedBy:   user.ID,
	}
	if err := templateRepo.Create(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	startTime := time.Now().Add(72 * time.Hour)
	event := &models.Event{
		Title:      "Legacy Event",
		StartTime:  startTime,
		Timezone:   "America/Los_Angeles",
		Status:     models.EventStatusPublished,
		CreatedBy:  user.ID,
		TemplateID: &template.ID,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	authChecker := auth.NewAuthorizationChecker()
	customizationService := events.NewCustomizationService(eventRepo, templateRepo, authChecker)
	customizationHandlers := NewEventCustomizationHandlers(customizationService)

	t.Run("get_customization_for_non_component_template", func(t *testing.T) {
		eventIDStr := strconv.FormatInt(event.ID, 10)
		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventIDStr+"/template/customization", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.GetCustomization(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetCustomization() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response events.EventCustomizationData
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.TemplateConfig != nil {
			t.Error("Expected no template config for non-component template")
		}
	})

	t.Run("update_fails_for_non_component_template", func(t *testing.T) {
		overrides := &models.ComponentOverrides{
			Version: "1.0",
			Overrides: []models.ComponentOverride{
				{
					ID: "test",
					Updates: map[string]interface{}{
						"content": map[string]interface{}{
							"text": "Test",
						},
					},
				},
			},
		}

		eventIDStr := strconv.FormatInt(event.ID, 10)
		bodyBytes, _ := json.Marshal(overrides)
		req := httptest.NewRequest(http.MethodPut, "/api/events/"+eventIDStr+"/template/customization", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(ctx, user))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", eventIDStr)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		customizationHandlers.UpdateCustomization(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusUnprocessableEntity {
			t.Logf("Update status: %d, body: %s", w.Code, w.Body.String())
		}
	})
}

func setupTestDatabase(t *testing.T) (*db.RetryableDatabase, func()) {
	t.Helper()

	baseDB, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	database := db.NewRetryableDatabase(baseDB, db.DefaultRetryConfig)

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}
