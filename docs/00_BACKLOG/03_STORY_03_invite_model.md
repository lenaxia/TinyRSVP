# User Story: Invite Model & Repository

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-07

---

## User Story

As a **system developer**, I want **invite data model and repository** so that **invites can be stored, retrieved, and managed in the database**.

---

## Acceptance Criteria

- [x] Invite struct defined with all required fields
- [x] Invite repository interface defined
- [x] Repository implements CRUD operations
- [x] Repository handles token hash lookups
- [x] Repository supports filtering by event ID
- [x] Repository supports filtering by status
- [x] Repository handles duplicate email detection
- [x] All database operations use transactions where appropriate
- [x] Optimistic locking for concurrent updates (not needed - no version field)
- [x] Comprehensive test coverage (>90%)

---

## Technical Details

### Package Location
- `internal/models/invite.go` - Invite struct
- `internal/db/repositories/invite_repository.go` - Repository implementation
- `internal/db/repositories/invite_repository_test.go` - Tests

### Invite Model

```go
package models

import "time"

type Invite struct {
    ID           int64      `json:"id"`
    EventID      int64      `json:"event_id"`
    Name         *string    `json:"name,omitempty"`
    Email        *string    `json:"email,omitempty"`
    TokenHash    string     `json:"-"`
    MaxPlusOnes  int        `json:"max_plus_ones"`
    Status       string     `json:"status"`
    SentAt       *time.Time `json:"sent_at,omitempty"`
    ViewedAt     *time.Time `json:"viewed_at,omitempty"`
    Unsubscribed bool       `json:"unsubscribed"`
    EmailInvalid bool       `json:"email_invalid"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
    ExpiresAt    time.Time  `json:"expires_at"`
}

const (
    InviteStatusDraft     = "draft"
    InviteStatusSent      = "sent"
    InviteStatusViewed    = "viewed"
    InviteStatusResponded = "responded"
    InviteStatusRevoked   = "revoked"
)
```

### Repository Interface

```go
package repositories

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type InviteRepository interface {
    Create(ctx context.Context, invite *models.Invite) error
    CreateBatch(ctx context.Context, invites []*models.Invite) error
    GetByID(ctx context.Context, id int64) (*models.Invite, error)
    GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error)
    Update(ctx context.Context, invite *models.Invite) error
    Delete(ctx context.Context, id int64) error
    ListByEventID(ctx context.Context, eventID int64, filters InviteFilters) ([]*models.Invite, error)
    CountByEventID(ctx context.Context, eventID int64) (int, error)
    GetStats(ctx context.Context, eventID int64) (*InviteStats, error)
    FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error)
    DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type InviteFilters struct {
    Status       *string
    Unsubscribed *bool
    EmailInvalid *bool
    Limit        int
    Offset       int
}

type InviteStats struct {
    Total      int
    Draft      int
    Sent       int
    Viewed     int
    Responded  int
    Revoked    int
}
```

---

## Subtasks

### Model Implementation
- [x] Create `internal/models/invite.go`
- [x] Define `Invite` struct with all fields
- [x] Define invite status constants
- [x] Add JSON tags for API serialization
- [x] Add validation tags if using validator library
- [x] Implement `Validate()` method for business rules

### Repository Implementation
- [x] Create `InviteRepository` interface
- [x] Implement `inviteRepository` struct
- [x] Implement `Create()` - insert single invite
- [x] Implement `CreateBatch()` - bulk insert with transaction
- [x] Implement `GetByID()` - retrieve by primary key
- [x] Implement `GetByTokenHash()` - retrieve by token hash (unique index)
- [x] Implement `Update()` - update invite fields
- [x] Implement `Delete()` - soft or hard delete
- [x] Implement `ListByEventID()` - list with filters and pagination
- [x] Implement `CountByEventID()` - count invites for event
- [x] Implement `GetStats()` - aggregate statistics by status
- [x] Implement `FindDuplicateEmails()` - check for existing emails
- [x] Implement `DeleteExpired()` - cleanup expired tokens

### Testing
- [x] Test invite struct validation
- [x] Test Create() with valid invite
- [x] Test Create() with duplicate token hash (should fail)
- [x] Test CreateBatch() with multiple invites
- [x] Test CreateBatch() rollback on error
- [x] Test GetByID() found and not found
- [x] Test GetByTokenHash() found and not found
- [x] Test Update() with valid changes
- [x] Test Update() with concurrent modifications (not applicable - no version field)
- [x] Test Delete() removes invite
- [x] Test ListByEventID() with various filters
- [x] Test ListByEventID() pagination
- [x] Test CountByEventID() accuracy
- [x] Test GetStats() aggregation
- [x] Test FindDuplicateEmails() detection
- [x] Test DeleteExpired() cleanup

### Documentation
- [x] Document invite model fields
- [x] Document status transitions
- [x] Document repository methods
- [x] Add usage examples

---

## Dependencies

**Depends on:**
- Story 00: Token Generation (for token hash)
- Story 01: Token Hashing (for hash storage)
- Epic 02: Events (foreign key to events table)

**Blocks:**
- Story 04: Individual Invite
- Story 05: Bulk CSV Import
- Story 06: Manual Invite

---

## Testing Strategy

### Unit Tests

1. **Model Validation Tests**
   ```go
   func TestInvite_Validate(t *testing.T) {
       tests := []struct {
           name    string
           invite  *models.Invite
           wantErr bool
       }{
           {
               name: "valid invite",
               invite: &models.Invite{
                   EventID:     1,
                   TokenHash:   "valid-hash",
                   MaxPlusOnes: 2,
                   Status:      models.InviteStatusDraft,
                   ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
               },
               wantErr: false,
           },
           {
               name: "invalid status",
               invite: &models.Invite{
                   EventID:     1,
                   TokenHash:   "valid-hash",
                   MaxPlusOnes: 2,
                   Status:      "invalid",
                   ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
               },
               wantErr: true,
           },
           {
               name: "negative max_plus_ones",
               invite: &models.Invite{
                   EventID:     1,
                   TokenHash:   "valid-hash",
                   MaxPlusOnes: -1,
                   Status:      models.InviteStatusDraft,
                   ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
               },
               wantErr: true,
           },
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := tt.invite.Validate()
               if (err != nil) != tt.wantErr {
                   t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
               }
           })
       }
   }
   ```

2. **Repository CRUD Tests**
   ```go
   func TestInviteRepository_Create(t *testing.T)
   func TestInviteRepository_GetByID(t *testing.T)
   func TestInviteRepository_GetByTokenHash(t *testing.T)
   func TestInviteRepository_Update(t *testing.T)
   func TestInviteRepository_Delete(t *testing.T)
   ```

3. **Repository Query Tests**
   ```go
   func TestInviteRepository_ListByEventID(t *testing.T)
   func TestInviteRepository_ListByEventID_Filters(t *testing.T)
   func TestInviteRepository_ListByEventID_Pagination(t *testing.T)
   func TestInviteRepository_CountByEventID(t *testing.T)
   func TestInviteRepository_GetStats(t *testing.T)
   ```

4. **Duplicate Detection Tests**
   ```go
   func TestInviteRepository_FindDuplicateEmails(t *testing.T) {
       // Setup: Create invites with known emails
       // Test: Check for duplicates
       // Verify: Returns correct duplicate list
   }
   ```

5. **Batch Operation Tests**
   ```go
   func TestInviteRepository_CreateBatch(t *testing.T) {
       // Test successful batch insert
       // Test rollback on error
       // Test partial success handling
   }
   ```

6. **Cleanup Tests**
   ```go
   func TestInviteRepository_DeleteExpired(t *testing.T) {
       // Setup: Create expired and valid invites
       // Test: Delete expired
       // Verify: Only expired deleted, count correct
   }
   ```

---

## Database Schema

```sql
CREATE TABLE invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT,
    email TEXT,
    token_hash TEXT NOT NULL UNIQUE,
    max_plus_ones INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    sent_at TIMESTAMP,
    viewed_at TIMESTAMP,
    unsubscribed BOOLEAN NOT NULL DEFAULT FALSE,
    email_invalid BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    CHECK (status IN ('draft', 'sent', 'viewed', 'responded', 'revoked')),
    CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10)
);

CREATE INDEX idx_invites_event_id ON invites(event_id);
CREATE INDEX idx_invites_token_hash ON invites(token_hash);
CREATE INDEX idx_invites_email ON invites(email);
CREATE INDEX idx_invites_status ON invites(status);
CREATE INDEX idx_invites_expires_at ON invites(expires_at);
```

---

## Validation Rules

### Field Constraints
- `event_id`: Required, must reference existing event
- `name`: Optional, max 100 characters
- `email`: Optional, max 255 characters, valid email format
- `token_hash`: Required, unique, 44 characters
- `max_plus_ones`: Required, 0-10
- `status`: Required, one of defined constants
- `expires_at`: Required, must be future date

### Business Rules
- Email required for sent invites
- Cannot change status from responded to draft
- Cannot change status from revoked
- Expired invites cannot be viewed
- Unsubscribed invites cannot receive emails

---

## Status Transitions

```
draft → sent → viewed → responded
  ↓       ↓       ↓         ↓
  └───────┴───────┴─────→ revoked
```

Valid transitions:
- `draft` → `sent`, `revoked`
- `sent` → `viewed`, `revoked`
- `viewed` → `responded`, `revoked`
- `responded` → (terminal state)
- `revoked` → (terminal state)

---

## Error Handling

| Error Condition | Error Type | HTTP Status | User Message |
|----------------|------------|-------------|--------------|
| Invite not found | `NotFoundError` | 404 | "Invite not found" |
| Duplicate token hash | `ConflictError` | 409 | "Token already exists" |
| Invalid event ID | `ValidationError` | 400 | "Invalid event ID" |
| Invalid status | `ValidationError` | 400 | "Invalid invite status" |
| Expired invite | `ValidationError` | 400 | "Invite has expired" |
| Database error | `InternalError` | 500 | "Database operation failed" |

---

## Performance Considerations

1. **Indexes**
   - Primary key on `id`
   - Unique index on `token_hash` (fast lookups)
   - Index on `event_id` (list by event)
   - Index on `email` (duplicate detection)
   - Index on `status` (filtering)
   - Index on `expires_at` (cleanup queries)

2. **Batch Operations**
   - Use transactions for batch inserts
   - Prepare statements for bulk operations
   - Consider batch size limits (500 max)

3. **Query Optimization**
   - Use `COUNT(*)` for statistics
   - Use `EXISTS` for duplicate checks
   - Limit result sets with pagination

---

## References

- **HLD:** Section 6 (Invite & Guest Access Model)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md) Section 3
- **Database:** [`migrations/sqlite/000001_initial_schema.up.sql`](../../migrations/sqlite/000001_initial_schema.up.sql)
- **Similar:** Event Repository implementation

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Invite model defined and validated
- [x] Repository interface complete
- [x] All repository methods implemented
- [x] Unit tests written and passing (>90% coverage)
- [x] Integration tests with database
- [x] Documentation complete
- [x] Code reviewed
- [x] No linter warnings
- [x] Performance benchmarks acceptable
