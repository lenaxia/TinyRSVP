# Worklog: Protect /metrics Endpoint and Role-Gate Admin Nav

**Date:** 2026-08-01  
**PR:** [#61](https://github.com/lenaxia/TinyRSVP/pull/61)  
**Issues:** #7, #13

## Summary

Two security/UX fixes: gated the raw Prometheus `/metrics` endpoint behind an IP allowlist (loopback default + configurable), and role-gated the Admin navigation link so non-admins no longer see it (and hit a 403).

## Work Completed

- [x] **#7 — `/metrics` IP allowlist**: New `MetricsIPAllowlist` middleware (`internal/middleware/metrics_allowlist.go`) permits loopback (`127.0.0.0/8`, `::1`) by default and any IPs/CIDRs in the new `METRICS_TRUSTED_IPS` env var. Client IP read via `GetRealIP` so it works behind a proxy. Wired in `cmd/server/main.go`; updated the startup log line (`protection: ip-allowlist`).
- [x] Extracted shared IP/CIDR parsing + validation into `config/ip_list.go`, reused by forward-auth (removed duplicated validation in `forward_auth.go`).
- [x] Added `MetricsConfig{ TrustedIPs }` to the Config struct, loaded via `loadFromEnv`, validated via `validateMetrics`.
- [x] Documented `METRICS_TRUSTED_IPS` in `.env.example`.
- [x] **#13 — admin nav role-gating**: Added `IsAdmin bool` to each staff page-data struct + the invites data map; populated via shared `isAdminRequest(r)` helper (`internal/handlers/nav.go`) reading role from the auth context; gated the link with `{{if .IsAdmin}}` in `navigation.html`.

## Decisions Made

### Decision 1: IP allowlist vs RequireAuth for /metrics
**Context:** Issue #7 suggested "RequireAuth or IP allowlist".  
**Decision:** IP allowlist. Prometheus scrapers cannot authenticate via browser session cookies, so RequireAuth would break scraping. The human-readable `/admin/metrics` dashboard (behind admin auth) is unchanged and remains for operators.

### Decision 2: loopback default vs deny-all default
**Context:** When `METRICS_TRUSTED_IPS` is unset, what should the default be?  
**Decision:** Loopback-only. This keeps same-host Prometheus scraping working out of the box (the common homelab case) while closing public exposure. Operators scraping from another host/container set `METRICS_TRUSTED_IPS` to the scraper IP or Docker subnet.

### Decision 3: per-struct IsAdmin field vs embedded NavData
**Context:** ~12 page-data structs render the nav.  
**Decision:** Add `IsAdmin bool` directly to each struct (flat field) + a shared `isAdminRequest(r)` helper, rather than introducing an embedded `NavData` type. Lower risk (no field-removal refactor, no reliance on field promotion), matches the existing flat `ActivePage` style.

## Files Changed

- `internal/middleware/metrics_allowlist.go` — new IP allowlist middleware
- `internal/config/ip_list.go` — shared IP/CIDR parse + validate helpers
- `internal/config/config.go` — `MetricsConfig` + load/validate wiring
- `internal/config/forward_auth.go` — refactored to use shared helpers
- `cmd/server/main.go` — wrap metrics handler with allowlist; update log
- `.env.example` — documented `METRICS_TRUSTED_IPS`
- `internal/handlers/nav.go` — `isAdminRequest(r)` helper
- `internal/handlers/{admin,dashboard,events_web,invites_web,metrics,settings,rsvp_summary,template_editor}.go` — `IsAdmin` field + population
- `templates/web/partials/navigation.html` — `{{if .IsAdmin}}` around Admin link
- `templates/web/rsvp_summary_test.go` — add `IsAdmin` to local test struct

## Tests

- [x] `internal/middleware/metrics_allowlist_test.go` — 11 cases (loopback default allow/deny, explicit IP/CIDR, X-Forwarded-For, malformed addr)
- [x] `internal/config/metrics_test.go` — env parsing + validation
- [x] `templates/web/partials/navigation_test.go` — admin link shown/hidden; Dashboard/Events always present
- [x] Updated `forward_auth_test.go` for improved error messages
- [x] All 39 non-browser packages pass; browser tests unaffected (skip in this sandbox)

## Notes

- The `review` GitHub check shows FAILURE, but the formal review is **APPROVE**; the check failure is the bot failing to git-push (403 permission error), not a code rejection.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** All non-browser tests pass  
**Confidence:** HIGH  
**Production Ready:** Yes
