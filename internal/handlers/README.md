# Handlers Package

## Purpose

HTTP request handlers for TinyRSVP application endpoints.

## Rules

- All handlers implement `http.Handler` interface
- Handlers are stateless and thread-safe
- Use strongly-typed request/response structs
- Return proper HTTP status codes
- Set appropriate Content-Type headers
- Use context for timeouts and cancellation

## Structure

- `auth.go` - Authentication route handlers
- `auth_test.go` - Authentication handler unit tests
- `auth_integration_test.go` - Authentication flow integration tests
- `health.go` - Health check endpoint (liveness probe)
- `health_test.go` - Health check tests
- `readiness.go` - Readiness check endpoint (readiness probe)
- `readiness_test.go` - Readiness check tests
- `router.go` - Router types, interfaces, and `NewRouter` orchestrator
- `router_setup.go` - Route registration functions (middleware, auth, pages, API, RSVP, static)
- `router_test.go` - Router tests
- `router_auth_test.go` - Router authentication integration tests
- `errors.go` - Centralized error handling

## Health Check Endpoints

### `/health` - Liveness Probe

**Purpose:** Determine if the application is alive and running.

**Response:**
- `200 OK` - Application is alive
- Returns basic application status and version

**Usage:**
```go
healthHandler := handlers.NewHealthHandler("0.1.0")
http.Handle("/health", healthHandler)
```

**Kubernetes Integration:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
```

### `/ready` - Readiness Probe

**Purpose:** Determine if the application can serve traffic.

**Checks:**
- Database connectivity
- Migration status

**Response:**
- `200 OK` - Application is ready to serve traffic
- `503 Service Unavailable` - Application is not ready

**Usage:**
```go
readinessHandler := handlers.NewReadinessHandler("0.1.0", database, migrator)
http.Handle("/ready", readinessHandler)
```

**Kubernetes Integration:**
```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Response Format

All health endpoints return JSON:

```json
{
  "status": "healthy",
  "timestamp": "2026-01-06T19:00:00Z",
  "version": "0.1.0",
  "checks": {
    "database": {
      "status": "healthy",
      "message": "Connected",
      "latency_ms": 2
    },
    "migrations": {
      "status": "healthy",
      "message": "Up to date",
      "version": 2
    }
  }
}
```

## Status Values

- `healthy` - Component is functioning normally
- `degraded` - Component is functioning but with issues (not currently used)
- `unhealthy` - Component is not functioning

## Testing

Run tests with timeout:
```bash
go test -timeout 30s ./internal/handlers/...
```

Run with coverage:
```bash
go test -timeout 30s -cover ./internal/handlers/...
```

## Authentication Endpoints

### `GET /login` - Login Page

**Purpose:** Display login page with optional return URL.

**Query Parameters:**
- `return` (optional) - URL to redirect to after successful login (validated)

**Response:**
- `200 OK` - Login page HTML
- `400 Bad Request` - Invalid return URL (open redirect attempt)

**Security:**
- Return URL validation prevents open redirect attacks
- Only allows absolute paths starting with `/`
- Rejects external URLs, protocol-relative URLs, and dangerous protocols

**Usage:**
```go
authHandlers := handlers.NewAuthHandlers(authenticator)
router := handlers.NewRouter(&handlers.RouterHandlers{
    AuthHandlers: authHandlers,
})
```

### `GET /auth/oidc/login` - OIDC Login

**Purpose:** Initiate OIDC authentication flow.

**Query Parameters:**
- `return` (optional) - URL to redirect to after successful login

**Response:**
- `302 Found` - Redirect to OIDC provider
- `400 Bad Request` - Invalid return URL
- `500 Internal Server Error` - Authentication system error

**Flow:**
1. Validate return URL
2. Delegate to authenticator's `HandleLogin()`
3. Authenticator redirects to OIDC provider

### `GET /auth/oidc/callback` - OIDC Callback

**Purpose:** Handle OIDC provider callback after authentication.

**Query Parameters:**
- `code` - Authorization code from OIDC provider
- `state` - State parameter for CSRF protection

**Response:**
- `302 Found` - Redirect to dashboard or return URL
- `401 Unauthorized` - Authentication failed

**Flow:**
1. Delegate to authenticator's `HandleCallback()`
2. Authenticator validates state, exchanges code for token
3. Creates user session
4. Redirects to return URL or dashboard

### `POST /logout` - Logout

**Purpose:** Clear user session and logout.

**Method:** POST only (CSRF protection)

**Headers Required:**
- `X-CSRF-Token` - Valid CSRF token (or in form body as `csrf_token`)

**Response:**
- `302 Found` - Redirect to `/login`
- `405 Method Not Allowed` - Non-POST request
- `403 Forbidden` - Missing or invalid CSRF token
- `500 Internal Server Error` - Logout failed

**Security:**
- POST-only to prevent CSRF attacks
- Requires valid CSRF token
- Clears session via authenticator

## Key Files

- `auth.go` - Authentication route handlers with return URL validation
- `health.go` - Simple liveness check handler
- `readiness.go` - Complex readiness check with database and migration validation
- `errors.go` - Centralized error handling with content negotiation
- `router.go` - Router types, interfaces, and `NewRouter` orchestrator
- `router_setup.go` - Route registration functions (extracted from NewRouter)
