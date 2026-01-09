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

## Recommended Middleware Order

```
1. Recovery          - Catch panics (must be first)
2. RequestID         - Generate tracking ID
3. RealIP            - Extract client IP
4. Logging           - Log requests (needs ID and IP)
5. Timeout           - Enforce deadlines
6. Security Headers  - Set CSP, HSTS, etc.
7. Rate Limiting     - Per-IP rate limiting (needs RealIP)
8. Authentication    - Session validation
9. RBAC              - Permission checking
10. CSRF             - Token validation
11. Handler          - Business logic
```

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
| **Full Chain** | **4,798 ns/op** | **33 allocs** | **2,824 B** |

Full chain overhead: ~5µs per request (well within acceptable range)

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
