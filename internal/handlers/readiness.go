package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/lenaxia/tinyrsvp/internal/db"
)

type ReadinessHandler struct {
	version  string
	database db.Database
	migrator db.Migrator
	logger   *slog.Logger
}

func NewReadinessHandler(version string, database db.Database, migrator db.Migrator) *ReadinessHandler {
	return &ReadinessHandler{
		version:  version,
		database: database,
		migrator: migrator,
		logger:   slog.Default(),
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
		h.logger.Error("Database health check failed", "message", dbCheck.Message)
	}

	migrationCheck := h.checkMigrations(ctx)
	response.Checks["migrations"] = migrationCheck
	if migrationCheck.Status == StatusUnhealthy {
		response.Status = StatusUnhealthy
		h.logger.Error("Migration health check failed", "message", migrationCheck.Message)
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
		if err == migrate.ErrNilVersion {
			return HealthCheck{
				Status:  StatusUnhealthy,
				Message: "No migrations applied",
			}
		}
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
