# 0175 — Admin dashboard redesign + reusable partials cookbook

**Date:** 2026-07-08
**Branch:** `feature/admin-dashboard-and-partials`
**Type:** UI refactor + technical-debt cleanup
**Status:** ⚠️ Complete (with follow-ups noted below)

## Summary

Redesigned the admin dashboard from three unstyled stats cards + three
dead-looking nav "boxes" into a proper ops-overview page with drilldowns
and system panels. Along the way I extracted the UI patterns that were
being copy-pasted across ~10 templates into partials + a component CSS file,
and documented them in a cookbook so future work doesn't reinvent them.

## What was broken (before)

Inspected on 2026-07-07 pre-rebase, main tag `v0.3.0`:

1. `templates/web/admin_dashboard.html`: the "Quick Actions" section used
   classes `.action-grid` and `.action-card` that **had zero CSS backing**.
   Rendered as unstyled boxes with no border, no padding, no hover.
2. `partials/components.html` already defined `stats-card`, `empty-state`,
   `loading-state`, `error-state` partials, but **every consumer template
   ignored them and copy-pasted the HTML**. Documented but not enforced.
3. `dashboard.css` `.stats-grid` stepped from 1 → 2 → 4 columns via media
   queries. With three admin cards this produced an orphan on desktop.
4. `admin_metrics.css` and `admin_settings.css` had hardcoded fallback
   colors (`#f5f5f5`, `#d4edda`, `#721c24`) that broke dark mode. The
   design-system rule "tokens only" was only enforced on `dashboard.css`
   via `TestDashboardNoHardcodedColors`.
5. `admin_dashboard.html` had no drilldowns — clicking a stats card did
   nothing.
6. `user_management.html` reimplemented `dashboard-header` inline instead
   of using the `page-header` partial.

## What changed

### New reusable partials (`templates/web/partials/components.html`)

Added: `section`, `action-card`, `status-badge`, `metric-tile`,
`definition-list`. Extended existing `stats-card` with `Href`, `Icon`, and
`Accent` (color variant) options. All partials have unit tests in
`templates/web/components_partials_test.go` that assert both rendering and
HTML escaping.

### New CSS file (`static/css/components.css`)

Backs every partial. Design tokens only, dark-mode-safe via existing
`variables.css` semantics. Guarded by:

- `TestComponentsRequiredClasses` — every class the partials produce exists
- `TestComponentsNoHardcodedColors` — no hex/rgb literals
- `TestComponentsUsesDesignTokens` — references `--spacing-*`, `--color-*`,
  etc.
- `TestComponentsActionCardIsInteractive` — hover + focus states exist
  (this is the "action-card boxes were dead" fix, made testable)
- `TestComponentsMetricTileGridIsFlexible` — auto-fit/minmax layout

Loaded ambient in `partials/base.html` — every page gets it via `css-common`.

### Handler wiring

`AdminDashboardHandler` gained an optional `AdminSystemHealthProvider`
dependency via `SetSystemHealth()`. When set, the page data includes
`EmailQueue` and `DBPool` KPIs. Failures in the provider are logged but
don't block the page (best-effort ops overview). `MetricsDataSource`
already implements the interface, so wiring is one line in `main.go` and
`tests/integration`.

Tests in `internal/handlers/admin_system_health_test.go` cover:
1. Data flows through when provider is set (happy path)
2. `nil` provider = `nil` health data, business stats still render
   (backward-compat guarantee for anyone constructing the handler alone)
3. Provider errors don't blank the page

### Admin dashboard redesign

`admin_dashboard.html` now shows:

- **At-a-glance strip** (`metric-tile-grid`): 4 tiles — Users, Events,
  Invites, System Health. Users → `/admin/users`, Events → `/events`,
  System Health → `/admin/metrics`. Invites has no dedicated list page yet.
- **System panels**: when `SetSystemHealth` is wired, renders DB Pool and
  Email Queue panels with the top KPIs + "View details →" link to
  `/admin/metrics`. Both hidden gracefully when data unavailable.
- **Quick Actions**: styled `action-card` grid (Users, Settings, Metrics,
  Prometheus). No longer dead boxes.

### Migrated templates

Rewrote to use the new partials:

- `admin_dashboard.html` — full redesign
- `admin_settings.html` — `.ui-section` + `.definition-list` + `status-badge`
  for redacted-secret states. Retained the `••••••••` redaction chip
  alongside the "Set" badge so the visual clue that secrets ARE set (not
  just "labeled set") is preserved. Both `admin_metrics.css` and
  `admin_settings.css` slimmed to page-only concerns (~30 lines each,
  down from ~110).
- `admin_metrics.html` — `.ui-section` + `.metric-tile-grid` +
  `.definition-list` + `status-badge`
- `user_management.html` — `page-header` + `data-table` + `empty-state` +
  `error-state` + `loading-state` partials

### CSS polish (`static/css/dashboard.css`)

- `.stats-grid` now uses `auto-fit minmax(200px, 1fr)` — no more desktop
  orphans regardless of card count. Enforced by
  `TestDashboardStatsGridIsAutoFit`.
- `.stats-card` gained a proper hover state (border color + `--shadow-md`
  + subtle lift transform). Anchor variant has focus outline.
- Removed hardcoded `rgba(0, 0, 0, 0.05)` shadow.

### Documentation

- `templates/web/PATTERNS.md` (new): cookbook mapping design needs to
  partial + CSS class. "I want a big number → use metric-tile", "I need a
  key/value grid → use .definition-list", etc. Includes explicit
  DON'T-DO list.
- `templates/web/partials/README.md`: rewritten to reflect the extended
  component set + new "adding a new partial" TDD workflow.

## Tests

- **CSS**: 100% pass, incl. new tests for `components.css`,
  `admin_metrics.css`, `admin_settings.css` (no hardcoded colors), and
  `dashboard.css` (auto-fit grid).
- **Template partials**: new `templates/web/components_partials_test.go`
  covers stats-card (5 tests), section (3), action-card (3), status-badge
  (3), metric-tile (3), definition-list (3). All include XSS-escaping
  assertions.
- **Handler**: 3 new tests in `admin_system_health_test.go`.
- **Playwright UX**: new `admin_dashboard_flow_test.go` — 8 tests
  covering page load, at-a-glance stats present, metric tiles are
  drilldown links (`a.metric-tile`), action-cards have non-zero
  computed `border-radius` (proves `components.css` loads), auth guard,
  and admin_settings + admin_metrics use the new partials.
- **Integration**: `tests/integration/post_merge_verification_test.go`
  updated to include `partials/components.html` in ParseFiles and wire
  `SetSystemHealth` for the /admin drilldown links to render.

Full suite: **all 35 packages pass**. `-race` clean on affected packages.

## Confidence & production readiness

**Confidence:** HIGH (~90%)

Rationale:
- All tests pass (unit, template, handler, integration, browser UX)
- Zero hardcoded colors introduced anywhere (guarded)
- Dark mode inherits correctly from existing token system
- Backward-compat: handlers still work without `SetSystemHealth`; existing
  templates that use `stats-card`/`empty-state`/etc. still work with
  ambient components.html loading

Reasons not 100%:
- Playwright browser tests skipped ~2 assertions the harness marks as
  "renderer crashed in sandbox" — same behavior as pre-existing tests
  and out of my control here.
- I only visually inspected the redesign via Playwright locator counts,
  not human eyes on rendered pixels. If the balance/density feels off,
  a follow-up polish pass may be needed.

**Production ready:** Yes for the admin dashboard, settings, metrics,
user_management pages. Other pages (dashboard, event_list, invite_list,
event_form, rsvp_page) were **not migrated** in this change — they're a
separate follow-up.

## Follow-ups (Epic 10 candidates)

1. Migrate remaining templates (`dashboard.html`, `event_list.html`,
   `invite_list.html`, `rsvp_page.html`, `confirmation.html`,
   `rsvp_summary.html`, `event_form.html`, `event_detail.html`,
   `event_customization.html`, `template_editor.html`) to use
   `components.css` partials.
2. Delete `.dashboard-header` from `dashboard.css` once every consumer
   uses `page-header` partial — currently kept because a few unmigrated
   templates still reference it.
3. Consider adding an "Invites" list page and wiring the Invites
   metric-tile to it.
4. Playwright renderer instability under seccomp — the pre-existing
   `AsAdminPage` helper already handles this with `t.Skip`, but the
   silent skips can hide real regressions. Follow-up to add a summary
   log at end of run.

## Files touched

New files:
- `static/css/components.css`
- `static/css/components_test.go`
- `static/css/admin_pages_test.go`
- `templates/web/components_partials_test.go`
- `templates/web/PATTERNS.md`
- `internal/handlers/admin_system_health_test.go`
- `tests/ux_playwright/admin_dashboard_flow_test.go`
- `docs/01_WORKLOG/0175_2026-07-08_admin_dashboard_and_partials.md` (this file)

Modified:
- `README-LLM.md` — branch table entry
- `cmd/server/main.go` — include `partials/components.html` in every
  ParseFiles call; wire `SetSystemHealth`; funcmap on user_management
- `internal/handlers/admin.go` — `AdminSystemHealthProvider` + `SetSystemHealth`
- `static/css/dashboard.css` — auto-fit grid, hover polish, no rgba()
- `static/css/admin_metrics.css` — slimmed to page-only concerns
- `static/css/admin_settings.css` — slimmed, added `.redacted-value`
- `templates/web/admin_dashboard.html` — full redesign
- `templates/web/admin_metrics.html` — migrated to partials
- `templates/web/admin_settings.html` — migrated to partials
- `templates/web/user_management.html` — migrated to partials
- `templates/web/partials/base.html` — include components.css in css-common
- `templates/web/partials/components.html` — extended
- `templates/web/partials/README.md` — updated docs
- `templates/web/testhelper_test.go` — `parseWithBase` now includes
  components.html; new `parsePartialsOnly` for partial-level tests
- `tests/uxserver/server.go` — include `partials/components.html`
- `tests/integration/post_merge_verification_test.go` — same + wire
  SetSystemHealth for drilldown-link tests
