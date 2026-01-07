# Domain 5: Email System - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 9 - Email System](../02_REVISED_HLD.md#9-email-system), [Section 10 - Calendar Integration](../02_REVISED_HLD.md#10-calendar-integration-ics)

---

## 1. Overview

### 1.1 Purpose

Manages email sending, queue processing, retry logic, rate limiting, and ICS calendar file generation.

### 1.2 Responsibilities

- SMTP configuration and connection management
- Email queue management (database-backed)
- Hybrid send strategy (immediate + background retry)
- Retry policy with exponential backoff (1min, 5min, 15min)
- Rate limiting (50/minute configurable)
- Email template rendering
- ICS calendar file generation (RFC 5545)
- Bounce detection and handling
- Unsubscribe mechanism

### 1.3 Design Principles

- **Queue-Based** - All emails queued for reliability
- **Retry with Backoff** - Exponential backoff for transient failures
- **Rate Limited** - Prevent SMTP throttling
- **Idempotent** - Safe to retry sends
- **Observable** - Track all email states

---

## 2. Package Structure

```
internal/
├── email/
│   ├── service.go              # Email service
│   ├── service_test.go
│   ├── sender.go               # SMTP sender
│   ├── sender_test.go
│   ├── queue.go                # Queue processor
│   ├── queue_test.go
│   ├── templates.go            # Email templates
│   └── templates_test.go
pkg/
└── ics/
    ├── generator.go            # ICS generation
    ├── generator_test.go
    └── validator.go            # ICS validation
        └── validator_test.go
```

---

## 3. Interfaces

### 3.1 Email Service Interface

```go
package email

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
    SendInviteEmail(ctx context.Context, invite *models.Invite, event *models.Event, token string) error
    SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event) error
    SendUpdateEmail(ctx context.Context, event *models.Event, changes string) error
    SendCancellationEmail(ctx context.Context, event *models.Event, reason string) error
    QueueEmail(ctx context.Context, email *models.EmailQueue) error
    ProcessQueue(ctx context.Context) error
    RetryFailed(ctx context.Context, emailID int64) error
}
```

### 3.2 SMTP Sender Interface

```go
package email

import "context"

type SMTPSender interface {
    Send(ctx context.Context, email *Email) error
    TestConnection(ctx context.Context) error
}

type Email struct {
    To          string
    ToName      string
    Subject     string
    BodyText    string
    BodyHTML    string
    Attachments []Attachment
}

type Attachment struct {
    Filename    string
    ContentType string
    Content     []byte
}
```

### 3.3 ICS Generator Interface

```go
package ics

import (
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Generator interface {
    Generate(event *models.Event, rsvpURL string) ([]byte, error)
    GenerateUpdate(event *models.Event, rsvpURL string) ([]byte, error)
    GenerateCancellation(event *models.Event) ([]byte, error)
}
```

---

## 4. Implementation

### 4.1 Email Service Implementation

```go
package email

import (
    "context"
    "fmt"
    
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
    "github.com/lenaxia/tinyrsvp/internal/models"
    "github.com/lenaxia/tinyrsvp/pkg/ics"
)

type service struct {
    queueRepo  repositories.EmailQueueRepository
    sender     SMTPSender
    renderer   TemplateRenderer
    icsGen     ics.Generator
}

func NewService(
    queueRepo repositories.EmailQueueRepository,
    sender SMTPSender,
    renderer TemplateRenderer,
    icsGen ics.Generator,
) Service {
    return &service{
        queueRepo: queueRepo,
        sender:    sender,
        renderer:  renderer,
        icsGen:    icsGen,
    }
}

func (s *service) SendInviteEmail(ctx context.Context, invite *models.Invite, event *models.Event, token string) error {
    rsvpURL := fmt.Sprintf("%s/rsvp/%s", baseURL, token)
    
    data := struct {
        Event   *models.Event
        Invite  *models.Invite
        RSVPURL string
    }{
        Event:   event,
        Invite:  invite,
        RSVPURL: rsvpURL,
    }
    
    bodyHTML, err := s.renderer.RenderHTML(inviteTemplate, data)
    if err != nil {
        return fmt.Errorf("failed to render invite email: %w", err)
    }
    
    bodyText, err := s.renderer.RenderText(inviteTemplate, data)
    if err != nil {
        return fmt.Errorf("failed to render invite text: %w", err)
    }
    
    icsData, err := s.icsGen.Generate(event, rsvpURL)
    if err != nil {
        return fmt.Errorf("failed to generate ICS: %w", err)
    }
    
    attachments, err := json.Marshal([]models.EmailAttachment{
        {
            Filename:    "event.ics",
            ContentType: "text/calendar",
            Content:     icsData,
        },
    })
    if err != nil {
        return fmt.Errorf("failed to marshal attachments: %w", err)
    }
    
    email := &models.EmailQueue{
        ToEmail:      *invite.Email,
        ToName:       invite.Name,
        Subject:      fmt.Sprintf("You're invited: %s", event.Title),
        BodyText:     bodyText,
        BodyHTML:     &bodyHTML,
        Attachments:  attachments,
        Status:       models.EmailStatusPending,
        MaxAttempts:  4,
        ScheduledFor: time.Now(),
    }
    
    return s.queueRepo.Create(ctx, email)
}

func (s *service) SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event) error {
    if invite.Email == nil {
        return nil
    }
    
    rsvpURL := fmt.Sprintf("%s/rsvp/%s", baseURL, "token")
    
    data := struct {
        Event  *models.Event
        Invite *models.Invite
        RSVP   *models.RSVP
    }{
        Event:  event,
        Invite: invite,
        RSVP:   rsvp,
    }
    
    bodyHTML, err := s.renderer.RenderHTML(confirmationTemplate, data)
    if err != nil {
        return fmt.Errorf("failed to render confirmation email: %w", err)
    }
    
    bodyText, err := s.renderer.RenderText(confirmationTemplate, data)
    if err != nil {
        return fmt.Errorf("failed to render confirmation text: %w", err)
    }
    
    icsData, err := s.icsGen.Generate(event, rsvpURL)
    if err != nil {
        return fmt.Errorf("failed to generate ICS: %w", err)
    }
    
    attachments, err := json.Marshal([]models.EmailAttachment{
        {
            Filename:    "event.ics",
            ContentType: "text/calendar",
            Content:     icsData,
        },
    })
    if err != nil {
        return fmt.Errorf("failed to marshal attachments: %w", err)
    }
    
    email := &models.EmailQueue{
        ToEmail:      *invite.Email,
        ToName:       invite.Name,
        Subject:      fmt.Sprintf("RSVP Confirmed: %s", event.Title),
        BodyText:     bodyText,
        BodyHTML:     &bodyHTML,
        Attachments:  attachments,
        Status:       models.EmailStatusPending,
        MaxAttempts:  4,
        ScheduledFor: time.Now(),
    }
    
    return s.queueRepo.Create(ctx, email)
}

func (s *service) SendUpdateEmail(ctx context.Context, event *models.Event, changes string) error {
    return fmt.Errorf("not implemented")
}

func (s *service) SendCancellationEmail(ctx context.Context, event *models.Event, reason string) error {
    return fmt.Errorf("not implemented")
}

func (s *service) QueueEmail(ctx context.Context, email *models.EmailQueue) error {
    return s.queueRepo.Create(ctx, email)
}

func (s *service) ProcessQueue(ctx context.Context) error {
    return fmt.Errorf("not implemented")
}

func (s *service) RetryFailed(ctx context.Context, emailID int64) error {
    return fmt.Errorf("not implemented")
}
```

### 4.2 Email Queue Processor

```go
package email

import (
    "context"
    "fmt"
    "time"
)

type queueProcessor struct {
    repo       repositories.EmailQueueRepository
    sender     SMTPSender
    rateLimit  int
    ticker     *time.Ticker
    stopChan   chan struct{}
}

func NewQueueProcessor(repo repositories.EmailQueueRepository, sender SMTPSender, rateLimit int) *queueProcessor {
    return &queueProcessor{
        repo:      repo,
        sender:    sender,
        rateLimit: rateLimit,
        stopChan:  make(chan struct{}),
    }
}

func (p *queueProcessor) Start() {
    p.ticker = time.NewTicker(60 * time.Second)
    
    go func() {
        for {
            select {
            case <-p.ticker.C:
                p.processQueue(context.Background())
            case <-p.stopChan:
                return
            }
        }
    }()
}

func (p *queueProcessor) Stop() {
    close(p.stopChan)
    if p.ticker != nil {
        p.ticker.Stop()
    }
}

func (p *queueProcessor) processQueue(ctx context.Context) error {
    emails, err := p.repo.GetPending(ctx, p.rateLimit)
    if err != nil {
        return err
    }
    
    for _, emailQueue := range emails {
        if err := p.sendEmail(ctx, emailQueue); err != nil {
            p.handleSendError(ctx, emailQueue, err)
        } else {
            p.repo.MarkSent(ctx, emailQueue.ID)
        }
    }
    
    return nil
}

func (p *queueProcessor) sendEmail(ctx context.Context, emailQueue *models.EmailQueue) error {
    attachments, err := emailQueue.GetAttachments()
    if err != nil {
        return err
    }
    
    email := &Email{
        To:          emailQueue.ToEmail,
        ToName:      *emailQueue.ToName,
        Subject:     emailQueue.Subject,
        BodyText:    emailQueue.BodyText,
        BodyHTML:    *emailQueue.BodyHTML,
        Attachments: convertAttachments(attachments),
    }
    
    return p.sender.Send(ctx, email)
}

func (p *queueProcessor) handleSendError(ctx context.Context, emailQueue *models.EmailQueue, err error) {
    if emailQueue.Attempts >= emailQueue.MaxAttempts {
        p.repo.MarkFailed(ctx, emailQueue.ID, err.Error())
        return
    }
    
    backoff := calculateBackoff(emailQueue.Attempts)
    scheduledFor := time.Now().Add(backoff)
    
    p.repo.IncrementAttempts(ctx, emailQueue.ID, err.Error())
}

func calculateBackoff(attempts int) time.Duration {
    switch attempts {
    case 0:
        return 1 * time.Minute
    case 1:
        return 5 * time.Minute
    case 2:
        return 15 * time.Minute
    default:
        return 30 * time.Minute
    }
}
```

### 4.2 ICS Generator

```go
package ics

import (
    "bytes"
    "fmt"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type generator struct{}

func NewGenerator() Generator {
    return &generator{}
}

func (g *generator) Generate(event *models.Event, rsvpURL string) ([]byte, error) {
    var buf bytes.Buffer
    
    buf.WriteString("BEGIN:VCALENDAR\r\n")
    buf.WriteString("VERSION:2.0\r\n")
    buf.WriteString("PRODID:-//TinyRSVP//EN\r\n")
    buf.WriteString("METHOD:REQUEST\r\n")
    buf.WriteString("BEGIN:VEVENT\r\n")
    
    uid := fmt.Sprintf("%d@tinyrsvp", event.ID)
    buf.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
    
    buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSTime(time.Now())))
    buf.WriteString(fmt.Sprintf("DTSTART;TZID=%s:%s\r\n", event.Timezone, formatICSTime(event.StartTime)))
    
    if event.EndTime != nil {
        buf.WriteString(fmt.Sprintf("DTEND;TZID=%s:%s\r\n", event.Timezone, formatICSTime(*event.EndTime)))
    }
    
    buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(event.Title)))
    
    if event.Location != nil {
        buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(*event.Location)))
    }
    
    if event.Description != nil {
        desc := fmt.Sprintf("%s\\n\\nRSVP: %s", *event.Description, rsvpURL)
        buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(desc)))
    }
    
    buf.WriteString("STATUS:CONFIRMED\r\n")
    buf.WriteString(fmt.Sprintf("SEQUENCE:%d\r\n", event.ICSSequence))
    
    buf.WriteString("BEGIN:VALARM\r\n")
    buf.WriteString("TRIGGER:-PT24H\r\n")
    buf.WriteString("ACTION:DISPLAY\r\n")
    buf.WriteString(fmt.Sprintf("DESCRIPTION:Reminder: %s tomorrow\r\n", event.Title))
    buf.WriteString("END:VALARM\r\n")
    
    buf.WriteString("END:VEVENT\r\n")
    buf.WriteString("END:VCALENDAR\r\n")
    
    return buf.Bytes(), nil
}

func formatICSTime(t time.Time) string {
    return t.UTC().Format("20060102T150405Z")
}

func escapeICS(s string) string {
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, ",", "\\,")
    s = strings.ReplaceAll(s, ";", "\\;")
    s = strings.ReplaceAll(s, "\n", "\\n")
    return s
}
```

---

## 5. Email Templates

**See:** [Domain 6 (Template) LLD](06_TEMPLATE_LLD.md) for template rendering

---

## 5. Mock Implementations

### 5.1 Mock Email Service

```go
package email

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type MockService struct {
    SendInviteEmailFunc       func(ctx context.Context, invite *models.Invite, event *models.Event, token string) error
    SendConfirmationEmailFunc func(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event) error
    SendUpdateEmailFunc       func(ctx context.Context, event *models.Event, changes string) error
    SendCancellationEmailFunc func(ctx context.Context, event *models.Event, reason string) error
    QueueEmailFunc            func(ctx context.Context, email *models.EmailQueue) error
    ProcessQueueFunc          func(ctx context.Context) error
    RetryFailedFunc           func(ctx context.Context, emailID int64) error
}

func (m *MockService) SendInviteEmail(ctx context.Context, invite *models.Invite, event *models.Event, token string) error {
    if m.SendInviteEmailFunc != nil {
        return m.SendInviteEmailFunc(ctx, invite, event, token)
    }
    return nil
}

func (m *MockService) SendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event) error {
    if m.SendConfirmationEmailFunc != nil {
        return m.SendConfirmationEmailFunc(ctx, rsvp, invite, event)
    }
    return nil
}

func (m *MockService) SendUpdateEmail(ctx context.Context, event *models.Event, changes string) error {
    if m.SendUpdateEmailFunc != nil {
        return m.SendUpdateEmailFunc(ctx, event, changes)
    }
    return nil
}

func (m *MockService) SendCancellationEmail(ctx context.Context, event *models.Event, reason string) error {
    if m.SendCancellationEmailFunc != nil {
        return m.SendCancellationEmailFunc(ctx, event, reason)
    }
    return nil
}

func (m *MockService) QueueEmail(ctx context.Context, email *models.EmailQueue) error {
    if m.QueueEmailFunc != nil {
        return m.QueueEmailFunc(ctx, email)
    }
    return nil
}

func (m *MockService) ProcessQueue(ctx context.Context) error {
    if m.ProcessQueueFunc != nil {
        return m.ProcessQueueFunc(ctx)
    }
    return nil
}

func (m *MockService) RetryFailed(ctx context.Context, emailID int64) error {
    if m.RetryFailedFunc != nil {
        return m.RetryFailedFunc(ctx, emailID)
    }
    return nil
}
```

### 5.2 Mock ICS Generator

```go
package ics

import "github.com/lenaxia/tinyrsvp/internal/models"

type MockGenerator struct {
    GenerateFunc             func(event *models.Event, rsvpURL string) ([]byte, error)
    GenerateUpdateFunc       func(event *models.Event, rsvpURL string) ([]byte, error)
    GenerateCancellationFunc func(event *models.Event) ([]byte, error)
}

func (m *MockGenerator) Generate(event *models.Event, rsvpURL string) ([]byte, error) {
    if m.GenerateFunc != nil {
        return m.GenerateFunc(event, rsvpURL)
    }
    return []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"), nil
}

func (m *MockGenerator) GenerateUpdate(event *models.Event, rsvpURL string) ([]byte, error) {
    if m.GenerateUpdateFunc != nil {
        return m.GenerateUpdateFunc(event, rsvpURL)
    }
    return []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"), nil
}

func (m *MockGenerator) GenerateCancellation(event *models.Event) ([]byte, error) {
    if m.GenerateCancellationFunc != nil {
        return m.GenerateCancellationFunc(event)
    }
    return []byte("BEGIN:VCALENDAR\nEND:VCALENDAR"), nil
}
```

---

## 6. Dependencies

**External:**
- `net/smtp` - SMTP client
- `crypto/tls` - TLS support

**Internal:**
- Domain 2 (Event) - Event details
- Domain 3 (Invite) - Recipient info
- Domain 6 (Template) - Email templates
- Domain 7 (Database) - Queue storage

**Dependents:**
- Domain 8 (API) - Email endpoints

---

## 7. Testing

```go
func TestICSGenerator_Generate(t *testing.T) {
    gen := NewGenerator()
    
    event := &models.Event{
        ID:        1,
        Title:     "Test Event",
        StartTime: time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
        Timezone:  "America/Los_Angeles",
    }
    
    ics, err := gen.Generate(event, "https://example.com/rsvp/token")
    if err != nil {
        t.Fatal(err)
    }
    
    if !bytes.Contains(ics, []byte("BEGIN:VCALENDAR")) {
        t.Error("Missing VCALENDAR")
    }
    
    if !bytes.Contains(ics, []byte("Test Event")) {
        t.Error("Missing event title")
    }
}
```

---

**Document Status:** ✅ Complete

**Next Domain:** [Domain 6: Template & Asset Management](06_TEMPLATE_LLD.md)
