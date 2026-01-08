# User Story: RSVP Confirmation Email

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **guest**, I want **to receive a confirmation email after submitting my RSVP** so that **I have a record of my response**.

---

## Acceptance Criteria

- [ ] Confirmation email sent after successful RSVP submission
- [ ] Email sent after RSVP update
- [ ] Email includes RSVP summary (response, plus ones)
- [ ] Email includes event details
- [ ] Email includes preference answers
- [ ] ICS calendar file attached
- [ ] "Update RSVP" link included
- [ ] Email template mobile-friendly
- [ ] Email queued asynchronously (non-blocking)
- [ ] Retry logic for failed sends
- [ ] All tests pass with timeout

---

## Technical Details

### Email Template Location
- [`templates/email/rsvp_confirmation.html`](../../templates/email/rsvp_confirmation.html)
- [`templates/email/rsvp_confirmation.txt`](../../templates/email/rsvp_confirmation.txt)

### Email Data Structure

```go
type RSVPConfirmationEmail struct {
    GuestName      string
    EventTitle     string
    EventDate      string
    EventLocation  string
    Response       string
    PlusOnes       int
    Answers        []AnswerDisplay
    UpdateURL      string
    ICSAttachment  []byte
}

type AnswerDisplay struct {
    Question string
    Answer   string
}
```

### Email Sending Flow

```go
func (s *service) SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) error {
    // ... save RSVP ...
    
    // Queue confirmation email (async)
    go func() {
        if err := s.sendConfirmationEmail(context.Background(), rsvp, invite, event); err != nil {
            log.Printf("Failed to send confirmation email: %v", err)
        }
    }()
    
    return nil
}

func (s *service) sendConfirmationEmail(ctx context.Context, rsvp *models.RSVP, invite *models.Invite, event *models.Event) error {
    // Generate ICS file
    icsData, err := s.icsGenerator.Generate(event)
    if err != nil {
        return err
    }
    
    // Prepare email data
    emailData := &RSVPConfirmationEmail{
        GuestName:     invite.Name,
        EventTitle:    event.Title,
        EventDate:     formatEventDate(event),
        EventLocation: event.Location,
        Response:      string(rsvp.Response),
        PlusOnes:      rsvp.PlusOnes,
        UpdateURL:     generateUpdateURL(invite.Token),
        ICSAttachment: icsData,
    }
    
    // Queue email
    return s.emailQueue.Enqueue(ctx, &email.Message{
        To:          invite.Email,
        Subject:     fmt.Sprintf("RSVP Confirmation: %s", event.Title),
        HTMLBody:    renderTemplate("rsvp_confirmation.html", emailData),
        TextBody:    renderTemplate("rsvp_confirmation.txt", emailData),
        Attachments: []email.Attachment{
            {
                Filename: "event.ics",
                Content:  icsData,
                MimeType: "text/calendar",
            },
        },
    })
}
```

---

## Tasks

### Phase 1: Email Template (TDD)
- [ ] Create HTML email template
- [ ] Create plain text email template
- [ ] Add RSVP summary section
- [ ] Add event details section
- [ ] Add preference answers section
- [ ] Add update link
- [ ] Test template rendering

### Phase 2: Email Service Integration
- [ ] Implement sendConfirmationEmail method
- [ ] Generate ICS attachment
- [ ] Format email data
- [ ] Queue email for sending
- [ ] Write tests for email generation
- [ ] Write tests for ICS attachment

### Phase 3: Async Sending
- [ ] Implement async email sending
- [ ] Add error logging
- [ ] Add retry logic
- [ ] Test email queue integration
- [ ] Test failure scenarios

### Phase 4: Integration Testing
- [ ] Test full RSVP + email flow
- [ ] Test email content accuracy
- [ ] Test ICS attachment validity
- [ ] Test update link functionality
- [ ] Test retry on failure

---

## Email Template Structure

### HTML Template

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP Confirmation</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
        <h1 style="color: #28a745; margin: 0;">✓ RSVP Confirmed</h1>
    </div>
    
    <p>Hi {{.GuestName}},</p>
    
    <p>Thank you for responding to the invitation!</p>
    
    <div style="background: #fff; border: 1px solid #ddd; padding: 15px; border-radius: 4px; margin: 20px 0;">
        <h2 style="margin-top: 0;">Your Response</h2>
        <p><strong>Status:</strong> {{.Response | title}}</p>
        {{if gt .PlusOnes 0}}
        <p><strong>Guests:</strong> {{.PlusOnes}}</p>
        {{end}}
    </div>
    
    <div style="background: #fff; border: 1px solid #ddd; padding: 15px; border-radius: 4px; margin: 20px 0;">
        <h2 style="margin-top: 0;">Event Details</h2>
        <p><strong>{{.EventTitle}}</strong></p>
        <p>📅 {{.EventDate}}</p>
        <p>📍 {{.EventLocation}}</p>
    </div>
    
    {{if .Answers}}
    <div style="background: #fff; border: 1px solid #ddd; padding: 15px; border-radius: 4px; margin: 20px 0;">
        <h2 style="margin-top: 0;">Your Preferences</h2>
        {{range .Answers}}
        <p><strong>{{.Question}}:</strong> {{.Answer}}</p>
        {{end}}
    </div>
    {{end}}
    
    <p style="margin-top: 30px;">
        <a href="{{.UpdateURL}}" style="display: inline-block; background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">
            Update RSVP
        </a>
    </p>
    
    <p style="color: #666; font-size: 14px; margin-top: 30px;">
        A calendar file is attached to this email. Add it to your calendar so you don't forget!
    </p>
</body>
</html>
```

### Plain Text Template

```
RSVP CONFIRMED
==============

Hi {{.GuestName}},

Thank you for responding to the invitation!

YOUR RESPONSE
-------------
Status: {{.Response | title}}
{{if gt .PlusOnes 0}}Guests: {{.PlusOnes}}{{end}}

EVENT DETAILS
-------------
{{.EventTitle}}
Date: {{.EventDate}}
Location: {{.EventLocation}}

{{if .Answers}}
YOUR PREFERENCES
----------------
{{range .Answers}}
{{.Question}}: {{.Answer}}
{{end}}
{{end}}

Update your RSVP: {{.UpdateURL}}

A calendar file is attached to this email.
```

---

## Tasks

- [ ] Create HTML email template
- [ ] Create plain text email template
- [ ] Implement email sending logic
- [ ] Generate ICS attachment
- [ ] Queue email asynchronously
- [ ] Write tests for email generation
- [ ] Write tests for template rendering
- [ ] Integration test email sending

---

## Dependencies

**Depends on:**
- Story 02: RSVP Submission
- Story 10: Confirmation Page
- Epic 05: Email System (email queue)

**Blocks:**
- None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Email templates created
- [ ] Email sending implemented
- [ ] ICS attachment working
- [ ] Async sending working
- [ ] Tests passing
- [ ] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
- **Email:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
