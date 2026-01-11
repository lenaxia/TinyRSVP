package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

func main() {
	migrator := templates.NewThemeMigrator()

	themes := []struct {
		name     string
		filename string
		migrate  func() (*models.ComponentConfiguration, error)
	}{
		{"Plain Text", "plain-text.json", migrator.MigratePlainText},
		{"Wedding Elegance", "wedding-elegance.json", migrator.MigrateWeddingElegance},
		{"Birthday Celebration", "birthday-celebration.json", migrator.MigrateBirthdayCelebration},
		{"Corporate Professional", "corporate-professional.json", migrator.MigrateCorporateProfessional},
		{"Holiday Festive", "holiday-festive.json", migrator.MigrateHolidayFestive},
		{"Garden Party", "garden-party.json", migrator.MigrateGardenParty},
		{"Modern Minimalist", "modern-minimalist.json", migrator.MigrateModernMinimalist},
	}

	outputDir := "internal/templates/defaults/component_configs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	for _, theme := range themes {
		fmt.Printf("Migrating %s...\n", theme.name)

		config, err := theme.migrate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to migrate %s: %v\n", theme.name, err)
			os.Exit(1)
		}

		jsonBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal %s to JSON: %v\n", theme.name, err)
			os.Exit(1)
		}

		outputPath := filepath.Join(outputDir, theme.filename)
		if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", outputPath, err)
			os.Exit(1)
		}

		fmt.Printf("✓ Saved %s to %s\n", theme.name, outputPath)
	}

	fmt.Println("\nAll themes migrated successfully!")
}
