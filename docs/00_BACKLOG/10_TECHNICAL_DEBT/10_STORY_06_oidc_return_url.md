# User Story: OIDC Return URL Preservation

**Epic:** [10_EPIC_technical_debt.md](10_EPIC_technical_debt.md)  
**Priority:** Medium  
**Status:** Complete  
**Estimated Effort:** 0.5 days  
**Completed:** 2026-07-07

---

## User Story

As an **event manager**, I want **to be redirected to my original destination after login** so that **I don't lose my place when authentication is required**.

---

## Background

During validation of Epic 08 Story 06 (Authentication Routes), a gap was identified:

**Current Behavior:**
- User visits `/events/123/edit`
- Not authenticated, redirected to `/login?return=/events/123/edit`
- User logs in via OIDC
- User is redirected to `/dashboard` (hardcoded)
- Original destination `/events/123/edit` is lost

**Expected Behavior:**
- User should be redirected to `/events/123/edit` after successful login

**Root Cause:**
The return URL is validated and preserved in [`internal/handlers/auth.go`](../../internal/handlers/auth.go) but not passed through to the underlying authenticator in [`internal/auth/handlers.go`](../../internal/auth/handlers.go), which hardcodes the redirect to `/dashboard`.

---

## Acceptance Criteria

- [x] Return URL preserved through OIDC login flow
- [x] Return URL validated for security (no open redirects)
- [x] User redirected to original destination after login
- [x] Fallback to `/` if no return URL specified
- [x] Return URL stored in cookie during login initiation
- [x] Return URL retrieved from cookie after callback
- [x] Tests verify return URL preservation
- [x] Tests verify open redirect prevention still works

---

## Technical Details

### Package Location
- `internal/auth/handlers.go` - OIDC callback handler
- `internal/auth/session.go` - Session management
- `internal/handlers/auth.go` - Auth route handlers

### Current Implementation

**Login Handler** ([`internal/handlers/auth.go:44-56`](../../internal/handlers/auth.go:44)):
```go
func (h *AuthHandlers) OIDCLogin(w http.ResponseWriter, r *http.Request) {
    returnURL := r.URL.Query().Get("return")
    if returnURL == "" {
        returnURL = "/"
    }
    
    if !validateReturnURL(returnURL) {
        returnURL = "/"
    }
    
    // Return URL is validated but NOT passed to authenticator
    h.authenticator.HandleLogin(w, r)
}
```

**Callback Handler** ([`internal/auth/handlers.go:70`](../../internal/auth/handlers.go:70)):
```go
// Hardcoded redirect to /dashboard
http.Redirect(w, r, "/dashboard", http.StatusFound)
```

### Proposed Solution

**Step 1: Store return URL in session during login**
```go
func (h *AuthHandlers) OIDCLogin(w http.ResponseWriter, r *http.Request) {
    returnURL := r.URL.Query().Get("return")
    if returnURL == "" {
        returnURL = "/dashboard"
    }
    
    if !validateReturnURL(returnURL) {
        returnURL = "/dashboard"
    }
    
    // Store return URL in session before redirecting to OIDC
    session, _ := h.sessionStore.Get(r, "tinyrsvp_session")
    session.Values["return_url"] = returnURL
    session.Save(r, w)
    
    h.authenticator.HandleLogin(w, r)
}
```

**Step 2: Retrieve return URL in callback handler**
```go
func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ... existing authentication logic ...
    
    // Retrieve return URL from session
    session, _ := h.sessionMgr.Get(r, "tinyrsvp_session")
    returnURL, ok := session.Values["return_url"].(string)
    if !ok || returnURL == "" {
        returnURL = "/dashboard"
    }
    
    // Clean up return URL from session
    delete(session.Values, "return_url")
    session.Save(r, w)
    
    http.Redirect(w, r, returnURL, http.StatusFound)
}
```

---

## Tasks

- [x] Update OIDCLogin handler to store return URL in session
- [x] Update CallbackHandler to retrieve return URL from session
- [x] Add tests for return URL preservation
- [x] Add tests for fallback to /dashboard
- [x] Verify open redirect protection still works
- [x] Test with various return URLs
- [x] Update documentation

---

## Dependencies

**Depends on:** 
- Epic 01 (Auth implementation) - Complete
- Epic 08 Story 06 (Login routes) - Complete

**Blocks:** None (quality improvement)

---

## Testing Strategy

### Unit Tests

```go
func TestOIDCLogin_StoresReturnURL(t *testing.T)
func TestOIDCCallback_RetrievesReturnURL(t *testing.T)
func TestOIDCCallback_FallbackToDashboard(t *testing.T)
func TestOIDCCallback_ValidatesReturnURL(t *testing.T)
```

### Integration Tests

```go
func TestAuthFlow_PreservesReturnURL(t *testing.T) {
    // 1. Request protected resource
    // 2. Redirected to login with return URL
    // 3. Complete OIDC flow
    // 4. Verify redirected to original resource
}
```

---

## Security Considerations

- Return URL must be validated before storage (already implemented in auth.go)
- Return URL must be validated before redirect (add validation in callback)
- Session must be secure (HttpOnly, Secure, SameSite)
- Return URL should be removed from session after use (prevent replay)

---

## References

- **Validation Report:** Epic 08 Story 06 validation identified this gap
- **Related Story:** [08_STORY_06_login_routes.md](08_STORY_06_login_routes.md)
- **Auth Package:** [`internal/auth/`](../../internal/auth/)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Return URL preserved through OIDC flow
- [x] Tests passing (unit + integration)
- [x] Open redirect protection verified
- [x] Documentation updated
- [x] Code reviewed
- [x] No linter warnings

---

## Implementation Notes (2026-07-07)

**Approach:** Used a short-lived `oidc_return_url` cookie (10-minute MaxAge, HttpOnly, Secure, SameSite=Lax) instead of the session-based approach proposed in the original story. The OIDC flow does not yet have a session at login-initiation time (the session is only created in the callback), so a cookie is the natural carrier. The cookie is cleared after use to prevent replay.

**Files changed:**
- `internal/auth/session.go` — added `ReturnURLCookieName` and `ReturnURLMaxAge` constants
- `internal/auth/handlers.go` — `LoginHandler.ServeHTTP` now sets the cookie before calling `HandleLogin` and detects whether `HandleLogin` already wrote a response (OIDC redirects to provider) vs. left it untouched (forward-auth creates session and returns); `CallbackHandler.ServeHTTP` now reads the cookie as fallback when the query param is absent (OIDC provider drops custom params), validates it, and clears it.
- `internal/auth/handlers_test.go` — 8 new tests covering cookie storage, direct redirect for forward-auth, validated-URL storage, cookie retrieval in callback, query-param precedence, cookie clearing, fallback to `/`, and open-redirect blocking from a tampered cookie.

**Note on `internal/handlers/auth.go`:** The `AuthHandlers` struct (`OIDCLogin`, `OIDCCallback`, `ShowLogin`) exists but is **not wired** in `cmd/server/main.go` — the `AuthHandlers` field of `RouterHandlers` is always nil, so production uses `LoginHandler`/`CallbackHandler` from `internal/auth/handlers.go`. The OIDC return URL fix is therefore applied to the production code path. The dead `AuthHandlers` code is tracked separately as tech debt (could be removed or wired in a future story).
