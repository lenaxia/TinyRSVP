# User Story: Create Testing Documentation

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 hour
**Phase:** 4 - Cleanup & Documentation
**Completed:** 2026-07-07

---

## User Story

As a **developer**, I want **comprehensive testing documentation** so that **I understand when to use mocks vs real implementations and how to write effective tests**.

---

## Acceptance Criteria

- [x] `docs/TESTING.md` created
- [x] Testing philosophy documented
- [x] Test categories explained (unit/integration/e2e)
- [x] When to mock guidelines
- [x] gomock usage examples
- [x] Common testing patterns documented

---

## Document Outline

```markdown
# TinyRSVP Testing Guide

## Testing Philosophy
- Test-Driven Development
- Fast, isolated, deterministic tests
- Clear assertions

## Test Categories
### Unit Tests
- Purpose, speed, dependencies
- When to use
- Example

### Integration Tests
- Purpose, speed, dependencies
- When to use
- Example

### E2E Tests
- Purpose, speed, dependencies
- When to use
- Example

## When to Mock
- ALWAYS mock (external services, slow operations)
- SOMETIMES mock (databases, repositories, services)
- NEVER mock (validators, pure functions, code under test)

## Using Generated Mocks
### Basic Setup
- Creating controller
- Creating mocks
- Setting expectations

### Argument Matchers
- gomock.Any()
- gomock.Eq()
- Custom matchers

### Multiple Calls
- Times()
- MinTimes()
- AnyTimes()

### Return Values
- Return()
- Do()
- DoAndReturn()

### Call Order
- gomock.InOrder()

## Common Patterns
- Handler testing
- Service testing
- Repository testing
- Error scenario testing

## Test Data
- Using testutil helpers
- Creating test users/events
- Using builders (Phase 5)

## Running Tests
- All tests
- Specific package
- Specific test
- With coverage

## Troubleshooting
- Common errors
- Mock expectation failures
- Test flakiness
```

---

## Dependencies

**Depends on:** Story 14 (cleanup complete)

---

## Validation

- [x] Document reviewed by team
- [x] Examples tested and work
- [x] Links to code samples verified

---

## Implementation Notes (2026-07-07)

`docs/TESTING.md` was already in place and comprehensive. This pass:

- Verified the mock-package contents table against the actual generated mocks in `internal/testutil/mocks/{services,repositories,other}/` and completed the `other` list (added `MockProvider`, `MockTemplateValidator`, `MockJobsEventService`).
- Added a **UX Tests (Browser)** subsection covering `tests/ux/` (chromedp, in-process `httptest.NewServer`, `X-Test-User-ID` auth bypass, `SeedDefaults`).
- Expanded the **Running Tests** section with the recommended non-UX fast-feedback invocation, the UX-only invocation, race-detector guidance, and a note that `-timeout` is mandatory (enforced by the pre-commit hook).

All acceptance criteria met.
