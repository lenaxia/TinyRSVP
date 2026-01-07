# Auth Package

## Purpose

Provides OIDC authentication, session management, and user service for TinyRSVP admin and event manager authentication.

## Components

### OIDC Authenticator
- [`oidc.go`](oidc.go) - OIDC authentication implementation
- [`oidc_test.go`](oidc_test.go) - OIDC authenticator tests

### Session Management
- [`session.go`](session.go) - Session manager implementation
- [`session_test.go`](session_test.go) - Session manager tests

### User Service
- [`user_service.go`](user_service.go) - User creation and management
- [`user_service_test.go`](user_service_test.go) - User service tests

### HTTP Handlers
- [`handlers.go`](handlers.go) - HTTP handlers for login, callback, logout
- [`handlers_test.go`](handlers_test.go) - Handler tests
- [`login_test.go`](login_test.go) - Login handler tests
- [`callback_test.go`](callback_test.go) - Callback handler tests
- [`logout_test.go`](logout_test.go) - Logout handler tests

### Configuration
- [`config.go`](config.go) - Config conversion utilities

## Usage

### Initialize OIDC Authentication

```go
import (
    "github.com/yourusername/tinyrsvp/internal/auth"
    "github.com/yourusername/tinyrsvp/internal/config"
    "github.com/yourusername/tinyrsvp/internal/db/repositories"
)

appCfg, _ := config.Load()
database, _ := db.New(&appCfg.Database)

userRepo := repositories.NewUserRepository(database)
sessionRepo := repositories.NewSessionRepository(database)

userService := auth.NewUserService(userRepo)
sessionMgr := auth.NewSessionManager(sessionRepo, true)

oidcCfg := auth.NewOIDCConfigFromAppConfig(appCfg)
authenticator, err := auth.NewOIDCAuthenticator(oidcCfg, userService, sessionMgr)
if err != nil {
    log.Fatalf("Failed to create authenticator: %v", err)
}
```

### Wire HTTP Handlers

```go
loginHandler := auth.NewLoginHandler(authenticator)
callbackHandler := auth.NewCallbackHandler(authenticator, userService, sessionMgr)
logoutHandler := auth.NewLogoutHandler(authenticator)

http.Handle("/login", loginHandler)
http.Handle("/auth/callback", callbackHandler)
http.Handle("/logout", logoutHandler)
```

## Authentication Flow

1. User visits `/login`
2. Redirected to OIDC provider with state parameter
3. User authenticates with provider
4. Provider redirects to `/auth/callback` with authorization code
5. Exchange code for ID token
6. Verify ID token signature and claims
7. Create or update user in database
8. Create session
9. Set session cookie
10. Redirect to `/dashboard`

## Security Features

- State parameter CSRF protection
- ID token signature verification
- Secure session cookies (HttpOnly, Secure, SameSite=Lax)
- 7-day session expiration
- First user automatically becomes admin

## Testing

Run all auth tests:
```bash
go test -timeout 30s ./internal/auth/...
```

Run specific test suites:
```bash
go test -timeout 30s ./internal/auth/... -run TestOIDC
go test -timeout 30s ./internal/auth/... -run TestSession
go test -timeout 30s ./internal/auth/... -run TestUser
go test -timeout 30s ./internal/auth/... -run TestHandler
```

## Dependencies

- `github.com/coreos/go-oidc/v3/oidc` - OIDC library
- `golang.org/x/oauth2` - OAuth2 library
- `github.com/go-jose/go-jose/v4` - JWT/JWK library (testing)
- `internal/db/repositories` - User and session repositories
- `internal/models` - Domain models

## Environment Variables

Required when OIDC is enabled:
- `OIDC_ENABLED=true`
- `OIDC_ISSUER_URL` - OIDC provider URL (must be HTTPS)
- `OIDC_CLIENT_ID` - OAuth2 client ID
- `OIDC_CLIENT_SECRET` - OAuth2 client secret
- `OIDC_REDIRECT_URL` - Callback URL (e.g., https://rsvp.example.com/auth/callback)

## References

- **LLD:** [docs/lld/01_AUTH_LLD.md](../../docs/lld/01_AUTH_LLD.md)
- **Story:** [docs/00_BACKLOG/01_STORY_01_oidc_integration.md](../../docs/00_BACKLOG/01_STORY_01_oidc_integration.md)
