# Worklog 0174: Remove X-Test-User-ID Auth Bypass from Production Middleware

**Date:** 2026-07-08  
**Epic:** 09 (Security) / 14 (Bug Fixes)  
**Branch:** `fix/remove-test-auth-bypass`

---

## Summary

Removed the `X-Test-User-ID` header authentication bypass from production `RequireAuth` middleware. The bypass allowed any HTTP request with `X-Test-User-ID: <valid_user_id>` to authenticate as that user, bypassing all session validation. It was active in all builds with no build tag or environment gate.

## Approach

1. **Removed** the bypass code from `RequireAuth` in `internal/middleware/rbac.go`
2. **Created** `TestRequireAuth` in `internal/middleware/rbac_test_bypass.go` — a test-only wrapper that provides the same bypass behavior for test server setups
3. **Updated** `tests/uxserver/server.go` and `tests/integration/post_merge_verification_test.go` to use `TestRequireAuth` instead of `RequireAuth`
4. **Added** 5 security regression tests in `rbac_bypass_test.go`:
   - `TestRequireAuth_NoTestBypass` — verifies production `RequireAuth` does NOT honor the header (redirects to login)
   - `TestTestRequireAuth_BypassWorks` — verifies test wrapper DOES honor the header
   - `TestTestRequireAuth_FallsThroughWithoutHeader` — falls through to session auth when header absent
   - `TestTestRequireAuth_InvalidUserIDFallsThrough` — falls through on invalid user ID
   - `TestTestRequireAuth_NonExistentUserFallsThrough` — falls through on non-existent user

## Why this is safe for forward auth and OIDC

Both auth paths create real sessions via `sessionMgr.CreateSession` + `SetSessionCookie`:
- **Forward auth** (`internal/auth/forward_auth.go:41`): reads headers from trusted proxy, creates user, creates session, sets cookie
- **OIDC** (`internal/auth/handlers.go:65`): exchanges code for token, creates user, creates session, sets cookie

The `RequireAuth` middleware validates these sessions via `GetSessionFromRequest` → `GetSession`. The bypass was a shortcut that skipped all of this — removing it doesn't affect either auth path.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green  
**Security Impact**: Critical vulnerability closed. Public deployment is no longer blocked by this issue.
