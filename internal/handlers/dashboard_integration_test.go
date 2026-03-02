package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
	"go.uber.org/mock/gomock"
)

type testAuthMiddleware struct {
	requireAuthCheck bool
}

func (m *testAuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.requireAuthCheck {
			_, ok := auth.UserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (m *testAuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func TestDashboardRoute_Integration(t *testing.T) {
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
	mockDashSvc.EXPECT().GetDashboardStats(gomock.Any(), int64(1)).Return(stats, nil).AnyTimes()
	mockDashSvc.EXPECT().GetRecentActivity(gomock.Any(), int64(1), gomock.Any()).Return(activity, nil).AnyTimes()

	dashboardHandler := NewDashboardHandler(mockDashSvc)
	tmpl := template.Must(template.New("dashboard.html").Parse(`
		<div>Stats: {{.Stats.TotalEvents}}</div>
		<div>Activities: {{len .Activities}}</div>
	`))
	dashboardHandler.SetTemplates(tmpl)

	authMiddleware := &testAuthMiddleware{requireAuthCheck: true}

	router := NewRouter(&RouterHandlers{
		DashboardHandler: dashboardHandler,
		AuthMiddleware:   authMiddleware,
	})

	t.Run("authenticated user can access dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleEventManager,
		}
		ctx := auth.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("unauthenticated user is rejected by auth middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestDashboardRoute_WithoutHandler(t *testing.T) {
	router := NewRouter(&RouterHandlers{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
