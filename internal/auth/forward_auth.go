package auth

import (
	"fmt"
	"net/http"
	"strings"
)

type ForwardAuthConfig struct {
	UserHeader  string
	EmailHeader string
	TrustedIPs  []string
}

type forwardAuthenticator struct {
	config      *ForwardAuthConfig
	userService UserService
	sessionMgr  SessionManager
}

func NewForwardAuthenticator(cfg *ForwardAuthConfig, userService UserService, sessionMgr SessionManager) Authenticator {
	return &forwardAuthenticator{
		config:      cfg,
		userService: userService,
		sessionMgr:  sessionMgr,
	}
}

func (a *forwardAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	authResult, err := a.HandleCallback(w, r)
	if err != nil {
		return err
	}

	user, err := a.userService.GetOrCreateUser(r.Context(), authResult.Email, authResult.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to get or create user: %w", err)
	}

	session, err := a.sessionMgr.CreateSession(r.Context(), user.ID, r)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if err := a.sessionMgr.SetSessionCookie(w, session.ID); err != nil {
		return fmt.Errorf("failed to set session cookie: %w", err)
	}

	return nil
}

func (a *forwardAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	if err := a.validateTrustedProxy(r); err != nil {
		return nil, fmt.Errorf("untrusted proxy: %w", err)
	}

	username := r.Header.Get(a.config.UserHeader)
	if username == "" {
		return nil, fmt.Errorf("missing or empty user header: %s", a.config.UserHeader)
	}

	email := r.Header.Get(a.config.EmailHeader)
	if email == "" {
		return nil, fmt.Errorf("missing or empty email header: %s", a.config.EmailHeader)
	}

	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	name := r.Header.Get("Remote-Name")
	if name == "" {
		name = r.Header.Get("X-authentik-name")
	}
	if name == "" {
		name = username
	}

	return &AuthResult{
		Email:       email,
		Name:        name,
		OIDCSubject: nil,
	}, nil
}

func (a *forwardAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	sessionID, err := a.sessionMgr.GetSessionFromRequest(r)
	if err != nil {
		return nil
	}

	if err := a.sessionMgr.DeleteSession(r.Context(), sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	a.sessionMgr.ClearSessionCookie(w)
	return nil
}

func (a *forwardAuthenticator) validateTrustedProxy(r *http.Request) error {
	clientIP := getClientIP(r)
	if clientIP == "" {
		return fmt.Errorf("unable to determine client IP")
	}

	for _, trustedIP := range a.config.TrustedIPs {
		if clientIP == trustedIP {
			return nil
		}
	}

	return fmt.Errorf("request from untrusted IP: %s", clientIP)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if parts[0] == "" || parts[1] == "" {
		return false
	}

	return true
}
