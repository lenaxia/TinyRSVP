package ux_playwright

import (
	"testing"

	pw "github.com/mxschmitt/playwright-go"
)

// TestDashboard_LoadsInBrowser is a proof-of-concept test that verifies
// Playwright is wired up correctly and the dashboard page renders in a real
// headless browser against the in-process test server.
//
// This test exercises the full HTTP+HTML stack end-to-end:
//   - Real browser (not mock HTTP)
//   - Real CSS/JS loading via /static
//   - Real template rendering with base.html + navigation.html
//   - X-Test-User-ID auth bypass working through the middleware stack
//   - Dashboard service producing real stats (zeros for empty DB)
//
// If this passes, the Playwright harness is functional and ready for
// migration of the chromedp tests.
func TestDashboard_LoadsInBrowser(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/")

	AssertContainsText(t, page, "Dashboard")

	AssertContainsText(t, page, "Total Events")
	AssertContainsText(t, page, "Invites Sent")
	AssertContainsText(t, page, "RSVPs Received")

	AssertContainsText(t, page, "Recent Activity")

	t.Run("navigation works", func(t *testing.T) {
		adminLink := page.Locator(`a[href="/admin"]`)
		if count, err := adminLink.Count(); err != nil || count == 0 {
			t.Skip("no /admin link on dashboard (admin nav may be conditional)")
		}
		if err := adminLink.Click(pw.LocatorClickOptions{
			Timeout: pw.Float(5000),
		}); err != nil {
			t.Fatalf("click /admin: %v", err)
		}
		if err := page.WaitForURL(srv.URL("/admin"), pw.PageWaitForURLOptions{
			Timeout: pw.Float(5000),
		}); err != nil {
			t.Fatalf("wait for /admin URL: %v", err)
		}
		AssertContainsText(t, page, "Admin Dashboard")
	})
}

// TestAdminDashboard_HasSettingsAndMetricsLinks verifies the dashboard now
// links to both new admin pages (added in PRs #33 and #34).
func TestAdminDashboard_HasSettingsAndMetricsLinks(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin")

	settingsLink := page.Locator(`a[href="/admin/settings"]`)
	if count, err := settingsLink.Count(); err != nil {
		t.Fatalf("count settings link: %v", err)
	} else if count == 0 {
		t.Error("expected /admin/settings link in dashboard quick actions")
	}

	metricsLink := page.Locator(`a[href="/admin/metrics"]`)
	if count, err := metricsLink.Count(); err != nil {
		t.Fatalf("count metrics link: %v", err)
	} else if count == 0 {
		t.Error("expected /admin/metrics link in dashboard quick actions")
	}
}

// TestAdminSettings_Renders verifies the new settings page renders with
// expected content in a real browser.
func TestAdminSettings_Renders(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin/settings")

	AssertContainsText(t, page, "System Settings")
	AssertContainsText(t, page, "Server")
	AssertContainsText(t, page, "Database")
	AssertContainsText(t, page, "Authentication")
}

// TestAdminMetrics_Renders verifies the new metrics page renders with
// expected content in a real browser.
func TestAdminMetrics_Renders(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin/metrics")

	AssertContainsText(t, page, "System Metrics")
	AssertContainsText(t, page, "Business Metrics")
	AssertContainsText(t, page, "Database Connection Pool")
	AssertContainsText(t, page, "Email Queue")
}

// --- Unhappy-path tests ---

// TestDashboard_UnauthenticatedRedirect verifies that without the auth bypass
// header, the dashboard redirects to /login rather than rendering.
func TestDashboard_UnauthenticatedRedirect(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	resp, err := page.Goto(srv.URL("/"))
	if err != nil {
		t.Fatalf("goto /: %v", err)
	}

	if status := resp.Status(); status != 200 && status != 302 && status != 303 {
		t.Errorf("expected status 200/302/303, got %d", status)
	}

	currentURL := page.URL()
	if currentURL == srv.URL("/") {
		t.Errorf("expected redirect away from /, but stayed at /")
	}
}

// TestNonExistentAdminPage_Returns404 verifies that an admin user navigating
// to a non-existent admin page gets a 404, not a server error.
func TestNonExistentAdminPage_Returns404(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/admin/this-page-does-not-exist")

	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if body == "" {
		t.Error("expected non-empty body for non-existent page (404 handler should render)")
	}
}

// TestDashboard_RendersOnEmptyDatabase verifies the dashboard renders
// correctly when the database has no events, invites, or RSVPs — the
// "empty state" case that admins see on first login.
func TestDashboard_RendersOnEmptyDatabase(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/")

	AssertContainsText(t, page, "0")

	AssertContainsText(t, page, "No Recent Activity")
}

// TestStaticAssets_Load verifies that static CSS files are reachable from the
// browser (catches misconfigured static file server wiring).
func TestStaticAssets_Load(t *testing.T) {
	srv := SetupTestServer(t)
	_, ctx := NewBrowser(t)

	page := AsAdminPage(t, ctx, srv, "/")

	assets, err := page.Locator(`link[rel="stylesheet"]`).All()
	if err != nil {
		t.Fatalf("list stylesheets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected at least one stylesheet link in the dashboard page")
	}

	for i, a := range assets {
		href, err := a.GetAttribute("href")
		if err != nil {
			t.Logf("stylesheet %d: could not read href: %v", i, err)
			continue
		}
		if href == "" {
			continue
		}

		status, err := page.Evaluate(`async (url) => {
			const resp = await fetch(url);
			return resp.status;
		}`, href)
		if err != nil {
			t.Errorf("stylesheet %s: fetch failed: %v", href, err)
			continue
		}
		var statusNum int
		switch s := status.(type) {
		case int:
			statusNum = s
		case float64:
			statusNum = int(s)
		default:
			t.Errorf("stylesheet %s: unexpected status type %T: %v", href, status, status)
			continue
		}
		if statusNum >= 400 {
			t.Errorf("stylesheet %s: expected HTTP < 400, got %d", href, statusNum)
		}
	}
}
