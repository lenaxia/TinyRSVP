# User Story: Token Revocation

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **to revoke invite tokens** so that **I can cancel specific invites if needed (wrong person, security concern, etc.)**.

---

## Acceptance Criteria

- [ ] Event manager can revoke any invite for their event
- [ ] Revoked tokens cannot be used for RSVP
- [ ] Revocation is permanent (cannot be undone)
- [ ] Invite status changes to 'revoked'
- [ ] Revoked invites still visible in invite list
- [ ] Revocation reason optional but recommended
- [ ] Permission check: only event creator/managers
- [ ] HTTP API endpoint for revocation

---

## Technical Details

### Service Interface

```go
type RevokeInviteRequest struct {
    InviteID int64
    Reason   *string
}

func (s *service) RevokeInvite(ctx context.Context, req *RevokeInviteRequest) error
```

### HTTP Endpoint

```
POST /api/invites/:inviteId/revoke
Content-Type: application/json

{
    "reason": "Wrong email address"
}

Response 200 OK:
{
    "message": "Invite revoked successfully"
}
```

---

## Subtasks

### Implementation
- [ ] Add `RevokeInvite()` to service
- [ ] Update invite status to 'revoked'
- [ ] Store revocation reason (audit log)
- [ ] Check permissions
- [ ] Prevent revocation of already-responded invites
- [ ] Add HTTP handler endpoint

### Testing
- [ ] Test successful revocation
- [ ] Test permission checks
- [ ] Test revoked token rejection
- [ ] Test cannot revoke responded invite
- [ ] Test revocation reason storage

### Documentation
- [ ] API documentation
- [ ] Revocation policy
- [ ] Use cases

---

## Status Transitions

```
draft → revoked ✓
sent → revoked ✓
viewed → revoked ✓
responded → revoked ✗ (cannot revoke)
```

---

## References

- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md)
- **Story 03:** [03_STORY_03_invite_model.md](03_STORY_03_invite_model.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Revocation logic implemented
- [ ] Tests passing (>90% coverage)
- [ ] Documentation complete
- [ ] Code reviewed
