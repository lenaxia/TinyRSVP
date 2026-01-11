package templates

import (
	"context"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestSeeder_SeedThemes_FreshDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	err := seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("SeedThemes failed: %v", err)
	}

	filters := &repositories.TemplateFilters{
		Type: ptrTemplateType(models.TemplateTypeRSVPPage),
	}
	templates, err := repo.List(context.Background(), filters)
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	if len(templates) != 7 {
		t.Errorf("Expected 7 themes, got %d", len(templates))
	}

	expectedThemes := []string{
		"Simple & Clean",
		"Wedding Elegance",
		"Birthday Celebration",
		"Corporate Professional",
		"Holiday Festive",
		"Garden Party",
		"Modern Minimalist",
	}

	foundThemes := make(map[string]bool)
	for _, tmpl := range templates {
		foundThemes[tmpl.Name] = true
	}

	for _, name := range expectedThemes {
		if !foundThemes[name] {
			t.Errorf("Theme %q not found in seeded templates", name)
		}
	}
}

func TestSeeder_SeedThemes_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	err := seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("First SeedThemes failed: %v", err)
	}

	filters := &repositories.TemplateFilters{
		Type: ptrTemplateType(models.TemplateTypeRSVPPage),
	}
	firstTemplates, err := repo.List(context.Background(), filters)
	if err != nil {
		t.Fatalf("Failed to list templates after first seed: %v", err)
	}
	firstCount := len(firstTemplates)

	err = seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("Second SeedThemes failed: %v", err)
	}

	secondTemplates, err := repo.List(context.Background(), filters)
	if err != nil {
		t.Fatalf("Failed to list templates after second seed: %v", err)
	}
	secondCount := len(secondTemplates)

	if firstCount != secondCount {
		t.Errorf("Seeding not idempotent: first=%d, second=%d", firstCount, secondCount)
	}

	if secondCount != 7 {
		t.Errorf("Expected 7 themes after second seed, got %d", secondCount)
	}
}

func TestSeeder_SeedThemes_UpdatesExisting(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)

	existingTheme := &models.Template{
		Name:        "Simple & Clean",
		Type:        models.TemplateTypeRSVPPage,
		Description: "Old description",
		HTMLContent: "<div>Old content</div>",
		Category:    models.CategoryPlain,
		IsDefault:   true,
		IsActive:    true,
		Version:     1,
		CreatedBy:   0,
		Tags:        []string{"old"},
		SortOrder:   0,
	}

	err := repo.Create(context.Background(), existingTheme)
	if err != nil {
		t.Fatalf("Failed to create existing theme: %v", err)
	}

	originalID := existingTheme.ID

	seeder := NewSeeder(repo, 0)
	err = seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("SeedThemes failed: %v", err)
	}

	updated, err := repo.GetByID(context.Background(), originalID)
	if err != nil {
		t.Fatalf("Failed to get updated theme: %v", err)
	}

	if updated.Description == "Old description" {
		t.Error("Theme description was not updated")
	}

	if updated.HTMLContent == "<div>Old content</div>" {
		t.Error("Theme HTML content was not updated")
	}

	if updated.ID != originalID {
		t.Errorf("Theme ID changed from %d to %d", originalID, updated.ID)
	}
}

func TestSeeder_GetDefaultThemes_ReturnsSevenThemes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	themes := seeder.getDefaultThemes()

	if len(themes) != 7 {
		t.Errorf("Expected 7 default themes, got %d", len(themes))
	}

	plainCount := 0
	cardCount := 0
	defaultCount := 0

	for _, theme := range themes {
		if theme.Category == models.CategoryPlain {
			plainCount++
		}
		if theme.Category == models.CategoryCard {
			cardCount++
		}
		if theme.IsDefault {
			defaultCount++
		}

		if theme.Name == "" {
			t.Error("Theme has empty name")
		}
		if theme.Type != models.TemplateTypeRSVPPage {
			t.Errorf("Theme %q has wrong type: %s", theme.Name, theme.Type)
		}
		if theme.HTMLContent == "" {
			t.Errorf("Theme %q has empty HTML content", theme.Name)
		}
		if theme.Category == "" {
			t.Errorf("Theme %q has empty category", theme.Name)
		}
	}

	if plainCount != 1 {
		t.Errorf("Expected 1 plain theme, got %d", plainCount)
	}

	if cardCount != 6 {
		t.Errorf("Expected 6 card themes, got %d", cardCount)
	}

	if defaultCount != 1 {
		t.Errorf("Expected 1 default theme, got %d", defaultCount)
	}
}

func TestSeeder_GetDefaultThemes_CorrectSortOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	themes := seeder.getDefaultThemes()

	for i, theme := range themes {
		if theme.SortOrder != i {
			t.Errorf("Theme %q has sort order %d, expected %d", theme.Name, theme.SortOrder, i)
		}
	}
}

func TestSeeder_GetDefaultThemes_HasRequiredMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	themes := seeder.getDefaultThemes()

	for _, theme := range themes {
		if theme.Description == "" {
			t.Errorf("Theme %q missing description", theme.Name)
		}

		if len(theme.Tags) == 0 {
			t.Errorf("Theme %q has no tags", theme.Name)
		}

		if theme.Category == models.CategoryCard {
			if theme.ThumbnailURL == nil || *theme.ThumbnailURL == "" {
				t.Errorf("Card theme %q missing thumbnail URL", theme.Name)
			}
			if theme.ImageURL == nil || *theme.ImageURL == "" {
				t.Errorf("Card theme %q missing image URL", theme.Name)
			}
		}

		if theme.Category == models.CategoryPlain {
			if theme.ImageURL != nil && *theme.ImageURL != "" {
				t.Errorf("Plain theme %q should not have image URL", theme.Name)
			}
		}
	}
}

func TestSeeder_GetDefaultThemes_SystemThemes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	themes := seeder.getDefaultThemes()

	for _, theme := range themes {
		if theme.CreatedBy != 0 {
			t.Errorf("Theme %q has CreatedBy=%d, expected 0 for system themes", theme.Name, theme.CreatedBy)
		}

		if !theme.IsActive {
			t.Errorf("Theme %q is not active", theme.Name)
		}
	}
}

func TestSeeder_GetDefaultThemes_UniqueNames(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	themes := seeder.getDefaultThemes()
	names := make(map[string]bool)

	for _, theme := range themes {
		if names[theme.Name] {
			t.Errorf("Duplicate theme name: %q", theme.Name)
		}
		names[theme.Name] = true
	}
}

func TestSeeder_SeedThemes_PartialFailure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewTemplateRepository(db)
	seeder := NewSeeder(repo, 0)

	err := seeder.SeedThemes(context.Background())
	if err != nil {
		t.Errorf("SeedThemes should not return error even if some themes fail, got: %v", err)
	}
}

func ptrTemplateType(t models.TemplateType) *models.TemplateType {
	return &t
}
