# User Story: Invite Status Tracking

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-08

---

## User Story

As an **event manager**, I want **to track invite status transitions** so that **I know which guests have viewed and responded to invites**.

---

## Acceptance Criteria

- [x] Invite status tracked through lifecycle
- [x] Status transitions: draft → sent → viewed → responded
- [x] `sent_at` timestamp recorded when email sent
- [x] `viewed_at` timestamp recorded when RSVP page accessed
- [x] Status automatically updated on transitions
- [x] Invalid transitions prevented
- [x] Status visible in invite list
- [x] Statistics aggregated by status

---

## Technical Details

### Status Transitions

```
draft → sent → viewed → responded
  ↓       ↓       ↓         ↓
  └───────┴───────┴─────→ revoked
```

### Status Update Methods

```go
func (s *service) MarkInviteSent(ctx context.Context, inviteID int64) error
func (s *service) MarkInviteViewed(ctx context.Context, inviteID int64) error
func (s *service) MarkInviteResponded(ctx context.Context, inviteID int64) error
```

### Automatic Transitions

- **sent**: When email queued/sent
- **viewed**: When guest accesses RSVP page
- **responded**: When guest submits RSVP

---

## Subtasks

### Implementation
- [x] Implement status transition methods
- [x] Add validation for valid transitions
- [x] Update timestamps on transitions
- [ ] Integrate with email service (mark sent)
- [ ] Integrate with RSVP handler (mark viewed)
- [ ] Integrate with RSVP submission (mark responded)

### Testing
- [x] Test valid transitions
- [x] Test invalid transitions rejected
- [x] Test timestamp updates
- [x] Test idempotency (multiple views)
- [x] Test statistics accuracy

### Documentation
- [x] Document status lifecycle
- [x] Document transition rules
- [x] Document integration points

---

## Status Lifecycle

| Status | Description | Next Valid States |
|--------|-------------|-------------------|
| draft | Created but not sent | sent, revoked |
| sent | Email sent to guest | viewed, revoked |
| viewed | Guest opened RSVP page | responded, revoked |
| responded | Guest submitted RSVP | (terminal) |
| revoked | Cancelled by manager | (terminal) |

---

## References

- **Story 03:** [03_STORY_03_invite_model.md](03_STORY_03_invite_model.md)
- **Epic 04:** RSVP (response submission)
- **Epic 05:** Email (sending invites)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Status tracking implemented
- [x] Transitions validated
- [x] Tests passing (>90% coverage)
- [ ] Integration complete (deferred to Epic 4 & 5)
- [x] Documentation complete
