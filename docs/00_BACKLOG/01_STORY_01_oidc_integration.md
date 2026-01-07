# User Story: OIDC Authentication Integration

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 8 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As an **admin or event manager**, I want **to authenticate using OpenID Connect (OIDC)** so that **I can securely log in using my existing identity provider**.

---

## Acceptance Criteria

- [ ] OIDC provider discovery working
- [ ] Authorization code flow implemented
- [ ] State parameter CSRF protection functional
- [ ] ID token signature verification working
- [ ] User claims extracted correctly
- [ ] Session created on successful authentication
- [ ] Redirect to original URL after login
- [ ] Error handling for OIDC failures
- [ ] All tests pass with timeout

---

## Technical Details

### OIDC Configuration

```go
type OIDCConfig struct {
    IssuerURL    string
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}
```

### OIDC Authenticator Interface

```go
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

### Authentication Flow

```
1. User visits /login
2. Redirect to OIDC provider with state parameter
3. User authenticates with provider
4. Provider redirects to /auth/callback with code
5. Exchange code for tokens
6. Verify ID token signature
7. Extract user claims (email, name, subject)
8. Create/update user in database
9. Create session
10. Set session cookie
11. Redirect to dashboard
```

### Required Scopes

- `openid` - Required for OIDC
- `email` - User email address
- `profile` - User name and basic profile

### Security Measures

- State parameter prevents CSRF attacks
- ID token signature verification prevents tampering
- Token expiration checking
- Secure cookie for session (HttpOnly, Secure, SameSite)

---

## Tasks

### Phase 1: OIDC Setup (TDD)
- [ ] Write test for OIDC provider discovery
- [ ] Write test for OAuth2 config creation
- [ ] Write test for verifier initialization
- [ ] Implement `NewOIDCAuthenticator()` constructor
- [ ] Run tests (should pass)

### Phase 2: Login Handler (TDD)
- [ ] Write test for state generation
- [ ] Write test for state cookie creation
- [ ] Write test for authorization URL generation
- [ ] Write test for redirect to provider
- [ ] Implement `HandleLogin()` method
- [ ] Run tests (should pass)

### Phase 3: Callback Handler (TDD)
- [ ] Write test for state validation
- [ ] Write test for missing state cookie
- [ ] Write test for state mismatch
- [ ] Write test for missing authorization code
- [ ] Write test for code exchange
- [ ] Write test for ID token extraction
- [ ] Write test for ID token verification
- [ ] Write test for claims parsing
- [ ] Write test for missing email claim
- [ ] Implement `HandleCallback()` method
- [ ] Run tests (should pass)

### Phase 4: Logout Handler (TDD)
- [ ] Write test for session deletion
- [ ] Write test for cookie clearing
- [ ] Write test for missing session
- [ ] Implement `HandleLogout()` method
- [ ] Run tests (should pass)

### Phase 5: HTTP Handlers (TDD)
- [ ] Write test for `/login` endpoint
- [ ] Write test for `/auth/callback` endpoint
- [ ] Write test for `/logout` endpoint
- [ ] Implement HTTP handlers
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Add OIDC config to application config
- [ ] Wire authenticator into HTTP router
- [ ] Test with real OIDC provider (Keycloak/Authentik)
- [ ] Document OIDC setup instructions
- [ ] Update README with OIDC configuration

---

## Testing Requirements

### Unit Tests

```go
func TestOIDCAuthenticator_HandleLogin(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*httptest.ResponseRecorder, *http.Request)
        wantErr bool
    }{
        {
            name:    "successful redirect",
            setup:   func(w *httptest.ResponseRecorder, r *http.Request) {},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockProvider := setupMockOIDCProvider(t)
            auth, err := NewOIDCAuthenticator(testOIDCConfig(), mockUserService, mockSessionMgr)
            if err != nil {
                t.Fatalf("Failed to create authenticator: %v", err)
            }
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/login", nil)
            
            if tt.setup != nil {
                tt.setup(w, r)
            }
            
            err = auth.HandleLogin(w, r)
            if (err != nil) != tt.wantErr {
                t.Errorf("HandleLogin() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !tt.wantErr {
                if w.Code != http.StatusFound {
                    t.Errorf("Expected redirect status, got %d", w.Code)
                }
                
                location := w.Header().Get("Location")
                if location == "" {
                    t.Error("Expected Location header")
                }
                
                cookies := w.Result().Cookies()
                var stateCookie *http.Cookie
                for _, c := range cookies {
                    if c.Name == "oauth_state" {
                        stateCookie = c
                        break
                    }
                }
                
                if stateCookie == nil {
                    t.Error("Expected oauth_state cookie")
                }
            }
        })
    }
}

func TestOIDCAuthenticator_HandleCallback(t *testing.T) {
    tests := []struct {
        name         string
        code         string
        state        string
        cookieState  string
        setupMock    func(*MockOIDCProvider)
        wantErr      bool
        wantEmail    string
    }{
        {
            name:        "valid callback",
            code:        "valid_code",
            state:       "state123",
            cookieState: "state123",
            setupMock: func(m *MockOIDCProvider) {
                m.ExchangeFunc = func(ctx context.Context, code string) (*oauth2.Token, error) {
                    token := &oauth2.Token{}
                    token = token.WithExtra(map[string]interface{}{
                        "id_token": "valid_id_token",
                    })
                    return token, nil
                }
                m.VerifyFunc = func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
                    return &oidc.IDToken{Subject: "user123"}, nil
                }
                m.ClaimsFunc = func(v interface{}) error {
                    claims := v.(*struct {
                        Email string `json:"email"`
                        Name  string `json:"name"`
                    })
                    claims.Email = "user@example.com"
                    claims.Name = "Test User"
                    return nil
                }
            },
            wantErr:   false,
            wantEmail: "user@example.com",
        },
        {
            name:        "state mismatch",
            code:        "valid_code",
            state:       "state123",
            cookieState: "different_state",
            wantErr:     true,
        },
        {
            name:        "missing code",
            code:        "",
            state:       "state123",
            cookieState: "state123",
            wantErr:     true,
        },
        {
            name:        "missing state cookie",
            code:        "valid_code",
            state:       "state123",
            cookieState: "",
            wantErr:     true,
        },
        {
            name:        "token exchange failure",
            code:        "invalid_code",
            state:       "state123",
            cookieState: "state123",
            setupMock: func(m *MockOIDCProvider) {
                m.ExchangeFunc = func(ctx context.Context, code string) (*oauth2.Token, error) {
                    return nil, fmt.Errorf("exchange failed")
                }
            },
            wantErr: true,
        },
        {
            name:        "missing email claim",
            code:        "valid_code",
            state:       "state123",
            cookieState: "state123",
            setupMock: func(m *MockOIDCProvider) {
                m.ExchangeFunc = func(ctx context.Context, code string) (*oauth2.Token, error) {
                    token := &oauth2.Token{}
                    token = token.WithExtra(map[string]interface{}{
                        "id_token": "valid_id_token",
                    })
                    return token, nil
                }
                m.VerifyFunc = func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
                    return &oidc.IDToken{Subject: "user123"}, nil
                }
                m.ClaimsFunc = func(v interface{}) error {
                    return nil
                }
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockProvider := &MockOIDCProvider{}
            if tt.setupMock != nil {
                tt.setupMock(mockProvider)
            }
            
            auth := setupAuthenticatorWithMock(t, mockProvider)
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/auth/callback?code="+tt.code+"&state="+tt.state, nil)
            
            if tt.cookieState != "" {
                r.AddCookie(&http.Cookie{
                    Name:  "oauth_state",
                    Value: tt.cookieState,
                })
            }
            
            result, err := auth.HandleCallback(w, r)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("HandleCallback() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                if result == nil {
                    t.Error("Expected AuthResult, got nil")
                    return
                }
                
                if result.Email != tt.wantEmail {
                    t.Errorf("Expected email %q, got %q", tt.wantEmail, result.Email)
                }
                
                if result.OIDCSubject == nil {
                    t.Error("Expected OIDC subject, got nil")
                }
            }
        })
    }
}

func TestOIDCAuthenticator_HandleLogout(t *testing.T) {
    tests := []struct {
        name          string
        sessionID     string
        setupMock     func(*MockSessionManager)
        wantErr       bool
    }{
        {
            name:      "successful logout",
            sessionID: "session123",
            setupMock: func(m *MockSessionManager) {
                m.DeleteSessionFunc = func(ctx context.Context, sessionID string) error {
                    return nil
                }
            },
            wantErr: false,
        },
        {
            name:      "no session cookie",
            sessionID: "",
            wantErr:   false,
        },
        {
            name:      "delete session error",
            sessionID: "session123",
            setupMock: func(m *MockSessionManager) {
                m.DeleteSessionFunc = func(ctx context.Context, sessionID string) error {
                    return fmt.Errorf("delete failed")
                }
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSessionMgr := &MockSessionManager{}
            if tt.setupMock != nil {
                tt.setupMock(mockSessionMgr)
            }
            
            auth := setupAuthenticatorWithMocks(t, mockSessionMgr)
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("POST", "/logout", nil)
            
            if tt.sessionID != "" {
                r.AddCookie(&http.Cookie{
                    Name:  "tinyrsvp_session",
                    Value: tt.sessionID,
                })
            }
            
            err := auth.HandleLogout(w, r)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("HandleLogout() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

```go
func TestOIDC_FullFlow(t *testing.T) {
    testOIDCProvider := setupTestOIDCProvider(t)
    defer testOIDCProvider.Close()
    
    db := setupTestDB(t)
    defer db.Close()
    
    cfg := &OIDCConfig{
        IssuerURL:    testOIDCProvider.URL,
        ClientID:     "test-client",
        ClientSecret: "test-secret",
        RedirectURL:  "http://localhost:8080/auth/callback",
        Scopes:       []string{"openid", "email", "profile"},
    }
    
    userRepo := repositories.NewUserRepository(db)
    sessionRepo := repositories.NewSessionRepository(db)
    
    userService := auth.NewUserService(userRepo)
    sessionMgr := auth.NewSessionManager(sessionRepo, false)
    
    authenticator, err := auth.NewOIDCAuthenticator(cfg, userService, sessionMgr)
    if err != nil {
        t.Fatalf("Failed to create authenticator: %v", err)
    }
    
    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/login", nil)
    
    if err := authenticator.HandleLogin(w, r); err != nil {
        t.Fatalf("HandleLogin failed: %v", err)
    }
    
    if w.Code != http.StatusFound {
        t.Fatalf("Expected redirect, got %d", w.Code)
    }
}
```

---

## Dependencies

**Depends on:** 
- User model and repository (01_STORY_user_model.md)
- Session management (01_STORY_session_management.md)

**Blocks:** 
- User authentication flow
- Admin dashboard access

**External Dependencies:**
- `github.com/coreos/go-oidc/v3/oidc` - OIDC library
- `golang.org/x/oauth2` - OAuth2 library

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/auth/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] OIDC flow tested with real provider
- [ ] State CSRF protection verified
- [ ] ID token verification tested
- [ ] Error handling comprehensive
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### OIDC Provider Discovery

The authenticator uses OIDC Discovery to automatically retrieve provider configuration:
- Authorization endpoint
- Token endpoint
- JWKS (JSON Web Key Set) endpoint
- Supported scopes and claims

### State Parameter

The state parameter is a cryptographically random string that:
- Prevents CSRF attacks
- Stored in secure cookie with 5-minute expiration
- Verified on callback

### ID Token Verification

The ID token is verified for:
- Signature validity using provider's public keys
- Issuer matches configured issuer
- Audience matches client ID
- Token not expired

### Claims Extraction

Required claims:
- `sub` (subject) - Unique user identifier from provider
- `email` - User's email address
- `name` (optional) - User's display name

---

## Configuration Example

```yaml
auth:
  mode: oidc
  oidc:
    issuer_url: https://keycloak.example.com/realms/tinyrsvp
    client_id: tinyrsvp-app
    client_secret: your-client-secret
    redirect_url: https://rsvp.example.com/auth/callback
    scopes:
      - openid
      - email
      - profile
```

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md) - Section 5.1
- **OIDC Specification:** https://openid.net/specs/openid-connect-core-1_0.html
- **go-oidc Documentation:** https://github.com/coreos/go-oidc
