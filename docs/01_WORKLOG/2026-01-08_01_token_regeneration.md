# Token Regeneration Implementation

**Date:** 2026-01-08  
**Story:** [03_STORY_09_token_regeneration.md](../00_BACKLOG/03_STORY_09_token_regeneration.md)  
**Status:** Complete

---

## Summary

Implemented token regeneration functionality allowing event managers to generate new invite tokens when the original is compromised or lost. The old token is immediately invalidated and a new secure token is generated.

---

## Changes Made

### Service Layer

**File:** [`internal/invites/service.go`](../../internal/invites/service.go)

- Added `RegenerateTokenResponse` struct with `Token` and `RSVPURL` fields
- Added `RegenerateToken()` method to `InviteService` interface
- Implemented `RegenerateToken()` method in `inviteService`:
  - Validates invite exists
  - Prevents regeneration of revoked invites
  - Prevents regeneration of responded invites
  - Generates new secure token
  - Hashes new token
  - Updates invite with new token hash
  - Returns new token and RSVP URL

### Repository Layer

**File:** [`internal/db/repositories/invite_repository.go`](../../internal/db/repositories/invite_repository.go)

- Updated `Update()` method to include `token_hash` and `revocation_reason` in UPDATE statement
- This allows the service to update the token hash when regenerating

### Handler Layer

**File:** [`internal/handlers/invites_regenerate.go`](../../internal/handlers/invites_regenerate.go)

- Created `RegenerateInviteTokenService` interface
- Created `RegenerateInviteTokenHandlers` struct
- Implemented `RegenerateInviteToken()` HTTP handler:
  - Validates authentication
  - Validates invite ID parameter
  - Checks invite exists
  - Checks event exists
  - Validates permissions (event creator or admin)
  - Calls service to regenerate token
  - Returns new token and RSVP URL

**File:** [`cmd/server/main.go`](../../cmd/server/main.go)

- Registered regenerate invite handlers with router
- Added logging for regenerate endpoint

---

## Testing

### Service Tests

**File:** [`internal/invites/service_regenerate_test.go`](../../internal/invites/service_regenerate_test.go)

Created comprehensive integration tests:
- `TestRegenerateToken_Success` - Verifies successful regeneration
- `TestRegenerateToken_OldTokenInvalidated` - Confirms old token cannot be used
- `TestRegenerateToken_NewTokenWorks` - Confirms new token works
- `TestRegenerateToken_CannotRegenerateRevoked` - Prevents revoked invite regeneration
- `TestRegenerateToken_CannotRegenerateResponded` - Prevents responded invite regeneration
- `TestRegenerateToken_InviteNotFound` - Handles missing invite
- `TestRegenerateToken_PreservesSentStatus` - Preserves sent status
- `TestRegenerateToken_PreservesViewedStatus` - Preserves viewed status

### Handler Tests

**File:** [`internal/handlers/invites_regenerate_test.go`](../../internal/handlers/invites_regenerate_test.go)

Created HTTP handler tests:
- Successful regeneration by event creator
- Successful regeneration by admin
- Missing authentication
- Invalid invite ID
- Invite not found
- Event not found
- Permission denied
- Cannot regenerate revoked invite
- Cannot regenerate responded invite

### Test Results

All tests pass:
```
ok  	github.com/lenaxia/tinyrsvp/internal/invites	0.065s
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	1.012s
```

---

## API Endpoint

```
POST /api/invites/:inviteId/regenerate
Authorization: Required (event creator or admin)

Response 200 OK:
{
    "token": "b4G9lM0nO3qR6sU8wX1yZ5aB7cD9eF2gH4jK6m",
    "rsvp_url": "/rsvp/b4G9lM0nO3qR6sU8wX1yZ5aB7cD9eF2gH4jK6m"
}

Response 400 Bad Request:
{
    "error": "cannot regenerate token for revoked invite"
}

Response 403 Forbidden:
{
    "error": "permission denied"
}

Response 404 Not Found:
{
    "error": "invite not found"
}
```

---

## Security Considerations

1. **Immediate Invalidation**: Old token is immediately invalidated when new token is generated
2. **No Grace Period**: No overlap between old and new tokens
3. **Independent Tokens**: New token is completely independent from old token
4. **Permission Checks**: Only event creator or admin can regenerate
5. **Status Validation**: Cannot regenerate revoked or responded invites
6. **Audit Trail**: Token hash changes are tracked via `updated_at` timestamp

---

## Use Cases Supported

1. **Token Compromised**: Guest accidentally shared link publicly
2. **Token Lost**: Guest deleted email, needs new link
3. **Wrong Channel**: Sent via wrong medium, need to resend
4. **Testing**: Generate new token for testing purposes

---

## Technical Notes

- Token regeneration preserves all invite metadata (name, email, max_plus_ones, status)
- Status transitions remain unchanged (draft/sent/viewed preserved)
- The repository `Update()` method was enhanced to support token hash updates
- All existing tests continue to pass after changes

---

## Next Steps

Story complete. Ready for:
- Story 10: Invite Tracking
- Story 11: Invite Listing
