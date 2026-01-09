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

- [ ] CSRF tokens generated per session
- [ ] Tokens validated on POST/PUT/DELETE requests
- [ ] Tokens embedded in forms
- [ ] Tokens validated from headers for AJAX
- [ ] SameSite cookie attribute set
- [ ] Token rotation on use
- [ ] Failed validation returns 403
- [ ] Tokens expire with session
- [ ] Double-submit cookie pattern
- [ ] Configurable token length

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

- [ ] Implement CSRF token generation
- [ ] Implement token validation
- [ ] Add form token injection
- [ ] Add header token validation
- [ ] Configure SameSite cookies
- [ ] Test CSRF protection
- [ ] Document CSRF usage

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

- [ ] All acceptance criteria met
- [ ] CSRF protection implemented
- [ ] Tokens validated
- [ ] Tests passing
- [ ] Documentation complete
