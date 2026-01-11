# Story 10.16: Auth Test Compilation Fix

**Epic:** 10 - Technical Debt & Improvements  
**Priority:** Medium  
**Status:** Not Started  
**Identified:** 2026-01-11 (Epic 11 Phase 1 Validation)

---

## User Story

As a **developer**, I want **auth tests to compile successfully** so that **I can run the complete test suite**.

---

## Problem Statement

During Epic 11 Phase 1 validation, auth tests fail to compile due to missing imports or incorrect package references in `internal/auth/login_redirect_test.go`.

---

## Compilation Errors

```
internal/auth/login_redirect_test.go:107:74: undefined: models
internal/auth/login_redirect_test.go:108:14: undefined: models
internal/auth/login_redirect_test.go:112:14: undefined: models
internal/auth/login_redirect_test.go:115:26: cannot use func(userID int64) error {…} 
    (value of type func(userID int64) error) as 
    func(ctx context.Context, userID int64) error value in struct literal
internal/auth/login_redirect_test.go:121:62: undefined: models
internal/auth/login_redirect_test.go:122:14: undefined: models
```

---

## Acceptance Criteria

- [ ] Auth tests compile successfully
- [ ] No undefined references to `models`
- [ ] Function signatures match expected types
- [ ] All auth tests run and pass
- [ ] No compilation errors in auth package

---

## Technical Approach

### Issue 1: Undefined `models`

**Likely Causes:**
1. Missing import: `"github.com/lenaxia/tinyrsvp/internal/models"`
2. Incorrect package reference

**Fix:**
```go
import (
    "github.com/lenaxia/tinyrsvp/internal/models"
    // ... other imports
)
```

### Issue 2: Function Signature Mismatch

**Error:**
```
cannot use func(userID int64) error as func(ctx context.Context, userID int64) error
```

**Fix:**
Update mock function to include `context.Context` parameter:
```go
UpdateLastLogin: func(ctx context.Context, userID int64) error {
    return nil
},
```

---

## Tasks

- [ ] Open `internal/auth/login_redirect_test.go`
- [ ] Add missing `models` import
- [ ] Update function signatures to include `context.Context`
- [ ] Verify all references to `models` are correct
- [ ] Run auth tests to verify compilation
- [ ] Run auth tests to verify they pass
- [ ] Check for similar issues in other auth test files

---

## Testing Requirements

### Compilation
- [ ] `go build ./internal/auth` succeeds
- [ ] `go test -c ./internal/auth` succeeds

### Test Execution
- [ ] All auth tests compile
- [ ] All auth tests run
- [ ] All auth tests pass (or have documented failures)

---

## Dependencies

**Prerequisites:**
- None (can be done independently)

**Blocks:**
- Clean test suite
- Auth package testing

---

## Notes

- This is a compilation error, not a logic error
- Likely caused by refactoring that updated function signatures
- Tests were not updated to match
- Medium priority since it blocks auth testing

---

## Estimated Effort

**Size:** Small (15-30 minutes)
- Add import statement
- Update function signatures
- Verify compilation
- Run tests

---

**Status:** Ready for Implementation
