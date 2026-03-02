# Worklog: Token Expiration & Cleanup Implementation

**Date:** 2026-01-07  
**Story:** [03_STORY_07_token_expiration.md](../00_BACKLOG/03_STORY_07_token_expiration.md)  
**Status:** Complete

---

## Summary

Implemented automatic token expiration and cleanup functionality for invite tokens, including expiration validation, scheduled cleanup jobs, and manual admin cleanup endpoint.

---

## Changes Made

### 1. Service Layer (`internal/invites/service.go`)

**Added expiration check to token validation:**
- Modified [`GetInviteByToken()`](internal/invites/service.go:114) to check if invite has expired
- Returns error if `ExpiresAt` is before current time
- Prevents expired tokens from being used for RSVP

**Added cleanup method:**
- Implemented [`CleanupExpiredTokens()`](internal/invites/service.go:310) method
- Calls repository's `DeleteExpired()` method
- Returns count of deleted tokens
- Proper error handling and wrapping

**Updated interface:**
- Added `CleanupExpiredTokens(ctx context.Context) (int64, error)` to [`InviteService`](internal/invites/service.go:40) interface

### 2. Handler Layer (`internal/handlers/invites_cleanup.go`)

**Created new cleanup handler:**
- [`NewCleanupHandler()`](internal/handlers/invites_cleanup.go:13) constructor
- POST-only endpoint
- Returns JSON with deleted count and success message
- Proper error handling with 500 status on failure

### 3. Main Application (`cmd/server/main.go`)

**Added scheduled cleanup job:**
- Background goroutine running every 24 hours
- Calls `CleanupExpiredTokens()` on schedule
- Logs success/failure with statistics
- Properly integrated with shutdown context

**Registered cleanup endpoint:**
- Path: `/api/invites/cleanup`
- Method: POST
- Protection: Admin only (requires auth + admin role)
- Logged endpoint registration

### 4. Testing

**Service tests (`internal/invites/service_test.go`):**
- [`TestInviteService_GetInviteByToken_ExpiredToken`](internal/invites/service_test.go:534) - Tests expired token rejection
- [`TestInviteService_CleanupExpiredTokens`](internal/invites/service_test.go:607) - Tests cleanup functionality
- Multiple test cases covering happy and unhappy paths

**Handler tests (`internal/handlers/invites_cleanup_test.go`):**
- [`TestCleanupExpiredTokensHandler_Success`](internal/handlers/invites_cleanup_test.go:56) - Tests successful cleanup
- [`TestCleanupExpiredTokensHandler_NoExpiredTokens`](internal/handlers/invites_cleanup_test.go:82) - Tests no tokens to clean
- [`TestCleanupExpiredTokensHandler_ServiceError`](internal/handlers/invites_cleanup_test.go:108) - Tests error handling
- [`TestCleanupExpiredTokensHandler_InvalidMethod`](internal/handlers/invites_cleanup_test.go:134) - Tests method validation

**Mock updates:**
- Updated [`mockImportService`](internal/handlers/invites_import_test.go:22) with `CleanupExpiredTokens` method
- Updated [`mockManualInviteService`](internal/handlers/invites_manual_test.go:21) with `CleanupExpiredTokens` method
- Created [`mockInviteServiceWithCleanup`](internal/handlers/invites_cleanup_test.go:18) for cleanup tests

---

## Test Results

```
go test -timeout 30s -race -cover ./...
```

**All tests passing:**
- internal/auth: 88.2% coverage
- internal/invites: 92.0% coverage
- internal/handlers: 91.3% coverage
- No race conditions detected

---

## Technical Notes

### Expiration Logic

Tokens expire based on the `ExpiresAt` field in the invite model:
- Set during invite creation (event start time + 30 days)
- Validated on token retrieval
- Used by cleanup job to identify expired tokens

### Cleanup Schedule

- **Automatic:** Runs every 24 hours via background goroutine
- **Manual:** Admin-only POST endpoint at `/api/invites/cleanup`
- Both use same service method for consistency

### Security

- Manual cleanup endpoint protected by admin middleware
- Expiration check prevents expired token usage
- Cleanup job runs in background without blocking main thread

---

## API Endpoint

### POST /api/invites/cleanup

**Protection:** Admin only

**Response (Success):**
```json
{
  "deleted": 5,
  "message": "Expired tokens cleaned up successfully"
}
```

**Response (Error):**
```json
{
  "error": "error message"
}
```

---

## Dependencies

- Repository layer already had `DeleteExpired()` method
- Invite model already had `ExpiresAt` field
- No database schema changes required

---

## Next Steps

None - story complete. Ready for next story in Epic 3.

---

## Verification

- [x] All tests pass
- [x] Build succeeds
- [x] Race detector clean
- [x] Coverage >90%
- [x] Story checklist updated
