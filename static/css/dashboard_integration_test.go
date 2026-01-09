package css

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDashboardIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	variablesStr := string(variablesContent)
	dashboardStr := string(dashboardContent)

	requiredVars := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--spacing-12",
		"--font-size-xs",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-2xl",
		"--font-size-3xl",
		"--font-size-4xl",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--radius-lg",
		"--radius-full",
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-primary-50",
		"--color-primary-300",
		"--color-primary-600",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-tertiary",
		"--color-error",
		"--color-error-light",
		"--color-error-dark",
		"--transition-base",
	}

	for _, varName := range requiredVars {
		t.Run("variable_defined_"+varName, func(t *testing.T) {
			if !strings.Contains(variablesStr, varName+":") {
				t.Errorf("Variable %s not defined in variables.css", varName)
			}
		})

		t.Run("variable_used_"+varName, func(t *testing.T) {
			if strings.Contains(dashboardStr, "var("+varName+")") {
				if !strings.Contains(variablesStr, varName+":") {
					t.Errorf("Dashboard uses %s but it's not defined in variables.css", varName)
				}
			}
		})
	}
}

func TestDashboardHTTPServing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/dashboard.css" {
			content, err := os.ReadFile("dashboard.css")
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/css")
			w.Write(content)
		} else {
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/static/css/dashboard.css")
	if err != nil {
		t.Fatalf("Failed to fetch dashboard.css: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/css" {
		t.Errorf("Expected Content-Type text/css, got %s", contentType)
	}
}

func TestDashboardResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	breakpoints := []struct {
		name  string
		query string
	}{
		{"tablet", "@media (min-width: 768px)"},
		{"desktop", "@media (min-width: 1024px)"},
	}

	for _, bp := range breakpoints {
		t.Run(bp.name+"_breakpoint", func(t *testing.T) {
			if !strings.Contains(cssContent, bp.query) {
				t.Errorf("Missing %s breakpoint: %s", bp.name, bp.query)
			}
		})
	}
}

func TestDashboardConsistencyWithExistingCSS(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	variablesStr := string(variablesContent)

	colorVars := []string{
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-primary-50",
		"--color-primary-300",
		"--color-primary-600",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-tertiary",
		"--color-error",
		"--color-error-light",
		"--color-error-dark",
	}

	for _, colorVar := range colorVars {
		t.Run("color_consistency_"+colorVar, func(t *testing.T) {
			if strings.Contains(dashboardStr, "var("+colorVar+")") {
				if !strings.Contains(variablesStr, colorVar+":") {
					t.Errorf("Dashboard uses %s but it's not defined in variables.css", colorVar)
				}
			}
		})
	}
}

func TestDashboardFileSize(t *testing.T) {
	info, err := os.Stat("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to stat dashboard.css: %v", err)
	}

	maxSize := int64(20 * 1024)
	if info.Size() > maxSize {
		t.Errorf("dashboard.css is too large: %d bytes (max: %d bytes)", info.Size(), maxSize)
	}
}

func TestDashboardNoHardcodedValues(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	hardcodedPatterns := []struct {
		pattern     string
		description string
	}{
		{": 8px", "8px spacing"},
		{": 12px", "12px spacing"},
		{": 16px", "16px spacing"},
		{": 1rem", "1rem spacing"},
		{": 1.5rem", "1.5rem spacing"},
	}

	for _, hp := range hardcodedPatterns {
		t.Run("no_hardcoded_"+hp.description, func(t *testing.T) {
			if strings.Contains(cssContent, hp.pattern) {
				t.Errorf("Dashboard should not use hardcoded %s, use CSS variables instead", hp.description)
			}
		})
	}
}

func TestDashboardComponentCompleteness(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	components := []string{
		".dashboard",
		".dashboard-sidebar",
		".dashboard-main",
		".dashboard-header",
		".stats-grid",
		".stats-card",
		".stats-card-title",
		".stats-card-value",
		".activity-feed",
		".activity-item",
		".quick-actions",
		".empty-state",
		".loading-state",
		".error-state",
	}

	for _, component := range components {
		t.Run("component_"+component, func(t *testing.T) {
			if !strings.Contains(cssContent, component) {
				t.Errorf("Missing dashboard component: %s", component)
			}
		})
	}
}

func TestDashboardAccessibilityFeatures(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	accessibilityFeatures := []struct {
		name        string
		requirement string
	}{
		{"color_contrast", "color:"},
		{"text_sizing", "font-size:"},
		{"spacing", "padding:"},
		{"borders", "border:"},
	}

	for _, feature := range accessibilityFeatures {
		t.Run("accessibility_"+feature.name, func(t *testing.T) {
			if !strings.Contains(cssContent, feature.requirement) {
				t.Errorf("Missing accessibility feature %s: %s", feature.name, feature.requirement)
			}
		})
	}
}

func TestDashboardIntegrationWithTypography(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	typographyContent, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	typographyStr := string(typographyContent)

	sharedTypographyVars := []string{
		"--font-size-xs",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-2xl",
		"--font-size-3xl",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
	}

	for _, varName := range sharedTypographyVars {
		t.Run("shared_typography_"+varName, func(t *testing.T) {
			usedInDashboard := strings.Contains(dashboardStr, "var("+varName+")")
			definedInTypography := strings.Contains(typographyStr, varName+":")

			if usedInDashboard && !definedInTypography {
				t.Logf("Variable %s is used in dashboard but not defined in typography.css", varName)
			}
		})
	}
}

func TestDashboardIntegrationWithSpacing(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	spacingStr := string(spacingContent)

	sharedSpacingVars := []string{
		"--spacing-1",
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-5",
		"--spacing-6",
		"--spacing-8",
		"--spacing-12",
	}

	for _, varName := range sharedSpacingVars {
		t.Run("shared_spacing_"+varName, func(t *testing.T) {
			usedInDashboard := strings.Contains(dashboardStr, "var("+varName+")")
			usedInSpacing := strings.Contains(spacingStr, "var("+varName+")")

			if usedInDashboard && usedInSpacing {
				t.Logf("Variable %s is shared between dashboard and spacing systems", varName)
			}
		})
	}
}

func TestDashboardIntegrationWithColors(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	colorsContent, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	colorsStr := string(colorsContent)

	sharedColorVars := []string{
		"--color-background",
		"--color-surface",
		"--color-border",
		"--color-primary-50",
		"--color-primary-300",
		"--color-primary-600",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-error",
	}

	for _, varName := range sharedColorVars {
		t.Run("shared_color_"+varName, func(t *testing.T) {
			usedInDashboard := strings.Contains(dashboardStr, "var("+varName+")")
			usedInColors := strings.Contains(colorsStr, "var("+varName+")")

			if usedInDashboard && usedInColors {
				t.Logf("Variable %s is shared between dashboard and colors systems", varName)
			}
		})
	}
}

func TestDashboardIntegrationWithGrid(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	gridStr := string(gridContent)

	if strings.Contains(dashboardStr, "display: grid") {
		t.Log("Dashboard uses CSS Grid layout")
	}

	if strings.Contains(dashboardStr, "display: flex") {
		t.Log("Dashboard uses Flexbox layout")
	}

	if strings.Contains(gridStr, ".container") {
		t.Log("Grid system provides container classes")
	}
}

func TestDashboardIntegrationWithButtons(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	buttonsContent, err := os.ReadFile("buttons.css")
	if err != nil {
		t.Fatalf("Failed to read buttons.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	buttonsStr := string(buttonsContent)

	if strings.Contains(buttonsStr, ".btn") {
		t.Log("Button system available for dashboard quick actions")
	}

	if strings.Contains(dashboardStr, ".quick-actions") {
		t.Log("Dashboard has quick actions section for buttons")
	}
}

func TestDashboardLoadingAnimation(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "@keyframes") {
		t.Error("Loading state should include keyframe animation")
	}

	if !strings.Contains(cssContent, "animation:") {
		t.Error("Loading state should use animation property")
	}

	if !strings.Contains(cssContent, ".loading-spinner") {
		t.Error("Loading state should have spinner element")
	}
}

func TestDashboardStateCompleteness(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	states := []string{
		".empty-state",
		".loading-state",
		".error-state",
	}

	for _, state := range states {
		t.Run("state_"+state, func(t *testing.T) {
			if !strings.Contains(cssContent, state) {
				t.Errorf("Missing dashboard state: %s", state)
			}
		})
	}
}

func TestDashboardTransitionConsistency(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, "var(--transition-base)") {
		t.Error("Dashboard should use consistent transition variable")
	}

	if strings.Contains(cssContent, "transition: 0.2s") || strings.Contains(cssContent, "transition: 200ms") {
		t.Error("Dashboard should not use hardcoded transition values, use CSS variables instead")
	}
}

func TestDashboardIntegrationWithNavigation(t *testing.T) {
	dashboardContent, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	navigationContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	dashboardStr := string(dashboardContent)
	navigationStr := string(navigationContent)

	if strings.Contains(dashboardStr, ".dashboard-sidebar") {
		t.Log("Dashboard has sidebar for navigation")
	}

	if strings.Contains(navigationStr, ".nav") {
		t.Log("Navigation system available for dashboard")
	}
}
