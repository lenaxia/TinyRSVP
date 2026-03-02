# Individual Invite Creation - Complete

**Date:** 2026-01-07  
**Story:** [03_STORY_04_individual_invite.md](../00_BACKLOG/03_STORY_04_individual_invite.md)  
**Status:** ✅ Complete

---

## Summary

Implemented individual invite creation functionality with full HTTP API endpoint, comprehensive validation, permission checks, and test coverage exceeding 90%.

---

## What Was Implemented

### 1. Service Layer (`internal/invites/service_individual.go`)

Created `IndividualInviteService` interface with `CreateIndividualInvite` method:

**Key Features:**
- Email validation using regex pattern
- Case-insensitive duplicate email detection
- Event status validation (rejects cancelled/archived events)
- Permission checking (event creator or admin)
- Max plus ones validation against event limits
- Automatic token generation and hashing
- Expiration calculation (event start + 30 days)
- Normalized email storage (lowercase, trimmed)

**Request/Response Types:**
```go
type CreateIndividualInviteRequest struct {
    EventID     int64
    Name        *string
    Email       string
    MaxPlusOnes *int
}

type CreateIndividualInviteResponse struct {
    Invite *models.Invite
    Token  string
}
```

### 2. HTTP Handler (`internal/handlers/invites.go`)

Created `InviteHandlers` with POST endpoint:

**Endpoint:** `POST /api/events/:eventId/invites`

**Features:**
- Authentication required (extracts user from context)
- Event ID parsing and validation
- Request body parsing with proper error handling
- Service error mapping to HTTP status codes
- RSVP URL generation in response

**Response Format:**
```json
{
    "invite": {
        "id": 123,
        "event_id": 1,
        "email": "guest@example.com",
        "name": "John Doe",
        "max_plus_ones": 2,
        "status": "draft",
        "expires_at": "2026-02-15T00:00:00Z",
        "created_at": "2026-01-07T20:00:00Z",
        "updated_at": "2026-01-07T20:00:00Z"
    },
    "token": "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p",
    "rsvp_url": "https://rsvp.example.com/rsvp/a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p"
}
```

### 3. Comprehensive Testing

**Service Tests (`internal/invites/service_individual_test.go`):**
- ✅ Successful invite creation
- ✅ Missing email validation
- ✅ Invalid email format (5 test cases)
- ✅ Duplicate email detection
- ✅ Event not found
- ✅ Cancelled event rejection
- ✅ Archived event rejection
- ✅ Permission denied for non-creator
- ✅ Permission granted for admin
- ✅ Max plus ones default behavior
- ✅ Max plus ones custom value
- ✅ Max plus ones exceeded validation

**Handler Tests (`internal/handlers/invites_test.go`):**
- ✅ Successful invite creation
- ✅ Invalid JSON handling
- ✅ Missing email validation
- ✅ Unauthorized request (no user in context)
- ✅ Invalid event ID
- ✅ Service error: not found (404)
- ✅ Service error: permission denied (403)
- ✅ Service error: conflict (409)
- ✅ Service error: validation (400)
- ✅ Service error: internal (500)

**Integration Tests (`internal/handlers/invites_integration_test.go`):**
- ✅ Full create invite flow with database
- ✅ Duplicate email detection with real database
- ✅ Permission enforcement with multiple users

---

## Test Results

```
=== All Tests Passing ===
internal/invites:   90.4% coverage
internal/handlers:  93.0% coverage

Total: 25 test cases
- 12 service unit tests
- 10 handler unit tests  
- 3 integration tests
```

---

## Validation Rules Implemented

### Email Validation
- ✅ Required field
- ✅ Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- ✅ Max 255 characters
- ✅ Case-insensitive duplicate check
- ✅ Normalized to lowercase

### Name Validation
- ✅ Optional field
- ✅ Max 100 characters

### Max Plus Ones Validation
- ✅ Optional (defaults to event's max_plus_ones)
- ✅ Range: 0-10
- ✅ Cannot exceed event's max_plus_ones

### Event Validation
- ✅ Event must exist
- ✅ Event must not be cancelled
- ✅ Event must not be archived

### Permission Checks
- ✅ User must be event creator OR admin
- ✅ Event managers without ownership denied

---

## Error Handling

All error conditions properly handled with appropriate HTTP status codes:

| Error Condition | HTTP Status | Error Type |
|----------------|-------------|------------|
| Missing email | 400 | ValidationError |
| Invalid email | 400 | ValidationError |
| Duplicate email | 409 | ConflictError |
| Max plus ones exceeded | 400 | ValidationError |
| Event not found | 404 | NotFoundError |
| Permission denied | 403 | PermissionDeniedError |
| Cancelled event | 400 | ValidationError |
| Archived event | 400 | ValidationError |
| No authentication | 401 | - |
| Internal errors | 500 | - |

---

## Security Features

1. **Token Security**
   - Generated using crypto/rand (256 bits entropy)
   - Only hash stored in database
   - Plain token returned once in API response
   - Token never logged

2. **Permission Enforcement**
   - User authentication required
   - Event ownership verified
   - Admin override supported
   - Event status checked

3. **Email Privacy**
   - Normalized and stored lowercase
   - Case-insensitive duplicate detection
   - Not exposed in error messages

4. **Input Validation**
   - All fields validated before processing
   - Length limits enforced
   - Format validation for email
   - Range validation for max_plus_ones

---

## Files Created/Modified

### Created Files
- `internal/invites/service_individual.go` - Individual invite service implementation
- `internal/invites/service_individual_test.go` - Service unit tests
- `internal/handlers/invites.go` - HTTP handlers
- `internal/handlers/invites_test.go` - Handler unit tests
- `internal/handlers/invites_integration_test.go` - Integration tests

### Modified Files
- `internal/invites/service_test.go` - Updated mock repository with findDuplicateEmailsFunc
- `docs/00_BACKLOG/03_STORY_04_individual_invite.md` - Updated status and checklists

---

## API Usage Example

```bash
# Create an invite
curl -X POST https://rsvp.example.com/api/events/1/invites \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "email": "guest@example.com",
    "name": "John Doe",
    "max_plus_ones": 2
  }'

# Response
{
  "invite": {
    "id": 123,
    "event_id": 1,
    "email": "guest@example.com",
    "name": "John Doe",
    "max_plus_ones": 2,
    "status": "draft",
    "expires_at": "2026-02-15T00:00:00Z",
    "created_at": "2026-01-07T20:00:00Z",
    "updated_at": "2026-01-07T20:00:00Z"
  },
  "token": "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p",
  "rsvp_url": "https://rsvp.example.com/rsvp/a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p"
}
```

---

## Next Steps

This story is complete and unblocks:
- Story 05: Bulk CSV Import
- Story 06: Manual Invite
- Epic 05: Email (sending invites)

---

## Notes

- All acceptance criteria met
- Test coverage exceeds 90% requirement
- No linter warnings
- All existing tests still passing
- Ready for integration with email sending functionality
