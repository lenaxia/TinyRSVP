# Email Package

## Purpose

Provides email service interface for sending confirmation emails after RSVP submission and updates, along with a background queue processor for reliable email delivery.

## Structure

- `config.go` - Email configuration management with environment variable loading
- `config_test.go` - Configuration tests
- `service.go` - Email service interface and mock implementation
- `processor.go` - Background queue processor for email delivery
- `processor_test.go` - Queue processor tests
- `smtp_sender.go` - SMTP email sender implementation
- `smtp_sender_test.go` - SMTP sender tests
- `rate_limiter.go` - Rate limiting for email sending
- `rate_limiter_test.go` - Rate limiter tests
- `renderer.go` - Email template rendering
- `renderer_test.go` - Renderer tests
- `stubs.go` - Test stubs for email components

## Configuration

The email package uses a centralized configuration system that loads settings from environment variables.

### Loading Configuration

```go
import "github.com/lenaxia/tinyrsvp/internal/email"

config, err := email.LoadConfig()
if err != nil {
    log.Fatalf("Failed to load email config: %v", err)
}

sender, err := email.NewSMTPSender(config)
if err != nil {
    log.Fatalf("Failed to create SMTP sender: %v", err)
}
```

### Required Environment Variables

- `SMTP_HOST` - SMTP server hostname (e.g., smtp.gmail.com)
- `SMTP_FROM_EMAIL` - From email address (must be valid email format)

### Optional Environment Variables (with defaults)

- `SMTP_PORT` - SMTP port (default: 587)
- `SMTP_USERNAME` - SMTP authentication username
- `SMTP_PASSWORD` - SMTP authentication password
- `SMTP_FROM_NAME` - From name for emails
- `SMTP_TLS` - Use TLS encryption (default: true)
- `SMTP_SKIP_VERIFY` - Skip certificate verification (default: false)
- `SMTP_TIMEOUT` - Connection timeout (default: 30s)
- `EMAIL_RATE_LIMIT` - Emails per minute (default: 50)
- `EMAIL_TEST_ON_STARTUP` - Test connection on startup (default: true)
- `MAX_RETRY_ATTEMPTS` - Max retry attempts (default: 4, range: 1-10)
- `QUEUE_POLL_INTERVAL` - Queue polling interval (default: 60s)
- `QUEUE_BATCH_SIZE` - Batch size for processing (default: 50)

### Security Features

- Password sanitization via `Config.Sanitized()` method
- Passwords are never logged
- TLS encryption enabled by default
- Certificate verification enabled by default

### Provider Examples

See [`docs/00_BACKLOG/05_STORY_07_email_configuration.md`](../../docs/00_BACKLOG/05_STORY_07_email_configuration.md) for configuration examples for Gmail, SendGrid, AWS SES, and Mailgun.

## Interfaces

### Email Service
```go
type Service interface {
    SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error
}
```

### Queue Processor
```go
type QueueProcessor interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    ProcessBatch(ctx context.Context) error
}
```

### SMTP Sender
```go
type SMTPSender interface {
    Send(ctx context.Context, msg *SMTPMessage) error
}
```

### Rate Limiter
```go
type RateLimiter interface {
	Allow() bool
	AvailableSlots() int
	WaitTime() time.Duration
	Record()
	Reset()
}
```

Implementation: Sliding window algorithm with thread-safe operations.
Default: 50 emails per minute (configurable).

## Usage

### In RSVP Service

```go
import "github.com/lenaxia/tinyrsvp/internal/email"

// Create service with email support
rsvpService := rsvp.NewServiceWithEmail(
    database,
    inviteService,
    inviteRepo,
    eventRepo,
    rsvpRepo,
    answerRepo,
    questionRepo,
    emailService,
)
```

### Mock for Testing

```go
mockEmail := &email.MockService{
    SendConfirmationEmailFunc: func(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error {
        return nil
    },
}
```

## Queue Processor

The queue processor is a background worker that processes pending emails from the database queue.

### Features
- Configurable poll interval (default 60 seconds)
- Batch processing with configurable size
- Rate limiting integration
- Exponential backoff retry logic (1min, 5min, 15min, 30min)
- Optimistic locking to prevent duplicate sends
- Graceful shutdown support
- Automatic failure handling after max attempts

### Usage
```go
processor := email.NewQueueProcessor(
    emailQueueRepo,
    smtpSender,
    rateLimiter,
    50,              // batch size
    60*time.Second,  // poll interval
)

// Start processor in background
go func() {
    if err := processor.Start(ctx); err != nil {
        log.Printf("Queue processor error: %v", err)
    }
}()

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := processor.Stop(ctx); err != nil {
    log.Printf("Failed to stop processor: %v", err)
}
```

### Retry Logic
- Attempt 1: Retry after 1 minute
- Attempt 2: Retry after 5 minutes
- Attempt 3: Retry after 15 minutes
- Attempt 4+: Retry after 30 minutes
- After max attempts: Mark as permanently failed

## Implementation Notes

- Email service is optional - RSVP service checks for nil before sending
- Emails are sent asynchronously in goroutines
- Email failures are logged but don't block RSVP submission
- No email sent if invite has no email address
- Queue processor uses optimistic locking to prevent duplicate sends
- Rate limiter is checked before each send operation

## Implementation Status

Completed:
- Story 01: Email queue repository ✓
- Story 02: Email queue processor ✓
- Story 03: SMTP sender with TLS support ✓
- Story 04: Template rendering with Go html/template ✓
- Story 06: Rate limiting implementation ✓
- Story 07: Email configuration management ✓

Remaining (Epic 05):
- Story 05: Retry logic (partially complete)
- Story 08: Monitoring and observability

## Rate Limiter

The rate limiter uses a sliding window algorithm to enforce email sending limits:

### Features
- Sliding window algorithm for accurate rate limiting
- Thread-safe for concurrent access
- Configurable limit (default: 50 emails per minute)
- Automatic cleanup of expired timestamps
- Returns available slots for batch sizing
- Provides wait time until next available slot

### Usage
```go
// Create rate limiter with 50 emails per minute
limiter := email.NewRateLimiter(50)

// Check if send is allowed
if limiter.Allow() {
    // Send email
    err := sender.Send(ctx, msg)
    if err == nil {
        // Record successful send
        limiter.Record()
    }
}

// Get available slots for batch sizing
available := limiter.AvailableSlots()

// Get wait time if at limit
waitTime := limiter.WaitTime()
if waitTime > 0 {
    time.Sleep(waitTime)
}
```

### Algorithm
The sliding window algorithm maintains a list of timestamps for recent sends:
1. On each operation, expired timestamps (older than 1 minute) are removed
2. `Allow()` checks if current count is below the limit
3. `Record()` adds a new timestamp for the current send
4. `AvailableSlots()` returns remaining capacity
5. `WaitTime()` calculates time until oldest timestamp expires

## Dependencies

- `internal/models` - RSVP, Invite, Event, RSVPAnswer models
- `pkg/ics` - ICS calendar file generation

## Related Files

- Email templates: `templates/email/rsvp_confirmation.{html,txt}`
- ICS generator: `pkg/ics/generator.go`
- Email queue model: `internal/models/email_queue.go`
