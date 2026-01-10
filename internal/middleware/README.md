# Middleware Package

## Purpose

This package provides HTTP middleware for TinyRSVP, including core request processing, authentication, and authorization.

## Core Middleware

### Recovery

Catches panics in HTTP handlers and returns a 500 Internal Server Error without leaking stack traces to clients.

**Usage:**
```go
mux.Use(middleware.Recovery)
```

**Behavior:**
- Catches all panics in downstream handlers
- Logs panic message and stack trace to server logs
- Returns 500 Internal Server Error to client
- Does not leak sensitive information to client
- Allows subsequent requests to continue normally

**Performance:** <1µs overhead per request

### RequestID

Generates or uses existing unique request IDs for request tracking and correlation.

**Usage:**
```go
mux.Use(middleware.RequestID)
```

**Behavior:**
- Checks for existing X-Request-ID header
- Generates new 32-character hex ID if not present
- Injects ID into request context
- Sets X-Request-ID response header
- Enables request correlation across logs

**Context Access:**
```go
requestID := middleware.GetRequestID(r.Context())
```

**Performance:** ~1µs overhead per request

### RealIP

Extracts the real client IP address from proxy headers or RemoteAddr.

**Usage:**
```go
mux.Use(middleware.RealIP)
```

**Behavior:**
- Checks X-Real-IP header first
- Falls back to X-Forwarded-For (uses first IP)
- Falls back to RemoteAddr if no headers present
- Trims whitespace from extracted IPs
- Injects IP into request context

**Context Access:**
```go
realIP := middleware.GetRealIP(r.Context())
```

**Performance:** <1µs overhead per request

### Logging

Logs HTTP requests with method, path, status code, duration, and request ID.

**Usage:**
```go
logger := log.New(os.Stdout, "", log.LstdFlags)
mux.Use(middleware.Logging(logger))
```

**Behavior:**
- Captures response status code
- Measures request duration
- Logs after request completes
- Includes request ID from context
- Uses custom responseWriter to track status

**Log Format:**
```
GET /api/events 200 1.234ms abc123def456
```

**Performance:** <1µs overhead per request

### Timeout

Enforces request timeout using context cancellation.

**Usage:**
```go
mux.Use(middleware.Timeout(30 * time.Second))
```

**Behavior:**
- Creates timeout context for request
- Runs handler in goroutine
- Returns 504 Gateway Timeout if exceeded
- Propagates panics from handler
- Cancels context on completion

**Performance:** ~8µs overhead per request

## Middleware Chaining

### Chain Composer

Combines multiple middleware functions in the correct execution order.

**Usage:**
```go
handler := middleware.Chain(
    middleware.Recovery,
    middleware.RequestID,
    middleware.RealIP,
    middleware.Logging(logger),
    middleware.Timeout(30 * time.Second),
)(finalHandler)
```

**Execution Order:**
Middleware executes in the order provided to Chain:
1. Recovery (outermost - catches all panics)
2. RequestID (generates ID early)
3. RealIP (extracts IP early)
4. Logging (logs with ID and IP)
5. Timeout (enforces deadline)
6. Handler (business logic)

Then unwinds in reverse order.

**Performance:** ~5µs total overhead for full chain

### Security Headers

Sets comprehensive security headers on all responses to protect against common web vulnerabilities.

**Usage:**
```go
mux.Use(middleware.SecurityHeaders(nil))
```

**Behavior:**
- Sets Content-Security-Policy (CSP) header
- Sets Strict-Transport-Security (HSTS) header
- Sets X-Content-Type-Options header
- Sets X-Frame-Options header
- Sets X-XSS-Protection header
- Sets Referrer-Policy header
- Sets Permissions-Policy header
- Fully configurable via SecurityHeadersConfig
- Supports CSP report-only mode
- Supports CSP violation reporting

**Default Headers:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**Custom Configuration:**
```go
config := &middleware.SecurityHeadersConfig{
    HSTSMaxAge:            intPtr(63072000),
    HSTSIncludeSubDomains: true,
    HSTSPreload:           true,
    CSPScriptSrc:          []string{"'self'", "'nonce-abc123'"},
    CSPReportURI:          "/api/csp-report",
    XFrameOptions:         "SAMEORIGIN",
}
mux.Use(middleware.SecurityHeaders(config))
```

**Performance:** ~1.7µs overhead per request

### CSP Violation Reporting

Endpoint for receiving and logging Content Security Policy violation reports.

**Usage:**
```go
mux.Handle("/api/csp-report", middleware.CSPReportHandler(logger))
```

**Behavior:**
- Accepts POST requests only
- Validates Content-Type (application/csp-report or application/json)
- Limits request body to 10KB
- Parses CSP violation reports
- Logs violations with request ID and client IP
- Returns 204 No Content on success

**CSP Configuration with Reporting:**
```go
config := &middleware.SecurityHeadersConfig{
    CSPReportURI: "/api/csp-report",
}
```

**Performance:** Minimal overhead, async logging

## Recommended Middleware Order

```
1. Recovery          - Catch panics (must be first)
2. RequestID         - Generate tracking ID
3. RealIP            - Extract client IP
4. Logging           - Log requests (needs ID and IP)
5. Timeout           - Enforce deadlines
6. Security Headers  - Set CSP, HSTS, etc.
7. CSRF              - Token validation (before auth)
8. Rate Limiting     - Per-IP rate limiting (needs RealIP)
9. Authentication    - Session validation
10. RBAC             - Permission checking
11. Handler          - Business logic
```

## Rate Limiting

Protects against abuse and DoS attacks using per-IP rate limiting with sliding window algorithm.

**Usage:**
```go
limiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
    RequestsPerMinute: 100,
    BurstSize:         100,
    WhitelistedIPs:    []string{"10.0.0.1"},
    BlacklistedIPs:    []string{"10.0.0.99"},
})
defer limiter.Stop()

mux.Use(middleware.RateLimit(limiter, middleware.RateLimitConfig{
    AnonymousLimit:      100,
    AuthenticatedLimit:  300,
    AdminLimit:          1000,
}))
```

**Behavior:**
- Per-IP rate limiting with sliding window
- Different limits for anonymous/authenticated/admin users
- Returns 429 Too Many Requests when limit exceeded
- Sets rate limit headers on all responses
- Supports IP whitelist (unlimited requests)
- Supports IP blacklist (all requests denied)
- In-memory storage with automatic cleanup
- Tracks metrics (total/allowed/denied requests)

**Rate Limit Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704931200
```

**429 Response Headers:**
```
Retry-After: 45
```

**Configuration:**
- `RequestsPerMinute`: Base rate limit per IP
- `BurstSize`: Maximum tokens available
- `WindowDuration`: Time window for rate limiting (default: 1 minute)
- `CleanupInterval`: How often to clean expired entries (default: 5 minutes)
- `WhitelistedIPs`: IPs that bypass rate limiting
- `BlacklistedIPs`: IPs that are always denied

**Default Limits:**
- Anonymous: 100 requests/minute
- Authenticated: 300 requests/minute
- Admin: 1000 requests/minute

**Metrics:**
```go
metrics := limiter.GetMetrics()
// metrics.TotalRequests
// metrics.AllowedRequests
// metrics.DeniedRequests
// metrics.ActiveIPs
```

**Performance:** ~2µs overhead per request


## CSRF Protection

Protects against Cross-Site Request Forgery attacks using the double-submit cookie pattern with token rotation.

**Usage:**
```go
mux.Use(middleware.CSRF(32))
```

**Behavior:**
- Generates cryptographically secure tokens per request
- Uses double-submit cookie pattern (cookie + form/header)
- Validates tokens on POST/PUT/DELETE/PATCH requests
- Allows GET/HEAD/OPTIONS/TRACE without validation
- Rotates tokens after each successful validation
- Sets SameSite=Strict cookie attribute
- Supports both form field and header validation
- Header validation takes precedence (for AJAX)

**Token Sources (in order of precedence):**
1. X-CSRF-Token header (for AJAX requests)
2. csrf_token form field (for form submissions)

**Cookie Attributes:**
```
Name: csrf_token
Path: /
SameSite: Strict
HttpOnly: false (JavaScript needs to read it)
Secure: false (set to true in production)
```

**Context Access:**
```go
token := middleware.GetCSRFToken(r.Context())
```

**Template Usage:**
```html
<form method="POST">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
    <!-- form fields -->
</form>
```

**AJAX Usage:**
```javascript
fetch('/api/endpoint', {
    method: 'POST',
    headers: {
        'X-CSRF-Token': getCookie('csrf_token'),
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
});
```

**Security Features:**
- Constant-time token comparison (timing attack resistant)
- Cryptographically secure random token generation
- Token rotation prevents replay attacks
- Double-submit pattern prevents cookie injection
- SameSite=Strict prevents cross-site cookie sending

**Performance:** ~2µs overhead per request


## Authentication & Authorization Middleware

### RequireAuth

Validates session and injects user and session into request context.

**Usage:**
```go
authMiddleware := middleware.RequireAuth(sessionMgr, userService)
mux.Handle("/protected", authMiddleware(handler))
```

**Behavior:**
- Extracts session ID from cookie
- Validates session (not expired)
- Loads user from database
- Refreshes session last accessed time
- Injects user and session into context
- Returns 401 Unauthorized on any failure

### RequireAdmin

Restricts access to admin users only. Must be used after RequireAuth.

**Usage:**
```go
authMiddleware := middleware.RequireAuth(sessionMgr, userService)
adminMiddleware := middleware.RequireAdmin
mux.Handle("/admin", authMiddleware(adminMiddleware(handler)))
```

**Behavior:**
- Checks for user in context
- Verifies user has admin role
- Returns 401 if no user in context
- Returns 403 Forbidden if user is not admin

### RequireEventManager

Allows both admin and event manager users. Must be used after RequireAuth.

**Usage:**
```go
authMiddleware := middleware.RequireAuth(sessionMgr, userService)
managerMiddleware := middleware.RequireEventManager
mux.Handle("/events", authMiddleware(managerMiddleware(handler)))
```

**Behavior:**
- Checks for user in context
- Verifies user has admin or event manager role
- Returns 401 if no user in context
- Returns 403 Forbidden if user lacks required role

## Context Values

The middleware package uses typed context keys for storing values:

```go
const (
    RequestIDKey contextKey = "requestID"
    RealIPKey    contextKey = "realIP"
)
```

Authentication middleware (from auth package) injects:
- `user` - Current authenticated user (*models.User)
- `session` - Current session (*models.Session)

Retrieve in handlers using:
```go
requestID := middleware.GetRequestID(r.Context())
realIP := middleware.GetRealIP(r.Context())
user, ok := auth.UserFromContext(r.Context())
session, ok := auth.SessionFromContext(r.Context())
```

## Error Responses

- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Authenticated but insufficient permissions
- **500 Internal Server Error**: Panic recovered
- **504 Gateway Timeout**: Request exceeded timeout

## Public Endpoints

These endpoints should NOT have auth middleware:
- `/` - Home page
- `/invite/:token` - Guest RSVP page
- `/health` - Health check
- `/readiness` - Readiness check
- `/static/*` - Static assets

## Performance Metrics

Measured on Intel Core Ultra 7 165U:

| Middleware | Overhead | Allocations | Memory |
|------------|----------|-------------|--------|
| Recovery | 188 ns/op | 4 allocs | 208 B |
| RequestID | 1,279 ns/op | 15 allocs | 1,424 B |
| RealIP | 663 ns/op | 8 allocs | 608 B |
| Logging | 306 ns/op | 7 allocs | 272 B |
| Timeout | 1,754 ns/op | 13 allocs | 1,120 B |
| Security Headers | 1,694 ns/op | 15 allocs | 1,152 B |
| **Full Chain** | **~6.5µs** | **~48 allocs** | **~4 KB** |

Full chain overhead: ~6.5µs per request (well within acceptable range)

## Testing

All middleware functions have comprehensive test coverage:
- Unit tests for each middleware
- Integration tests for full chain
- Performance benchmarks
- Middleware ordering verification
- Context propagation tests
- Error handling tests

Run tests:
```bash
go test -timeout 30s ./internal/middleware/...
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./internal/middleware/...
```

## Example: Full Middleware Stack

```go
package main

import (
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/middleware"
)

func main() {
    logger := log.New(os.Stdout, "", log.LstdFlags)
    
    handler := middleware.Chain(
        middleware.Recovery,
        middleware.RequestID,
        middleware.RealIP,
        middleware.Logging(logger),
        middleware.Timeout(30 * time.Second),
    )(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := middleware.GetRequestID(r.Context())
        realIP := middleware.GetRealIP(r.Context())
        
        logger.Printf("Processing request %s from %s", requestID, realIP)
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    }))
    
    http.ListenAndServe(":8080", handler)
}
```

## Security Considerations

- Recovery middleware does not leak stack traces to clients
- Request IDs are validated for length and format
- Real IP extraction respects proxy headers
- Logging sanitizes sensitive data
- Timeout prevents resource exhaustion
- All middleware designed to fail safely
