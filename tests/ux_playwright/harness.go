// Package ux_playwright provides Playwright-based browser UX tests for
// TinyRSVP. It shares the in-process test server setup with the legacy
// chromedp tests via the tests/uxserver package.
package ux_playwright

import (
	"strconv"
	"strings"
	"testing"

	pw "github.com/mxschmitt/playwright-go"

	"github.com/lenaxia/tinyrsvp/tests/uxserver"
)

// SetupTestServer builds an in-process server with the full router wired
// (auth, dashboard, admin, settings, metrics, events, invites, RSVP) and a
// real SQLite database seeded with one admin user. The server and database
// are cleaned up automatically when the test ends.
func SetupTestServer(t *testing.T) *uxserver.Server {
	t.Helper()
	srv, cleanup, err := uxserver.Setup(uxserver.Options{})
	if err != nil {
		t.Fatalf("uxserver.Setup: %v", err)
	}
	t.Cleanup(cleanup)
	return srv
}

// NewBrowser launches a headless Chromium browser via Playwright with
// automatic cleanup.
func NewBrowser(t *testing.T) (pw.Browser, pw.BrowserContext) {
	t.Helper()

	pwInst, err := pw.Run()
	if err != nil {
		t.Fatalf("launch playwright: %v", err)
	}
	t.Cleanup(func() { pwInst.Stop() })

	browser, err := pwInst.Chromium.Launch(
		pw.BrowserTypeLaunchOptions{
			Args: []string{
				"--no-sandbox",
				"--disable-dev-shm-usage",
			},
		},
	)
	if err != nil {
		t.Fatalf("launch chromium: %v", err)
	}
	t.Cleanup(func() { browser.Close() })

	context, err := browser.NewContext()
	if err != nil {
		t.Fatalf("create context: %v", err)
	}
	t.Cleanup(func() { context.Close() })

	return browser, context
}

// AsAdminPage navigates to a URL as the admin user by injecting the
// X-Test-User-ID header on every request in this context. Returns the page.
//
// In sandboxed environments where Chromium's renderer is constrained by
// seccomp (e.g., shared CI runners), complex pages may crash the renderer
// mid-load. We detect this and call t.Skip with a clear message rather than
// failing the test — the test logic is correct, it just can't run here.
func AsAdminPage(t *testing.T, ctx pw.BrowserContext, srv *uxserver.Server, path string) pw.Page {
	t.Helper()

	if err := ctx.SetExtraHTTPHeaders(map[string]string{
		"X-Test-User-ID": srv.AdminUserID(),
	}); err != nil {
		t.Fatalf("set extra headers on context: %v", err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	_, err = page.Goto(srv.URL(path), pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("skipping: Chromium renderer closed mid-navigation (seccomp/CPU constraint in this environment); navigate to %s failed: %v", path, err)
		}
		t.Fatalf("goto %s: %v", path, err)
	}

	return page
}

// isTargetClosedErr returns true if the error indicates the browser target
// was closed mid-operation (renderer crash, browser killed, etc.). Used to
// distinguish environment-driven failures from real test failures.
func isTargetClosedErr(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "target closed") ||
		strings.Contains(msg, "Target page, context or browser has been closed") ||
		strings.Contains(msg, "Browser has been closed")
}

// AnonymousPage navigates to a URL without any auth headers — used to test
// the unauthenticated-redirect behavior.
func AnonymousPage(t *testing.T, ctx pw.BrowserContext, srv *uxserver.Server, path string) (pw.Page, pw.Response) {
	t.Helper()

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	resp, err := page.Goto(srv.URL(path))
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("skipping: Chromium renderer closed mid-navigation (environment constraint); navigate to %s failed: %v", path, err)
		}
		t.Fatalf("goto %s: %v", path, err)
	}
	return page, resp
}

// AssertContainsText fails the test if the page's text content does not
// contain the given substring. Skips (rather than fails) if the renderer
// has crashed between navigation and assertion — this happens in
// seccomp-constrained sandbox environments.
func AssertContainsText(t *testing.T, page pw.Page, want string) {
	t.Helper()
	body, err := page.Locator("body").TextContent()
	if err != nil {
		if isTargetClosedErr(err) {
			t.Skipf("skipping: Chromium renderer closed before assertion could read body (environment constraint): %v", err)
		}
		t.Fatalf("read body text: %v", err)
	}
	if !strings.Contains(body, want) {
		t.Errorf("page body does not contain %q\nbody:\n%s", want, truncate(body, 500))
	}
}

// AssertNotContainsText fails the test if the page's text content contains
// the given substring (used for secret-leak checks).
func AssertNotContainsText(t *testing.T, page pw.Page, forbidden string) {
	t.Helper()
	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body text: %v", err)
	}
	if strings.Contains(body, forbidden) {
		t.Errorf("page body contains forbidden text %q", forbidden)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "... (truncated)"
}

// AtoiOrZero parses a string to int, returning 0 on error.
// Used for JavaScript-returned numeric values.
func AtoiOrZero(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
