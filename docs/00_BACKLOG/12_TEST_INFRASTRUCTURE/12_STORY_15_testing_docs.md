# User Story: Create Testing Documentation

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 hour
**Phase:** 4 - Cleanup & Documentation

---

## User Story

As a **developer**, I want **comprehensive testing documentation** so that **I understand when to use mocks vs real implementations and how to write effective tests**.

---

## Acceptance Criteria

- [ ] `docs/TESTING.md` created
- [ ] Testing philosophy documented
- [ ] Test categories explained (unit/integration/e2e)
- [ ] When to mock guidelines
- [ ] gomock usage examples
- [ ] Common testing patterns documented

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

- [ ] Document reviewed by team
- [ ] Examples tested and work
- [ ] Links to code samples verified
