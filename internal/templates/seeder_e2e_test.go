package templates

import (
	"context"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestSeeder_EndToEnd_ApplicationStartup(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	templateRepo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(templateRepo, 0)

	err = seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("SeedThemes failed: %v", err)
	}

	filters := &repositories.TemplateFilters{
		Type: ptrTemplateType(models.TemplateTypeRSVPPage),
	}
	themes, err := templateRepo.List(context.Background(), filters)
	if err != nil {
		t.Fatalf("Failed to list themes: %v", err)
	}

	if len(themes) != 7 {
		t.Fatalf("Expected 7 themes, got %d", len(themes))
	}

	defaultTheme, err := templateRepo.GetDefaultByType(context.Background(), models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("Failed to get default theme: %v", err)
	}

	if defaultTheme.Name != "Simple & Clean" {
		t.Errorf("Expected default theme to be 'Simple & Clean', got %q", defaultTheme.Name)
	}

	if defaultTheme.Category != models.CategoryPlainText {
		t.Errorf("Expected default theme category to be 'plain-text', got %q", defaultTheme.Category)
	}

	for _, theme := range themes {
		if theme.CreatedBy != 0 {
			t.Errorf("Theme %q has CreatedBy=%d, expected 0 for system themes", theme.Name, theme.CreatedBy)
		}

		if !theme.IsActive {
			t.Errorf("Theme %q is not active", theme.Name)
		}

		if theme.HTMLContent == "" {
			t.Errorf("Theme %q has empty HTML content", theme.Name)
		}

		if theme.CSSContent == nil || *theme.CSSContent == "" {
			t.Errorf("Theme %q has empty CSS content", theme.Name)
		}

		if theme.Category == models.CategoryCard {
			if theme.ThumbnailURL == nil || *theme.ThumbnailURL == "" {
				t.Errorf("Card theme %q missing thumbnail URL", theme.Name)
			}
			if theme.ImageURL == nil || *theme.ImageURL == "" {
				t.Errorf("Card theme %q missing image URL", theme.Name)
			}
		}
	}

	err = seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("Second SeedThemes failed: %v", err)
	}

	themesAfterSecondSeed, err := templateRepo.List(context.Background(), filters)
	if err != nil {
		t.Fatalf("Failed to list themes after second seed: %v", err)
	}

	if len(themesAfterSecondSeed) != 7 {
		t.Errorf("Expected 7 themes after second seed, got %d (not idempotent)", len(themesAfterSecondSeed))
	}

	for i, theme := range themesAfterSecondSeed {
		if theme.ID != themes[i].ID {
			t.Errorf("Theme %q ID changed from %d to %d after second seed", theme.Name, themes[i].ID, theme.ID)
		}
	}
}

func TestSeeder_EndToEnd_ThemeRetrieval(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	templateRepo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(templateRepo, 0)

	err = seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("SeedThemes failed: %v", err)
	}

	testCases := []struct {
		name     string
		category models.TemplateCategory
		expected int
	}{
		{"Plain-text themes", models.CategoryPlainText, 1},
		{"Wedding-elegance themes", models.CategoryWeddingElegance, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			themes, err := templateRepo.GetTemplatesByCategory(context.Background(), tc.category)
			if err != nil {
				t.Fatalf("Failed to get themes by category: %v", err)
			}

			if len(themes) != tc.expected {
				t.Errorf("Expected %d %s, got %d", tc.expected, tc.name, len(themes))
			}
		})
	}
}

func TestSeeder_EndToEnd_ThemeByName(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	templateRepo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(templateRepo, 0)

	err = seeder.SeedThemes(context.Background())
	if err != nil {
		t.Fatalf("SeedThemes failed: %v", err)
	}

	themeNames := []string{
		"Simple & Clean",
		"Wedding Elegance",
		"Birthday Celebration",
		"Corporate Professional",
		"Holiday Festive",
		"Garden Party",
		"Modern Minimalist",
	}

	for _, name := range themeNames {
		t.Run(name, func(t *testing.T) {
			theme, err := templateRepo.GetByNameAndType(context.Background(), name, models.TemplateTypeRSVPPage)
			if err != nil {
				t.Fatalf("Failed to get theme %q: %v", name, err)
			}

			if theme.Name != name {
				t.Errorf("Expected name %q, got %q", name, theme.Name)
			}

			if theme.Type != models.TemplateTypeRSVPPage {
				t.Errorf("Expected type %q, got %q", models.TemplateTypeRSVPPage, theme.Type)
			}

			if theme.CreatedBy != 0 {
				t.Errorf("Expected CreatedBy=0 for system theme, got %d", theme.CreatedBy)
			}
		})
	}
}
