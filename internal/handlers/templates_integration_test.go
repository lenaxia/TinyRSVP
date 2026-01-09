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

func setupHandlerTestDB(t *testing.T) (db.Database, func()) {
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
		database.Close()
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

var handlerUserCounter int64 = 0

func createHandlerTestUser(t *testing.T, database db.Database, role models.UserRole) *models.User {
	t.Helper()
	handlerUserCounter++

	user := &models.User{
		Email: fmt.Sprintf("handler%d@example.com", handlerUserCounter),
		Name:  fmt.Sprintf("Handler User %d", handlerUserCounter),
		Role:  role,
	}

	query := `
		INSERT INTO users (email, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := database.Exec(context.Background(), query, user.Email, user.Name, user.Role, now, now)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now

	return user
}

func TestTemplateHandlers_Integration_FullCRUDFlow(t *testing.T) {
	database, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	engine := templates.NewEngine()
	validator := templates.NewValidator(engine)
	service := templates.NewService(repo, validator)
	handler := NewTemplateHandlers(service)

	user := createHandlerTestUser(t, database, models.RoleEventManager)

	createReq := map[string]interface{}{
		"name":         "Integration Test Template",
		"type":         "rsvp_page",
		"description":  "Integration test",
		"html_content": "<h1>{{.Event.Title}}</h1>",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
	req = req.WithContext(auth.WithUser(context.Background(), user))

	w := httptest.NewRecorder()
	handler.CreateTemplate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("CreateTemplate status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var createResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResp)
	templateData := createResp["template"].(map[string]interface{})
	templateID := int64(templateData["id"].(float64))

	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/templates/%d", templateID), nil)
	getReq = getReq.WithContext(auth.WithUser(context.Background(), user))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", templateID))
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.GetTemplate(w, getReq)

	if w.Code != http.StatusOK {
		t.Fatalf("GetTemplate status = %d, want %d", w.Code, http.StatusOK)
	}

	updateReq := map[string]interface{}{
		"name": "Updated Integration Template",
	}

	body, _ = json.Marshal(updateReq)
	putReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/templates/%d", templateID), bytes.NewReader(body))
	putReq = putReq.WithContext(auth.WithUser(context.Background(), user))
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", templateID))
	putReq = putReq.WithContext(context.WithValue(putReq.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.UpdateTemplate(w, putReq)

	if w.Code != http.StatusOK {
		t.Fatalf("UpdateTemplate status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/templates/%d", templateID), nil)
	deleteReq = deleteReq.WithContext(auth.WithUser(context.Background(), user))
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", templateID))
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.DeleteTemplate(w, deleteReq)

	if w.Code != http.StatusNoContent {
		t.Fatalf("DeleteTemplate status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestTemplateHandlers_Integration_PermissionEnforcement(t *testing.T) {
	database, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	engine := templates.NewEngine()
	validator := templates.NewValidator(engine)
	service := templates.NewService(repo, validator)
	handler := NewTemplateHandlers(service)

	user1 := createHandlerTestUser(t, database, models.RoleEventManager)
	user2 := createHandlerTestUser(t, database, models.RoleEventManager)

	createReq := map[string]interface{}{
		"name":         "User1 Template",
		"type":         "rsvp_page",
		"html_content": "<h1>{{.Event.Title}}</h1>",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
	req = req.WithContext(auth.WithUser(context.Background(), user1))

	w := httptest.NewRecorder()
	handler.CreateTemplate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("CreateTemplate status = %d, want %d", w.Code, http.StatusCreated)
	}

	var createResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResp)
	templateData := createResp["template"].(map[string]interface{})
	templateID := int64(templateData["id"].(float64))

	updateReq := map[string]interface{}{
		"name": "User2 Trying to Update",
	}

	body, _ = json.Marshal(updateReq)
	putReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/templates/%d", templateID), bytes.NewReader(body))
	putReq = putReq.WithContext(auth.WithUser(context.Background(), user2))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", templateID))
	putReq = putReq.WithContext(context.WithValue(putReq.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.UpdateTemplate(w, putReq)

	if w.Code != http.StatusForbidden {
		t.Errorf("UpdateTemplate by different user status = %d, want %d", w.Code, http.StatusForbidden)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/templates/%d", templateID), nil)
	deleteReq = deleteReq.WithContext(auth.WithUser(context.Background(), user2))
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", templateID))
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.DeleteTemplate(w, deleteReq)

	if w.Code != http.StatusForbidden {
		t.Errorf("DeleteTemplate by different user status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestTemplateHandlers_Integration_ListWithFilters(t *testing.T) {
	database, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := templates.NewSeeder(repo, 1)
	engine := templates.NewEngine()
	validator := templates.NewValidator(engine)
	service := templates.NewService(repo, validator)
	handler := NewTemplateHandlers(service)

	user := createHandlerTestUser(t, database, models.RoleEventManager)

	err := seeder.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	for i := 0; i < 2; i++ {
		createReq := map[string]interface{}{
			"name":         fmt.Sprintf("Custom Template %d", i+1),
			"type":         "rsvp_page",
			"html_content": "<h1>{{.Event.Title}}</h1>",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
		req = req.WithContext(auth.WithUser(context.Background(), user))

		w := httptest.NewRecorder()
		handler.CreateTemplate(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("CreateTemplate status = %d, want %d", w.Code, http.StatusCreated)
		}
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/templates?type=rsvp_page", nil)
	listReq = listReq.WithContext(auth.WithUser(context.Background(), user))

	w := httptest.NewRecorder()
	handler.ListTemplates(w, listReq)

	if w.Code != http.StatusOK {
		t.Fatalf("ListTemplates status = %d, want %d", w.Code, http.StatusOK)
	}

	var listResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&listResp)
	templatesList := listResp["templates"].([]interface{})

	if len(templatesList) < 3 {
		t.Errorf("Expected at least 3 RSVP templates, got %d", len(templatesList))
	}
}
