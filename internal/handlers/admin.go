package handlers

import (
	"context"
	"html/template"
	"log/slog"
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

// AdminSystemHealthProvider supplies best-effort operational health data for
// the admin dashboard's system panels. Both methods can fail without
// blocking the page — the handler treats a nil return as "not available".
type AdminSystemHealthProvider interface {
	GetEmailQueueStatus(ctx context.Context) (*EmailQueueMetrics, error)
	GetDBStats() (*DBPoolMetrics, error)
}

type AdminDashboardStats = admin.AdminStats

type AdminDashboardHandler struct {
	service      AdminDashboardService
	systemHealth AdminSystemHealthProvider
	templates    *template.Template
	lastPageData *AdminDashboardPageData
}

type AdminDashboardPageData struct {
	ActivePage string
	IsAdmin    bool
	User       *models.User
	Stats      *AdminDashboardStats
	EmailQueue *EmailQueueMetrics
	DBPool     *DBPoolMetrics
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

// SetSystemHealth wires the optional operational health provider. When set,
// the admin dashboard's rendered data includes email queue + DB pool KPIs.
// When left unset, the page still renders with just business stats.
func (h *AdminDashboardHandler) SetSystemHealth(p AdminSystemHealthProvider) {
	h.systemHealth = p
}

// LastPageData returns the most recently rendered page data. Exposed so
// tests can assert on what would be passed to the template without needing
// to parse a real template — the render itself is exercised by template
// tests in templates/web/.
func (h *AdminDashboardHandler) LastPageData() *AdminDashboardPageData {
	return h.lastPageData
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
		HandleError(w, r, err)
		return
	}

	data := &AdminDashboardPageData{
		ActivePage: "admin",
		IsAdmin:    isAdminRequest(r),
		User:       user,
		Stats:      stats,
	}

	if h.systemHealth != nil {
		if q, err := h.systemHealth.GetEmailQueueStatus(r.Context()); err != nil {
			slog.Warn("admin dashboard: email queue status unavailable", "error", err)
		} else {
			data.EmailQueue = q
		}
		if db, err := h.systemHealth.GetDBStats(); err != nil {
			slog.Warn("admin dashboard: db pool stats unavailable", "error", err)
		} else {
			data.DBPool = db
		}
	}

	h.lastPageData = data
	h.renderPage(w, http.StatusOK, data)
}

func (h *AdminDashboardHandler) renderPage(w http.ResponseWriter, status int, data *AdminDashboardPageData) {
	renderHTML(w, h.templates, "admin_dashboard.html", status, data)
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
	IsAdmin    bool
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
		HandleError(w, r, err)
		return
	}

	total, err := h.service.CountUsers(r.Context())
	if err != nil {
		slog.Warn("user management: failed to get user count", "error", err)
		h.renderPage(w, http.StatusOK, &UserManagementPageData{
			ActivePage: "admin",
			IsAdmin:    isAdminRequest(r),
			User:       user,
			Users:      users,
			Error:      "Failed to get user count",
		})
		return
	}

	csrfToken := middleware.GetCSRFToken(r.Context())

	data := &UserManagementPageData{
		ActivePage: "admin",
		IsAdmin:    isAdminRequest(r),
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
	renderHTML(w, h.templates, "user_management.html", status, data)
}
