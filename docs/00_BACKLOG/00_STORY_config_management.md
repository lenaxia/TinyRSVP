# User Story: Configuration Management

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 3 hours

---

## User Story

As a **developer**, I want **environment-based configuration management** so that **the application can be configured for different environments without code changes**.

---

## Acceptance Criteria

- [ ] Configuration loaded from environment variables
- [ ] Type-safe configuration struct defined
- [ ] All required configuration fields validated on startup
- [ ] Default values provided for optional fields
- [ ] Configuration errors fail fast with clear messages
- [ ] Structured logging package integrated (slog)
- [ ] No sensitive data in logs
- [ ] Configuration can be loaded in tests
- [ ] All tests pass with timeout

---

## Technical Details

### Configuration Structure

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    OIDC     OIDCConfig
    Email    EmailConfig
    Storage  StorageConfig
    Security SecurityConfig
}

type ServerConfig struct {
    Host         string
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
    BaseURL      string
}

type DatabaseConfig struct {
    Type         string
    Path         string
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  time.Duration
}

type OIDCConfig struct {
    Enabled      bool
    IssuerURL    string
    ClientID     string
    ClientSecret string
    RedirectURL  string
}

type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    SMTPUser     string
    SMTPPassword string
    FromEmail    string
    FromName     string
}

type StorageConfig struct {
    Type      string
    LocalPath string
    S3Bucket  string
    S3Region  string
    S3Endpoint string
}

type SecurityConfig struct {
    SessionDuration time.Duration
    TokenExpiry     time.Duration
    HMACSecretKey   string
}
```

### Environment Variables

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_IDLE_TIMEOUT=120s
SERVER_BASE_URL=https://rsvp.example.com

# Database
DATABASE_TYPE=sqlite
DATABASE_PATH=/data/tinyrsvp.db
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_MAX_LIFETIME=5m

# OIDC (Optional)
OIDC_ENABLED=false
OIDC_ISSUER_URL=
OIDC_CLIENT_ID=
OIDC_CLIENT_SECRET=
OIDC_REDIRECT_URL=

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
EMAIL_FROM=noreply@example.com
EMAIL_FROM_NAME=TinyRSVP

# Storage
STORAGE_TYPE=local
STORAGE_LOCAL_PATH=/data/uploads
STORAGE_S3_BUCKET=
STORAGE_S3_REGION=
STORAGE_S3_ENDPOINT=

# Security
SECURITY_SESSION_DURATION=168h
SECURITY_TOKEN_EXPIRY=720h
SECURITY_HMAC_SECRET=
```

### Validation Rules

**Server:**
- Port: 1-65535
- Timeouts: > 0
- BaseURL: Valid URL format

**Database:**
- Type: "sqlite" or "postgres"
- Path: Non-empty for SQLite
- MaxOpenConns: > 0
- MaxIdleConns: > 0, <= MaxOpenConns

**OIDC:**
- If enabled: IssuerURL, ClientID, ClientSecret, RedirectURL required
- IssuerURL: Valid HTTPS URL
- RedirectURL: Valid URL

**Email:**
- SMTPHost: Non-empty
- SMTPPort: 1-65535
- FromEmail: Valid email format

**Storage:**
- Type: "local" or "s3"
- If local: LocalPath required
- If s3: S3Bucket, S3Region required

**Security:**
- SessionDuration: >= 1h, <= 720h (30 days)
- TokenExpiry: >= 24h, <= 8760h (365 days)
- HMACSecretKey: Auto-generated if empty, >= 32 bytes

---

## Tasks

### Phase 1: Configuration Structure (TDD)
- [ ] Write test for loading valid configuration
- [ ] Write test for missing required fields
- [ ] Write test for invalid values
- [ ] Implement `Config` struct
- [ ] Implement `Load()` function
- [ ] Run tests (should pass)

### Phase 2: Environment Variable Parsing (TDD)
- [ ] Write test for parsing string values
- [ ] Write test for parsing integer values
- [ ] Write test for parsing duration values
- [ ] Write test for parsing boolean values
- [ ] Implement parsing functions
- [ ] Run tests (should pass)

### Phase 3: Validation (TDD)
- [ ] Write test for server validation
- [ ] Write test for database validation
- [ ] Write test for OIDC validation
- [ ] Write test for email validation
- [ ] Write test for storage validation
- [ ] Write test for security validation
- [ ] Implement `Validate()` method
- [ ] Run tests (should pass)

### Phase 4: Default Values (TDD)
- [ ] Write test for default server config
- [ ] Write test for default database config
- [ ] Write test for default security config
- [ ] Implement `SetDefaults()` method
- [ ] Run tests (should pass)

### Phase 5: Structured Logging Setup (TDD)
- [ ] Write test for logger initialization
- [ ] Write test for log level configuration
- [ ] Write test for JSON log output format
- [ ] Implement logger setup using log/slog
- [ ] Configure log levels (DEBUG, INFO, WARN, ERROR)
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Update `cmd/server/main.go` to load config
- [ ] Initialize structured logger
- [ ] Add config logging (mask sensitive fields)
- [ ] Test with various environment configurations
- [ ] Document all environment variables

---

## Testing Requirements

### Unit Tests

```go
func TestConfig_Load(t *testing.T) {
    tests := []struct {
        name    string
        env     map[string]string
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid minimal config",
            env: map[string]string{
                "SERVER_PORT": "8080",
                "DATABASE_PATH": "/tmp/test.db",
                "SMTP_HOST": "localhost",
                "EMAIL_FROM": "test@example.com",
            },
            wantErr: false,
        },
        {
            name: "missing required field",
            env: map[string]string{
                "SERVER_PORT": "8080",
            },
            wantErr: true,
            errMsg: "DATABASE_PATH is required",
        },
        {
            name: "invalid port",
            env: map[string]string{
                "SERVER_PORT": "99999",
                "DATABASE_PATH": "/tmp/test.db",
            },
            wantErr: true,
            errMsg: "invalid port",
        },
        {
            name: "invalid duration",
            env: map[string]string{
                "SERVER_PORT": "8080",
                "DATABASE_PATH": "/tmp/test.db",
                "SERVER_READ_TIMEOUT": "invalid",
            },
            wantErr: true,
            errMsg: "invalid duration",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            setTestEnv(t, tt.env)
            
            cfg, err := Load()
            if (err != nil) != tt.wantErr {
                t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && tt.errMsg != "" {
                if err == nil || !contains(err.Error(), tt.errMsg) {
                    t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
                }
            }
            
            if !tt.wantErr && cfg == nil {
                t.Error("Expected config, got nil")
            }
        })
    }
}

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: &Config{
                Server: ServerConfig{
                    Port: 8080,
                    ReadTimeout: 10 * time.Second,
                },
                Database: DatabaseConfig{
                    Type: "sqlite",
                    Path: "/tmp/test.db",
                    MaxOpenConns: 25,
                    MaxIdleConns: 5,
                },
            },
            wantErr: false,
        },
        {
            name: "invalid port",
            config: &Config{
                Server: ServerConfig{
                    Port: 0,
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestConfig_MaskSensitive(t *testing.T) {
    cfg := &Config{
        OIDC: OIDCConfig{
            ClientSecret: "secret123",
        },
        Email: EmailConfig{
            SMTPPassword: "password123",
        },
        Security: SecurityConfig{
            HMACSecretKey: "hmac-secret-key",
        },
    }
    
    masked := cfg.String()
    
    if contains(masked, "secret123") {
        t.Error("ClientSecret should be masked")
    }
    if contains(masked, "password123") {
        t.Error("SMTPPassword should be masked")
    }
    if contains(masked, "hmac-secret-key") {
        t.Error("HMACSecretKey should be masked")
    }
    if !contains(masked, "***") {
        t.Error("Expected masked values to contain ***")
    }
}
```

### Test Helpers

```go
func setTestEnv(t *testing.T, env map[string]string) {
    t.Helper()
    
    for k, v := range env {
        if err := os.Setenv(k, v); err != nil {
            t.Fatalf("Failed to set env var %s: %v", k, err)
        }
    }
    
    t.Cleanup(func() {
        for k := range env {
            os.Unsetenv(k)
        }
    })
}

func contains(s, substr string) bool {
    return len(s) > 0 && len(substr) > 0 && 
           len(s) >= len(substr) && 
           s != substr &&
           (s[:len(substr)] == substr || 
            s[len(s)-len(substr):] == substr ||
            len(s) > len(substr) && 
            func() bool {
                for i := 0; i <= len(s)-len(substr); i++ {
                    if s[i:i+len(substr)] == substr {
                        return true
                    }
                }
                return false
            }())
}
```

---

## Dependencies

**Depends on:** [00_STORY_go_module_setup.md](00_STORY_go_module_setup.md)  
**Blocks:** All other stories requiring configuration

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/config/...`)
- [ ] Test coverage >= 80%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Sensitive data properly masked in logs
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Notes

- Never log sensitive configuration values (passwords, secrets, tokens)
- Use `***` to mask sensitive fields in String() method
- Auto-generate HMAC secret if not provided
- Fail fast on startup if configuration is invalid
- Provide helpful error messages for configuration issues

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **HLD:** Section 20 (Deployment)
- **12-Factor App:** https://12factor.net/config
