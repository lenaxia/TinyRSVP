# Event Auto-Archiving Implementation

**Date:** 2026-01-08  
**Story:** [02_STORY_06_auto_archiving.md](../00_BACKLOG/02_STORY_06_auto_archiving.md)  
**Status:** Complete

---

## Summary

Implemented automatic event archiving system that runs as a scheduled background job. Events older than 30 days with published or cancelled status are automatically archived daily.

---

## Changes Made

### 1. Fixed GetEventsToArchive Query

**File:** [`internal/db/repositories/event_repository.go`](../../internal/db/repositories/event_repository.go:402)

**Issue:** Query was excluding only archived events, but should only include published or cancelled events.

**Fix:**
```go
// Before: WHERE status != ?
// After: WHERE status IN (?, ?)
```

**Test Update:** Added test case for draft events to verify they are not included.

### 2. Created Jobs Package

**Location:** `internal/jobs/`

**Files Created:**
- [`README.md`](../../internal/jobs/README.md) - Package documentation
- [`archiver.go`](../../internal/jobs/archiver.go) - EventArchiver implementation
- [`archiver_test.go`](../../internal/jobs/archiver_test.go) - Unit tests
- [`archiver_integration_test.go`](../../internal/jobs/archiver_integration_test.go) - Integration tests

### 3. EventArchiver Implementation

**Key Features:**
- Idempotent operation (safe to run multiple times)
- Graceful error handling (continues on individual failures)
- Context cancellation support
- Comprehensive logging
- Returns error if any events fail to archive

**Interface:**
```go
type EventService interface {
    GetEventsToArchive(ctx context.Context) ([]*models.Event, error)
    ArchiveEvent(ctx context.Context, id int64) error
}
```

### 4. Scheduler Integration

**File:** [`cmd/server/main.go`](../../cmd/server/main.go:322)

**Implementation:**
- Runs every 24 hours
- Uses system user context for authorization
- Graceful shutdown support via cleanupCtx
- Logs all activity

---

## Test Coverage

### Unit Tests (7 tests)
- ✅ Constructor initialization
- ✅ No events to archive
- ✅ Single event archiving
- ✅ Multiple events archiving
- ✅ Partial failure handling
- ✅ GetEventsToArchive error handling
- ✅ Idempotency
- ✅ Context cancellation

### Integration Tests (1 test)
- ✅ Full archiving cycle with real database
- ✅ Verifies only published/cancelled events archived
- ✅ Verifies draft events not archived
- ✅ Verifies recent events not archived
- ✅ Verifies idempotency on second run

**All tests pass with timeout.**

---

## Technical Decisions

### 1. System User Context

The archiving job runs with a synthetic system user context:
```go
systemUser := &models.User{
    ID:    0,
    Email: "system@tinyrsvp.local",
    Name:  "System",
    Role:  models.RoleAdmin,
}
```

This allows the job to bypass normal authentication requirements while still respecting authorization checks.

### 2. 24-Hour Schedule

Hardcoded to run every 24 hours. Configuration support can be added in future if needed.

### 3. Error Handling Strategy

Job continues processing remaining events even if individual events fail. Returns error at end if any failures occurred, allowing monitoring systems to detect issues.

### 4. Context Cancellation

Job checks `ctx.Done()` in the loop to support graceful shutdown during server termination.

---

## Acceptance Criteria Status

- [x] Scheduled job runs daily
- [x] Job identifies events older than 30 days
- [x] Job only archives published or cancelled events
- [x] Job updates event status to archived
- [x] Job is idempotent (safe to run multiple times)
- [x] Job logs archiving activity
- [x] Job handles errors gracefully
- [x] All tests pass with timeout

---

## Definition of Done Status

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Job runs successfully
- [x] Logging implemented
- [x] Documentation complete
- [ ] Changes committed to git

---

## Next Steps

1. Commit changes
2. Consider adding configuration for:
   - Archive threshold (currently hardcoded to 30 days)
   - Schedule interval (currently hardcoded to 24 hours)
   - Job enable/disable flag

---

## Files Modified

- `internal/db/repositories/event_repository.go` - Fixed GetEventsToArchive query
- `internal/db/repositories/event_repository_test.go` - Added draft event test case
- `cmd/server/main.go` - Added event archiving scheduler

## Files Created

- `internal/jobs/README.md`
- `internal/jobs/archiver.go`
- `internal/jobs/archiver_test.go`
- `internal/jobs/archiver_integration_test.go`
