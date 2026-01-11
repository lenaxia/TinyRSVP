package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestTemplateRepository_GetByNameAndType_Found(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	ctx := context.Background()

	template := &models.Template{
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		Description: "Test description",
		HTMLContent: "<div>Test</div>",
		Category:    models.CategoryPlain,
		IsDefault:   false,
		IsActive:    true,
		Version:     1,
		CreatedBy:   user.ID,
		Tags:        []string{"test"},
		SortOrder:   0,
	}

	err := repo.Create(ctx, template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	found, err := repo.GetByNameAndType(ctx, "Test Theme", models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetByNameAndType failed: %v", err)
	}

	if found.Name != template.Name {
		t.Errorf("Expected name %q, got %q", template.Name, found.Name)
	}

	if found.Type != template.Type {
		t.Errorf("Expected type %q, got %q", template.Type, found.Type)
	}

	if found.ID == 0 {
		t.Error("Expected non-zero ID")
	}
}

func TestTemplateRepository_GetByNameAndType_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	ctx := context.Background()

	_, err := repo.GetByNameAndType(ctx, "Nonexistent Theme", models.TemplateTypeRSVPPage)
	
	if err == nil {
		t.Fatal("Expected error for nonexistent template")
	}

	var notFoundErr *models.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTemplateRepository_GetByNameAndType_DifferentTypes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	ctx := context.Background()

	template1 := &models.Template{
		Name:        "Same Name",
		Type:        models.TemplateTypeRSVPPage,
		Description: "RSVP Page",
		HTMLContent: "<div>RSVP</div>",
		Category:    models.CategoryPlain,
		IsDefault:   false,
		IsActive:    true,
		Version:     1,
		CreatedBy:   user.ID,
		Tags:        []string{},
		SortOrder:   0,
	}

	template2 := &models.Template{
		Name:        "Same Name",
		Type:        models.TemplateTypeInviteEmail,
		Description: "Invite Email",
		HTMLContent: "<div>Email</div>",
		TextContent: stringPtr("Email text"),
		Category:    models.CategoryPlain,
		IsDefault:   false,
		IsActive:    true,
		Version:     1,
		CreatedBy:   user.ID,
		Tags:        []string{},
		SortOrder:   0,
	}

	err := repo.Create(ctx, template1)
	if err != nil {
		t.Fatalf("Failed to create template1: %v", err)
	}

	err = repo.Create(ctx, template2)
	if err != nil {
		t.Fatalf("Failed to create template2: %v", err)
	}

	found, err := repo.GetByNameAndType(ctx, "Same Name", models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetByNameAndType failed: %v", err)
	}

	if found.Type != models.TemplateTypeRSVPPage {
		t.Errorf("Expected type %q, got %q", models.TemplateTypeRSVPPage, found.Type)
	}

	if found.Description != "RSVP Page" {
		t.Errorf("Got wrong template, description: %q", found.Description)
	}
}

func TestTemplateRepository_GetByNameAndType_CaseSensitive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	ctx := context.Background()

	template := &models.Template{
		Name:        "Test Theme",
		Type:        models.TemplateTypeRSVPPage,
		Description: "Test",
		HTMLContent: "<div>Test</div>",
		Category:    models.CategoryPlain,
		IsDefault:   false,
		IsActive:    true,
		Version:     1,
		CreatedBy:   user.ID,
		Tags:        []string{},
		SortOrder:   0,
	}

	err := repo.Create(ctx, template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	_, err = repo.GetByNameAndType(ctx, "test theme", models.TemplateTypeRSVPPage)
	if err == nil {
		t.Error("Expected error for case-different name, but got none")
	}
}
