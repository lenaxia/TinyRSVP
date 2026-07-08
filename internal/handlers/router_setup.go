package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	customMiddleware "github.com/lenaxia/tinyrsvp/internal/middleware"
)

// setupMiddleware configures the global middleware chain on the router.
// Extracted from NewRouter for readability.
func setupMiddleware(r chi.Router, h *RouterHandlers, logger *log.Logger) {
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

	if h.MetricsMiddleware != nil {
		r.Use(h.MetricsMiddleware)
	}
}

// registerInfrastructureRoutes sets up health, readiness, metrics, and CSP
// reporting endpoints. These are the non-application routes.
func registerInfrastructureRoutes(r chi.Router, h *RouterHandlers, logger *log.Logger) {
	r.NotFound(NotFoundHandler)
	r.MethodNotAllowed(MethodNotAllowedHandler)

	r.Handle("/api/csp-report", customMiddleware.CSPReportHandler(logger))

	if h.HealthHandler != nil {
		r.Handle("/health", h.HealthHandler)
	} else {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})
	}

	if h.ReadinessHandler != nil {
		r.Handle("/ready", h.ReadinessHandler)
	}

	if h.MetricsHandler != nil {
		r.Handle("/metrics", h.MetricsHandler)
	}
}

// registerAuthRoutes sets up login, logout, OIDC, and forward-auth routes.
func registerAuthRoutes(r chi.Router, h *RouterHandlers) {
	if h.AuthHandlers != nil {
		r.Get("/login", h.AuthHandlers.ShowLogin)
		r.Get("/auth/oidc/login", h.AuthHandlers.OIDCLogin)
		r.Get("/auth/oidc/callback", h.AuthHandlers.OIDCCallback)
		r.Post("/logout", h.AuthHandlers.Logout)
	} else if h.LoginHandler != nil {
		r.Handle("/login", h.LoginHandler)
		r.Get("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		if h.CallbackHandler != nil {
			r.Handle("/auth/callback", h.CallbackHandler)
		} else {
			r.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		}
		if h.LogoutHandler != nil {
			r.Handle("/logout", h.LogoutHandler)
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
}

// registerPageRoutes sets up the dashboard, admin, settings, and metrics
// web pages.
func registerPageRoutes(r chi.Router, h *RouterHandlers) {
	r.Get("/not-implemented", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, notImplementedHTML)
	})

	if h.DashboardHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/", h.AuthMiddleware.RequireAuth(
			http.HandlerFunc(h.DashboardHandler.Dashboard),
		))
	} else {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}

	if h.AdminDashboardHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/admin", h.AuthMiddleware.RequireAuth(
			h.AuthMiddleware.RequireAdmin(
				http.HandlerFunc(h.AdminDashboardHandler.AdminDashboard),
			),
		))
	}

	if h.UserManagementHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/admin/users", h.AuthMiddleware.RequireAuth(
			h.AuthMiddleware.RequireAdmin(
				http.HandlerFunc(h.UserManagementHandler.UserManagementPage),
			),
		))
	}

	if h.SettingsHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/admin/settings", h.AuthMiddleware.RequireAuth(
			h.AuthMiddleware.RequireAdmin(
				http.HandlerFunc(h.SettingsHandler.SettingsPage),
			),
		))
	}

	if h.AdminMetricsHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/admin/metrics", h.AuthMiddleware.RequireAuth(
			h.AuthMiddleware.RequireAdmin(
				http.HandlerFunc(h.AdminMetricsHandler.MetricsPage),
			),
		))
	}
}

// registerAPIRoutes sets up the /api/* routes including events, invites,
// images, templates, users, and admin endpoints.
func registerAPIRoutes(r chi.Router, h *RouterHandlers) {
	apiRouter := chi.NewRouter()

	if h.AuthMiddleware != nil {
		apiRouter.Use(func(next http.Handler) http.Handler {
			return h.AuthMiddleware.RequireAuth(next)
		})
	}

	if h.EventHandlers != nil {
		apiRouter.Route("/events", func(r chi.Router) {
			r.Get("/", h.EventHandlers.ListEvents)
			r.Post("/", h.EventHandlers.CreateEvent)
			r.Get("/{id}", h.EventHandlers.GetEvent)
			r.Put("/{id}", h.EventHandlers.UpdateEvent)
			r.Delete("/{id}", h.EventHandlers.DeleteEvent)
		})
	} else {
		apiRouter.Route("/events", func(r chi.Router) {
			r.Get("/", unauthorized)
			r.Post("/", unauthorized)
			r.Get("/{id}", unauthorized)
			r.Put("/{id}", unauthorized)
			r.Delete("/{id}", unauthorized)
		})
	}

	if h.QuestionHandlers != nil {
		h.QuestionHandlers.RegisterRoutes(apiRouter)
	}

	apiRouter.Route("/events/{eventId}/invites", func(r chi.Router) {
		if h.ListInviteHandlers != nil {
			r.Get("/", h.ListInviteHandlers.ListInvites)
		}
		if h.InviteHandlers != nil {
			r.Post("/", h.InviteHandlers.CreateInvite)
		}
		if h.ImportInviteHandlers != nil {
			r.Post("/import", h.ImportInviteHandlers.ImportInvites)
		}
		if h.ManualInviteHandlers != nil {
			r.Post("/manual", h.ManualInviteHandlers.CreateManualInvite)
		}
	})

	apiRouter.Route("/invites/{inviteId}", func(r chi.Router) {
		if h.GetInviteHandlers != nil {
			r.Get("/", h.GetInviteHandlers.GetInvite)
		}
		if h.UpdateInviteHandlers != nil {
			r.Put("/", h.UpdateInviteHandlers.UpdateInvite)
		}
		if h.DeleteInviteHandlers != nil {
			r.Delete("/", h.DeleteInviteHandlers.DeleteInvite)
		}
		if h.RevokeInviteHandlers != nil {
			r.Post("/revoke", h.RevokeInviteHandlers.RevokeInvite)
		}
		if h.RegenerateInviteHandlers != nil {
			r.Post("/regenerate", h.RegenerateInviteHandlers.RegenerateInviteToken)
		}
		if h.SendInviteHandlers != nil {
			r.Post("/send", h.SendInviteHandlers.SendInvite)
		}
	})

	if h.ImageHandlers != nil {
		h.ImageHandlers.RegisterRoutes(apiRouter)
	}

	if h.RSVPSummaryHandler != nil {
		apiRouter.Get("/events/{id}/rsvp-summary", h.RSVPSummaryHandler.GetRSVPSummary)
	}

	if h.TemplateHandlers != nil {
		h.TemplateHandlers.RegisterRoutes(apiRouter)
	}

	if h.CustomizationHandlers != nil {
		h.CustomizationHandlers.RegisterRoutes(apiRouter)
	}

	if h.UserHandler != nil && h.AuthMiddleware != nil {
		r.Route("/api/users", func(r chi.Router) {
			r.Use(h.AuthMiddleware.RequireAuth)
			r.Use(h.AuthMiddleware.RequireAdmin)

			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodPost {
						if err := r.ParseForm(); err == nil {
							if override := r.FormValue("_method"); override != "" {
								r.Method = strings.ToUpper(override)
							}
						}
					}
					next.ServeHTTP(w, r)
				})
			})

			r.Get("/", h.UserHandler.ListUsers)
			r.Get("/{id}", h.UserHandler.GetUser)
			r.Patch("/{id}", h.UserHandler.UpdateUserRole)
			r.Delete("/{id}", h.UserHandler.DeleteUser)
		})
	}

	if h.CleanupHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/api/invites/cleanup", h.AuthMiddleware.RequireAuth(
			h.AuthMiddleware.RequireAdmin(h.CleanupHandler),
		))
	}

	if h.EmailHealthHandler != nil && h.AuthMiddleware != nil {
		r.Handle("/api/email/health", h.AuthMiddleware.RequireAuth(
			h.AuthMiddleware.RequireAdmin(h.EmailHealthHandler),
		))
	}

	if h.CleanupHandler == nil {
		apiRouter.Post("/invites/cleanup", unauthorized)
	}

	r.Mount("/api", apiRouter)
}

// registerRSVPRoutes sets up the RSVP, confirmation, unsubscribe, and
// calendar endpoints.
func registerRSVPRoutes(r chi.Router, h *RouterHandlers) {
	if h.RSVPHandler != nil {
		r.Route("/rsvp/{token}", func(r chi.Router) {
			r.Get("/", h.RSVPHandler.GetRSVPPage)
			r.Post("/", h.RSVPHandler.SubmitRSVP)
			r.Put("/", h.RSVPHandler.UpdateRSVP)
			r.Get("/confirmation", h.RSVPHandler.GetConfirmationPage)
		})
		r.Get("/unsubscribe/{token}", h.RSVPHandler.Unsubscribe)
		r.Get("/api/calendar/{token}", h.RSVPHandler.GetCalendar)
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
}

// registerStaticRoutes sets up static file serving and asset handling.
func registerStaticRoutes(r chi.Router, h *RouterHandlers) {
	if h.AssetHandler != nil {
		r.HandleFunc("/assets/*", h.AssetHandler.ServeAsset)
	}

	if h.StaticFileServer != nil {
		r.Handle("/static/*", h.StaticFileServer)
	} else {
		r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
		})
	}
}

// unauthorized is a helper that returns 401 for stub routes when handlers
// are not wired.
func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}
