# User Story: Authentication Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1 day

---

## User Story

As an **event manager**, I want **to log in and out of the application** so that **I can securely access my events and manage invites**.

---

## Acceptance Criteria

- [ ] GET /login - Display login page
- [ ] GET /auth/oidc/login - Redirect to OIDC provider
- [ ] GET /auth/oidc/callback - Handle OIDC callback
- [ ] POST /logout - Logout and clear session
- [ ] Session cookie set on login
- [ ] Session cleared on logout
- [ ] Redirect to original URL after login
- [ ] Error handling for failed login
- [ ] CSRF protection on logout
- [ ] Rate limiting on login attempts

---

## Technical Details

### Package Location
- `internal/handlers/auth.go` - Auth handlers
- `internal/handlers/auth_test.go` - Tests
- `templates/web/login.html` - Login page

### Routes

```go
r.Route("/auth", func(r chi.Router) {
    r.Get("/login", handlers.ShowLogin)
    r.Get("/oidc/login", handlers.OIDCLogin)
    r.Get("/oidc/callback", handlers.OIDCCallback)
})

r.Post("/logout", handlers.Logout)
```

---

## Tasks

- [ ] Implement login page handler
- [ ] Implement OIDC login redirect
- [ ] Implement OIDC callback handler
- [ ] Implement logout handler
- [ ] Add session management
- [ ] Add redirect handling
- [ ] Test authentication flow
- [ ] Test error cases

---

## Dependencies

**Depends on:** 
- 08_STORY_00_router_setup.md
- 08_STORY_01_middleware_chain.md
- Epic 01 (Auth implementation)

**Blocks:** 08_STORY_07_dashboard_route.md

---

## Testing Strategy

```go
func TestShowLogin(t *testing.T)
func TestOIDCLogin(t *testing.T)
func TestOIDCCallback_Success(t *testing.T)
func TestOIDCCallback_Error(t *testing.T)
func TestLogout(t *testing.T)
```

---

## Handler Implementations

### Login Page

```go
func (h *Handlers) ShowLogin(w http.ResponseWriter, r *http.Request) {
    returnURL := r.URL.Query().Get("return")
    if returnURL == "" {
        returnURL = "/"
    }
    
    data := struct {
        ReturnURL string
    }{
        ReturnURL: returnURL,
    }
    
    h.templates.Render(w, "login.html", data)
}
```

### OIDC Login

```go
func (h *Handlers) OIDCLogin(w http.ResponseWriter, r *http.Request) {
    returnURL := r.URL.Query().Get("return")
    
    state := generateState()
    h.sessions.Set(r.Context(), "oauth_state", state)
    h.sessions.Set(r.Context(), "return_url", returnURL)
    
    authURL := h.oidc.AuthCodeURL(state)
    http.Redirect(w, r, authURL, http.StatusFound)
}
```

### OIDC Callback

```go
func (h *Handlers) OIDCCallback(w http.ResponseWriter, r *http.Request) {
    state := r.URL.Query().Get("state")
    savedState := h.sessions.Get(r.Context(), "oauth_state")
    
    if state != savedState {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    code := r.URL.Query().Get("code")
    token, err := h.oidc.Exchange(r.Context(), code)
    if err != nil {
        http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
        return
    }
    
    user, err := h.auth.GetOrCreateUser(r.Context(), token)
    if err != nil {
        http.Error(w, "Failed to get user", http.StatusInternalServerError)
        return
    }
    
    h.sessions.Set(r.Context(), "user_id", user.ID)
    
    returnURL := h.sessions.Get(r.Context(), "return_url")
    if returnURL == "" {
        returnURL = "/"
    }
    
    http.Redirect(w, r, returnURL, http.StatusFound)
}
```

### Logout

```go
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
    h.sessions.Destroy(r.Context())
    http.Redirect(w, r, "/login", http.StatusFound)
}
```

---

## Security Considerations

- State parameter prevents CSRF
- Session cookies use HttpOnly and Secure flags
- Rate limit login attempts
- Log failed login attempts
- Validate return URL to prevent open redirect

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Auth Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All routes implemented
- [ ] Authentication flow working
- [ ] Session management functional
- [ ] Tests passing
- [ ] Documentation complete
