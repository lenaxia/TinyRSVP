package templates

import (
	"context"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockTemplateRepository struct {
	templates map[models.TemplateType]*models.Template
	createErr error
	getErr    error
}

func newMockTemplateRepository() *mockTemplateRepository {
	return &mockTemplateRepository{
		templates: make(map[models.TemplateType]*models.Template),
	}
}

func (m *mockTemplateRepository) Create(ctx context.Context, template *models.Template) error {
	if m.createErr != nil {
		return m.createErr
	}
	template.ID = int64(len(m.templates) + 1)
	m.templates[template.Type] = template
	return nil
}

func (m *mockTemplateRepository) GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	tmpl, exists := m.templates[templateType]
	if !exists {
		return nil, &models.NotFoundError{Resource: "Template", ID: 0}
	}
	return tmpl, nil
}

func (m *mockTemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) List(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	return nil, nil
}

func (m *mockTemplateRepository) Update(ctx context.Context, template *models.Template) error {
	return nil
}

func (m *mockTemplateRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockTemplateRepository) SetActive(ctx context.Context, id int64, active bool) error {
	return nil
}

func TestNewSeeder(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	if seeder == nil {
		t.Fatal("NewSeeder() returned nil")
	}
}

func TestSeeder_SeedDefaults_CreatesAllTemplates(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	expectedTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range expectedTypes {
		tmpl, exists := repo.templates[typ]
		if !exists {
			t.Errorf("Template type %s was not created", typ)
			continue
		}

		if tmpl.Type != typ {
			t.Errorf("Template type mismatch: got %s, want %s", tmpl.Type, typ)
		}

		if !tmpl.IsDefault {
			t.Errorf("Template %s is not marked as default", typ)
		}

		if !tmpl.IsActive {
			t.Errorf("Template %s is not marked as active", typ)
		}

		if tmpl.Version != 1 {
			t.Errorf("Template %s version = %d, want 1", typ, tmpl.Version)
		}

		if tmpl.CreatedBy != 1 {
			t.Errorf("Template %s created_by = %d, want 1", typ, tmpl.CreatedBy)
		}

		if tmpl.HTMLContent == "" {
			t.Errorf("Template %s has empty HTML content", typ)
		}
	}
}

func TestSeeder_SeedDefaults_InviteEmailHasTextContent(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	tmpl := repo.templates[models.TemplateTypeInviteEmail]
	if tmpl.TextContent == nil {
		t.Error("Invite email template missing text content")
	} else if *tmpl.TextContent == "" {
		t.Error("Invite email template has empty text content")
	}
}

func TestSeeder_SeedDefaults_PageTemplatesNoTextContent(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	pageTypes := []models.TemplateType{
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range pageTypes {
		tmpl := repo.templates[typ]
		if tmpl.TextContent != nil {
			t.Errorf("Page template %s should not have text content", typ)
		}
	}
}

func TestSeeder_SeedDefaults_SkipsExistingTemplates(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()

	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("First SeedDefaults() error = %v", err)
	}

	firstCount := len(repo.templates)

	err = seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("Second SeedDefaults() error = %v", err)
	}

	secondCount := len(repo.templates)

	if firstCount != secondCount {
		t.Errorf("Second SeedDefaults() created new templates: first=%d, second=%d", firstCount, secondCount)
	}

	if firstCount != 3 {
		t.Errorf("Expected 3 templates, got %d", firstCount)
	}
}

func TestSeeder_SeedDefaults_TemplateNames(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	expectedNames := map[models.TemplateType]string{
		models.TemplateTypeInviteEmail:      "Default Invite Email",
		models.TemplateTypeRSVPPage:         "Default RSVP Page",
		models.TemplateTypeConfirmationPage: "Default Confirmation Page",
	}

	for typ, expectedName := range expectedNames {
		tmpl := repo.templates[typ]
		if tmpl.Name != expectedName {
			t.Errorf("Template %s name = %q, want %q", typ, tmpl.Name, expectedName)
		}
	}
}

func TestSeeder_SeedDefaults_ContextCancellation(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := seeder.SeedDefaults(ctx)
	if err == nil {
		t.Error("SeedDefaults() with cancelled context should return error")
	}
}

func TestSeeder_SeedDefaults_RepositoryError(t *testing.T) {
	repo := newMockTemplateRepository()
	repo.createErr = &models.ValidationError{Field: "test", Message: "test error"}
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err == nil {
		t.Error("SeedDefaults() should return error when repository fails")
	}
}

func TestSeeder_SeedDefaults_ValidatesTemplateContent(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	for typ, tmpl := range repo.templates {
		if len(tmpl.Name) < 3 {
			t.Errorf("Template %s name too short: %q", typ, tmpl.Name)
		}

		if len(tmpl.HTMLContent) < 100 {
			t.Errorf("Template %s HTML content suspiciously short: %d bytes", typ, len(tmpl.HTMLContent))
		}

		if typ == models.TemplateTypeInviteEmail && tmpl.TextContent != nil {
			if len(*tmpl.TextContent) < 50 {
				t.Errorf("Template %s text content suspiciously short: %d bytes", typ, len(*tmpl.TextContent))
			}
		}
	}
}

func TestSeeder_SeedDefaults_ZeroCreatedBy(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 0)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err == nil {
		t.Error("SeedDefaults() with createdBy=0 should return validation error")
	}
}

func TestSeeder_SeedDefaults_TemplateContentHasVariables(t *testing.T) {
	repo := newMockTemplateRepository()
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	tests := []struct {
		templateType models.TemplateType
		mustContain  []string
	}{
		{
			templateType: models.TemplateTypeInviteEmail,
			mustContain:  []string{"{{.Event.Title}}", "{{.RSVPURL}}", "{{formatDateTime"},
		},
		{
			templateType: models.TemplateTypeRSVPPage,
			mustContain:  []string{"{{.Event.Title}}", "{{.Token}}", "{{range .Questions}}"},
		},
		{
			templateType: models.TemplateTypeConfirmationPage,
			mustContain:  []string{"{{.Event.Title}}", ".RSVP.Response", "{{.Token}}"},
		},
	}

	for _, tt := range tests {
		tmpl := repo.templates[tt.templateType]
		for _, mustContain := range tt.mustContain {
			if !contains(tmpl.HTMLContent, mustContain) {
				t.Errorf("Template %s HTML content missing variable: %s", tt.templateType, mustContain)
			}
		}
	}
}
