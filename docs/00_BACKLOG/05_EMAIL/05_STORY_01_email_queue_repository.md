# User Story: Email Queue Repository

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day

---

## User Story

As a **system**, I want **a database repository for email queue operations** so that **emails can be reliably stored, retrieved, and tracked through their lifecycle**.

---

## Acceptance Criteria

- [x] Repository interface defined with all CRUD operations
- [x] Create email queue entry
- [x] Get pending emails (scheduled for now or earlier)
- [x] Get email by ID
- [x] Update email status (pending → sending → sent/failed)
- [x] Increment attempt counter with error message
- [x] Mark email as sent with timestamp
- [x] Mark email as failed with error details
- [x] Mark email as cancelled
- [x] Query emails by status
- [x] Query emails by recipient
- [x] All operations are atomic and thread-safe
- [x] All tests pass with timeout
- [x] Integration tests with real database

---

## Technical Details

### Repository Interface

```go
package repositories

import (
    "context"
    "time"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type EmailQueueRepository interface {
    // Create adds a new email to the queue
    Create(ctx context.Context, email *models.EmailQueue) error
    
    // GetByID retrieves an email by its ID
    GetByID(ctx context.Context, id int64) (*models.EmailQueue, error)
    
    // GetPending retrieves emails ready to send (status=pending, scheduled_for <= now)
    // Limited by maxCount for rate limiting
    GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error)
    
    // GetByStatus retrieves emails with specific status
    GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error)
    
    // GetByRecipient retrieves emails for a specific recipient
    GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error)
    
    // UpdateStatus updates the email status
    UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error
    
    // IncrementAttempts increments attempt counter and records error
    IncrementAttempts(ctx context.Context, id int64, errorMsg string) error
    
    // MarkSending marks email as currently being sent
    MarkSending(ctx context.Context, id int64) error
    
    // MarkSent marks email as successfully sent
    MarkSent(ctx context.Context, id int64) error
    
    // MarkFailed marks email as permanently failed
    MarkFailed(ctx context.Context, id int64, errorMsg string) error
    
    // MarkCancelled marks email as cancelled
    MarkCancelled(ctx context.Context, id int64) error
    
    // Reschedule updates the scheduled_for time for retry
    Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error
    
    // GetStats returns queue statistics
    GetStats(ctx context.Context) (*EmailQueueStats, error)
}

type EmailQueueStats struct {
    PendingCount  int
    SendingCount  int
    SentCount     int
    FailedCount   int
    TotalCount    int
}
```

### Implementation Notes

1. **Atomic Operations**: Use database transactions for status updates
2. **Optimistic Locking**: Prevent concurrent processing of same email
3. **Indexes**: Add indexes on status, scheduled_for, to_email for performance
4. **Error Handling**: Store last error message for debugging
5. **Timestamps**: Track created_at, updated_at, sent_at automatically

### Database Queries

```sql
-- Get pending emails ready to send
SELECT * FROM email_queue 
WHERE status = 'pending' 
  AND scheduled_for <= ?
ORDER BY scheduled_for ASC, priority DESC
LIMIT ?;

-- Mark as sending (with optimistic lock)
UPDATE email_queue 
SET status = 'sending', updated_at = ?
WHERE id = ? AND status = 'pending';

-- Increment attempts
UPDATE email_queue
SET attempts = attempts + 1,
    last_error = ?,
    updated_at = ?
WHERE id = ?;

-- Mark as sent
UPDATE email_queue
SET status = 'sent',
    sent_at = ?,
    updated_at = ?
WHERE id = ?;

-- Get queue statistics
SELECT 
    status,
    COUNT(*) as count
FROM email_queue
GROUP BY status;
```

---

## Tasks

### Phase 1: Interface Definition (TDD)
- [x] Define EmailQueueRepository interface
- [x] Define EmailQueueStats struct
- [x] Document all methods with examples
- [x] Review interface with Epic 05 requirements

### Phase 2: Implementation (TDD)
- [x] Write test for Create operation
- [x] Implement Create operation
- [x] Write test for GetByID operation
- [x] Implement GetByID operation
- [x] Write test for GetPending operation
- [x] Implement GetPending operation
- [x] Write test for status update operations
- [x] Implement status update operations
- [x] Write test for IncrementAttempts
- [x] Implement IncrementAttempts
- [x] Write test for Reschedule
- [x] Implement Reschedule

### Phase 3: Query Operations (TDD)
- [x] Write test for GetByStatus
- [x] Implement GetByStatus
- [x] Write test for GetByRecipient
- [x] Implement GetByRecipient
- [x] Write test for GetStats
- [x] Implement GetStats

### Phase 4: Integration Testing
- [x] Test with SQLite database
- [x] Test concurrent access scenarios
- [x] Test transaction rollback on errors
- [x] Test optimistic locking
- [x] Test with large datasets (1000+ emails)
- [x] Verify index usage with EXPLAIN

### Phase 5: Edge Cases
- [x] Test with nil context
- [x] Test with cancelled context
- [x] Test with invalid IDs
- [x] Test with missing emails
- [x] Test status transition validation
- [x] Test duplicate email prevention

---

## Dependencies

**Depends on:**
- Epic 04 Story 11: EmailQueue model already exists
- Database migrations for email_queue table

**Blocks:**
- Story 02: Email Queue Processor
- Story 05: Retry Logic

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Interface defined and documented
- [x] Implementation complete
- [x] All unit tests passing
- [x] Integration tests passing
- [x] Edge cases covered
- [x] Code reviewed
- [x] Documentation updated

---

## Test Requirements

### Unit Tests
```go
func TestEmailQueueRepository_Create(t *testing.T) {
    tests := []struct {
        name    string
        email   *models.EmailQueue
        wantErr bool
    }{
        {
            name: "valid email",
            email: &models.EmailQueue{
                ToEmail:      "test@example.com",
                Subject:      "Test",
                BodyText:     "Test body",
                Status:       models.EmailStatusPending,
                MaxAttempts:  4,
                ScheduledFor: time.Now(),
            },
            wantErr: false,
        },
        {
            name: "missing required fields",
            email: &models.EmailQueue{
                ToEmail: "test@example.com",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(ctx, tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestEmailQueueRepository_GetPending(t *testing.T) {
    // Setup: Create emails with different scheduled times
    // Test: Retrieve only emails scheduled for now or earlier
    // Verify: Correct emails returned in priority order
}

func TestEmailQueueRepository_ConcurrentAccess(t *testing.T) {
    // Test concurrent GetPending calls
    // Verify no email is returned twice
    // Verify optimistic locking works
}
```

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **Model:** [`internal/models/email_queue.go`](../../internal/models/email_queue.go)
- **Database:** [`migrations/sqlite/000001_initial_schema.up.sql`](../../migrations/sqlite/000001_initial_schema.up.sql)
