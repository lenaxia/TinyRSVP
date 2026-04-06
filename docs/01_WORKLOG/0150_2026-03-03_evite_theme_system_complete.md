# Worklog 0150: Evite-Style Theme System — Complete

**Date:** 2026-03-03
**Author:** AI Assistant
**Type:** Feature + Bug Fix
**Status:** Complete

---

## Executive Summary

Implemented a polished Evite-style RSVP theming system. Users can pick a pre-designed theme from a gallery, optionally swap the header image, and optionally set a custom accent color. Themes render with their full visual design (colors, shapes, SVG header graphics) on both the preview modal and the live RSVP page.

**Test result:** 23/23 packages pass, 0 regressions.

---

## Goal

Replace the coarse `CategoryCard` / `CategoryPlain` two-bucket category system with per-theme categories that directly match CSS filenames. Each of the six styled themes gets its own category constant (e.g. `CategoryWeddingElegance = "wedding-elegance"`), which is used as both the `data-event-theme` HTML attribute value and the CSS filename stem (`/static/css/themes/wedding-elegance.css`).

Custom accent color (`custom_theme_color`) is suppressed when a named theme is active — the theme defines its own palette. The override only applies to plain/no-theme pages.

---

## Architecture Changes

### New per-theme `TemplateCategory` constants (`internal/models/template.go`)

Six new constants added alongside the existing `CategoryPlain` and `CategoryCard`:

```
CategoryWeddingElegance     = "wedding-elegance"
CategoryBirthdayCelebration = "birthday-celebration"
CategoryCorporatePro        = "corporate-professional"
CategoryHolidayFestive      = "holiday-festive"
CategoryGardenParty         = "garden-party"
CategoryModernMinimalist    = "modern-minimalist"
CategoryPlainText           = "plain-text"
```

### `getThemeSlug()` (`internal/handlers/templates.go`)

Maps a `TemplateCategory` to the CSS/attribute slug used in the live page. Per-theme categories pass through directly. Legacy coarse categories (`CategoryCard`, `CategoryModern`, etc.) map to their nearest named equivalent for backwards compatibility with any manually created templates that still carry those values.

`CategoryPlain` is special: it maps to `"plain-text"` inside `getThemeSlug()`, but the RSVP page handler gates the call behind `theme.Category != models.CategoryPlain`, producing `themeCategory = ""` (no CSS, no attribute) for truly plain events.

### `getThemeColor()` (`internal/handlers/templates.go`)

Returns empty string when the active theme is a named theme (any category other than `CategoryPlain` and `CategoryPlainText`). This suppresses the custom color override CSS block for themed pages.

### RSVP page handler (`internal/handlers/rsvp.go`)

Sets `themeCategory = getThemeSlug(theme.Category)` instead of `string(theme.Category)`. Uses the theme's own `ImageURL` as the default header image, overridden by `event.CustomThemeImageURL` when set and non-empty.

### Template HTML (`templates/web/rsvp_page.html`)

When `ThemeCategory` is non-empty the page body wraps content in `.rsvp-card` / `.rsvp-card-header` / `.rsvp-card-content` and renders the header image as `<img class="theme-header-image" ...>`.

### Event handler form parsing (`internal/handlers/events_web.go`)

`UpdateEventFromForm` reads `template_id`, `custom_theme_color`, and `custom_theme_image_url` from the submitted form.

### Invite expiry cascade (`internal/events/service.go`)

When an event's start date changes, existing invite expiry dates are updated proportionally via `inviteRepo.UpdateExpiresAtByEventID`. `events.NewService` gains an `inviteRepo` parameter.

### `UpdateExpiresAtByEventID` (`internal/db/repositories/invite_repository.go`)

New repository method that bulk-updates `expires_at` for all non-revoked invites belonging to an event.

### Migration `000014` (`migrations/sqlite/`)

Fixes any existing DB rows that were seeded with the legacy `category = 'card'` value, rewriting them to the correct per-theme category strings.

### Theme CSS countdown overrides (`static/css/themes/*.css`)

All six theme CSS files received `.event-countdown` color overrides so the countdown timer respects the theme palette.

### Dev overlay (`docker-compose.dev.yml`)

Created a dev Docker Compose overlay that bind-mounts `static/` and `templates/` into the running container so CSS/JS/template changes take effect without a full image rebuild.

---

## Bugs Fixed

### Test: `mockInviteRepository` missing `UpdateExpiresAtByEventID`

The new `InviteRepository` interface method was not implemented on the hand-rolled mock structs in five test files across `internal/invites/`. Added the no-op stub to:

- `service_test.go` — `mockInviteRepository`
- `service_send_test.go` — `mockSendInviteRepo`
- `service_update_test.go` — `mockUpdateInviteRepo`

(`service_import_test.go` already had the stub; `service_individual_test.go` uses the shared mock from `service_test.go`.)

### Test: `TestThemePreviewIntegration_ResponsiveImageDisplay`

`HandleThemePreview` produced a plain `<img class="theme-header-image" ...>` tag. The test expected the image to carry inline responsive styles. Fixed by adding `style="width: 100%; object-fit: cover; max-height: 400px;"` to the img element.

### Test: `TestSeeder_EndToEnd_ApplicationStartup`

Asserted `defaultTheme.Category == models.CategoryPlain`, but "Simple & Clean" now has `CategoryPlainText`. Updated assertion to `CategoryPlainText`.

### Test: `TestSeeder_EndToEnd_ThemeRetrieval`

Queried `GetTemplatesByCategory` for `CategoryPlain` (expect 1) and `CategoryCard` (expect 6). No themes carry those categories anymore. Updated to query `CategoryPlainText` (1) and `CategoryWeddingElegance` (1 spot-check).

### Test: `TestSeeder_GetDefaultThemes_ReturnsSevenThemes`

Counted `CategoryPlain` and `CategoryCard` themes; both returned 0. Replaced with counts for `CategoryPlainText` (expect 1) and the six named-theme categories (expect 6 total).

---

## Files Changed

| File | Change |
|---|---|
| `internal/models/template.go` | Added 7 per-theme `TemplateCategory` constants |
| `internal/templates/seeder.go` | Each theme uses its own per-theme category; Birthday Celebration `HTMLContent` fix |
| `internal/handlers/templates.go` | `getThemeSlug()` per-theme mapping; `getThemeColor()` suppresses override for named themes; `HandleThemePreview` inline img styles |
| `internal/handlers/rsvp.go` | `themeCategory = getThemeSlug(...)` instead of raw string; theme image URL logic |
| `internal/handlers/events_web.go` | `UpdateEventFromForm` reads `template_id`, `custom_theme_color`, `custom_theme_image_url` |
| `templates/web/rsvp_page.html` | `.rsvp-card` wrapper + `theme-header-image` class when theme is active |
| `internal/events/service.go` | `inviteRepo` field; cascade invite expiry on event start date change |
| `internal/db/repositories/invite_repository.go` | Added `UpdateExpiresAtByEventID` method and interface entry |
| `internal/testutil/mocks/repositories/mock_invite_repository.go` | Mock for `UpdateExpiresAtByEventID` |
| `cmd/server/main.go` | Passes `inviteRepo` to `events.NewService` |
| `internal/invites/service_test.go` | Added `UpdateExpiresAtByEventID` stub to `mockInviteRepository` |
| `internal/invites/service_send_test.go` | Added `UpdateExpiresAtByEventID` stub to `mockSendInviteRepo` |
| `internal/invites/service_update_test.go` | Added `UpdateExpiresAtByEventID` stub to `mockUpdateInviteRepo` |
| `internal/events/service_publicid_test.go` | Fixed 4 calls to `NewService` (missing `nil` inviteRepo arg) |
| `internal/templates/seeder_e2e_test.go` | Updated category assertions to `CategoryPlainText`; updated `GetTemplatesByCategory` test cases |
| `internal/templates/seeder_integration_test.go` | Replaced `CategoryPlain`/`CategoryCard` counts with `CategoryPlainText` + named-theme counts |
| `migrations/sqlite/000014_fix_theme_categories.up.sql` | Rewrites legacy `'card'` category rows to per-theme values |
| `migrations/sqlite/000014_fix_theme_categories.down.sql` | Reverts to `'card'` |
| `static/css/themes/wedding-elegance.css` | `.event-countdown` color overrides |
| `static/css/themes/birthday-celebration.css` | `.event-countdown` color overrides |
| `static/css/themes/corporate-professional.css` | `.event-countdown` color overrides |
| `static/css/themes/holiday-festive.css` | `.event-countdown` color overrides |
| `static/css/themes/garden-party.css` | `.event-countdown` color overrides |
| `static/css/themes/modern-minimalist.css` | `.event-countdown` color overrides |
| `docker-compose.dev.yml` | Created: dev overlay with bind mounts for `static/` and `templates/` |

---

## Key Discoveries

- **Template/static files are baked into the Docker image binary.** The dev overlay (`docker-compose.dev.yml`) bind-mounts `static/` and `templates/` so changes take effect without a rebuild.
- **`getThemeSlug(CategoryPlain)` returns `"plain-text"`, not `"".`** The RSVP handler gates the call with `theme.Category != models.CategoryPlain` to produce an empty slug for plain events, which suppresses the theme CSS link and `data-event-theme` attribute.
- **`CategoryCard` is now a legacy-only constant.** No seeded theme uses it. The `getThemeSlug` fallback maps it to `"wedding-elegance"` to handle any manually created templates that still carry the old value.

---

## Test Results

```
ok  github.com/lenaxia/tinyrsvp/internal/admin
ok  github.com/lenaxia/tinyrsvp/internal/assets
ok  github.com/lenaxia/tinyrsvp/internal/auth
ok  github.com/lenaxia/tinyrsvp/internal/config
ok  github.com/lenaxia/tinyrsvp/internal/db
ok  github.com/lenaxia/tinyrsvp/internal/db/repositories
ok  github.com/lenaxia/tinyrsvp/internal/email
ok  github.com/lenaxia/tinyrsvp/internal/events
ok  github.com/lenaxia/tinyrsvp/internal/handlers
ok  github.com/lenaxia/tinyrsvp/internal/invites
ok  github.com/lenaxia/tinyrsvp/internal/jobs
ok  github.com/lenaxia/tinyrsvp/internal/middleware
ok  github.com/lenaxia/tinyrsvp/internal/models
ok  github.com/lenaxia/tinyrsvp/internal/rsvp
ok  github.com/lenaxia/tinyrsvp/internal/storage
ok  github.com/lenaxia/tinyrsvp/internal/templates
ok  github.com/lenaxia/tinyrsvp/internal/templates/defaults
ok  github.com/lenaxia/tinyrsvp/internal/testutil
ok  github.com/lenaxia/tinyrsvp/internal/testutil/builders
ok  github.com/lenaxia/tinyrsvp/cmd/server
ok  github.com/lenaxia/tinyrsvp/pkg/eventid
ok  github.com/lenaxia/tinyrsvp/pkg/ics
ok  github.com/lenaxia/tinyrsvp/pkg/token
```

23/23 packages pass. 0 regressions. Docker image rebuilt and confirmed healthy on port 8080.

---

## Status

**Status:** ✅ Complete
**Test Pass Rate:** 100% (23/23 packages)
**Confidence:** HIGH (95%)
**Production Ready:** Yes
**Known Issues:** None
