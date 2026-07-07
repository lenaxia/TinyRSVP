# Worklog 0164: Playwright UX Test Harness Setup

**Date:** 2026-07-07  
**Epic:** 12 (Test Infrastructure)  
**Branch:** `feat/playwright-setup`  
**PR:** #36

---

## Summary

Replaced the chromedp UX test framework with Playwright (via `playwright-go`). Playwright downloads its own pinned Chromium browser, which unblocks UX testing in sandboxed environments where system Chrome isn't installed and can't be `apt-get install`ed.

## Motivation

The existing chromedp tests (`tests/ux/`) require system-installed Chrome, which isn't available in this sandbox environment. As a result, no browser-level UX tests have run since the project started — only unit and integration tests. This was the gap that prompted the "verify the work" question earlier: I had merged 5 PRs of changes (admin pages, event list stats, dashboard clickable events, metrics middleware fix) without any browser-level validation.

Playwright solves this by downloading a pinned Chromium binary as part of its install. The only remaining requirement is the shared library dependencies (libglib, libnss, libxcb, libgbm, etc.), which `playwright install-deps chromium` normally handles via apt.

## Sandbox environment workaround

This sandbox has no root and no apt sources configured, so `playwright install-deps` fails. I wrote `scripts/install_playwright_deps.sh` which:

1. Downloads ~25 `.deb` packages directly from the Debian mirror via HTTP
2. Extracts them with `dpkg-deb -x` into `/tmp/pwlibs/extracted/` (no root)
3. Returns the `LD_LIBRARY_PATH` to set

`scripts/run_playwright_tests.sh` automates the whole flow: downloads deps if missing, sets `LD_LIBRARY_PATH`, runs the tests.

On a normal Debian/Ubuntu system with root access, none of this is needed — `playwright install-deps chromium` handles it natively.

## Files Added

| File | Purpose |
|---|---|
| `tests/ux_playwright/harness.go` | `SetupTestServer`, `NewBrowser`, `AsAdminPage`, `AssertContainsText`, `AssertNotContainsText` helpers |
| `tests/ux_playwright/dashboard_poc_test.go` | 4 PoC tests + unhappy-path tests |
| `scripts/run_playwright_tests.sh` | Launcher script with LD_LIBRARY_PATH setup |
| `scripts/install_playwright_deps.sh` | Downloads Chromium's shared lib deps from Debian mirror |
| `go.mod` / `go.sum` | Adds `github.com/mxschmitt/playwright-go` |

## Harness design

The harness mirrors the existing chromedp `tests/ux/server_test.go` pattern but is minimal:

- `SetupTestServer(t)` — wires a real router with dashboard, admin, settings, metrics, RSVP handlers against a file-backed SQLite database, returns the test server + admin user ID
- `NewBrowser(t)` — launches headless Chromium via Playwright, returns a browser context with automatic cleanup
- `AsAdminPage(t, ctx, srv, path)` — sets the `X-Test-User-ID` header on the browser context, creates a page, navigates to `path`, waits for network idle

Currently wires only what's needed for the PoC tests. The full migration (all 5 chromedp test files: 27 tests total) is a follow-up that will extend the harness with events, invites, and full RSVP service wiring.

## Tests

### Happy-path (PoC, originally in this PR)

- `TestDashboard_LoadsInBrowser` — dashboard renders with stats cards and Recent Activity section; clicking `/admin` navigates successfully
- `TestAdminDashboard_HasSettingsAndMetricsLinks` — admin dashboard links to `/admin/settings` and `/admin/metrics`
- `TestAdminSettings_Renders` — settings page renders with Server, Database, Authentication sections
- `TestAdminMetrics_Renders` — metrics page renders with Business Metrics, DB Pool, Email Queue sections

### Unhappy-path (added per review feedback)

- `TestDashboard_UnauthenticatedRedirect` — no `X-Test-User-ID` header → redirect to `/login`
- `TestNonExistentAdminPage_Returns404` — admin visiting non-existent `/admin/nonexistent` gets 404 (not a server error)
- `TestDashboard_RendersOnEmptyDatabase` — empty DB shows zero counts and "No Recent Activity" empty state
- `TestStaticAssets_Load` — every stylesheet link returns HTTP < 400 (catches broken static file server wiring)

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** All PoC tests pass in ~10 seconds  
**Confidence:** HIGH — verified against real headless Chromium in this environment  
**Production Ready:** Yes — follow-up migration is a separate story
