# Domain 8: API & HTTP Handlers - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 18 - API Routes](../02_REVISED_HLD.md#18-api-routes), [Section 19 - Request Flow](../02_REVISED_HLD.md#19-request-flow)

---

## 1. Overview

### 1.1 Purpose

Orchestrates all domains, provides HTTP API, handles routing, middleware, input validation, error formatting, and security headers.

### 1.2 Responsibilities

- HTTP router configuration
- Request/response handling
- Input validation and sanitization
- Error response formatting
- CSRF protection
- Security headers (CSP, HSTS, etc.)
- Rate limiting
- Health check endpoint
- Metrics endpoint (Prometheus)
- All 50+ API routes

### 1.3 Design Principles

- **Middleware Chain** - Composable request processing
- **Fail Fast** - Validate early, return errors immediately
- **Standardized Errors** - Consistent error format
- **Secure by Default** - Security headers on all responses
- **Observable** - Metrics on all endpoints

---

## 2. Package Structure

```
internal/
├── handlers/
│   ├── auth.go                 # Auth endpoints
│   ├── auth_test.go
│   ├── events.go               # Event endpoints
│   ├── events_test.go
│   ├── invites.go              # Invite endpoints
│   ├── invites_test.go
│   ├── rsvp.go                 # RSVP endpoints
│   ├── rsvp_test.go
│   ├── templates.go            # Template endpoints
│   ├── templates_test.go
│   ├── health.go               # Health check
│   ├── health_test.go
│   └── errors.go               # Error handling
│       └── errors_test.go
├── middleware/
│   ├── auth.go                 # Auth middleware
│   ├── session.go              # Session middleware
│   ├── rbac.go                 # RBAC middleware
│   ├── csrf.go                 # CSRF protection
│   ├── security.go             # Security headers
│   ├── ratelimit.go            # Rate limiting
│   ├── logging.go              # Request logging
│   └── recovery.go             # Panic recovery
cmd/
└── server/
    └── main.go                 # Application entrypoint
```

---

## 3. Interfaces

### 3.1 Handler Interface

```go
package handlers

import "net/http"

type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

### 3.2 Middleware Interface

```go
package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler
```

---

## 4. Implementation

### 4.1 Router Setup

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
)

func setupRouter(deps *Dependencies) http.Handler {
    r := chi.NewRouter()
    
    r.Use(middleware.Recovery())
    r.Use(middleware.RequestLogger())
    r.Use(middleware.SecurityHeaders())
    r.Use(middleware.RateLimit(100, time.Minute))
    
    r.Get("/health", handlers.NewHealthHandler(deps.DB).ServeHTTP)
    r.Get("/metrics", handlers.NewMetricsHandler().ServeHTTP)
    
    r.Route("/auth", func(r chi.Router) {
        r.Get("/login", handlers.NewLoginHandler(deps.Auth).ServeHTTP)
        r.Get("/callback", handlers.NewCallbackHandler(deps.Auth, deps.UserService, deps.SessionMgr).ServeHTTP)
        r.Post("/logout", handlers.NewLogoutHandler(deps.Auth).ServeHTTP)
    })
    
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth(deps.SessionMgr, deps.UserService))
        r.Use(middleware.CSRF())
        
        r.Route("/events", func(r chi.Router) {
            r.Get("/", handlers.NewListEventsHandler(deps.EventService).ServeHTTP)
            r.Post("/", handlers.NewCreateEventHandler(deps.EventService).ServeHTTP)
            r.Get("/{id}", handlers.NewGetEventHandler(deps.EventService).ServeHTTP)
            r.Put("/{id}", handlers.NewUpdateEventHandler(deps.EventService).ServeHTTP)
            r.Delete("/{id}", handlers.NewDeleteEventHandler(deps.EventService).ServeHTTP)
        })
        
        r.Route("/invites", func(r chi.Router) {
            r.Post("/", handlers.NewCreateInviteHandler(deps.InviteService).ServeHTTP)
            r.Post("/bulk", handlers.NewBulkInviteHandler(deps.InviteService).ServeHTTP)
        })
    })
    
    r.Route("/rsvp/{token}", func(r chi.Router) {
        r.Get("/", handlers.NewRSVPPageHandler(deps.RSVPService, deps.InviteService).ServeHTTP)
        r.Post("/", handlers.NewSubmitRSVPHandler(deps.RSVPService).ServeHTTP)
    })
    
    r.Get("/unsubscribe/{token}", handlers.NewUnsubscribeHandler(deps.InviteService).ServeHTTP)
    
    return r
}
```

### 4.2 Background Jobs

```go
package main

import (
    "context"
    "time"
)

type BackgroundJobs struct {
    emailProcessor *email.QueueProcessor
    sessionMgr     auth.SessionManager
    inviteService  invites.Service
    eventService   events.Service
    auditRepo      repositories.AuditLogRepository
    stopChan       chan struct{}
}

func NewBackgroundJobs(deps *Dependencies) *BackgroundJobs {
    return &BackgroundJobs{
        emailProcessor: email.NewQueueProcessor(deps.EmailQueueRepo, deps.SMTPSender, 50),
        sessionMgr:     deps.SessionMgr,
        inviteService:  deps.InviteService,
        eventService:   deps.EventService,
        auditRepo:      deps.AuditLogRepo,
        stopChan:       make(chan struct{}),
    }
}

func (j *BackgroundJobs) Start() {
    j.emailProcessor.Start()
    
    go j.runSessionCleanup()
    go j.runTokenCleanup()
    go j.runEventArchive()
    go j.runAuditLogCleanup()
}

func (j *BackgroundJobs) Stop() {
    close(j.stopChan)
    j.emailProcessor.Stop()
}

func (j *BackgroundJobs) runSessionCleanup() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := context.Background()
            count, err := j.sessionMgr.CleanupExpired(ctx)
            if err != nil {
                log.Printf("Session cleanup error: %v", err)
            } else {
                log.Printf("Cleaned up %d expired sessions", count)
            }
        case <-j.stopChan:
            return
        }
    }
}

func (j *BackgroundJobs) runTokenCleanup() {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := context.Background()
            if err := j.inviteService.CleanupExpiredTokens(ctx); err != nil {
                log.Printf("Token cleanup error: %v", err)
            }
        case <-j.stopChan:
            return
        }
    }
}

func (j *BackgroundJobs) runEventArchive() {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := context.Background()
            events, err := j.eventService.GetEventsToArchive(ctx)
            if err != nil {
                log.Printf("Event archive query error: %v", err)
                continue
            }
            
            for _, event := range events {
                if err := j.eventService.ArchiveEvent(ctx, event.ID); err != nil {
                    log.Printf("Failed to archive event %d: %v", event.ID, err)
                }
            }
        case <-j.stopChan:
            return
        }
    }
}

func (j *BackgroundJobs) runAuditLogCleanup() {
    ticker := time.NewTicker(7 * 24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ctx := context.Background()
            cutoff := time.Now().AddDate(-1, 0, 0)
            count, err := j.auditRepo.DeleteOlderThan(ctx, cutoff)
            if err != nil {
                log.Printf("Audit log cleanup error: %v", err)
            } else {
                log.Printf("Cleaned up %d old audit logs", count)
            }
        case <-j.stopChan:
            return
        }
    }
}
```

### 4.2 Auth Middleware

```go
package middleware

import (
    "context"
    "net/http"
)

func RequireAuth(sessionMgr SessionManager, userService UserService) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sessionID, err := sessionMgr.GetSessionFromRequest(r)
            if err != nil {
                http.Redirect(w, r, "/auth/login", http.StatusFound)
                return
            }
            
            session, err := sessionMgr.GetSession(r.Context(), sessionID)
            if err != nil {
                sessionMgr.ClearSessionCookie(w)
                http.Redirect(w, r, "/auth/login", http.StatusFound)
                return
            }
            
            user, err := userService.GetUserByID(r.Context(), session.UserID)
            if err != nil {
                http.Error(w, "User not found", http.StatusUnauthorized)
                return
            }
            
            ctx := auth.WithUser(r.Context(), user)
            ctx = auth.WithSession(ctx, session)
            
            sessionMgr.RefreshSession(ctx, sessionID)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 4.3 RBAC Middleware

```go
package middleware

import (
    "net/http"
)

func RequireAdmin() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, ok := auth.UserFromContext(r.Context())
            if !ok || !user.IsAdmin() {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func RequireEventManager() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, ok := auth.UserFromContext(r.Context())
            if !ok || !user.IsEventManager() {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 4.4 Security Headers Middleware

```go
package middleware

import "net/http"

func SecurityHeaders() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            w.Header().Set("X-Content-Type-Options", "nosniff")
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
            w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'")
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 4.5 Error Response

```go
package handlers

import (
    "encoding/json"
    "net/http"
)

type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Field   string                 `json:"field,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func WriteError(w http.ResponseWriter, code string, message string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteStatus(status)
    
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: ErrorDetail{
            Code:    code,
            Message: message,
        },
    })
}

func WriteValidationError(w http.ResponseWriter, field, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteStatus(http.StatusBadRequest)
    
    json.NewEncoder(w).Encode(ErrorResponse{
        Error: ErrorDetail{
            Code:    "VALIDATION_ERROR",
            Message: message,
            Field:   field,
        },
    })
}
```

---

## 5. Request Flow

```
HTTP Request
    ↓
Recovery Middleware (panic handling)
    ↓
Logging Middleware (request logging)
    ↓
Security Headers Middleware (CSP, HSTS, etc.)
    ↓
Rate Limit Middleware (per-IP limiting)
    ↓
Auth Middleware (session validation)
    ↓
RBAC Middleware (permission checking)
    ↓
CSRF Middleware (token validation)
    ↓
Handler (business logic)
    ↓
Response
```

---

## 6. API Routes

### 6.1 Auth Routes

- `GET /auth/login` - Initiate login
- `GET /auth/callback` - OIDC callback
- `POST /auth/logout` - Logout

### 6.2 Event Routes

- `GET /events` - List events
- `POST /events` - Create event
- `GET /events/{id}` - Get event
- `PUT /events/{id}` - Update event
- `DELETE /events/{id}` - Delete event
- `POST /events/{id}/publish` - Publish event
- `POST /events/{id}/cancel` - Cancel event

### 6.3 Invite Routes

- `GET /events/{id}/invites` - List invites
- `POST /events/{id}/invites` - Create invite
- `POST /events/{id}/invites/bulk` - Bulk create
- `POST /events/{id}/invites/import` - CSV import
- `PUT /invites/{id}` - Update invite
- `DELETE /invites/{id}` - Delete invite
- `POST /invites/{id}/revoke` - Revoke invite
- `POST /invites/{id}/regenerate` - Regenerate token
- `POST /invites/{id}/send` - Send invite email

### 6.4 RSVP Routes

- `GET /rsvp/{token}` - RSVP page
- `POST /rsvp/{token}` - Submit RSVP
- `GET /unsubscribe/{token}` - Unsubscribe from reminder emails

### 6.5 Admin Routes

- `GET /admin/users` - List users
- `POST /admin/users/{id}/role` - Update role
- `GET /admin/settings` - System settings
- `POST /admin/smtp/test` - Test SMTP

### 6.6 Utility Routes

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

---

## 7. Dependencies

**All Domains:**
- Domain 1 (Auth) - Authentication
- Domain 2 (Event) - Event operations
- Domain 3 (Invite) - Invite operations
- Domain 4 (RSVP) - RSVP operations
- Domain 5 (Email) - Email operations
- Domain 6 (Template) - Template operations
- Domain 7 (Database) - Data persistence

---

## 8. Testing

```go
func TestEventHandler_CreateEvent(t *testing.T) {
    handler := NewCreateEventHandler(mockEventService)
    
    tests := []struct {
        name       string
        body       string
        user       *models.User
        wantStatus int
    }{
        {
            name: "valid event",
            body: `{"title":"Test","start_time":"2026-06-15T18:00:00Z","timezone":"America/Los_Angeles"}`,
            user: &models.User{Role: models.RoleEventManager},
            wantStatus: http.StatusCreated,
        },
        {
            name: "unauthorized",
            body: `{"title":"Test"}`,
            user: nil,
            wantStatus: http.StatusUnauthorized,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/events", strings.NewReader(tt.body))
            if tt.user != nil {
                req = req.WithContext(auth.WithUser(req.Context(), tt.user))
            }
            
            w := httptest.NewRecorder()
            handler.ServeHTTP(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
            }
        })
    }
}
```

---

## 9. Health Check Implementation

```go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/yourusername/tinyrsvp/internal/db"
)

type HealthHandler struct {
    db db.Database
}

func NewHealthHandler(database db.Database) *HealthHandler {
    return &HealthHandler{db: database}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    checks := make(map[string]string)
    allHealthy := true
    
    if err := h.db.Ping(ctx); err != nil {
        checks["database"] = "unhealthy: " + err.Error()
        allHealthy = false
    } else {
        checks["database"] = "ok"
    }
    
    _, err := h.db.Exec(ctx, "SELECT 1")
    if err != nil {
        checks["database_write"] = "unhealthy: " + err.Error()
        allHealthy = false
    } else {
        checks["database_write"] = "ok"
    }
    
    status := "healthy"
    statusCode := http.StatusOK
    if !allHealthy {
        status = "unhealthy"
        statusCode = http.StatusServiceUnavailable
    }
    
    response := map[string]interface{}{
        "status":  status,
        "checks":  checks,
        "version": "0.1.0",
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

---

## 10. Metrics Implementation

```go
package handlers

import (
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    eventsTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "tinyrsvp_events_total",
            Help: "Total number of events by status",
        },
        []string{"status"},
    )
    
    invitesTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "tinyrsvp_invites_total",
            Help: "Total number of invites by status",
        },
        []string{"status"},
    )
    
    rsvpsTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "tinyrsvp_rsvps_total",
            Help: "Total number of RSVPs by response",
        },
        []string{"response"},
    )
    
    emailsTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "tinyrsvp_emails_total",
            Help: "Total number of emails by status",
        },
        []string{"status"},
    )
    
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tinyrsvp_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "tinyrsvp_http_request_duration_seconds",
            Help:    "HTTP request latency in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    dbConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "tinyrsvp_db_connections",
            Help: "Number of database connections",
        },
    )
    
    emailQueueSize = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "tinyrsvp_email_queue_size",
            Help: "Number of pending emails in queue",
        },
    )
)

func init() {
    prometheus.MustRegister(eventsTotal)
    prometheus.MustRegister(invitesTotal)
    prometheus.MustRegister(rsvpsTotal)
    prometheus.MustRegister(emailsTotal)
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(dbConnections)
    prometheus.MustRegister(emailQueueSize)
}

type MetricsHandler struct {
    handler http.Handler
}

func NewMetricsHandler() *MetricsHandler {
    return &MetricsHandler{
        handler: promhttp.Handler(),
    }
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.handler.ServeHTTP(w, r)
}
```

---

## 11. Security

**CSRF Protection:** Token in form/header
**Security Headers:** CSP, HSTS, X-Frame-Options
**Rate Limiting:** 100 requests/minute per IP
**Input Validation:** Sanitize all inputs
**Error Messages:** User-friendly, no sensitive data

---

## 12. Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| VALIDATION_ERROR | 400 | Input validation failed |
| UNAUTHORIZED | 401 | Authentication required |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Concurrent modification |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server error |
| SERVICE_UNAVAILABLE | 503 | Service unavailable |

---

## 11. Dependencies

**External:**
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/prometheus/client_golang` - Metrics

**Internal:** All domains (orchestration layer)

---

**Document Status:** ✅ Complete

**All LLD Documents:** Complete
