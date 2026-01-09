package css

import (
	"os"
	"strings"
	"testing"
)

func TestNavigationCSSExists(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("navigation.css should exist: %v", err)
	}
	if len(content) == 0 {
		t.Error("navigation.css should not be empty")
	}
}

func TestNavigationHeaderStyles(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	tests := []struct {
		name     string
		selector string
		wantErr  bool
	}{
		{"header class exists", ".header", false},
		{"nav class exists", ".nav", false},
		{"logo class exists", ".logo", false},
		{"nav-toggle class exists", ".nav-toggle", false},
		{"nav-menu class exists", ".nav-menu", false},
		{"nav-link class exists", ".nav-link", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(css, tt.selector) {
				t.Errorf("Expected selector %s to exist in navigation.css", tt.selector)
			}
		})
	}
}

func TestNavigationUsesVariables(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	requiredVars := []string{
		"var(--color-background)",
		"var(--color-border)",
		"var(--spacing-",
		"var(--color-text-primary)",
		"var(--transition-",
	}

	for _, varName := range requiredVars {
		if !strings.Contains(css, varName) {
			t.Errorf("Expected navigation.css to use CSS variable %s", varName)
		}
	}
}

func TestNavigationResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	breakpoints := []string{
		"@media (min-width: 768px)",
	}

	for _, bp := range breakpoints {
		if !strings.Contains(css, bp) {
			t.Errorf("Expected navigation.css to contain responsive breakpoint: %s", bp)
		}
	}
}

func TestNavigationMobileFirstApproach(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, ".nav-toggle") {
		t.Error("Expected mobile hamburger toggle to exist")
	}

	mediaQueryIndex := strings.Index(css, "@media (min-width: 768px)")
	navToggleIndex := strings.Index(css, ".nav-toggle")

	if mediaQueryIndex != -1 && navToggleIndex != -1 && mediaQueryIndex < navToggleIndex {
		t.Error("Mobile styles should come before desktop media queries (mobile-first)")
	}
}

func TestNavigationTouchTargets(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, "44px") && !strings.Contains(css, "2.75rem") {
		t.Error("Expected navigation to have 44px minimum touch targets for mobile accessibility")
	}
}

func TestNavigationActiveState(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	activeSelectors := []string{
		".nav-link.active",
		".nav-link:hover",
		".nav-link:focus",
	}

	for _, selector := range activeSelectors {
		if !strings.Contains(css, selector) {
			t.Errorf("Expected navigation.css to contain active/hover/focus state: %s", selector)
		}
	}
}

func TestNavigationKeyboardAccessibility(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	focusSelectors := []string{
		":focus",
		"outline",
	}

	foundFocus := false
	for _, selector := range focusSelectors {
		if strings.Contains(css, selector) {
			foundFocus = true
			break
		}
	}

	if !foundFocus {
		t.Error("Expected navigation.css to include focus styles for keyboard accessibility")
	}
}

func TestNavigationStickyHeader(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, ".header.sticky") && !strings.Contains(css, "position: sticky") {
		t.Error("Expected navigation.css to support sticky header positioning")
	}
}

func TestNavigationZIndex(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, "z-index") && !strings.Contains(css, "var(--z-index-") {
		t.Error("Expected navigation to use z-index for proper layering")
	}
}

func TestNavigationMenuToggleState(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, ".nav-menu.open") && !strings.Contains(css, ".nav-menu.active") {
		t.Error("Expected navigation to have open/active state for mobile menu")
	}
}

func TestNavigationHamburgerIcon(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	hamburgerSelectors := []string{
		".nav-toggle span",
		".nav-toggle::before",
		".nav-toggle::after",
	}

	foundHamburger := false
	for _, selector := range hamburgerSelectors {
		if strings.Contains(css, selector) {
			foundHamburger = true
			break
		}
	}

	if !foundHamburger {
		t.Error("Expected navigation to have hamburger icon styling")
	}
}

func TestNavigationDesktopHidesToggle(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	mediaQuery := "@media (min-width: 768px)"
	mediaIndex := strings.Index(css, mediaQuery)
	if mediaIndex == -1 {
		t.Fatal("Expected desktop media query to exist")
	}

	afterMedia := css[mediaIndex:]
	if !strings.Contains(afterMedia, ".nav-toggle") || !strings.Contains(afterMedia, "display: none") {
		t.Error("Expected nav-toggle to be hidden on desktop")
	}
}

func TestNavigationDesktopShowsMenu(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	mediaQuery := "@media (min-width: 768px)"
	mediaIndex := strings.Index(css, mediaQuery)
	if mediaIndex == -1 {
		t.Fatal("Expected desktop media query to exist")
	}

	afterMedia := css[mediaIndex:]
	if !strings.Contains(afterMedia, ".nav-menu") {
		t.Error("Expected nav-menu to have desktop styles")
	}
	if !strings.Contains(afterMedia, "display: flex") && !strings.Contains(afterMedia, "display:flex") {
		t.Error("Expected nav-menu to use flexbox on desktop")
	}
}

func TestNavigationTransitions(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, "transition") {
		t.Error("Expected navigation to use transitions for smooth interactions")
	}
}

func TestNavigationNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("navigation.css")
	if err != nil {
		t.Fatalf("Failed to read navigation.css: %v", err)
	}
	css := string(content)

	hardcodedColorPatterns := []string{
		"#fff",
		"#000",
		"rgb(",
		"rgba(",
	}

	for _, pattern := range hardcodedColorPatterns {
		if strings.Contains(css, pattern) {
			t.Errorf("Expected navigation.css to use CSS variables instead of hardcoded colors like %s", pattern)
		}
	}
}
