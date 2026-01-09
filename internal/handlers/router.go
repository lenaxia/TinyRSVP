package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrEmptyParameter   = errors.New("parameter cannot be empty")
)

type Router struct {
	mux      chi.Router
	handlers interface{}
}

func NewRouter(handlers interface{}) *Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.NotFound(NotFoundHandler)
	r.MethodNotAllowed(MethodNotAllowedHandler)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/events", func(r chi.Router) {
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

		r.Route("/invites", func(r chi.Router) {
			r.Post("/cleanup", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
		})

		r.Route("/templates", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})
		})
	})

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

	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return &Router{
		mux:      r,
		handlers: handlers,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	if IsAPIRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Route not found",
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>404 - Not Found</title>
    <style>
        body { font-family: sans-serif; text-align: center; padding: 50px; }
        h1 { color: #333; }
    </style>
</head>
<body>
    <h1>404 - Page Not Found</h1>
    <p>The page you are looking for does not exist.</p>
</body>
</html>`)
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	if IsAPIRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Method not allowed",
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>405 - Method Not Allowed</title>
    <style>
        body { font-family: sans-serif; text-align: center; padding: 50px; }
        h1 { color: #333; }
    </style>
</head>
<body>
    <h1>405 - Method Not Allowed</h1>
    <p>The HTTP method used is not allowed for this resource.</p>
</body>
</html>`)
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

func GetInviteIDFromRequest(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "inviteId")
	return GetInt64Param(idStr)
}

func GetTokenFromRequest(r *http.Request) (string, error) {
	token := chi.URLParam(r, "token")
	return GetStringParam(token)
}

func GetUserIDFromRequest(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "userId")
	return GetInt64Param(idStr)
}
