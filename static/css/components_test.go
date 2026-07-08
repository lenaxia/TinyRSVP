package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// components.css provides the CSS backing for reusable template partials
// defined in templates/web/partials/components.html and page_header.html.
// Every class here is expected to be used by at least one partial; conversely
// every partial that renders a class name (except one-off page-specific
// classes) should be backed by a rule in components.css.
//
// Design tokens only — hardcoded colors and pixel values must not appear.

func TestComponentsFileExists(t *testing.T) {
	if _, err := os.Stat("components.css"); os.IsNotExist(err) {
		t.Fatal("components.css file does not exist")
	}
}

func TestComponentsValidCSS(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	if strings.TrimSpace(css) == "" {
		t.Fatal("components.css is empty")
	}

	open := strings.Count(css, "{")
	close := strings.Count(css, "}")
	if open != close {
		t.Errorf("Mismatched braces: %d open, %d close", open, close)
	}
}

func TestComponentsRequiredClasses(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	required := []string{
		".ui-section",
		".ui-section-header",
		".ui-section-title",
		".ui-section-description",
		".ui-section-body",
		".action-grid",
		".action-card",
		".action-card-icon",
		".action-card-title",
		".action-card-description",
		".status-badge",
		".status-badge-success",
		".status-badge-warning",
		".status-badge-error",
		".status-badge-info",
		".status-badge-neutral",
		".definition-list",
		".metric-tile",
		".metric-tile-value",
		".metric-tile-label",
		".metric-tile-grid",
		".data-table",
		".panel",
		".panel-header",
		".panel-body",
		".panel-footer",
	}

	for _, cls := range required {
		t.Run(strings.TrimPrefix(cls, "."), func(t *testing.T) {
			if !strings.Contains(css, cls) {
				t.Errorf("Missing required class: %s", cls)
			}
		})
	}
}

func TestComponentsStatsCardAccentModifiers(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	// Accent modifiers layer semantic color onto .stats-card / .metric-tile so
	// the admin dashboard can visually differentiate categories (Users,
	// Events, Invites, Health) without cluttering the markup.
	accents := []string{
		".stats-card-accent-primary",
		".stats-card-accent-success",
		".stats-card-accent-warning",
		".stats-card-accent-error",
	}

	for _, cls := range accents {
		t.Run(strings.TrimPrefix(cls, "."), func(t *testing.T) {
			if !strings.Contains(css, cls) {
				t.Errorf("Missing accent modifier: %s", cls)
			}
		})
	}
}

func TestComponentsActionCardIsInteractive(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	// Action cards are anchors — they must have hover and focus states so the
	// admin dashboard "Quick Actions" section isn't a wall of dead-looking
	// boxes (which is what it looked like before this change).
	needsStates := []string{
		".action-card:hover",
		".action-card:focus",
	}

	for _, sel := range needsStates {
		t.Run(sel, func(t *testing.T) {
			if !strings.Contains(css, sel) {
				t.Errorf("Missing interactive state: %s", sel)
			}
		})
	}
}

func TestComponentsUsesDesignTokens(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	needed := []string{
		"var(--spacing-",
		"var(--color-",
		"var(--radius-",
		"var(--font-size-",
		"var(--font-weight-",
		"var(--transition-",
	}

	for _, v := range needed {
		t.Run("uses_"+v, func(t *testing.T) {
			if !strings.Contains(css, v) {
				t.Errorf("components.css should use %s", v)
			}
		})
	}
}

func TestComponentsNoHardcodedColors(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := stripCSSComments(string(content))

	hex := regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)
	if match := hex.FindString(css); match != "" {
		t.Errorf("hardcoded hex color %q — use CSS variables instead", match)
	}

	rgb := regexp.MustCompile(`rgba?\(`)
	if rgb.MatchString(css) {
		t.Error("hardcoded rgb/rgba — use CSS variables instead")
	}
}

func TestComponentsHasResponsiveBreakpoints(t *testing.T) {
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	if !strings.Contains(css, "@media") {
		t.Error("components.css should include at least one @media query for responsive behavior")
	}
	if !strings.Contains(css, "min-width: 768px") {
		t.Error("components.css should include tablet breakpoint (min-width: 768px)")
	}
}

func TestComponentsMetricTileGridIsFlexible(t *testing.T) {
	// .metric-tile-grid backs the admin dashboard "at a glance" strip and the
	// admin_metrics business-metrics section. It must be auto-fit so 3, 4, or
	// 5 tiles all lay out without orphans on desktop — a bug in the previous
	// dashboard.css .stats-grid which was hardcoded to 4 columns.
	content, err := os.ReadFile("components.css")
	if err != nil {
		t.Fatalf("read components.css: %v", err)
	}
	css := string(content)

	gridBlock := extractRuleBlock(css, ".metric-tile-grid")
	if gridBlock == "" {
		t.Fatal(".metric-tile-grid rule not found")
	}
	if !strings.Contains(gridBlock, "auto-fit") && !strings.Contains(gridBlock, "auto-fill") {
		t.Errorf(".metric-tile-grid should use auto-fit or auto-fill for orphan-free responsive layout; got:\n%s", gridBlock)
	}
	if !strings.Contains(gridBlock, "minmax(") {
		t.Errorf(".metric-tile-grid should use minmax() for a sensible min column width; got:\n%s", gridBlock)
	}
}

// stripCSSComments removes `/* ... */` blocks so that regex assertions about
// hardcoded colors don't false-positive on the documentation comments that
// mention hex / rgb literally.
func stripCSSComments(css string) string {
	commentRE := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return commentRE.ReplaceAllString(css, "")
}

// extractRuleBlock returns the body of the first CSS rule matching the given
// selector at the top level (not inside @media). Returns "" if not found.
func extractRuleBlock(css, selector string) string {
	// Find the selector followed by a "{" (allowing whitespace and
	// possibly other selectors joined by comma before the brace).
	idx := 0
	for {
		found := strings.Index(css[idx:], selector)
		if found == -1 {
			return ""
		}
		start := idx + found
		// Look forward for the opening brace on the same rule.
		brace := strings.Index(css[start:], "{")
		if brace == -1 {
			return ""
		}
		braceAt := start + brace

		// Ensure the selector between `start` and `braceAt` doesn't contain a
		// closing brace (which would mean we matched inside a different rule
		// or after one ended).
		if strings.Contains(css[start:braceAt], "}") {
			idx = start + len(selector)
			continue
		}

		// Walk forward to find the matching close brace.
		depth := 1
		i := braceAt + 1
		for i < len(css) && depth > 0 {
			switch css[i] {
			case '{':
				depth++
			case '}':
				depth--
			}
			i++
		}
		if depth != 0 {
			return ""
		}
		return css[braceAt+1 : i-1]
	}
}
