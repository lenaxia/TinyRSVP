package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	customMiddleware "github.com/lenaxia/tinyrsvp/internal/middleware"
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

	CleanupHandler     http.Handler
	EmailHealthHandler http.Handler
	MetricsHandler     http.Handler

	AuthMiddleware AuthMiddlewareInterface

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
	Unsubscribe(w http.ResponseWriter, r *http.Request)
}

type RSVPSummaryHandlerInterface interface {
	GetRSVPSummary(w http.ResponseWriter, r *http.Request)
}

type UserHandlerInterface interface {
	ListUsers(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request, userID string)
	UpdateUserRole(w http.ResponseWriter, r *http.Request, userID string)
	DeleteUser(w http.ResponseWriter, r *http.Request, userID string)
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

	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.Recovery(next)
	})
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.RequestID(next)
	})
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.RealIP(next)
	})
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.Logging(logger)(next)
	})
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.Timeout(30 * time.Second)(next)
	})
	hstsMaxAge := 31536000
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.SecurityHeaders(&customMiddleware.SecurityHeadersConfig{
			HSTSMaxAge: &hstsMaxAge,
		})(next)
	})
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.CSRF(32)(next)
	})

	rateLimiter := customMiddleware.NewRateLimiter(customMiddleware.RateLimiterConfig{
		RequestsPerMinute: 300,
		BurstSize:         300,
	})
	r.Use(func(next http.Handler) http.Handler {
		return customMiddleware.RateLimit(rateLimiter, customMiddleware.RateLimitConfig{
			AnonymousLimit:     300,
			AuthenticatedLimit: 900,
			AdminLimit:         3000,
		})(next)
	})

	r.NotFound(NotFoundHandler)
	r.MethodNotAllowed(MethodNotAllowedHandler)

	r.Handle("/api/csp-report", customMiddleware.CSPReportHandler(logger))

	if handlers.HealthHandler != nil {
		r.Handle("/health", handlers.HealthHandler)
	} else {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})
	}

	if handlers.ReadinessHandler != nil {
		r.Handle("/ready", handlers.ReadinessHandler)
	}

	if handlers.MetricsHandler != nil {
		r.Handle("/metrics", handlers.MetricsHandler)
	}

	if handlers.AuthHandlers != nil {
		r.Get("/login", handlers.AuthHandlers.ShowLogin)
		r.Get("/auth/oidc/login", handlers.AuthHandlers.OIDCLogin)
		r.Get("/auth/oidc/callback", handlers.AuthHandlers.OIDCCallback)
		r.Post("/logout", handlers.AuthHandlers.Logout)
	} else if handlers.LoginHandler != nil {
		r.Handle("/login", handlers.LoginHandler)
		r.Get("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		if handlers.CallbackHandler != nil {
			r.Handle("/auth/callback", handlers.CallbackHandler)
		} else {
			r.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		}

		if handlers.LogoutHandler != nil {
			r.Handle("/logout", handlers.LogoutHandler)
		} else {
			r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		}
	} else {
		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}

	if handlers.DashboardHandler != nil && handlers.AuthMiddleware != nil {
		r.Handle("/", handlers.AuthMiddleware.RequireAuth(
			http.HandlerFunc(handlers.DashboardHandler.Dashboard),
		))
	} else {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}

	if handlers.AdminDashboardHandler != nil && handlers.AuthMiddleware != nil {
		r.Handle("/admin", handlers.AuthMiddleware.RequireAuth(
			handlers.AuthMiddleware.RequireAdmin(
				http.HandlerFunc(handlers.AdminDashboardHandler.AdminDashboard),
			),
		))
	}

	if handlers.UserManagementHandler != nil && handlers.AuthMiddleware != nil {
		r.Handle("/admin/users", handlers.AuthMiddleware.RequireAuth(
			handlers.AuthMiddleware.RequireAdmin(
				http.HandlerFunc(handlers.UserManagementHandler.UserManagementPage),
			),
		))
	}

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
				}
				r.Post("/", handlers.EventWebHandlers.UpdateEventFromForm)
				r.Post("/publish", handlers.EventWebHandlers.PublishEventAction)
				r.Post("/cancel", handlers.EventWebHandlers.CancelEventAction)
				r.Post("/delete", handlers.EventWebHandlers.DeleteEventAction)
			})
		})
	}

	if handlers.InviteWebHandlers != nil && handlers.AuthMiddleware != nil {
		r.Route("/events/{eventId}/invites", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return handlers.AuthMiddleware.RequireAuth(next)
			})

			r.Get("/", handlers.InviteWebHandlers.ListInvitesPage)
		})
	}

	apiRouter := chi.NewRouter()
	if handlers.AuthMiddleware != nil {
		apiRouter.Use(func(next http.Handler) http.Handler {
			return handlers.AuthMiddleware.RequireAuth(next)
		})
	}

	if handlers.EventHandlers != nil {
		apiRouter.Route("/events", func(r chi.Router) {
			r.Get("/", handlers.EventHandlers.ListEvents)
			r.Post("/", handlers.EventHandlers.CreateEvent)
			r.Get("/{id}", handlers.EventHandlers.GetEvent)
			r.Put("/{id}", handlers.EventHandlers.UpdateEvent)
			r.Delete("/{id}", handlers.EventHandlers.DeleteEvent)
		})
	} else {
		apiRouter.Route("/events", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
			r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
		})
	}

	if handlers.QuestionHandlers != nil {
		handlers.QuestionHandlers.RegisterRoutes(apiRouter)
	}

	apiRouter.Route("/events/{eventId}/invites", func(r chi.Router) {
		if handlers.ListInviteHandlers != nil {
			r.Get("/", handlers.ListInviteHandlers.ListInvites)
		}
		if handlers.InviteHandlers != nil {
			r.Post("/", handlers.InviteHandlers.CreateInvite)
		}
		if handlers.ImportInviteHandlers != nil {
			r.Post("/import", handlers.ImportInviteHandlers.ImportInvites)
		}
		if handlers.ManualInviteHandlers != nil {
			r.Post("/manual", handlers.ManualInviteHandlers.CreateManualInvite)
		}
	})

	apiRouter.Route("/invites/{inviteId}", func(r chi.Router) {
		if handlers.GetInviteHandlers != nil {
			r.Get("/", handlers.GetInviteHandlers.GetInvite)
		}
		if handlers.UpdateInviteHandlers != nil {
			r.Put("/", handlers.UpdateInviteHandlers.UpdateInvite)
		}
		if handlers.DeleteInviteHandlers != nil {
			r.Delete("/", handlers.DeleteInviteHandlers.DeleteInvite)
		}
		if handlers.RevokeInviteHandlers != nil {
			r.Post("/revoke", handlers.RevokeInviteHandlers.RevokeInvite)
		}
		if handlers.RegenerateInviteHandlers != nil {
			r.Post("/regenerate", handlers.RegenerateInviteHandlers.RegenerateInviteToken)
		}
		if handlers.SendInviteHandlers != nil {
			r.Post("/send", handlers.SendInviteHandlers.SendInvite)
		}
	})

	if handlers.ImageHandlers != nil {
		handlers.ImageHandlers.RegisterRoutes(apiRouter)
	}

	if handlers.RSVPSummaryHandler != nil {
		apiRouter.Get("/events/{id}/rsvp-summary", handlers.RSVPSummaryHandler.GetRSVPSummary)
	}

	if handlers.TemplateHandlers != nil {
		handlers.TemplateHandlers.RegisterRoutes(apiRouter)
	}

	if handlers.CustomizationHandlers != nil {
		handlers.CustomizationHandlers.RegisterRoutes(apiRouter)
	}

	if handlers.UserHandler != nil {
		if handlers.AuthMiddleware != nil {
			r.Handle("/api/users", handlers.AuthMiddleware.RequireAuth(
				handlers.AuthMiddleware.RequireAdmin(
					http.HandlerFunc(handlers.UserHandler.ListUsers),
				),
			))

			r.HandleFunc("/api/users/", func(w http.ResponseWriter, req *http.Request) {
				path := strings.TrimPrefix(req.URL.Path, "/api/users/")
				if path == "" {
					http.NotFound(w, req)
					return
				}

				parts := strings.Split(path, "/")
				userID := parts[0]

				switch req.Method {
				case http.MethodGet:
					handlers.AuthMiddleware.RequireAuth(
						handlers.AuthMiddleware.RequireAdmin(
							http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								handlers.UserHandler.GetUser(w, r, userID)
							}),
						),
					).ServeHTTP(w, req)
				case http.MethodPatch:
					handlers.AuthMiddleware.RequireAuth(
						handlers.AuthMiddleware.RequireAdmin(
							http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								handlers.UserHandler.UpdateUserRole(w, r, userID)
							}),
						),
					).ServeHTTP(w, req)
				case http.MethodDelete:
					handlers.AuthMiddleware.RequireAuth(
						handlers.AuthMiddleware.RequireAdmin(
							http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								handlers.UserHandler.DeleteUser(w, r, userID)
							}),
						),
					).ServeHTTP(w, req)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			})
		}
	}

	if handlers.CleanupHandler != nil && handlers.AuthMiddleware != nil {
		r.Handle("/api/invites/cleanup", handlers.AuthMiddleware.RequireAuth(
			handlers.AuthMiddleware.RequireAdmin(handlers.CleanupHandler),
		))
	}

	if handlers.EmailHealthHandler != nil && handlers.AuthMiddleware != nil {
		r.Handle("/api/email/health", handlers.AuthMiddleware.RequireAuth(
			handlers.AuthMiddleware.RequireAdmin(handlers.EmailHealthHandler),
		))
	}

	if handlers.CleanupHandler == nil {
		apiRouter.Post("/invites/cleanup", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
	}

	r.Mount("/api", apiRouter)

	if handlers.TemplateEditorHandlers != nil {
		handlers.TemplateEditorHandlers.RegisterRoutes(r)
	}

	if handlers.RSVPHandler != nil {
		r.Route("/rsvp/{token}", func(r chi.Router) {
			r.Get("/", handlers.RSVPHandler.GetRSVPPage)
			r.Post("/", handlers.RSVPHandler.SubmitRSVP)
			r.Put("/", handlers.RSVPHandler.UpdateRSVP)
			r.Get("/confirmation", handlers.RSVPHandler.GetConfirmationPage)
		})
		r.Get("/unsubscribe/{token}", handlers.RSVPHandler.Unsubscribe)
	} else {
		r.Route("/rsvp/{token}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Put("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Get("/confirmation", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
		r.Get("/unsubscribe/{token}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}

	if handlers.AssetHandler != nil {
		r.HandleFunc("/assets/*", handlers.AssetHandler.ServeAsset)
	}

	if handlers.StaticFileServer != nil {
		r.Handle("/static/*", handlers.StaticFileServer)
	} else {
		r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
		})
	}

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
