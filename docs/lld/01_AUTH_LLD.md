# Domain 1: Authentication & Authorization - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 4 - Authentication & Authorization](../02_REVISED_HLD.md#4-authentication--authorization)

---

## 1. Overview

### 1.1 Purpose

Provides authentication and authorization for admin and event manager users, including OIDC integration, forward auth support, session management, and role-based access control.

### 1.2 Responsibilities

- OIDC authentication flow (authorization code flow)
- Forward auth header validation
- Database-backed session management
- User creation and role assignment
- Bootstrap admin creation
- Permission checking middleware
- Session cookie management
- Logout handling

### 1.3 Design Principles

- **Interface-Based** - Pluggable authentication providers
- **Secure by Default** - Secure cookies, HTTPS enforcement
- **Stateless Auth** - Session ID in cookie, data in database
- **Fail Closed** - Deny access on error
- **Audit Everything** - Log all auth events

---

## 2. Package Structure

```
internal/
├── auth/
│   ├── auth.go                  # Core auth logic
│   ├── auth_test.go
│   ├── oidc.go                  # OIDC provider implementation
│   ├── oidc_test.go
│   ├── forward.go               # Forward auth implementation
│   ├── forward_test.go
│   ├── session.go               # Session management
│   ├── session_test.go
│   ├── permissions.go           # Permission checking
│   ├── permissions_test.go
│   └── mock_auth.go             # Mock for testing
├── middleware/
│   ├── auth.go                  # Auth middleware
│   ├── auth_test.go
│   ├── session.go               # Session middleware
│   ├── session_test.go
│   ├── rbac.go                  # RBAC middleware
│   └── rbac_test.go
```

---

## 3. Data Models

### 3.1 Auth Configuration

```go
package auth

type Config struct {
    Mode              AuthMode
    OIDCConfig        *OIDCConfig
    ForwardAuthConfig *ForwardAuthConfig
    SessionDuration   time.Duration
}

type AuthMode string

const (
    AuthModeOIDC        AuthMode = "oidc"
    AuthModeForwardAuth AuthMode = "forward_auth"
)

type OIDCConfig struct {
    IssuerURL    string
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}

type ForwardAuthConfig struct {
    UserHeader  string
    EmailHeader string
    TrustedIPs  []string
}
```

### 3.2 Session Cookie

```go
package auth

const (
    SessionCookieName = "tinyrsvp_session"
    SessionDuration   = 7 * 24 * time.Hour
)

type SessionCookie struct {
    Name     string
    Value    string
    Path     string
    Domain   string
    MaxAge   int
    Secure   bool
    HttpOnly bool
    SameSite http.SameSite
}

func NewSessionCookie(sessionID string, secure bool) *SessionCookie {
    return &SessionCookie{
        Name:     SessionCookieName,
        Value:    sessionID,
        Path:     "/",
        MaxAge:   int(SessionDuration.Seconds()),
        Secure:   secure,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
}
```

### 3.3 Auth Context

```go
package auth

import "context"

type contextKey string

const (
    userContextKey    contextKey = "user"
    sessionContextKey contextKey = "session"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
    return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*models.User, bool) {
    user, ok := ctx.Value(userContextKey).(*models.User)
    return user, ok
}

func WithSession(ctx context.Context, session *models.Session) context.Context {
    return context.WithValue(ctx, sessionContextKey, session)
}

func SessionFromContext(ctx context.Context) (*models.Session, bool) {
    session, ok := ctx.Value(sessionContextKey).(*models.Session)
    return session, ok
}
```

---

## 4. Interfaces

### 4.1 Authenticator Interface

```go
package auth

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
    "net/http"
)

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
```

### 4.2 Session Manager Interface

```go
package auth

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
    "net/http"
)

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
```

### 4.3 Authorization Checker Interface

```go
package auth

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type AuthorizationChecker interface {
    CanCreateEvent(ctx context.Context, user *models.User) bool
    CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool
    CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool
    CanManageUsers(ctx context.Context, user *models.User) bool
    CanConfigureSystem(ctx context.Context, user *models.User) bool
    IsAdmin(user *models.User) bool
    IsEventManager(user *models.User) bool
}
```

### 4.4 User Service Interface

```go
package auth

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

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
```

---

## 5. Implementation Details

### 5.1 OIDC Authentication

```go
package auth

import (
    "context"
    "fmt"
    "net/http"
    
    "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
)

type oidcAuthenticator struct {
    provider     *oidc.Provider
    oauth2Config oauth2.Config
    verifier     *oidc.IDTokenVerifier
    userService  UserService
    sessionMgr   SessionManager
}

func NewOIDCAuthenticator(cfg *OIDCConfig, userService UserService, sessionMgr SessionManager) (Authenticator, error) {
    ctx := context.Background()
    
    provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
    if err != nil {
        return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
    }
    
    oauth2Config := oauth2.Config{
        ClientID:     cfg.ClientID,
        ClientSecret: cfg.ClientSecret,
        RedirectURL:  cfg.RedirectURL,
        Endpoint:     provider.Endpoint(),
        Scopes:       append([]string{oidc.ScopeOpenID, "email", "profile"}, cfg.Scopes...),
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
    state := generateRandomState()
    
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

func (a *oidcAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
    ctx := r.Context()
    
    stateCookie, err := r.Cookie("oauth_state")
    if err != nil {
        return nil, fmt.Errorf("missing state cookie: %w", err)
    }
    
    if r.URL.Query().Get("state") != stateCookie.Value {
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
        return nil, fmt.Errorf("missing id_token")
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
```

### 5.2 Forward Auth Implementation

```go
package auth

import (
    "context"
    "fmt"
    "net/http"
    "strings"
)

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
    return a.validateAndCreateSession(w, r)
}

func (a *forwardAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
    user := r.Header.Get(a.config.UserHeader)
    email := r.Header.Get(a.config.EmailHeader)
    
    if user == "" || email == "" {
        return nil, fmt.Errorf("missing required headers")
    }
    
    if !strings.Contains(email, "@") {
        return nil, fmt.Errorf("invalid email format")
    }
    
    return &AuthResult{
        Email:       email,
        Name:        user,
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

func (a *forwardAuthenticator) validateAndCreateSession(w http.ResponseWriter, r *http.Request) error {
    result, err := a.HandleCallback(w, r)
    if err != nil {
        return err
    }
    
    user, err := a.userService.GetOrCreateUser(r.Context(), result.Email, result.Name, result.OIDCSubject)
    if err != nil {
        return fmt.Errorf("failed to get or create user: %w", err)
    }
    
    session, err := a.sessionMgr.CreateSession(r.Context(), user.ID, r)
    if err != nil {
        return fmt.Errorf("failed to create session: %w", err)
    }
    
    return a.sessionMgr.SetSessionCookie(w, session.ID)
}
```

### 5.3 Session Manager Implementation

```go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "net/http"
    "time"
    
    "github.com/yourusername/tinyrsvp/internal/db/repositories"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type sessionManager struct {
    repo   repositories.SessionRepository
    secure bool
}

func NewSessionManager(repo repositories.SessionRepository, secure bool) SessionManager {
    return &sessionManager{
        repo:   repo,
        secure: secure,
    }
}

func (m *sessionManager) CreateSession(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
    sessionID, err := generateSessionID()
    if err != nil {
        return nil, fmt.Errorf("failed to generate session ID: %w", err)
    }
    
    ipAddress := getClientIP(r)
    userAgent := r.UserAgent()
    
    session := &models.Session{
        ID:             sessionID,
        UserID:         userID,
        CreatedAt:      time.Now(),
        ExpiresAt:      time.Now().Add(SessionDuration),
        LastAccessedAt: time.Now(),
        IPAddress:      &ipAddress,
        UserAgent:      &userAgent,
    }
    
    if err := m.repo.Create(ctx, session); err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }
    
    return session, nil
}

func (m *sessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
    session, err := m.repo.GetByID(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    if session.IsExpired() {
        m.repo.Delete(ctx, sessionID)
        return nil, fmt.Errorf("session expired")
    }
    
    return session, nil
}

func (m *sessionManager) RefreshSession(ctx context.Context, sessionID string) error {
    return m.repo.UpdateLastAccessed(ctx, sessionID)
}

func (m *sessionManager) DeleteSession(ctx context.Context, sessionID string) error {
    return m.repo.Delete(ctx, sessionID)
}

func (m *sessionManager) DeleteUserSessions(ctx context.Context, userID int64) error {
    return m.repo.DeleteByUserID(ctx, userID)
}

func (m *sessionManager) CleanupExpired(ctx context.Context) (int64, error) {
    return m.repo.DeleteExpired(ctx)
}

func (m *sessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
    cookie := NewSessionCookie(sessionID, m.secure)
    http.SetCookie(w, &http.Cookie{
        Name:     cookie.Name,
        Value:    cookie.Value,
        Path:     cookie.Path,
        MaxAge:   cookie.MaxAge,
        Secure:   cookie.Secure,
        HttpOnly: cookie.HttpOnly,
        SameSite: cookie.SameSite,
    })
    return nil
}

func (m *sessionManager) ClearSessionCookie(w http.ResponseWriter) error {
    http.SetCookie(w, &http.Cookie{
        Name:     SessionCookieName,
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        Secure:   m.secure,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    })
    return nil
}

func (m *sessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
    cookie, err := r.Cookie(SessionCookieName)
    if err != nil {
        return "", fmt.Errorf("session cookie not found: %w", err)
    }
    return cookie.Value, nil
}

func generateSessionID() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

func getClientIP(r *http.Request) string {
    if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
        return strings.Split(ip, ",")[0]
    }
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }
    return r.RemoteAddr
}
```

### 5.4 User Service Implementation

```go
package auth

import (
    "context"
    "fmt"
    
    "github.com/yourusername/tinyrsvp/internal/db/repositories"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type userService struct {
    repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
    return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
    isFirst, err := s.repo.IsFirstUser(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to check if first user: %w", err)
    }
    
    role := models.RoleEventManager
    if isFirst {
        role = models.RoleAdmin
    }
    
    user := &models.User{
        Email:       email,
        Name:        name,
        Role:        role,
        OIDCSubject: oidcSubject,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}

func (s *userService) GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
    var user *models.User
    var err error
    
    if oidcSubject != nil {
        user, err = s.repo.GetByOIDCSubject(ctx, *oidcSubject)
        if err == nil {
            if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
                return nil, fmt.Errorf("failed to update last login: %w", err)
            }
            return user, nil
        }
    }
    
    user, err = s.repo.GetByEmail(ctx, email)
    if err == nil {
        if oidcSubject != nil && user.OIDCSubject == nil {
            user.OIDCSubject = oidcSubject
            if err := s.repo.Update(ctx, user); err != nil {
                return nil, fmt.Errorf("failed to update user: %w", err)
            }
        }
        if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
            return nil, fmt.Errorf("failed to update last login: %w", err)
        }
        return user, nil
    }
    
    return s.CreateUser(ctx, email, name, oidcSubject)
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    return s.repo.GetByEmail(ctx, email)
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
    return s.repo.Update(ctx, user)
}

func (s *userService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
    user, err := s.repo.GetByID(ctx, userID)
    if err != nil {
        return err
    }
    
    user.Role = role
    return s.repo.Update(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
    return s.repo.Delete(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
    return s.repo.List(ctx, limit, offset)
}
```

### 5.5 Authorization Checker Implementation

```go
package auth

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type authorizationChecker struct{}

func NewAuthorizationChecker() AuthorizationChecker {
    return &authorizationChecker{}
}

func (c *authorizationChecker) IsAdmin(user *models.User) bool {
    return user.Role == models.RoleAdmin
}

func (c *authorizationChecker) IsEventManager(user *models.User) bool {
    return user.Role == models.RoleEventManager || user.Role == models.RoleAdmin
}

func (c *authorizationChecker) CanCreateEvent(ctx context.Context, user *models.User) bool {
    return c.IsEventManager(user)
}

func (c *authorizationChecker) CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool {
    if c.IsAdmin(user) {
        return true
    }
    return user.ID == event.CreatedBy
}

func (c *authorizationChecker) CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool {
    if c.IsAdmin(user) {
        return true
    }
    if user.ID == event.CreatedBy {
        return event.Status == models.EventStatusDraft || event.Status == models.EventStatusPublished
    }
    return false
}

func (c *authorizationChecker) CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool {
    return c.IsEventManager(user)
}

func (c *authorizationChecker) CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool {
    return c.CanEditEvent(ctx, user, event)
}

func (c *authorizationChecker) CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool {
    return c.CanEditEvent(ctx, user, event)
}

func (c *authorizationChecker) CanManageUsers(ctx context.Context, user *models.User) bool {
    return c.IsAdmin(user)
}

func (c *authorizationChecker) CanConfigureSystem(ctx context.Context, user *models.User) bool {
    return c.IsAdmin(user)
}
```

---

## 6. Dependencies

### 6.1 External Libraries

```go
import (
    "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
)
```

### 6.2 Internal Dependencies

- **Domain 7 (Database)** - User and session repositories
  - `repositories.UserRepository`
  - `repositories.SessionRepository`

### 6.3 Dependents

- **Domain 2 (Event)** - Permission checks for event operations
- **Domain 3 (Invite)** - Permission checks for invite operations
- **Domain 8 (API)** - Auth middleware for all protected endpoints

**See:** [LLD Index - Dependency Graph](../04_LLD_INDEX.md#dependency-graph)

---

## 7. Testing Strategy

### 7.1 Unit Tests

```go
func TestOIDCAuthenticator_HandleCallback(t *testing.T) {
    tests := []struct {
        name        string
        code        string
        state       string
        cookieState string
        wantErr     bool
    }{
        {
            name:        "valid callback",
            code:        "valid_code",
            state:       "state123",
            cookieState: "state123",
            wantErr:     false,
        },
        {
            name:        "state mismatch",
            code:        "valid_code",
            state:       "state123",
            cookieState: "different",
            wantErr:     true,
        },
        {
            name:        "missing code",
            code:        "",
            state:       "state123",
            cookieState: "state123",
            wantErr:     true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            
        })
    }
}

func TestUserService_GetOrCreateUser(t *testing.T) {
    tests := []struct {
        name        string
        email       string
        userName    string
        oidcSubject *string
        setupDB     func(*testing.T, repositories.UserRepository)
        wantRole    models.UserRole
        wantErr     bool
    }{
        {
            name:     "first user becomes admin",
            email:    "first@example.com",
            userName: "First User",
            setupDB:  func(t *testing.T, repo repositories.UserRepository) {},
            wantRole: models.RoleAdmin,
            wantErr:  false,
        },
        {
            name:     "second user becomes event manager",
            email:    "second@example.com",
            userName: "Second User",
            setupDB: func(t *testing.T, repo repositories.UserRepository) {
                repo.Create(context.Background(), &models.User{
                    Email: "first@example.com",
                    Role:  models.RoleAdmin,
                })
            },
            wantRole: models.RoleEventManager,
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDB(t)
            defer db.Close()
            
            repo := repositories.NewUserRepository(db)
            tt.setupDB(t, repo)
            
            service := NewUserService(repo)
            user, err := service.GetOrCreateUser(context.Background(), tt.email, tt.userName, tt.oidcSubject)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("GetOrCreateUser() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !tt.wantErr && user.Role != tt.wantRole {
                t.Errorf("Expected role %v, got %v", tt.wantRole, user.Role)
            }
        })
    }
}
```

### 7.2 Mock Implementations

```go
package auth

type MockAuthenticator struct {
    HandleLoginFunc    func(w http.ResponseWriter, r *http.Request) error
    HandleCallbackFunc func(w http.ResponseWriter, r *http.Request) (*AuthResult, error)
    HandleLogoutFunc   func(w http.ResponseWriter, r *http.Request) error
}

func (m *MockAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) error {
    if m.HandleLoginFunc != nil {
        return m.HandleLoginFunc(w, r)
    }
    return nil
}

func (m *MockAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
    if m.HandleCallbackFunc != nil {
        return m.HandleCallbackFunc(w, r)
    }
    return &AuthResult{Email: "test@example.com", Name: "Test User"}, nil
}

func (m *MockAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) error {
    if m.HandleLogoutFunc != nil {
        return m.HandleLogoutFunc(w, r)
    }
    return nil
}
```

---

## 8. Security Considerations

### 8.1 Session Security

**Secure Cookie Attributes:**
- `HttpOnly: true` - Prevents JavaScript access
- `Secure: true` - HTTPS only
- `SameSite: Lax` - CSRF protection
- 7-day expiration

**Session ID Generation:**
- 32 bytes cryptographically secure random
- Base64-URL encoded
- Collision probability: negligible

### 8.2 OIDC Security

**ID Token Validation:**
- Signature verification
- Issuer/audience validation
- Expiration check
- State parameter CSRF protection

### 8.3 Forward Auth Security

**Requirements:**
- App behind trusted proxy only
- Proxy strips client headers
- Validate header format

---

## 9. Performance Considerations

**Session Lookup:** Indexed by session ID
**User Lookup:** Indexed by email and OIDC subject
**Connection Pool:** 25 max connections

---

## 10. Error Scenarios

**OIDC Provider Down:** Show error with retry
**Invalid Token:** Redirect to login
**Missing Email:** Show error, contact admin
**Session Expired:** Redirect to login
**Missing Headers:** HTTP 401

---

## 11. Examples

See implementation sections above for complete examples.

---

## 12. Open Questions

**None** - All design decisions finalized

---

**Document Status:** ✅ Complete

**Next Domain:** [Domain 2: Event Management](02_EVENT_LLD.md)