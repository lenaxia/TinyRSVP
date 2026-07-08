package ux_playwright

import (
	"strings"
	"testing"

	pw "github.com/mxschmitt/playwright-go"
)

// TestPW_AdminDashboard_PageLoads verifies that /admin renders successfully
// when authenticated as admin and produces the expected structural elements
// (page-header, at-a-glance metric strip, quick actions).
func TestPW_AdminDashboard_PageLoads(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin")

	AssertContainsText(t, page, "Admin Dashboard")
	AssertContainsText(t, page, "Quick Actions")
}

// TestPW_AdminDashboard_ShowsBusinessStats asserts that the four at-a-glance
// metric tiles (Users, Events, Invites, System Health) are all present.
// This is the ops-overview strip added in the admin dashboard redesign.
func TestPW_AdminDashboard_ShowsBusinessStats(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin")

	labels := []string{"Total Users", "Total Events", "Total Invites", "System Health"}
	for _, label := range labels {
		body, err := page.Locator("body").TextContent()
		if err != nil {
			if isTargetClosedErr(err) {
				t.Skipf("renderer closed: %v", err)
			}
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(body, label) {
			t.Errorf("admin dashboard should show metric tile %q", label)
		}
	}
}

// TestPW_AdminDashboard_MetricTilesLinkToDrilldowns confirms that metric
// tiles for Users and Events are drilldown links, not dead boxes. This is
// the key behavioral difference from the pre-redesign dashboard.
func TestPW_AdminDashboard_MetricTilesLinkToDrilldowns(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin")

	// Expect an <a> with class "metric-tile" pointing at /admin/users.
	usersTile := page.Locator(`a.metric-tile[href="/admin/users"]`)
	count, err := usersTile.Count()
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("renderer closed: %v", err)
		}
		t.Fatalf("count usersTile: %v", err)
	}
	if count == 0 {
		t.Error("expected a Users drilldown link (a.metric-tile[href=/admin/users])")
	}

	eventsTile := page.Locator(`a.metric-tile[href="/events"]`)
	count, err = eventsTile.Count()
	if err != nil {
		t.Fatalf("count eventsTile: %v", err)
	}
	if count == 0 {
		t.Error("expected an Events drilldown link (a.metric-tile[href=/events])")
	}
}

// TestPW_AdminDashboard_QuickActionsAreStyledLinks confirms the "Quick
// Actions" section renders styled action-card links. Pre-redesign, the
// .action-card class had no CSS at all — this asserts the fix.
func TestPW_AdminDashboard_QuickActionsAreStyledLinks(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin")

	actionCards := page.Locator("a.action-card")
	count, err := actionCards.Count()
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("renderer closed: %v", err)
		}
		t.Fatalf("count action-cards: %v", err)
	}
	if count < 3 {
		t.Errorf("expected at least 3 action-cards in Quick Actions, got %d", count)
	}

	// At least one of them should have a computed border-radius (proving
	// components.css is applied). Pre-redesign this was 0px because
	// .action-card had no rules.
	radius, err := actionCards.First().Evaluate(
		"el => getComputedStyle(el).borderRadius", nil,
	)
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("renderer closed: %v", err)
		}
		t.Fatalf("get border-radius: %v", err)
	}
	if s, ok := radius.(string); !ok || s == "" || s == "0px" {
		t.Errorf("expected action-card to have non-zero border-radius (components.css should be loaded); got %v", radius)
	}
}

// TestPW_AdminDashboard_UnauthenticatedRedirect confirms the auth guard.
func TestPW_AdminDashboard_UnauthenticatedRedirect(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page, _ := AnonymousPage(t, ctx, srv, "/admin")

	currentURL := page.URL()
	if !strings.Contains(currentURL, "/login") {
		t.Errorf("Expected /admin to redirect to /login when unauthenticated, got: %s", currentURL)
	}
}

// TestPW_AdminSettings_UsesNewPartials confirms the migrated settings page
// renders using .ui-section-title (from the section partial) and
// .definition-list (replacing the ad-hoc .settings-grid).
func TestPW_AdminSettings_UsesNewPartials(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin/settings")

	assertLocatorHasCount := func(sel string, want int) {
		loc := page.Locator(sel)
		count, err := loc.Count()
		if err != nil {
			if isTargetClosedErr(err) {
				t.Skipf("renderer closed: %v", err)
			}
			t.Fatalf("count %s: %v", sel, err)
		}
		if count < want {
			t.Errorf("expected at least %d matches for %s, got %d", want, sel, count)
		}
	}

	assertLocatorHasCount(".ui-section-title", 3) // Server, Database, Auth, Email, Storage, Security, Token
	assertLocatorHasCount(".definition-list", 3)
}

// TestPW_AdminMetrics_UsesNewPartials confirms the migrated metrics page
// uses metric-tile-grid + definition-list.
func TestPW_AdminMetrics_UsesNewPartials(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin/metrics")

	metricGrid := page.Locator(".metric-tile-grid")
	count, err := metricGrid.Count()
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("renderer closed: %v", err)
		}
		t.Fatalf("count metric-tile-grid: %v", err)
	}
	if count == 0 {
		t.Error("expected metric-tile-grid on /admin/metrics")
	}

	// Body should reference at least one business metric label.
	AssertContainsText(t, page, "Total Users")
}

// TestPW_ComponentsCSS_Loaded asserts that the shared components.css is
// actually served (returns 200 with body containing a known rule).
func TestPW_ComponentsCSS_Loaded(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page, resp := AnonymousPage(t, ctx, srv, "/static/css/components.css")
	if resp != nil && resp.Status() != 200 {
		t.Errorf("expected 200 for components.css, got %d", resp.Status())
	}

	body, err := page.Locator("body").TextContent()
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("renderer closed: %v", err)
		}
		t.Fatalf("read body: %v", err)
	}
	// pw serves the raw CSS as text — look for a signature rule.
	if !strings.Contains(body, ".action-card") {
		t.Errorf("components.css missing .action-card selector; got first 500 chars:\n%s", firstN(body, 500))
	}
	if !strings.Contains(body, ".metric-tile") {
		t.Error("components.css missing .metric-tile selector")
	}
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Sanity: playwright import must be present so unused-import isn't a bug.
var _ = pw.String
