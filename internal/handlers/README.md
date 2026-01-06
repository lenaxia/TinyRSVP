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

- `health.go` - Health check endpoint (liveness probe)
- `health_test.go` - Health check tests
- `readiness.go` - Readiness check endpoint (readiness probe)
- `readiness_test.go` - Readiness check tests

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

## Key Files

- `health.go` - Simple liveness check handler
- `readiness.go` - Complex readiness check with database and migration validation
