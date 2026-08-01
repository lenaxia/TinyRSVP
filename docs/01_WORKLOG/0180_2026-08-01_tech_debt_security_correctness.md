# Worklog: Tech-Debt Fixes — Security & Correctness Cluster

**Date:** 2026-08-01  
**Branch:** `fix/tech-debt-security-correctness`

## Summary

First batch from the deep tech-debt review: fixed 9 live bugs (XSS, pagination count, silent data loss, config/logging/error-handling inconsistencies). All were verified against code before changing.

## Changes

### Security
1. **C1 — Theme preview XSS (CRITICAL)** — `HandleThemePreview` interpolated query params (`title`, `location`, `description`) unescaped via `fmt.Fprintf`. Moved to `html/template` rendering (auto-escaping); CSS color still gated by `isValidHexColor`. Added 4 XSS regression tests.

### Correctness
2. **C2 — ListEvents pagination Total bug** — Both API and web handlers set `Total: len(results)`, truncating the total whenever `limit` limits the page. Added `CountByFilters` to the event repo and `CountEvents` to the service (honoring creator/status filters + authz scoping); wired into both handlers. Worklog 0175 falsely claimed this was done — now actually done.
3. **C3 — RSVP repo drops adults_count/kids_count** — `RSVPRepository` INSERT/SELECT/UPDATE omitted the columns while `rsvp/service.go` worked around it with direct SQL. Added the columns to all 6 repo sites and to the GetStats guest sum.
4. **H1 — help_text never persisted** — Column + model + template existed but repo INSERT/SELECT/UPDATE and handler request structs omitted it. Wired `help_text` end-to-end.
5. **H2 — ICS SEQUENCE never incremented** — `UpdateWithVersion` now does `ics_sequence = ics_sequence + 1`, so calendar clients see reschedules (RFC 5545).

### Logging
6. **H8 — `logError` on `log.Printf`** — Central error path now uses `slog.Error`/`slog.Warn` with structured fields.
7. **H9 — `fmt.Printf` in confirmation render** — Now `slog.Error`.
8. Removed `log.Printf` double-logs before `HandleError` in dashboard/admin/metrics; converted non-fatal warnings to `slog.Warn`.

### Error handling
9. **H10 — RSVP string-sniffs error messages** — `handleSubmitError`/`handleUpdateError` mapped status codes via `strings.Contains(err.Error(), "expired")` etc. The service now returns typed `ForbiddenError`s (expired/revoked/cancelled); `toAPIError` gained `DeadlinePassedError` → 403 and `rsvp.ErrDuplicateRSVP` → 409 mappings; the JSON path delegates to `HandleError`.

### Docs
10. **C4 — OIDC redirect path mismatch** — `.env.example` documented `/auth/oidc/callback` but production registers `/auth/callback`. Fixed the doc to match reality. The unfinished `handlers.AuthHandlers` OIDC refactor is an intended future feature (auth system) — flagged for a separate completion PR, not removed.

## Tests
- New: `theme_preview_xss_test.go` (4 cases), pagination-total regression case.
- Updated: rsvp service tests (typed ForbiddenError), handler test mocks (+CountEvents/CountByFilters), rsvp integration assertion (canonical not-found message).
- Full suite: all 39 non-browser packages pass; `go build`/`go vet` clean.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** All non-browser tests pass  
**Confidence:** HIGH  
**Production Ready:** Yes
