# Epic 3 Story 05 - Bulk CSV Import Gaps Fixed

**Date:** 2026-01-07  
**Status:** ✅ Complete  
**Story:** [`docs/00_BACKLOG/03_STORY_05_bulk_csv_import.md`](../00_BACKLOG/03_STORY_05_bulk_csv_import.md)

---

## Summary

Fixed all critical and major gaps in the bulk CSV import implementation identified in the previous worklog. The import endpoint is now fully functional with proper security, validation, and integration with the rest of the system.

---

## Gaps Addressed

### Gap 1: Handler Not Registered (CRITICAL) ✅
**Issue:** ImportInviteHandlers was created but never registered in main.go, making the endpoint inaccessible.

**Solution:**
- Registered ImportInviteHandlers in [`cmd/server/main.go`](../../cmd/server/main.go:230)
- Added proper logging for the import endpoint registration
- Endpoint now accessible at `POST /api/events/:eventId/invites/import`

### Gap 2: Service Architecture Mismatch (CRITICAL) ✅
**Issue:** Two separate service interfaces with no integration:
- IndividualInviteService (used in main.go) - only CreateIndividualInvite()
- InviteService (required by ImportInviteHandlers) - has ImportCSV() but not instantiated

**Solution:**
- Updated [`cmd/server/main.go`](../../cmd/server/main.go:125) to instantiate both services:
  - `inviteService` (InviteService) - for bulk operations
  - `individualInviteService` (IndividualInviteService) - for individual operations
- InviteHandlers uses individualInviteService
- ImportInviteHandlers uses inviteService
- Both services share the same token generator and repositories

### Gap 3: Missing Permission Checks (CRITICAL SECURITY) ✅
**Issue:** Import handler didn't check user permissions, allowing any authenticated user to import invites to any event.

**Solution:**
- Added event retrieval in [`ImportInvites`](../../internal/handlers/invites.go:169) handler
- Implemented permission check: user must be admin OR event creator
- Returns 403 Forbidden if unauthorized
- Added comprehensive tests in [`invites_import_permission_test.go`](../../internal/handlers/invites_import_permission_test.go)

### Gap 4: Hardcoded Expiration Time (MAJOR) ✅
**Issue:** Import handler hardcoded expiration to 60 days instead of event.StartTime + 30 days.

**Solution:**
- Updated [`ImportInvites`](../../internal/handlers/invites.go:203) to calculate: `event.StartTime.Add(30 * 24 * time.Hour)`
- Now consistent with individual invite creation
- Added test [`TestImportInvitesHandler_CorrectExpirationTime`](../../internal/handlers/invites_import_permission_test.go:235) to verify

### Gap 5: Missing Event Validation (MAJOR) ✅
**Issue:** Import service didn't validate event status, could create invites for cancelled/archived events.

**Solution:**
- Added event status checks in [`ImportInvites`](../../internal/handlers/invites.go:178) handler:
  - Reject cancelled events with 400 Bad Request
  - Reject archived events with 400 Bad Request
- Added tests for both scenarios
- Validation happens before file processing for efficiency

### Gap 6: Inconsistent Default MaxPlusOnes (MEDIUM) ✅
**Issue:** Import handler passed defaultMaxPlusOnes: 0 instead of using event's MaxPlusOnes setting.

**Solution:**
- Updated [`ImportInvites`](../../internal/handlers/invites.go:203) to pass `event.MaxPlusOnes`
- Now consistent with event configuration
- Added test [`TestImportInvitesHandler_CorrectDefaultMaxPlusOnes`](../../internal/handlers/invites_import_permission_test.go:289) to verify

---

## Implementation Details

### Service Architecture

**Before:**
```go
// main.go
inviteService := invites.NewIndividualInviteService(...)
inviteHandlers := handlers.NewInviteHandlers(inviteService, ...)
// ImportInviteHandlers never registered
```

**After:**
```go
// main.go
inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)
individualInviteService := invites.NewIndividualInviteService(tokenGenerator, inviteRepo, eventRepo)

inviteHandlers := handlers.NewInviteHandlers(individualInviteService, cfg.Server.BaseURL)
inviteHandlers.RegisterRoutes(chiRouter)

importInviteHandlers := handlers.NewImportInviteHandlers(inviteService, eventRepo, cfg.Server.BaseURL)
importInviteHandlers.RegisterRoutes(chiRouter)
```

### Handler Updates

**ImportInviteHandlers struct:**
```go
type ImportInviteHandlers struct {
    service   invites.InviteService
    eventRepo repositories.EventRepository  // Added
    baseURL   string
}
```

**ImportInvites method flow:**
1. Authenticate user
2. Parse and validate event ID
3. **Retrieve event from database** (NEW)
4. **Validate event status** (NEW)
5. **Check user permission** (NEW)
6. Parse multipart form
7. Validate file
8. Read CSV data
9. **Calculate expiration from event.StartTime** (FIXED)
10. **Pass event.MaxPlusOnes as default** (FIXED)
11. Call service.ImportCSV()
12. Return result

---

## Test Coverage

### New Tests Added

**Permission Tests** ([`invites_import_permission_test.go`](../../internal/handlers/invites_import_permission_test.go)):
- `TestImportInvitesHandler_PermissionDenied_NotAdmin_NotCreator` - Verifies 403 for unauthorized users
- `TestImportInvitesHandler_PermissionGranted_Admin` - Verifies admin can import
- `TestImportInvitesHandler_PermissionGranted_Creator` - Verifies creator can import
- `TestImportInvitesHandler_EventNotFound` - Verifies 404 for non-existent events
- `TestImportInvitesHandler_CancelledEvent` - Verifies 400 for cancelled events
- `TestImportInvitesHandler_ArchivedEvent` - Verifies 400 for archived events
- `TestImportInvitesHandler_CorrectExpirationTime` - Verifies expiration calculation
- `TestImportInvitesHandler_CorrectDefaultMaxPlusOnes` - Verifies MaxPlusOnes default

### Existing Tests Updated

All existing tests in [`invites_import_test.go`](../../internal/handlers/invites_import_test.go) and [`invites_import_integration_test.go`](../../internal/handlers/invites_import_integration_test.go) updated to:
- Pass EventRepository to NewImportInviteHandlers constructor
- Use helper function `createMockEventRepo()` for consistent test setup

---

## Test Results

```bash
$ go test -timeout 30s ./...
ok  	github.com/lenaxia/tinyrsvp/internal/auth	(cached)
ok  	github.com/lenaxia/tinyrsvp/internal/config	(cached)
ok  	github.com/lenaxia/tinyrsvp/internal/db	(cached)
ok  	github.com/lenaxia/tinyrsvp/internal/db/repositories	(cached)
ok  	github.com/lenaxia/tinyrsvp/internal/events	(cached)
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	0.782s
ok  	github.com/lenaxia/tinyrsvp/internal/invites	0.010s
ok  	github.com/lenaxia/tinyrsvp/internal/middleware	(cached)
ok  	github.com/lenaxia/tinyrsvp/internal/models	(cached)
ok  	github.com/lenaxia/tinyrsvp/pkg/token	(cached)
ok  	github.com/lenaxia/tinyrsvp/tests/e2e	0.746s
```

All tests passing ✅

---

## Security Improvements

1. **Permission Enforcement:** Only admins and event creators can import invites
2. **Event Status Validation:** Cannot import to cancelled or archived events
3. **Event Existence Check:** Returns 404 for non-existent events
4. **Proper Error Messages:** Clear feedback for security violations

---

## API Behavior

### Endpoint
`POST /api/events/:eventId/invites/import`

### Authorization
- Requires authentication
- User must be admin OR event creator
- Returns 403 Forbidden if unauthorized

### Event Validation
- Event must exist (404 if not found)
- Event must not be cancelled (400 if cancelled)
- Event must not be archived (400 if archived)

### Expiration Calculation
- Calculated as: `event.StartTime + 30 days`
- Consistent with individual invite creation

### Default MaxPlusOnes
- Uses event's MaxPlusOnes setting
- Can be overridden per-invite in CSV

---

## Files Modified

1. [`cmd/server/main.go`](../../cmd/server/main.go)
   - Instantiate unified InviteService
   - Register ImportInviteHandlers
   - Add logging for import endpoint

2. [`internal/handlers/invites.go`](../../internal/handlers/invites.go)
   - Add EventRepository to ImportInviteHandlers
   - Implement permission checks
   - Fix expiration calculation
   - Fix defaultMaxPlusOnes
   - Add event status validation

3. [`internal/handlers/invites_import_test.go`](../../internal/handlers/invites_import_test.go)
   - Update all tests for new constructor signature
   - Add helper function for mock event repo

4. [`internal/handlers/invites_import_integration_test.go`](../../internal/handlers/invites_import_integration_test.go)
   - Update integration tests for new constructor

5. [`internal/handlers/invites_import_permission_test.go`](../../internal/handlers/invites_import_permission_test.go) (NEW)
   - Comprehensive permission and validation tests

---

## Verification

### Manual Testing Checklist
- [ ] Import endpoint accessible at `/api/events/:eventId/invites/import`
- [ ] Non-creator/non-admin receives 403
- [ ] Admin can import to any event
- [ ] Creator can import to their own event
- [ ] Cannot import to cancelled event
- [ ] Cannot import to archived event
- [ ] Expiration set to event.StartTime + 30 days
- [ ] Default MaxPlusOnes uses event setting

### Automated Testing
- ✅ All unit tests passing
- ✅ All integration tests passing
- ✅ Permission tests passing
- ✅ Validation tests passing

---

## Next Steps

Epic 3 Story 05 is now complete with all gaps addressed. The bulk CSV import feature is:
- ✅ Fully functional
- ✅ Properly secured
- ✅ Correctly integrated
- ✅ Well tested

Ready to proceed with remaining Epic 3 stories:
- Story 06: Manual Invite Creation
- Story 07: Token Expiration
- Story 08: Token Revocation
- Story 09: Token Regeneration
- Story 10: Invite Tracking
- Story 11: Invite Listing

---

## Technical Notes

### Service Pattern
The dual-service approach allows:
- **InviteService**: Core invite operations (create, import, revoke, list)
- **IndividualInviteService**: Business logic for individual invites (permission checks, event validation)

This separation maintains clean architecture while avoiding code duplication.

### Permission Model
Permission checks follow the established pattern:
1. Admin users have full access
2. Event creators have access to their events
3. All other users denied

### Event Validation
Event status validation prevents:
- Creating invites for events that won't happen (cancelled)
- Creating invites for past events (archived)
- Wasting resources on invalid operations

---

**Commit:** `301c33d` - Fix Epic 3 Story 05 gaps
