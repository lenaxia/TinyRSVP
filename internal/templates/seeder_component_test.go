package templates

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestSeeder_SeedThemesWithComponentConfigs(t *testing.T) {
	mockRepo := &mockServiceTemplateRepository{
		GetByNameAndTypeFunc: func(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error) {
			return nil, &models.NotFoundError{Resource: "Template"}
		},
		CreateFunc: func(ctx context.Context, template *models.Template) error {
			template.ID = 1
			return nil
		},
	}

	seeder := NewSeeder(mockRepo, 1)
	ctx := context.Background()

	if err := seeder.SeedThemes(ctx); err != nil {
		t.Fatalf("SeedThemes failed: %v", err)
	}
}

func TestSeeder_LoadComponentConfig(t *testing.T) {
	seeder := NewSeeder(nil, 1)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"Plain Text", "plain-text.json", false},
		{"Wedding Elegance", "wedding-elegance.json", false},
		{"Birthday Celebration", "birthday-celebration.json", false},
		{"Corporate Professional", "corporate-professional.json", false},
		{"Holiday Festive", "holiday-festive.json", false},
		{"Garden Party", "garden-party.json", false},
		{"Modern Minimalist", "modern-minimalist.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configJSON := seeder.loadComponentConfig(tt.filename)
			if configJSON == "" && !tt.wantErr {
				t.Errorf("Expected non-empty config for %s", tt.name)
				return
			}

			if configJSON != "" {
				var config models.ComponentConfiguration
				if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
					t.Errorf("Failed to parse component config JSON for %s: %v", tt.name, err)
					return
				}

				if config.Version != "1.0" {
					t.Errorf("Expected version 1.0, got %s", config.Version)
				}

				if len(config.Components) == 0 {
					t.Errorf("Expected at least one component for %s", tt.name)
				}

				t.Logf("%s has %d components", tt.name, len(config.Components))
			}
		})
	}
}

func TestSeeder_ThemesHaveComponentConfigs(t *testing.T) {
	seeder := NewSeeder(nil, 1)
	themes := seeder.getDefaultThemes()

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			if theme.ComponentConfig == nil {
				t.Errorf("Theme %s is missing ComponentConfig", theme.Name)
				return
			}

			var config models.ComponentConfiguration
			if err := json.Unmarshal([]byte(*theme.ComponentConfig), &config); err != nil {
				t.Errorf("Failed to parse ComponentConfig for %s: %v", theme.Name, err)
				return
			}

			if config.Version == "" {
				t.Errorf("ComponentConfig for %s has empty version", theme.Name)
			}

			if config.Metadata.Name == "" {
				t.Errorf("ComponentConfig for %s has empty metadata name", theme.Name)
			}

			if len(config.Components) == 0 {
				t.Errorf("ComponentConfig for %s has no components", theme.Name)
			}

			t.Logf("%s: %d components, category=%s", theme.Name, len(config.Components), config.Metadata.Category)
		})
	}
}
