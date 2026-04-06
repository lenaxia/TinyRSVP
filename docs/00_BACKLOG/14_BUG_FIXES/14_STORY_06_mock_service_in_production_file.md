# STORY: Move MockService Out of Production Package

**Epic:** 14 - Bug Fixes & Code Gaps  
**Story ID:** 14_STORY_06  
**Priority:** Low  
**Estimated Effort:** 30 minutes  
**Severity:** Low — no runtime impact; test infrastructure code living in a production package

---

## Problem

`internal/email/service.go:13-23` defines a `MockService` struct:

```go
// MockService is a mock implementation of Service for testing
type MockService struct {
    SendConfirmationEmailFunc  func(...) error
    SendConfirmationEmailCalls int
}
```

This is test infrastructure in a non-test file. It is compiled into the production binary, visible in the package's exported API, and creates confusion about what is production code vs. test code. The project already has a generated mock at `internal/testutil/mocks/services/mock_email_service.go` that provides the same capability via gomock — `MockService` here is redundant.

---

## Acceptance Criteria

- [ ] `MockService` struct removed from `internal/email/service.go`
- [ ] All usages of `email.MockService` in test files updated to use `testutil/mocks/services.MockEmailService` (the generated mock)
- [ ] No production file imports or references `MockService`
- [ ] All 32 non-browser packages pass
- [ ] Update `docs/00_BACKLOG/14_BUG_FIXES/README.md`: mark this story complete

---

## Technical Approach

1. Search for all references to `email.MockService` or `&email.MockService{}`
2. Replace with `mocksvcs.MockEmailService` from `internal/testutil/mocks/services`
3. Delete the `MockService` struct and `SendConfirmationEmail` method from `internal/email/service.go`
4. Run tests

```bash
grep -rn "email\.MockService\|&MockService{" internal/ --include="*.go"
```

---

## Files to Change

- `internal/email/service.go` — remove `MockService` struct
- Any test file that directly references `email.MockService` — update to use generated mock

---

## Testing

```bash
go build ./...
go test -timeout 30s ./internal/email/...
go test -timeout 30s ./...
```

---

## Status

- **Status:** Not Started
