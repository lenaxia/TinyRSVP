# User Story: Health Check and Readiness Endpoints

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)
**Priority:** High
**Status:** ✅ Complete
**Estimated Effort:** 3 hours
**Completed:** 2026-01-06

---

## User Story

As an **operator**, I want **health check and readiness endpoints** so that **I can monitor application status and ensure proper deployment**.

---

## Acceptance Criteria

- [x] `/health` endpoint returns application health status
- [x] `/ready` endpoint returns readiness status
- [x] Database connectivity checked in health endpoint
- [x] Proper HTTP status codes returned (200 for healthy, 503 for unhealthy)
- [x] Response includes component status details
- [x] Endpoints accessible without authentication
- [x] All tests pass with timeout

---

## Technical Details

### Endpoints

**Health Endpoint:** `GET /health`
- Purpose: Liveness probe - is the application running?
- Returns: 200 if application is alive, 503 if dead
- Checks: Basic application health

**Readiness Endpoint:** `GET /ready`
- Purpose: Readiness probe - can the application serve traffic?
- Returns: 200 if ready, 503 if not ready
- Checks: Database connectivity, migrations applied

### Response Format

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

### Status Values

- `healthy` - Component is functioning normally
- `degraded` - Component is functioning but with issues
- `unhealthy` - Component is not functioning

### HTTP Status Codes

- `200 OK` - All checks passed
- `503 Service Unavailable` - One or more checks failed

---

## Tasks

### Phase 1: Health Check Handler (TDD)
- [x] Write test for health endpoint returns 200
- [x] Write test for health endpoint response format
- [x] Write test for health endpoint includes version
- [x] Implement health check handler
- [x] Run tests (should pass)

### Phase 2: Readiness Check Handler (TDD)
- [x] Write test for readiness endpoint returns 200 when ready
- [x] Write test for readiness endpoint returns 503 when not ready
- [x] Write test for database connectivity check
- [x] Write test for migration version check
- [x] Implement readiness check handler
- [x] Run tests (should pass)

### Phase 3: Component Health Checks (TDD)
- [x] Write test for database health check
- [x] Write test for database health check failure
- [x] Write test for database latency measurement
- [x] Implement database health checker
- [x] Run tests (should pass)

### Phase 4: Integration
- [x] Register health endpoints in router
- [x] Add health check logging
- [x] Test with real database
- [x] Document endpoints

---

## Implementation

### Health Check Types

```go
package handlers

import (
    "context"
    "time"
)

type HealthStatus string

const (
    StatusHealthy   HealthStatus = "healthy"
    StatusDegraded  HealthStatus = "degraded"
    StatusUnhealthy HealthStatus = "unhealthy"
)

type HealthCheck struct {
    Status    HealthStatus `json:"status"`
    Message   string       `json:"message,omitempty"`
    LatencyMs *int64       `json:"latency_ms,omitempty"`
    Version   *uint        `json:"version,omitempty"`
}

type HealthResponse struct {
    Status    HealthStatus            `json:"status"`
    Timestamp time.Time               `json:"timestamp"`
    Version   string                  `json:"version"`
    Checks    map[string]HealthCheck  `json:"checks"`
}
```

### Health Handler

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
)

type HealthHandler struct {
    version string
}

func NewHealthHandler(version string) *HealthHandler {
    return &HealthHandler{
        version: version,
    }
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    response := HealthResponse{
        Status:    StatusHealthy,
        Timestamp: time.Now().UTC(),
        Version:   h.version,
        Checks:    make(map[string]HealthCheck),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
```

### Readiness Handler

```go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/db"
)

type ReadinessHandler struct {
    version  string
    database db.Database
    migrator db.Migrator
}

func NewReadinessHandler(version string, database db.Database, migrator db.Migrator) *ReadinessHandler {
    return &ReadinessHandler{
        version:  version,
        database: database,
        migrator: migrator,
    }
}

func (h *ReadinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    response := HealthResponse{
        Status:    StatusHealthy,
        Timestamp: time.Now().UTC(),
        Version:   h.version,
        Checks:    make(map[string]HealthCheck),
    }
    
    dbCheck := h.checkDatabase(ctx)
    response.Checks["database"] = dbCheck
    if dbCheck.Status == StatusUnhealthy {
        response.Status = StatusUnhealthy
    }
    
    migrationCheck := h.checkMigrations(ctx)
    response.Checks["migrations"] = migrationCheck
    if migrationCheck.Status == StatusUnhealthy {
        response.Status = StatusUnhealthy
    }
    
    statusCode := http.StatusOK
    if response.Status == StatusUnhealthy {
        statusCode = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

func (h *ReadinessHandler) checkDatabase(ctx context.Context) HealthCheck {
    start := time.Now()
    
    err := h.database.Ping(ctx)
    latency := time.Since(start).Milliseconds()
    
    if err != nil {
        return HealthCheck{
            Status:  StatusUnhealthy,
            Message: "Database unreachable: " + err.Error(),
        }
    }
    
    return HealthCheck{
        Status:    StatusHealthy,
        Message:   "Connected",
        LatencyMs: &latency,
    }
}

func (h *ReadinessHandler) checkMigrations(ctx context.Context) HealthCheck {
    version, dirty, err := h.migrator.Version(ctx)
    if err != nil {
        return HealthCheck{
            Status:  StatusUnhealthy,
            Message: "Cannot determine migration version: " + err.Error(),
        }
    }
    
    if dirty {
        return HealthCheck{
            Status:  StatusUnhealthy,
            Message: "Migrations in dirty state",
            Version: &version,
        }
    }
    
    return HealthCheck{
        Status:  StatusHealthy,
        Message: "Up to date",
        Version: &version,
    }
}
```

---

## Testing Requirements

### Unit Tests

```go
func TestHealthHandler(t *testing.T) {
    handler := NewHealthHandler("0.1.0")
    
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    var response HealthResponse
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    
    if response.Status != StatusHealthy {
        t.Errorf("Expected status healthy, got %s", response.Status)
    }
    
    if response.Version != "0.1.0" {
        t.Errorf("Expected version 0.1.0, got %s", response.Version)
    }
}

func TestReadinessHandler_Healthy(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    migrator, err := db.NewMigrator(db.DB(), "../../migrations/sqlite")
    if err != nil {
        t.Fatalf("Failed to create migrator: %v", err)
    }
    
    if err := migrator.Up(context.Background()); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }
    
    handler := NewReadinessHandler("0.1.0", db, migrator)
    
    req := httptest.NewRequest("GET", "/ready", nil)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    var response HealthResponse
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    
    if response.Status != StatusHealthy {
        t.Errorf("Expected status healthy, got %s", response.Status)
    }
    
    if _, ok := response.Checks["database"]; !ok {
        t.Error("Expected database check in response")
    }
    
    if _, ok := response.Checks["migrations"]; !ok {
        t.Error("Expected migrations check in response")
    }
}

func TestReadinessHandler_DatabaseUnhealthy(t *testing.T) {
    db := setupTestDB(t)
    db.Close()
    
    migrator, _ := db.NewMigrator(db.DB(), "../../migrations/sqlite")
    
    handler := NewReadinessHandler("0.1.0", db, migrator)
    
    req := httptest.NewRequest("GET", "/ready", nil)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusServiceUnavailable {
        t.Errorf("Expected status 503, got %d", w.Code)
    }
    
    var response HealthResponse
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    
    if response.Status != StatusUnhealthy {
        t.Errorf("Expected status unhealthy, got %s", response.Status)
    }
    
    dbCheck, ok := response.Checks["database"]
    if !ok {
        t.Fatal("Expected database check in response")
    }
    
    if dbCheck.Status != StatusUnhealthy {
        t.Errorf("Expected database status unhealthy, got %s", dbCheck.Status)
    }
}
```

---

## Kubernetes Integration

### Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

---

## Dependencies

**Depends on:** 
- [00_STORY_go_module_setup.md](00_STORY_go_module_setup.md)
- [00_STORY_database_connection.md](00_STORY_database_connection.md)
- [00_STORY_database_migrations.md](00_STORY_database_migrations.md)

**Blocks:** 
- [00_STORY_docker_setup.md](00_STORY_docker_setup.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/handlers/...`)
- [x] Test coverage >= 80%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Health endpoint functional
- [x] Readiness endpoint functional
- [x] Database checks working
- [x] Migration checks working
- [x] Documentation complete
- [x] Changes committed to git

---

## Notes

### Health vs Readiness

**Health (Liveness):**
- Is the application alive?
- Should Kubernetes restart the pod?
- Checks basic application functionality

**Readiness:**
- Can the application serve traffic?
- Should Kubernetes route traffic to this pod?
- Checks dependencies (database, external services)

### Best Practices
- Keep health checks lightweight
- Set appropriate timeouts (5 seconds max)
- Return detailed status for debugging
- Don't include authentication on health endpoints
- Log health check failures

### Monitoring Integration
- Prometheus metrics can be added later
- Health endpoints provide basic monitoring
- Use for alerting on service degradation

---

## References

- **README-LLM.md:** TDD Requirements
- **Kubernetes Health Checks:** https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
- **Health Check Pattern:** https://microservices.io/patterns/observability/health-check-api.html
