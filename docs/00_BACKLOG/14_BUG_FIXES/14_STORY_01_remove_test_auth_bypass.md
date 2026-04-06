# STORY: Remove X-Test-User-ID Auth Bypass from Production Middleware

**Epic:** 14 - Bug Fixes & Code Gaps  
**Story ID:** 14_STORY_01  
**Priority:** Critical  
**Estimated Effort:** 2 hours  
**Severity:** Critical — authentication bypass active in all deployments

---

## Problem

`internal/middleware/rbac.go:16` contains an unconditional authentication bypass:

```go
if testUserID := r.Header.Get("X-Test-User-ID"); testUserID != "" {
    userID, err := strconv.ParseInt(testUserID, 10, 64)
    // ...
    user, err := userService.GetUserByID(r.Context(), userID)
    // ...
    ctx := auth.WithUser(r.Context(), user)
    next.ServeHTTP(w, r.WithContext(ctx))
    return
}
```

There is no build tag, no environment variable gate, no conditional. Any HTTP request to a running production server with the header `X-Test-User-ID: <valid_user_id>` is authenticated as that user, bypassing all session validation.

This was added to support `tests/ux/server_test.go` which uses `asAdmin(userID)` to inject auth headers for browser tests. The intent was test-only but the implementation is in production code.

---

## Acceptance Criteria

- [ ] `internal/middleware/rbac.go` no longer contains the `X-Test-User-ID` check in any form that executes in a normal build
- [ ] The UX tests in `tests/ux/` continue to work (the test server setup must use an alternative auth mechanism)
- [ ] All 32 non-browser packages pass
- [ ] Update `docs/00_BACKLOG/08_API/README.md`: remove ISSUE-1 from Known Issues, mark criterion "Authentication bypass removed" complete
- [ ] Update `docs/00_BACKLOG/09_SECURITY/README.md`: mark the pre-identified finding as resolved
- [ ] Update `docs/00_BACKLOG/14_BUG_FIXES/README.md`: mark this story complete

---

## Technical Approach

### Option A: Build Tag (Recommended)

Move the bypass into a separate file gated with a build tag:

```go
// internal/middleware/rbac_test_hook.go
//go:build testing

package middleware

// enableTestAuthBypass is set to true when compiled with -tags testing.
// This file must never be included in production builds.
var enableTestAuthBypass = true
```

```go
// internal/middleware/rbac.go — production file
// No X-Test-User-ID logic here.
var enableTestAuthBypass = false
```

Then in `RequireAuth`:
```go
if enableTestAuthBypass {
    if testUserID := r.Header.Get("X-Test-User-ID"); testUserID != "" {
        // ... bypass logic
    }
}
```

UX tests then compile with `go test -tags testing ./tests/ux/...`.

### Option B: Replace with Test Server Auth

Refactor `tests/ux/server_test.go` to not use header injection at all. Instead, create a pre-seeded session in the test DB and set the session cookie on the test HTTP client directly. This is the cleanest approach but requires more refactoring of the UX test setup.

### Option C: Environment Variable Gate

Check `os.Getenv("TINYRSVP_TEST_AUTH") == "1"` before activating the bypass. Less safe than a build tag (env can be accidentally set) but simpler to implement.

**Recommendation:** Option A (build tag) for safety; Option B if the UX test suite is being refactored anyway.

---

## Files to Change

- `internal/middleware/rbac.go` — remove or gate the bypass
- `tests/ux/server_test.go` — update `asAdmin()` to match chosen approach
- Possibly: `internal/middleware/rbac_test_hook.go` (new file for build tag approach)

---

## Testing

```bash
# Verify bypass is gone in normal build
go build ./...
curl -H "X-Test-User-ID: 1" http://localhost:8080/events/  # should redirect to login

# Verify UX tests still compile and pass (requires Chrome)
go test -tags testing -timeout 120s ./tests/ux/...
```

---

## Status

- **Status:** Not Started
