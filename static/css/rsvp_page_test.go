package css

import (
	"os"
	"strings"
	"testing"
)

func TestRSVPPageCSSExists(t *testing.T) {
	_, err := os.Stat("rsvp_page.css")
	if err != nil {
		t.Fatalf("rsvp_page.css file does not exist: %v", err)
	}
}

func TestRSVPPageCSSNotEmpty(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	if len(content) == 0 {
		t.Error("rsvp_page.css is empty")
	}
}

func TestRSVPPageCSSValidSyntax(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	openBraces := strings.Count(css, "{")
	closeBraces := strings.Count(css, "}")

	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d opening, %d closing", openBraces, closeBraces)
	}
}

func TestRSVPPageCSSContainsEventDetailsSection(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".rsvp-page",
		".event-details",
		".event-title",
		".event-info",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestRSVPPageCSSContainsRSVPFormSection(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".rsvp-form",
		".response-options",
		".response-option",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestRSVPPageCSSContainsPlusOnesSection(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".plus-ones-selector",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestRSVPPageCSSContainsQuestionsSection(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".preference-questions",
		".question-item",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestRSVPPageCSSContainsMobileOptimizations(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "@media") {
		t.Error("Missing media queries for responsive design")
	}
}

func TestRSVPPageCSSContainsLoadingState(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".btn-loading") || !strings.Contains(css, ".loading") {
		t.Error("Missing loading state styles")
	}
}

func TestRSVPPageCSSContainsErrorState(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".error-message",
		".alert",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestRSVPPageCSSContainsTouchFriendlyControls(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "min-height: 44px") && !strings.Contains(css, "min-height:44px") {
		t.Error("Missing touch-friendly minimum height (44px) for interactive elements")
	}
}

func TestRSVPPageCSSUsesVariables(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "var(--") {
		t.Error("CSS should use CSS variables for consistency")
	}
}
