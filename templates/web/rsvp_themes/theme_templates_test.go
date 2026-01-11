package rsvp_themes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var expectedHTMLTemplates = []string{
	"plain-text.html",
	"wedding-elegance.html",
	"birthday-celebration.html",
	"corporate-professional.html",
	"holiday-festive.html",
	"garden-party.html",
	"modern-minimalist.html",
}

func TestThemeHTMLTemplatesExist(t *testing.T) {
	for _, htmlFile := range expectedHTMLTemplates {
		t.Run(htmlFile, func(t *testing.T) {
			path := filepath.Join(".", htmlFile)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("HTML template not found: %s, error: %v", path, err)
			}

			if info.Size() == 0 {
				t.Errorf("HTML template is empty: %s", path)
			}
		})
	}
}

func TestThemeHTMLContainsRequiredStructure(t *testing.T) {
	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"data-event-theme=",
		"<head>",
		"<body>",
		"rsvp-container",
		"rsvp-card",
		"event-title",
		"rsvp-form",
	}

	for _, htmlFile := range expectedHTMLTemplates {
		t.Run(htmlFile, func(t *testing.T) {
			content, err := os.ReadFile(htmlFile)
			if err != nil {
				t.Fatalf("Failed to read HTML template: %v", err)
			}

			htmlContent := string(content)

			for _, element := range requiredElements {
				if !strings.Contains(htmlContent, element) {
					t.Errorf("HTML template %s missing required element: %s", htmlFile, element)
				}
			}
		})
	}
}

func TestThemeHTMLContainsGoTemplateVariables(t *testing.T) {
	requiredVars := []string{
		"{{.Event.Title}}",
		"{{.CSRFToken}}",
	}

	for _, htmlFile := range expectedHTMLTemplates {
		t.Run(htmlFile, func(t *testing.T) {
			content, err := os.ReadFile(htmlFile)
			if err != nil {
				t.Fatalf("Failed to read HTML template: %v", err)
			}

			htmlContent := string(content)

			for _, varName := range requiredVars {
				if !strings.Contains(htmlContent, varName) {
					t.Errorf("HTML template %s missing required Go template variable: %s", htmlFile, varName)
				}
			}
		})
	}
}

func TestThemeHTMLLinksToCorrectCSS(t *testing.T) {
	for _, htmlFile := range expectedHTMLTemplates {
		t.Run(htmlFile, func(t *testing.T) {
			content, err := os.ReadFile(htmlFile)
			if err != nil {
				t.Fatalf("Failed to read HTML template: %v", err)
			}

			htmlContent := string(content)
			themeName := strings.TrimSuffix(htmlFile, ".html")
			expectedCSSLink := "/static/css/themes/" + themeName + ".css"

			if !strings.Contains(htmlContent, expectedCSSLink) {
				t.Errorf("HTML template %s does not link to correct CSS file: %s", htmlFile, expectedCSSLink)
			}
		})
	}
}

func TestThemeHTMLHasCorrectDataAttribute(t *testing.T) {
	for _, htmlFile := range expectedHTMLTemplates {
		t.Run(htmlFile, func(t *testing.T) {
			content, err := os.ReadFile(htmlFile)
			if err != nil {
				t.Fatalf("Failed to read HTML template: %v", err)
			}

			htmlContent := string(content)
			themeName := strings.TrimSuffix(htmlFile, ".html")
			expectedAttr := `data-event-theme="` + themeName + `"`

			if !strings.Contains(htmlContent, expectedAttr) {
				t.Errorf("HTML template %s missing correct data-event-theme attribute: %s", htmlFile, expectedAttr)
			}
		})
	}
}

func TestCardBasedThemesHaveHeaderImage(t *testing.T) {
	cardBasedThemes := []string{
		"wedding-elegance.html",
		"birthday-celebration.html",
		"corporate-professional.html",
		"holiday-festive.html",
		"garden-party.html",
		"modern-minimalist.html",
	}

	for _, htmlFile := range cardBasedThemes {
		t.Run(htmlFile, func(t *testing.T) {
			content, err := os.ReadFile(htmlFile)
			if err != nil {
				t.Fatalf("Failed to read HTML template: %v", err)
			}

			htmlContent := string(content)

			if !strings.Contains(htmlContent, "rsvp-card-header") {
				t.Errorf("Card-based theme %s missing rsvp-card-header element", htmlFile)
			}

			if !strings.Contains(htmlContent, "theme-header-image") {
				t.Errorf("Card-based theme %s missing theme-header-image element", htmlFile)
			}
		})
	}
}

func TestPlainTextThemeHasNoHeaderImage(t *testing.T) {
	content, err := os.ReadFile("plain-text.html")
	if err != nil {
		t.Fatalf("Failed to read plain-text.html: %v", err)
	}

	htmlContent := string(content)

	if strings.Contains(htmlContent, "rsvp-card-header") {
		t.Errorf("Plain text theme should not have rsvp-card-header element")
	}

	if strings.Contains(htmlContent, "theme-header-image") {
		t.Errorf("Plain text theme should not have theme-header-image element")
	}
}

func TestThemeHTMLCount(t *testing.T) {
	expectedCount := 7
	actualCount := len(expectedHTMLTemplates)

	if actualCount != expectedCount {
		t.Errorf("Expected %d HTML templates, got %d", expectedCount, actualCount)
	}
}
