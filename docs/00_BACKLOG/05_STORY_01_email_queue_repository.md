# User Story: Email Queue Repository

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **system**, I want **a database repository for email queue operations** so that **emails can be reliably stored, retrieved, and tracked through their lifecycle**.

---

## Acceptance Criteria

- [ ] Repository interface defined with all CRUD operations
- [ ] Create email queue entry
- [ ] Get pending emails (scheduled for now or earlier)
- [ ] Get email by ID
- [ ] Update email status (pending → sending → sent/failed)
- [ ] Increment attempt counter with error message
- [ ] Mark email as sent with timestamp
- [ ] Mark email as failed with error details
- [ ] Mark email as cancelled
- [ ] Query emails by status
- [ ] Query emails by recipient
- [ ] All operations are atomic and thread-safe
- [ ] All tests pass with timeout
- [ ] Integration tests with real database

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
- [ ] Define EmailQueueRepository interface
- [ ] Define EmailQueueStats struct
- [ ] Document all methods with examples
- [ ] Review interface with Epic 05 requirements

### Phase 2: Implementation (TDD)
- [ ] Write test for Create operation
- [ ] Implement Create operation
- [ ] Write test for GetByID operation
- [ ] Implement GetByID operation
- [ ] Write test for GetPending operation
- [ ] Implement GetPending operation
- [ ] Write test for status update operations
- [ ] Implement status update operations
- [ ] Write test for IncrementAttempts
- [ ] Implement IncrementAttempts
- [ ] Write test for Reschedule
- [ ] Implement Reschedule

### Phase 3: Query Operations (TDD)
- [ ] Write test for GetByStatus
- [ ] Implement GetByStatus
- [ ] Write test for GetByRecipient
- [ ] Implement GetByRecipient
- [ ] Write test for GetStats
- [ ] Implement GetStats

### Phase 4: Integration Testing
- [ ] Test with SQLite database
- [ ] Test concurrent access scenarios
- [ ] Test transaction rollback on errors
- [ ] Test optimistic locking
- [ ] Test with large datasets (1000+ emails)
- [ ] Verify index usage with EXPLAIN

### Phase 5: Edge Cases
- [ ] Test with nil context
- [ ] Test with cancelled context
- [ ] Test with invalid IDs
- [ ] Test with missing emails
- [ ] Test status transition validation
- [ ] Test duplicate email prevention

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

- [ ] All acceptance criteria met
- [ ] Interface defined and documented
- [ ] Implementation complete
- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] Edge cases covered
- [ ] Code reviewed
- [ ] Documentation updated

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
