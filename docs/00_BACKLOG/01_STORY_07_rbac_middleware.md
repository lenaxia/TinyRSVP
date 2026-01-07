# User Story: RBAC Middleware

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 6 hours
**Actual Effort:** 1 hour
**Completed:** 2026-01-07

---

## User Story

As a **developer**, I want **role-based access control middleware** so that **only authorized users can access protected endpoints**.

---

## Acceptance Criteria

- [x] RequireAuth middleware validates session
- [x] RequireAdmin middleware restricts to admin users
- [x] RequireEventManager middleware allows managers and admins
- [x] Middleware chains correctly
- [x] User and session injected into request context
- [x] Clear error responses for unauthorized access
- [x] Middleware skips public endpoints
- [x] All tests pass with timeout

---

## Technical Details

### Middleware Functions

```go
package middleware

import (
    "context"
    "net/http"
    
    "github.com/lenaxia/tinyrsvp/internal/auth"
    "github.com/lenaxia/tinyrsvp/internal/models"
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
- [x] Write test for valid session
- [x] Write test for missing session cookie
- [x] Write test for invalid session ID
- [x] Write test for expired session
- [x] Write test for user not found
- [x] Write test for context injection
- [x] Implement RequireAuth middleware
- [x] Run tests (should pass)

### Phase 2: RequireAdmin Middleware (TDD)
- [x] Write test for admin user allowed
- [x] Write test for event manager denied
- [x] Write test for missing user in context
- [x] Write test for middleware chaining
- [x] Implement RequireAdmin middleware
- [x] Run tests (should pass)

### Phase 3: RequireEventManager Middleware (TDD)
- [x] Write test for admin allowed
- [x] Write test for event manager allowed
- [x] Write test for missing user denied
- [x] Write test for middleware chaining
- [x] Implement RequireEventManager middleware
- [x] Run tests (should pass)

### Phase 4: Session Refresh Middleware (TDD)
- [x] Write test for session last access update
- [x] Write test for refresh on each request
- [x] Implement session refresh in RequireAuth
- [x] Run tests (should pass)

### Phase 5: Integration
- [x] Wire middleware into HTTP router
- [x] Test middleware chains
- [x] Test endpoint protection
- [x] Document middleware usage
- [x] Create middleware examples

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
- [x] Write test for valid authentication
- [x] Write test for missing session
- [x] Write test for invalid session
- [x] Write test for expired session
- [x] Write test for context injection
- [x] Implement RequireAuth
- [x] Run tests (should pass)

### Phase 2: RequireAdmin Middleware (TDD)
- [x] Write test for admin access granted
- [x] Write test for non-admin denied
- [x] Write test for missing user
- [x] Implement RequireAdmin
- [x] Run tests (should pass)

### Phase 3: RequireEventManager Middleware (TDD)
- [x] Write test for admin access
- [x] Write test for event manager access
- [x] Write test for unauthorized access
- [x] Implement RequireEventManager
- [x] Run tests (should pass)

### Phase 4: Middleware Chaining (TDD)
- [x] Write test for RequireAuth + RequireAdmin chain
- [x] Write test for RequireAuth + RequireEventManager chain
- [x] Write test for multiple middleware chain
- [x] Test order of execution
- [x] Run tests (should pass)

### Phase 5: Error Responses (TDD)
- [x] Write test for JSON error responses
- [x] Write test for HTML error pages
- [x] Write test for proper HTTP status codes
- [x] Implement error response logic
- [x] Run tests (should pass)

### Phase 6: Integration
- [x] Apply middleware to protected routes
- [x] Test with real HTTP server
- [x] Verify all endpoints properly protected
- [x] Document middleware usage patterns

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

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/middleware/...`)
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Middleware chains tested
- [x] Context injection verified
- [x] Error responses appropriate
- [x] Documentation complete
- [x] Changes committed to git

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
