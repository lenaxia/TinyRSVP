package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestEventListIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	eventListContent, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	variables := string(variablesContent)
	eventList := string(eventListContent)

	requiredVariables := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--spacing-10",
		"--spacing-12",
		"--spacing-16",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-tertiary",
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-border-focus",
		"--color-primary-600",
		"--color-success",
		"--color-success-light",
		"--color-warning",
		"--color-warning-light",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-xl",
		"--font-size-2xl",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--radius-md",
		"--radius-lg",
		"--radius-full",
		"--line-height-relaxed",
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
		"var(--line-height-",
	}

	for _, pattern := range usedVariables {
		if !strings.Contains(eventList, pattern) {
			t.Errorf("event_list.css should use variable pattern: %s", pattern)
		}
	}
}

func TestEventListIntegrationWithGrid(t *testing.T) {
	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	eventListContent, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	grid := string(gridContent)
	eventList := string(eventListContent)

	if !strings.Contains(grid, "display: grid") {
		t.Error("grid.css should define grid display")
	}

	if !strings.Contains(eventList, "display: grid") {
		t.Error("event_list.css should use CSS Grid for layout")
	}

	if !strings.Contains(eventList, "grid-template-columns") {
		t.Error("event_list.css should use grid-template-columns")
	}
}

func TestEventListHTTPServing(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(content)
	})

	req := httptest.NewRequest("GET", "/static/css/event_list.css", nil)
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

func TestEventListFileSize(t *testing.T) {
	info, err := os.Stat("event_list.css")
	if err != nil {
		t.Fatalf("Failed to stat event_list.css: %v", err)
	}

	maxSize := int64(50 * 1024)
	if info.Size() > maxSize {
		t.Errorf("event_list.css is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestEventListResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
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

func TestEventListAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	accessibilityFeatures := []string{
		":focus",
		"outline:",
		"cursor:",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(css, feature) {
			t.Errorf("Missing accessibility feature: %s", feature)
		}
	}
}

func TestEventListAnimations(t *testing.T) {
	content, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	css := string(content)

	if !strings.Contains(css, "@keyframes") {
		t.Error("event_list.css should include keyframe animations")
	}

	if !strings.Contains(css, "animation:") {
		t.Error("event_list.css should use animations")
	}

	if !strings.Contains(css, "transition:") {
		t.Error("event_list.css should use transitions for smooth interactions")
	}
}

func TestEventListIntegrationWithButtons(t *testing.T) {
	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	eventListContent, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	buttons := string(buttonsContent)
	eventList := string(eventListContent)

	if !strings.Contains(buttons, ".btn") {
		t.Error("buttons.css should define .btn class")
	}

	if strings.Contains(eventList, ".btn") {
		t.Log("event_list.css references button styles from buttons.css")
	}
}

func TestEventListIntegrationWithForms(t *testing.T) {
	formsContent, err := os.ReadFile("forms.css")
	if err != nil {
		t.Fatalf("Failed to read forms.css: %v", err)
	}

	eventListContent, err := os.ReadFile("event_list.css")
	if err != nil {
		t.Fatalf("Failed to read event_list.css: %v", err)
	}

	forms := string(formsContent)
	eventList := string(eventListContent)

	if !strings.Contains(forms, "input") || !strings.Contains(forms, "select") {
		t.Error("forms.css should define input and select styles")
	}

	formElements := []string{
		"input",
		"select",
	}

	for _, element := range formElements {
		if strings.Contains(eventList, element) {
			t.Logf("event_list.css uses form element: %s", element)
		}
	}
}
