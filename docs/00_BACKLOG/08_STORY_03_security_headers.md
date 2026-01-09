# User Story: Security Headers Middleware

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As a **security engineer**, I want **comprehensive security headers on all responses** so that **the application is protected against common web vulnerabilities**.

---

## Acceptance Criteria

- [ ] Content-Security-Policy (CSP) header set
- [ ] Strict-Transport-Security (HSTS) header set
- [ ] X-Content-Type-Options header set
- [ ] X-Frame-Options header set
- [ ] X-XSS-Protection header set
- [ ] Referrer-Policy header set
- [ ] Permissions-Policy header set
- [ ] Headers configurable via environment
- [ ] Headers tested on all routes
- [ ] CSP violations logged

---

## Technical Details

### Package Location
- `internal/middleware/security_headers.go` - Security headers middleware
- `internal/middleware/security_headers_test.go` - Tests

### Security Headers

```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Content-Security-Policy", buildCSP())
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        next.ServeHTTP(w, r)
    })
}
```

---

## Tasks

- [ ] Implement security headers middleware
- [ ] Configure CSP policy
- [ ] Configure HSTS policy
- [ ] Make headers configurable
- [ ] Add CSP violation reporting
- [ ] Test headers on all routes
- [ ] Document security headers
- [ ] Security audit

---

## Dependencies

**Depends on:** 08_STORY_01_middleware_chain.md

**Blocks:** None

---

## Content Security Policy

```go
func buildCSP() string {
    directives := []string{
        "default-src 'self'",
        "script-src 'self' 'unsafe-inline'",
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' data: https:",
        "font-src 'self'",
        "connect-src 'self'",
        "frame-ancestors 'none'",
        "base-uri 'self'",
        "form-action 'self'",
    }
    return strings.Join(directives, "; ")
}
```

---

## Testing Strategy

```go
func TestSecurityHeaders(t *testing.T)
func TestCSP(t *testing.T)
func TestHSTS(t *testing.T)
```

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **OWASP:** Security Headers Best Practices

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Security headers implemented
- [ ] Headers tested
- [ ] CSP configured
- [ ] Documentation complete
- [ ] Security audit passed
