# Epic: Email System & Calendar Integration

**Priority:** High  
**Status:** Not Started  
**Target Version:** v0  
**Estimated Effort:** 1.5 weeks

---

## Overview

Implement reliable email delivery system with queue management, retry logic, and ICS calendar file generation. Support invite emails, confirmation emails, event updates, and cancellation notifications.

**Goal:** Enable automated email delivery with calendar attachments, ensuring reliable delivery through retry mechanisms and proper SMTP integration.

---

## Success Criteria

- [ ] SMTP configuration validated on startup
- [ ] Email queue processes messages reliably
- [ ] Hybrid send strategy (immediate + background retry)
- [ ] Retry policy with exponential backoff (4 attempts)
- [ ] Rate limiting enforced (50/minute configurable)
- [ ] ICS calendar files generated correctly (RFC 5545)
- [ ] Bounce handling for failed deliveries
- [ ] Email templates support HTML and plain text
- [ ] Unsubscribe mechanism functional

---

## User Stories

### Phase 1: SMTP Integration
- [ ] [`05_STORY_00_smtp_configuration.md`](05_STORY_smtp_configuration.md) - SMTP config and validation
- [ ] [`05_STORY_01_smtp_connection.md`](05_STORY_smtp_connection.md) - SMTP connection management
- [ ] [`05_STORY_02_email_sending.md`](05_STORY_email_sending.md) - Basic email sending

### Phase 2: Email Queue
- [ ] [`05_STORY_03_email_queue_model.md`](05_STORY_email_queue_model.md) - Queue table and repository
- [ ] [`05_STORY_04_queue_processor.md`](05_STORY_queue_processor.md) - Background queue processor
- [ ] [`05_STORY_05_retry_policy.md`](05_STORY_retry_policy.md) - Exponential backoff retry

### Phase 3: Email Types
- [ ] [`05_STORY_06_invite_email.md`](05_STORY_invite_email.md) - Invitation email template
- [ ] [`05_STORY_07_confirmation_email.md`](05_STORY_confirmation_email.md) - RSVP confirmation email
- [ ] [`05_STORY_08_update_email.md`](05_STORY_update_email.md) - Event update notification
- [ ] [`05_STORY_09_cancellation_email.md`](05_STORY_cancellation_email.md) - Event cancellation email

### Phase 4: Calendar Integration
- [ ] [`05_STORY_10_ics_generation.md`](05_STORY_ics_generation.md) - ICS file generation (RFC 5545)
- [ ] [`05_STORY_11_ics_updates.md`](05_STORY_ics_updates.md) - ICS updates with SEQUENCE
- [ ] [`05_STORY_12_ics_cancellation.md`](05_STORY_ics_cancellation.md) - ICS cancellation

### Phase 5: Reliability
- [ ] [`05_STORY_13_rate_limiting.md`](05_STORY_rate_limiting.md) - Email rate limiting
- [ ] [`05_STORY_14_bounce_handling.md`](05_STORY_bounce_handling.md) - Bounce detection
- [ ] [`05_STORY_15_unsubscribe.md`](05_STORY_unsubscribe.md) - Unsubscribe mechanism

---

## Dependencies

**Depends on:** Epic 00 (Foundation), Epic 02 (Events), Epic 03 (Invites), Epic 06 (Templates)  
**Blocks:** None (supporting system)

---

## Technical Overview

### Hybrid Send Strategy

```
Email queued
     ↓
Immediate send attempt
     ↓
Success? → Mark sent, done
     ↓ Failure
Schedule retry
     ↓
Background worker picks up
     ↓
Retry with backoff
```

### Retry Policy

```
Attempt 1: Immediate
Attempt 2: +1 minute
Attempt 3: +5 minutes
Attempt 4: +15 minutes
After 4 attempts: Mark failed, notify admin
```

### Email Queue States

```
pending  → Waiting to be sent
sending  → Currently being sent
sent     → Successfully delivered
failed   → Permanently failed
cancelled → Manually cancelled
```

### ICS File Structure

```
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//TinyRSVP//EN
METHOD:REQUEST
BEGIN:VEVENT
UID:{event.id}@{domain}
DTSTAMP:{generation_time}
DTSTART;TZID={timezone}:{start_time}
DTEND;TZID={timezone}:{end_time}
SUMMARY:{title}
DESCRIPTION:{description}
LOCATION:{location}
SEQUENCE:{version}
END:VEVENT
END:VCALENDAR
```

---

## Technical Decisions

### Queue Storage: Database
- Survives application restarts
- Queryable for monitoring
- Supports atomic updates
- Enables retry scheduling

### Hybrid Send Strategy
- Immediate attempt for fast delivery
- Background retry for reliability
- Best of both approaches
- Graceful degradation

### Rate Limiting
- Prevents SMTP throttling
- Configurable limit (default 50/min)
- Rolling window tracking
- Delays batches if limit reached

### ICS Sequence Numbers
- Starts at 0
- Increments on each update
- Tells calendar clients to update
- Same UID = same event

---

## Email Templates

### Invite Email
```
Subject: You're invited: {event.title}
Body: Event details + RSVP link
Attachment: event.ics
```

### Confirmation Email
```
Subject: RSVP Confirmed: {event.title}
Body: RSVP summary + update link
Attachment: event.ics
```

### Update Email
```
Subject: Event Updated: {event.title}
Body: Changes summary + event details
Attachment: updated.ics (SEQUENCE++)
```

### Cancellation Email
```
Subject: Event Cancelled: {event.title}
Body: Cancellation reason
Attachment: cancelled.ics (STATUS:CANCELLED)
```

---

## SMTP Configuration

### Required Settings
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=user@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM=noreply@example.com
SMTP_FROM_NAME=TinyRSVP
```

### Optional Settings
```
SMTP_TLS=true
SMTP_SKIP_VERIFY=false
SMTP_TIMEOUT=30
EMAIL_RATE_LIMIT=50
```

### Validation
- Test connection on startup
- Fail fast if invalid
- "Test Email" button in admin UI

---

## Bounce Handling

### Bounce Types
- **Hard bounce** (5xx): Mailbox doesn't exist → mark email_invalid
- **Soft bounce** (4xx): Temporary issue → retry
- **Block**: Spam filter → mark failed, notify admin

### Admin Notification
- Daily digest of failed emails
- Includes: guest name, email, error
- Admin can manually resend or update email

---

## Rate Limiting

### Implementation
```
Track sends in rolling 60-second window
If count >= limit:
  - Delay next batch
  - Wait until window resets
  - Resume sending
```

### Monitoring
- Expose metrics: emails_sent_per_minute
- Alert if consistently at limit
- Suggest increasing SMTP limit

---

## References

- **HLD:** Section 9 (Email System), Section 10 (Calendar Integration)
- **LLD:** [`lld/05_EMAIL_LLD.md`](../lld/05_EMAIL_LLD.md)
- **Database:** email_queue table
- **Standards:** RFC 5545 (iCalendar), RFC 5322 (Email), CAN-SPAM Act

---

## Testing Strategy

### Unit Tests
- ICS generation
- Retry policy logic
- Rate limiting
- Bounce classification
- Template rendering

### Integration Tests
- Full email send flow
- Queue processing
- SMTP connection handling
- Retry mechanism
- ICS import into calendar apps

### Manual Tests
1. Configure SMTP (Gmail, SendGrid, etc.)
2. Send test invite
3. Verify email received
4. Import ICS into calendar
5. Update event, verify calendar updates
6. Cancel event, verify calendar removes

### Edge Cases
- SMTP server down
- Authentication failure
- Recipient rejected
- Rate limit exceeded
- Malformed ICS
- Missing timezone

---

## Compliance

### CAN-SPAM Act
- Sender identification in all emails
- Physical address in footer
- Unsubscribe link in reminders
- Process unsubscribe within 10 days

### Unsubscribe
- Link: `/unsubscribe/{token}`
- Sets invite.unsubscribed = true
- Stops reminder emails only
- Does not affect RSVP ability

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| SMTP provider throttling | High | Rate limiting, retry policy |
| Email deliverability | High | SPF/DKIM/DMARC setup guide |
| Queue processing failure | Medium | Idempotent operations, monitoring |
| ICS compatibility issues | Medium | Test with major calendar apps |
| Bounce loop | Low | Max retry attempts, admin notification |

---

## Performance Targets

- Email queued: <100ms
- Immediate send: <5 seconds
- Queue processing: 60 second intervals
- Rate limit: 50 emails/minute
- ICS generation: <50ms

---

## Monitoring

### Key Metrics
- `email_queue_size` - Pending emails
- `email_send_rate` - Emails per minute
- `email_failures` - Failed sends
- `email_retry_count` - Retry attempts
- `ics_generation_time` - ICS generation latency

### Alerts
- Queue size > 1000
- Failure rate > 10%
- SMTP connection failures
- Rate limit consistently hit

---

## Definition of Done

- [ ] All user stories complete
- [ ] SMTP integration working
- [ ] Email queue processing reliably
- [ ] Retry policy functional
- [ ] Rate limiting enforced
- [ ] ICS files generated correctly
- [ ] All email types working
- [ ] Bounce handling implemented
- [ ] Unsubscribe functional
- [ ] All tests passing
- [ ] Calendar integration verified
- [ ] Documentation updated
