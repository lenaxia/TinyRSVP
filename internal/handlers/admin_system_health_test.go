package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

// stubSystemHealth is a hand-rolled implementation of the
// AdminSystemHealthProvider interface used by the admin dashboard for its
// "ops at a glance" panels. We use a hand-rolled stub (not gomock) to avoid
// piling more entries onto the generated-mock set for a single small
// interface.
type stubSystemHealth struct {
	emailQueueFunc func(ctx context.Context) (*EmailQueueMetrics, error)
	dbStatsFunc    func() (*DBPoolMetrics, error)
}

func (s *stubSystemHealth) GetEmailQueueStatus(ctx context.Context) (*EmailQueueMetrics, error) {
	if s.emailQueueFunc != nil {
		return s.emailQueueFunc(ctx)
	}
	return nil, errors.New("not configured")
}

func (s *stubSystemHealth) GetDBStats() (*DBPoolMetrics, error) {
	if s.dbStatsFunc != nil {
		return s.dbStatsFunc()
	}
	return nil, errors.New("not configured")
}

// TestAdminDashboard_ExposesSystemHealthWhenProviderSet verifies that when an
// AdminSystemHealthProvider is wired, its data is included in the rendered
// page data. This is the mechanism that lets the redesigned admin dashboard
// show DB pool + email queue KPIs alongside the business stats.
func TestAdminDashboard_ExposesSystemHealthWhenProviderSet(t *testing.T) {
	adminStub := &stubAdminService{
		stats: &AdminDashboardStats{TotalUsers: 3, TotalEvents: 4, TotalInvites: 5},
	}
	healthStub := &stubSystemHealth{
		emailQueueFunc: func(ctx context.Context) (*EmailQueueMetrics, error) {
			return &EmailQueueMetrics{QueueSize: 2, SendingCount: 1, FailedCount: 0, Healthy: true}, nil
		},
		dbStatsFunc: func() (*DBPoolMetrics, error) {
			return &DBPoolMetrics{
				OpenConnections:    5,
				InUse:              1,
				Idle:               4,
				MaxOpenConnections: 25,
			}, nil
		},
	}

	handler := NewAdminDashboardHandler(adminStub)
	handler.SetSystemHealth(healthStub)

	data := runAdminDashboard(t, handler)

	if data.EmailQueue == nil {
		t.Fatal("EmailQueue should be populated when provider set")
	}
	if data.EmailQueue.QueueSize != 2 {
		t.Errorf("EmailQueue.QueueSize=%d want 2", data.EmailQueue.QueueSize)
	}
	if data.DBPool == nil {
		t.Fatal("DBPool should be populated when provider set")
	}
	if data.DBPool.OpenConnections != 5 {
		t.Errorf("DBPool.OpenConnections=%d want 5", data.DBPool.OpenConnections)
	}
}

// TestAdminDashboard_NoSystemHealthProviderMeansNoHealthData asserts the
// backward-compatible behavior: if no provider is wired, the handler still
// renders successfully with just the business stats — no crashes, no
// half-populated panels.
func TestAdminDashboard_NoSystemHealthProviderMeansNoHealthData(t *testing.T) {
	adminStub := &stubAdminService{
		stats: &AdminDashboardStats{TotalUsers: 3, TotalEvents: 4, TotalInvites: 5},
	}
	handler := NewAdminDashboardHandler(adminStub)
	// Deliberately do not call SetSystemHealth.

	data := runAdminDashboard(t, handler)
	if data.EmailQueue != nil {
		t.Errorf("EmailQueue should be nil when no provider set, got %+v", data.EmailQueue)
	}
	if data.DBPool != nil {
		t.Errorf("DBPool should be nil when no provider set, got %+v", data.DBPool)
	}
	// Business stats must still work.
	if data.Stats == nil || data.Stats.TotalUsers != 3 {
		t.Errorf("business stats missing, got %+v", data.Stats)
	}
}

// TestAdminDashboard_SystemHealthFailuresDoNotBlockRender asserts that if
// the health provider returns an error, the page still renders with the
// business stats. Ops overview is best-effort — a broken email queue check
// should not blank out the admin dashboard.
func TestAdminDashboard_SystemHealthFailuresDoNotBlockRender(t *testing.T) {
	adminStub := &stubAdminService{
		stats: &AdminDashboardStats{TotalUsers: 3, TotalEvents: 4, TotalInvites: 5},
	}
	healthStub := &stubSystemHealth{
		emailQueueFunc: func(ctx context.Context) (*EmailQueueMetrics, error) {
			return nil, errors.New("email checker down")
		},
		dbStatsFunc: func() (*DBPoolMetrics, error) {
			return nil, errors.New("db check failed")
		},
	}

	handler := NewAdminDashboardHandler(adminStub)
	handler.SetSystemHealth(healthStub)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	user := &models.User{ID: 1, Email: "a@a", Name: "A", Role: models.RoleAdmin}
	req = req.WithContext(auth.WithUser(req.Context(), user))
	w := httptest.NewRecorder()

	handler.AdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

// stubAdminService is a small hand-rolled stub for AdminDashboardService,
// avoiding a gomock controller for tests that only care about the stats
// value being propagated.
type stubAdminService struct {
	stats *AdminDashboardStats
	err   error
}

func (s *stubAdminService) GetAdminStats(_ context.Context) (*AdminDashboardStats, error) {
	return s.stats, s.err
}

// runAdminDashboard exercises the handler against a real HTTP recorder and
// retrieves the resolved AdminDashboardPageData via the exposed lastPageData
// getter (added for testability — see admin.go).
func runAdminDashboard(t *testing.T, handler *AdminDashboardHandler) *AdminDashboardPageData {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	user := &models.User{ID: 1, Email: "a@a", Name: "A", Role: models.RoleAdmin}
	req = req.WithContext(auth.WithUser(req.Context(), user))
	w := httptest.NewRecorder()

	handler.AdminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	data := handler.LastPageData()
	if data == nil {
		t.Fatal("LastPageData returned nil")
	}
	return data
}
