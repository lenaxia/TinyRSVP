# Email Package

## Purpose

Provides email service interface for sending confirmation emails after RSVP submission and updates.

## Structure

- `service.go` - Email service interface and mock implementation

## Interface

```go
type Service interface {
    SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error
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

## Implementation Notes

- Email service is optional - RSVP service checks for nil before sending
- Emails are sent asynchronously in goroutines
- Email failures are logged but don't block RSVP submission
- No email sent if invite has no email address

## Future Implementation

The full email service implementation (Epic 05) will include:
- Template rendering with Go html/template
- Email queue processing with database persistence
- SMTP sender with TLS support
- Retry logic with exponential backoff (1min, 5min, 15min)
- Rate limiting (50/minute configurable)
- ICS attachment generation

## Dependencies

- `internal/models` - RSVP, Invite, Event, RSVPAnswer models
- `pkg/ics` - ICS calendar file generation

## Related Files

- Email templates: `templates/email/rsvp_confirmation.{html,txt}`
- ICS generator: `pkg/ics/generator.go`
- Email queue model: `internal/models/email_queue.go`
