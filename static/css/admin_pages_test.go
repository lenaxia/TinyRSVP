package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The admin_metrics and admin_settings stylesheets shipped with hardcoded
// hex color fallbacks like #f5f5f5, #d4edda, #721c24 that break in dark mode
// because they never resolve to the theme's actual surface / semantic colors.
// The rest of the design system enforces "tokens only" via TestDashboardNoHardcodedColors
// and TestColorsNoHardcodedColors; this brings those two admin stylesheets
// under the same guardrail.

func TestAdminMetricsNoHardcodedColors(t *testing.T) {
	assertNoHardcodedColors(t, "admin_metrics.css")
}

func TestAdminSettingsNoHardcodedColors(t *testing.T) {
	assertNoHardcodedColors(t, "admin_settings.css")
}

func TestAdminMetricsUsesDesignTokens(t *testing.T) {
	assertUsesDesignTokens(t, "admin_metrics.css")
}

func TestAdminSettingsUsesDesignTokens(t *testing.T) {
	assertUsesDesignTokens(t, "admin_settings.css")
}

// The old .stats-grid stepped from 1 → 2 → 4 columns which produced an
// orphan card whenever a page had 3 stats (as the admin dashboard does).
// Auto-fit + minmax removes the orphan without needing to hand-pick
// per-page grid counts.
func TestDashboardStatsGridIsAutoFit(t *testing.T) {
	content, err := os.ReadFile("dashboard.css")
	if err != nil {
		t.Fatalf("read dashboard.css: %v", err)
	}
	css := string(content)

	block := extractRuleBlock(css, ".stats-grid")
	if block == "" {
		t.Fatal(".stats-grid rule not found")
	}
	if !strings.Contains(block, "auto-fit") && !strings.Contains(block, "auto-fill") {
		t.Errorf(".stats-grid should use auto-fit or auto-fill so pages with 3, 4 or 5 cards all lay out without orphans; got:\n%s", block)
	}
	if !strings.Contains(block, "minmax(") {
		t.Errorf(".stats-grid should use minmax() for a sensible min column width; got:\n%s", block)
	}
}

func assertNoHardcodedColors(t *testing.T, filename string) {
	t.Helper()
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	css := stripCSSComments(string(content))

	hex := regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)
	if match := hex.FindString(css); match != "" {
		t.Errorf("%s: hardcoded hex color %q — use CSS variables instead", filename, match)
	}

	rgb := regexp.MustCompile(`rgba?\(`)
	if rgb.MatchString(css) {
		t.Errorf("%s: hardcoded rgb/rgba — use CSS variables instead", filename)
	}
}

func assertUsesDesignTokens(t *testing.T, filename string) {
	t.Helper()
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	css := string(content)

	if !strings.Contains(css, "var(--color-") {
		t.Errorf("%s should reference --color-* design tokens", filename)
	}
	if !strings.Contains(css, "var(--spacing-") {
		t.Errorf("%s should reference --spacing-* design tokens", filename)
	}
}
