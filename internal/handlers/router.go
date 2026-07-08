package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type RouteInfo struct {
	Method  string
	Pattern string
}

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrEmptyParameter   = errors.New("parameter cannot be empty")
)

type Router struct {
	mux      chi.Router
	handlers *RouterHandlers
}

type RouterHandlers struct {
	AuthHandlers     *AuthHandlers
	LoginHandler     http.Handler
	CallbackHandler  http.Handler
	LogoutHandler    http.Handler
	HealthHandler    http.Handler
	ReadinessHandler http.Handler

	DashboardHandler         DashboardHandlerInterface
	EventHandlers            EventHandlerInterface
	EventWebHandlers         EventWebHandlerInterface
	QuestionHandlers         QuestionHandlerInterface
	InviteHandlers           InviteHandlerInterface
	InviteWebHandlers        InviteWebHandlerInterface
	ImportInviteHandlers     ImportInviteHandlerInterface
	ManualInviteHandlers     ManualInviteHandlerInterface
	RevokeInviteHandlers     RevokeInviteHandlerInterface
	RegenerateInviteHandlers RegenerateInviteHandlerInterface
	ListInviteHandlers       ListInviteHandlerInterface
	GetInviteHandlers        GetInviteHandlerInterface
	UpdateInviteHandlers     UpdateInviteHandlerInterface
	DeleteInviteHandlers     DeleteInviteHandlerInterface
	SendInviteHandlers       SendInviteHandlerInterface
	ImageHandlers            RouteRegistrar
	RSVPHandler              RSVPHandlerInterface
	RSVPSummaryHandler       RSVPSummaryHandlerInterface
	UserHandler              UserHandlerInterface
	TemplateHandlers         TemplateHandlerInterface
	TemplateEditorHandlers   RouteRegistrar
	CustomizationHandlers    CustomizationHandlerInterface
	AssetHandler             AssetHandlerInterface

	AdminDashboardHandler AdminDashboardHandlerInterface
	UserManagementHandler UserManagementHandlerInterface
	SettingsHandler       SettingsHandlerInterface
	AdminMetricsHandler   AdminMetricsHandlerInterface

	CleanupHandler     http.Handler
	EmailHealthHandler http.Handler
	MetricsHandler     http.Handler

	AuthMiddleware AuthMiddlewareInterface

	MetricsMiddleware func(http.Handler) http.Handler

	StaticFileServer http.Handler

	Logger *log.Logger
}

type DashboardHandlerInterface interface {
	Dashboard(w http.ResponseWriter, r *http.Request)
}

type RouteRegistrar interface {
	RegisterRoutes(r chi.Router)
}

type EventHandlerInterface interface {
	ListEvents(w http.ResponseWriter, r *http.Request)
	CreateEvent(w http.ResponseWriter, r *http.Request)
	GetEvent(w http.ResponseWriter, r *http.Request)
	UpdateEvent(w http.ResponseWriter, r *http.Request)
	DeleteEvent(w http.ResponseWriter, r *http.Request)
}

type EventWebHandlerInterface interface {
	ListEventsPage(w http.ResponseWriter, r *http.Request)
	NewEventForm(w http.ResponseWriter, r *http.Request)
	EditEventForm(w http.ResponseWriter, r *http.Request)
	GetEventPage(w http.ResponseWriter, r *http.Request)
	CreateEventFromForm(w http.ResponseWriter, r *http.Request)
	UpdateEventFromForm(w http.ResponseWriter, r *http.Request)
	PublishEventAction(w http.ResponseWriter, r *http.Request)
	CancelEventAction(w http.ResponseWriter, r *http.Request)
	DeleteEventAction(w http.ResponseWriter, r *http.Request)
}

type QuestionHandlerInterface interface {
	RegisterRoutes(r chi.Router)
}

type InviteHandlerInterface interface {
	CreateInvite(w http.ResponseWriter, r *http.Request)
}

type InviteWebHandlerInterface interface {
	ListInvitesPage(w http.ResponseWriter, r *http.Request)
}

type ImportInviteHandlerInterface interface {
	ImportInvites(w http.ResponseWriter, r *http.Request)
}

type ManualInviteHandlerInterface interface {
	CreateManualInvite(w http.ResponseWriter, r *http.Request)
}

type RevokeInviteHandlerInterface interface {
	RevokeInvite(w http.ResponseWriter, r *http.Request)
}

type RegenerateInviteHandlerInterface interface {
	RegenerateInviteToken(w http.ResponseWriter, r *http.Request)
}

type ListInviteHandlerInterface interface {
	ListInvites(w http.ResponseWriter, r *http.Request)
}

type GetInviteHandlerInterface interface {
	GetInvite(w http.ResponseWriter, r *http.Request)
}

type UpdateInviteHandlerInterface interface {
	UpdateInvite(w http.ResponseWriter, r *http.Request)
}

type DeleteInviteHandlerInterface interface {
	DeleteInvite(w http.ResponseWriter, r *http.Request)
}

type SendInviteHandlerInterface interface {
	SendInvite(w http.ResponseWriter, r *http.Request)
}

type RSVPHandlerInterface interface {
	GetRSVPPage(w http.ResponseWriter, r *http.Request)
	SubmitRSVP(w http.ResponseWriter, r *http.Request)
	UpdateRSVP(w http.ResponseWriter, r *http.Request)
	GetConfirmationPage(w http.ResponseWriter, r *http.Request)
	GetCalendar(w http.ResponseWriter, r *http.Request)
	Unsubscribe(w http.ResponseWriter, r *http.Request)
}

type RSVPSummaryHandlerInterface interface {
	GetRSVPSummary(w http.ResponseWriter, r *http.Request)
}

type UserHandlerInterface interface {
	ListUsers(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	UpdateUserRole(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
}

type TemplateHandlerInterface interface {
	RegisterRoutes(r chi.Router)
}

type CustomizationHandlerInterface interface {
	RegisterRoutes(r chi.Router)
	CustomizationPage(w http.ResponseWriter, r *http.Request)
}

type AssetHandlerInterface interface {
	ServeAsset(w http.ResponseWriter, r *http.Request)
}

type AdminDashboardHandlerInterface interface {
	AdminDashboard(w http.ResponseWriter, r *http.Request)
}

type UserManagementHandlerInterface interface {
	UserManagementPage(w http.ResponseWriter, r *http.Request)
}

type SettingsHandlerInterface interface {
	SettingsPage(w http.ResponseWriter, r *http.Request)
}

type AdminMetricsHandlerInterface interface {
	MetricsPage(w http.ResponseWriter, r *http.Request)
}

type AuthMiddlewareInterface interface {
	RequireAuth(next http.Handler) http.Handler
	RequireAdmin(next http.Handler) http.Handler
}

func NewRouter(handlers *RouterHandlers) *Router {
	r := chi.NewRouter()

	if handlers == nil {
		handlers = &RouterHandlers{}
	}

	logger := handlers.Logger
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// Global middleware chain
	setupMiddleware(r, handlers, logger)

	// Infrastructure routes (health, ready, metrics, CSP)
	registerInfrastructureRoutes(r, handlers, logger)

	// Auth routes (login, logout, OIDC, forward-auth)
	registerAuthRoutes(r, handlers)

	// Page routes (dashboard, admin, settings, metrics)
	registerPageRoutes(r, handlers)

	// Event web UI routes (requires auth middleware for nested routes)
	if handlers.EventWebHandlers != nil && handlers.AuthMiddleware != nil {
		r.Route("/events", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return handlers.AuthMiddleware.RequireAuth(next)
			})

			r.Get("/", handlers.EventWebHandlers.ListEventsPage)
			r.Get("/new", handlers.EventWebHandlers.NewEventForm)
			r.Post("/", handlers.EventWebHandlers.CreateEventFromForm)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handlers.EventWebHandlers.GetEventPage)
				r.Get("/edit", handlers.EventWebHandlers.EditEventForm)
				if handlers.CustomizationHandlers != nil {
					r.Get("/customize", handlers.CustomizationHandlers.CustomizationPage)
					r.Post("/customize", handlers.CustomizationHandlers.CustomizationPage)
				}
			})
		})
	}

	// Invite web UI routes (separate from /events to match original path pattern)
	if handlers.InviteWebHandlers != nil && handlers.AuthMiddleware != nil {
		r.Route("/events/{eventId}/invites", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return handlers.AuthMiddleware.RequireAuth(next)
			})
			r.Get("/", handlers.InviteWebHandlers.ListInvitesPage)
		})
	}

	// API routes (/api/*)
	registerAPIRoutes(r, handlers)

	// Template editor
	if handlers.TemplateEditorHandlers != nil {
		handlers.TemplateEditorHandlers.RegisterRoutes(r)
	}

	// RSVP routes (public, no auth)
	registerRSVPRoutes(r, handlers)

	// Static files and assets
	registerStaticRoutes(r, handlers)

	return &Router{
		mux:      r,
		handlers: handlers,
	}
}
func (router *Router) ListRoutes() []RouteInfo {
	var routes []RouteInfo

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes = append(routes, RouteInfo{
			Method:  method,
			Pattern: route,
		})
		return nil
	}

	chi.Walk(router.mux, walkFunc)

	return routes
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	HandleError(w, r, NewNotFoundError("Route not found"))
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	err := &APIError{
		StatusCode: http.StatusMethodNotAllowed,
		Code:       "METHOD_NOT_ALLOWED",
		Message:    "Method not allowed",
	}
	HandleError(w, r, err)
}

func IsAPIRequest(r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		return true
	}

	acceptHeader := r.Header.Get("Accept")
	return strings.Contains(acceptHeader, "application/json")
}

func GetInt64Param(value string) (int64, error) {
	if value == "" {
		return 0, fmt.Errorf("%w: empty value", ErrEmptyParameter)
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidParameter, err)
	}

	if id < 0 {
		return 0, fmt.Errorf("%w: negative value not allowed", ErrInvalidParameter)
	}

	return id, nil
}

func GetStringParam(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%w: empty value", ErrEmptyParameter)
	}
	return trimmed, nil
}

func GetEventIDFromRequest(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return GetInt64Param(idStr)
}

func GetTokenFromRequest(r *http.Request) (string, error) {
	token := chi.URLParam(r, "token")
	return GetStringParam(token)
}

const notImplementedHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Coming Soon - TinyRSVP</title>
    <link rel="stylesheet" href="/static/css/base.css">
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/typography.css">
    <link rel="stylesheet" href="/static/css/colors.css">
    <link rel="stylesheet" href="/static/css/spacing.css">
    <link rel="stylesheet" href="/static/css/buttons.css">
    <link rel="stylesheet" href="/static/css/app_navigation.css">
    <link rel="stylesheet" href="/static/css/mobile_optimization.css">
    <link rel="stylesheet" href="/static/css/theme_toggle.css">
</head>
<body>
    <div style="max-width:480px;margin:var(--spacing-16,4rem) auto;text-align:center;padding:var(--spacing-8,2rem)">
        <div style="font-size:64px;margin-bottom:1rem">🚧</div>
        <h1 style="margin-bottom:0.75rem">Coming Soon</h1>
        <p style="color:var(--color-text-secondary,#666);margin-bottom:1.5rem;line-height:1.6">
            This feature is not yet implemented. It&#39;s on the roadmap and will be available in a future release.
        </p>
        <a href="javascript:history.back()" class="btn btn-secondary">Go Back</a>
    </div>
    <script src="/static/js/theme_controller.js"></script>
</body>
</html>`
