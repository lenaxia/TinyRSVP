# User Story: Generate Service Mocks

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 30 minutes
**Phase:** 2 - Mock Generation Setup

---

## User Story

As a **developer**, I want **generated mocks for service interfaces** so that **I can test handlers without implementing manual service mocks**.

---

## Acceptance Criteria

- [ ] Mock for events.Service generated
- [ ] Mock for invites.InviteService generated
- [ ] Mock for rsvp.Service generated
- [ ] Mock for templates.Service generated
- [ ] Mock for email.Service generated
- [ ] All mocks compile and can be used

---

## Services to Mock

1. events.Service (8 methods)
2. invites.InviteService (16 methods)
3. rsvp.Service (2 methods)
4. templates.Service (11 methods)
5. email.Service (1 method - replaces manual mock)

---

## Dependencies

**Depends on:** Story 05  
**Blocks:** Story 09

---

## Validation

```bash
./scripts/generate_mocks.sh
go build ./internal/testutil/mocks/...
```
