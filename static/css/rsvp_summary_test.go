package css

import (
	"os"
	"strings"
	"testing"
)

func TestRSVPSummaryCSS_FileExists(t *testing.T) {
	_, err := os.Stat("rsvp_summary.css")
	if err != nil {
		t.Fatalf("rsvp_summary.css file should exist: %v", err)
	}
}

func TestRSVPSummaryCSS_ContainsMainClass(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".rsvp-summary") {
		t.Error("Expected CSS to contain .rsvp-summary class")
	}
}

func TestRSVPSummaryCSS_ContainsStatsGrid(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".stats-grid") {
		t.Error("Expected CSS to contain .stats-grid class")
	}
}

func TestRSVPSummaryCSS_ContainsStatCard(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".stat-card") {
		t.Error("Expected CSS to contain .stat-card class")
	}
}

func TestRSVPSummaryCSS_ContainsResponseRateStyles(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".response-rate-card",
		".response-rate-circle",
		".response-rate-svg",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Expected CSS to contain %s class", class)
		}
	}
}

func TestRSVPSummaryCSS_ContainsChartStyles(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".chart-container",
		".chart-bars",
		".chart-bar",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Expected CSS to contain %s class", class)
		}
	}
}

func TestRSVPSummaryCSS_ContainsQuestionStyles(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	requiredClasses := []string{
		".questions-grid",
		".question-card",
		".response-bar",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Expected CSS to contain %s class", class)
		}
	}
}

func TestRSVPSummaryCSS_ContainsResponsiveDesign(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "@media") {
		t.Error("Expected CSS to contain responsive media queries")
	}

	if !strings.Contains(css, "min-width: 768px") {
		t.Error("Expected CSS to contain tablet breakpoint")
	}

	if !strings.Contains(css, "min-width: 1024px") {
		t.Error("Expected CSS to contain desktop breakpoint")
	}
}

func TestRSVPSummaryCSS_ContainsLoadingState(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".rsvp-summary-loading") {
		t.Error("Expected CSS to contain loading state styles")
	}

	if !strings.Contains(css, ".loading-spinner") {
		t.Error("Expected CSS to contain loading spinner styles")
	}
}

func TestRSVPSummaryCSS_ContainsErrorState(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".rsvp-summary-error") {
		t.Error("Expected CSS to contain error state styles")
	}
}

func TestRSVPSummaryCSS_UsesCSSVariables(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	requiredVars := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--font-size-",
	}

	for _, varPrefix := range requiredVars {
		if !strings.Contains(css, varPrefix) {
			t.Errorf("Expected CSS to use CSS variables with prefix %s", varPrefix)
		}
	}
}

func TestRSVPSummaryCSS_ContainsFilterStyles(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".filter-group") {
		t.Error("Expected CSS to contain filter-group styles")
	}
}

func TestRSVPSummaryCSS_ContainsExportButtonStyles(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read CSS file: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".export-btn") {
		t.Error("Expected CSS to contain export button styles")
	}
}
