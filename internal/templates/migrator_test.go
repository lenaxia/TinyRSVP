package templates

import (
	"encoding/json"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestThemeMigrator_MigratePlainText(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigratePlainText()
	if err != nil {
		t.Fatalf("Failed to migrate plain-text theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	if config.Metadata.Name != "Plain Text" {
		t.Errorf("Expected name 'Plain Text', got %s", config.Metadata.Name)
	}

	if config.Metadata.Category != "simple" {
		t.Errorf("Expected category 'simple', got %s", config.Metadata.Category)
	}

	if len(config.Components) == 0 {
		t.Fatal("Expected at least one component")
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Plain Text Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_MigrateWeddingElegance(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateWeddingElegance()
	if err != nil {
		t.Fatalf("Failed to migrate wedding-elegance theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	if config.Metadata.Name != "Wedding Elegance" {
		t.Errorf("Expected name 'Wedding Elegance', got %s", config.Metadata.Name)
	}

	if config.Metadata.Category != "card" {
		t.Errorf("Expected category 'card', got %s", config.Metadata.Category)
	}

	if len(config.Components) == 0 {
		t.Fatal("Expected at least one component")
	}

	hasHeaderImage := false
	hasTitleText := false
	for _, comp := range config.Components {
		if comp.ID == "header-image" && comp.Type == models.ComponentTypeImage {
			hasHeaderImage = true
		}
		if comp.ID == "title-text" && comp.Type == models.ComponentTypeTextBox {
			hasTitleText = true
		}
	}

	if !hasHeaderImage {
		t.Error("Expected header-image component")
	}

	if !hasTitleText {
		t.Error("Expected title-text component")
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Wedding Elegance Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_MigrateBirthdayCelebration(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateBirthdayCelebration()
	if err != nil {
		t.Fatalf("Failed to migrate birthday-celebration theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Metadata.Name != "Birthday Celebration" {
		t.Errorf("Expected name 'Birthday Celebration', got %s", config.Metadata.Name)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Birthday Celebration Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_MigrateCorporateProfessional(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateCorporateProfessional()
	if err != nil {
		t.Fatalf("Failed to migrate corporate-professional theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Metadata.Name != "Corporate Professional" {
		t.Errorf("Expected name 'Corporate Professional', got %s", config.Metadata.Name)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Corporate Professional Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_MigrateHolidayFestive(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateHolidayFestive()
	if err != nil {
		t.Fatalf("Failed to migrate holiday-festive theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Metadata.Name != "Holiday Festive" {
		t.Errorf("Expected name 'Holiday Festive', got %s", config.Metadata.Name)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Holiday Festive Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_MigrateGardenParty(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateGardenParty()
	if err != nil {
		t.Fatalf("Failed to migrate garden-party theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Metadata.Name != "Garden Party" {
		t.Errorf("Expected name 'Garden Party', got %s", config.Metadata.Name)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Garden Party Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_MigrateModernMinimalist(t *testing.T) {
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateModernMinimalist()
	if err != nil {
		t.Fatalf("Failed to migrate modern-minimalist theme: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil configuration")
	}

	if config.Metadata.Name != "Modern Minimalist" {
		t.Errorf("Expected name 'Modern Minimalist', got %s", config.Metadata.Name)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config to JSON: %v", err)
	}

	t.Logf("Modern Minimalist Configuration:\n%s", string(jsonBytes))
}

func TestThemeMigrator_AllThemes(t *testing.T) {
	migrator := NewThemeMigrator()

	themes := []struct {
		name    string
		migrate func() (*models.ComponentConfiguration, error)
	}{
		{"Plain Text", migrator.MigratePlainText},
		{"Wedding Elegance", migrator.MigrateWeddingElegance},
		{"Birthday Celebration", migrator.MigrateBirthdayCelebration},
		{"Corporate Professional", migrator.MigrateCorporateProfessional},
		{"Holiday Festive", migrator.MigrateHolidayFestive},
		{"Garden Party", migrator.MigrateGardenParty},
		{"Modern Minimalist", migrator.MigrateModernMinimalist},
	}

	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			config, err := theme.migrate()
			if err != nil {
				t.Fatalf("Failed to migrate %s: %v", theme.name, err)
			}

			if config == nil {
				t.Fatalf("Expected non-nil configuration for %s", theme.name)
			}

			if config.Version == "" {
				t.Errorf("Expected version for %s", theme.name)
			}

			if config.Metadata.Name == "" {
				t.Errorf("Expected metadata name for %s", theme.name)
			}

			if len(config.Components) == 0 {
				t.Errorf("Expected at least one component for %s", theme.name)
			}

			for i, comp := range config.Components {
				if comp.ID == "" {
					t.Errorf("Component %d in %s has empty ID", i, theme.name)
				}
				if !comp.Type.IsValid() {
					t.Errorf("Component %d in %s has invalid type: %s", i, theme.name, comp.Type)
				}
			}
		})
	}
}
