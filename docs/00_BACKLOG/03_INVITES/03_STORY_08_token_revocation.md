# User Story: Token Revocation

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-08

---

## User Story

As an **event manager**, I want **to revoke invite tokens** so that **I can cancel specific invites if needed (wrong person, security concern, etc.)**.

---

## Acceptance Criteria

- [x] Event manager can revoke any invite for their event
- [x] Revoked tokens cannot be used for RSVP
- [x] Revocation is permanent (cannot be undone)
- [x] Invite status changes to 'revoked'
- [x] Revoked invites still visible in invite list
- [x] Revocation reason optional but recommended
- [x] Permission check: only event creator/managers
- [x] HTTP API endpoint for revocation

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
- [x] Add `RevokeInvite()` to service
- [x] Update invite status to 'revoked'
- [x] Store revocation reason (audit log)
- [x] Check permissions
- [x] Prevent revocation of already-responded invites
- [x] Add HTTP handler endpoint

### Testing
- [x] Test successful revocation
- [x] Test permission checks
- [x] Test revoked token rejection
- [x] Test cannot revoke responded invite
- [x] Test revocation reason storage

### Documentation
- [x] API documentation
- [x] Revocation policy
- [x] Use cases

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

- [x] All acceptance criteria met
- [x] Revocation logic implemented
- [x] Tests passing (>90% coverage)
- [x] Documentation complete
- [x] Code reviewed
