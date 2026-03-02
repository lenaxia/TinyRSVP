# User Story: Token Regeneration

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-08

---

## User Story

As an **event manager**, I want **to regenerate invite tokens** so that **I can issue a new link if the original was compromised or lost**.

---

## Acceptance Criteria

- [x] Event manager can regenerate token for any invite
- [x] Old token immediately invalidated
- [x] New token generated securely
- [x] New token returned once to caller
- [x] Invite status preserved (draft/sent/viewed)
- [x] Cannot regenerate revoked invites
- [x] Cannot regenerate responded invites
- [x] Permission check: only event creator/managers
- [x] HTTP API endpoint for regeneration

---

## Technical Details

### Service Interface

```go
type RegenerateTokenResponse struct {
    Token   string
    RSVPURL string
}

func (s *service) RegenerateToken(ctx context.Context, inviteID int64) (*RegenerateTokenResponse, error)
```

### HTTP Endpoint

```
POST /api/invites/:inviteId/regenerate
Content-Type: application/json

Response 200 OK:
{
    "token": "b4G9lM0nO3qR6sU8wX1yZ5aB7cD9eF2gH4jK6m",
    "rsvp_url": "https://rsvp.example.com/rsvp/b4G9lM0nO3qR6sU8wX1yZ5aB7cD9eF2gH4jK6m"
}
```

---

## Subtasks

### Implementation
- [x] Add `RegenerateToken()` to service
- [x] Validate invite can be regenerated
- [x] Generate new secure token
- [x] Hash new token
- [x] Update invite with new token hash
- [x] Return new token
- [x] Add HTTP handler endpoint

### Testing
- [x] Test successful regeneration
- [x] Test old token invalidated
- [x] Test new token works
- [x] Test cannot regenerate revoked
- [x] Test cannot regenerate responded
- [x] Test permission checks

### Documentation
- [x] API documentation
- [x] Use cases
- [x] Security considerations

---

## Use Cases

1. **Token Compromised**: Guest accidentally shared link publicly
2. **Token Lost**: Guest deleted email, needs new link
3. **Wrong Channel**: Sent via wrong medium, need to resend
4. **Testing**: Generate new token for testing purposes

---

## Security Considerations

- Old token immediately invalid (no grace period)
- New token completely independent
- Regeneration logged for audit
- Rate limiting to prevent abuse

---

## References

- **Story 00:** [03_STORY_00_token_generation.md](03_STORY_00_token_generation.md)
- **Story 08:** [03_STORY_08_token_revocation.md](03_STORY_08_token_revocation.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Regeneration logic implemented
- [x] Tests passing (>90% coverage)
- [x] Documentation complete
- [x] Code reviewed
