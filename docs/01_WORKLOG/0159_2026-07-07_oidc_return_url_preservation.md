# Worklog 0159: OIDC Return URL Preservation (Epic 10 Story 06)

**Date:** 2026-07-07  
**Epic:** 10 (Technical Debt)  
**Story:** [10_STORY_06_oidc_return_url.md](../00_BACKLOG/10_TECHNICAL_DEBT/10_STORY_06_oidc_return_url.md)  
**Branch:** `fix/oidc-return-url-preservation`

---

## Summary

Implemented return URL preservation through the OIDC login flow so that users are redirected to their original destination after authentication, instead of always landing on `/`. The fix uses a short-lived cookie as the carrier, since the OIDC provider's redirect drops custom query parameters.

## Problem

When an unauthenticated user visited a protected page (e.g. `/events/42/edit`), the auth middleware redirected them to `/login?return=/events/42/edit`. The `LoginHandler` validated the return URL but then called `HandleLogin`, which for OIDC redirected the user to the external provider. The provider only echoes back `code` and `state` on its callback — the `return` parameter was lost. After authentication, the `CallbackHandler` had no record of where the user originally wanted to go, so it always redirected to `/`.

## Root Cause

`LoginHandler.ServeHTTP` (`internal/auth/handlers.go`) read and validated the `return` query parameter but had no mechanism to carry it across the external OIDC redirect. The `CallbackHandler.ServeHTTP` read `return` from the query (which works for forward-auth, where there is no external redirect) but had no fallback for the OIDC case.

## Solution

A short-lived `oidc_return_url` cookie (10-minute MaxAge, HttpOnly, Secure, SameSite=Lax) carries the validated return URL from login initiation through the OIDC provider redirect to the callback.

### Login Handler Changes

1. Validate the `return` query parameter up front (open-redirect protection already existed in `ValidateReturnURL`).
2. Set the validated return URL in the `oidc_return_url` cookie **before** calling `HandleLogin`.
3. Wrap the `http.ResponseWriter` to detect whether `HandleLogin` already wrote a response:
   - **OIDC:** `HandleLogin` writes a 302 redirect to the provider. The wrapper records this; we do not issue a second redirect.
   - **Forward-auth:** `HandleLogin` creates the session, sets the session cookie, and returns without writing a response. We issue the 302 redirect to the return URL ourselves.

This response-writer detection replaces the previous unconditional `http.Redirect` at the end of `LoginHandler.ServeHTTP`, which would have corrupted the OIDC provider redirect by writing a second Location header.

### Callback Handler Changes

1. Resolve the return URL with query-parameter precedence, then cookie fallback:
   - Query `return` param (forward-auth preserves it through its callback).
   - `oidc_return_url` cookie (OIDC preserves it through the provider redirect).
2. Validate whichever value is found via `ValidateReturnURL` (defense-in-depth: a tampered cookie with an absolute URL is rejected and falls back to `/`).
3. Clear the `oidc_return_url` cookie (MaxAge=-1) so it cannot be replayed.
4. Redirect to the validated URL.

## Files Changed

| File | Change |
|------|--------|
| `internal/auth/session.go` | Added `ReturnURLCookieName` (`"oidc_return_url"`) and `ReturnURLMaxAge` (`10 * time.Minute`) constants |
| `internal/auth/handlers.go` | `LoginHandler.ServeHTTP`: set cookie + response-writer detection; `CallbackHandler.ServeHTTP`: cookie fallback + validation + clearing |
| `internal/auth/handlers_test.go` | 8 new tests (TDD: written first, all failed, then implemented) |

## Tests Added (TDD)

All 8 tests written before implementation (red phase), then implementation made them pass (green phase):

| Test | Covers |
|------|--------|
| `TestLoginHandler_StoresReturnURLInCookie` | OIDC case: cookie set, redirect goes to provider |
| `TestLoginHandler_DirectRedirectWhenAuthDoesNotRedirect` | Forward-auth case: no provider redirect, direct redirect to return URL |
| `TestLoginHandler_StoresValidatedReturnURL` | Invalid return URL normalized to `/` before storing |
| `TestCallbackHandler_RetrievesReturnURLFromCookie` | Cookie is the source when query param absent (OIDC) |
| `TestCallbackHandler_QueryReturnTakesPrecedenceOverCookie` | Query param wins when both present (forward-auth) |
| `TestCallbackHandler_ClearsReturnURLCookie` | Cookie deleted after use (MaxAge < 0) |
| `TestCallbackHandler_FallbackToRootWhenNoReturnURL` | No cookie + no query → redirect to `/` |
| `TestCallbackHandler_InvalidReturnURLInCookieFallsBackToRoot` | Tampered cookie with `https://evil.com` blocked → `/` |

## Pre-existing Regressions Also Fixed

This branch also includes the cherry-picked commit `b178e57` from PR #25 (`fix/test-regressions-2026-07`) to resolve pre-existing test failures on `main`:

- `internal/assets/service_integration_test.go` — EXIF stripping assertion relaxed to 150% tolerance
- `internal/handlers/events_integration_test.go` — hardcoded `2026-06-15` dates replaced with `time.Now().Add(24h)`
- `internal/handlers/events_web_integration_test.go` — same date pattern
- `templates/web/confirmation.html` — added missing JS includes + ARIA labels

## Test Results

```
go test -timeout 60s $(go list ./... | grep -v '/tests/ux')
```

All packages pass (excluding UX tests which require headless Chrome). Auth package passes with race detector:

```
ok  github.com/lenaxia/tinyrsvp/internal/auth   3.286s  (with -race)
```

## Note on Dead Code

`internal/handlers/auth.go` contains an `AuthHandlers` struct (`OIDCLogin`, `OIDCCallback`, `ShowLogin`) that is **not wired** in `cmd/server/main.go` — the `AuthHandlers` field of `RouterHandlers` is always nil, so production uses `LoginHandler`/`CallbackHandler` from `internal/auth/handlers.go`. The return URL fix targets the production code path. The dead `AuthHandlers` code is tracked as separate tech debt (its `OIDCCallback` is a stub that discards the `AuthResult` and never creates a session).

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** 8/8 new tests pass; full suite green (excluding UX)  
**Confidence:** HIGH  
**Production Ready:** Yes — pending PR review
