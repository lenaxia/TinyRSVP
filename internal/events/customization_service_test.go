package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockAuthzChecker struct{}

func (m *mockAuthzChecker) CanCreateEvent(ctx context.Context, user *models.User) bool {
	return user.Role == models.RoleEventManager || user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	return user.ID == event.CreatedBy || user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	return user.ID == event.CreatedBy || user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	return user.ID == event.CreatedBy || user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool {
	return user.ID == event.CreatedBy || user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool {
	return user.ID == event.CreatedBy || user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanManageUsers(ctx context.Context, user *models.User) bool {
	return user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) CanConfigureSystem(ctx context.Context, user *models.User) bool {
	return user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) IsAdmin(user *models.User) bool {
	return user.Role == models.RoleAdmin
}

func (m *mockAuthzChecker) IsEventManager(user *models.User) bool {
	return user.Role == models.RoleEventManager || user.Role == models.RoleAdmin
}

func TestCustomizationService_GetEventCustomization(t *testing.T) {
	templateID := int64(1)
	componentConfig := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "title-text",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
			},
		},
	}
	configJSON, _ := json.Marshal(componentConfig)
	configStr := string(configJSON)

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#ff0000",
					},
				},
			},
		},
	}
	overridesJSON, _ := json.Marshal(overrides)
	overridesStr := string(overridesJSON)

	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:                 1,
				CreatedBy:          1,
				TemplateID:         &templateID,
				ComponentOverrides: &overridesStr,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:              1,
				ComponentConfig: &configStr,
			}, nil
		},
	}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 1, Role: models.RoleEventManager})

	result, err := service.GetEventCustomization(ctx, 1)
	if err != nil {
		t.Fatalf("GetEventCustomization failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.TemplateConfig == nil {
		t.Error("Expected template config")
	}

	if result.EventOverrides == nil {
		t.Error("Expected event overrides")
	}
}

func TestCustomizationService_GetEventCustomization_NoTemplate(t *testing.T) {
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 1, Role: models.RoleEventManager})

	_, err := service.GetEventCustomization(ctx, 1)
	if err == nil {
		t.Error("Expected error when event has no template")
	}
}

func TestCustomizationService_GetEventCustomization_PermissionDenied(t *testing.T) {
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 2, Role: models.RoleEventManager})

	_, err := service.GetEventCustomization(ctx, 1)
	if err == nil {
		t.Error("Expected permission denied error")
	}

	if _, ok := err.(*models.PermissionDeniedError); !ok {
		t.Errorf("Expected PermissionDeniedError, got %T", err)
	}
}

func TestCustomizationService_UpdateEventCustomization(t *testing.T) {
	updateCalled := 0
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
		UpdateComponentOverridesFunc: func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
			updateCalled++
			return nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 1, Role: models.RoleEventManager})

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#00ff00",
					},
				},
			},
		},
		Additions: []models.Component{},
		Removals:  []string{},
	}

	err := service.UpdateEventCustomization(ctx, 1, overrides)
	if err != nil {
		t.Fatalf("UpdateEventCustomization failed: %v", err)
	}

	if updateCalled != 1 {
		t.Errorf("Expected UpdateComponentOverrides to be called once, got %d", updateCalled)
	}
}

func TestCustomizationService_UpdateEventCustomization_PermissionDenied(t *testing.T) {
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 2, Role: models.RoleEventManager})

	overrides := &models.ComponentOverrides{
		Version:   "1.0",
		Overrides: []models.ComponentOverride{},
		Additions: []models.Component{},
		Removals:  []string{},
	}

	err := service.UpdateEventCustomization(ctx, 1, overrides)
	if err == nil {
		t.Error("Expected permission denied error")
	}

	if _, ok := err.(*models.PermissionDeniedError); !ok {
		t.Errorf("Expected PermissionDeniedError, got %T", err)
	}
}

func TestCustomizationService_UpdateEventCustomization_ValidationError(t *testing.T) {
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 1, Role: models.RoleEventManager})

	overrides := &models.ComponentOverrides{
		Version: "",
		Overrides: []models.ComponentOverride{
			{
				ID:      "",
				Updates: map[string]interface{}{},
			},
		},
	}

	err := service.UpdateEventCustomization(ctx, 1, overrides)
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestCustomizationService_PreviewEventCustomization(t *testing.T) {
	templateID := int64(1)
	componentConfig := &models.ComponentConfiguration{
		Version: "1.0",
		Components: []models.Component{
			{
				ID:      "title-text",
				Type:    models.ComponentTypeTextBox,
				ZIndex:  10,
				Visible: true,
			},
		},
	}
	configJSON, _ := json.Marshal(componentConfig)
	configStr := string(configJSON)

	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:         1,
				CreatedBy:  1,
				TemplateID: &templateID,
			}, nil
		},
	}

	mockTemplateRepo := &mockTemplateRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
			return &models.Template{
				ID:              1,
				ComponentConfig: &configStr,
			}, nil
		},
	}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 1, Role: models.RoleEventManager})

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#0000ff",
					},
				},
			},
		},
	}

	result, err := service.PreviewEventCustomization(ctx, 1, overrides)
	if err != nil {
		t.Fatalf("PreviewEventCustomization failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(result.Components))
	}
}

func TestCustomizationService_ResetEventCustomization(t *testing.T) {
	overrides := &models.ComponentOverrides{
		Version:   "1.0",
		Overrides: []models.ComponentOverride{},
	}
	overridesJSON, _ := json.Marshal(overrides)
	overridesStr := string(overridesJSON)

	deleteCalled := 0
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:                 1,
				CreatedBy:          1,
				ComponentOverrides: &overridesStr,
			}, nil
		},
		DeleteComponentOverridesFunc: func(ctx context.Context, eventID int64) error {
			deleteCalled++
			return nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 1, Role: models.RoleEventManager})

	err := service.ResetEventCustomization(ctx, 1)
	if err != nil {
		t.Fatalf("ResetEventCustomization failed: %v", err)
	}

	if deleteCalled != 1 {
		t.Errorf("Expected DeleteComponentOverrides to be called once, got %d", deleteCalled)
	}
}

func TestCustomizationService_ResetEventCustomization_PermissionDenied(t *testing.T) {
	mockRepo := &mockEventRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
	}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	ctx := createTestContext(t, &models.User{ID: 2, Role: models.RoleEventManager})

	err := service.ResetEventCustomization(ctx, 1)
	if err == nil {
		t.Error("Expected permission denied error")
	}

	if _, ok := err.(*models.PermissionDeniedError); !ok {
		t.Errorf("Expected PermissionDeniedError, got %T", err)
	}
}

func TestCustomizationService_ValidateEventCustomization(t *testing.T) {
	mockRepo := &mockEventRepository{}
	mockTemplateRepo := &mockTemplateRepository{}

	service := NewCustomizationService(mockRepo, mockTemplateRepo, &mockAuthzChecker{})

	tests := []struct {
		name      string
		overrides *models.ComponentOverrides
		wantErr   bool
	}{
		{
			name: "valid overrides",
			overrides: &models.ComponentOverrides{
				Version: "1.0",
				Overrides: []models.ComponentOverride{
					{
						ID: "title-text",
						Updates: map[string]interface{}{
							"content": map[string]interface{}{
								"color": "#ff0000",
							},
						},
					},
				},
				Additions: []models.Component{},
				Removals:  []string{},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			overrides: &models.ComponentOverrides{
				Version:   "",
				Overrides: []models.ComponentOverride{},
			},
			wantErr: true,
		},
		{
			name: "empty override ID",
			overrides: &models.ComponentOverrides{
				Version: "1.0",
				Overrides: []models.ComponentOverride{
					{
						ID:      "",
						Updates: map[string]interface{}{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "nil overrides",
			overrides: &models.ComponentOverrides{
				Version:   "1.0",
				Overrides: nil,
			},
			wantErr: false,
		},
		{
			name: "empty removal ID",
			overrides: &models.ComponentOverrides{
				Version:  "1.0",
				Removals: []string{""},
			},
			wantErr: true,
		},
		{
			name: "invalid addition component",
			overrides: &models.ComponentOverrides{
				Version: "1.0",
				Additions: []models.Component{
					{
						ID:      "",
						Type:    models.ComponentTypeTextBox,
						Visible: true,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateEventCustomization(tt.overrides)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEventCustomization() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockTemplateRepository struct {
	GetByIDFunc func(ctx context.Context, id int64) (*models.Template, error)
}

func (m *mockTemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "Template", ID: id}
}

func createTestContext(t *testing.T, user *models.User) context.Context {
	t.Helper()
	return auth.WithUser(context.Background(), user)
}
