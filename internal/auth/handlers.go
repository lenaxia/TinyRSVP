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

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.HandleLogin(w, r); err != nil {
		log.Printf("Login error: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
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

	http.Redirect(w, r, "/dashboard", http.StatusFound)
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
