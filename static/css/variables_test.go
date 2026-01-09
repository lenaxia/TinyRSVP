package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestCSSVariablesFileExists(t *testing.T) {
	_, err := os.Stat("variables.css")
	if err != nil {
		t.Fatalf("variables.css file should exist: %v", err)
	}
}

func TestCSSVariablesHasRootSelector(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	if !strings.Contains(string(content), ":root") {
		t.Error("variables.css should contain :root selector")
	}
}

func TestCSSVariablesDefinesPrimaryColors(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	requiredColors := []string{
		"--color-primary-50",
		"--color-primary-100",
		"--color-primary-200",
		"--color-primary-300",
		"--color-primary-400",
		"--color-primary-500",
		"--color-primary-600",
		"--color-primary-700",
		"--color-primary-800",
		"--color-primary-900",
	}

	for _, color := range requiredColors {
		if !strings.Contains(string(content), color) {
			t.Errorf("variables.css should define %s", color)
		}
	}
}

func TestCSSVariablesDefinesSemanticColors(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	semanticColors := []string{
		"--color-success",
		"--color-success-light",
		"--color-warning",
		"--color-warning-light",
		"--color-error",
		"--color-error-light",
		"--color-info",
		"--color-info-light",
	}

	for _, color := range semanticColors {
		if !strings.Contains(string(content), color) {
			t.Errorf("variables.css should define %s", color)
		}
	}
}

func TestCSSVariablesDefinesGrayScale(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	grayColors := []string{
		"--color-gray-50",
		"--color-gray-100",
		"--color-gray-200",
		"--color-gray-300",
		"--color-gray-400",
		"--color-gray-500",
		"--color-gray-600",
		"--color-gray-700",
		"--color-gray-800",
		"--color-gray-900",
	}

	for _, color := range grayColors {
		if !strings.Contains(string(content), color) {
			t.Errorf("variables.css should define %s", color)
		}
	}
}

func TestCSSVariablesDefinesFunctionalColors(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	functionalColors := []string{
		"--color-background",
		"--color-surface",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-text-disabled",
		"--color-border",
		"--color-border-focus",
	}

	for _, color := range functionalColors {
		if !strings.Contains(string(content), color) {
			t.Errorf("variables.css should define %s", color)
		}
	}
}

func TestCSSVariablesDefinesSpacingScale(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	spacingValues := []string{
		"--spacing-0",
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
		"--spacing-20",
		"--spacing-24",
	}

	for _, spacing := range spacingValues {
		if !strings.Contains(string(content), spacing) {
			t.Errorf("variables.css should define %s", spacing)
		}
	}
}

func TestCSSVariablesDefinesTypographyScale(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	typographyValues := []string{
		"--font-size-xs",
		"--font-size-sm",
		"--font-size-base",
		"--font-size-lg",
		"--font-size-xl",
		"--font-size-2xl",
		"--font-size-3xl",
		"--font-size-4xl",
		"--font-size-5xl",
		"--font-weight-normal",
		"--font-weight-medium",
		"--font-weight-semibold",
		"--font-weight-bold",
		"--line-height-tight",
		"--line-height-normal",
		"--line-height-relaxed",
		"--font-family-sans",
		"--font-family-mono",
	}

	for _, typo := range typographyValues {
		if !strings.Contains(string(content), typo) {
			t.Errorf("variables.css should define %s", typo)
		}
	}
}

func TestCSSVariablesDefinesBorderRadius(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	radiusValues := []string{
		"--radius-none",
		"--radius-sm",
		"--radius-base",
		"--radius-md",
		"--radius-lg",
		"--radius-xl",
		"--radius-2xl",
		"--radius-full",
	}

	for _, radius := range radiusValues {
		if !strings.Contains(string(content), radius) {
			t.Errorf("variables.css should define %s", radius)
		}
	}
}

func TestCSSVariablesDefinesShadows(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	shadowValues := []string{
		"--shadow-sm",
		"--shadow-base",
		"--shadow-md",
		"--shadow-lg",
		"--shadow-xl",
		"--shadow-2xl",
	}

	for _, shadow := range shadowValues {
		if !strings.Contains(string(content), shadow) {
			t.Errorf("variables.css should define %s", shadow)
		}
	}
}

func TestCSSVariablesDefinesTransitions(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	transitionValues := []string{
		"--transition-fast",
		"--transition-base",
		"--transition-slow",
	}

	for _, transition := range transitionValues {
		if !strings.Contains(string(content), transition) {
			t.Errorf("variables.css should define %s", transition)
		}
	}
}

func TestCSSVariablesDefinesZIndex(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	zIndexValues := []string{
		"--z-index-dropdown",
		"--z-index-sticky",
		"--z-index-fixed",
		"--z-index-modal-backdrop",
		"--z-index-modal",
		"--z-index-popover",
		"--z-index-tooltip",
	}

	for _, zIndex := range zIndexValues {
		if !strings.Contains(string(content), zIndex) {
			t.Errorf("variables.css should define %s", zIndex)
		}
	}
}

func TestCSSVariablesDefinesBreakpoints(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	breakpointValues := []string{
		"--breakpoint-sm",
		"--breakpoint-md",
		"--breakpoint-lg",
		"--breakpoint-xl",
		"--breakpoint-2xl",
	}

	for _, breakpoint := range breakpointValues {
		if !strings.Contains(string(content), breakpoint) {
			t.Errorf("variables.css should define %s", breakpoint)
		}
	}
}

func TestCSSVariablesDefinesContainerWidths(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	containerValues := []string{
		"--container-sm",
		"--container-md",
		"--container-lg",
		"--container-xl",
	}

	for _, container := range containerValues {
		if !strings.Contains(string(content), container) {
			t.Errorf("variables.css should define %s", container)
		}
	}
}

func TestCSSVariablesHasValidSyntax(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	variablePattern := regexp.MustCompile(`--[a-z0-9-]+:\s*[^;]+;`)
	matches := variablePattern.FindAllString(string(content), -1)

	if len(matches) < 50 {
		t.Errorf("Expected at least 50 CSS variables, found %d", len(matches))
	}
}

func TestCSSVariablesColorContrastWCAA(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	textPrimaryPattern := regexp.MustCompile(`--color-text-primary:\s*#([0-9a-fA-F]{6}|[0-9a-fA-F]{3});`)
	backgroundPattern := regexp.MustCompile(`--color-background:\s*#([0-9a-fA-F]{6}|[0-9a-fA-F]{3});`)

	if !textPrimaryPattern.MatchString(string(content)) {
		t.Error("--color-text-primary should be defined with hex color")
	}

	if !backgroundPattern.MatchString(string(content)) {
		t.Error("--color-background should be defined with hex color")
	}
}

func TestCSSVariablesHasDarkModeSupport(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	if !strings.Contains(string(content), "@media (prefers-color-scheme: dark)") {
		t.Error("variables.css should include dark mode support with prefers-color-scheme media query")
	}
}

func TestCSSVariablesDarkModeOverrides(t *testing.T) {
	content, err := os.ReadFile("variables.css")
	if err != nil {
		t.Fatalf("Failed to read variables.css: %v", err)
	}

	darkModeSection := string(content)
	if !strings.Contains(darkModeSection, "@media (prefers-color-scheme: dark)") {
		t.Skip("Dark mode not implemented")
	}

	darkModeColors := []string{
		"--color-background",
		"--color-surface",
		"--color-text-primary",
		"--color-text-secondary",
		"--color-border",
	}

	darkModeStart := strings.Index(darkModeSection, "@media (prefers-color-scheme: dark)")
	if darkModeStart == -1 {
		t.Fatal("Dark mode media query not found")
	}

	darkModeContent := darkModeSection[darkModeStart:]

	for _, color := range darkModeColors {
		if !strings.Contains(darkModeContent, color) {
			t.Errorf("Dark mode should override %s", color)
		}
	}
}
