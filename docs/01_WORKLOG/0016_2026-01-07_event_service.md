# Event Service Layer Implementation

**Date:** 2026-01-07  
**Story:** [02_STORY_03_event_service.md](../00_BACKLOG/02_STORY_03_event_service.md)  
**Status:** Complete

---

## Summary

Implemented the Event Service Layer following TDD principles, providing business logic for event operations with validation, authorization, and state management.

---

## Changes Made

### New Files Created

1. **`internal/events/service.go`**
   - Implemented Service interface with 9 methods
   - Permission checking on all operations
   - State transition validation
   - Optimistic locking support
   - User context extraction

2. **`internal/events/service_test.go`**
   - Comprehensive unit tests for all service methods
   - Mock implementations for dependencies
   - 87.8% test coverage
   - All tests passing

### Modified Files

1. **`internal/models/errors.go`**
   - Added `PermissionDeniedError` type
   - Supports optional ID field for resource-specific errors

2. **`internal/models/user.go`**
   - Added `RoleGuest` constant for testing

3. **`internal/auth/permissions_test.go`**
   - Fixed test using non-existent `EventStatusCompleted`
   - Removed duplicate test case

4. **`internal/events/README.md`**
   - Updated to document Service layer
   - Added usage examples
   - Updated structure section

---

## Implementation Details

### Service Interface

```go
type Service interface {
    CreateEvent(ctx context.Context, event *models.Event) error
    GetEvent(ctx context.Context, id int64) (*models.Event, error)
    UpdateEvent(ctx context.Context, event *models.Event) error
    DeleteEvent(ctx context.Context, id int64) error
    ListEvents(ctx context.Context, filters ListFilters) ([]*models.Event, error)
    PublishEvent(ctx context.Context, id int64) error
    CancelEvent(ctx context.Context, id int64, reason string) error
    ArchiveEvent(ctx context.Context, id int64) error
    GetEventsToArchive(ctx context.Context) ([]*models.Event, error)
}
```

### Permission Model

- **CreateEvent**: EventManager or Admin
- **GetEvent**: EventManager or Admin
- **UpdateEvent**: Owner or Admin
- **DeleteEvent**: Owner or Admin (draft/published only)
- **ListEvents**: EventManager (own events) or Admin (all events)
- **PublishEvent**: Owner or Admin
- **CancelEvent**: Owner or Admin
- **ArchiveEvent**: Admin only
- **GetEventsToArchive**: Admin only

### State Transitions

Enforced through validator:
- Draft → Published
- Draft → Cancelled
- Published → Cancelled
- Published → Archived
- Cancelled → Archived

---

## Test Results

```
go test -timeout 30s ./internal/events/...
PASS
ok      github.com/lenaxia/tinyrsvp/internal/events    0.013s

go test -timeout 30s -cover ./internal/events/...
ok      github.com/lenaxia/tinyrsvp/internal/events    0.006s    coverage: 87.8% of statements

go test -timeout 30s ./...
ok      github.com/lenaxia/tinyrsvp/internal/auth      2.685s
ok      github.com/lenaxia/tinyrsvp/internal/config    (cached)
ok      github.com/lenaxia/tinyrsvp/internal/db        (cached)
ok      github.com/lenaxia/tinyrsvp/internal/db/repositories    (cached)
ok      github.com/lenaxia/tinyrsvp/internal/events    (cached)
ok      github.com/lenaxia/tinyrsvp/internal/handlers  (cached)
ok      github.com/lenaxia/tinyrsvp/internal/middleware        (cached)
ok      github.com/lenaxia/tinyrsvp/internal/models    (cached)
ok      github.com/lenaxia/tinyrsvp/tests/e2e          (cached)
```

All tests passing across entire codebase.

---

## Key Design Decisions

1. **User Context Extraction**: All service methods extract user from context first, returning PermissionDeniedError if missing

2. **Permission-First Approach**: Authorization checks happen before any business logic or database operations

3. **Optimistic Locking**: UpdateEvent uses UpdateWithVersion from repository to prevent concurrent modification issues

4. **Status Preservation**: UpdateEvent preserves existing event status, preventing accidental status changes through update operations

5. **Admin Override**: Admins can perform all operations except those explicitly restricted (like GetEventsToArchive which is admin-only anyway)

6. **Non-Admin Filtering**: ListEvents automatically filters to user's own events for non-admins

---

## Testing Approach

Followed strict TDD:
1. Wrote comprehensive tests first
2. Implemented service to make tests pass
3. Verified all tests pass
4. Checked coverage (87.8%)

Test coverage includes:
- Happy paths for all operations
- Permission denial scenarios
- Validation errors
- Repository errors
- State transition errors
- Optimistic locking conflicts
- Missing user context
- Not found errors

---

## Next Steps

This completes Epic 2 Story 3. The event service layer is now ready for:
- Event HTTP handlers (Story 4)
- Integration with invite system
- Frontend integration

---

## References

- **Story:** [02_STORY_03_event_service.md](../00_BACKLOG/02_STORY_03_event_service.md)
- **Epic:** [02_EPIC_events.md](../00_BACKLOG/02_EPIC_events.md)
- **LLD:** [02_EVENT_LLD.md](../lld/02_EVENT_LLD.md)
