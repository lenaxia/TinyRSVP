package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

type mockTemplateService struct {
	CreateTemplateFunc       func(ctx context.Context, template *models.Template) error
	GetTemplateFunc          func(ctx context.Context, id int64) (*models.Template, error)
	GetTemplateForEventFunc  func(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error)
	GetDefaultTemplateFunc   func(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
	UpdateTemplateFunc       func(ctx context.Context, template *models.Template) error
	DeleteTemplateFunc       func(ctx context.Context, id int64) error
	SetActiveFunc            func(ctx context.Context, id int64, active bool) error
	SetDefaultFunc           func(ctx context.Context, id int64) error
	ListTemplatesFunc        func(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error)
	PreviewTemplateFunc      func(ctx context.Context, req *templates.PreviewRequest) (*templates.PreviewResponse, error)
	GetComponentRendererFunc func() *templates.ComponentRenderer
	RenderRSVPPageFunc       func(w io.Writer, event *models.Event, template *models.Template) error
}

func (m *mockTemplateService) CreateTemplate(ctx context.Context, template *models.Template) error {
	if m.CreateTemplateFunc != nil {
		return m.CreateTemplateFunc(ctx, template)
	}
	template.ID = 1
	return nil
}

func (m *mockTemplateService) GetTemplate(ctx context.Context, id int64) (*models.Template, error) {
	if m.GetTemplateFunc != nil {
		return m.GetTemplateFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "Template", ID: id}
}

func (m *mockTemplateService) GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	if m.GetTemplateForEventFunc != nil {
		return m.GetTemplateForEventFunc(ctx, eventID, templateType)
	}
	return nil, &models.NotFoundError{Resource: "Template"}
}

func (m *mockTemplateService) GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	if m.GetDefaultTemplateFunc != nil {
		return m.GetDefaultTemplateFunc(ctx, templateType)
	}
	return nil, &models.NotFoundError{Resource: "Template"}
}

func (m *mockTemplateService) UpdateTemplate(ctx context.Context, template *models.Template) error {
	if m.UpdateTemplateFunc != nil {
		return m.UpdateTemplateFunc(ctx, template)
	}
	return nil
}

func (m *mockTemplateService) DeleteTemplate(ctx context.Context, id int64) error {
	if m.DeleteTemplateFunc != nil {
		return m.DeleteTemplateFunc(ctx, id)
	}
	return nil
}

func (m *mockTemplateService) SetActive(ctx context.Context, id int64, active bool) error {
	if m.SetActiveFunc != nil {
		return m.SetActiveFunc(ctx, id, active)
	}
	return nil
}

func (m *mockTemplateService) SetDefault(ctx context.Context, id int64) error {
	if m.SetDefaultFunc != nil {
		return m.SetDefaultFunc(ctx, id)
	}
	return nil
}

func (m *mockTemplateService) ListTemplates(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc(ctx, filters)
	}
	return []*models.Template{}, nil
}

func (m *mockTemplateService) PreviewTemplate(ctx context.Context, req *templates.PreviewRequest) (*templates.PreviewResponse, error) {
	if m.PreviewTemplateFunc != nil {
		return m.PreviewTemplateFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockTemplateService) GetComponentRenderer() *templates.ComponentRenderer {
	if m.GetComponentRendererFunc != nil {
		return m.GetComponentRendererFunc()
	}
	return nil
}

func (m *mockTemplateService) RenderRSVPPage(w io.Writer, event *models.Event, template *models.Template) error {
	if m.RenderRSVPPageFunc != nil {
		return m.RenderRSVPPageFunc(w, event, template)
	}
	return nil
}

func TestTemplateHandlers_CreateTemplate(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		user           *models.User
		serviceErr     error
		wantStatus     int
		wantErrMessage string
	}{
		{
			name: "valid template",
			body: map[string]interface{}{
				"name":         "Custom Invite",
				"type":         "invite_email",
				"description":  "My custom template",
				"html_content": "<h1>{{.Event.Title}}</h1>",
				"text_content": "{{.Event.Title}}",
			},
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			wantStatus: http.StatusCreated,
		},
		{
			name: "unauthorized",
			body: map[string]interface{}{
				"name":         "Custom Invite",
				"type":         "invite_email",
				"html_content": "<h1>Test</h1>",
			},
			user:           nil,
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "authentication required",
		},
		{
			name:           "invalid JSON",
			body:           "invalid json",
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid request body",
		},
		{
			name: "validation error",
			body: map[string]interface{}{
				"name":         "",
				"type":         "invite_email",
				"html_content": "<h1>Test</h1>",
			},
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			serviceErr:     &models.ValidationError{Field: "name", Message: "name is required"},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				CreateTemplateFunc: func(ctx context.Context, template *models.Template) error {
					return tt.serviceErr
				},
			}

			handler := NewTemplateHandlers(mockService)

			var body []byte
			var err error
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/templates", bytes.NewReader(body))
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Accept", "application/json")
			if tt.user != nil {
				req = req.WithContext(auth.WithUser(context.Background(), tt.user))
			}

			w := httptest.NewRecorder()
			handler.CreateTemplate(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrMessage != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if errMsg, ok := resp["message"].(string); !ok || !containsStr(errMsg, tt.wantErrMessage) {
					t.Errorf("Error message = %v, want to contain %s", errMsg, tt.wantErrMessage)
				}
			}
		})
	}
}

func TestTemplateHandlers_GetTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		user           *models.User
		template       *models.Template
		serviceErr     error
		wantStatus     int
		wantErrMessage string
	}{
		{
			name:       "get existing template",
			templateID: "1",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			template: &models.Template{
				ID:   1,
				Name: "Test Template",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "invalid template ID",
			templateID:     "invalid",
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid template ID",
		},
		{
			name:           "template not found",
			templateID:     "99999",
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			serviceErr:     &models.NotFoundError{Resource: "Template", ID: int64(99999)},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "Template not found",
		},
		{
			name:           "unauthorized",
			templateID:     "1",
			user:           nil,
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					if tt.serviceErr != nil {
						return nil, tt.serviceErr
					}
					return tt.template, nil
				},
			}

			handler := NewTemplateHandlers(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/templates/"+tt.templateID, nil)
			req.Header.Set("Accept", "application/json")
			if tt.user != nil {
				req = req.WithContext(auth.WithUser(context.Background(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.templateID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.GetTemplate(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrMessage != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if errMsg, ok := resp["message"].(string); !ok || !containsStr(errMsg, tt.wantErrMessage) {
					t.Errorf("Error message = %v, want to contain %s", errMsg, tt.wantErrMessage)
				}
			}
		})
	}
}

func TestTemplateHandlers_UpdateTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		body           interface{}
		user           *models.User
		existing       *models.Template
		getErr         error
		updateErr      error
		wantStatus     int
		wantErrMessage string
	}{
		{
			name:       "valid update",
			templateID: "1",
			body: map[string]interface{}{
				"name":         "Updated Template",
				"html_content": "<h1>Updated</h1>",
			},
			user: &models.User{ID: 1, Role: models.RoleEventManager},
			existing: &models.Template{
				ID:          1,
				Name:        "Original",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Original</h1>",
				CreatedBy:   1,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "invalid template ID",
			templateID:     "invalid",
			body:           map[string]interface{}{},
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid template ID",
		},
		{
			name:       "forbidden",
			templateID: "1",
			body:       map[string]interface{}{"name": "Updated"},
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			existing: &models.Template{
				ID:        1,
				CreatedBy: 1,
			},
			updateErr:      &models.ForbiddenError{Message: "You can only edit your own templates"},
			wantStatus:     http.StatusForbidden,
			wantErrMessage: "only edit your own",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				GetTemplateFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					if tt.existing != nil {
						return tt.existing, nil
					}
					return nil, &models.NotFoundError{Resource: "Template", ID: id}
				},
				UpdateTemplateFunc: func(ctx context.Context, template *models.Template) error {
					return tt.updateErr
				},
			}

			handler := NewTemplateHandlers(mockService)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/templates/"+tt.templateID, bytes.NewReader(body))
			req.Header.Set("Accept", "application/json")
			if tt.user != nil {
				req = req.WithContext(auth.WithUser(context.Background(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.templateID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.UpdateTemplate(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrMessage != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if errMsg, ok := resp["message"].(string); !ok || !containsStr(errMsg, tt.wantErrMessage) {
					t.Errorf("Error message = %v, want to contain %s", errMsg, tt.wantErrMessage)
				}
			}
		})
	}
}

func TestTemplateHandlers_DeleteTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		user           *models.User
		serviceErr     error
		wantStatus     int
		wantErrMessage string
	}{
		{
			name:       "successful delete",
			templateID: "1",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			wantStatus: http.StatusNoContent,
		},
		{
			name:           "invalid template ID",
			templateID:     "invalid",
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid template ID",
		},
		{
			name:           "template not found",
			templateID:     "99999",
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			serviceErr:     &models.NotFoundError{Resource: "Template", ID: int64(99999)},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "Template not found",
		},
		{
			name:           "forbidden",
			templateID:     "1",
			user:           &models.User{ID: 1, Role: models.RoleEventManager},
			serviceErr:     &models.ForbiddenError{Message: "You can only delete your own templates"},
			wantStatus:     http.StatusForbidden,
			wantErrMessage: "only delete your own",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				DeleteTemplateFunc: func(ctx context.Context, id int64) error {
					return tt.serviceErr
				},
			}

			handler := NewTemplateHandlers(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/templates/"+tt.templateID, nil)
			req.Header.Set("Accept", "application/json")
			if tt.user != nil {
				req = req.WithContext(auth.WithUser(context.Background(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.templateID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.DeleteTemplate(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrMessage != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if errMsg, ok := resp["message"].(string); !ok || !containsStr(errMsg, tt.wantErrMessage) {
					t.Errorf("Error message = %v, want to contain %s", errMsg, tt.wantErrMessage)
				}
			}
		})
	}
}

func TestTemplateHandlers_ListTemplates(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		user       *models.User
		templates  []*models.Template
		wantStatus int
		wantCount  int
	}{
		{
			name:  "list templates",
			query: "",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			templates: []*models.Template{
				{ID: 1, Name: "Template 1"},
				{ID: 2, Name: "Template 2"},
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "empty list",
			query:      "",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			templates:  []*models.Template{},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:       "filter by type",
			query:      "?type=invite_email",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			templates:  []*models.Template{{ID: 1, Type: models.TemplateTypeInviteEmail}},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTemplateService{
				ListTemplatesFunc: func(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
					return tt.templates, nil
				},
			}

			handler := NewTemplateHandlers(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/templates"+tt.query, nil)
			req.Header.Set("Accept", "application/json")
			if tt.user != nil {
				req = req.WithContext(auth.WithUser(context.Background(), tt.user))
			}

			w := httptest.NewRecorder()
			handler.ListTemplates(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				templates := resp["templates"].([]interface{})
				if len(templates) != tt.wantCount {
					t.Errorf("Template count = %d, want %d", len(templates), tt.wantCount)
				}
			}
		})
	}
}

func containsStr(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPreviewTemplate_Success(t *testing.T) {
	mockService := &mockTemplateService{
		PreviewTemplateFunc: func(ctx context.Context, req *templates.PreviewRequest) (*templates.PreviewResponse, error) {
			return &templates.PreviewResponse{
				HTMLPreview: "<h1>Sample Event</h1>",
				TextPreview: "Sample Event",
			}, nil
		},
	}

	handlers := NewTemplateHandlers(mockService)

	reqBody := `{
		"type": "invite_email",
		"html_content": "<h1>{{.Event.Title}}</h1>",
		"text_content": "{{.Event.Title}}"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/templates/preview", strings.NewReader(reqBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(context.Background(), user))

	w := httptest.NewRecorder()
	handlers.PreviewTemplate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["html_preview"] == nil {
		t.Error("Expected html_preview in response")
	}

	if resp["text_preview"] == nil {
		t.Error("Expected text_preview in response")
	}
}

func TestPreviewTemplate_InvalidJSON(t *testing.T) {
	mockService := &mockTemplateService{}
	handlers := NewTemplateHandlers(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/templates/preview", strings.NewReader("invalid json"))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(context.Background(), user))

	w := httptest.NewRecorder()
	handlers.PreviewTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPreviewTemplate_ValidationError(t *testing.T) {
	mockService := &mockTemplateService{
		PreviewTemplateFunc: func(ctx context.Context, req *templates.PreviewRequest) (*templates.PreviewResponse, error) {
			return nil, &models.ValidationError{
				Field:   "html_content",
				Message: "Template syntax error",
			}
		},
	}

	handlers := NewTemplateHandlers(mockService)

	reqBody := `{
		"type": "invite_email",
		"html_content": "{{.Event.Title"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/templates/preview", strings.NewReader(reqBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(context.Background(), user))

	w := httptest.NewRecorder()
	handlers.PreviewTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPreviewTemplate_Unauthorized(t *testing.T) {
	mockService := &mockTemplateService{}
	handlers := NewTemplateHandlers(mockService)

	reqBody := `{
		"type": "invite_email",
		"html_content": "<h1>{{.Event.Title}}</h1>"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/templates/preview", strings.NewReader(reqBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	w := httptest.NewRecorder()
	handlers.PreviewTemplate(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
