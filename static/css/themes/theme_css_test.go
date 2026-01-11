package themes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var expectedCSSThemes = []string{
	"plain-text.css",
	"wedding-elegance.css",
	"birthday-celebration.css",
	"corporate-professional.css",
	"holiday-festive.css",
	"garden-party.css",
	"modern-minimalist.css",
}

func TestThemeCSSFilesExist(t *testing.T) {
	for _, cssFile := range expectedCSSThemes {
		t.Run(cssFile, func(t *testing.T) {
			path := filepath.Join(".", cssFile)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("CSS file not found: %s, error: %v", path, err)
			}

			if info.Size() == 0 {
				t.Errorf("CSS file is empty: %s", path)
			}
		})
	}
}

func TestThemeCSSContainsRequiredVariables(t *testing.T) {
	requiredVars := []string{
		"--theme-primary",
		"--theme-secondary",
		"--theme-accent",
		"--theme-font-heading",
		"--theme-font-body",
	}

	for _, cssFile := range expectedCSSThemes {
		t.Run(cssFile, func(t *testing.T) {
			content, err := os.ReadFile(cssFile)
			if err != nil {
				t.Fatalf("Failed to read CSS file: %v", err)
			}

			cssContent := string(content)

			for _, varName := range requiredVars {
				if !strings.Contains(cssContent, varName) {
					t.Errorf("CSS file %s missing required variable: %s", cssFile, varName)
				}
			}
		})
	}
}

func TestThemeCSSHasDarkModeSupport(t *testing.T) {
	cardBasedThemes := []string{
		"wedding-elegance.css",
		"birthday-celebration.css",
		"corporate-professional.css",
		"holiday-festive.css",
		"garden-party.css",
		"modern-minimalist.css",
	}

	for _, cssFile := range cardBasedThemes {
		t.Run(cssFile, func(t *testing.T) {
			content, err := os.ReadFile(cssFile)
			if err != nil {
				t.Fatalf("Failed to read CSS file: %v", err)
			}

			cssContent := string(content)

			if !strings.Contains(cssContent, `[data-theme="dark"]`) {
				t.Errorf("CSS file %s missing dark mode support", cssFile)
			}
		})
	}
}

func TestThemeCSSUsesDataAttribute(t *testing.T) {
	for _, cssFile := range expectedCSSThemes {
		t.Run(cssFile, func(t *testing.T) {
			content, err := os.ReadFile(cssFile)
			if err != nil {
				t.Fatalf("Failed to read CSS file: %v", err)
			}

			cssContent := string(content)
			themeName := strings.TrimSuffix(cssFile, ".css")

			expectedAttr := `[data-event-theme="` + themeName + `"]`
			if !strings.Contains(cssContent, expectedAttr) {
				t.Errorf("CSS file %s missing data-event-theme attribute selector: %s", cssFile, expectedAttr)
			}
		})
	}
}

func TestThemeCSSCount(t *testing.T) {
	expectedCount := 7
	actualCount := len(expectedCSSThemes)

	if actualCount != expectedCount {
		t.Errorf("Expected %d CSS theme files, got %d", expectedCount, actualCount)
	}
}
