package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type MetricsDataSource interface {
	GetAdminStats(ctx context.Context) (*AdminMetricsStats, error)
	GetEmailQueueStatus(ctx context.Context) (*EmailQueueMetrics, error)
	GetDBStats() (*DBPoolMetrics, error)
}

type AdminMetricsStats struct {
	TotalUsers   int
	TotalEvents  int
	TotalInvites int
}

type EmailQueueMetrics struct {
	QueueSize    int
	SendingCount int
	FailedCount  int
	Healthy      bool
}

type DBPoolMetrics struct {
	OpenConnections int
	InUse           int
	Idle            int
	WaitCount       int64
	WaitDuration    time.Duration
	MaxOpenConnections int
}

type MetricsHandler struct {
	source    MetricsDataSource
	templates *template.Template
}

type MetricsPageData struct {
	ActivePage  string
	IsAdmin     bool
	User        *models.User
	Stats       *AdminMetricsStats
	EmailQueue  *EmailQueueMetrics
	DBPool      *DBPoolMetrics
	Error       string
	GeneratedAt time.Time
}

func NewMetricsHandler(source MetricsDataSource) *MetricsHandler {
	return &MetricsHandler{source: source}
}

func (h *MetricsHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *MetricsHandler) MetricsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "view admin metrics",
			Resource: "Admin Metrics",
		})
		return
	}

	data := &MetricsPageData{
		ActivePage:  "admin",
		IsAdmin:     isAdminRequest(r),
		User:        user,
		GeneratedAt: time.Now(),
	}

	stats, err := h.source.GetAdminStats(r.Context())
	if err != nil {
		data.Error = fmt.Sprintf("Failed to load stats: %v", err)
		h.renderPage(w, http.StatusOK, data)
		return
	}
	data.Stats = stats

	emailQueue, err := h.source.GetEmailQueueStatus(r.Context())
	if err != nil {
		slog.Warn("metrics: failed to load email queue status", "error", err)
	} else {
		data.EmailQueue = emailQueue
	}

	dbPool, err := h.source.GetDBStats()
	if err != nil {
		slog.Warn("metrics: failed to load DB pool stats", "error", err)
	} else {
		data.DBPool = dbPool
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *MetricsHandler) renderPage(w http.ResponseWriter, status int, data *MetricsPageData) {
	renderHTML(w, h.templates, "admin_metrics.html", status, data)
}

func NewDBPoolMetricsFromSQLDB(db *sql.DB) *DBPoolMetrics {
	stats := db.Stats()
	return &DBPoolMetrics{
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxOpenConnections: stats.MaxOpenConnections,
	}
}
