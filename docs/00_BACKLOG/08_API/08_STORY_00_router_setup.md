# User Story: HTTP Router Setup

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-09

---

## User Story

As a **developer**, I want **a well-structured HTTP router with organized route groups** so that **the application can handle all HTTP requests efficiently and maintainably**.

---

## Acceptance Criteria

- [x] HTTP router configured using chi or gorilla/mux
- [x] Route groups organized by domain (auth, events, invites, rsvp, admin)
- [x] Sub-routers for nested resources
- [x] Route parameters properly extracted
- [x] Method-based routing (GET, POST, PUT, DELETE)
- [x] 404 handler for unknown routes
- [x] 405 handler for unsupported methods
- [x] Router supports middleware attachment
- [x] All routes documented in code
- [x] Route listing available for debugging

---

## Technical Details

### Package Location
- `cmd/server/main.go` - Router initialization
- `internal/handlers/router.go` - Route configuration
- `internal/handlers/router_test.go` - Router tests

### Router Structure

```go
type Router struct {
    mux chi.Router
    handlers *Handlers
}

func NewRouter(handlers *Handlers) *Router {
    r := chi.NewRouter()
    
    // Global middleware
    r.Use(middleware.Recovery)
    r.Use(middleware.Logger)
    
    // Route groups
    r.Route("/auth", func(r chi.Router) {
        // Auth routes
    })
    
    r.Route("/events", func(r chi.Router) {
        // Event routes
    })
    
    return &Router{mux: r, handlers: handlers}
}
```

### Route Groups

```
/auth/*          - Authentication routes
/events/*        - Event management
/invites/*       - Invite management
/rsvp/{token}    - Guest RSVP (no auth)
/admin/*         - Admin functions
/assets/*        - Static assets
/health          - Health check
/metrics         - Metrics endpoint
```

---

## Tasks

### Router Setup
- [x] Choose router library (chi recommended)
- [x] Create router initialization function
- [x] Configure route groups
- [x] Set up sub-routers for nested resources
- [x] Add 404 handler
- [x] Add 405 handler
- [x] Add route parameter extraction helpers

### Route Organization
- [x] Create auth route group
- [x] Create events route group
- [x] Create invites route group
- [x] Create RSVP route group
- [x] Create admin route group
- [x] Create utility routes (health, metrics)
- [x] Create static asset routes

### Testing
- [x] Test route matching
- [x] Test route parameters
- [x] Test 404 handling
- [x] Test 405 handling
- [x] Test route group isolation
- [x] Test middleware attachment
- [x] Integration test full router

### Documentation
- [x] Document route structure
- [x] Document route parameters
- [x] Document route groups
- [x] Add inline route documentation

---

## Dependencies

**Depends on:** None (foundational)

**Blocks:** All other Epic 08 stories

---

## Testing Strategy

### Unit Tests

```go
func TestRouter_RouteMatching(t *testing.T)
func TestRouter_RouteParameters(t *testing.T)
func TestRouter_404Handler(t *testing.T)
func TestRouter_405Handler(t *testing.T)
func TestRouter_RouteGroups(t *testing.T)
```

### Integration Tests

```go
func TestRouter_Integration(t *testing.T) {
    // Test all routes accessible
    // Test middleware chain
    // Test route isolation
}
```

---

## Router Configuration

### Chi Router Example

```go
r := chi.NewRouter()

// Global middleware
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// Auth routes
r.Route("/auth", func(r chi.Router) {
    r.Get("/login", handlers.Login)
    r.Get("/callback", handlers.Callback)
    r.Post("/logout", handlers.Logout)
})

// Protected routes
r.Group(func(r chi.Router) {
    r.Use(authMiddleware)
    
    r.Route("/events", func(r chi.Router) {
        r.Get("/", handlers.ListEvents)
        r.Post("/", handlers.CreateEvent)
        r.Get("/{id}", handlers.GetEvent)
        r.Put("/{id}", handlers.UpdateEvent)
        r.Delete("/{id}", handlers.DeleteEvent)
    })
})

// Public RSVP routes
r.Route("/rsvp/{token}", func(r chi.Router) {
    r.Get("/", handlers.RSVPPage)
    r.Post("/", handlers.SubmitRSVP)
})
```

---

## Route Parameter Extraction

```go
func GetEventID(r *http.Request) (int64, error) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        return 0, ErrInvalidEventID
    }
    return id, nil
}

func GetToken(r *http.Request) string {
    return chi.URLParam(r, "token")
}
```

---

## Error Handlers

### 404 Handler

```go
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    if isAPIRequest(r) {
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Route not found",
        })
    } else {
        renderTemplate(w, "404.html", nil)
    }
}
```

### 405 Handler

```go
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusMethodNotAllowed)
    if isAPIRequest(r) {
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Method not allowed",
        })
    } else {
        renderTemplate(w, "405.html", nil)
    }
}
```

---

## Performance Considerations

- Use compiled route patterns
- Minimize middleware overhead
- Cache route lookups
- Use efficient parameter extraction
- Avoid reflection in hot paths

---

## Security Considerations

- Validate all route parameters
- Sanitize path traversal attempts
- Limit route parameter length
- Log suspicious route access
- Rate limit route discovery attempts

---

## References

- **HLD:** Section 18 (API Routes)
- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Chi Router:** https://github.com/go-chi/chi
- **Gorilla Mux:** https://github.com/gorilla/mux

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Router configured and tested
- [x] All route groups defined
- [x] 404/405 handlers working
- [x] Route parameters extracted correctly
- [x] Unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Documentation complete
- [x] Code reviewed
- [x] No linter warnings
