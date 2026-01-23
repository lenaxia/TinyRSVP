# User Story: Migrate Handler Tests

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 4-5 hours
**Phase:** 3 - Migration with Validation

---

## User Story

As a **developer**, I want **all handler tests migrated to use generated mocks** so that **handler testing is consistent and maintainable across the codebase**.

---

## Acceptance Criteria

- [ ] All 81 handler test files migrated
- [ ] Manual mocks removed from handler tests
- [ ] All tests pass
- [ ] Committed in batches (5-10 files at a time)
- [ ] Code is more readable

---

## Scope

**Target Files:** 81 handler test files
- Event handlers (~15 files)
- Invite handlers (~20 files)
- RSVP handlers (~5 files)
- Admin handlers (~5 files)
- Template handlers (~10 files)
- Middleware tests (~10 files)
- Other handlers (~16 files)

---

## Migration Strategy

1. Work in categories (events, invites, etc.)
2. Batch 5-10 files at a time
3. Run tests after each batch
4. Commit after each successful batch
5. Use patterns from validation phase

---

## Dependencies

**Depends on:** Story 10 (reflect/adjust - must be PROCEED)  
**Blocks:** Story 14 (cleanup)

---

## Progress Tracking

Track migration in batches:
- [ ] Batch 1: Event handlers (files 1-10)
- [ ] Batch 2: Event handlers (files 11-15)
- [ ] Batch 3: Invite handlers (files 1-10)
- [ ] Batch 4: Invite handlers (files 11-20)
- [ ] Batch 5: RSVP + Admin handlers
- [ ] Batch 6: Template handlers
- [ ] Batch 7: Middleware tests
- [ ] Batch 8: Remaining handlers
