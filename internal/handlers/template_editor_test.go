package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

type mockEditorService struct {
	GetEditableTemplateFunc     func(ctx context.Context, templateID int64) (*templates.EditableTemplate, error)
	UpdateComponentsFunc        func(ctx context.Context, templateID int64, components []models.Component) error
	AddComponentFunc            func(ctx context.Context, templateID int64, component models.Component) error
	RemoveComponentFunc         func(ctx context.Context, templateID int64, componentID string) error
	UpdateComponentPropertyFunc func(ctx context.Context, templateID int64, componentID string, property string, value interface{}) error
	ReorderComponentsFunc       func(ctx context.Context, templateID int64, componentIDs []string) error
	PreviewChangesFunc          func(ctx context.Context, templateID int64, changes *templates.ComponentChanges) (*models.ComponentConfiguration, error)
}

func (m *mockEditorService) GetEditableTemplate(ctx context.Context, templateID int64) (*templates.EditableTemplate, error) {
	if m.GetEditableTemplateFunc != nil {
		return m.GetEditableTemplateFunc(ctx, templateID)
	}
	return nil, nil
}

func (m *mockEditorService) UpdateComponents(ctx context.Context, templateID int64, components []models.Component) error {
	if m.UpdateComponentsFunc != nil {
		return m.UpdateComponentsFunc(ctx, templateID, components)
	}
	return nil
}

func (m *mockEditorService) AddComponent(ctx context.Context, templateID int64, component models.Component) error {
	if m.AddComponentFunc != nil {
		return m.AddComponentFunc(ctx, templateID, component)
	}
	return nil
}

func (m *mockEditorService) RemoveComponent(ctx context.Context, templateID int64, componentID string) error {
	if m.RemoveComponentFunc != nil {
		return m.RemoveComponentFunc(ctx, templateID, componentID)
	}
	return nil
}

func (m *mockEditorService) UpdateComponentProperty(ctx context.Context, templateID int64, componentID string, property string, value interface{}) error {
	if m.UpdateComponentPropertyFunc != nil {
		return m.UpdateComponentPropertyFunc(ctx, templateID, componentID, property, value)
	}
	return nil
}

func (m *mockEditorService) ReorderComponents(ctx context.Context, templateID int64, componentIDs []string) error {
	if m.ReorderComponentsFunc != nil {
		return m.ReorderComponentsFunc(ctx, templateID, componentIDs)
	}
	return nil
}

func (m *mockEditorService) PreviewChanges(ctx context.Context, templateID int64, changes *templates.ComponentChanges) (*models.ComponentConfiguration, error) {
	if m.PreviewChangesFunc != nil {
		return m.PreviewChangesFunc(ctx, templateID, changes)
	}
	return nil, nil
}

func TestTemplateEditorHandlers_GetComponents(t *testing.T) {
	mockService := &mockEditorService{}
	handlers := NewTemplateEditorHandlers(mockService)

	t.Run("returns component config successfully", func(t *testing.T) {
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

		mockService.GetEditableTemplateFunc = func(ctx context.Context, templateID int64) (*templates.EditableTemplate, error) {
			return &templates.EditableTemplate{
				Template: &models.Template{
					ID:   templateID,
					Name: "Test Template",
				},
				ComponentConfig: config,
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/templates/1/components", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.GetComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["component_config"] == nil {
			t.Error("Expected component_config in response")
		}
	})

	t.Run("requires authentication", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/templates/1/components", nil)
		req.Header.Set("Accept", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.GetComponents(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid template ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/templates/invalid/components", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.GetComponents(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestTemplateEditorHandlers_UpdateComponents(t *testing.T) {
	mockService := &mockEditorService{}
	handlers := NewTemplateEditorHandlers(mockService)

	t.Run("updates components successfully", func(t *testing.T) {
		mockService.UpdateComponentsFunc = func(ctx context.Context, templateID int64, components []models.Component) error {
			return nil
		}

		reqBody := map[string]interface{}{
			"components": []map[string]interface{}{
				{
					"id":   "new-component",
					"type": "TextBox",
					"position": map[string]interface{}{
						"mode": "absolute",
						"x":    "50%",
						"y":    "100px",
					},
					"dimensions": map[string]interface{}{
						"width":  "80%",
						"height": "auto",
					},
					"zIndex":  10,
					"visible": true,
				},
			},
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/templates/1/components", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.UpdateComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("requires authentication", func(t *testing.T) {
		reqBody := map[string]interface{}{"components": []interface{}{}}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/templates/1/components", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.UpdateComponents(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/templates/1/components", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.UpdateComponents(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestTemplateEditorHandlers_PreviewComponents(t *testing.T) {
	mockService := &mockEditorService{}
	handlers := NewTemplateEditorHandlers(mockService)

	t.Run("generates preview successfully", func(t *testing.T) {
		previewConfig := &models.ComponentConfiguration{
			Version: "1.0",
			Components: []models.Component{
				{
					ID:     "test-component",
					Type:   models.ComponentTypeTextBox,
					ZIndex: 20,
					Position: models.Position{
						Mode: models.PositionModeAbsolute,
					},
					Dimensions: models.Dimensions{
						Width:  "100%",
						Height: "auto",
					},
					Visible: true,
				},
			},
		}

		mockService.PreviewChangesFunc = func(ctx context.Context, templateID int64, changes *templates.ComponentChanges) (*models.ComponentConfiguration, error) {
			return previewConfig, nil
		}

		reqBody := map[string]interface{}{
			"updates": []map[string]interface{}{
				{
					"component_id": "test-component",
					"property":     "zIndex",
					"value":        20,
				},
			},
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/templates/1/components/preview", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.PreviewComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["preview"] == nil {
			t.Error("Expected preview in response")
		}
	})
}

func TestTemplateEditorHandlers_ValidateComponents(t *testing.T) {
	mockService := &mockEditorService{}
	handlers := NewTemplateEditorHandlers(mockService)

	t.Run("validates components successfully", func(t *testing.T) {
		config := &models.ComponentConfiguration{
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

		mockService.GetEditableTemplateFunc = func(ctx context.Context, templateID int64) (*templates.EditableTemplate, error) {
			return &templates.EditableTemplate{
				Template: &models.Template{
					ID:   templateID,
					Name: "Test",
				},
				ComponentConfig: config,
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/templates/1/components/validate", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.ValidateComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if valid, ok := response["valid"].(bool); !ok || !valid {
			t.Error("Expected valid to be true")
		}
	})

	t.Run("returns validation errors", func(t *testing.T) {
		config := &models.ComponentConfiguration{
			Version: "1.0",
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

		mockService.GetEditableTemplateFunc = func(ctx context.Context, templateID int64) (*templates.EditableTemplate, error) {
			return &templates.EditableTemplate{
				Template: &models.Template{
					ID:   templateID,
					Name: "Test",
				},
				ComponentConfig: config,
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/templates/1/components/validate", nil)
		req.Header.Set("Accept", "application/json")
		req = req.WithContext(auth.WithUser(context.Background(), &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handlers.ValidateComponents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if valid, ok := response["valid"].(bool); !ok || valid {
			t.Error("Expected valid to be false")
		}

		if response["errors"] == nil {
			t.Error("Expected errors in response")
		}
	})
}
