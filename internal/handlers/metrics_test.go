package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

type mockMetricsDataSource struct {
	statsFunc    func(ctx context.Context) (*AdminMetricsStats, error)
	emailFunc    func(ctx context.Context) (*EmailQueueMetrics, error)
	dbFunc       func() (*DBPoolMetrics, error)
}

func (m *mockMetricsDataSource) GetAdminStats(ctx context.Context) (*AdminMetricsStats, error) {
	if m.statsFunc != nil {
		return m.statsFunc(ctx)
	}
	return &AdminMetricsStats{TotalUsers: 10, TotalEvents: 5, TotalInvites: 20}, nil
}

func (m *mockMetricsDataSource) GetEmailQueueStatus(ctx context.Context) (*EmailQueueMetrics, error) {
	if m.emailFunc != nil {
		return m.emailFunc(ctx)
	}
	return &EmailQueueMetrics{QueueSize: 3, SendingCount: 1, FailedCount: 0, Healthy: true}, nil
}

func (m *mockMetricsDataSource) GetDBStats() (*DBPoolMetrics, error) {
	if m.dbFunc != nil {
		return m.dbFunc()
	}
	return &DBPoolMetrics{OpenConnections: 5, InUse: 2, Idle: 3, MaxOpenConnections: 10}, nil
}

func TestMetricsPage_Success(t *testing.T) {
	handler := NewMetricsHandler(&mockMetricsDataSource{})
	handler.SetTemplates(testTemplate(t, "admin_metrics.html"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/metrics", nil)
	ctx := testutil.CreateAdminContext()
	r = r.WithContext(ctx)

	handler.MetricsPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestMetricsPage_Unauthorized(t *testing.T) {
	handler := NewMetricsHandler(&mockMetricsDataSource{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/metrics", nil)
	r.Header.Set("Accept", "application/json")

	handler.MetricsPage(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for missing user, got %d", http.StatusForbidden, w.Code)
	}
}

func TestMetricsPage_ServiceError(t *testing.T) {
	handler := NewMetricsHandler(&mockMetricsDataSource{
		statsFunc: func(ctx context.Context) (*AdminMetricsStats, error) {
			return nil, fmt.Errorf("database unavailable")
		},
	})
	handler.SetTemplates(testTemplate(t, "admin_metrics.html"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/metrics", nil)
	r = r.WithContext(testutil.CreateAdminContext())

	handler.MetricsPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d (page renders with error), got %d", http.StatusOK, w.Code)
	}
}

func TestMetricsPage_PartialFailure(t *testing.T) {
	handler := NewMetricsHandler(&mockMetricsDataSource{
		emailFunc: func(ctx context.Context) (*EmailQueueMetrics, error) {
			return nil, fmt.Errorf("email checker unavailable")
		},
		dbFunc: func() (*DBPoolMetrics, error) {
			return nil, fmt.Errorf("db stats unavailable")
		},
	})
	handler.SetTemplates(testTemplate(t, "admin_metrics.html"))

	data := &MetricsPageData{}
	_ = data

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/metrics", nil)
	r = r.WithContext(testutil.CreateAdminContext())

	handler.MetricsPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d with graceful degradation, got %d", http.StatusOK, w.Code)
	}
}

func TestNewDBPoolMetricsFromSQLDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	metrics := NewDBPoolMetricsFromSQLDB(db)

	if metrics.MaxOpenConnections != 5 {
		t.Errorf("MaxOpenConnections = %d, want 5", metrics.MaxOpenConnections)
	}
	if metrics.OpenConnections < 0 {
		t.Errorf("OpenConnections should be >= 0, got %d", metrics.OpenConnections)
	}
}
