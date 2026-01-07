# User Story: RBAC Middleware

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 6 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As a **developer**, I want **role-based access control middleware** so that **only authorized users can access protected endpoints**.

---

## Acceptance Criteria

- [ ] RequireAuth middleware validates session
- [ ] RequireAdmin middleware restricts to admin users
- [ ] RequireEventManager middleware allows managers and admins
- [ ] Middleware chains correctly
- [ ] User and session injected into request context
- [ ] Clear error responses for unauthorized access
- [ ] Middleware skips public endpoints
- [ ] All tests pass with timeout

---

## Technical Details

### Middleware Functions

```go
package middleware

import (
    "context"
    "net/http"
    
    "github.com/yourusername/tinyrsvp/internal/auth"
    "github.com/yourusername/tinyrsvp/internal/models"
)

func RequireAuth(sessionMgr auth.SessionManager, userService auth.UserService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sessionID, err := sessionMgr.GetSessionFromRequest(r)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            session, err := sessionMgr.GetSession(r.Context(), sessionID)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            user, err := userService.GetUserByID(r.Context(), session.UserID)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            ctx := auth.WithUser(r.Context(), user)
            ctx = auth.WithSession(ctx, session)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, ok := auth.UserFromContext(r.Context())
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        if user.Role != models.RoleAdmin {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func RequireEventManager(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, ok := auth.UserFromContext(r.Context())
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        if user.Role != models.RoleAdmin && user.Role != models.RoleEventManager {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Middleware Chaining

```go
mux := http.NewServeMux()

authMiddleware := RequireAuth(sessionMgr, userService)
adminMiddleware := RequireAdmin
managerMiddleware := RequireEventManager

mux.Handle("/api/users", authMiddleware(adminMiddleware(http.HandlerFunc(listUsersHandler))))
mux.Handle("/api/events", authMiddleware(managerMiddleware(http.HandlerFunc(listEventsHandler))))
```

---

## Tasks

### Phase 1: RequireAuth Middleware (TDD)
- [ ] Write test for valid session
- [ ] Write test for missing session cookie
- [ ] Write test for invalid session ID
- [ ] Write test for expired session
- [ ] Write test for user not found
- [ ] Write test for context injection
- [ ] Implement RequireAuth middleware
- [ ] Run tests (should pass)

### Phase 2: RequireAdmin Middleware (TDD)
- [ ] Write test for admin user allowed
- [ ] Write test for event manager denied
- [ ] Write test for missing user in context
- [ ] Write test for middleware chaining
- [ ] Implement RequireAdmin middleware
- [ ] Run tests (should pass)

### Phase 3: RequireEventManager Middleware (TDD)
- [ ] Write test for admin allowed
- [ ] Write test for event manager allowed
- [ ] Write test for missing user denied
- [ ] Write test for middleware chaining
- [ ] Implement RequireEventManager middleware
- [ ] Run tests (should pass)

### Phase 4: Session Refresh Middleware (TDD)
- [ ] Write test for session last access update
- [ ] Write test for refresh on each request
- [ ] Implement session refresh in RequireAuth
- [ ] Run tests (should pass)

### Phase 5: Integration
- [ ] Wire middleware into HTTP router
- [ ] Test middleware chains
- [ ] Test endpoint protection
- [ ] Document middleware usage
- [ ] Create middleware examples

---

## Testing Requirements

### Unit Tests

```go
func TestRequireAuth(t *testing.T) {
    tests := []struct {
        name           string
        cookie         *http.Cookie
        mockSession    *models.Session
        mockUser       *models.User
        mockGetSession error
        mockGetUser    error
        wantStatus     int
        wantContext    bool
    }{
        {
            name: "valid session and user",
            cookie: &http.Cookie{
                Name:  "tinyrsvp_session",
                Value: "session123",
            },
            mockSession: &models.Session{
                ID:     "session123",
                UserID: 1,
            },
            mockUser: &models.User{
                ID:    1,
                Email: "user@example.com",
                Role:  models.RoleEventManager,
            },
            wantStatus:  http.StatusOK,
            wantContext: true,
        },
        {
            name:       "missing session cookie",
            cookie:     nil,
            wantStatus: http.StatusUnauthorized,
        },
        {
            name: "invalid session ID",
            cookie: &http.Cookie{
                Name:  "tinyrsvp_session",
                Value: "invalid",
            },
            mockGetSession: fmt.Errorf("session not found"),
            wantStatus:     http.StatusUnauthorized,
        },
        {
            name: "expired session",
            cookie: &http.Cookie{
                Name:  "tinyrsvp_session",
                Value: "expired",
            },
            mockSession: &models.Session{
                ID:        "expired",
                UserID:    1,
                ExpiresAt: time.Now().Add(-1 * time.Hour),
            },
            wantStatus: http.StatusUnauthorized,
        },
        {
            name: "user not found",
            cookie: &http.Cookie{
                Name:  "tinyrsvp_session",
                Value: "session123",
            },
            mockSession: &models.Session{
                ID:     "session123",
                UserID: 999,
            },
            mockGetUser: fmt.Errorf("user not found"),
            wantStatus:  http.StatusUnauthorized,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSessionMgr := &MockSessionManager{
                GetSessionFromRequestFunc: func(r *http.Request) (string, error) {
                    if tt.cookie == nil {
                        return "", http.ErrNoCookie
                    }
                    return tt.cookie.Value, nil
                },
                GetSessionFunc: func(ctx context.Context, sessionID string) (*models.Session, error) {
                    if tt.mockGetSession != nil {
                        return nil, tt.mockGetSession
                    }
                    return tt.mockSession, nil
                },
            }
            
            mockUserService := &MockUserService{
                GetUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
                    if tt.mockGetUser != nil {
                        return nil, tt.mockGetUser
                    }
                    return tt.mockUser, nil
                },
            }
            
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                if tt.wantContext {
                    user, ok := auth.UserFromContext(r.Context())
                    if !ok {
                        t.Error("Expected user in context")
                    }
                    if user.ID != tt.mockUser.ID {
                        t.Errorf("Expected user ID %d, got %d", tt.mockUser.ID, user.ID)
                    }
                    
                    session, ok := auth.SessionFromContext(r.Context())
                    if !ok {
                        t.Error("Expected session in context")
                    }
                    if session.ID != tt.mockSession.ID {
                        t.Errorf("Expected session ID %q, got %q", tt.mockSession.ID, session.ID)
                    }
                }
                w.WriteHeader(http.StatusOK)
            })
            
            middleware := RequireAuth(mockSessionMgr, mockUserService)
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/protected", nil)
            
            if tt.cookie != nil {
                r.AddCookie(tt.cookie)
            }
            
            middleware(handler).ServeHTTP(w, r)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
            }
        })
    }
}

func TestRequireAdmin(t *testing.T) {
    tests := []struct {
        name       string
        user       *models.User
        wantStatus int
    }{
        {
            name: "admin allowed",
            user: &models.User{
                ID:   1,
                Role: models.RoleAdmin,
            },
            wantStatus: http.StatusOK,
        },
        {
            name: "event manager denied",
            user: &models.User{
                ID:   2,
                Role: models.RoleEventManager,
            },
            wantStatus: http.StatusForbidden,
        },
        {
            name:       "no user in context",
            user:       nil,
            wantStatus: http.StatusUnauthorized,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            })
            
            middleware := RequireAdmin
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/admin", nil)
            
            if tt.user != nil {
                ctx := auth.WithUser(r.Context(), tt.user)
                r = r.WithContext(ctx)
            }
            
            middleware(handler).ServeHTTP(w, r)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
            }
        })
    }
}

func TestRequireEventManager(t *testing.T) {
    tests := []struct {
        name       string
        user       *models.User
        wantStatus int
    }{
        {
            name: "admin allowed",
            user: &models.User{
                ID:   1,
                Role: models.RoleAdmin,
            },
            wantStatus: http.StatusOK,
        },
        {
            name: "event manager allowed",
            user: &models.User{
                ID:   2,
                Role: models.RoleEventManager,
            },
            wantStatus: http.StatusOK,
        },
        {
            name:       "no user denied",
            user:       nil,
            wantStatus: http.StatusUnauthorized,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            })
            
            middleware := RequireEventManager
            
            w := httptest.NewRecorder()
            r := httptest.NewRequest("GET", "/events", nil)
            
            if tt.user != nil {
                ctx := auth.WithUser(r.Context(), tt.user)
                r = r.WithContext(ctx)
            }
            
            middleware(handler).ServeHTTP(w, r)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
            }
        })
    }
}

func TestMiddlewareChaining(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewUserRepository(db)
    sessionRepo := repositories.NewSessionRepository(db)
    
    user := &models.User{
        Email: "admin@example.com",
        Name:  "Admin",
        Role:  models.RoleAdmin,
    }
    repo.Create(context.Background(), user)
    
    sessionMgr := auth.NewSessionManager(sessionRepo, false)
    userService := auth.NewUserService(repo)
    
    r := httptest.NewRequest("GET", "/", nil)
    session, err := sessionMgr.CreateSession(context.Background(), user.ID, r)
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, ok := auth.UserFromContext(r.Context())
        if !ok {
            t.Error("Expected user in context")
        }
        if user.Role != models.RoleAdmin {
            t.Error("Expected admin user")
        }
        w.WriteHeader(http.StatusOK)
    })
    
    authMiddleware := RequireAuth(sessionMgr, userService)
    adminMiddleware := RequireAdmin
    
    chainedHandler := authMiddleware(adminMiddleware(handler))
    
    w := httptest.NewRecorder()
    r = httptest.NewRequest("GET", "/admin/endpoint", nil)
    r.AddCookie(&http.Cookie{
        Name:  "tinyrsvp_session",
        Value: session.ID,
    })
    
    chainedHandler.ServeHTTP(w, r)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

---

## Tasks

### Phase 1: RequireAuth Middleware (TDD)
- [ ] Write test for valid authentication
- [ ] Write test for missing session
- [ ] Write test for invalid session
- [ ] Write test for expired session
- [ ] Write test for context injection
- [ ] Implement RequireAuth
- [ ] Run tests (should pass)

### Phase 2: RequireAdmin Middleware (TDD)
- [ ] Write test for admin access granted
- [ ] Write test for non-admin denied
- [ ] Write test for missing user
- [ ] Implement RequireAdmin
- [ ] Run tests (should pass)

### Phase 3: RequireEventManager Middleware (TDD)
- [ ] Write test for admin access
- [ ] Write test for event manager access
- [ ] Write test for unauthorized access
- [ ] Implement RequireEventManager
- [ ] Run tests (should pass)

### Phase 4: Middleware Chaining (TDD)
- [ ] Write test for RequireAuth + RequireAdmin chain
- [ ] Write test for RequireAuth + RequireEventManager chain
- [ ] Write test for multiple middleware chain
- [ ] Test order of execution
- [ ] Run tests (should pass)

### Phase 5: Error Responses (TDD)
- [ ] Write test for JSON error responses
- [ ] Write test for HTML error pages
- [ ] Write test for proper HTTP status codes
- [ ] Implement error response logic
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Apply middleware to protected routes
- [ ] Test with real HTTP server
- [ ] Verify all endpoints properly protected
- [ ] Document middleware usage patterns

---

## Dependencies

**Depends on:** 
- Session management (01_STORY_03_session_management.md)
- User service (01_STORY_04_user_model.md)
- Auth context functions

**Blocks:** 
- All protected API endpoints
- User management CRUD (01_STORY_06_user_crud.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/middleware/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Middleware chains tested
- [ ] Context injection verified
- [ ] Error responses appropriate
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### Middleware Order

Middleware is applied in reverse order of wrapping:

```go
handler = RequireAuth(RequireAdmin(handler))
```

Execution order:
1. RequireAuth validates session
2. RequireAdmin checks role
3. Handler executes

### Context Values

The RequireAuth middleware injects:
- `user` - Current authenticated user
- `session` - Current session

These can be retrieved in handlers using:
```go
user, ok := auth.UserFromContext(r.Context())
session, ok := auth.SessionFromContext(r.Context())
```

### Error Handling

- 401 Unauthorized: Missing or invalid authentication
- 403 Forbidden: Authenticated but insufficient permissions
- Include WWW-Authenticate header on 401

### Public Endpoints

These endpoints should NOT have auth middleware:
- `/` - Home page
- `/invite/:token` - Guest RSVP page
- `/health` - Health check
- `/readiness` - Readiness check
- `/static/*` - Static assets

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md)
- **Go Middleware Pattern:** https://www.alexedwards.net/blog/making-and-using-middleware
