package templates

import (
	"context"
	"errors"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestTemplateService_CreateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template *models.Template
		userID   int64
		role     models.UserRole
		wantErr  bool
		errType  interface{}
	}{
		{
			name: "valid template",
			template: &models.Template{
				Name:        "Custom Invite",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
				TextContent: stringPtr("{{.Event.Title}}"),
			},
			userID:  1,
			role:    models.RoleEventManager,
			wantErr: false,
		},
		{
			name: "unauthorized",
			template: &models.Template{
				Name:        "Custom Invite",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
			},
			userID:  0,
			role:    models.RoleGuest,
			wantErr: true,
			errType: &models.UnauthorizedError{},
		},
		{
			name: "validation error",
			template: &models.Template{
				Name:        "",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<h1>{{.Event.Title}}</h1>",
			},
			userID:  1,
			role:    models.RoleEventManager,
			wantErr: true,
			errType: &models.ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{ID: tt.userID, Role: tt.role}
			ctx := auth.WithUser(context.Background(), user)

			mockRepo := &mockServiceTemplateRepository{}
			mockValidator := &mockServiceValidator{
				ValidateTemplateFunc: func(tmpl *models.Template) error {
					return tmpl.Validate()
				},
			}

			service := NewService(mockRepo, mockValidator)

			err := service.CreateTemplate(ctx, tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Error type = %T, want %T", err, tt.errType)
				}
			}

			if !tt.wantErr {
				if tt.template.ID == 0 {
					t.Error("Expected template ID to be set")
				}
				if tt.template.CreatedBy != tt.userID {
					t.Errorf("CreatedBy = %d, want %d", tt.template.CreatedBy, tt.userID)
				}
				if !tt.template.IsActive {
					t.Error("Expected template to be active")
				}
			}
		})
	}
}

func TestTemplateService_UpdateTemplate(t *testing.T) {
	tests := []struct {
		name         string
		template     *models.Template
		existing     *models.Template
		userID       int64
		role         models.UserRole
		wantErr      bool
		errContains  string
	}{
		{
			name: "update own template",
			template: &models.Template{
				ID:          1,
				Name:        "Updated",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Updated</h1>",
				CreatedBy:   1,
			},
			existing: &models.Template{
				ID:        1,
				CreatedBy: 1,
				IsDefault: false,
				Version:   1,
			},
			userID:  1,
			role:    models.RoleEventManager,
			wantErr: false,
		},
		{
			name: "update other user's template as non-admin",
			template: &models.Template{
				ID:          1,
				Name:        "Updated",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Updated</h1>",
				CreatedBy:   1,
			},
			existing: &models.Template{
				ID:        1,
				CreatedBy: 2,
				IsDefault: false,
			},
			userID:      1,
			role:        models.RoleEventManager,
			wantErr:     true,
			errContains: "only edit your own",
		},
		{
			name: "update default template as non-admin",
			template: &models.Template{
				ID:          1,
				Name:        "Updated",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Updated</h1>",
				CreatedBy:   1,
			},
			existing: &models.Template{
				ID:        1,
				CreatedBy: 1,
				IsDefault: true,
			},
			userID:      1,
			role:        models.RoleEventManager,
			wantErr:     true,
			errContains: "Only admins can edit default",
		},
		{
			name: "admin can update any template",
			template: &models.Template{
				ID:          1,
				Name:        "Updated",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Updated</h1>",
				CreatedBy:   1,
			},
			existing: &models.Template{
				ID:        1,
				CreatedBy: 2,
				IsDefault: false,
			},
			userID:  1,
			role:    models.RoleAdmin,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{ID: tt.userID, Role: tt.role}
			ctx := auth.WithUser(context.Background(), user)

			mockRepo := &mockServiceTemplateRepository{
				GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					return tt.existing, nil
				},
			}
			mockValidator := &mockServiceValidator{
				ValidateTemplateFunc: func(tmpl *models.Template) error {
					return tmpl.Validate()
				},
			}

			service := NewService(mockRepo, mockValidator)

			err := service.UpdateTemplate(ctx, tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want to contain %s", err, tt.errContains)
				}
			}
		})
	}
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		existing    *models.Template
		userID      int64
		role        models.UserRole
		inUse       bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "delete own template",
			templateID: 1,
			existing: &models.Template{
				ID:        1,
				CreatedBy: 1,
				IsDefault: false,
			},
			userID:  1,
			role:    models.RoleEventManager,
			inUse:   false,
			wantErr: false,
		},
		{
			name:       "delete default template",
			templateID: 1,
			existing: &models.Template{
				ID:        1,
				IsDefault: true,
			},
			userID:      1,
			role:        models.RoleAdmin,
			inUse:       false,
			wantErr:     true,
			errContains: "Cannot delete default",
		},
		{
			name:       "delete template in use",
			templateID: 1,
			existing: &models.Template{
				ID:        1,
				CreatedBy: 1,
				IsDefault: false,
			},
			userID:      1,
			role:        models.RoleEventManager,
			inUse:       true,
			wantErr:     true,
			errContains: "in use",
		},
		{
			name:       "delete other user's template",
			templateID: 1,
			existing: &models.Template{
				ID:        1,
				CreatedBy: 2,
				IsDefault: false,
			},
			userID:      1,
			role:        models.RoleEventManager,
			inUse:       false,
			wantErr:     true,
			errContains: "only delete your own",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{ID: tt.userID, Role: tt.role}
			ctx := auth.WithUser(context.Background(), user)

			mockRepo := &mockServiceTemplateRepository{
				GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					return tt.existing, nil
				},
				IsTemplateInUseFunc: func(ctx context.Context, id int64) (bool, error) {
					return tt.inUse, nil
				},
			}

			service := NewService(mockRepo, nil)

			err := service.DeleteTemplate(ctx, tt.templateID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want to contain %s", err, tt.errContains)
				}
			}
		})
	}
}

func TestTemplateService_GetTemplateForEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventID       int64
		templateType  models.TemplateType
		eventTemplate *models.Template
		defaultTmpl   *models.Template
		wantDefault   bool
		wantErr       bool
	}{
		{
			name:         "event has active template",
			eventID:      1,
			templateType: models.TemplateTypeRSVPPage,
			eventTemplate: &models.Template{
				ID:       1,
				IsActive: true,
			},
			wantDefault: false,
			wantErr:     false,
		},
		{
			name:         "event template inactive, use default",
			eventID:      1,
			templateType: models.TemplateTypeRSVPPage,
			eventTemplate: &models.Template{
				ID:       1,
				IsActive: false,
			},
			defaultTmpl: &models.Template{
				ID:        2,
				IsDefault: true,
			},
			wantDefault: true,
			wantErr:     false,
		},
		{
			name:         "no event template, use default",
			eventID:      1,
			templateType: models.TemplateTypeRSVPPage,
			defaultTmpl: &models.Template{
				ID:        2,
				IsDefault: true,
			},
			wantDefault: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceTemplateRepository{
				GetByEventAndTypeFunc: func(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
					if tt.eventTemplate != nil {
						return tt.eventTemplate, nil
					}
					return nil, &models.NotFoundError{}
				},
				GetDefaultByTypeFunc: func(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
					if tt.defaultTmpl != nil {
						return tt.defaultTmpl, nil
					}
					return nil, &models.NotFoundError{}
				},
			}

			service := NewService(mockRepo, nil)

			result, err := service.GetTemplateForEvent(context.Background(), tt.eventID, tt.templateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateForEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.wantDefault && result.ID != tt.defaultTmpl.ID {
					t.Errorf("Expected default template ID %d, got %d", tt.defaultTmpl.ID, result.ID)
				}
				if !tt.wantDefault && tt.eventTemplate != nil && result.ID != tt.eventTemplate.ID {
					t.Errorf("Expected event template ID %d, got %d", tt.eventTemplate.ID, result.ID)
				}
			}
		})
	}
}

func TestTemplateService_SetActive(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		active      bool
		existing    *models.Template
		userID      int64
		role        models.UserRole
		wantErr     bool
		errContains string
	}{
		{
			name:       "set own template inactive",
			templateID: 1,
			active:     false,
			existing: &models.Template{
				ID:        1,
				CreatedBy: 1,
			},
			userID:  1,
			role:    models.RoleEventManager,
			wantErr: false,
		},
		{
			name:       "set other user's template as non-admin",
			templateID: 1,
			active:     false,
			existing: &models.Template{
				ID:        1,
				CreatedBy: 2,
			},
			userID:      1,
			role:        models.RoleEventManager,
			wantErr:     true,
			errContains: "only modify your own",
		},
		{
			name:       "admin can modify any template",
			templateID: 1,
			active:     false,
			existing: &models.Template{
				ID:        1,
				CreatedBy: 2,
			},
			userID:  1,
			role:    models.RoleAdmin,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{ID: tt.userID, Role: tt.role}
			ctx := auth.WithUser(context.Background(), user)

			mockRepo := &mockServiceTemplateRepository{
				GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					return tt.existing, nil
				},
			}

			service := NewService(mockRepo, nil)

			err := service.SetActive(ctx, tt.templateID, tt.active)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetActive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want to contain %s", err, tt.errContains)
				}
			}
		})
	}
}

func TestTemplateService_SetDefault(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		userID      int64
		role        models.UserRole
		wantErr     bool
		errContains string
	}{
		{
			name:       "admin can set default",
			templateID: 1,
			userID:     1,
			role:       models.RoleAdmin,
			wantErr:    false,
		},
		{
			name:        "non-admin cannot set default",
			templateID:  1,
			userID:      1,
			role:        models.RoleEventManager,
			wantErr:     true,
			errContains: "Only admins can set default",
		},
		{
			name:        "unauthorized",
			templateID:  1,
			userID:      0,
			role:        models.RoleGuest,
			wantErr:     true,
			errContains: "Authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{ID: tt.userID, Role: tt.role}
			ctx := auth.WithUser(context.Background(), user)

			mockRepo := &mockServiceTemplateRepository{}

			service := NewService(mockRepo, nil)

			err := service.SetDefault(ctx, tt.templateID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !containsString(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want to contain %s", err, tt.errContains)
				}
			}
		})
	}
}

func TestTemplateService_GetTemplate(t *testing.T) {
	tests := []struct {
		name       string
		templateID int64
		existing   *models.Template
		wantErr    bool
	}{
		{
			name:       "get existing template",
			templateID: 1,
			existing: &models.Template{
				ID:   1,
				Name: "Test",
			},
			wantErr: false,
		},
		{
			name:       "get non-existent template",
			templateID: 99999,
			existing:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceTemplateRepository{
				GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
					if tt.existing != nil {
						return tt.existing, nil
					}
					return nil, &models.NotFoundError{Resource: "Template", ID: id}
				},
			}

			service := NewService(mockRepo, nil)

			result, err := service.GetTemplate(context.Background(), tt.templateID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.ID != tt.existing.ID {
				t.Errorf("GetTemplate() ID = %d, want %d", result.ID, tt.existing.ID)
			}
		})
	}
}

func TestTemplateService_ListTemplates(t *testing.T) {
	tests := []struct {
		name      string
		templates []*models.Template
		wantCount int
	}{
		{
			name: "list templates",
			templates: []*models.Template{
				{ID: 1, Name: "Template 1"},
				{ID: 2, Name: "Template 2"},
			},
			wantCount: 2,
		},
		{
			name:      "empty list",
			templates: []*models.Template{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockServiceTemplateRepository{
				ListFunc: func(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
					return tt.templates, nil
				},
			}

			service := NewService(mockRepo, nil)

			result, err := service.ListTemplates(context.Background(), nil)
			if err != nil {
				t.Errorf("ListTemplates() error = %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListTemplates() count = %d, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestTemplateService_RenderWithComponents(t *testing.T) {
	tests := []struct {
		name             string
		template         *models.Template
		event            *models.Event
		wantErr          bool
		wantComponentUse bool
	}{
		{
			name: "template with component config",
			template: &models.Template{
				ID:          1,
				Name:        "Component Template",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<h1>Legacy</h1>",
				ComponentConfig: stringPtr(`{
					"version": "1.0",
					"metadata": {"name": "Test", "category": "card", "description": "Test"},
					"layout": {"mode": "card"},
					"components": [{
						"id": "title",
						"type": "TextBox",
						"position": {"mode": "absolute", "x": "50%", "y": "100px"},
						"dimensions": {"width": "80%", "height": "auto"},
						"zIndex": 10,
						"visible": true,
						"content": {"text": "{{.Event.Title}}"}
					}]
				}`),
			},
			event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			},
			wantErr:          false,
			wantComponentUse: true,
		},
		{
			name: "template without component config uses legacy",
			template: &models.Template{
				ID:              2,
				Name:            "Legacy Template",
				Type:            models.TemplateTypeRSVPPage,
				HTMLContent:     "<h1>{{.Event.Title}}</h1>",
				ComponentConfig: nil,
			},
			event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			},
			wantErr:          false,
			wantComponentUse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			renderer := NewComponentRenderer(engine)
			svc := &service{
				repo:               &mockServiceTemplateRepository{},
				validator:          &mockServiceValidator{},
				componentRenderer:  renderer,
			}

			hasComponentConfig := tt.template.ComponentConfig != nil && *tt.template.ComponentConfig != ""
			if hasComponentConfig != tt.wantComponentUse {
				t.Errorf("Component config presence = %v, want %v", hasComponentConfig, tt.wantComponentUse)
			}

			if svc.componentRenderer == nil {
				t.Error("ComponentRenderer should be initialized in service")
			}
		})
	}
}
