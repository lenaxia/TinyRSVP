# User Story: Individual Invite Creation

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-07

---

## User Story

As an **event manager**, I want **to create individual invites with email addresses** so that **I can invite specific guests to my event**.

---

## Acceptance Criteria

- [x] Event manager can create invite for their event
- [x] Invite requires valid email address
- [x] Guest name is optional
- [x] Max plus ones defaults to event's max_plus_ones
- [x] Max plus ones cannot exceed event's max_plus_ones
- [x] Token generated automatically and securely
- [x] Token hash stored in database (not plain token)
- [x] Token returned once to caller (for email/display)
- [x] Invite expires 30 days after event date
- [x] Duplicate emails detected and rejected
- [x] Permission check: only event creator/managers can create invites
- [x] Invite created in 'draft' status
- [x] HTTP API endpoint for invite creation

---

## Technical Details

### Package Location
- `internal/invites/service.go` - Invite service
- `internal/invites/service_test.go` - Service tests
- `internal/handlers/invites.go` - HTTP handlers
- `internal/handlers/invites_test.go` - Handler tests

### Service Interface

```go
type Service interface {
    CreateInvite(ctx context.Context, req *CreateInviteRequest) (*CreateInviteResponse, error)
}

type CreateInviteRequest struct {
    EventID     int64
    Name        *string
    Email       string
    MaxPlusOnes *int
}

type CreateInviteResponse struct {
    Invite *models.Invite
    Token  string
}
```

### HTTP Endpoint

```
POST /api/events/:eventId/invites
Content-Type: application/json

{
    "email": "guest@example.com",
    "name": "John Doe",
    "max_plus_ones": 2
}

Response 201 Created:
{
    "invite": {
        "id": 123,
        "event_id": 1,
        "email": "guest@example.com",
        "name": "John Doe",
        "max_plus_ones": 2,
        "status": "draft",
        "expires_at": "2026-02-15T00:00:00Z",
        "created_at": "2026-01-07T20:00:00Z"
    },
    "token": "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p",
    "rsvp_url": "https://rsvp.example.com/rsvp/a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p"
}
```

---

## Subtasks

### Service Implementation
- [x] Create `InviteService` interface
- [x] Implement `CreateInvite()` method
- [x] Validate email format
- [x] Check for duplicate email in event
- [x] Verify event exists and is not cancelled
- [x] Check user has permission to create invites
- [x] Validate max_plus_ones against event limit
- [x] Generate secure token
- [x] Hash token for storage
- [x] Calculate expiration date (event date + 30 days)
- [x] Create invite in database
- [x] Return invite with plain token

### Handler Implementation
- [x] Create POST `/api/events/:eventId/invites` endpoint
- [x] Parse and validate request body
- [x] Extract user from context
- [x] Call invite service
- [x] Format response with RSVP URL
- [x] Handle errors appropriately

### Testing
- [x] Test successful invite creation
- [x] Test duplicate email rejection
- [x] Test invalid email format
- [x] Test max_plus_ones validation
- [x] Test permission checks
- [x] Test non-existent event
- [x] Test cancelled event
- [x] Test token generation uniqueness
- [x] Test expiration date calculation
- [x] Integration test full flow

### Documentation
- [x] API endpoint documentation
- [x] Request/response examples
- [x] Error codes and messages
- [x] Permission requirements

---

## Dependencies

**Depends on:**
- Story 00: Token Generation
- Story 01: Token Hashing
- Story 03: Invite Model
- Epic 02: Events (event must exist)
- Epic 01: Auth (permission checks)

**Blocks:**
- Story 05: Bulk CSV Import
- Story 06: Manual Invite
- Epic 05: Email (sending invites)

---

## Testing Strategy

### Unit Tests

1. **Service Tests**
   ```go
   func TestInviteService_CreateInvite_Success(t *testing.T)
   func TestInviteService_CreateInvite_DuplicateEmail(t *testing.T)
   func TestInviteService_CreateInvite_InvalidEmail(t *testing.T)
   func TestInviteService_CreateInvite_MaxPlusOnesExceeded(t *testing.T)
   func TestInviteService_CreateInvite_PermissionDenied(t *testing.T)
   func TestInviteService_CreateInvite_EventNotFound(t *testing.T)
   func TestInviteService_CreateInvite_CancelledEvent(t *testing.T)
   ```

2. **Handler Tests**
   ```go
   func TestInviteHandler_CreateInvite_Success(t *testing.T)
   func TestInviteHandler_CreateInvite_InvalidJSON(t *testing.T)
   func TestInviteHandler_CreateInvite_MissingEmail(t *testing.T)
   func TestInviteHandler_CreateInvite_Unauthorized(t *testing.T)
   ```

3. **Integration Tests**
   ```go
   func TestCreateInvite_Integration(t *testing.T) {
       // Create event
       // Create invite
       // Verify token works
       // Verify expiration correct
   }
   ```

---

## Validation Rules

### Email Validation
- Required field
- Valid email format: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- Max 255 characters
- Case-insensitive duplicate check

### Name Validation
- Optional field
- Max 100 characters
- Sanitized for XSS

### Max Plus Ones Validation
- Optional (defaults to event's max_plus_ones)
- Integer 0-10
- Cannot exceed event's max_plus_ones

### Event Validation
- Event must exist
- Event must not be cancelled
- Event must not be archived

---

## Permission Checks

User must have one of:
- Created the event (`created_by` = user_id)
- Admin role
- Event manager role with access to event

---

## Error Handling

| Error Condition | Error Type | HTTP Status | User Message |
|----------------|------------|-------------|--------------|
| Missing email | `ValidationError` | 400 | "Email is required" |
| Invalid email | `ValidationError` | 400 | "Invalid email format" |
| Duplicate email | `ConflictError` | 409 | "Email already invited to this event" |
| Max plus ones exceeded | `ValidationError` | 400 | "Max plus ones exceeds event limit" |
| Event not found | `NotFoundError` | 404 | "Event not found" |
| Permission denied | `PermissionDeniedError` | 403 | "Not authorized to create invites" |
| Cancelled event | `ValidationError` | 400 | "Cannot invite to cancelled event" |
| Token generation failed | `InternalError` | 500 | "Failed to generate invite token" |

---

## Business Logic

### Expiration Calculation
```go
expiresAt := event.StartTime.Add(30 * 24 * time.Hour)
```

### Max Plus Ones Default
```go
if req.MaxPlusOnes == nil {
    maxPlusOnes = event.MaxPlusOnes
} else {
    maxPlusOnes = *req.MaxPlusOnes
    if maxPlusOnes > event.MaxPlusOnes {
        return ErrMaxPlusOnesExceeded
    }
}
```

### Duplicate Detection
```go
existing, err := s.repo.FindByEventAndEmail(ctx, eventID, email)
if err == nil && existing != nil {
    return ErrDuplicateEmail
}
```

---

## Security Considerations

1. **Token Security**
   - Token generated using crypto/rand
   - Only hash stored in database
   - Token returned once in response
   - Token never logged (only last 6 chars)

2. **Permission Enforcement**
   - Verify user owns event or is admin
   - Check event status before allowing invites
   - Validate all input fields

3. **Email Privacy**
   - Email stored encrypted (future enhancement)
   - Email not exposed in logs
   - Duplicate check case-insensitive

4. **XSS Prevention**
   - Sanitize name field
   - Escape output in templates
   - Use JSON encoding for API

---

## References

- **HLD:** Section 6.1 (Invite Creation)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md) Section 4.3
- **Story 00:** [03_STORY_00_token_generation.md](03_STORY_00_token_generation.md)
- **Story 03:** [03_STORY_03_invite_model.md](03_STORY_03_invite_model.md)
- **Similar:** Event creation flow

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Service layer implemented and tested
- [x] HTTP handler implemented and tested
- [x] Unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Permission checks working
- [x] Validation rules enforced
- [x] Error handling complete
- [x] Documentation updated
- [x] Code reviewed
- [x] No linter warnings
