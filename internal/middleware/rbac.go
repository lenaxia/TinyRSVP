package middleware

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/lenaxia/tinyrsvp/internal/auth"
)

func RequireAuth(sessionMgr auth.SessionManager, userService auth.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := sessionMgr.GetSessionFromRequest(r)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			session, err := sessionMgr.GetSession(r.Context(), sessionID)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			user, err := userService.GetUserByID(r.Context(), session.UserID)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			if err := sessionMgr.RefreshSession(r.Context(), sessionID); err != nil {
				slog.Warn("Failed to refresh session, continuing with request", "error", err, "session_id", sessionID)
			}

			ctx := auth.WithUser(r.Context(), user)
			ctx = auth.WithSession(ctx, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Path
	if r.URL.RawQuery != "" {
		returnURL += "?" + r.URL.RawQuery
	}

	http.Redirect(w, r, "/login?return="+url.QueryEscape(returnURL), http.StatusSeeOther)
}

func RequireAdmin(authChecker auth.AuthorizationChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok {
				w.Header().Set("WWW-Authenticate", "Cookie")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authChecker.IsAdmin(user) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireEventManager(authChecker auth.AuthorizationChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok {
				w.Header().Set("WWW-Authenticate", "Cookie")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authChecker.IsEventManager(user) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
