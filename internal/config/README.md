# Config Package

## Purpose

Provides environment-based configuration management and structured logging for TinyRSVP.

## Rules

- All configuration loaded from environment variables
- Type-safe configuration structs (NO `map[string]interface{}`)
- Fail fast on startup if configuration is invalid
- Sensitive data masked in logs
- HMAC secret auto-generated if not provided

## Structure

- [`config.go`](config.go) - Configuration structs and loading logic
- [`config_test.go`](config_test.go) - Configuration loading tests
- [`validation_test.go`](validation_test.go) - Validation tests
- [`logger.go`](logger.go) - Structured logging setup using log/slog
- [`logger_test.go`](logger_test.go) - Logger tests

## Key Functions

- `Load()` - Load and validate configuration from environment
- `InitLogger(level)` - Initialize structured JSON logger
- `GetLogLevelFromEnv()` - Get log level from LOG_LEVEL env var
- `Config.String()` - String representation with masked sensitive fields
- `Config.Validate()` - Validate all configuration fields

## Environment Variables

### Server
- `SERVER_HOST` - Server host (default: 0.0.0.0)
- `SERVER_PORT` - Server port (default: 8080)
- `SERVER_READ_TIMEOUT` - Read timeout (default: 10s)
- `SERVER_WRITE_TIMEOUT` - Write timeout (default: 10s)
- `SERVER_IDLE_TIMEOUT` - Idle timeout (default: 120s)
- `SERVER_BASE_URL` - Base URL (REQUIRED)

### Database
- `DATABASE_TYPE` - Database type: sqlite or postgres (default: sqlite)
- `DATABASE_PATH` - Database path (REQUIRED for SQLite)
- `DATABASE_MAX_OPEN_CONNS` - Max open connections (default: 25)
- `DATABASE_MAX_IDLE_CONNS` - Max idle connections (default: 5)
- `DATABASE_MAX_LIFETIME` - Connection max lifetime (default: 5m)

### OIDC
- `OIDC_ENABLED` - Enable OIDC (default: false)
- `OIDC_ISSUER_URL` - OIDC issuer URL (required if enabled)
- `OIDC_CLIENT_ID` - OIDC client ID (required if enabled)
- `OIDC_CLIENT_SECRET` - OIDC client secret (required if enabled)
- `OIDC_REDIRECT_URL` - OIDC redirect URL (required if enabled)

### Email
- `SMTP_HOST` - SMTP host (REQUIRED)
- `SMTP_PORT` - SMTP port (default: 587)
- `SMTP_USER` - SMTP username (optional)
- `SMTP_PASSWORD` - SMTP password (optional)
- `EMAIL_FROM` - From email address (REQUIRED)
- `EMAIL_FROM_NAME` - From name (default: TinyRSVP)

### Storage
- `STORAGE_TYPE` - Storage type: local or s3 (default: local)
- `STORAGE_LOCAL_PATH` - Local storage path (default: /data/uploads)
- `STORAGE_S3_BUCKET` - S3 bucket name (required if type=s3)
- `STORAGE_S3_REGION` - S3 region (required if type=s3)
- `STORAGE_S3_ENDPOINT` - S3 endpoint (optional)

### Security
- `SECURITY_SESSION_DURATION` - Session duration (default: 168h)
- `SECURITY_TOKEN_EXPIRY` - Token expiry (default: 720h)
- `SECURITY_HMAC_SECRET` - HMAC secret key (auto-generated if empty)

### Logging
- `LOG_LEVEL` - Log level: DEBUG, INFO, WARN, ERROR (default: INFO)

## Usage Example

```go
package main

import (
    "log"
    "github.com/yourusername/tinyrsvp/internal/config"
)

func main() {
    logLevel := config.GetLogLevelFromEnv()
    logger := config.InitLogger(logLevel)
    
    cfg, err := config.Load()
    if err != nil {
        logger.Error("Failed to load configuration", "error", err)
        os.Exit(1)
    }
    
    logger.Info("Configuration loaded", "config", cfg.String())
}
```

## Testing

Run tests with timeout:
```bash
go test -timeout 30s ./internal/config/...
```

Check coverage:
```bash
go test -timeout 30s -cover ./internal/config/...
```

## Validation Rules

- Server port: 1-65535
- All timeouts: > 0
- Database type: sqlite or postgres
- Max idle conns: <= max open conns
- OIDC issuer URL: HTTPS required when enabled
- Email format: must contain @
- Storage type: local or s3
- Session duration: 1h - 720h (30 days)
- Token expiry: 24h - 8760h (365 days)
- HMAC secret: >= 32 bytes
