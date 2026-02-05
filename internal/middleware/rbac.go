package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/lenaxia/tinyrsvp/internal/auth"
)

func RequireAuth(sessionMgr auth.SessionManager, userService auth.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Test bypass: Allow tests to bypass authentication with X-Test-User header
			if testUserID := r.Header.Get("X-Test-User-ID"); testUserID != "" {
				// Convert test user ID to int64
				userID, err := strconv.ParseInt(testUserID, 10, 64)
				if err != nil {
					slog.Warn("Invalid X-Test-User-ID header, falling through to normal auth", "value", testUserID, "error", err)
				} else {
					// Get test user from database
					user, err := userService.GetUserByID(r.Context(), userID)
					if err != nil {
						slog.Warn("Failed to get test user, falling through to normal auth", "user_id", userID, "error", err)
					} else {
						// Create test session context
						ctx := auth.WithUser(r.Context(), user)
						slog.Debug("Test authentication bypass enabled", "user_id", userID, "user_email", user.Email)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Normal authentication flow
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
