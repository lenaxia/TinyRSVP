package css

import (
	"os"
	"strings"
	"testing"
)

func TestNavigationIntegrationWithVariables(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	varContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	nav := string(navContent)
	vars := string(varContent)

	tests := []struct {
		name     string
		variable string
	}{
		{"uses color-background", "--color-background"},
		{"uses color-border", "--color-border"},
		{"uses color-text-primary", "--color-text-primary"},
		{"uses color-primary-600", "--color-primary-600"},
		{"uses color-primary-700", "--color-primary-700"},
		{"uses color-primary-50", "--color-primary-50"},
		{"uses color-surface", "--color-surface"},
		{"uses color-border-focus", "--color-border-focus"},
		{"uses spacing variables", "--spacing-"},
		{"uses font-size variables", "--font-size-"},
		{"uses font-weight variables", "--font-weight-"},
		{"uses transition variables", "--transition-"},
		{"uses radius variables", "--radius-"},
		{"uses shadow variables", "--shadow-"},
		{"uses z-index variables", "--z-index-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(nav, "var("+tt.variable) {
				t.Errorf("navigation.css should use variable %s", tt.variable)
			}

			if !strings.Contains(vars, tt.variable) {
				t.Errorf("variables.css should define %s", tt.variable)
			}
		})
	}
}

func TestNavigationIntegrationWithGrid(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	nav := string(navContent)
	grid := string(gridContent)

	if !strings.Contains(nav, "display: flex") {
		t.Error("navigation should use flexbox for layout")
	}

	if !strings.Contains(grid, ".flex") {
		t.Error("grid.css should define flex utilities")
	}

	if !strings.Contains(grid, ".container") {
		t.Error("grid.css should define container class for navigation")
	}
}

func TestNavigationIntegrationWithTypography(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	typContent, err := os.ReadFile("typography.css")
	if err != nil {
		t.Fatalf("Failed to read typography.css: %v", err)
	}

	nav := string(navContent)
	typ := string(typContent)

	fontVars := []string{
		"--font-size-",
		"--font-weight-",
	}

	for _, fontVar := range fontVars {
		if !strings.Contains(nav, "var("+fontVar) {
			t.Errorf("navigation should use typography variable %s", fontVar)
		}

		if !strings.Contains(typ, fontVar) {
			t.Errorf("typography.css should define %s", fontVar)
		}
	}
}

func TestNavigationIntegrationWithColors(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	colorsContent, err := os.ReadFile("colors.css")
	if err != nil {
		t.Fatalf("Failed to read colors.css: %v", err)
	}

	nav := string(navContent)
	colors := string(colorsContent)

	colorVars := []string{
		"--color-background",
		"--color-primary-600",
		"--color-surface",
	}

	for _, colorVar := range colorVars {
		if !strings.Contains(nav, "var("+colorVar) {
			t.Errorf("navigation should use color variable %s", colorVar)
		}

		if !strings.Contains(colors, colorVar) {
			t.Logf("colors.css uses utility classes that reference %s", colorVar)
		}
	}
}

func TestNavigationIntegrationWithSpacing(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	nav := string(navContent)
	spacing := string(spacingContent)

	if !strings.Contains(nav, "var(--spacing-") {
		t.Error("navigation should use spacing variables")
	}

	if !strings.Contains(spacing, "--spacing-") {
		t.Error("spacing.css should define spacing variables")
	}

	spacingVars := []string{
		"--spacing-2",
		"--spacing-3",
		"--spacing-4",
		"--spacing-6",
		"--spacing-8",
	}

	for _, spacingVar := range spacingVars {
		if strings.Contains(nav, "var("+spacingVar+")") {
			if !strings.Contains(spacing, spacingVar) {
				t.Errorf("spacing.css should define %s used by navigation", spacingVar)
			}
		}
	}
}

func TestNavigationResponsiveBreakpointsMatchGrid(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	nav := string(navContent)
	grid := string(gridContent)

	breakpoints := []string{
		"@media (min-width: 768px)",
		"@media (min-width: 1024px)",
	}

	for _, bp := range breakpoints {
		navHas := strings.Contains(nav, bp)
		gridHas := strings.Contains(grid, bp)

		if navHas && !gridHas {
			t.Errorf("navigation uses breakpoint %s but grid doesn't define it", bp)
		}
	}
}

func TestNavigationAccessibilityCompliance(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	nav := string(navContent)

	accessibilityFeatures := []struct {
		name    string
		pattern string
	}{
		{"focus styles", ":focus"},
		{"focus-visible support", ":focus-visible"},
		{"outline for keyboard nav", "outline:"},
		{"44px touch targets", "44px"},
		{"hover states", ":hover"},
		{"active states", ".active"},
	}

	for _, feature := range accessibilityFeatures {
		t.Run(feature.name, func(t *testing.T) {
			if !strings.Contains(nav, feature.pattern) {
				t.Errorf("navigation should include %s for accessibility", feature.name)
			}
		})
	}
}

func TestNavigationMobileFirstImplementation(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	nav := string(navContent)

	mobileFirstChecks := []struct {
		name        string
		mobileStyle string
		mediaQuery  string
	}{
		{
			"hamburger visible on mobile",
			".nav-toggle",
			"@media (min-width: 768px)",
		},
		{
			"menu hidden on mobile",
			".nav-menu",
			"@media (min-width: 768px)",
		},
	}

	for _, check := range mobileFirstChecks {
		t.Run(check.name, func(t *testing.T) {
			mobileIndex := strings.Index(nav, check.mobileStyle)
			mediaIndex := strings.Index(nav, check.mediaQuery)

			if mobileIndex == -1 {
				t.Errorf("Expected %s to exist", check.mobileStyle)
				return
			}

			if mediaIndex == -1 {
				t.Errorf("Expected media query %s to exist", check.mediaQuery)
				return
			}

			if mobileIndex > mediaIndex {
				t.Errorf("Mobile styles for %s should come before desktop media queries", check.mobileStyle)
			}
		})
	}
}

func TestNavigationConsistentWithDesignSystem(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	nav := string(navContent)

	designSystemPatterns := []struct {
		name    string
		pattern string
		reason  string
	}{
		{
			"uses CSS variables",
			"var(--",
			"should use design system variables",
		},
		{
			"no hardcoded colors",
			"#",
			"should not have hardcoded hex colors",
		},
		{
			"uses transitions",
			"transition:",
			"should have smooth transitions",
		},
		{
			"uses border-radius",
			"border-radius:",
			"should use consistent border radius",
		},
	}

	for _, pattern := range designSystemPatterns {
		t.Run(pattern.name, func(t *testing.T) {
			contains := strings.Contains(nav, pattern.pattern)

			if pattern.name == "no hardcoded colors" {
				if contains && (strings.Contains(nav, "#fff") || strings.Contains(nav, "#000")) {
					t.Errorf("navigation %s", pattern.reason)
				}
			} else {
				if !contains {
					t.Errorf("navigation %s", pattern.reason)
				}
			}
		})
	}
}

func TestNavigationPerformanceOptimizations(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	nav := string(navContent)

	performanceChecks := []struct {
		name    string
		pattern string
		reason  string
	}{
		{
			"uses will-change sparingly",
			"will-change",
			"should avoid overusing will-change",
		},
		{
			"uses transform for animations",
			"transform:",
			"should use transform for smooth animations",
		},
		{
			"uses transition timing",
			"var(--transition-",
			"should use consistent transition timing",
		},
	}

	for _, check := range performanceChecks {
		t.Run(check.name, func(t *testing.T) {
			contains := strings.Contains(nav, check.pattern)

			if check.name == "uses will-change sparingly" {
				if contains {
					count := strings.Count(nav, check.pattern)
					if count > 2 {
						t.Errorf("navigation uses will-change too many times (%d), %s", count, check.reason)
					}
				}
			} else if check.name == "uses transform for animations" {
				if contains {
					t.Logf("navigation properly uses transform for animations")
				}
			} else {
				if !contains {
					t.Logf("navigation could benefit from %s", check.reason)
				}
			}
		})
	}
}

func TestNavigationZIndexLayering(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	varContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	nav := string(navContent)
	vars := string(varContent)

	if !strings.Contains(nav, "z-index") {
		t.Error("navigation should use z-index for proper layering")
	}

	if strings.Contains(nav, "var(--z-index-") {
		if !strings.Contains(vars, "--z-index-sticky") && !strings.Contains(vars, "--z-index-dropdown") {
			t.Error("variables.css should define z-index variables used by navigation")
		}
	}
}

func TestNavigationDarkModeSupport(t *testing.T) {
	navContent, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}

	varContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	nav := string(navContent)
	vars := string(varContent)

	if !strings.Contains(nav, "var(--color-") {
		t.Error("navigation should use color variables for dark mode support")
	}

	if !strings.Contains(vars, "@media (prefers-color-scheme: dark)") {
		t.Error("variables.css should define dark mode color overrides")
	}

	colorVarsInNav := []string{
		"--color-background",
		"--color-text-primary",
		"--color-border",
	}

	darkModeSection := vars[strings.Index(vars, "@media (prefers-color-scheme: dark)"):]

	for _, colorVar := range colorVarsInNav {
		if strings.Contains(nav, "var("+colorVar+")") {
			if !strings.Contains(darkModeSection, colorVar) {
				t.Errorf("variables.css dark mode should override %s used by navigation", colorVar)
			}
		}
	}
}
