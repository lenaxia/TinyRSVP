package templates

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestMigratedThemes_RenderCorrectly(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	migrator := NewThemeMigrator()

	event := &models.Event{
		ID:          1,
		Title:       "Test Event",
		Description: stringPtr("This is a test event"),
		Location:    stringPtr("Test Location"),
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
	}

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

			configJSON := mustMarshalJSON(t, config)
			template := &models.Template{
				ID:              1,
				Name:            theme.name,
				Type:            models.TemplateTypeRSVPPage,
				ComponentConfig: &configJSON,
			}

			var buf bytes.Buffer
			if err := renderer.Render(&buf, event, template); err != nil {
				t.Fatalf("Failed to render %s: %v", theme.name, err)
			}

			html := buf.String()
			if html == "" {
				t.Fatal("Rendered HTML is empty")
			}

			if !strings.Contains(html, "Test Event") {
				t.Error("Rendered HTML does not contain event title")
			}

			if !strings.Contains(html, "Test Location") {
				t.Error("Rendered HTML does not contain event location")
			}

			if !strings.Contains(html, "component-canvas") {
				t.Error("Rendered HTML does not contain component-canvas")
			}

			t.Logf("%s rendered successfully (%d bytes)", theme.name, len(html))
		})
	}
}

func TestMigratedThemes_ComponentStructure(t *testing.T) {
	migrator := NewThemeMigrator()

	themes := []struct {
		name              string
		migrate           func() (*models.ComponentConfiguration, error)
		expectedMinComps  int
		hasHeaderImage    bool
		hasTitleText      bool
		hasDateText       bool
		hasLocationText   bool
		hasBackground     bool
	}{
		{
			name:             "Plain Text",
			migrate:          migrator.MigratePlainText,
			expectedMinComps: 4,
			hasHeaderImage:   false,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
		{
			name:             "Wedding Elegance",
			migrate:          migrator.MigrateWeddingElegance,
			expectedMinComps: 5,
			hasHeaderImage:   true,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
		{
			name:             "Birthday Celebration",
			migrate:          migrator.MigrateBirthdayCelebration,
			expectedMinComps: 5,
			hasHeaderImage:   true,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
		{
			name:             "Corporate Professional",
			migrate:          migrator.MigrateCorporateProfessional,
			expectedMinComps: 5,
			hasHeaderImage:   true,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
		{
			name:             "Holiday Festive",
			migrate:          migrator.MigrateHolidayFestive,
			expectedMinComps: 5,
			hasHeaderImage:   true,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
		{
			name:             "Garden Party",
			migrate:          migrator.MigrateGardenParty,
			expectedMinComps: 5,
			hasHeaderImage:   true,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
		{
			name:             "Modern Minimalist",
			migrate:          migrator.MigrateModernMinimalist,
			expectedMinComps: 5,
			hasHeaderImage:   true,
			hasTitleText:     true,
			hasDateText:      true,
			hasLocationText:  true,
			hasBackground:    true,
		},
	}

	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			config, err := theme.migrate()
			if err != nil {
				t.Fatalf("Failed to migrate %s: %v", theme.name, err)
			}

			if len(config.Components) < theme.expectedMinComps {
				t.Errorf("Expected at least %d components, got %d", theme.expectedMinComps, len(config.Components))
			}

			componentMap := make(map[string]models.Component)
			for _, comp := range config.Components {
				componentMap[comp.ID] = comp
			}

			if theme.hasBackground {
				if comp, exists := componentMap["page-background"]; !exists {
					t.Error("Missing page-background component")
				} else if comp.Type != models.ComponentTypeBackground {
					t.Errorf("page-background has wrong type: %s", comp.Type)
				}
			}

			if theme.hasHeaderImage {
				if comp, exists := componentMap["header-image"]; !exists {
					t.Error("Missing header-image component")
				} else if comp.Type != models.ComponentTypeImage {
					t.Errorf("header-image has wrong type: %s", comp.Type)
				}
			}

			if theme.hasTitleText {
				if comp, exists := componentMap["title-text"]; !exists {
					t.Error("Missing title-text component")
				} else if comp.Type != models.ComponentTypeTextBox {
					t.Errorf("title-text has wrong type: %s", comp.Type)
				}
			}

			if theme.hasDateText {
				if comp, exists := componentMap["date-text"]; !exists {
					t.Error("Missing date-text component")
				} else if comp.Type != models.ComponentTypeTextBox {
					t.Errorf("date-text has wrong type: %s", comp.Type)
				}
			}

			if theme.hasLocationText {
				if comp, exists := componentMap["location-text"]; !exists {
					t.Error("Missing location-text component")
				} else if comp.Type != models.ComponentTypeTextBox {
					t.Errorf("location-text has wrong type: %s", comp.Type)
				}
			}
		})
	}
}

func TestMigratedThemes_ResponsiveSupport(t *testing.T) {
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

			hasResponsive := false
			for _, comp := range config.Components {
				if comp.Responsive != nil {
					hasResponsive = true
					if comp.Responsive.Mobile != nil {
						t.Logf("%s component %s has mobile responsive config", theme.name, comp.ID)
					}
				}
			}

			if !hasResponsive {
				t.Logf("Warning: %s has no responsive configurations", theme.name)
			}
		})
	}
}

func TestMigratedThemes_TemplateVariables(t *testing.T) {
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

	requiredVars := []string{
		"{{.Event.Title}}",
		"{{formatDateTime .Event.StartTime}}",
		"{{.Event.Location}}",
	}

	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			config, err := theme.migrate()
			if err != nil {
				t.Fatalf("Failed to migrate %s: %v", theme.name, err)
			}

			allContent := ""
			for _, comp := range config.Components {
				if comp.Content != nil {
					if comp.Content.TextBox != nil {
						allContent += comp.Content.TextBox.Text + " "
					}
					if comp.Content.Image != nil {
						allContent += comp.Content.Image.Src + " "
					}
				}
			}

			for _, reqVar := range requiredVars {
				if !strings.Contains(allContent, reqVar) {
					t.Errorf("%s is missing required template variable: %s", theme.name, reqVar)
				}
			}
		})
	}
}

func TestMigratedThemes_CustomImageSupport(t *testing.T) {
	migrator := NewThemeMigrator()

	themes := []struct {
		name    string
		migrate func() (*models.ComponentConfiguration, error)
	}{
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

			hasCustomImageSupport := false
			for _, comp := range config.Components {
				if comp.Type == models.ComponentTypeImage {
					if comp.Content != nil && comp.Content.Image != nil {
						if strings.Contains(comp.Content.Image.Src, "{{if .Event.CustomThemeImageURL}}") {
							hasCustomImageSupport = true
							t.Logf("%s supports custom theme images", theme.name)
							break
						}
					}
				}
			}

			if !hasCustomImageSupport {
				t.Errorf("%s does not support custom theme images", theme.name)
			}
		})
	}
}

func TestMigratedThemes_ZIndexOrdering(t *testing.T) {
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

			for _, comp := range config.Components {
				if comp.Type == models.ComponentTypeBackground && comp.ZIndex != 0 {
					t.Errorf("Background component should have zIndex 0, got %d", comp.ZIndex)
				}

				if comp.Type == models.ComponentTypeImage && comp.ZIndex < 1 {
					t.Errorf("Image component should have zIndex >= 1, got %d", comp.ZIndex)
				}

				if comp.Type == models.ComponentTypeTextBox && comp.ZIndex < 10 {
					t.Errorf("TextBox component should have zIndex >= 10, got %d", comp.ZIndex)
				}
			}
		})
	}
}

func TestMigratedThemes_LayoutConfiguration(t *testing.T) {
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

			if config.Layout.Mode != "card" {
				t.Errorf("Expected layout mode 'card', got %s", config.Layout.Mode)
			}

			if config.Layout.CardWidth == "" {
				t.Error("CardWidth should not be empty")
			}

			if config.Layout.CardMaxWidth == "" {
				t.Error("CardMaxWidth should not be empty")
			}

			if config.Layout.BackgroundColor == "" {
				t.Error("BackgroundColor should not be empty")
			}
		})
	}
}

func TestMigratedThemes_ComponentVisibility(t *testing.T) {
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

			for _, comp := range config.Components {
				if !comp.Visible {
					t.Errorf("Component %s is not visible by default", comp.ID)
				}
			}
		})
	}
}

func TestMigratedThemes_WithOverrides(t *testing.T) {
	engine := NewEngine()
	renderer := NewComponentRenderer(engine)
	migrator := NewThemeMigrator()

	config, err := migrator.MigrateWeddingElegance()
	if err != nil {
		t.Fatalf("Failed to migrate Wedding Elegance: %v", err)
	}

	configJSON := mustMarshalJSON(t, config)

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color":    "#8b4789",
						"fontSize": "3rem",
					},
				},
			},
		},
	}

	overridesJSON := mustMarshalJSON(t, overrides)

	event := &models.Event{
		ID:                 1,
		Title:              "Custom Wedding",
		Description:        stringPtr("A beautiful custom wedding"),
		Location:           stringPtr("Custom Venue"),
		StartTime:          time.Now().Add(24 * time.Hour),
		Timezone:           "America/Los_Angeles",
		ComponentOverrides: &overridesJSON,
	}

	template := &models.Template{
		ID:              1,
		Name:            "Wedding Elegance",
		Type:            models.TemplateTypeRSVPPage,
		ComponentConfig: &configJSON,
	}

	var buf bytes.Buffer
	if err := renderer.Render(&buf, event, template); err != nil {
		t.Fatalf("Failed to render with overrides: %v", err)
	}

	html := buf.String()
	if html == "" {
		t.Fatal("Rendered HTML is empty")
	}

	if !strings.Contains(html, "Custom Wedding") {
		t.Error("Rendered HTML does not contain custom event title")
	}

	t.Logf("Rendered with overrides successfully (%d bytes)", len(html))
}

func TestMigratedThemes_BackwardCompatibility(t *testing.T) {
	seeder := NewSeeder(nil, 1)

	themes := seeder.getDefaultThemes()

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			if theme.HTMLContent == "" {
				t.Error("HTMLContent should not be empty (backward compatibility)")
			}

			if theme.ComponentConfig == nil || *theme.ComponentConfig == "" {
				t.Error("ComponentConfig should not be empty (new system)")
			}

			t.Logf("%s has both HTMLContent (%d bytes) and ComponentConfig (%d bytes)",
				theme.Name, len(theme.HTMLContent), len(*theme.ComponentConfig))
		})
	}
}

func mustMarshalJSON(t *testing.T, v interface{}) string {
	t.Helper()
	bytes := mustMarshalJSONBytes(t, v)
	return string(bytes)
}

func mustMarshalJSONBytes(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	return data
}
