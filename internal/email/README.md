# Email Package

## Purpose

Provides email service interface for sending confirmation emails after RSVP submission and updates, along with a background queue processor for reliable email delivery.

## Structure

- `service.go` - Email service interface and mock implementation
- `processor.go` - Background queue processor for email delivery
- `processor_test.go` - Queue processor tests

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
}
```

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

## Future Implementation

The full email service implementation (Epic 05) will include:
- Story 03: SMTP sender with TLS support
- Story 04: Template rendering with Go html/template
- Story 06: Rate limiting implementation
- Story 07: Email configuration management
- Story 08: Monitoring and observability

## Dependencies

- `internal/models` - RSVP, Invite, Event, RSVPAnswer models
- `pkg/ics` - ICS calendar file generation

## Related Files

- Email templates: `templates/email/rsvp_confirmation.{html,txt}`
- ICS generator: `pkg/ics/generator.go`
- Email queue model: `internal/models/email_queue.go`
