# Story 09: Validation Migration Report (In Progress)

**Status:** 1/3 files migrated  
**Date:** 2026-01-22  
**Phase:** 3 - Migration with Validation

---

## Summary

Successfully migrated 1 of 3 planned validation files. Discovered critical architectural issue with import cycles that blocks same-package migrations.

---

## Migrations Completed

### ✅ File 1: Handler Test (internal/handlers/invites_get_test.go)

**Status:** SUCCESS  
**Before:** 314 lines  
**After:** 244 lines  
**Reduction:** 70 lines (22%)

**Changes:**
- Removed 74 lines of manual mock definitions
- Replaced `mockGetInviteService` with `mocks.NewMockInviteService(ctrl)`
- Replaced `mockGetInviteEventRepo` with `mocks.NewMockEventRepository(ctrl)`
- Replaced local `stringPtr()` with `testutil.StringPtr()`
- Added gomock controller setup/teardown

**Test Results:** ✅ All 5 tests pass

**Improvements:**
- ✅ Cleaner, more readable code
- ✅ Type-safe mock expectations
- ✅ Declarative `.EXPECT()` pattern
- ✅ Automatic verification via `ctrl.Finish()`
- ✅ No manual mock maintenance needed

**Commit:** `21b7154`

---

## Migrations Attempted

### ❌ File 2: Service Test - Admin (internal/admin/service_test.go)

**Status:** BLOCKED - Missing Interfaces  
**Size:** 119 lines  
**Issue:** Custom counter interfaces not in generated mocks

**Problem:**
The admin service uses custom small interfaces:
- `UserCounter` interface with `CountUsers(ctx) (int, error)`
- `EventCounter` interface with `CountEvents(ctx) (int, error)`  
- `InviteCounter` interface with `CountInvites(ctx) (int, error)`

These interfaces are defined in `internal/admin/service.go` but were not identified during our mock generation phase (Stories 06-08).

**Finding:** We need a process to identify ALL interfaces that need mocking, including domain-specific interfaces defined in service files.

---

### ❌ File 3: Service Test - Dashboard (internal/events/dashboard_service_test.go)

**Status:** BLOCKED - Import Cycle  
**Size:** 233 lines  
**Issue:** Import cycle when importing mocks package

**Problem:**
```
import cycle:
  internal/events (test)
  → internal/testutil/mocks
  → internal/events (from mock_event_service.go)
```

The `internal/testutil/mocks` package imports `internal/events` because `mock_event_service.go` needs to reference the `events.Service` interface. When a test in `internal/events` tries to import mocks, it creates a cycle.

**Impact:** **CRITICAL** - Blocks ALL same-package service test migrations

**Root Cause:**
All mocks are in a single package (`internal/testutil/mocks`). When ANY mock in that package imports a package X, then tests in package X cannot import the mocks package.

Files affected:
- `mock_event_service.go` imports `internal/events`
- `mock_template_service.go` imports `internal/templates`
- `mock_rsvp_service.go` imports `internal/rsvp`
- `mock_invite_service.go` imports `internal/invites`
- `mock_email_service.go` imports `internal/email`

**This blocks migrations for:**
- All tests in `internal/events/` (~10 test files)
- All tests in `internal/templates/` (~3 test files)
- All tests in `internal/rsvp/` (~2 test files)
- All tests in `internal/invites/` (~5 test files)
- All tests in `internal/email/` (~2 test files)

**Estimated impact:** ~22 test files cannot be migrated with current approach

---

## Key Findings

### Finding 1: Import Cycle Issue (CRITICAL)

**Severity:** HIGH - Blocks ~22 test files  
**Type:** Architectural

**Description:**
Service mocks in `internal/testutil/mocks` create import cycles because they reference their source packages. This prevents tests in those packages from using the mocks.

**Examples:**
- `mock_event_service.go` imports `events` → tests in `events` package blocked
- `mock_template_service.go` imports `templates` → tests in `templates` package blocked

**Proposed Solutions:**

1. **Option A: Split mocks by type** (Recommended)
   - `internal/testutil/mocks/repositories/` - repo mocks only
   - `internal/testutil/mocks/services/` - service mocks only
   - Tests can import `mocks/repositories` without cycles

2. **Option B: Generate mocks in test packages**
   - Generate mocks directly in each package's test files
   - No import cycles, but mocks not reusable across packages

3. **Option C: Interface-only packages**
   - Move service interfaces to separate `*iface` packages
   - Example: `events.Service` → `eventsiface.Service`
   - Mocks import interface packages, avoiding cycles

**Recommendation:** Option A is cleanest and maintains reusability

---

### Finding 2: Missing Interface Discovery

**Severity:** MEDIUM  
**Type:** Process

**Description:**
Our mock generation phase (Stories 06-08) missed domain-specific interfaces defined in service files (e.g., UserCounter, EventCounter, InviteCounter in admin/service.go).

**Impact:**
Tests using these interfaces cannot be migrated without either:
1. Generating additional mocks
2. Keeping manual mocks for these interfaces

**Root Cause:**
Our interface discovery focused on:
- Repository interfaces in `internal/db/repositories/`
- Main service interfaces (Service, InviteService, etc.)
- Auth interfaces

We didn't scan for small domain-specific interfaces embedded in service files.

**Proposed Solution:**
Add a Story 08.5 to scan ALL .go files for `type.*interface` and generate comprehensive mock catalog.

---

### Finding 3: Context Helper Limitations

**Severity:** LOW  
**Type:** Usability

**Description:**
The `testutil.CreateAdminContext()` helpers create a fresh context from `context.Background()`, which doesn't work well for HTTP handler tests that need to preserve chi's RouteContext.

**Example:**
```go
// Doesn't work - overwrites route context
rctx := chi.NewRouteContext()
req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
req = req.WithContext(testutil.CreateAdminContext()) // ❌ Loses route context
```

**Solution:**
Context helpers are best for service-layer tests. For HTTP tests, continue using `auth.WithUser()` directly.

**Impact:** Minor - doesn't block migrations, just reduces utility of context helpers for HTTP tests

---

## Migration Patterns That Work

### Pattern 1: Handler Tests (Cross-Package Mocks)

✅ **Works when:** Test imports mocks from different packages

```go
// internal/handlers/invites_get_test.go
import (
    "github.com/lenaxia/tinyrsvp/internal/testutil/mocks"  // ✅ No cycle
    "go.uber.org/mock/gomock"
)

func TestHandler(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockService := mocks.NewMockInviteService(ctrl)
    mockRepo := mocks.NewMockEventRepository(ctrl)
    
    mockService.EXPECT().GetInviteByID(gomock.Any(), int64(1)).Return(invite, nil)
    
    handler := NewHandler(mockService, mockRepo)
    // ... test
}
```

**Success Rate:** 100% for handler tests  
**Files Tested:** 1/1 (invites_get_test.go)

---

### Pattern 2: Service Tests (Same-Package)

❌ **BLOCKED:** Import cycle prevents using mocks from same package

```go
// internal/events/dashboard_service_test.go
import (
    "github.com/lenaxia/tinyrsvp/internal/testutil/mocks"  // ❌ Import cycle!
)
```

**Error:**
```
import cycle not allowed in test:
  internal/events → internal/testutil/mocks → internal/events
```

**Success Rate:** 0% for same-package service tests  
**Files Tested:** 0/2 attempted

---

## Recommendations for Story 10

### Decision: ADJUST ⚠️

**Rationale:**
- Pattern works well for handler tests (22% code reduction, clearer expectations)
- Critical architectural issue blocks 22+ test files
- Issue is fixable with mock package restructuring
- Benefits are clear where it works

**Recommended Adjustments:**

1. **CRITICAL: Restructure mock package** (Highest Priority)
   - Split `internal/testutil/mocks/` into:
     - `internal/testutil/mocks/repositories/` (no import cycles)
     - `internal/testutil/mocks/services/` (service mocks)
     - `internal/testutil/mocks/other/` (validators, storage, etc.)
   - Update `scripts/generate_mocks.sh` to generate into subdirectories
   - Update imports in migrated test (invites_get_test.go)

2. **HIGH: Complete interface discovery** (High Priority)
   - Scan ALL .go files for interface definitions
   - Generate mocks for domain-specific interfaces (UserCounter, etc.)
   - Document which interfaces need mocks

3. **MEDIUM: Update context helpers** (Nice to have)
   - Add helper that takes existing context: `testutil.WithAdminUser(ctx)`
   - Preserve existing context values (like RouteContext)

4. **LOW: Document patterns** (Documentation)
   - Update testutil README with import cycle warnings
   - Add examples for handler vs service tests
   - Document when to use which mock import path

---

## Metrics

### Test Files Examined: 3
- ✅ internal/handlers/invites_get_test.go (314 → 244 lines, -22%)
- ⏸️ internal/admin/service_test.go (blocked - missing interfaces)
- ⏸️ internal/events/dashboard_service_test.go (blocked - import cycle)

### Lines Saved: 70 lines (22% reduction)

### Tests Passing: 5/5 (100%)

### Issues Discovered: 3
1. **CRITICAL:** Import cycle blocks 22+ test files
2. **MEDIUM:** Missing domain-specific interface mocks
3. **LOW:** Context helper limitations for HTTP tests

---

## Next Steps

1. **Implement recommended adjustments** (Story 10 tasks)
2. **Retry dashboard service test** after mock restructuring
3. **Complete 2 more validation migrations** to reach 3/3 target
4. **Make final PROCEED/ADJUST/ABORT decision**

---

## Conclusion

The migration pattern shows strong benefits where it works (22% reduction, clearer code, better safety), but a critical architectural issue with import cycles blocks a significant portion of tests.

**Recommendation:** ADJUST the approach by restructuring the mock package, then proceed with full migration.

The pattern is sound; the implementation needs refinement.
