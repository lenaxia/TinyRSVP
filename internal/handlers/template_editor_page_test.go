package handlers

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

type mockEditorPageService struct {
	getEditableTemplateFunc func(ctx context.Context, id int64) (*templates.EditableTemplate, error)
}

func (m *mockEditorPageService) GetEditableTemplate(ctx context.Context, id int64) (*templates.EditableTemplate, error) {
	if m.getEditableTemplateFunc != nil {
		return m.getEditableTemplateFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockEditorPageService) UpdateComponents(ctx context.Context, templateID int64, components []models.Component) error {
	return errors.New("not implemented")
}

func (m *mockEditorPageService) PreviewChanges(ctx context.Context, templateID int64, changes *templates.ComponentChanges) (*models.ComponentConfiguration, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEditorPageService) AddComponent(ctx context.Context, templateID int64, component models.Component) error {
	return errors.New("not implemented")
}

func (m *mockEditorPageService) RemoveComponent(ctx context.Context, templateID int64, componentID string) error {
	return errors.New("not implemented")
}

func (m *mockEditorPageService) UpdateComponentProperty(ctx context.Context, templateID int64, componentID string, property string, value interface{}) error {
	return errors.New("not implemented")
}

func (m *mockEditorPageService) ReorderComponents(ctx context.Context, templateID int64, componentIDs []string) error {
	return errors.New("not implemented")
}

func TestGetEditorPage_Success(t *testing.T) {
	template := &models.Template{
		ID:          1,
		Name:        "Test Template",
		Type:        models.TemplateTypeRSVPPage,
		Category:    models.CategoryCard,
		Description: "Test description",
	}

	componentConfig := &models.ComponentConfiguration{
		Version: "1.0",
		Metadata: models.ConfigMetadata{
			Name:        "Test Template",
			Category:    "card",
			Description: "Test description",
		},
		Components: []models.Component{
			{
				ID:      "title-text",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    testutil.StringPtr("50%"),
					Y:    testutil.StringPtr("200px"),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
			},
		},
	}

	service := &mockEditorPageService{
		getEditableTemplateFunc: func(ctx context.Context, id int64) (*templates.EditableTemplate, error) {
			if id == 1 {
				return &templates.EditableTemplate{
					Template:        template,
					ComponentConfig: componentConfig,
				}, nil
			}
			return nil, &models.NotFoundError{Resource: "template"}
		},
	}

	handler := NewTemplateEditorHandlers(service)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/templates/1/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestGetEditorPage_Unauthorized(t *testing.T) {
	service := &mockEditorPageService{}
	handler := NewTemplateEditorHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/templates/1/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetEditorPage_InvalidID(t *testing.T) {
	service := &mockEditorPageService{}
	handler := NewTemplateEditorHandlers(service)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/templates/invalid/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetEditorPage_NotFound(t *testing.T) {
	service := &mockEditorPageService{
		getEditableTemplateFunc: func(ctx context.Context, id int64) (*templates.EditableTemplate, error) {
			return nil, &models.NotFoundError{Resource: "template"}
		},
	}

	handler := NewTemplateEditorHandlers(service)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/templates/999/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetEditorPage_ServiceError(t *testing.T) {
	service := &mockEditorPageService{
		getEditableTemplateFunc: func(ctx context.Context, id int64) (*templates.EditableTemplate, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewTemplateEditorHandlers(service)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/templates/1/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetEditorPage_ZeroID(t *testing.T) {
	service := &mockEditorPageService{}
	handler := NewTemplateEditorHandlers(service)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/templates/0/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetEditorPage_NegativeID(t *testing.T) {
	service := &mockEditorPageService{}
	handler := NewTemplateEditorHandlers(service)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/templates/-1/edit", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetEditorPage_ProductionTemplateSet_RendersEditorPage proves that the
// production template set (as constructed in main.go) contains
// "template_editor.html" and renders it without error. This is the regression
// test for the dead-code bug — verifying wiring is complete end-to-end.
func TestGetEditorPage_ProductionTemplateSet_RendersEditorPage(t *testing.T) {
	// Mirror the ParseFiles call from cmd/server/main.go exactly.
	tmpl, err := template.New("template_editor.html").ParseFiles(
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
		"../../templates/web/template_editor.html",
	)
	if err != nil {
		t.Fatalf("Failed to parse production template set: %v", err)
	}

	tmplModel := &models.Template{
		ID:   1,
		Name: "Test Template",
		Type: models.TemplateTypeRSVPPage,
	}

	service := &mockEditorPageService{
		getEditableTemplateFunc: func(ctx context.Context, id int64) (*templates.EditableTemplate, error) {
			return &templates.EditableTemplate{
				Template:        tmplModel,
				ComponentConfig: nil,
			}, nil
		},
	}

	handler := NewTemplateEditorHandlers(service)
	handler.SetTemplates(tmpl)

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}

	req := httptest.NewRequest(http.MethodGet, "/templates/1/edit", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handler.GetEditorPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()

	if strings.Contains(body, "Failed to render page") {
		t.Error("Response must not contain render error fallback")
	}
	if strings.Contains(body, "Template engine not initialized") {
		t.Error("Response must not contain inline fallback — SetTemplates not called")
	}
	if !strings.Contains(body, "Test Template") {
		t.Errorf("Expected template name in response body, got: %s", body)
	}
}
