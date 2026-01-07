# User Story: Forward Auth Integration

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 4 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As an **admin or event manager**, I want **to authenticate using forward auth headers** so that **I can use my existing reverse proxy authentication (Authelia/Authentik)**.

---

## Acceptance Criteria

- [ ] Forward auth headers validated correctly
- [ ] Email and username extracted from headers
- [ ] Header format validation implemented
- [ ] User created/updated on authentication
- [ ] Session created on successful validation
- [ ] IP validation for trusted proxies working
- [ ] Clear error messages for missing/invalid headers
- [ ] All tests pass with timeout

---

## Technical Details

### Forward Auth Configuration

```go
type ForwardAuthConfig struct {
    UserHeader  string
    EmailHeader string
    TrustedIPs  []string
}
```

### Common Header Names

**Authelia:**
- `Remote-User` - Username
- `Remote-Email` - Email address
- `Remote-Name` - Full name
- `Remote-Groups` - Groups (comma-separated)

**Authentik:**
- `X-authentik-username` - Username
- `X-authentik-email` - Email address
- `X-authentik-name` - Full name
- `X-authentik-groups` - Groups

**Generic:**
- `X-Forwarded-User` - Username
- `X-Forwarded-Email` - Email address

### Authentication Flow

```
1. Request arrives at reverse proxy
2. Proxy authenticates user
3. Proxy forwards request with auth headers
4. App validates headers present
5. App validates request from trusted proxy
6. App extracts user information
7. App creates/updates user in database
8. App creates session
9. App sets session cookie
10. Request continues to handler
```

### Security Considerations

- **Header Stripping:** Proxy MUST strip client-supplied headers
- **IP Validation:** Only accept headers from trusted proxy IPs
- **Email Validation:** Validate email format
- **No OIDC Subject:** Forward auth doesn't provide OIDC subject

---

## Tasks

### Phase 1: Configuration (TDD)
- [ ] Write test for forward auth config parsing
- [ ] Write test for missing required config
- [ ] Write test for invalid IP format
- [ ] Implement config validation
- [ ] Run tests (should pass)

### Phase 2: Header Validation (TDD)
- [ ] Write test for valid headers
- [ ] Write test for missing user header
- [ ] Write test for missing email header
- [ ] Write test for invalid email format
- [ ] Write test for empty header values
- [ ] Implement header validation
- [ ] Run tests (should pass)

### Phase 3: IP Validation (TDD)
- [ ] Write test for trusted IP validation
- [ ] Write test for untrusted IP rejection
- [ ] Write test for X-Forwarded-For parsing
- [ ] Write test for X-Real-IP parsing
- [ ] Write test for direct connection IP
- [ ] Implement IP validation
- [ ] Run tests (should pass)

### Phase 4: Authenticator Implementation (TDD)
- [ ] Write test for HandleLogin (auto-validates)
- [ ] Write test for HandleCallback (extracts headers)
- [ ] Write test for HandleLogout
- [ ] Write test for user creation on first auth
- [ ] Write test for user update on subsequent auth
- [ ] Implement forwardAuthenticator
- [ ] Run tests (should pass)

### Phase 5: HTTP Handlers (TDD)
- [ ] Write test for middleware integration
- [ ] Write test for header extraction in middleware
- [ ] Write test for error responses
- [ ] Implement HTTP handlers
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Add forward auth config to application config
- [ ] Wire authenticator into HTTP router
- [ ] Test with Authelia/Authentik
- [ ] Document forward auth setup
- [ ] Update README with proxy configuration examples

---

## Testing Requirements

### Unit Tests

```go
func TestForwardAuthenticator_HandleCallback(t *testing.T) {
    tests := []struct {
        name      string
        headers   map[string]string
        wantErr   bool
        wantEmail string
        wantName  string
    }{
        {
            name: "valid authelia headers",
            headers: map[string]string{
                "Remote-User":  "testuser",
                "Remote-Email": "test@example.com",
                "Remote-Name":  "Test User",
            },
            wantErr:   false,
            wantEmail: "test@example.com",
            wantName:  "Test User",
        },
        {
            name: "valid authentik headers",
            headers: map[string]string{
                "X-authentik-username": "testuser",
                "X-authentik-email":    "test@example.com",
                "X-authentik-name":     "Test User",
            },
            wantErr:   false,
            wantEmail: "test@example.com",
            wantName:  "Test User",
        },
        {
            name: "missing email header",
            headers: map[string]string{
                "Remote-User": "testuser",
            },
            wantErr: true,
        },
        {
            name: "missing user header",
            headers: map[string]string{
                "Remote-Email": "test@example.com",
            },
            wantErr: true,
        },
        {
            name: "invalid email format",
            headers: map[string]string{
                "Remote-User":  "testuser",
                "Remote-Email": "not-an-email",
            },
            wantErr: true,
        },
        {
            name: "empty email",
            headers: map[string]string{
                "Remote-User":  "testuser",
                "Remote-Email": "",
            },
            wantErr: true,
        },
        {
            name: "empty user",
            headers: map[string]string{
                "Remote-User":  "",
                "Remote-Email": "test@example.com",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := &ForwardAuthConfig{
                UserHeader:  "Remote-User",
                EmailHeader: "Remote-Email",
                TrustedIPs:  []string{"127.0.0.1"},
            }
            
            auth := NewForwardAuthenticator(cfg, mockUserService, mockSessionMgr)
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/", nil)
            
            for k, v := range tt.headers {
                r.Header.Set(k, v)
            }
            
            result, err := auth.HandleCallback(w, r)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("HandleCallback() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                if result.Email != tt.wantEmail {
                    t.Errorf("Expected email %q, got %q", tt.wantEmail, result.Email)
                }
                
                if result.Name != tt.wantName {
                    t.Errorf("Expected name %q, got %q", tt.wantName, result.Name)
                }
                
                if result.OIDCSubject != nil {
                    t.Error("Expected nil OIDC subject for forward auth")
                }
            }
        })
    }
}

func TestForwardAuthenticator_IPValidation(t *testing.T) {
    tests := []struct {
        name       string
        trustedIPs []string
        remoteAddr string
        headers    map[string]string
        wantErr    bool
    }{
        {
            name:       "trusted direct connection",
            trustedIPs: []string{"127.0.0.1"},
            remoteAddr: "127.0.0.1:12345",
            wantErr:    false,
        },
        {
            name:       "untrusted direct connection",
            trustedIPs: []string{"127.0.0.1"},
            remoteAddr: "192.168.1.100:12345",
            wantErr:    true,
        },
        {
            name:       "trusted via X-Forwarded-For",
            trustedIPs: []string{"10.0.0.1"},
            remoteAddr: "192.168.1.100:12345",
            headers: map[string]string{
                "X-Forwarded-For": "10.0.0.1, 192.168.1.100",
            },
            wantErr: false,
        },
        {
            name:       "trusted via X-Real-IP",
            trustedIPs: []string{"10.0.0.1"},
            remoteAddr: "192.168.1.100:12345",
            headers: map[string]string{
                "X-Real-IP": "10.0.0.1",
            },
            wantErr: false,
        },
        {
            name:       "untrusted via X-Forwarded-For",
            trustedIPs: []string{"10.0.0.1"},
            remoteAddr: "192.168.1.100:12345",
            headers: map[string]string{
                "X-Forwarded-For": "192.168.1.50, 192.168.1.100",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := &ForwardAuthConfig{
                UserHeader:  "Remote-User",
                EmailHeader: "Remote-Email",
                TrustedIPs:  tt.trustedIPs,
            }
            
            auth := NewForwardAuthenticator(cfg, mockUserService, mockSessionMgr)
            
            r := httptest.NewRequest("GET", "/", nil)
            r.RemoteAddr = tt.remoteAddr
            
            for k, v := range tt.headers {
                r.Header.Set(k, v)
            }
            
            err := validateTrustedProxy(auth, r)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("validateTrustedProxy() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestForwardAuthenticator_HandleLogin(t *testing.T) {
    tests := []struct {
        name    string
        headers map[string]string
        wantErr bool
    }{
        {
            name: "valid headers creates session",
            headers: map[string]string{
                "Remote-User":  "testuser",
                "Remote-Email": "test@example.com",
            },
            wantErr: false,
        },
        {
            name: "invalid headers returns error",
            headers: map[string]string{
                "Remote-User": "testuser",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := &ForwardAuthConfig{
                UserHeader:  "Remote-User",
                EmailHeader: "Remote-Email",
                TrustedIPs:  []string{"127.0.0.1"},
            }
            
            mockUserService := &MockUserService{
                GetOrCreateUserFunc: func(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
                    return &models.User{ID: 1, Email: email, Name: name}, nil
                },
            }
            
            mockSessionMgr := &MockSessionManager{
                CreateSessionFunc: func(ctx context.Context, userID int64, r *http.Request) (*models.Session, error) {
                    return &models.Session{ID: "session123", UserID: userID}, nil
                },
                SetSessionCookieFunc: func(w http.ResponseWriter, sessionID string) error {
                    return nil
                },
            }
            
            auth := NewForwardAuthenticator(cfg, mockUserService, mockSessionMgr)
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/", nil)
            
            for k, v := range tt.headers {
                r.Header.Set(k, v)
            }
            
            err := auth.HandleLogin(w, r)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("HandleLogin() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

```go
func TestForwardAuth_WithAuthelia(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    cfg := &ForwardAuthConfig{
        UserHeader:  "Remote-User",
        EmailHeader: "Remote-Email",
        TrustedIPs:  []string{"127.0.0.1"},
    }
    
    userRepo := repositories.NewUserRepository(db)
    sessionRepo := repositories.NewSessionRepository(db)
    
    userService := auth.NewUserService(userRepo)
    sessionMgr := auth.NewSessionManager(sessionRepo, false)
    
    authenticator := auth.NewForwardAuthenticator(cfg, userService, sessionMgr)
    
    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/dashboard", nil)
    r.Header.Set("Remote-User", "testuser")
    r.Header.Set("Remote-Email", "test@example.com")
    r.Header.Set("Remote-Name", "Test User")
    
    if err := authenticator.HandleLogin(w, r); err != nil {
        t.Fatalf("HandleLogin failed: %v", err)
    }
    
    cookies := w.Result().Cookies()
    if len(cookies) == 0 {
        t.Fatal("Expected session cookie")
    }
}
```

---

## Dependencies

**Depends on:** 
- User model and repository (01_STORY_04_user_model.md)
- Session management (01_STORY_03_session_management.md)

**Blocks:** 
- User authentication flow
- Reverse proxy integration

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/auth/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Tested with Authelia headers
- [ ] Tested with Authentik headers
- [ ] IP validation working
- [ ] Header validation comprehensive
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### Header Security

Forward auth is only secure when:
1. Reverse proxy strips ALL auth headers from client requests
2. Only trusted proxy can set auth headers
3. Application validates request comes from trusted proxy
4. Headers cannot be spoofed by clients

### Email Validation

Email validation checks:
- Contains @ symbol
- Non-empty local part
- Non-empty domain part
- Valid character set

### Proxy Configuration

The application must be configured with the exact IP address(es) of the trusted reverse proxy. Requests from any other source will be rejected.

### No OIDC Subject

Unlike OIDC authentication, forward auth doesn't provide an OIDC subject. Users are identified solely by email address.

---

## Configuration Example

```yaml
auth:
  mode: forward_auth
  forward_auth:
    user_header: Remote-User
    email_header: Remote-Email
    trusted_ips:
      - 10.0.0.1
      - 172.16.0.1
```

### Authelia Configuration

```yaml
server:
  headers:
    authorization:
      Remote-User: "{{ .Subject }}"
      Remote-Email: "{{ .Email }}"
      Remote-Name: "{{ .DisplayName }}"
      Remote-Groups: "{{ join .Groups \",\" }}"
```

### Authentik Configuration

In Authentik proxy provider, configure headers:
- `X-authentik-username: {{ user.username }}`
- `X-authentik-email: {{ user.email }}`
- `X-authentik-name: {{ user.name }}`

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md) - Section 5.2
- **Authelia Documentation:** https://www.authelia.com/integration/proxies/introduction/
- **Authentik Documentation:** https://goauthentik.io/docs/providers/proxy/
