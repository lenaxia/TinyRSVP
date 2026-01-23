# User Story: Migrate Repository Tests

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 1 hour
**Phase:** 3 - Migration with Validation

---

## User Story

As a **developer**, I want **repository tests updated to use generated mocks where applicable** so that **all mock usage is consistent**.

---

## Acceptance Criteria

- [ ] Tests mocking db.Database updated
- [ ] Error scenario tests use generated mocks
- [ ] All tests pass

---

## Scope

Most repository tests use real in-memory SQLite. Only update tests that currently mock db.Database for error scenarios.

**Files to Update:**
- Tests with mocked Database for error handling
- Tests needing transaction mocking

---

## Pattern

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockDB := mocks.NewMockDatabase(ctrl)
mockDB.EXPECT().
    Exec(gomock.Any(), gomock.Any(), gomock.Any()).
    Return(nil, errors.New("database error"))
```

---

## Dependencies

**Depends on:** Story 10

---

## Note

This is the smallest migration story since most repository tests correctly use real databases.
