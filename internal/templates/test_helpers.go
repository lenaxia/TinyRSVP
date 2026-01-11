package templates

import (
	"context"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockServiceTemplateRepository struct {
	CreateFunc                  func(ctx context.Context, template *models.Template) error
	GetByIDFunc                 func(ctx context.Context, id int64) (*models.Template, error)
	GetByEventAndTypeFunc       func(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error)
	GetDefaultByTypeFunc        func(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
	GetByNameAndTypeFunc        func(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error)
	ListFunc                    func(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error)
	UpdateFunc                  func(ctx context.Context, template *models.Template) error
	DeleteFunc                  func(ctx context.Context, id int64) error
	SetActiveFunc               func(ctx context.Context, id int64, active bool) error
	IsTemplateInUseFunc         func(ctx context.Context, id int64) (bool, error)
	SetDefaultFunc              func(ctx context.Context, id int64) error
	GetTemplatesByCategoryFunc  func(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error)
	ListThemesFunc              func(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error)
	GetComponentConfigFunc      func(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error)
	UpdateComponentConfigFunc   func(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error
	ValidateComponentConfigFunc func(ctx context.Context, config *models.ComponentConfiguration) error
}

func (m *mockServiceTemplateRepository) Create(ctx context.Context, template *models.Template) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, template)
	}
	template.ID = 1
	template.Version = 1
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	return nil
}

func (m *mockServiceTemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "Template", ID: id}
}

func (m *mockServiceTemplateRepository) GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	if m.GetByEventAndTypeFunc != nil {
		return m.GetByEventAndTypeFunc(ctx, eventID, templateType)
	}
	return nil, &models.NotFoundError{Resource: "Template"}
}

func (m *mockServiceTemplateRepository) GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	if m.GetDefaultByTypeFunc != nil {
		return m.GetDefaultByTypeFunc(ctx, templateType)
	}
	return nil, &models.NotFoundError{Resource: "Template"}
}

func (m *mockServiceTemplateRepository) GetByNameAndType(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error) {
	if m.GetByNameAndTypeFunc != nil {
		return m.GetByNameAndTypeFunc(ctx, name, templateType)
	}
	return nil, &models.NotFoundError{Resource: "Template"}
}

func (m *mockServiceTemplateRepository) List(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return []*models.Template{}, nil
}

func (m *mockServiceTemplateRepository) Update(ctx context.Context, template *models.Template) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, template)
	}
	template.Version++
	template.UpdatedAt = time.Now()
	return nil
}

func (m *mockServiceTemplateRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockServiceTemplateRepository) SetActive(ctx context.Context, id int64, active bool) error {
	if m.SetActiveFunc != nil {
		return m.SetActiveFunc(ctx, id, active)
	}
	return nil
}

func (m *mockServiceTemplateRepository) IsTemplateInUse(ctx context.Context, id int64) (bool, error) {
	if m.IsTemplateInUseFunc != nil {
		return m.IsTemplateInUseFunc(ctx, id)
	}
	return false, nil
}

func (m *mockServiceTemplateRepository) SetDefault(ctx context.Context, id int64) error {
	if m.SetDefaultFunc != nil {
		return m.SetDefaultFunc(ctx, id)
	}
	return nil
}

func (m *mockServiceTemplateRepository) GetTemplatesByCategory(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error) {
	if m.GetTemplatesByCategoryFunc != nil {
		return m.GetTemplatesByCategoryFunc(ctx, category)
	}
	return []*models.Template{}, nil
}

func (m *mockServiceTemplateRepository) ListThemes(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error) {
	if m.ListThemesFunc != nil {
		return m.ListThemesFunc(ctx, templateType, category)
	}
	return []*models.Template{}, nil
}

func (m *mockServiceTemplateRepository) GetComponentConfig(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error) {
	if m.GetComponentConfigFunc != nil {
		return m.GetComponentConfigFunc(ctx, templateID)
	}
	return nil, nil
}

func (m *mockServiceTemplateRepository) UpdateComponentConfig(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error {
	if m.UpdateComponentConfigFunc != nil {
		return m.UpdateComponentConfigFunc(ctx, templateID, config)
	}
	return nil
}

func (m *mockServiceTemplateRepository) ValidateComponentConfig(ctx context.Context, config *models.ComponentConfiguration) error {
	if m.ValidateComponentConfigFunc != nil {
		return m.ValidateComponentConfigFunc(ctx, config)
	}
	return nil
}

type mockServiceValidator struct {
	ValidateTemplateFunc func(template *models.Template) error
}

func (m *mockServiceValidator) ValidateTemplate(template *models.Template) error {
	if m.ValidateTemplateFunc != nil {
		return m.ValidateTemplateFunc(template)
	}
	return nil
}

func (m *mockServiceValidator) ValidateSyntax(content string, templateType models.TemplateType) error {
	return nil
}

func (m *mockServiceValidator) ValidateVariables(content string, allowedVars []string) error {
	return nil
}

func (m *mockServiceValidator) ValidateSize(content string, maxBytes int) error {
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func containsString(s, substr string) bool {
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
