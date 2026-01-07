package middleware

import (
	"log/slog"
	"net/http"

	"github.com/lenaxia/tinyrsvp/internal/auth"
)

func RequireAuth(sessionMgr auth.SessionManager, userService auth.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := sessionMgr.GetSessionFromRequest(r)
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Cookie")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			session, err := sessionMgr.GetSession(r.Context(), sessionID)
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Cookie")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := userService.GetUserByID(r.Context(), session.UserID)
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Cookie")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
