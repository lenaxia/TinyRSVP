package templates

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupSeederTestDB(t *testing.T) (db.Database, func()) {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		database.Close()
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (1, 'system@test.com', 'System', 'admin', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		database.Close()
		t.Fatalf("Failed to create system user: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

func TestSeeder_Integration_SeedDefaults(t *testing.T) {
	database, cleanup := setupSeederTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	templateTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range templateTypes {
		tmpl, err := repo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Errorf("GetDefaultByType(%s) error = %v", typ, err)
			continue
		}

		if tmpl == nil {
			t.Errorf("No default template found for type %s", typ)
			continue
		}

		if !tmpl.IsDefault {
			t.Errorf("Template %s is not marked as default", typ)
		}

		if !tmpl.IsActive {
			t.Errorf("Template %s is not marked as active", typ)
		}

		if tmpl.HTMLContent == "" {
			t.Errorf("Template %s has empty HTML content", typ)
		}

		if typ == models.TemplateTypeInviteEmail {
			if tmpl.TextContent == nil || *tmpl.TextContent == "" {
				t.Errorf("Invite email template missing or empty text content")
			}
		}
	}
}

func TestSeeder_Integration_IdempotentSeeding(t *testing.T) {
	database, cleanup := setupSeederTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)

	ctx := context.Background()

	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("First SeedDefaults() error = %v", err)
	}

	firstTemplates := make(map[models.TemplateType]*models.Template)
	templateTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range templateTypes {
		tmpl, err := repo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Fatalf("GetDefaultByType(%s) error = %v", typ, err)
		}
		firstTemplates[typ] = tmpl
	}

	err = seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("Second SeedDefaults() error = %v", err)
	}

	for _, typ := range templateTypes {
		tmpl, err := repo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Fatalf("GetDefaultByType(%s) after second seed error = %v", typ, err)
		}

		if tmpl.ID != firstTemplates[typ].ID {
			t.Errorf("Template %s ID changed after second seed: %d -> %d", typ, firstTemplates[typ].ID, tmpl.ID)
		}

		if tmpl.Version != firstTemplates[typ].Version {
			t.Errorf("Template %s version changed after second seed: %d -> %d", typ, firstTemplates[typ].Version, tmpl.Version)
		}
	}
}

func TestSeeder_Integration_TemplatesParseable(t *testing.T) {
	database, cleanup := setupSeederTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	templateTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range templateTypes {
		tmpl, err := repo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Fatalf("GetDefaultByType(%s) error = %v", typ, err)
		}

		_, err = engine.Parse(tmpl.HTMLContent)
		if err != nil {
			t.Errorf("Failed to parse HTML template for %s: %v", typ, err)
		}

		if tmpl.TextContent != nil {
			_, err = engine.Parse(*tmpl.TextContent)
			if err != nil {
				t.Errorf("Failed to parse text template for %s: %v", typ, err)
			}
		}
	}
}

func TestSeeder_Integration_TemplatesRenderable(t *testing.T) {
	database, cleanup := setupSeederTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	testData := map[models.TemplateType]interface{}{
		models.TemplateTypeInviteEmail: struct {
			Event struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}
			Invite struct {
				Name  string
				Email string
			}
			RSVPURL     string
			MaxPlusOnes int
		}{
			Event: struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   time.Now(),
				Timezone:    "America/Los_Angeles",
				Location:    "Test Location",
			},
			Invite: struct {
				Name  string
				Email string
			}{
				Name:  "Test User",
				Email: "test@example.com",
			},
			RSVPURL:     "https://example.com/rsvp/token",
			MaxPlusOnes: 2,
		},
		models.TemplateTypeRSVPPage: struct {
			Event struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}
			Token       string
			MaxPlusOnes int
			Questions   []struct {
				ID           int64
				QuestionText string
				QuestionType string
				Required     bool
				HelpText     string
			}
			RSVP struct {
				Response string
				PlusOnes int
			}
		}{
			Event: struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   time.Now(),
				Timezone:    "America/Los_Angeles",
				Location:    "Test Location",
			},
			Token:       "test-token",
			MaxPlusOnes: 2,
			Questions:   []struct {
				ID           int64
				QuestionText string
				QuestionType string
				Required     bool
				HelpText     string
			}{},
			RSVP: struct {
				Response string
				PlusOnes int
			}{
				Response: "yes",
				PlusOnes: 1,
			},
		},
		models.TemplateTypeConfirmationPage: struct {
			Event struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}
			Token string
			RSVP  struct {
				Response string
				PlusOnes int
				Notes    string
			}
			Answers []struct {
				QuestionText  string
				AnswerDisplay string
			}
		}{
			Event: struct {
				Title        string
				Description  string
				StartTime    time.Time
				EndTime      *time.Time
				Timezone     string
				Location     string
				RSVPDeadline *time.Time
			}{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   time.Now(),
				Timezone:    "America/Los_Angeles",
				Location:    "Test Location",
			},
			Token: "test-token",
			RSVP: struct {
				Response string
				PlusOnes int
				Notes    string
			}{
				Response: "yes",
				PlusOnes: 1,
			},
			Answers: []struct {
				QuestionText  string
				AnswerDisplay string
			}{},
		},
	}

	for typ, data := range testData {
		tmpl, err := repo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Fatalf("GetDefaultByType(%s) error = %v", typ, err)
		}

		parsedHTML, err := engine.Parse(tmpl.HTMLContent)
		if err != nil {
			t.Fatalf("Failed to parse HTML template for %s: %v", typ, err)
		}

		result, err := engine.ExecuteToString(parsedHTML, data)
		if err != nil {
			t.Errorf("Failed to render HTML template for %s: %v", typ, err)
		}

		if len(result) == 0 {
			t.Errorf("Rendered HTML template for %s is empty", typ)
		}

		if !contains(result, "Test Event") {
			t.Errorf("Rendered HTML template for %s missing event title", typ)
		}

		if tmpl.TextContent != nil {
			parsedText, err := engine.Parse(*tmpl.TextContent)
			if err != nil {
				t.Fatalf("Failed to parse text template for %s: %v", typ, err)
			}

			textResult, err := engine.ExecuteToString(parsedText, data)
			if err != nil {
				t.Errorf("Failed to render text template for %s: %v", typ, err)
			}

			if len(textResult) == 0 {
				t.Errorf("Rendered text template for %s is empty", typ)
			}

			if !contains(textResult, "Test Event") {
				t.Errorf("Rendered text template for %s missing event title", typ)
			}
		}
	}
}

func TestSeeder_Integration_TemplatesValidated(t *testing.T) {
	database, cleanup := setupSeederTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()
	validator := NewValidator(engine)

	ctx := context.Background()
	err := seeder.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	templateTypes := []models.TemplateType{
		models.TemplateTypeInviteEmail,
		models.TemplateTypeRSVPPage,
		models.TemplateTypeConfirmationPage,
	}

	for _, typ := range templateTypes {
		tmpl, err := repo.GetDefaultByType(ctx, typ)
		if err != nil {
			t.Fatalf("GetDefaultByType(%s) error = %v", typ, err)
		}

		err = validator.ValidateTemplate(tmpl)
		if err != nil {
			t.Errorf("Default template %s failed validation: %v", typ, err)
		}
	}
}
