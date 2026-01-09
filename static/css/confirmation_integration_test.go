package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestConfirmationCSSServing(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/css/confirmation.css", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("CSS file should not be empty")
	}
}

func TestConfirmationCSSVariableConsistency(t *testing.T) {
	confirmationCSS, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	variablesCSS, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	confirmation := string(confirmationCSS)
	variables := string(variablesCSS)

	usedVariables := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--color-success",
		"--color-success-light",
		"--color-success-dark",
		"--color-error-light",
		"--color-error-dark",
		"--color-warning-light",
		"--color-warning-dark",
		"--color-primary-600",
		"--color-primary-700",
		"--color-gray-600",
		"--color-gray-700",
		"--color-surface",
		"--color-background",
		"--color-text-primary",
		"--color-text-label",
		"--color-border",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-xl",
		"--font-size-3xl",
		"--font-size-4xl",
		"--font-weight-bold",
		"--radius-base",
		"--transition-base",
	}

	for _, variable := range usedVariables {
		if strings.Contains(confirmation, variable) && !strings.Contains(variables, variable) {
			t.Errorf("Variable %s used in confirmation.css but not defined in variables.css", variable)
		}
	}
}

func TestConfirmationCSSMobileFirst(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	baseStylesIndex := strings.Index(css, ".confirmation {")
	if baseStylesIndex == -1 {
		t.Fatal("Base styles not found")
	}

	tabletBreakpointIndex := strings.Index(css, "@media (min-width: 768px)")
	desktopBreakpointIndex := strings.Index(css, "@media (min-width: 1024px)")

	if tabletBreakpointIndex != -1 && tabletBreakpointIndex < baseStylesIndex {
		t.Error("Mobile-first approach violated: tablet breakpoint appears before base styles")
	}

	if desktopBreakpointIndex != -1 && desktopBreakpointIndex < baseStylesIndex {
		t.Error("Mobile-first approach violated: desktop breakpoint appears before base styles")
	}

	if tabletBreakpointIndex != -1 && desktopBreakpointIndex != -1 {
		if desktopBreakpointIndex < tabletBreakpointIndex {
			t.Error("Breakpoints should be in ascending order: tablet before desktop")
		}
	}
}

func TestConfirmationCSSPrintOptimization(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	printMediaIndex := strings.Index(css, "@media print")
	if printMediaIndex == -1 {
		t.Fatal("Print media query not found")
	}

	printSection := css[printMediaIndex:]

	printOptimizations := []string{
		"display: none",
		"page-break-inside: avoid",
	}

	for _, optimization := range printOptimizations {
		if !strings.Contains(printSection, optimization) {
			t.Errorf("Print optimization missing: %s", optimization)
		}
	}
}

func TestConfirmationCSSAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	accessibilityFeatures := []string{
		":focus",
		"outline:",
		"outline-offset:",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(css, feature) {
			t.Errorf("Accessibility feature missing: %s", feature)
		}
	}
}

func TestConfirmationCSSResponsiveButtons(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, ".calendar-download") {
		t.Error("Calendar download button styles missing")
	}

	if !strings.Contains(css, ".update-rsvp") {
		t.Error("Update RSVP button styles missing")
	}

	mobileBreakpoint := "@media (max-width: 480px)"
	if !strings.Contains(css, mobileBreakpoint) {
		t.Error("Mobile breakpoint for buttons missing")
	}

	mobileSection := css[strings.Index(css, mobileBreakpoint):]
	if !strings.Contains(mobileSection, "width: 100%") {
		t.Error("Mobile buttons should be full width")
	}
}

func TestConfirmationCSSStatusColors(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	statusClasses := []string{
		".response-yes",
		".response-no",
		".response-maybe",
	}

	for _, class := range statusClasses {
		if !strings.Contains(css, class) {
			t.Errorf("Status class missing: %s", class)
		}
	}

	colorVariables := []string{
		"--color-success",
		"--color-error",
		"--color-warning",
	}

	for _, variable := range colorVariables {
		if !strings.Contains(css, variable) {
			t.Errorf("Color variable not used: %s", variable)
		}
	}
}

func TestConfirmationCSSLayoutStructure(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	layoutComponents := []string{
		".confirmation-success",
		".confirmation-summary",
		".confirmation-details",
		".answer-list",
		".confirmation-actions",
	}

	for _, component := range layoutComponents {
		if !strings.Contains(css, component) {
			t.Errorf("Layout component missing: %s", component)
		}
	}
}

func TestConfirmationCSSFlexboxUsage(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	flexProperties := []string{
		"display: flex",
		"flex-wrap:",
		"gap:",
		"justify-content:",
		"align-items:",
	}

	foundFlexbox := false
	for _, property := range flexProperties {
		if strings.Contains(css, property) {
			foundFlexbox = true
			break
		}
	}

	if !foundFlexbox {
		t.Error("CSS should use flexbox for layout")
	}
}

func TestConfirmationCSSTransitions(t *testing.T) {
	content, err := os.ReadFile("confirmation.css")
	if err != nil {
		t.Fatalf("Failed to read confirmation.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "transition:") {
		t.Error("CSS should include transitions for smooth interactions")
	}

	if !strings.Contains(css, ":hover") {
		t.Error("CSS should include hover states")
	}
}
