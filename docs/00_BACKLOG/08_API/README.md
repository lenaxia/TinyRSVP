# Epic: API & HTTP Layer

**Priority:** High
**Status:** ⚠️ Mostly Complete — 1 critical security issue, 1 dead feature  
**Target Version:** v0
**Completed:** 2026-01-10
**Confidence:** MEDIUM (80%)
**Test Pass Rate:** 100% (all tests passing as of 2026-03-05)
**Production Ready:** No — X-Test-User-ID bypass must be removed before public deployment

---

## Current Status

**Last Updated:** 2026-04-06

All routes implemented and tests pass. Two issues found during code-level validation:

1. **`X-Test-User-ID` auth bypass is active in production** (Critical — see Known Issues)
2. **Template editor routes are registered in docs but never registered in router** (Dead code)

---

## Known Issues (Code-Verified)

**ISSUE-1 (Critical): X-Test-User-ID authentication bypass**
- `internal/middleware/rbac.go:16`: `if testUserID := r.Header.Get("X-Test-User-ID"); testUserID != "" {`
- No build tag, no env flag, no conditional. Any HTTP request with this header bypasses session auth and impersonates any user by ID.
- Used by `tests/ux/server_test.go` for browser tests. Should be gated by a build tag (`//go:build testing`) or env flag.
- Must be removed or gated before public deployment.

**ISSUE-2: Template editor routes unreachable**
- `internal/handlers/template_editor.go` implements full component editor UI (`/templates/{id}/edit`, `/api/templates/{id}/components`).
- `NewTemplateEditorHandlers` is never called in `cmd/server/main.go`.
- Routes are never registered in `router.go`.
- Feature is completely dead. Route table in docs is aspirational for this path.

---



## Overview

Implement complete HTTP API layer that orchestrates all domains. Includes routing, middleware chain, request/response handling, error formatting, CSRF protection, security headers, and rate limiting.

**Goal:** Wire all components together into a cohesive web application with proper HTTP handling, security, and user experience.

---

## Success Criteria

- [x] All 50+ API routes implemented
- [x] Middleware chain properly ordered
- [x] CSRF protection on all mutations
- [x] Security headers on all responses
- [x] Rate limiting per IP address
- [x] Input validation and sanitization
- [x] Error responses user-friendly
- [x] Health check and metrics endpoints
- [x] Static asset serving
- [x] Mobile-responsive web UI

**Status:** All criteria met. 100% test pass rate.

---

## User Stories

### Phase 1: HTTP Infrastructure
- [x] [`08_STORY_00_router_setup.md`](08_STORY_router_setup.md) - HTTP router configuration
- [x] [`08_STORY_01_middleware_chain.md`](08_STORY_middleware_chain.md) - Middleware ordering and composition
- [x] [`08_STORY_02_error_handling.md`](08_STORY_error_handling.md) - Error response formatting

### Phase 2: Security Middleware
- [x] [`08_STORY_03_security_headers.md`](08_STORY_security_headers.md) - CSP, HSTS, X-Frame-Options
- [x] [`08_STORY_04_csrf_protection.md`](08_STORY_csrf_protection.md) - CSRF token generation and validation
- [x] [`08_STORY_05_rate_limiting.md`](08_STORY_rate_limiting.md) - Per-IP rate limiting

### Phase 3: Authentication Routes
- [x] [`08_STORY_06_login_routes.md`](08_STORY_login_routes.md) - Login, logout, callback
- [x] [`08_STORY_07_dashboard_route.md`](08_STORY_dashboard_route.md) - Main dashboard

### Phase 4: Event Routes
- [x] [`08_STORY_08_event_routes.md`](08_STORY_event_routes.md) - Event CRUD endpoints
- [x] [`08_STORY_09_event_ui.md`](08_STORY_event_ui.md) - Event management UI

### Phase 5: Invite Routes
- [x] [`08_STORY_10_invite_routes.md`](08_STORY_invite_routes.md) - Invite CRUD endpoints
- [x] [`08_STORY_11_invite_ui.md`](08_STORY_invite_ui.md) - Invite management UI
- [x] [`08_STORY_12_csv_upload_route.md`](08_STORY_csv_upload_route.md) - CSV import endpoint (integrated in Story 10)

### Phase 6: RSVP Routes
- [x] [`08_STORY_13_rsvp_routes.md`](08_STORY_rsvp_routes.md) - RSVP submission endpoints
- [x] [`08_STORY_14_rsvp_ui.md`](08_STORY_rsvp_ui.md) - Guest RSVP page

### Phase 7: Admin Routes
- [x] [`08_STORY_15_admin_routes.md`](08_STORY_admin_routes.md) - User management, settings
- [x] [`08_STORY_16_admin_ui.md`](08_STORY_admin_ui.md) - Admin dashboard

### Phase 8: Utility Routes
- [x] [`08_STORY_17_health_metrics.md`](08_STORY_health_metrics.md) - Health check and metrics
- [x] [`08_STORY_18_static_assets.md`](08_STORY_static_assets.md) - Static file serving

---

## Dependencies

**Depends on:** All other epics (orchestration layer)  
**Blocks:** None (final integration)

---

## Technical Overview

### Middleware Chain Order

```
1. Recovery (panic handling)
2. Logging (request logging)
3. Security Headers (CSP, HSTS, etc.)
4. Rate Limiting (per-IP)
5. Authentication (session validation)
6. RBAC (permission checking)
7. CSRF (token validation)
8. Handler (business logic)
```

### Route Categories

```
/auth/*          - Authentication (login, logout, callback)
/events/*        - Event management (CRUD)
/invites/*       - Invite management (CRUD, CSV)
/rsvp/{token}    - Guest RSVP (no auth required)
/admin/*         - Admin functions (user mgmt, settings)
/assets/*        - Static assets (images, CSS, JS)
/health          - Health check
/metrics         - Prometheus metrics
```

### Request Flow

```
HTTP Request
     ↓
Middleware Chain
     ↓
Route Handler
     ↓
Service Layer
     ↓
Repository Layer
     ↓
Database
     ↓
Response
```

---

## Technical Decisions

### Router: net/http with chi/mux
- Standard library compatible
- Middleware support
- Route parameters
- Sub-routers for organization

### Error Responses
- JSON for API endpoints
- HTML for web pages
- User-friendly messages
- Field-specific validation errors

### CSRF Protection
- Token per session
- Hidden form field
- Header validation for AJAX
- SameSite cookies as backup

### Rate Limiting
- Per IP address
- Sliding window algorithm
- Configurable limits
- 429 status with Retry-After

---

## API Routes

### Authentication Routes
```
GET  /login                  - Show login page
GET  /auth/oidc/login        - Redirect to OIDC provider
GET  /auth/oidc/callback     - OIDC callback handler
POST /logout                 - Logout and clear session
```

### Event Routes
```
GET    /events               - List events
GET    /events/new           - New event form
POST   /events               - Create event
GET    /events/{id}          - View event details
GET    /events/{id}/edit     - Edit event form
PUT    /events/{id}          - Update event
DELETE /events/{id}          - Delete (archive) event
POST   /events/{id}/publish  - Publish draft event
POST   /events/{id}/cancel   - Cancel event
```

### Invite Routes
```
GET    /events/{id}/invites           - List invites
GET    /events/{id}/invites/new       - New invite form
POST   /events/{id}/invites           - Create invite
POST   /events/{id}/invites/bulk      - CSV import
GET    /invites/{id}                  - View invite
PUT    /invites/{id}                  - Update invite
DELETE /invites/{id}                  - Delete invite
POST   /invites/{id}/revoke           - Revoke token
POST   /invites/{id}/regenerate       - Regenerate token
POST   /invites/{id}/send             - Send invite email
```

### RSVP Routes (No Auth)
```
GET  /rsvp/{token}           - RSVP page
POST /rsvp/{token}           - Submit RSVP
GET  /rsvp/{token}/confirm   - Confirmation page
GET  /unsubscribe/{token}    - Unsubscribe from reminders
```

### Admin Routes
```
GET    /admin                - Admin dashboard
GET    /admin/users          - List users
POST   /admin/users          - Create user
PUT    /admin/users/{id}     - Update user
DELETE /admin/users/{id}     - Delete user
GET    /admin/settings       - System settings
PUT    /admin/settings       - Update settings
```

### Utility Routes
```
GET /health                  - Health check
GET /metrics                 - Prometheus metrics
GET /assets/*                - Static assets
```

---

## Security Headers

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'
```

---

## Rate Limiting

### Limits
- Anonymous: 100 requests/minute
- Authenticated: 300 requests/minute
- Admin: 1000 requests/minute

### Implementation
- Leaky Bucket
- Per IP address tracking
- Redis-free (in-memory in v0)
- Configurable limits

---

## References

- **HLD:** Section 18 (API Routes), Section 19 (Request Flow)
- **LLD:** [`lld/08_API_LLD.md`](../lld/08_API_LLD.md)
- **Security:** Section 16 (Security)

---

## Testing Strategy

### Unit Tests
- Middleware functions
- Route handlers
- Error formatting
- CSRF token generation
- Rate limiting logic

### Integration Tests
- Full request flows
- Middleware chain
- Authentication enforcement
- Permission checking
- Error handling

### End-to-End Tests
- Complete user workflows
- Event creation to RSVP
- CSV import to email send
- Login to logout

### Security Tests
- CSRF protection
- XSS prevention
- SQL injection prevention
- Rate limit enforcement
- Security header validation

---

## UI Components

### Admin Dashboard
- Event list with status
- Quick stats (total events, RSVPs)
- Recent activity
- Navigation menu

### Event Management
- Event form (create/edit)
- Event list with filters
- Event details with RSVP summary
- Invite management

### Guest RSVP Page
- Event details
- Response selection
- Plus ones input
- Preference questions
- Submit button

### Confirmation Page
- Thank you message
- RSVP summary
- Calendar download
- Update link

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| CSRF attacks | High | Token validation, SameSite cookies |
| Rate limit bypass | Medium | Multiple tracking methods, IP validation |
| XSS vulnerabilities | High | Template auto-escaping, input sanitization |
| Broken middleware chain | High | Comprehensive integration tests |
| Poor mobile experience | Medium | Mobile-first design, responsive testing |

---

## Performance Targets

- Page load: <2 seconds
- API response: <500ms
- Static assets: <100ms
- Health check: <50ms
- Metrics: <100ms

---

## Definition of Done

- [x] All user stories complete
- [x] All 50+ routes implemented
- [x] Middleware chain functional
- [x] CSRF protection working
- [x] Security headers set
- [x] Rate limiting enforced
- [x] All UIs mobile-responsive
- [x] Error handling comprehensive
- [x] Health and metrics working
- [x] All tests passing (100%)
- [x] Security review passed
- [x] Performance targets met
- [x] Documentation updated
