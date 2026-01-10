package handlers

import (
	"fmt"
	"html/template"
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
		h.renderPage(w, http.StatusOK, &DashboardPageData{
			ActivePage: "dashboard",
			User:       user,
			Error:      "Failed to load dashboard statistics",
		})
		return
	}

	activity, err := h.service.GetRecentActivity(r.Context(), user.ID, 10)
	if err != nil {
		h.renderPage(w, http.StatusOK, &DashboardPageData{
			ActivePage: "dashboard",
			User:       user,
			Stats:      stats,
			Error:      "Failed to load recent activity",
		})
		return
	}

	data := &DashboardPageData{
		ActivePage: "dashboard",
		User:       user,
		Stats:      stats,
		Activities: activity,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *DashboardHandler) renderPage(w http.ResponseWriter, status int, data *DashboardPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - TinyRSVP</title>
</head>
<body>
    <h1>Dashboard</h1>
    %s
</body>
</html>`, func() string {
		if data.Error != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.Error)
		}
		if data.Stats != nil {
			return fmt.Sprintf("<p>Total Events: %d</p>", data.Stats.TotalEvents)
		}
		return "<p>Loading...</p>"
	}())
}
