package middleware

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lenaxia/tinyrsvp/internal/auth"
)

// TestRequireAuth wraps RequireAuth with an X-Test-User-ID header bypass
// for use in test server setups ONLY. Production code must use RequireAuth
// directly.
//
// The bypass checks for the X-Test-User-ID header. If present and valid,
// it loads the user from the database and sets it in the request context,
// skipping session validation. If the header is absent or invalid, it
// falls through to normal session-based authentication.
//
// This function exists so that browser-based UX tests (Playwright) and
// integration tests can authenticate as a specific user without going
// through the full OIDC or forward-auth flow. It MUST NOT be wired into
// the production router.
func TestRequireAuth(sessionMgr auth.SessionManager, userService auth.UserService) func(http.Handler) http.Handler {
	realAuth := RequireAuth(sessionMgr, userService)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if testUserID := r.Header.Get("X-Test-User-ID"); testUserID != "" {
				userID, err := strconv.ParseInt(testUserID, 10, 64)
				if err != nil {
					slog.Warn("Invalid X-Test-User-ID header, falling through to normal auth", "value", testUserID, "error", err)
				} else {
					user, err := userService.GetUserByID(r.Context(), userID)
					if err != nil {
						slog.Warn("Failed to get test user, falling through to normal auth", "user_id", userID, "error", err)
					} else {
						ctx := auth.WithUser(r.Context(), user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			realAuth(next).ServeHTTP(w, r)
		})
	}
}
