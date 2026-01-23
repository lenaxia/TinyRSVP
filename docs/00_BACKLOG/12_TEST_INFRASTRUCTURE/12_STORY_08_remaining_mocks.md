# User Story: Generate Remaining Mocks

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 30 minutes
**Phase:** 2 - Mock Generation Setup

---

## User Story

As a **developer**, I want **mocks for all remaining interfaces** so that **the mock generation infrastructure is complete**.

---

## Acceptance Criteria

- [ ] Mocks for all remaining repository interfaces
- [ ] Mocks for token.Generator
- [ ] Mocks for validator interfaces
- [ ] Mock for storage.Provider
- [ ] All mocks compile
- [ ] Script generates all 21+ mocks

---

## Remaining Interfaces

1. repositories.RSVPRepository
2. repositories.TemplateRepository
3. repositories.AnswerRepository
4. repositories.QuestionRepository
5. repositories.ConfigRepository
6. repositories.SessionRepository
7. repositories.EmailQueueRepository
8. token.Generator
9. events.Validator
10. templates.Validator
11. storage.Provider

---

## Dependencies

**Depends on:** Story 05

---

## Validation

```bash
./scripts/generate_mocks.sh
ls -la internal/testutil/mocks/
# Should see 21+ mock files
go build ./internal/testutil/mocks/...
```
