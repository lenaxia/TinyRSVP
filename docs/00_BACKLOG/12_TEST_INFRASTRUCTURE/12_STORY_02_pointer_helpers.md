# User Story: Centralize Pointer Helpers

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 30 minutes
**Phase:** 1 - Foundation

---

## User Story

As a **developer**, I want **centralized pointer helper functions** so that **I don't need to redefine stringPtr, intPtr, etc. in every test file**.

---

## Acceptance Criteria

- [ ] `internal/testutil/pointers.go` created
- [ ] All 6 pointer helpers implemented
- [ ] Helper functions have clear documentation
- [ ] Test coverage for pointer helpers (basic smoke tests)
- [ ] Package compiles and tests pass
- [ ] Used 198 times across codebase (will migrate in Phase 4)

---

## Technical Details

### Problem Statement

**Current State:**
- `stringPtr()` defined in 10+ test files
- `intPtr()` defined in 8+ test files
- `boolPtr()` defined in 5+ test files
- Total: 18+ duplicate definitions across 35+ files
- 198 usage sites across the codebase

**Impact:**
- Maintenance burden (update 18 places for any change)
- Inconsistent implementations (some have different names)
- Code bloat (~200 lines of duplicate code)

### Implementation

**File:** `internal/testutil/pointers.go`

```go
package testutil

import "time"

// StringPtr returns a pointer to the given string value.
// Useful for populating optional string fields in test data.
//
// Example:
//
//	invite := &models.Invite{
//	    Email: testutil.StringPtr("test@example.com"),
//	}
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int value.
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to the given int64 value.
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to the given time.Time value.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// Float64Ptr returns a pointer to the given float64 value.
func Float64Ptr(f float64) *float64 {
	return &f
}
```

---

## Tasks (TDD Approach)

### Task 1: Write Tests First
- [ ] Create `internal/testutil/pointers_test.go`
- [ ] Write test for StringPtr
- [ ] Write test for IntPtr
- [ ] Write test for Int64Ptr
- [ ] Write test for BoolPtr (with true/false cases)
- [ ] Write test for TimePtr
- [ ] Write test for Float64Ptr
- [ ] Run tests: `go test ./internal/testutil` (should FAIL)

### Task 2: Implement Pointer Helpers
- [ ] Create `internal/testutil/pointers.go`
- [ ] Implement StringPtr with documentation
- [ ] Implement IntPtr with documentation
- [ ] Implement Int64Ptr with documentation
- [ ] Implement BoolPtr with documentation
- [ ] Implement TimePtr with documentation
- [ ] Implement Float64Ptr with documentation
- [ ] Run tests: `go test ./internal/testutil` (should PASS)

### Task 3: Update Documentation
- [ ] Add examples to godoc comments
- [ ] Update `internal/testutil/README.md` with usage examples
- [ ] Verify godoc renders: `go doc testutil.StringPtr`

### Task 4: Validation
- [ ] All tests pass: `go test ./internal/testutil -v`
- [ ] Package compiles: `go build ./internal/testutil`
- [ ] Can be imported in test files
- [ ] Examples work as documented

---

## Dependencies

**Depends on:** Story 01 (testutil package structure)  
**Blocks:** Story 14 (cleanup old pointer helpers)

---

## Usage Examples

**Before (35+ files with duplicates):**
```go
// In internal/handlers/invites_get_test.go:1256
func stringPtr(s string) *string { return &s }

// In internal/rsvp/service_test.go:58
func strPtr(s string) *string { return &s }

// In internal/events/service_test.go:45
func stringPtr(s string) *string { return &s }

// ... 15+ more definitions
```

**After (single definition, used everywhere):**
```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

email := testutil.StringPtr("test@example.com")
maxPlusOnes := testutil.IntPtr(2)
allowMaybe := testutil.BoolPtr(true)
deadline := testutil.TimePtr(time.Now().Add(24 * time.Hour))
```

---

## Migration Note

This story creates the centralized helpers but does NOT migrate existing code. The actual migration of 198 usage sites across 35+ files will happen in Story 14 (Phase 4: Cleanup).

For now, both the old and new helpers will coexist. This allows us to validate the pattern before committing to full migration.

---

## Validation

```bash
# Run tests
go test ./internal/testutil -v

# Check test coverage
go test ./internal/testutil -cover

# Verify godoc
go doc testutil.StringPtr
go doc testutil.IntPtr

# Test import in existing test file
# Add to any test file temporarily:
import "github.com/lenaxia/tinyrsvp/internal/testutil"
ptr := testutil.StringPtr("test")
```

---

## Notes

Keep these functions simple - they're just one-liners. The value is in consolidation, not complexity.

Document clearly with examples since these will be used hundreds of times across the codebase.
