package css

import (
	"os"
	"strings"
	"testing"
)

func TestConfirmationCSSExists(t *testing.T) {
	if _, err := os.Stat("confirmation.css"); os.IsNotExist(err) {
		t.Fatal("confirmation.css does not exist")
	}
}

func TestConfirmationCSSContent(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".confirmation",
		".confirmation-success",
		".confirmation-summary",
		".confirmation-details",
		".confirmation-actions",
		".response-status",
		".answer-list",
		".calendar-download",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Missing required class: %s", class)
		}
	}
}

func TestConfirmationUsesVariables(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	requiredVariables := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--font-",
		"var(--radius-",
	}

	for _, variable := range requiredVariables {
		if !strings.Contains(css, variable) {
			t.Errorf("CSS should use variable pattern: %s", variable)
		}
	}
}

func TestConfirmationResponsive(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	breakpoints := []string{
		"@media (min-width: 768px)",
		"@media (min-width: 1024px)",
	}

	for _, breakpoint := range breakpoints {
		if !strings.Contains(css, breakpoint) {
			t.Errorf("Missing responsive breakpoint: %s", breakpoint)
		}
	}
}

func TestConfirmationSuccessComponents(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	successComponents := []string{
		".confirmation-success-icon",
		".confirmation-success-title",
		".confirmation-success-message",
	}

	for _, component := range successComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing success component: %s", component)
		}
	}
}

func TestConfirmationSummaryComponents(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	summaryComponents := []string{
		".confirmation-summary-title",
		".confirmation-summary-item",
		".response-yes",
		".response-no",
		".response-maybe",
	}

	for _, component := range summaryComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing summary component: %s", component)
		}
	}
}

func TestConfirmationAnswerComponents(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	answerComponents := []string{
		".answer-item",
		".answer-label",
		".answer-value",
	}

	for _, component := range answerComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing answer component: %s", component)
		}
	}
}

func TestConfirmationPrintStyles(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "@media print") {
		t.Error("CSS should include print styles")
	}
}

func TestConfirmationAccessibility(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ":focus") {
		t.Error("CSS should include focus styles for accessibility")
	}
}

func TestConfirmationNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	hardcodedPatterns := []string{
		"#fff",
		"#000",
		"16px",
		"14px",
	}

	for _, pattern := range hardcodedPatterns {
		if strings.Contains(css, pattern) {
			t.Errorf("CSS should not contain hardcoded value: %s (use CSS variables instead)", pattern)
		}
	}
}

func TestConfirmationActionButtons(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	actionComponents := []string{
		".confirmation-actions",
		".calendar-download",
		".update-rsvp",
	}

	for _, component := range actionComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Missing action component: %s", component)
		}
	}
}
