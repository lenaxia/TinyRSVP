# User Story: Remove Manual Mocks and Duplicates

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 hour
**Phase:** 4 - Cleanup & Documentation

---

## User Story

As a **developer**, I want **all manual mock definitions and duplicate helpers removed** so that **the codebase is clean and maintainable**.

---

## Acceptance Criteria

- [ ] All 92 manual mock definitions removed
- [ ] All 18 duplicate pointer helpers removed
- [ ] All duplicate setupTestDB functions removed
- [ ] All tests still pass
- [ ] Code review completed

---

## Cleanup Tasks

### Remove Manual Mocks
- [ ] Search for `type mock.*struct` in test files
- [ ] Delete entire mock type definitions
- [ ] Delete mock method implementations
- [ ] Verify tests still pass

### Remove Duplicate Helpers
- [ ] Search for `func stringPtr` definitions
- [ ] Search for `func intPtr` definitions
- [ ] Search for `func boolPtr` definitions
- [ ] Remove all occurrences (now using testutil)

### Remove Duplicate setupTestDB
- [ ] Search for `func setupTestDB` definitions
- [ ] Remove all occurrences (now using testutil)

---

## Verification Script

```bash
# Ensure no manual mocks remain
rg "type mock.*struct" internal/ --include="*_test.go"
# Should return no results

# Ensure no duplicate pointer helpers
rg "func.*Ptr\(" internal/ --include="*_test.go"
# Should only show testutil imports

# Run all tests
go test ./... -v
```

---

## Dependencies

**Depends on:** Stories 11, 12, 13 (all migrations complete)  
**Blocks:** Story 15 (documentation references clean code)

---

## Expected Impact

**Lines Removed:** ~15,000 lines of duplicate test code
**Files Modified:** 52 test files
**Net Result:** 87% reduction in test infrastructure code
