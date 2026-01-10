package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

func setupTemplateIntegrationTest(t *testing.T) (db.Database, *TemplateHandlers, *models.User) {
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

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	templateRepo := repositories.NewTemplateRepository(database)
	templateEngine := templates.NewEngine()
	templateValidator := templates.NewValidator(templateEngine)
	templateService := templates.NewService(templateRepo, templateValidator)

	templateHandlers := NewTemplateHandlers(templateService)

	return database, templateHandlers, user
}

func TestTemplateHandlers_FullStackIntegration(t *testing.T) {
	database, templateHandlers, user := setupTemplateIntegrationTest(t)
	defer database.Close()

	ctx := auth.WithUser(context.Background(), user)

	var createdTemplateID int64

	t.Run("POST /api/templates - create template", func(t *testing.T) {
		reqBody := CreateTemplateRequest{
			Name:        "Test Template",
			Type:        string(models.TemplateTypeInviteEmail),
			Description: "Test description",
			HTMLContent: "<h1>{{.Event.Title}}</h1>",
			TextContent: strPtr("{{.Event.Title}}"),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		templateHandlers.CreateTemplate(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		templateData, ok := response["template"].(map[string]interface{})
		if !ok {
			t.Fatal("Response missing template field")
		}

		if templateData["name"] != "Test Template" {
			t.Errorf("Expected name 'Test Template', got %v", templateData["name"])
		}

		if id, ok := templateData["id"].(float64); ok {
			createdTemplateID = int64(id)
		}
	})

	t.Run("GET /api/templates - list templates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		templateHandlers.ListTemplates(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response ListTemplatesResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Templates) == 0 {
			t.Error("Expected at least one template in list")
		}
	})

	t.Run("GET /api/templates/{id} - get template", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template created in previous test")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/templates/%d", createdTemplateID), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", createdTemplateID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.GetTemplate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		templateData, ok := response["template"].(map[string]interface{})
		if !ok {
			t.Fatal("Response missing template field")
		}

		if templateData["name"] != "Test Template" {
			t.Errorf("Expected name 'Test Template', got %v", templateData["name"])
		}
	})

	t.Run("PUT /api/templates/{id} - update template", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template created in previous test")
		}

		updateReq := UpdateTemplateRequest{
			Description: strPtr("Updated description"),
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/templates/%d", createdTemplateID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", createdTemplateID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.UpdateTemplate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		templateData, ok := response["template"].(map[string]interface{})
		if !ok {
			t.Fatal("Response missing template field")
		}

		if templateData["description"] != "Updated description" {
			t.Errorf("Expected description 'Updated description', got %v", templateData["description"])
		}
	})

	t.Run("POST /api/templates/{id}/set-active - set template active/inactive", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template created in previous test")
		}

		setActiveReq := SetActiveRequest{
			Active: false,
		}
		body, _ := json.Marshal(setActiveReq)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/templates/%d/set-active", createdTemplateID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", createdTemplateID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.SetActive(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("DELETE /api/templates/{id} - delete template", func(t *testing.T) {
		template := &models.Template{
			Name:        "Delete Test Template",
			Type:        models.TemplateTypeInviteEmail,
			Description: "To be deleted",
			HTMLContent: "<div>Delete me</div>",
			TextContent: strPtr("Delete me"),
		}
		if err := templateHandlers.service.CreateTemplate(ctx, template); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/templates/%d", template.ID), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", template.ID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.DeleteTemplate(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	})

	t.Run("Unauthenticated request returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)

		w := httptest.NewRecorder()
		templateHandlers.ListTemplates(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestTemplateHandlers_FullStackIntegration_WithRouter(t *testing.T) {
	database, templateHandlers, user := setupTemplateIntegrationTest(t)
	defer database.Close()

	ctx := auth.WithUser(context.Background(), user)

	router := chi.NewRouter()
	templateHandlers.RegisterRoutes(router)

	var createdTemplateID int64

	t.Run("POST /api/templates via router", func(t *testing.T) {
		reqBody := CreateTemplateRequest{
			Name:        "Router Test Template",
			Type:        string(models.TemplateTypeRSVPPage),
			Description: "Testing via router",
			HTMLContent: "<div>{{.Event.Title}}</div>",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		templateData, ok := response["template"].(map[string]interface{})
		if !ok {
			t.Fatal("Response missing template field")
		}

		if id, ok := templateData["id"].(float64); ok {
			createdTemplateID = int64(id)
		}
	})

	t.Run("GET /api/templates/{id} via router", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template created in previous test")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/templates/%d", createdTemplateID), nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("PUT /api/templates/{id} via router", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template created in previous test")
		}

		updateReq := UpdateTemplateRequest{
			Description: strPtr("Updated via router"),
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/templates/%d", createdTemplateID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("DELETE /api/templates/{id} via router", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template created in previous test")
		}

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/templates/%d", createdTemplateID), nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	})

	t.Run("GET /api/templates via router - list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response ListTemplatesResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
	})
}

func TestTemplateHandlers_PermissionEnforcement(t *testing.T) {
	database, templateHandlers, user := setupTemplateIntegrationTest(t)
	defer database.Close()

	ctx := auth.WithUser(context.Background(), user)

	otherUserRepo := repositories.NewUserRepository(database)
	otherUser := &models.User{
		Email: "other@example.com",
		Name:  "Other User",
		Role:  models.RoleEventManager,
	}
	if err := otherUserRepo.Create(context.Background(), otherUser); err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	template := &models.Template{
		Name:        "User Template",
		Type:        models.TemplateTypeInviteEmail,
		Description: "Owned by first user",
		HTMLContent: "<div>Test</div>",
		TextContent: strPtr("Test"),
	}
	if err := templateHandlers.service.CreateTemplate(ctx, template); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	t.Run("cannot update other user's template", func(t *testing.T) {
		otherCtx := auth.WithUser(context.Background(), otherUser)

		updateReq := UpdateTemplateRequest{
			Description: strPtr("Trying to update"),
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/templates/%d", template.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
		req = req.WithContext(otherCtx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", template.ID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.UpdateTemplate(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusForbidden, w.Code, w.Body.String())
		}
	})

	t.Run("cannot delete other user's template", func(t *testing.T) {
		otherCtx := auth.WithUser(context.Background(), otherUser)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/templates/%d", template.ID), nil)
		req = req.WithContext(otherCtx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", template.ID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.DeleteTemplate(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusForbidden, w.Code, w.Body.String())
		}
	})
}

func TestTemplateHandlers_DefaultTemplateProtection(t *testing.T) {
	database, templateHandlers, _ := setupTemplateIntegrationTest(t)
	defer database.Close()

	adminRepo := repositories.NewUserRepository(database)
	admin := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := adminRepo.Create(context.Background(), admin); err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	adminCtx := auth.WithUser(context.Background(), admin)

	defaultTemplate := &models.Template{
		Name:        "Default Template",
		Type:        models.TemplateTypeInviteEmail,
		Description: "System default",
		HTMLContent: "<div>Default</div>",
		TextContent: strPtr("Default"),
		IsDefault:   true,
	}
	if err := templateHandlers.service.CreateTemplate(adminCtx, defaultTemplate); err != nil {
		t.Fatalf("Failed to create default template: %v", err)
	}

	t.Run("cannot delete default template", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/templates/%d", defaultTemplate.ID), nil)
		req = req.WithContext(adminCtx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprintf("%d", defaultTemplate.ID))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		templateHandlers.DeleteTemplate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
		}
	})
}

func strPtr(s string) *string {
	return &s
}
