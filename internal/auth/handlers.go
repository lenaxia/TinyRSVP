package auth

import (
	"log"
	"net/http"
)

type LoginHandler struct {
	auth Authenticator
}

func NewLoginHandler(auth Authenticator) http.Handler {
	return &LoginHandler{auth: auth}
}

// headerWrittenResponseWriter tracks whether WriteHeader/Write was called,
// so the caller can decide whether to write its own response.
type headerWrittenResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (rw *headerWrittenResponseWriter) WriteHeader(code int) {
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *headerWrittenResponseWriter) Write(b []byte) (int, error) {
	rw.wroteHeader = true
	return rw.ResponseWriter.Write(b)
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Query().Get("return")
	validatedURL, err := ValidateReturnURL(returnURL)
	if err != nil {
		log.Printf("Invalid return URL rejected: %s (error: %v)", returnURL, err)
		validatedURL = "/"
	}

	// Persist the return URL in a short-lived cookie so it survives the
	// external OIDC provider redirect (the provider only echoes back code
	// and state, dropping any custom query parameters).
	http.SetCookie(w, &http.Cookie{
		Name:     ReturnURLCookieName,
		Value:    validatedURL,
		Path:     "/",
		MaxAge:   int(ReturnURLMaxAge.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Wrap the response writer to detect whether HandleLogin already wrote a
	// response (OIDC redirects to the provider) vs. left it untouched
	// (forward-auth creates the session and returns without redirecting).
	rw := &headerWrittenResponseWriter{ResponseWriter: w}
	if err := h.auth.HandleLogin(rw, r); err != nil {
		log.Printf("Login error: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Forward-auth case: HandleLogin did not redirect, so issue the redirect
	// to the return URL now. For OIDC, HandleLogin already wrote the
	// provider redirect — issuing a second redirect would corrupt the
	// response.
	if !rw.wroteHeader {
		http.Redirect(w, r, validatedURL, http.StatusFound)
	}
}

type CallbackHandler struct {
	auth        Authenticator
	userService UserService
	sessionMgr  SessionManager
}

func NewCallbackHandler(auth Authenticator, userService UserService, sessionMgr SessionManager) http.Handler {
	return &CallbackHandler{
		auth:        auth,
		userService: userService,
		sessionMgr:  sessionMgr,
	}
}

func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.auth.HandleCallback(w, r)
	if err != nil {
		log.Printf("Callback error: %v", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetOrCreateUser(r.Context(), result.Email, result.Name, result.OIDCSubject)
	if err != nil {
		log.Printf("User creation error: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if err := h.userService.UpdateLastLogin(r.Context(), user.ID); err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	session, err := h.sessionMgr.CreateSession(r.Context(), user.ID, r)
	if err != nil {
		log.Printf("Session creation error: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	if err := h.sessionMgr.SetSessionCookie(w, session.ID); err != nil {
		log.Printf("Cookie setting error: %v", err)
		http.Error(w, "Failed to set session cookie", http.StatusInternalServerError)
		return
	}

	// Resolve the return URL: query parameter takes precedence (forward-auth
	// preserves it), then fall back to the cookie set during login (OIDC
	// preserves it through the provider redirect).
	returnURL := r.URL.Query().Get("return")
	if returnURL == "" {
		if cookie, cookieErr := r.Cookie(ReturnURLCookieName); cookieErr == nil {
			returnURL = cookie.Value
		}
	}
	validatedURL, err := ValidateReturnURL(returnURL)
	if err != nil {
		log.Printf("Invalid return URL rejected: %s (error: %v)", returnURL, err)
		validatedURL = "/"
	}

	// Clear the return URL cookie so it cannot be replayed.
	http.SetCookie(w, &http.Cookie{
		Name:     ReturnURLCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, validatedURL, http.StatusFound)
}

type LogoutHandler struct {
	auth Authenticator
}

func NewLogoutHandler(auth Authenticator) http.Handler {
	return &LogoutHandler{auth: auth}
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.auth.HandleLogout(w, r); err != nil {
		log.Printf("Logout error: %v", err)
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
