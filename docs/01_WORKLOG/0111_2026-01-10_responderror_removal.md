# Worklog: Remove respondError() and Centralize Error Handling

**Date:** 2026-01-10  
**Story:** [08_STORY_02_error_handling.md](../00_BACKLOG/08_STORY_02_error_handling.md)  
**Status:** ✅ Complete

---

## Objective

Remove the legacy `respondError()` function that bypassed centralized error handling and migrate all 108+ call sites to use `HandleError()` directly.

---

## Problem Statement

The `respondError()` wrapper function had critical issues:

1. **No request context** - Could not log with request ID
2. **No content negotiation** - Always returned JSON, never HTML for browser requests
3. **No error wrapping** - Could not preserve underlying errors
4. **Bypassed centralized system** - Most handlers didn't get proper error handling

This created inconsistency across the application and prevented proper error tracking and user experience.

---

## Solution

### 1. Removed respondError() Function

Deleted the legacy function from `internal/handlers/errors.go` entirely.

### 2. Migrated All Call Sites

Replaced all 108+ `respondError(w, status, message)` calls with `HandleError(w, r, NewXxxError(message))`:

**Files Updated:**
- `internal/handlers/events.go` (~13 call sites)
- `internal/handlers/invites.go` (~10 call sites)
- `internal/handlers/invites_list.go` (~12 call sites)
- `internal/handlers/invites_manual.go` (~8 call sites)
- `internal/handlers/invites_regenerate.go` (~8 call sites)
- `internal/handlers/invites_revoke.go` (~9 call sites)
- `internal/handlers/templates.go` (~25 call sites)
- `internal/handlers/questions.go` (~10 call sites)
- `internal/handlers/images.go` (~8 call sites)

### 3. Simplified Helper Functions

Replaced complex error mapping logic in helper functions with direct `HandleError()` calls:

```go
// Before: Complex error type checking and mapping
func handleServiceError(w http.ResponseWriter, err error) {
    var notFoundErr *models.NotFoundError
    var permErr *models.PermissionDeniedError
    // ... 20+ lines of error type checking
    switch {
    case errors.As(err, &notFoundErr):
        respondError(w, http.StatusNotFound, "event not found")
    // ... more cases
    }
}

// After: Simple delegation to centralized handler
func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
    HandleError(w, r, err)
}
```

### 4. Enhanced toAPIError()

Added support for additional error types:

```go
// Version conflicts
var versionConflictErr *models.VersionConflictError
if errors.As(err, &versionConflictErr) {
    return &APIError{
        StatusCode: http.StatusConflict,
        Code:       "VERSION_CONFLICT",
        Message:    "version conflict",
        Err:        err,
    }
}

// Assets validation errors
var assetsValidationErr *assets.ValidationError
if errors.As(err, &assetsValidationErr) {
    return &APIError{
        StatusCode: http.StatusBadRequest,
        Code:       "VALIDATION_ERROR",
        Message:    assetsValidationErr.Error(),
        Err:        err,
    }
}

// Invalid state transitions
if err.Error() == "invalid state transition" {
    return &APIError{
        StatusCode: http.StatusBadRequest,
        Code:       "BAD_REQUEST",
        Message:    "invalid state transition",
        Err:        err,
    }
}
```

### 5. Updated All Tests

**Test Changes:**
- Added `Accept: application/json` headers to all test requests expecting JSON
- Updated test expectations to match centralized error messages
- Fixed helper functions (`createMultipartRequest`) to set Accept headers
- Changed tests using `context.DeadlineExceeded` to use generic errors where 500 was expected

**Key Insight:** Tests failing with HTML responses instead of JSON indicated that content negotiation was **working correctly** - tests just needed to explicitly request JSON.

---

## Results

### ✅ All Tests Passing

```bash
go test -timeout 30s ./internal/handlers/...
ok      github.com/lenaxia/tinyrsvp/internal/handlers    0.900s
```

### ✅ Content Negotiation Verified

Integration tests confirm:
- JSON responses when `Accept: application/json` header present
- HTML responses when `Accept: text/html` header present or no Accept header
- Proper fallback behavior

### ✅ Request ID Logging Verified

All errors now logged with request IDs:
```
2026/01/09 16:49:31 [unique-request-id-12345] Error 404: Resource not found (code=NOT_FOUND)
2026/01/09 16:49:31 [] Error 404: Resource not found (code=NOT_FOUND)  # No request ID case handled
```

### ✅ Error Type Mapping Working

- NotFoundError → 404
- ValidationError → 400
- ConflictError → 409
- VersionConflictError → 409
- PermissionDeniedError → 403
- UnauthorizedError → 401
- Generic errors → 500
- context.DeadlineExceeded → 504

---

## Lessons Learned

### 1. Avoid Wrapper Functions That Bypass Core Systems

Creating convenience wrappers like `respondError()` that bypass centralized systems creates technical debt and inconsistency. Always use the core system directly.

### 2. Content Negotiation Requires Explicit Headers in Tests

When implementing content negotiation, tests must explicitly set the `Accept` header. Without it, the default behavior (HTML) is used, which can cause test failures if JSON is expected.

### 3. Error Type Completeness Matters

The centralized error handler must support ALL error types used in the application. Missing types (like `assets.ValidationError`) cause errors to be incorrectly mapped to 500 Internal Server Error.

### 4. Test Expectations Must Match Implementation

When refactoring error handling, test expectations may need updates to match the new centralized error messages. This is expected and correct - don't hack the implementation to match old test expectations.

### 5. Helper Functions Should Delegate, Not Duplicate

Helper functions like `handleServiceError()` should simply delegate to the centralized handler rather than duplicating error mapping logic. This ensures consistency and reduces maintenance burden.

---

## Impact

### Code Quality

- **Eliminated 108+ instances** of bypassed error handling
- **Removed duplicate error mapping logic** from 5+ helper functions
- **Centralized all error handling** through single `HandleError()` function
- **Improved error logging** with request context

### User Experience

- **Consistent error responses** across all endpoints
- **Proper content negotiation** - browsers get HTML, APIs get JSON
- **Better error messages** with request IDs for support

### Maintainability

- **Single source of truth** for error handling
- **Easier to add new error types** - just update `toAPIError()`
- **No duplicate logic** to maintain across handlers
- **Clear pattern** for future development

---

## Next Steps

None - story is complete. Error handling is now fully centralized and working correctly.

---

## References

- **Story:** [08_STORY_02_error_handling.md](../00_BACKLOG/08_STORY_02_error_handling.md)
- **Gap Analysis:** [2026-01-09_35_error_handling_gaps.md](2026-01-09_35_error_handling_gaps.md)
- **Commit:** 9c98f40
