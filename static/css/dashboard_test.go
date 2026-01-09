package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestDashboardFileExists(t *testing.T) {
	if _, err := os.Stat("dashboard.css"); os.IsNotExist(err) {
		t.Fatal("dashboard.css file does not exist")
	}
}

func TestDashboardValidCSS(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if strings.TrimSpace(cssContent) == "" {
		t.Fatal("dashboard.css is empty")
	}

	openBraces := strings.Count(cssContent, "{")
	closeBraces := strings.Count(cssContent, "}")
	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}
}

func TestDashboardLayoutClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".dashboard") {
		t.Error("Missing .dashboard class")
	}

	requiredProperties := []string{
		"display:",
		"min-height:",
	}

	for _, prop := range requiredProperties {
		t.Run("dashboard_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".dashboard class should have %s property", prop)
			}
		})
	}
}

func TestDashboardSidebarClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".dashboard-sidebar") {
		t.Error("Missing .dashboard-sidebar class")
	}

	requiredProperties := []string{
		"width:",
		"background-color:",
		"padding:",
	}

	for _, prop := range requiredProperties {
		t.Run("sidebar_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".dashboard-sidebar should have %s property", prop)
			}
		})
	}
}

func TestDashboardMainClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".dashboard-main") {
		t.Error("Missing .dashboard-main class")
	}

	requiredProperties := []string{
		"flex:",
		"padding:",
	}

	for _, prop := range requiredProperties {
		t.Run("main_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".dashboard-main should have %s property", prop)
			}
		})
	}
}

func TestDashboardStatsCardClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".stats-card") {
		t.Error("Missing .stats-card class")
	}

	requiredProperties := []string{
		"background-color:",
		"padding:",
		"border-radius:",
		"border:",
	}

	for _, prop := range requiredProperties {
		t.Run("stats_card_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".stats-card should have %s property", prop)
			}
		})
	}
}

func TestDashboardStatsCardTitle(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".stats-card-title") {
		t.Error("Missing .stats-card-title class")
	}

	requiredProperties := []string{
		"font-size:",
		"color:",
		"margin-bottom:",
	}

	for _, prop := range requiredProperties {
		t.Run("stats_card_title_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".stats-card-title should have %s property", prop)
			}
		})
	}
}

func TestDashboardStatsCardValue(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".stats-card-value") {
		t.Error("Missing .stats-card-value class")
	}

	requiredProperties := []string{
		"font-size:",
		"font-weight:",
		"color:",
	}

	for _, prop := range requiredProperties {
		t.Run("stats_card_value_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".stats-card-value should have %s property", prop)
			}
		})
	}
}

func TestDashboardActivityFeedClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".activity-feed") {
		t.Error("Missing .activity-feed class")
	}

	requiredProperties := []string{
		"background-color:",
		"padding:",
		"border-radius:",
	}

	for _, prop := range requiredProperties {
		t.Run("activity_feed_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".activity-feed should have %s property", prop)
			}
		})
	}
}

func TestDashboardActivityItemClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".activity-item") {
		t.Error("Missing .activity-item class")
	}

	requiredProperties := []string{
		"padding:",
		"border-bottom:",
	}

	for _, prop := range requiredProperties {
		t.Run("activity_item_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".activity-item should have %s property", prop)
			}
		})
	}
}

func TestDashboardEmptyStateClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".empty-state") {
		t.Error("Missing .empty-state class")
	}

	requiredProperties := []string{
		"text-align:",
		"padding:",
		"color:",
	}

	for _, prop := range requiredProperties {
		t.Run("empty_state_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".empty-state should have %s property", prop)
			}
		})
	}
}

func TestDashboardLoadingStateClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".loading-state") {
		t.Error("Missing .loading-state class")
	}

	requiredProperties := []string{
		"text-align:",
		"padding:",
	}

	for _, prop := range requiredProperties {
		t.Run("loading_state_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".loading-state should have %s property", prop)
			}
		})
	}
}

func TestDashboardErrorStateClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".error-state") {
		t.Error("Missing .error-state class")
	}

	requiredProperties := []string{
		"background-color:",
		"padding:",
		"border-radius:",
		"color:",
	}

	for _, prop := range requiredProperties {
		t.Run("error_state_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".error-state should have %s property", prop)
			}
		})
	}
}

func TestDashboardUsesVariables(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	requiredVars := []string{
		"var(--spacing-",
		"var(--font-size-",
		"var(--font-weight-",
		"var(--radius-",
		"var(--color-",
	}

	for _, varPrefix := range requiredVars {
		t.Run("uses_"+varPrefix, func(t *testing.T) {
			if !strings.Contains(cssContent, varPrefix) {
				t.Errorf("Dashboard should use CSS variables with prefix: %s", varPrefix)
			}
		})
	}
}

func TestDashboardNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	hexColorPattern := regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
	if hexColorPattern.MatchString(cssContent) {
		t.Error("Dashboard should not use hardcoded hex colors, use CSS variables instead")
	}

	rgbPattern := regexp.MustCompile(`rgb\(`)
	if rgbPattern.MatchString(cssContent) {
		t.Error("Dashboard should not use hardcoded rgb colors, use CSS variables instead")
	}
}

func TestDashboardResponsiveLayout(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	breakpoints := []string{
		"@media (min-width: 768px)",
		"@media (min-width: 1024px)",
	}

	for _, bp := range breakpoints {
		t.Run("breakpoint_"+bp, func(t *testing.T) {
			if !strings.Contains(cssContent, bp) {
				t.Errorf("Missing responsive breakpoint: %s", bp)
			}
		})
	}
}

func TestDashboardStatsGridClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".stats-grid") {
		t.Error("Missing .stats-grid class")
	}

	requiredProperties := []string{
		"display:",
		"gap:",
	}

	for _, prop := range requiredProperties {
		t.Run("stats_grid_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".stats-grid should have %s property", prop)
			}
		})
	}
}

func TestDashboardQuickActionsClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".quick-actions") {
		t.Error("Missing .quick-actions class")
	}

	requiredProperties := []string{
		"display:",
		"gap:",
		"margin-top:",
	}

	for _, prop := range requiredProperties {
		t.Run("quick_actions_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".quick-actions should have %s property", prop)
			}
		})
	}
}

func TestDashboardHeaderClass(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".dashboard-header") {
		t.Error("Missing .dashboard-header class")
	}

	requiredProperties := []string{
		"display:",
		"justify-content:",
		"align-items:",
		"margin-bottom:",
	}

	for _, prop := range requiredProperties {
		t.Run("dashboard_header_property_"+prop, func(t *testing.T) {
			if !strings.Contains(cssContent, prop) {
				t.Errorf(".dashboard-header should have %s property", prop)
			}
		})
	}
}

func TestDashboardMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("Failed to read dashboard.css: %v", err)
	}

	cssContent := string(content)

	dashboardIndex := strings.Index(cssContent, ".dashboard {")
	if dashboardIndex == -1 {
		t.Fatal(".dashboard base styles not found")
	}

	mediaQueryIndex := strings.Index(cssContent, "@media (min-width: 768px)")
	if mediaQueryIndex == -1 {
		t.Fatal("Tablet media query not found")
	}

	if dashboardIndex > mediaQueryIndex {
		t.Error("Base styles should come before media queries (mobile-first approach)")
	}
}
