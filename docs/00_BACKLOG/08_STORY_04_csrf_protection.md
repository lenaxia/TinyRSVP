# User Story: CSRF Protection

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1 day

---

## User Story

As a **security engineer**, I want **CSRF protection on all state-changing operations** so that **the application is protected against cross-site request forgery attacks**.

---

## Acceptance Criteria

- [x] CSRF tokens generated per session
- [x] Tokens validated on POST/PUT/DELETE requests
- [x] Tokens embedded in forms
- [x] Tokens validated from headers for AJAX
- [x] SameSite cookie attribute set
- [x] Token rotation on use
- [x] Failed validation returns 403
- [x] Tokens expire with session
- [x] Double-submit cookie pattern
- [x] Configurable token length

---

## Technical Details

### Package Location
- `internal/middleware/csrf.go` - CSRF middleware
- `internal/middleware/csrf_test.go` - Tests

### CSRF Implementation

```go
func CSRF(secret string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if isSafeMethod(r.Method) {
                next.ServeHTTP(w, r)
                return
            }
            
            token := getCSRFToken(r)
            if !validateCSRFToken(token, secret) {
                http.Error(w, "Invalid CSRF token", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## Tasks

- [x] Implement CSRF token generation
- [x] Implement token validation
- [x] Add form token injection
- [x] Add header token validation
- [x] Configure SameSite cookies
- [x] Test CSRF protection
- [x] Document CSRF usage

---

## Dependencies

**Depends on:** 08_STORY_01_middleware_chain.md

**Blocks:** All mutation route stories

---

## Testing Strategy

```go
func TestCSRF_TokenGeneration(t *testing.T)
func TestCSRF_TokenValidation(t *testing.T)
func TestCSRF_SafeMethods(t *testing.T)
func TestCSRF_InvalidToken(t *testing.T)
```

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **OWASP:** CSRF Prevention Cheat Sheet

---

## Definition of Done

- [x] All acceptance criteria met
- [x] CSRF protection implemented
- [x] Tokens validated
- [x] Tests passing
- [x] Documentation complete

---

## Status

**Status:** Complete (Including Template Integration)
**Completed:** 2026-01-10
**Template Integration:** 2026-01-10
**Implementation Notes:**
- CSRF middleware implemented with double-submit cookie pattern
- Token rotation on every state-changing request
- Constant-time comparison for timing attack resistance
- Configurable token length (default 32 bytes)
- Comprehensive test coverage (unit, integration, benchmark)
- Integrated into router middleware chain
- Documentation updated in middleware README
- **Template Integration Complete:**
  - CSRF tokens injected into handler template data (rsvp.go)
  - CSRF hidden fields added to all forms (rsvp_page.html, event_form.html)
  - JavaScript helper created (static/js/csrf.js)
  - Integration tests verify end-to-end functionality (10 tests)
  - All tests passing
