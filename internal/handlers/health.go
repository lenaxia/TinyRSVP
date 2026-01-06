package handlers

import (
	"encoding/json"
	"net/http"
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
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks"`
}

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
