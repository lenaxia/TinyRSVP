# Middleware Package

## Purpose

This package provides HTTP middleware for authentication and authorization in TinyRSVP.

## Middleware Functions

### RequireAuth

Validates session and injects user and session into request context.

**Usage:**
```go
authMiddleware := middleware.RequireAuth(sessionMgr, userService)
mux.Handle("/protected", authMiddleware(handler))
```

**Behavior:**
- Extracts session ID from cookie
- Validates session (not expired)
- Loads user from database
- Refreshes session last accessed time
- Injects user and session into context
- Returns 401 Unauthorized on any failure

### RequireAdmin

Restricts access to admin users only. Must be used after RequireAuth.

**Usage:**
```go
authMiddleware := middleware.RequireAuth(sessionMgr, userService)
adminMiddleware := middleware.RequireAdmin
mux.Handle("/admin", authMiddleware(adminMiddleware(handler)))
```

**Behavior:**
- Checks for user in context
- Verifies user has admin role
- Returns 401 if no user in context
- Returns 403 Forbidden if user is not admin

### RequireEventManager

Allows both admin and event manager users. Must be used after RequireAuth.

**Usage:**
```go
authMiddleware := middleware.RequireAuth(sessionMgr, userService)
managerMiddleware := middleware.RequireEventManager
mux.Handle("/events", authMiddleware(managerMiddleware(handler)))
```

**Behavior:**
- Checks for user in context
- Verifies user has admin or event manager role
- Returns 401 if no user in context
- Returns 403 Forbidden if user lacks required role

## Middleware Chaining

Middleware is applied in reverse order of wrapping:

```go
handler = RequireAuth(RequireAdmin(handler))
```

Execution order:
1. RequireAuth validates session
2. RequireAdmin checks role
3. Handler executes

## Context Values

RequireAuth injects these values into the request context:
- `user` - Current authenticated user (*models.User)
- `session` - Current session (*models.Session)

Retrieve in handlers using:
```go
user, ok := auth.UserFromContext(r.Context())
session, ok := auth.SessionFromContext(r.Context())
```

## Error Responses

- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Authenticated but insufficient permissions

## Public Endpoints

These endpoints should NOT have auth middleware:
- `/` - Home page
- `/invite/:token` - Guest RSVP page
- `/health` - Health check
- `/readiness` - Readiness check
- `/static/*` - Static assets

## Testing

All middleware functions have comprehensive unit tests covering:
- Happy path scenarios
- Missing authentication
- Invalid sessions
- Expired sessions
- Missing users
- Role-based access control
- Middleware chaining
- Context injection

Run tests:
```bash
go test -timeout 30s ./internal/middleware/...
```
