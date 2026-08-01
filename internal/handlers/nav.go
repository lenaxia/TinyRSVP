package handlers

import (
	"net/http"

	"github.com/lenaxia/tinyrsvp/internal/auth"
)

// isAdminRequest reports whether the authenticated user for r has the admin
// role. Used by page handlers to decide whether to render admin-only UI such
// as the Admin navigation link.
func isAdminRequest(r *http.Request) bool {
	user, ok := auth.UserFromContext(r.Context())
	return ok && user.IsAdmin()
}
