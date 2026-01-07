# User Story: Manual Invite Generation

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **to generate invite tokens without email addresses** so that **I can manually distribute invite links via other channels (SMS, messaging apps, printed cards)**.

---

## Acceptance Criteria

- [x] Event manager can create invite without email
- [x] Guest name is optional
- [x] Token generated and displayed immediately
- [x] RSVP URL provided for copying
- [x] Invite created in 'draft' status
- [x] Max plus ones configurable
- [x] Multiple manual invites can be created
- [x] Token never shown again after initial display
- [x] Permission check: only event creator/managers

---

## Technical Details

### Service Interface

```go
type CreateManualInviteRequest struct {
    EventID     int64
    Name        *string
    MaxPlusOnes *int
}

type CreateManualInviteResponse struct {
    Invite  *models.Invite
    Token   string
    RSVPURL string
}

func (s *service) CreateManualInvite(ctx context.Context, req *CreateManualInviteRequest) (*CreateManualInviteResponse, error)
```

### HTTP Endpoint

```
POST /api/events/:eventId/invites/manual
Content-Type: application/json

{
    "name": "John Doe",
    "max_plus_ones": 2
}

Response 201 Created:
{
    "invite": {
        "id": 123,
        "event_id": 1,
        "name": "John Doe",
        "max_plus_ones": 2,
        "status": "draft",
        "expires_at": "2026-02-15T00:00:00Z"
    },
    "token": "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p",
    "rsvp_url": "https://rsvp.example.com/rsvp/a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p"
}
```

---

## Subtasks

### Implementation
- [x] Add `CreateManualInvite()` to service
- [x] Validate event exists and not cancelled
- [x] Check user permissions
- [x] Generate secure token
- [x] Create invite without email
- [x] Return token and RSVP URL
- [x] Add HTTP handler endpoint
- [ ] Add UI for manual invite creation

### Testing
- [x] Test successful manual invite creation
- [x] Test permission checks
- [x] Test token generation
- [x] Test RSVP URL format
- [x] Test multiple manual invites
- [x] Integration test

### Documentation
- [x] API documentation
- [ ] Use case examples
- [ ] UI instructions

---

## Use Cases

1. **SMS Distribution**: Generate token, send via SMS
2. **Messaging Apps**: Share RSVP link in WhatsApp/Telegram
3. **Printed Cards**: Print QR code with RSVP URL
4. **In-Person**: Show QR code on phone for scanning
5. **Social Media**: Share link in private messages

---

## UI Considerations

Display token prominently with:
- Copy button for token
- Copy button for full RSVP URL
- QR code generation option
- Warning: "Save this link - it won't be shown again"
- Option to print or download

---

## References

- **Story 04:** [03_STORY_04_individual_invite.md](03_STORY_04_individual_invite.md)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Service method implemented
- [x] HTTP endpoint implemented
- [x] Tests passing (>90% coverage)
- [x] Documentation complete
- [x] Code reviewed
