package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

func TestDashboardHandler_Dashboard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stats := &events.DashboardStats{
		TotalEvents:     5,
		DraftEvents:     2,
		PublishedEvents: 3,
		TotalInvites:    50,
		PendingInvites:  10,
		TotalRSVPs:      40,
		AcceptedRSVPs:   30,
		DeclinedRSVPs:   10,
		ResponseRate:    80,
	}

	activity := []*events.ActivityItem{
		{
			Icon:        "📅",
			Title:       "Event Created",
			Description: "New Year Party created",
			Time:        "2 hours ago",
		},
	}

	mockDashSvc := services.NewMockDashboardService(ctrl)
	mockDashSvc.EXPECT().GetDashboardStats(gomock.Any(), int64(1)).Return(stats, nil)
	mockDashSvc.EXPECT().GetRecentActivity(gomock.Any(), int64(1), gomock.Any()).Return(activity, nil)

	handler := NewDashboardHandler(mockDashSvc)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDashSvc := services.NewMockDashboardService(ctrl)
	handler := NewDashboardHandler(mockDashSvc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.Dashboard(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDashboardHandler_Dashboard_StatsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDashSvc := services.NewMockDashboardService(ctrl)
	mockDashSvc.EXPECT().GetDashboardStats(gomock.Any(), int64(1)).Return(nil, &models.NotFoundError{Resource: "Stats"})

	handler := NewDashboardHandler(mockDashSvc)

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

	// Primary data load failure now uses HandleError (returns proper
	// error status, not HTTP 200 with in-page error).
	if w.Code == http.StatusOK {
		t.Error("expected non-200 status for primary stats failure, got 200 (should use HandleError)")
	}
}

func TestDashboardHandler_Dashboard_NoTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stats := &events.DashboardStats{
		TotalEvents: 5,
	}

	mockDashSvc := services.NewMockDashboardService(ctrl)
	mockDashSvc.EXPECT().GetDashboardStats(gomock.Any(), int64(1)).Return(stats, nil)
	mockDashSvc.EXPECT().GetRecentActivity(gomock.Any(), int64(1), gomock.Any()).Return([]*events.ActivityItem{}, nil)

	handler := NewDashboardHandler(mockDashSvc)

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

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d when templates are nil, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDashboardStats_CalculateResponseRate(t *testing.T) {
	tests := []struct {
		name         string
		totalRSVPs   int
		totalInvites int
		expectedRate int
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
