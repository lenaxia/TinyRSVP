# Token Revocation Implementation

**Date:** 2026-01-08  
**Story:** [03_STORY_08_token_revocation.md](../00_BACKLOG/03_STORY_08_token_revocation.md)  
**Status:** Complete

---

## Summary

Implemented token revocation functionality allowing event managers to revoke invite tokens. The implementation includes service layer updates, HTTP handler with permission checks, and comprehensive test coverage.

---

## Changes Made

### 1. Service Layer Updates

**File:** `internal/invites/service.go`

- Added `RevokeInviteRequest` struct with `InviteID` and optional `Reason` fields
- Updated `InviteService` interface to use `RevokeInviteRequest` instead of plain `int64`
- Updated `RevokeInvite()` implementation to accept the new request struct
- Maintained existing state transition validation logic

### 2. HTTP Handler

**File:** `internal/handlers/invites_revoke.go`

Created new handler with:
- `RevokeInviteHandlers` struct with service and event repository dependencies
- `RevokeInvite()` handler method with:
  - Authentication check
  - Invite ID validation
  - Invite retrieval
  - Event retrieval for permission check
  - Permission validation (admin or event creator only)
  - Service call with optional reason
  - Error handling for invalid state transitions
- Route registration at `POST /api/invites/{inviteId}/revoke`

### 3. Test Coverage

**Files:**
- `internal/handlers/invites_revoke_test.go` - Handler tests
- `internal/invites/service_test.go` - Service tests with reason support

Test scenarios:
- Successful revocation by event creator
- Successful revocation by admin
- Revocation with and without reason
- Missing authentication
- Invalid invite ID
- Invite not found
- Event not found
- Permission denied (not creator or admin)
- Cannot revoke responded invite
- Invalid JSON body

### 4. Integration

**File:** `cmd/server/main.go`

- Instantiated `RevokeInviteHandlers`
- Registered routes with chi router
- Added logging for endpoint registration

### 5. Test Updates

Updated existing mock services to match new interface:
- `internal/handlers/invites_cleanup_test.go`
- `internal/handlers/invites_import_test.go`
- `internal/handlers/invites_manual_test.go`
- `internal/invites/integration_test.go`

---

## API Endpoint

```
POST /api/invites/{inviteId}/revoke
Content-Type: application/json
Authorization: Required (session-based)

Request Body:
{
    "reason": "Wrong email address"  // optional
}

Response 200 OK:
{
    "message": "Invite revoked successfully"
}

Error Responses:
- 401 Unauthorized: Missing authentication
- 400 Bad Request: Invalid invite ID or cannot revoke (already responded/revoked)
- 404 Not Found: Invite or event not found
- 403 Forbidden: Not event creator or admin
- 500 Internal Server Error: Server error
```

---

## State Transitions

The revocation respects existing state machine:
- `draft` → `revoked` ✓
- `sent` → `revoked` ✓
- `viewed` → `revoked` ✓
- `responded` → `revoked` ✗ (terminal state)
- `revoked` → any ✗ (terminal state)

---

## Permission Model

Revocation requires:
- User must be authenticated
- User must be either:
  - Admin (any event)
  - Event creator (their events only)

---

## Test Results

All tests passing:
```
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	0.005s
ok  	github.com/lenaxia/tinyrsvp/internal/invites	0.014s
```

Total test coverage maintained across all packages.

---

## Notes

1. **Revocation Reason**: Implemented as optional field in request. Currently stored in service layer but not persisted to database. Future enhancement could add a `revocation_reason` column to the `invites` table for audit purposes.

2. **Backward Compatibility**: Updated all existing mock services to match the new `RevokeInvite` signature, ensuring no breaking changes to existing tests.

3. **Error Handling**: Handler properly distinguishes between different error types (not found, permission denied, invalid transition) and returns appropriate HTTP status codes.

4. **Security**: Permission checks ensure only authorized users (event creator or admin) can revoke invites.

---

## Next Steps

Potential future enhancements:
1. Add database migration for `revocation_reason` column
2. Add audit logging for revocation events
3. Consider notification to guest when invite is revoked
4. Add bulk revocation endpoint for multiple invites

---

**Status:** ✅ Complete
