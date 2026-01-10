package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockDashboardService struct {
	stats    *events.DashboardStats
	activity []*events.ActivityItem
	err      error
}

func (m *mockDashboardService) GetDashboardStats(ctx context.Context, userID int64) (*events.DashboardStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func (m *mockDashboardService) GetRecentActivity(ctx context.Context, userID int64, limit int) ([]*events.ActivityItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.activity, nil
}

func TestDashboardHandler_Dashboard_Success(t *testing.T) {
	stats := &events.DashboardStats{
		TotalEvents:      5,
		DraftEvents:      2,
		PublishedEvents:  3,
		TotalInvites:     50,
		PendingInvites:   10,
		TotalRSVPs:       40,
		AcceptedRSVPs:    30,
		DeclinedRSVPs:    10,
		ResponseRate:     80,
	}

	activity := []*events.ActivityItem{
		{
			Icon:        "📅",
			Title:       "Event Created",
			Description: "New Year Party created",
			Time:        "2 hours ago",
		},
	}

	service := &mockDashboardService{
		stats:    stats,
		activity: activity,
	}

	handler := NewDashboardHandler(service)

	tmpl := template.Must(template.New("dashboard.html").Parse(`
		<div>Stats: {{.Stats.TotalEvents}}</div>
		<div>Activities: {{len .Activities}}</div>
	`))
	handler.SetTemplates(tmpl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}
}

func TestDashboardHandler_Dashboard_NoUser(t *testing.T) {
	service := &mockDashboardService{}
	handler := NewDashboardHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.Dashboard(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDashboardHandler_Dashboard_StatsError(t *testing.T) {
	service := &mockDashboardService{
		err: &models.NotFoundError{Resource: "Stats"},
	}

	handler := NewDashboardHandler(service)

	tmpl := template.Must(template.New("dashboard.html").Parse(`<div>Error: {{.Error}}</div>`))
	handler.SetTemplates(tmpl)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDashboardHandler_Dashboard_NoTemplate(t *testing.T) {
	stats := &events.DashboardStats{
		TotalEvents: 5,
	}

	service := &mockDashboardService{
		stats:    stats,
		activity: []*events.ActivityItem{},
	}

	handler := NewDashboardHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Dashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDashboardStats_CalculateResponseRate(t *testing.T) {
	tests := []struct {
		name          string
		totalRSVPs    int
		totalInvites  int
		expectedRate  int
	}{
		{
			name:         "80% response rate",
			totalRSVPs:   80,
			totalInvites: 100,
			expectedRate: 80,
		},
		{
			name:         "0% response rate",
			totalRSVPs:   0,
			totalInvites: 100,
			expectedRate: 0,
		},
		{
			name:         "100% response rate",
			totalRSVPs:   100,
			totalInvites: 100,
			expectedRate: 100,
		},
		{
			name:         "no invites",
			totalRSVPs:   0,
			totalInvites: 0,
			expectedRate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &events.DashboardStats{
				TotalRSVPs:   tt.totalRSVPs,
				TotalInvites: tt.totalInvites,
			}
			stats.CalculateResponseRate()

			if stats.ResponseRate != tt.expectedRate {
				t.Errorf("expected response rate %d, got %d", tt.expectedRate, stats.ResponseRate)
			}
		})
	}
}

func TestActivityItem_FormatTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now",
			time:     now,
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "30 minutes ago",
			time:     now.Add(-30 * time.Minute),
			expected: "30 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "2 hours ago",
			time:     now.Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "1 day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "3 days ago",
			time:     now.Add(-72 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "8 days ago",
			time:     now.Add(-192 * time.Hour),
			expected: now.Add(-192 * time.Hour).Format("Jan 2, 2006"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := events.FormatTimeAgo(tt.time)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
