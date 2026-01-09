package main

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

func TestTemplateSeeding_OnStartup(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	templateRepo := repositories.NewTemplateRepository(database)

	systemUser := &models.User{
		Email: "system@tinyrsvp.local",
		Name:  "System",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, systemUser); err != nil {
		t.Fatalf("Failed to create system user: %v", err)
	}

	if systemUser.ID == 0 {
		t.Fatal("System user ID should be non-zero after creation")
	}

	seeder := templates.NewSeeder(templateRepo, systemUser.ID)
	if err := seeder.SeedDefaults(ctx); err != nil {
		t.Fatalf("Failed to seed templates: %v", err)
	}

	templateTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range templateTypes {
		tmpl, err := templateRepo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Errorf("GetDefaultByType(%s) error = %v", typ, err)
			continue
		}
		if tmpl == nil {
			t.Errorf("No default template found for type %s", typ)
			continue
		}
		if !tmpl.IsDefault {
			t.Errorf("Template for type %s is not marked as default", typ)
		}
		if !tmpl.IsActive {
			t.Errorf("Template for type %s is not marked as active", typ)
		}
		if tmpl.CreatedBy != systemUser.ID {
			t.Errorf("Template for type %s has wrong creator: got %d, want %d", typ, tmpl.CreatedBy, systemUser.ID)
		}
	}

	if err := seeder.SeedDefaults(ctx); err != nil {
		t.Errorf("Second call to SeedDefaults should not error: %v", err)
	}

	for _, typ := range templateTypes {
		isDefault := true
		templates, err := templateRepo.List(ctx, &repositories.TemplateFilters{
			Type:      &typ,
			IsDefault: &isDefault,
		})
		if err != nil {
			t.Errorf("List(%s) error = %v", typ, err)
			continue
		}
		if len(templates) != 1 {
			t.Errorf("Expected exactly 1 default template for type %s, got %d", typ, len(templates))
		}
	}
}

func TestTemplateSeeding_IdempotentOnStartup(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	templateRepo := repositories.NewTemplateRepository(database)

	systemUser := &models.User{
		Email: "system@tinyrsvp.local",
		Name:  "System",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, systemUser); err != nil {
		t.Fatalf("Failed to create system user: %v", err)
	}

	seeder := templates.NewSeeder(templateRepo, systemUser.ID)

	for i := 0; i < 3; i++ {
		if err := seeder.SeedDefaults(ctx); err != nil {
			t.Fatalf("SeedDefaults call %d failed: %v", i+1, err)
		}
	}

	templateTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range templateTypes {
		isDefault := true
		templates, err := templateRepo.List(ctx, &repositories.TemplateFilters{
			Type:      &typ,
			IsDefault: &isDefault,
		})
		if err != nil {
			t.Errorf("List(%s) error = %v", typ, err)
			continue
		}
		if len(templates) != 1 {
			t.Errorf("After 3 seeding calls, expected exactly 1 default template for type %s, got %d", typ, len(templates))
		}
	}
}
