# User Story: Token Regeneration

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **to regenerate invite tokens** so that **I can issue a new link if the original was compromised or lost**.

---

## Acceptance Criteria

- [ ] Event manager can regenerate token for any invite
- [ ] Old token immediately invalidated
- [ ] New token generated securely
- [ ] New token returned once to caller
- [ ] Invite status preserved (draft/sent/viewed)
- [ ] Cannot regenerate revoked invites
- [ ] Cannot regenerate responded invites
- [ ] Permission check: only event creator/managers
- [ ] HTTP API endpoint for regeneration

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
- [ ] Add `RegenerateToken()` to service
- [ ] Validate invite can be regenerated
- [ ] Generate new secure token
- [ ] Hash new token
- [ ] Update invite with new token hash
- [ ] Return new token
- [ ] Add HTTP handler endpoint

### Testing
- [ ] Test successful regeneration
- [ ] Test old token invalidated
- [ ] Test new token works
- [ ] Test cannot regenerate revoked
- [ ] Test cannot regenerate responded
- [ ] Test permission checks

### Documentation
- [ ] API documentation
- [ ] Use cases
- [ ] Security considerations

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

- [ ] All acceptance criteria met
- [ ] Regeneration logic implemented
- [ ] Tests passing (>90% coverage)
- [ ] Documentation complete
- [ ] Code reviewed
