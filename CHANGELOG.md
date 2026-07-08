# Changelog

All notable changes to TinyRSVP will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-03-05

Initial beta release. Feature-complete for v0 scope. Suitable for homelab / self-hosted use.

### Added

**Core event management**
- Create, edit, publish, cancel, and archive events
- Event deadlines with automatic enforcement
- Auto-archiving of past events (daily background job)
- Timezone-aware date/time handling

**Invite system**
- Cryptographically secure 256-bit invite tokens (HMAC-hashed in database)
- Individual invite creation and management
- Bulk CSV import
- Token revocation and regeneration
- Invite expiration and cleanup

**RSVP handling**
- Yes / No / Maybe responses
- Plus-ones with configurable limits
- Custom preference questions (text, boolean)
- RSVP updates after initial submission
- RSVP deadline enforcement

**Email**
- SMTP invite and confirmation emails
- ICS calendar attachment generation
- Queued email delivery with retry and rate limiting
- Configurable SMTP (supports Gmail, Mailgun, self-hosted, etc.)

**Template system**
- 7 built-in RSVP page themes (Classic, Minimal, Evite-style, and more)
- Custom header images per event
- Color overrides per event
- Go `html/template` based, XSS-safe

**Authentication**
- Forward auth support (Traefik + Authelia, Caddy, Authentik proxy, etc.)
- OIDC support via `go-oidc` (implemented; not yet integration-tested against a real provider in beta)
- Session management with secure cookies
- Role-based access control (admin / event manager)

**Infrastructure**
- SQLite database with automatic migrations
- Local filesystem storage for uploaded images
- Docker / Docker Compose single-container deployment
- Prometheus metrics endpoint
- Health and readiness endpoints
- Graceful shutdown

**UI**
- Mobile-responsive design (mobile-first CSS)
- Dashboard with event overview
- Event management UI
- Invite list and management UI
- Admin panel (user management, system stats)
- RSVP summary view per event

### Known limitations

- **Storage**: Local filesystem only. S3-compatible storage is planned for v1.
- **Database**: SQLite only. PostgreSQL is planned for v1.
- **OIDC**: Implemented but not integration-tested against a real provider. Forward auth is the tested auth path for this release.
- **Auth requirement**: An external auth provider (forward auth proxy or OIDC provider) is required. There are no local username/password accounts by design.
- **Scale**: Designed for small to medium events (up to a few hundred guests per event).

### Security notes

- `TOKEN_SECRET` is required at startup — there is no insecure default fallback.
- All invite tokens are HMAC-hashed in the database; raw tokens are never stored.
- HTTPS is required (TLS termination at reverse proxy).

---

## [0.4.0] - 2026-07-08

Admin dashboard redesign, reusable UI partials + component CSS, and a security fix removing a production auth bypass.

### Added

**Admin dashboard redesign**
- Ops-at-a-glance strip: 4 metric tiles (Users, Events, Invites, System Health) with drilldowns to their respective admin pages
- System panels: Database connection pool and Email queue KPIs rendered inline on the admin dashboard, each linking to `/admin/metrics` for full details
- Quick Actions grid: properly styled action-cards for Users, Settings, Metrics, and raw Prometheus scrape endpoint. Previously the `.action-card` / `.action-grid` classes had **no CSS backing at all** — they rendered as unstyled boxes.
- Optional `AdminSystemHealthProvider` wired into `AdminDashboardHandler`. Best-effort — provider failures are logged but do not block page rendering.

**Reusable UI components**
- New `static/css/components.css` — CSS home for every reusable partial. Design tokens only (no hardcoded hex/rgb). Loaded ambient via `partials/base.html` so any page can use partials without per-file plumbing.
- New partials in `templates/web/partials/components.html`: `section`, `action-card`, `status-badge`, `metric-tile`, `definition-list`. Existing `stats-card` extended with optional `Href` (drilldown), `Icon`, and `Accent` (semantic color variant).
- `templates/web/PATTERNS.md` — cookbook mapping design needs to partial + CSS class, with an explicit DON'T-DO list to prevent future duplication.
- `templates/web/partials/README.md` — rewritten with the extended component set and a TDD workflow for adding new partials.

**Test coverage for the new UI**
- 9 CSS tests asserting class presence, design-token usage, no hardcoded colors, and auto-fit grid layout for `metric-tile-grid`
- 5 CSS guardrail tests bringing `admin_metrics.css` and `admin_settings.css` under the same "tokens only" rule that already covered `dashboard.css`
- 21 template-partial render tests, all including HTML-escaping (XSS) assertions
- 3 handler tests for `SetSystemHealth` — happy path, nil-provider backward-compat, provider errors don't blank the page
- 8 Playwright browser tests covering metric-tile drilldown links, action-card computed styles (proves components.css loaded), and admin_settings/admin_metrics migration

### Fixed

**Security**
- **Removed the `X-Test-User-ID` auth bypass from production middleware.** The `RequireAuth` middleware previously honored any HTTP request with a valid user ID header, bypassing all session authentication. No build tag, no env gate — active in every deployment. Anyone able to enumerate user IDs got full admin access. Fix moves the bypass to a `TestRequireAuth` wrapper in `rbac_test_bypass.go` used only by the test harness, plus 5 security regression tests. (Called out as a known limitation in v0.3.0.)

**UI**
- `.stats-grid` in `dashboard.css` stepped from 1 → 2 → 4 columns via media queries, producing an orphan card whenever a page had 3 stats. Now uses `auto-fit minmax(200px, 1fr)` so 3-, 4-, or 5-card layouts all lay out without desktop orphans.
- `.stats-card` hover state used a hardcoded `rgba(0, 0, 0, 0.05)` shadow. Now uses the `--shadow-md` design token, giving correct behavior in both light and dark themes.
- `admin_metrics.css` and `admin_settings.css` contained hardcoded fallback colors (`#f5f5f5`, `#d4edda`, `#721c24`) that broke dark mode. All replaced with design tokens.

### Changed

**Template system**
- `admin_dashboard.html`, `admin_settings.html`, `admin_metrics.html`, and `user_management.html` all migrated to use the new reusable partials. Per-page CSS (`admin_metrics.css`, `admin_settings.css`) slimmed from ~110 lines each to ~30 lines of truly page-specific concerns.
- Every page now gets `partials/components.html` and `components.css` loaded automatically via `partials/base.html` — no per-page plumbing needed.

**Handler**
- `AdminDashboardHandler` gained a `LastPageData()` getter exposed for testing (matches the `Handler.SetTemplates` pattern already used elsewhere).

### Known limitations (carried forward)

- Not every consumer template has been migrated yet. `dashboard.html`, `event_list.html`, `invite_list.html`, `rsvp_page.html`, `confirmation.html`, `rsvp_summary.html`, `event_form.html`, `event_detail.html`, `event_customization.html`, and `template_editor.html` still hand-roll patterns that could use the new partials. Marked as an Epic 10 follow-up.
- `GetRecentActivity` still loads all user events/invites/RSVPs into memory (from v0.3.0).
- OIDC implementation still not integration-tested against a real provider (from v0.1.0).

---

## [0.3.0] - 2026-07-08

Major quality, testing, and infrastructure release. No breaking changes to user-facing functionality, but significant internal improvements in performance, test coverage, and developer experience.

### Added

**New features (Epic 10)**
- Event list stats: per-event invite count, RSVP count, and accept count displayed on the event list page (single SQL query, no N+1)
- Dashboard clickable activity: activity items now link to the relevant event
- Admin settings page (`/admin/settings`): read-only view of server configuration with all secrets redacted
- Admin metrics dashboard (`/admin/metrics`): business counts, DB connection pool stats, and email queue status

**Testing infrastructure**
- Playwright-based browser UX test harness (replaces chromedp dependency for UX tests)
- 35 browser-level tests (dashboard, event creation, invite management, RSVP flow)
- Post-merge integration tests verifying admin pages, metrics middleware, and secret redaction
- Coverage gap closure: 74.9% → 79.2% statement coverage across all packages
- `tests/uxserver/` shared test server package (eliminates duplication between browser test frameworks)
- `scripts/run_playwright_tests.sh` and `scripts/install_playwright_deps.sh` for sandboxed environments

**CI/CD**
- AI workflow migration: modular three-workflow architecture (issue triage, AI command routing, PR review) with 12 slash commands
- 17 TinyRSVP-specific prompt files for AI assistant
- 99-assertion bash test suite for `route-command.sh` command routing

### Fixed

**Critical bugs**
- Metrics middleware was never wired into the router — `/metrics` endpoint returned zero counters for all HTTP request metrics
- Timeout middleware race condition: concurrent writes to `http.ResponseWriter` caused corrupted responses and intermittent 504s. Replaced with `http.TimeoutHandler`
- Invite `revocation_reason` column was written by `Update` but never read by `GetByID`, `GetByTokenHash`, or `GetByEventIDs` — a subsequent update would clobber the saved reason to NULL
- Event list pagination hardcoded page size 10 in template while handler used `limit=50` — pagination links showed wrong page count
- OIDC return URL lost during provider redirect — users always landed on `/` instead of their original destination. Fixed via short-lived cookie carrying the validated return URL

**Test regressions**
- Hardcoded `2026-06-15` dates in integration tests replaced with dynamic `time.Now().Add(24h)` for future-date safety
- EXIF stripping test assertion relaxed to 150% tolerance for JPEG re-encoding overhead
- `confirmation.html` missing JS includes and ARIA labels

### Changed

**Architecture improvements (codebase cleanup)**
- Dashboard stats: replaced in-memory scan (load ALL events + invites + RSVPs, iterate in Go) with single SQL aggregation query using `LEFT JOIN` + `COUNT(DISTINCT CASE WHEN ...)`. O(1) queries regardless of data volume
- Router refactor: split 716-line `NewRouter` monolith into 7 focused route-registration functions in `router_setup.go`
- `/api/users` routing: replaced hand-rolled path parsing and per-method auth wrapping with proper chi routes
- Page handlers: primary data load failures now use `HandleError` (proper HTTP status codes) instead of HTTP 200 with in-page error text
- Consolidated duplicate `ListFilters` struct (was duplicated between `events` and `repositories` packages)
- Removed dead DB-backed HMAC secret methods (`GetHMACSecret`/`SetHMACSecret`) — never called in production, env var is the single source of truth
- Timeout middleware status code: 504 → 503 (matches HLD `SERVICE_UNAVAILABLE` spec)

**Documentation**
- `docs/TESTING.md` completed and verified against actual generated mocks
- Epic 10 backlog status synced with verified code state (stories 13-22 confirmed complete)
- 11 worklog entries documenting all work

### Security

- Admin settings page uses `SettingsView` DTO that redacts `SMTPPassword`, `OIDC.ClientSecret`, `Token.Secret`, and `Security.HMACSecretKey` — never passes raw secrets to the template layer
- Return URL validation through OIDC flow (prevents open redirect via tampered cookie)
- Return URL cookie cleared after use (prevents replay)

### Known limitations (carried forward)

- `X-Test-User-ID` header bypasses authentication in production code (`internal/middleware/rbac.go`) — deferred to Epic 09 (Security)
- OIDC implementation not integration-tested against a real provider
- `GetRecentActivity` still loads all user events/invites/RSVPs into memory (stats path is fixed, activity path is not)
- 12 empty no-op method bodies in `email/metrics.go` report 0% coverage (Go tooling limitation)

---

[0.4.0]: https://github.com/lenaxia/tinyrsvp/releases/tag/v0.4.0
[0.3.0]: https://github.com/lenaxia/tinyrsvp/releases/tag/v0.3.0
[0.1.0]: https://github.com/lenaxia/tinyrsvp/releases/tag/v0.1.0
