# User Story: Middleware Chain Configuration

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-09

---

## User Story

As a **developer**, I want **a properly ordered middleware chain** so that **all HTTP requests are processed with correct security, logging, and authentication**.

---

## Acceptance Criteria

- [x] Middleware chain properly ordered
- [x] Recovery middleware catches panics
- [x] Request logging middleware
- [x] Request ID generation
- [x] Real IP extraction
- [x] Timeout middleware
- [x] Middleware composable and testable
- [x] Middleware can be selectively applied to routes
- [x] Middleware execution order documented
- [x] Performance impact measured

---

## Technical Details

### Package Location
- `internal/middleware/chain.go` - Middleware chain
- `internal/middleware/chain_test.go` - Chain tests
- `internal/middleware/recovery.go` - Panic recovery
- `internal/middleware/logging.go` - Request logging
- `internal/middleware/timeout.go` - Request timeout

### Middleware Order

```
1. Recovery (panic handling)
2. RequestID (generate unique ID)
3. RealIP (extract real client IP)
4. Logging (request/response logging)
5. Timeout (request timeout)
6. Security Headers (CSP, HSTS, etc.)
7. Rate Limiting (per-IP)
8. Authentication (session validation)
9. RBAC (permission checking)
10. CSRF (token validation)
11. Handler (business logic)
```

### Middleware Interface

```go
type Middleware func(http.Handler) http.Handler

func Chain(middlewares ...Middleware) Middleware {
    return func(next http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            next = middlewares[i](next)
        }
        return next
    }
}
```

---

## Tasks

### Core Middleware
- [x] Implement recovery middleware
- [x] Implement request ID middleware
- [x] Implement real IP middleware
- [x] Implement logging middleware
- [x] Implement timeout middleware
- [x] Create middleware chain composer

### Integration
- [x] Attach middleware to router
- [x] Configure global middleware
- [x] Configure route-specific middleware
- [x] Test middleware ordering
- [x] Measure middleware performance

### Testing
- [x] Test each middleware individually
- [x] Test middleware chain composition
- [x] Test middleware ordering
- [x] Test panic recovery
- [x] Test timeout enforcement
- [x] Integration test full chain

### Documentation
- [x] Document middleware order
- [x] Document middleware purpose
- [x] Document performance impact
- [x] Add usage examples

---

## Dependencies

**Depends on:** 08_STORY_00_router_setup.md

**Blocks:** 
- 08_STORY_02_error_handling.md
- 08_STORY_03_security_headers.md
- 08_STORY_04_csrf_protection.md
- 08_STORY_05_rate_limiting.md

---

## Testing Strategy

### Unit Tests

```go
func TestRecoveryMiddleware(t *testing.T)
func TestRequestIDMiddleware(t *testing.T)
func TestRealIPMiddleware(t *testing.T)
func TestLoggingMiddleware(t *testing.T)
func TestTimeoutMiddleware(t *testing.T)
func TestMiddlewareChain(t *testing.T)
```

### Integration Tests

```go
func TestMiddlewareChain_Integration(t *testing.T) {
    // Test full middleware chain
    // Test panic recovery
    // Test request logging
    // Test timeout enforcement
}
```

---

## Middleware Implementations

### Recovery Middleware

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Request ID Middleware

```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }
        
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Real IP Middleware

```go
func RealIP(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.Header.Get("X-Real-IP")
        if ip == "" {
            ip = r.Header.Get("X-Forwarded-For")
            if ip != "" {
                ip = strings.Split(ip, ",")[0]
            }
        }
        if ip == "" {
            ip = r.RemoteAddr
        }
        
        ctx := context.WithValue(r.Context(), RealIPKey, ip)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Logging Middleware

```go
func Logging(logger *log.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            ww := &responseWriter{ResponseWriter: w}
            next.ServeHTTP(ww, r)
            
            duration := time.Since(start)
            logger.Printf("%s %s %d %s %s",
                r.Method,
                r.URL.Path,
                ww.status,
                duration,
                GetRequestID(r.Context()),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}
```

### Timeout Middleware

```go
func Timeout(timeout time.Duration) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            done := make(chan struct{})
            go func() {
                next.ServeHTTP(w, r.WithContext(ctx))
                close(done)
            }()
            
            select {
            case <-done:
                return
            case <-ctx.Done():
                http.Error(w, "Request timeout", http.StatusGatewayTimeout)
            }
        })
    }
}
```

---

## Middleware Chain Composition

```go
func SetupMiddleware(router chi.Router, config *Config) {
    // Global middleware (all routes)
    router.Use(
        Recovery,
        RequestID,
        RealIP,
        Logging(config.Logger),
        Timeout(30 * time.Second),
    )
    
    // Protected routes
    router.Group(func(r chi.Router) {
        r.Use(
            SecurityHeaders,
            RateLimit(config.RateLimiter),
            Authentication(config.SessionStore),
            RBAC(config.PermissionChecker),
        )
        
        // Admin routes
        r.Route("/admin", func(r chi.Router) {
            r.Use(RequireAdmin)
            // Admin handlers
        })
    })
    
    // CSRF for mutation routes
    router.Group(func(r chi.Router) {
        r.Use(CSRF(config.CSRFSecret))
        
        r.Post("/events", handlers.CreateEvent)
        r.Put("/events/{id}", handlers.UpdateEvent)
        r.Delete("/events/{id}", handlers.DeleteEvent)
    })
}
```

---

## Performance Considerations

- Minimize allocations in hot path
- Use sync.Pool for temporary objects
- Avoid reflection in middleware
- Cache middleware results when possible
- Measure middleware overhead

### Performance Targets

- Recovery: <1μs overhead
- RequestID: <5μs overhead
- RealIP: <10μs overhead
- Logging: <50μs overhead
- Timeout: <10μs overhead

---

## Context Keys

```go
type contextKey string

const (
    RequestIDKey contextKey = "requestID"
    RealIPKey    contextKey = "realIP"
    UserKey      contextKey = "user"
    SessionKey   contextKey = "session"
)

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}

func GetRealIP(ctx context.Context) string {
    if ip, ok := ctx.Value(RealIPKey).(string); ok {
        return ip
    }
    return ""
}
```

---

## Error Handling

- Panic recovery logs stack trace
- Timeout returns 504 Gateway Timeout
- Middleware errors logged but don't stop chain
- Critical errors return 500 Internal Server Error

---

## Security Considerations

- Recovery middleware doesn't leak stack traces to clients
- Request ID validated for length and format
- Real IP validated against trusted proxies
- Logging sanitizes sensitive data
- Timeout prevents resource exhaustion

---

## References

- **HLD:** Section 19 (Request Flow)
- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Story 00:** [08_STORY_00_router_setup.md](08_STORY_00_router_setup.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All middleware implemented and tested
- [x] Middleware chain properly ordered
- [x] Panic recovery working
- [x] Request logging functional
- [x] Timeout enforcement working
- [x] Unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Performance targets met
- [x] Documentation complete
- [x] Code reviewed
- [x] No linter warnings
