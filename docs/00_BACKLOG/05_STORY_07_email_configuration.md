# User Story: Email Configuration Management

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-08

---

## User Story

As a **system administrator**, I want **centralized email configuration management** so that **SMTP settings can be configured via environment variables and validated on startup**.

---

## Acceptance Criteria

- [x] SMTP configuration loaded from environment variables
- [x] Configuration validation on application startup
- [x] Test connection on startup (optional, configurable)
- [x] Secure password handling (no logging)
- [x] Support for multiple SMTP providers (Gmail, SendGrid, etc.)
- [x] Default values for optional settings
- [x] Configuration errors fail fast with clear messages
- [x] All tests pass with timeout
- [x] Documentation for all configuration options

---

## Technical Details

### Configuration Structure

```go
package email

import (
    "fmt"
    "time"
)

type Config struct {
    // SMTP Server
    SMTPHost     string
    SMTPPort     int
    SMTPUsername string
    SMTPPassword string
    
    // Email Headers
    FromEmail string
    FromName  string
    
    // TLS/Security
    UseTLS       bool
    SkipVerify   bool
    
    // Connection
    Timeout        time.Duration
    MaxConnections int
    
    // Rate Limiting
    RateLimit int // emails per minute
    
    // Queue Processing
    QueuePollInterval time.Duration
    QueueBatchSize    int
    
    // Retry Policy
    MaxRetryAttempts int
    
    // Testing
    TestOnStartup bool
}

func LoadConfig() (*Config, error) {
    config := &Config{
        // Defaults
        SMTPPort:          587,
        UseTLS:            true,
        SkipVerify:        false,
        Timeout:           30 * time.Second,
        MaxConnections:    10,
        RateLimit:         50,
        QueuePollInterval: 60 * time.Second,
        QueueBatchSize:    50,
        MaxRetryAttempts:  4,
        TestOnStartup:     true,
    }
    
    // Load from environment
    if host := os.Getenv("SMTP_HOST"); host != "" {
        config.SMTPHost = host
    }
    
    if port := os.Getenv("SMTP_PORT"); port != "" {
        p, err := strconv.Atoi(port)
        if err != nil {
            return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
        }
        config.SMTPPort = p
    }
    
    if username := os.Getenv("SMTP_USERNAME"); username != "" {
        config.SMTPUsername = username
    }
    
    if password := os.Getenv("SMTP_PASSWORD"); password != "" {
        config.SMTPPassword = password
    }
    
    if from := os.Getenv("SMTP_FROM_EMAIL"); from != "" {
        config.FromEmail = from
    }
    
    if name := os.Getenv("SMTP_FROM_NAME"); name != "" {
        config.FromName = name
    }
    
    if tls := os.Getenv("SMTP_TLS"); tls != "" {
        config.UseTLS = tls == "true"
    }
    
    if skip := os.Getenv("SMTP_SKIP_VERIFY"); skip != "" {
        config.SkipVerify = skip == "true"
    }
    
    if timeout := os.Getenv("SMTP_TIMEOUT"); timeout != "" {
        t, err := time.ParseDuration(timeout)
        if err != nil {
            return nil, fmt.Errorf("invalid SMTP_TIMEOUT: %w", err)
        }
        config.Timeout = t
    }
    
    if limit := os.Getenv("EMAIL_RATE_LIMIT"); limit != "" {
        l, err := strconv.Atoi(limit)
        if err != nil {
            return nil, fmt.Errorf("invalid EMAIL_RATE_LIMIT: %w", err)
        }
        config.RateLimit = l
    }
    
    if test := os.Getenv("EMAIL_TEST_ON_STARTUP"); test != "" {
        config.TestOnStartup = test == "true"
    }
    
    // Validate
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid email configuration: %w", err)
    }
    
    return config, nil
}

func (c *Config) Validate() error {
    if c.SMTPHost == "" {
        return fmt.Errorf("SMTP_HOST is required")
    }
    
    if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
        return fmt.Errorf("SMTP_PORT must be between 1 and 65535")
    }
    
    if c.FromEmail == "" {
        return fmt.Errorf("SMTP_FROM_EMAIL is required")
    }
    
    if !isValidEmail(c.FromEmail) {
        return fmt.Errorf("SMTP_FROM_EMAIL is not a valid email address")
    }
    
    if c.Timeout <= 0 {
        return fmt.Errorf("SMTP_TIMEOUT must be positive")
    }
    
    if c.RateLimit <= 0 {
        return fmt.Errorf("EMAIL_RATE_LIMIT must be positive")
    }
    
    if c.MaxRetryAttempts < 1 || c.MaxRetryAttempts > 10 {
        return fmt.Errorf("MAX_RETRY_ATTEMPTS must be between 1 and 10")
    }
    
    return nil
}

func (c *Config) Sanitized() *Config {
    sanitized := *c
    if sanitized.SMTPPassword != "" {
        sanitized.SMTPPassword = "***REDACTED***"
    }
    return &sanitized
}
```

---

## Tasks

### Phase 1: Configuration Structure (TDD)
- [x] Define Config struct
- [x] Write test for LoadConfig
- [x] Implement LoadConfig
- [x] Write test for default values
- [x] Verify defaults applied

### Phase 2: Validation (TDD)
- [x] Write test for required fields
- [x] Implement required field validation
- [x] Write test for email format
- [x] Implement email validation
- [x] Write test for port range
- [x] Implement port validation
- [x] Write test for positive values
- [x] Implement positive value checks

### Phase 3: Security (TDD)
- [x] Write test for password sanitization
- [x] Implement Sanitized() method
- [x] Write test for no password logging
- [x] Verify passwords never logged
- [x] Write test for secure defaults
- [x] Implement secure defaults

### Phase 4: Provider Presets (TDD)
- [x] Documentation for Gmail preset
- [x] Documentation for SendGrid preset
- [x] Documentation for AWS SES preset
- [x] Documentation for Mailgun preset

### Phase 5: Integration Testing
- [x] Test with valid configuration
- [x] Test with invalid configuration
- [x] Test with missing required fields
- [x] Test with environment variables
- [x] Test startup validation
- [x] Test connection testing

---

## Dependencies

**Depends on:**
- Story 03: SMTP Sender (uses config)

**Blocks:**
- All email stories (need configuration)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Config struct defined
- [x] LoadConfig implemented
- [x] Validation working
- [x] All unit tests passing
- [x] Security verified
- [x] Documentation complete
- [x] Code reviewed

---

## Test Requirements

### Unit Tests

```go
func TestConfig_LoadConfig(t *testing.T) {
    // Set environment variables
    os.Setenv("SMTP_HOST", "smtp.gmail.com")
    os.Setenv("SMTP_PORT", "587")
    os.Setenv("SMTP_USERNAME", "user@gmail.com")
    os.Setenv("SMTP_PASSWORD", "secret")
    os.Setenv("SMTP_FROM_EMAIL", "noreply@example.com")
    os.Setenv("SMTP_FROM_NAME", "TinyRSVP")
    defer clearEnv()
    
    config, err := LoadConfig()
    if err != nil {
        t.Fatal(err)
    }
    
    if config.SMTPHost != "smtp.gmail.com" {
        t.Errorf("SMTPHost = %s, want smtp.gmail.com", config.SMTPHost)
    }
}

func TestConfig_Validate_MissingRequired(t *testing.T) {
    tests := []struct {
        name   string
        config *Config
        want   string
    }{
        {
            name:   "missing host",
            config: &Config{SMTPPort: 587, FromEmail: "test@example.com"},
            want:   "SMTP_HOST is required",
        },
        {
            name:   "missing from email",
            config: &Config{SMTPHost: "smtp.example.com", SMTPPort: 587},
            want:   "SMTP_FROM_EMAIL is required",
        },
        {
            name:   "invalid port",
            config: &Config{SMTPHost: "smtp.example.com", SMTPPort: 99999, FromEmail: "test@example.com"},
            want:   "SMTP_PORT must be between 1 and 65535",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if err == nil {
                t.Error("Validate() error = nil, want error")
            } else if !strings.Contains(err.Error(), tt.want) {
                t.Errorf("Validate() error = %v, want %v", err, tt.want)
            }
        })
    }
}

func TestConfig_Sanitized(t *testing.T) {
    config := &Config{
        SMTPHost:     "smtp.example.com",
        SMTPPassword: "secret123",
        FromEmail:    "test@example.com",
    }
    
    sanitized := config.Sanitized()
    
    if sanitized.SMTPPassword != "***REDACTED***" {
        t.Errorf("Sanitized password = %s, want ***REDACTED***", sanitized.SMTPPassword)
    }
    
    // Original should be unchanged
    if config.SMTPPassword != "secret123" {
        t.Error("Original config was modified")
    }
}
```

---

## Environment Variables

### Required
```bash
SMTP_HOST=smtp.gmail.com
SMTP_FROM_EMAIL=noreply@example.com
```

### Optional (with defaults)
```bash
SMTP_PORT=587                    # Default: 587
SMTP_USERNAME=user@gmail.com     # Default: empty
SMTP_PASSWORD=app-password       # Default: empty
SMTP_FROM_NAME=TinyRSVP          # Default: empty
SMTP_TLS=true                    # Default: true
SMTP_SKIP_VERIFY=false           # Default: false
SMTP_TIMEOUT=30s                 # Default: 30s
EMAIL_RATE_LIMIT=50              # Default: 50/min
EMAIL_TEST_ON_STARTUP=true       # Default: true
MAX_RETRY_ATTEMPTS=4             # Default: 4
QUEUE_POLL_INTERVAL=60s          # Default: 60s
QUEUE_BATCH_SIZE=50              # Default: 50
```

---

## Provider Configuration Examples

### Gmail
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_FROM_NAME="Your App Name"
SMTP_TLS=true
EMAIL_RATE_LIMIT=50
```

### SendGrid
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME="Your App Name"
SMTP_TLS=true
EMAIL_RATE_LIMIT=500
```

### AWS SES
```bash
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=your-smtp-username
SMTP_PASSWORD=your-smtp-password
SMTP_FROM_EMAIL=verified@yourdomain.com
SMTP_FROM_NAME="Your App Name"
SMTP_TLS=true
EMAIL_RATE_LIMIT=14
```

### Mailgun
```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USERNAME=postmaster@yourdomain.mailgun.org
SMTP_PASSWORD=your-mailgun-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME="Your App Name"
SMTP_TLS=true
EMAIL_RATE_LIMIT=100
```

---

## Security Best Practices

1. **Never log passwords** - Use Sanitized() for logging
2. **Use app-specific passwords** - Not account passwords
3. **Enable TLS** - Always use encrypted connections
4. **Verify certificates** - Only skip in development
5. **Rotate credentials** - Regularly update passwords
6. **Limit permissions** - Use send-only credentials

---

## Validation Rules

| Field | Rule | Error Message |
|-------|------|---------------|
| SMTP_HOST | Required, non-empty | "SMTP_HOST is required" |
| SMTP_PORT | 1-65535 | "SMTP_PORT must be between 1 and 65535" |
| SMTP_FROM_EMAIL | Required, valid email | "SMTP_FROM_EMAIL is not a valid email" |
| SMTP_TIMEOUT | Positive duration | "SMTP_TIMEOUT must be positive" |
| EMAIL_RATE_LIMIT | Positive integer | "EMAIL_RATE_LIMIT must be positive" |
| MAX_RETRY_ATTEMPTS | 1-10 | "MAX_RETRY_ATTEMPTS must be between 1 and 10" |

---

## Startup Validation

```go
func ValidateEmailConfig(config *Config) error {
    // Validate configuration
    if err := config.Validate(); err != nil {
        return fmt.Errorf("email configuration invalid: %w", err)
    }
    
    // Test connection if enabled
    if config.TestOnStartup {
        sender, err := NewSMTPSender(config)
        if err != nil {
            return fmt.Errorf("failed to create SMTP sender: %w", err)
        }
        
        ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
        defer cancel()
        
        if err := sender.TestConnection(ctx); err != nil {
            return fmt.Errorf("SMTP connection test failed: %w", err)
        }
        
        log.Println("✓ Email configuration validated successfully")
    }
    
    return nil
}
```

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **Config:** [`internal/config/config.go`](../../internal/config/config.go)
- **Story 03:** [05_STORY_03_smtp_sender.md](05_STORY_03_smtp_sender.md)
