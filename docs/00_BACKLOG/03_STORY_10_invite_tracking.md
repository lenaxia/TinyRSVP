# User Story: Invite Status Tracking

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As an **event manager**, I want **to track invite status transitions** so that **I know which guests have viewed and responded to invites**.

---

## Acceptance Criteria

- [ ] Invite status tracked through lifecycle
- [ ] Status transitions: draft → sent → viewed → responded
- [ ] `sent_at` timestamp recorded when email sent
- [ ] `viewed_at` timestamp recorded when RSVP page accessed
- [ ] Status automatically updated on transitions
- [ ] Invalid transitions prevented
- [ ] Status visible in invite list
- [ ] Statistics aggregated by status

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
- [ ] Implement status transition methods
- [ ] Add validation for valid transitions
- [ ] Update timestamps on transitions
- [ ] Integrate with email service (mark sent)
- [ ] Integrate with RSVP handler (mark viewed)
- [ ] Integrate with RSVP submission (mark responded)

### Testing
- [ ] Test valid transitions
- [ ] Test invalid transitions rejected
- [ ] Test timestamp updates
- [ ] Test idempotency (multiple views)
- [ ] Test statistics accuracy

### Documentation
- [ ] Document status lifecycle
- [ ] Document transition rules
- [ ] Document integration points

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

- [ ] All acceptance criteria met
- [ ] Status tracking implemented
- [ ] Transitions validated
- [ ] Tests passing (>90% coverage)
- [ ] Integration complete
- [ ] Documentation complete
