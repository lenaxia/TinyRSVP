package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type DashboardHandler struct {
	service   events.DashboardService
	templates *template.Template
}

type DashboardPageData struct {
	ActivePage string
	IsAdmin    bool
	User       *models.User
	Stats      *events.DashboardStats
	Activities []*events.ActivityItem
	Error      string
	Loading    bool
}

func NewDashboardHandler(service events.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: service,
	}
}

func (h *DashboardHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "view dashboard",
			Resource: "Dashboard",
		})
		return
	}

	stats, err := h.service.GetDashboardStats(r.Context(), user.ID)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	activity, err := h.service.GetRecentActivity(r.Context(), user.ID, 10)
	if err != nil {
		slog.Warn("dashboard: failed to load recent activity", "user_id", user.ID, "error", err)
		h.renderPage(w, http.StatusOK, &DashboardPageData{
			ActivePage: "dashboard",
			IsAdmin:    isAdminRequest(r),
			User:       user,
			Stats:      stats,
			Error:      "Failed to load recent activity",
		})
		return
	}

	data := &DashboardPageData{
		ActivePage: "dashboard",
		IsAdmin:    isAdminRequest(r),
		User:       user,
		Stats:      stats,
		Activities: activity,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *DashboardHandler) renderPage(w http.ResponseWriter, status int, data *DashboardPageData) {
	renderHTML(w, h.templates, "dashboard.html", status, data)
}
