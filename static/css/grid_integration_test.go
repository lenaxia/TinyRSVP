package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestGridIntegrationWithVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	variables := string(variablesContent)
	grid := string(gridContent)

	tests := []struct {
		name     string
		variable string
	}{
		{"spacing-4 used in container", "--spacing-4"},
		{"container-md defined", "--container-md"},
		{"container-lg defined", "--container-lg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(variables, tt.variable) {
				t.Errorf("Variable %s not found in variables.css", tt.variable)
			}

			if !strings.Contains(grid, tt.variable) {
				t.Errorf("Variable %s not used in grid.css", tt.variable)
			}
		})
	}
}

func TestGridIntegrationWithSpacing(t *testing.T) {
	spacingContent, err := os.ReadFile("spacing.css")
	if err != nil {
		t.Fatalf("Failed to read spacing.css: %v", err)
	}

	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	spacing := string(spacingContent)
	grid := string(gridContent)

	if !strings.Contains(spacing, ".gap-") {
		t.Error("spacing.css should contain gap utilities")
	}

	if strings.Contains(grid, ".gap-") {
		t.Error("grid.css should not duplicate gap utilities from spacing.css")
	}
}

func TestGridBreakpointsMatchVariables(t *testing.T) {
	variablesContent, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	variables := string(variablesContent)
	grid := string(gridContent)

	tests := []struct {
		name       string
		breakpoint string
	}{
		{"md breakpoint", "768px"},
		{"lg breakpoint", "1024px"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varPattern := regexp.MustCompile(`--breakpoint-(?:md|lg):\s*` + regexp.QuoteMeta(tt.breakpoint))
			if !varPattern.MatchString(variables) {
				t.Errorf("Breakpoint %s not found in variables.css", tt.breakpoint)
			}

			gridPattern := regexp.MustCompile(`@media\s*\(min-width:\s*` + regexp.QuoteMeta(tt.breakpoint) + `\)`)
			if !gridPattern.MatchString(grid) {
				t.Errorf("Media query for %s not found in grid.css", tt.breakpoint)
			}
		})
	}
}

func TestGridResponsiveClassesComplete(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".md\\:grid-cols-") {
		t.Error("Missing md: responsive variants")
	}

	if !strings.Contains(cssContent, ".lg\\:grid-cols-") {
		t.Error("Missing lg: responsive variants")
	}
}

func TestGridFlexboxIntegration(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	flexClasses := []string{
		".flex",
		".flex-row",
		".flex-col",
		".items-center",
		".justify-between",
	}

	for _, class := range flexClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("Missing flexbox utility class: %s", class)
		}
	}
}

func TestGridContainerIntegration(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	if !strings.Contains(cssContent, ".container") {
		t.Fatal("Missing .container class")
	}

	containerPattern := regexp.MustCompile(`\.container\s*\{[^}]*width:\s*100%[^}]*margin-left:\s*auto[^}]*margin-right:\s*auto`)
	if !containerPattern.MatchString(cssContent) {
		t.Error(".container should be centered with auto margins")
	}
}

func TestGridSystemCompleteness(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	requiredFeatures := []struct {
		name    string
		pattern string
	}{
		{"Grid display", `\.grid\s*\{[^}]*display:\s*grid`},
		{"Grid columns", `\.grid-cols-\d+`},
		{"Responsive grid", `\.(?:md|lg)\\:grid-cols-\d+`},
		{"Flex display", `\.flex\s*\{[^}]*display:\s*flex`},
		{"Flex direction", `\.flex-(?:row|col)`},
		{"Align items", `\.items-`},
		{"Justify content", `\.justify-`},
		{"Container", `\.container`},
		{"Auto-fit", `\.grid-auto-fit`},
		{"Auto-fill", `\.grid-auto-fill`},
		{"Column span", `\.col-span-`},
	}

	for _, feature := range requiredFeatures {
		t.Run(feature.name, func(t *testing.T) {
			matched, err := regexp.MatchString(feature.pattern, cssContent)
			if err != nil {
				t.Fatalf("Invalid regex pattern: %v", err)
			}
			if !matched {
				t.Errorf("Missing feature: %s (pattern: %s)", feature.name, feature.pattern)
			}
		})
	}
}

func TestGridNoConflictsWithOtherCSS(t *testing.T) {
	gridContent, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	grid := string(gridContent)

	conflictingPatterns := []struct {
		name    string
		pattern string
		reason  string
	}{
		{"No color definitions", `color:\s*#[0-9a-fA-F]{3,6}`, "Colors should be in colors.css"},
		{"No font definitions", `font-family:`, "Fonts should be in typography.css"},
		{"No margin utilities", `\.m-\d+\s*\{`, "Margin utilities should be in spacing.css"},
		{"No padding utilities", `\.p-\d+\s*\{`, "Padding utilities should be in spacing.css"},
	}

	for _, conflict := range conflictingPatterns {
		t.Run(conflict.name, func(t *testing.T) {
			matched, err := regexp.MatchString(conflict.pattern, grid)
			if err != nil {
				t.Fatalf("Invalid regex pattern: %v", err)
			}
			if matched {
				t.Errorf("Found conflicting pattern: %s - %s", conflict.name, conflict.reason)
			}
		})
	}
}

func TestGridMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	baseGridPattern := regexp.MustCompile(`\.grid-cols-\d+\s*\{`)
	baseMatches := baseGridPattern.FindAllString(cssContent, -1)
	if len(baseMatches) == 0 {
		t.Error("Missing base (mobile) grid column classes")
	}

	mediaQueryPattern := regexp.MustCompile(`@media\s*\([^)]+\)`)
	mediaMatches := mediaQueryPattern.FindAllString(cssContent, -1)

	for _, media := range mediaMatches {
		if !strings.Contains(media, "min-width") {
			t.Errorf("Media query should use min-width for mobile-first: %s", media)
		}
	}
}

func TestGridResponsiveVariantsConsistency(t *testing.T) {
	content, err := os.ReadFile("grid.css")
	if err != nil {
		t.Fatalf("Failed to read grid.css: %v", err)
	}

	cssContent := string(content)

	baseClasses := []string{"grid-cols-2", "grid-cols-3", "grid-cols-4"}
	breakpoints := []string{"md", "lg"}

	for _, baseClass := range baseClasses {
		if !strings.Contains(cssContent, "."+baseClass) {
			t.Errorf("Missing base class: .%s", baseClass)
			continue
		}

		for _, bp := range breakpoints {
			responsiveClass := "." + bp + `\:` + baseClass
			if !strings.Contains(cssContent, responsiveClass) {
				t.Errorf("Missing responsive variant: .%s:%s", bp, baseClass)
			}
		}
	}
}
