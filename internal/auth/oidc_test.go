package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/tinyrsvp/internal/models"
)

func TestNewOIDCAuthenticator_ValidConfig(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if auth == nil {
		t.Fatal("Expected authenticator, got nil")
	}
}

func TestNewOIDCAuthenticator_InvalidIssuerURL(t *testing.T) {
	cfg := &OIDCConfig{
		IssuerURL:    "http://invalid-url-that-does-not-exist.local",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	auth, err := NewOIDCAuthenticatorWithContext(ctx, cfg, mockUserService, mockSessionMgr)
	if err == nil {
		t.Fatal("Expected error for invalid issuer URL, got nil")
	}

	if auth != nil {
		t.Fatal("Expected nil authenticator for invalid config")
	}
}

func TestNewOIDCAuthenticator_EmptyClientID(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err == nil {
		t.Fatal("Expected error for empty client ID, got nil")
	}

	if auth != nil {
		t.Fatal("Expected nil authenticator for invalid config")
	}
}

func TestNewOIDCAuthenticator_EmptyClientSecret(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err == nil {
		t.Fatal("Expected error for empty client secret, got nil")
	}

	if auth != nil {
		t.Fatal("Expected nil authenticator for invalid config")
	}
}

func TestNewOIDCAuthenticator_EmptyRedirectURL(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err == nil {
		t.Fatal("Expected error for empty redirect URL, got nil")
	}

	if auth != nil {
		t.Fatal("Expected nil authenticator for invalid config")
	}
}

func setupMockOIDCProvider(t *testing.T) *httptest.Server {
	t.Helper()

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			discoveryDoc := fmt.Sprintf(`{
				"issuer": "%s",
				"authorization_endpoint": "%s/authorize",
				"token_endpoint": "%s/token",
				"jwks_uri": "%s/jwks",
				"response_types_supported": ["code"],
				"subject_types_supported": ["public"],
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, serverURL, serverURL, serverURL, serverURL)
			w.Write([]byte(discoveryDoc))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	serverURL = server.URL
	return server
}

type MockUserService struct {
	CreateUserFunc      func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
	GetOrCreateUserFunc func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
	GetUserByIDFunc     func(ctx context.Context, id int64) (*models.User, error)
	GetUserByEmailFunc  func(ctx context.Context, email string) (*models.User, error)
	UpdateUserFunc      func(ctx context.Context, user *models.User) error
	UpdateUserRoleFunc  func(ctx context.Context, userID int64, role models.UserRole) error
	DeleteUserFunc      func(ctx context.Context, id int64) error
	ListUsersFunc       func(ctx context.Context, limit, offset int) ([]*models.User, error)
	CountUsersFunc      func(ctx context.Context) (int, error)
	CountAdminsFunc     func(ctx context.Context) (int, error)
}

func (m *MockUserService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, email, name, oidcSubject)
	}
	return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin}, nil
}

func (m *MockUserService) GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	if m.GetOrCreateUserFunc != nil {
		return m.GetOrCreateUserFunc(ctx, email, name, oidcSubject)
	}
	return &models.User{ID: 1, Email: email, Name: name, Role: models.RoleAdmin}, nil
}

func (m *MockUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return &models.User{ID: id, Email: "test@example.com", Name: "Test User", Role: models.RoleAdmin}, nil
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return &models.User{ID: 1, Email: email, Name: "Test User", Role: models.RoleAdmin}, nil
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockUserService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(ctx, userID, role)
	}
	return nil
}

func (m *MockUserService) DeleteUser(ctx context.Context, id int64) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	return []*models.User{}, nil
}

func (m *MockUserService) CountUsers(ctx context.Context) (int, error) {
	if m.CountUsersFunc != nil {
		return m.CountUsersFunc(ctx)
	}
	return 0, nil
}

func (m *MockUserService) CountAdmins(ctx context.Context) (int, error) {
	if m.CountAdminsFunc != nil {
		return m.CountAdminsFunc(ctx)
	}
	return 1, nil
}

type MockSessionManager struct {
	CreateSessionFunc         func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error)
	GetSessionFunc            func(ctx context.Context, sessionID string) (*models.Session, error)
	RefreshSessionFunc        func(ctx context.Context, sessionID string) error
	DeleteSessionFunc         func(ctx context.Context, sessionID string) error
	DeleteUserSessionsFunc    func(ctx context.Context, userID int64) error
	CleanupExpiredFunc        func(ctx context.Context) (int64, error)
	SetSessionCookieFunc      func(w http.ResponseWriter, sessionID string) error
	ClearSessionCookieFunc    func(w http.ResponseWriter) error
	GetSessionFromRequestFunc func(r *http.Request) (string, error)
}

func (m *MockSessionManager) CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, userID, r)
	}
	return &models.Session{ID: "test-session-id", UserID: userID}, nil
}

func (m *MockSessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(ctx, sessionID)
	}
	return &models.Session{ID: sessionID, UserID: 1}, nil
}

func (m *MockSessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	if m.RefreshSessionFunc != nil {
		return m.RefreshSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockSessionManager) DeleteUserSessions(ctx context.Context, userID int64) error {
	if m.DeleteUserSessionsFunc != nil {
		return m.DeleteUserSessionsFunc(ctx, userID)
	}
	return nil
}

func (m *MockSessionManager) CleanupExpired(ctx context.Context) (int64, error) {
	if m.CleanupExpiredFunc != nil {
		return m.CleanupExpiredFunc(ctx)
	}
	return 0, nil
}

func (m *MockSessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	if m.SetSessionCookieFunc != nil {
		return m.SetSessionCookieFunc(w, sessionID)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "tinyrsvp_session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   604800,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *MockSessionManager) ClearSessionCookie(w http.ResponseWriter) error {
	if m.ClearSessionCookieFunc != nil {
		return m.ClearSessionCookieFunc(w)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "tinyrsvp_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *MockSessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
	if m.GetSessionFromRequestFunc != nil {
		return m.GetSessionFromRequestFunc(r)
	}
	cookie, err := r.Cookie("tinyrsvp_session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
