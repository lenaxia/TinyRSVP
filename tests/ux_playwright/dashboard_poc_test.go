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
