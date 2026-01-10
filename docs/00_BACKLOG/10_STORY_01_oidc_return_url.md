# User Story: OIDC Return URL Preservation

**Epic:** [10_EPIC_technical_debt.md](10_EPIC_technical_debt.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

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

- [ ] Return URL preserved through OIDC login flow
- [ ] Return URL validated for security (no open redirects)
- [ ] User redirected to original destination after login
- [ ] Fallback to `/dashboard` if no return URL specified
- [ ] Return URL stored in session during login initiation
- [ ] Return URL retrieved from session after callback
- [ ] Tests verify return URL preservation
- [ ] Tests verify open redirect prevention still works

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

- [ ] Update OIDCLogin handler to store return URL in session
- [ ] Update CallbackHandler to retrieve return URL from session
- [ ] Add tests for return URL preservation
- [ ] Add tests for fallback to /dashboard
- [ ] Verify open redirect protection still works
- [ ] Test with various return URLs
- [ ] Update documentation

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

- [ ] All acceptance criteria met
- [ ] Return URL preserved through OIDC flow
- [ ] Tests passing (unit + integration)
- [ ] Open redirect protection verified
- [ ] Documentation updated
- [ ] Code reviewed
- [ ] No linter warnings
