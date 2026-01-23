# User Story: Migrate Service Tests

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 2-3 hours
**Phase:** 3 - Migration with Validation

---

## User Story

As a **developer**, I want **all service tests migrated to use generated mocks** so that **service layer testing is consistent**.

---

## Acceptance Criteria

- [ ] All ~30 service test files migrated
- [ ] Manual mocks removed
- [ ] All tests pass
- [ ] Integration tests preserve real DB pattern

---

## Scope

**Core Services:**
- events.Service tests
- invites.InviteService tests
- rsvp.Service tests
- templates.Service tests

**Supporting Services:**
- email.* tests
- admin.Service tests
- assets.* tests

**Utility Services:**
- Various validation and helper tests

---

## Special Considerations

**Integration Tests:**
- Keep real DB (SetupTestDB)
- Only mock external dependencies
- Use testutil.SetupTestDBWithMigrations

**Unit Tests:**
- Mock everything except code under test

---

## Dependencies

**Depends on:** Story 10  
**Blocks:** Story 14

---

## Validation

```bash
# After each batch
go test ./internal/events -v
go test ./internal/invites -v
go test ./internal/rsvp -v
```
