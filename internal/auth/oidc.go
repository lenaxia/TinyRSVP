package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/yourusername/tinyrsvp/internal/models"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type Authenticator interface {
	HandleLogin(w http.ResponseWriter, r *http.Request) error
	HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error)
	HandleLogout(w http.ResponseWriter, r *http.Request) error
}

type AuthResult struct {
	Email       string
	Name        string
	OIDCSubject *string
}

type UserService interface {
	CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
	GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type SessionManager interface {
	CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	RefreshSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteUserSessions(ctx context.Context, userID int64) error
	CleanupExpired(ctx context.Context) (int64, error)
	SetSessionCookie(w http.ResponseWriter, sessionID string) error
	ClearSessionCookie(w http.ResponseWriter) error
	GetSessionFromRequest(r *http.Request) (string, error)
}

type oidcAuthenticator struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	userService  UserService
	sessionMgr   SessionManager
}

func NewOIDCAuthenticator(cfg *OIDCConfig, userService UserService, sessionMgr SessionManager) (Authenticator, error) {
	return NewOIDCAuthenticatorWithContext(context.Background(), cfg, userService, sessionMgr)
}

func NewOIDCAuthenticatorWithContext(ctx context.Context, cfg *OIDCConfig, userService UserService, sessionMgr SessionManager) (Authenticator, error) {
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}

	if cfg.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL is required")
	}

	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "email", "profile"}
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	return &oidcAuthenticator{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		userService:  userService,
		sessionMgr:   sessionMgr,
	}, nil
}

func (a *oidcAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	state, err := generateRandomState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	authURL := a.oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
	return nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (a *oidcAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	ctx := r.Context()

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		return nil, fmt.Errorf("missing state cookie: %w", err)
	}

	queryState := r.URL.Query().Get("state")
	if queryState != stateCookie.Value {
		return nil, fmt.Errorf("state mismatch")
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("missing authorization code")
	}

	oauth2Token, err := a.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing id_token in response")
	}

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("missing email claim")
	}

	subject := idToken.Subject

	return &AuthResult{
		Email:       claims.Email,
		Name:        claims.Name,
		OIDCSubject: &subject,
	}, nil
}

func (a *oidcAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) error {
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
