package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/lenaxia/tinyrsvp/internal/auth"
)

type Authenticator interface {
	HandleLogin(w http.ResponseWriter, r *http.Request) error
	HandleCallback(w http.ResponseWriter, r *http.Request) (*auth.AuthResult, error)
	HandleLogout(w http.ResponseWriter, r *http.Request) error
}

type AuthHandlers struct {
	authenticator Authenticator
	userService   auth.UserService
	sessionMgr    auth.SessionManager
}

func NewAuthHandlers(authenticator Authenticator, userService auth.UserService, sessionMgr auth.SessionManager) *AuthHandlers {
	return &AuthHandlers{
		authenticator: authenticator,
		userService:   userService,
		sessionMgr:    sessionMgr,
	}
}

func (h *AuthHandlers) ShowLogin(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Query().Get("return")

	validatedURL, err := auth.ValidateReturnURL(returnURL)
	if err != nil {
		HandleError(w, r, NewBadRequestError("Invalid return URL"))
		return
	}

	data := struct {
		ReturnURL string
	}{
		ReturnURL: validatedURL,
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - TinyRSVP</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            max-width: 400px;
            margin: 100px auto;
            padding: 20px;
            text-align: center;
        }
        h1 {
            font-size: 32px;
            margin-bottom: 40px;
            color: #333;
        }
        .login-button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #007bff;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            font-size: 16px;
        }
        .login-button:hover {
            background-color: #0056b3;
        }
    </style>
</head>
<body>
    <h1>TinyRSVP</h1>
    <p>Please log in to continue</p>
    <a href="/auth/oidc/login?return={{.ReturnURL}}" class="login-button">Log In with OIDC</a>
</body>
</html>`

	t := template.Must(template.New("login").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	t.Execute(w, data)
}

func (h *AuthHandlers) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Query().Get("return")

	if _, err := auth.ValidateReturnURL(returnURL); err != nil {
		HandleError(w, r, NewBadRequestError("Invalid return URL"))
		return
	}

	if err := h.authenticator.HandleLogin(w, r); err != nil {
		HandleError(w, r, NewInternalError())
		return
	}
}

func (h *AuthHandlers) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	result, err := h.authenticator.HandleCallback(w, r)
	if err != nil || result == nil {
		HandleError(w, r, NewUnauthorizedError("Authentication failed"))
		return
	}

	user, err := h.userService.GetOrCreateUser(r.Context(), result.Email, result.Name, result.OIDCSubject)
	if err != nil {
		HandleError(w, r, NewInternalError())
		return
	}

	if err := h.userService.UpdateLastLogin(r.Context(), user.ID); err != nil {
		slog.Warn("failed to update last login", "user_id", user.ID, "error", err)
	}

	session, err := h.sessionMgr.CreateSession(r.Context(), user.ID, r)
	if err != nil {
		HandleError(w, r, NewInternalError())
		return
	}

	if err := h.sessionMgr.SetSessionCookie(w, session.ID); err != nil {
		HandleError(w, r, NewInternalError())
		return
	}

	// Resolve the return URL: query parameter takes precedence (forward-auth
	// preserves it), then fall back to the cookie set during login (OIDC
	// preserves it through the provider redirect).
	returnURL := r.URL.Query().Get("return")
	if returnURL == "" {
		if cookie, cookieErr := r.Cookie(auth.ReturnURLCookieName); cookieErr == nil {
			returnURL = cookie.Value
		}
	}
	validatedURL, err := auth.ValidateReturnURL(returnURL)
	if err != nil {
		validatedURL = "/"
	}

	// Clear the return URL cookie so it cannot be replayed.
	http.SetCookie(w, &http.Cookie{
		Name:     auth.ReturnURLCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, validatedURL, http.StatusFound)
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HandleError(w, r, &APIError{
			StatusCode: http.StatusMethodNotAllowed,
			Code:       "METHOD_NOT_ALLOWED",
			Message:    "Method not allowed",
		})
		return
	}

	if err := h.authenticator.HandleLogout(w, r); err != nil {
		HandleError(w, r, NewInternalError())
		return
	}

	h.sessionMgr.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}
