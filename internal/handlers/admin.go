package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/lenaxia/tinyrsvp/internal/admin"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type AdminDashboardService interface {
	GetAdminStats(ctx context.Context) (*admin.AdminStats, error)
}

type AdminDashboardStats = admin.AdminStats

type AdminDashboardHandler struct {
	service   AdminDashboardService
	templates *template.Template
}

type AdminDashboardPageData struct {
	ActivePage string
	User       *models.User
	Stats      *AdminDashboardStats
	Error      string
	Loading    bool
}

func NewAdminDashboardHandler(service AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		service: service,
	}
}

func (h *AdminDashboardHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *AdminDashboardHandler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "view admin dashboard",
			Resource: "Admin Dashboard",
		})
		return
	}

	stats, err := h.service.GetAdminStats(r.Context())
	if err != nil {
		log.Printf("Admin dashboard: failed to load stats: %v", err)
		HandleError(w, r, err)
		return
	}

	data := &AdminDashboardPageData{
		ActivePage: "admin",
		User:       user,
		Stats:      stats,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *AdminDashboardHandler) renderPage(w http.ResponseWriter, status int, data *AdminDashboardPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "admin_dashboard.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Dashboard - TinyRSVP</title>
</head>
<body>
    <h1>Admin Dashboard</h1>
    %s
</body>
</html>`, func() string {
		if data.Error != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.Error)
		}
		if data.Stats != nil {
			return fmt.Sprintf("<p>Total Users: %d</p>", data.Stats.TotalUsers)
		}
		return "<p>Loading...</p>"
	}())
}

type UserListService interface {
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	CountUsers(ctx context.Context) (int, error)
}

type UserManagementHandler struct {
	service   UserListService
	templates *template.Template
}

type UserManagementPageData struct {
	ActivePage string
	User       *models.User
	Users      []*models.User
	Total      int
	Limit      int
	Offset     int
	CSRFToken  string
	Error      string
	Success    string
	Loading    bool
}

func NewUserManagementHandler(service UserListService) *UserManagementHandler {
	return &UserManagementHandler{
		service: service,
	}
}

func (h *UserManagementHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *UserManagementHandler) UserManagementPage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "view user management",
			Resource: "User Management",
		})
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	users, err := h.service.ListUsers(r.Context(), limit, offset)
	if err != nil {
		log.Printf("User management: failed to load users: %v", err)
		HandleError(w, r, err)
		return
	}

	total, err := h.service.CountUsers(r.Context())
	if err != nil {
		log.Printf("User management: failed to get user count: %v", err)
		h.renderPage(w, http.StatusOK, &UserManagementPageData{
			ActivePage: "admin",
			User:       user,
			Users:      users,
			Error:      "Failed to get user count",
		})
		return
	}

	csrfToken := middleware.GetCSRFToken(r.Context())

	data := &UserManagementPageData{
		ActivePage: "admin",
		User:       user,
		Users:      users,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		CSRFToken:  csrfToken,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *UserManagementHandler) renderPage(w http.ResponseWriter, status int, data *UserManagementPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "user_management.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Management - TinyRSVP</title>
</head>
<body>
    <h1>User Management</h1>
    %s
</body>
</html>`, func() string {
		if data.Error != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.Error)
		}
		if data.Users != nil {
			return fmt.Sprintf("<p>Total Users: %d</p>", len(data.Users))
		}
		return "<p>Loading...</p>"
	}())
}
