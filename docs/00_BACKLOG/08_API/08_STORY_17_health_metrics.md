# User Story: Health Check and Metrics Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As a **system operator**, I want **health check and metrics endpoints** so that **I can monitor the application's health and performance**.

---

## Acceptance Criteria

- [ ] GET /health - Health check endpoint
- [ ] GET /readiness - Readiness check endpoint
- [ ] GET /metrics - Prometheus metrics endpoint
- [ ] Health check validates database connection
- [ ] Health check validates email service
- [ ] Readiness check for Kubernetes
- [ ] Metrics include HTTP request counts
- [ ] Metrics include response times
- [ ] Metrics include error rates
- [ ] No authentication required

---

## Technical Details

### Routes
```go
r.Get("/health", handlers.HealthCheck)
r.Get("/readiness", handlers.ReadinessCheck)
r.Get("/metrics", handlers.Metrics)
```

### Health Check Handler
```go
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
    checks := make(map[string]string)
    
    // Check database
    if err := h.db.Ping(r.Context()); err != nil {
        checks["database"] = "unhealthy"
    } else {
        checks["database"] = "healthy"
    }
    
    // Check email service
    if err := h.email.HealthCheck(r.Context()); err != nil {
        checks["email"] = "unhealthy"
    } else {
        checks["email"] = "healthy"
    }
    
    // Determine overall status
    status := "healthy"
    for _, v := range checks {
        if v == "unhealthy" {
            status = "unhealthy"
            w.WriteHeader(http.StatusServiceUnavailable)
            break
        }
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": status,
        "checks": checks,
    })
}
```

### Readiness Check Handler
```go
func (h *Handlers) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
    // Simple check that app is ready to serve traffic
    if err := h.db.Ping(r.Context()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ready"))
}
```

### Metrics Handler
```go
func (h *Handlers) Metrics(w http.ResponseWriter, r *http.Request) {
    // Prometheus metrics format
    promhttp.Handler().ServeHTTP(w, r)
}
```

---

## Tasks

- [ ] Implement health check handler
- [ ] Implement readiness check handler
- [ ] Implement metrics handler
- [ ] Add database health check
- [ ] Add email service health check
- [ ] Configure Prometheus metrics
- [ ] Test health endpoints
- [ ] Test metrics collection

---

## Dependencies

**Depends on:** 
- 08_STORY_00_router_setup.md
- Epic 00 (Database)
- Epic 05 (Email)

**Blocks:** None

---

## Health Check Response

```json
{
  "status": "healthy",
  "checks": {
    "database": "healthy",
    "email": "healthy"
  }
}
```

---

## Metrics

```
# HTTP request count
http_requests_total{method="GET",path="/events",status="200"} 1234

# HTTP request duration
http_request_duration_seconds{method="GET",path="/events"} 0.123

# Error count
http_errors_total{method="POST",path="/events",status="500"} 5
```

---

## Testing Strategy

```go
func TestHealthCheck_Healthy(t *testing.T)
func TestHealthCheck_DatabaseDown(t *testing.T)
func TestHealthCheck_EmailDown(t *testing.T)
func TestReadinessCheck_Ready(t *testing.T)
func TestReadinessCheck_NotReady(t *testing.T)
func TestMetrics(t *testing.T)
```

---

## Kubernetes Integration

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readiness
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Prometheus:** https://prometheus.io/

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Health check working
- [ ] Readiness check working
- [ ] Metrics collecting
- [ ] Tests passing
- [ ] Documentation complete
