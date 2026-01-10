# Worklog: Authentication Routes Implementation

**Date:** 2026-01-10  
**Story:** [08_STORY_06_login_routes.md](../00_BACKLOG/08_STORY_06_login_routes.md)  
**Status:** Complete

---

## Summary

Implemented authentication route handlers in `internal/handlers` package that integrate with the existing `internal/auth` authentication system. The implementation provides secure login, OIDC authentication, callback handling, and logout functionality with proper error handling, CSRF protection, and return URL validation.

---

## Changes Made

### New Files Created

1. **`internal/handlers/auth.go`** - Authentication handlers
   - `AuthHandlers` struct with authenticator dependency
   - `ShowLogin()` - Displays login page with return URL
   - `OIDCLogin()` - Initiates OIDC authentication flow
   - `OIDCCallback()` - Handles OIDC provider callback
   - `Logout()` - Handles logout with POST-only restriction
   - `validateReturnURL()` - Prevents open redirect vulnerabilities
   - `Authenticator` interface for dependency injection
   - `AuthResult` struct for authentication results

2. **`internal/handlers/auth_test.go`** - Unit tests
   - `TestShowLogin_ValidReturnURL` - Tests login page with various return URLs
   - `TestShowLogin_InvalidReturnURL` - Tests open redirect prevention
   - `TestOIDCLogin_RedirectsToProvider` - Tests OIDC redirect
   - `TestOIDCLogin_InvalidReturnURL` - Tests return URL validation
   - `TestOIDCLogin_AuthenticatorError` - Tests error handling
   - `TestOIDCCallback_Success` - Tests successful callback
   - `TestOIDCCallback_Error` - Tests callback error handling
   - `TestLogout_Success` - Tests successful logout
   - `TestLogout_MethodNotAllowed` - Tests POST-only restriction
   - `TestLogout_Error` - Tests logout error handling
   - `TestValidateReturnURL` - Tests URL validation logic
   - Mock authenticator for testing

3. **`internal/handlers/auth_integration_test.go`** - Integration tests
   - `TestAuthFlow_Integration_LoginToCallback` - Full login flow
   - `TestAuthFlow_Integration_LoginPageToOIDC` - Login page to OIDC
   - `TestAuthFlow_Integration_LogoutFlow` - Logout flow
   - `TestAuthFlow_Integration_ErrorHandling` - Error scenarios
   - `TestAuthFlow_Integration_ContentNegotiation` - JSON vs HTML responses
   - `TestAuthFlow_Integration_RequestIDPropagation` - Request ID tracking

4. **`internal/handlers/router_auth_test.go`** - Router integration tests
   - `TestRouter_WithAuthHandlers` - Tests all routes work
   - `TestRouter_WithAuthHandlers_ListRoutes` - Verifies route registration
   - `TestRouter_WithAuthHandlers_ReturnURLFlow` - Tests return URL preservation
   - `TestRouter_WithAuthHandlers_InvalidReturnURL` - Tests URL validation
   - `TestRouter_WithAuthHandlers_LogoutMethodRestriction` - Tests POST-only
   - `TestRouter_WithAuthHandlers_LogoutWithCSRF` - Tests CSRF protection

### Modified Files

1. **`internal/handlers/router.go`**
   - Added `AuthHandlers` field to `RouterHandlers` struct
   - Updated route registration to use `AuthHandlers` when provided
   - Maintains backward compatibility with old handler fields
   - Routes registered:
     - `GET /login` → `ShowLogin`
     - `GET /auth/oidc/login` → `OIDCLogin`
     - `GET /auth/oidc/callback` → `OIDCCallback`
     - `POST /logout` → `Logout`

2. **`internal/handlers/router_test.go`**
   - Updated route group tests to match new route structure
   - Maintained fallback route tests for backward compatibility

---

## Implementation Details

### Security Features

1. **Open Redirect Prevention**
   - `validateReturnURL()` function validates all return URLs
   - Rejects external URLs (e.g., `https://evil.com`)
   - Rejects protocol-relative URLs (e.g., `//evil.com`)
   - Rejects dangerous protocols (e.g., `javascript:`, `data:`)
   - Only allows absolute paths starting with `/`
   - Defaults to `/` if empty or invalid

2. **CSRF Protection**
   - Logout requires POST method only
   - CSRF middleware automatically protects POST requests
   - Tests verify CSRF token requirement
   - Integration test demonstrates proper CSRF flow

3. **Error Handling**
   - Uses centralized `HandleError()` function
   - Content negotiation (JSON vs HTML)
   - Request ID propagation for debugging
   - Proper HTTP status codes

4. **Return URL Flow**
   - Login page preserves return URL in OIDC login link
   - URL-encoded in query parameters
   - Validated before use
   - Defaults to `/` if not provided

### Handler Design

The handlers follow the established patterns:
- Dependency injection via constructor
- Thin handlers that delegate to authenticator
- Proper error handling with `HandleError()`
- Integration with router middleware (CSRF, rate limiting, etc.)

### Testing Strategy

**Unit Tests (auth_test.go):**
- Test each handler in isolation
- Mock authenticator for controlled behavior
- Multiple happy path scenarios
- Multiple unhappy path scenarios
- Edge cases (empty URLs, invalid URLs, etc.)

**Integration Tests (auth_integration_test.go):**
- Test complete authentication flows
- Test error handling across handlers
- Test content negotiation
- Test request ID propagation

**Router Integration Tests (router_auth_test.go):**
- Test handlers work through router
- Test CSRF protection
- Test route registration
- Test method restrictions

---

## Test Results

All tests passing:
```bash
go test -timeout 30s ./internal/handlers/...
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	0.740s
```

Full project tests:
```bash
go test -timeout 30s ./...
# All packages pass
```

---

## Usage Example

```go
import (
    "github.com/lenaxia/tinyrsvp/internal/handlers"
    "github.com/lenaxia/tinyrsvp/internal/auth"
)

// Create authenticator (OIDC or ForwardAuth)
authenticator, err := auth.NewOIDCAuthenticator(cfg, userService, sessionMgr)
if err != nil {
    log.Fatal(err)
}

// Create auth handlers
authHandlers := handlers.NewAuthHandlers(authenticator)

// Create router with auth handlers
router := handlers.NewRouter(&handlers.RouterHandlers{
    AuthHandlers: authHandlers,
})

// Start server
http.ListenAndServe(":8080", router)
```

---

## Routes Implemented

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/login` | `ShowLogin` | Display login page with return URL |
| GET | `/auth/oidc/login` | `OIDCLogin` | Redirect to OIDC provider |
| GET | `/auth/oidc/callback` | `OIDCCallback` | Handle OIDC callback |
| POST | `/logout` | `Logout` | Logout and clear session |

---

## Security Considerations

1. **Return URL Validation** - Prevents open redirect attacks
2. **CSRF Protection** - POST /logout requires valid CSRF token
3. **Method Restrictions** - Logout only accepts POST
4. **Error Handling** - Proper error messages without leaking details
5. **Rate Limiting** - Applied via router middleware
6. **Request ID Tracking** - All errors include request ID

---

## Integration Points

### With internal/auth Package

The handlers delegate to the `Authenticator` interface:
- `HandleLogin()` - Initiates authentication
- `HandleCallback()` - Processes provider callback
- `HandleLogout()` - Clears session

This allows the handlers to work with both:
- `OIDCAuthenticator` - For OIDC providers
- `ForwardAuthenticator` - For reverse proxy auth

### With Router Middleware

The handlers benefit from router middleware:
- Request ID generation
- Logging
- CSRF protection
- Rate limiting
- Security headers
- Timeout handling
- Recovery from panics

---

## Next Steps

The authentication routes are now ready for integration with:
1. Dashboard route (08_STORY_07)
2. Event management routes
3. User management routes

All protected routes should use the `AuthMiddleware` to require authentication.

---

## Notes

- Login page template is inline HTML (no separate template file needed yet)
- Can be extracted to `templates/web/login.html` if customization needed
- Authenticator interface allows easy testing with mocks
- All handlers use centralized error handling
- CSRF protection is automatic via router middleware
