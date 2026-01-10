package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRSVPSummaryIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	rsvpSummaryContent, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	variables := string(variablesContent)
	rsvpSummary := string(rsvpSummaryContent)

	requiredVariables := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--spacing-8",
		"--spacing-12",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-tertiary",
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-border-focus",
		"--color-primary-600",
		"--color-success",
		"--color-warning",
		"--color-info",
		"--color-error",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-xl",
		"--font-size-2xl",
		"--font-size-3xl",
		"--font-size-4xl",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--radius-md",
		"--radius-lg",
	}

	for _, variable := range requiredVariables {
		if !strings.Contains(variables, variable+":") {
			t.Errorf("variables.css missing required variable: %s", variable)
		}
	}

	usedVariables := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--font-",
		"var(--radius-",
	}

	for _, pattern := range usedVariables {
		if !strings.Contains(rsvpSummary, pattern) {
			t.Errorf("rsvp_summary.css should use variable pattern: %s", pattern)
		}
	}
}

func TestRSVPSummaryIntegrationWithGrid(t *testing.T) {
	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	rsvpSummaryContent, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	grid := string(gridContent)
	rsvpSummary := string(rsvpSummaryContent)

	if !strings.Contains(grid, "display: grid") {
		t.Error("grid.css should define grid display")
	}

	if !strings.Contains(rsvpSummary, "display: grid") {
		t.Error("rsvp_summary.css should use CSS Grid for layout")
	}

	if !strings.Contains(rsvpSummary, "grid-template-columns") {
		t.Error("rsvp_summary.css should use grid-template-columns")
	}
}

func TestRSVPSummaryHTTPServing(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/css/rsvp_summary.css", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}
}

func TestRSVPSummaryFileSize(t *testing.T) {
	info, err := os.Stat("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to stat rsvp_summary.css: %v", err)
	}

	maxSize := int64(50 * 1024)
	if info.Size() > maxSize {
		t.Errorf("rsvp_summary.css is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestRSVPSummaryResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	css := string(content)

	breakpoints := map[string]string{
		"tablet":  "@media (min-width: 768px)",
		"desktop": "@media (min-width: 1024px)",
	}

	for name, breakpoint := range breakpoints {
		if !strings.Contains(css, breakpoint) {
			t.Errorf("Missing %s breakpoint: %s", name, breakpoint)
		}
	}
}

func TestRSVPSummaryAccessibilityFeatures(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	forms := string(formsContent)

	accessibilityFeatures := []string{
		":focus",
		"outline:",
		"cursor:",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(forms, feature) {
			t.Errorf("forms.css missing accessibility feature: %s", feature)
		}
	}
}

func TestRSVPSummaryAnimations(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "@keyframes") {
		t.Error("rsvp_summary.css should include keyframe animations")
	}

	if !strings.Contains(css, "animation:") {
		t.Error("rsvp_summary.css should use animations")
	}

	if !strings.Contains(css, "transition:") {
		t.Error("rsvp_summary.css should use transitions for smooth interactions")
	}
}

func TestRSVPSummaryIntegrationWithButtons(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	rsvpSummaryContent, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	buttons := string(buttonsContent)
	rsvpSummary := string(rsvpSummaryContent)

	if !strings.Contains(buttons, ".btn") {
		t.Error("buttons.css should define .btn class")
	}

	if strings.Contains(rsvpSummary, ".btn") || strings.Contains(rsvpSummary, "btn-") {
		t.Log("rsvp_summary.css references button styles from buttons.css")
	}
}

func TestRSVPSummaryIntegrationWithForms(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	rsvpSummaryContent, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	forms := string(formsContent)
	rsvpSummary := string(rsvpSummaryContent)

	if !strings.Contains(forms, "input") || !strings.Contains(forms, "select") {
		t.Error("forms.css should define input and select styles")
	}

	formElements := []string{
		"select",
	}

	for _, element := range formElements {
		if strings.Contains(rsvpSummary, element) {
			t.Logf("rsvp_summary.css uses form element: %s", element)
		}
	}
}

func TestRSVPSummaryMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	css := string(content)

	mobileFirstIndicators := []string{
		".stats-grid",
		"@media (min-width:",
	}

	for _, indicator := range mobileFirstIndicators {
		if !strings.Contains(css, indicator) {
			t.Errorf("Missing mobile-first indicator: %s", indicator)
		}
	}
}

func TestRSVPSummaryChartVisualization(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	css := string(content)

	chartElements := []string{
		".chart-container",
		".chart-bars",
		".chart-bar",
		".chart-bar-yes",
		".chart-bar-no",
		".chart-bar-maybe",
		".chart-bar-pending",
	}

	for _, element := range chartElements {
		if !strings.Contains(css, element) {
			t.Errorf("Missing chart element: %s", element)
		}
	}
}

func TestRSVPSummaryResponseRateCircle(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.css")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.css: %v", err)
	}

	css := string(content)

	circleElements := []string{
		".response-rate-circle",
		".response-rate-svg",
		".response-rate-bg",
		".response-rate-fill",
		".response-rate-percentage",
	}

	for _, element := range circleElements {
		if !strings.Contains(css, element) {
			t.Errorf("Missing response rate circle element: %s", element)
		}
	}
}
