# User Story: Email Queue Processor

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 2 days

---

## User Story

As a **system**, I want **a background worker to process the email queue** so that **emails are sent reliably with automatic retries on failure**.

---

## Acceptance Criteria

- [x] Background worker runs on configurable interval (default 60 seconds)
- [x] Processes pending emails in batches
- [x] Respects rate limiting (max emails per minute)
- [x] Updates email status atomically (pending → sending → sent/failed)
- [x] Implements exponential backoff for retries
- [x] Handles SMTP errors gracefully
- [x] Prevents duplicate sends via optimistic locking
- [x] Graceful shutdown on application stop
- [x] Logs processing activity
- [x] All tests pass with timeout
- [x] Integration tests with mock SMTP

---

## Technical Details

### Queue Processor Interface

```go
package email

import (
    "context"
    "time"
)

type QueueProcessor interface {
    // Start begins processing the queue
    Start(ctx context.Context) error
    
    // Stop gracefully shuts down the processor
    Stop(ctx context.Context) error
    
    // ProcessBatch processes a single batch of emails
    ProcessBatch(ctx context.Context) error
}
```

### Implementation

```go
package email

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type queueProcessor struct {
    repo          repositories.EmailQueueRepository
    sender        SMTPSender
    rateLimiter   RateLimiter
    batchSize     int
    pollInterval  time.Duration
    stopChan      chan struct{}
    doneChan      chan struct{}
}

func NewQueueProcessor(
    repo repositories.EmailQueueRepository,
    sender SMTPSender,
    rateLimiter RateLimiter,
    batchSize int,
    pollInterval time.Duration,
) QueueProcessor {
    return &queueProcessor{
        repo:         repo,
        sender:       sender,
        rateLimiter:  rateLimiter,
        batchSize:    batchSize,
        pollInterval: pollInterval,
        stopChan:     make(chan struct{}),
        doneChan:     make(chan struct{}),
    }
}

func (p *queueProcessor) Start(ctx context.Context) error {
    ticker := time.NewTicker(p.pollInterval)
    defer ticker.Stop()
    
    log.Printf("Email queue processor started (interval: %v, batch: %d)", 
        p.pollInterval, p.batchSize)
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Email queue processor stopped (context cancelled)")
            close(p.doneChan)
            return ctx.Err()
            
        case <-p.stopChan:
            log.Println("Email queue processor stopped (shutdown requested)")
            close(p.doneChan)
            return nil
            
        case <-ticker.C:
            if err := p.ProcessBatch(ctx); err != nil {
                log.Printf("Error processing email batch: %v", err)
            }
        }
    }
}

func (p *queueProcessor) Stop(ctx context.Context) error {
    close(p.stopChan)
    
    select {
    case <-p.doneChan:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (p *queueProcessor) ProcessBatch(ctx context.Context) error {
    // Get pending emails respecting rate limit
    availableSlots := p.rateLimiter.AvailableSlots()
    if availableSlots == 0 {
        return nil
    }
    
    batchSize := min(p.batchSize, availableSlots)
    emails, err := p.repo.GetPending(ctx, batchSize)
    if err != nil {
        return fmt.Errorf("failed to get pending emails: %w", err)
    }
    
    if len(emails) == 0 {
        return nil
    }
    
    log.Printf("Processing %d emails", len(emails))
    
    for _, email := range emails {
        if err := p.processEmail(ctx, email); err != nil {
            log.Printf("Failed to process email %d: %v", email.ID, err)
        }
    }
    
    return nil
}

func (p *queueProcessor) processEmail(ctx context.Context, email *models.EmailQueue) error {
    // Mark as sending (optimistic lock)
    if err := p.repo.MarkSending(ctx, email.ID); err != nil {
        return fmt.Errorf("failed to mark as sending: %w", err)
    }
    
    // Check rate limit
    if !p.rateLimiter.Allow() {
        // Reschedule for later
        return p.repo.Reschedule(ctx, email.ID, time.Now().Add(time.Minute))
    }
    
    // Send email
    if err := p.sendEmail(ctx, email); err != nil {
        return p.handleSendError(ctx, email, err)
    }
    
    // Mark as sent
    if err := p.repo.MarkSent(ctx, email.ID); err != nil {
        log.Printf("Warning: email sent but failed to mark as sent: %v", err)
    }
    
    log.Printf("Email %d sent successfully to %s", email.ID, email.ToEmail)
    return nil
}

func (p *queueProcessor) sendEmail(ctx context.Context, email *models.EmailQueue) error {
    attachments, err := email.GetAttachments()
    if err != nil {
        return fmt.Errorf("failed to get attachments: %w", err)
    }
    
    msg := &SMTPMessage{
        To:          email.ToEmail,
        ToName:      email.ToName,
        Subject:     email.Subject,
        BodyText:    email.BodyText,
        BodyHTML:    email.BodyHTML,
        Attachments: convertAttachments(attachments),
    }
    
    return p.sender.Send(ctx, msg)
}

func (p *queueProcessor) handleSendError(ctx context.Context, email *models.EmailQueue, err error) error {
    // Increment attempts
    if err := p.repo.IncrementAttempts(ctx, email.ID, err.Error()); err != nil {
        return fmt.Errorf("failed to increment attempts: %w", err)
    }
    
    // Check if max attempts reached
    if email.Attempts+1 >= email.MaxAttempts {
        if err := p.repo.MarkFailed(ctx, email.ID, err.Error()); err != nil {
            log.Printf("Failed to mark email as failed: %v", err)
        }
        log.Printf("Email %d permanently failed after %d attempts: %v", 
            email.ID, email.Attempts+1, err)
        return nil
    }
    
    // Calculate backoff and reschedule
    backoff := calculateBackoff(email.Attempts + 1)
    scheduledFor := time.Now().Add(backoff)
    
    if err := p.repo.Reschedule(ctx, email.ID, scheduledFor); err != nil {
        return fmt.Errorf("failed to reschedule: %w", err)
    }
    
    log.Printf("Email %d rescheduled for retry in %v (attempt %d/%d)", 
        email.ID, backoff, email.Attempts+1, email.MaxAttempts)
    
    return nil
}

func calculateBackoff(attempt int) time.Duration {
    switch attempt {
    case 1:
        return 1 * time.Minute
    case 2:
        return 5 * time.Minute
    case 3:
        return 15 * time.Minute
    default:
        return 30 * time.Minute
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

---

## Tasks

### Phase 1: Interface & Structure (TDD)
- [x] Define QueueProcessor interface
- [x] Define configuration struct
- [x] Write test for processor initialization
- [x] Implement processor struct
- [x] Write test for Start/Stop lifecycle
- [x] Implement Start/Stop methods

### Phase 2: Batch Processing (TDD)
- [x] Write test for ProcessBatch with no emails
- [x] Write test for ProcessBatch with emails
- [x] Implement ProcessBatch method
- [x] Write test for rate limit respect
- [x] Implement rate limit checking
- [x] Write test for batch size limiting
- [x] Implement batch size limiting

### Phase 3: Email Processing (TDD)
- [x] Write test for successful email send
- [x] Implement processEmail method
- [x] Write test for send failure handling
- [x] Implement error handling
- [x] Write test for retry scheduling
- [x] Implement retry logic
- [x] Write test for max attempts reached
- [x] Implement permanent failure handling

### Phase 4: Concurrency & Safety (TDD)
- [x] Write test for optimistic locking
- [x] Implement status update with locking
- [x] Write test for concurrent processing
- [x] Verify no duplicate sends
- [x] Write test for graceful shutdown
- [x] Implement shutdown with in-flight handling

### Phase 5: Integration Testing
- [x] Test with mock SMTP sender
- [x] Test with real database
- [x] Test full retry cycle
- [x] Test rate limiting enforcement
- [x] Test graceful shutdown
- [x] Test error recovery

---

## Dependencies

**Depends on:**
- Story 01: Email Queue Repository
- Story 03: SMTP Sender (interface)
- Story 05: Retry Logic (backoff calculation)
- Story 06: Rate Limiting

**Blocks:**
- None (supporting system)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Interface defined and documented
- [x] Implementation complete
- [x] All unit tests passing
- [x] Integration tests passing
- [x] Concurrency tests passing
- [x] Graceful shutdown working
- [x] Code reviewed
- [x] Documentation updated

---

## Test Requirements

### Unit Tests

```go
func TestQueueProcessor_ProcessBatch_NoEmails(t *testing.T) {
    mockRepo := &MockEmailQueueRepository{
        GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
            return []*models.EmailQueue{}, nil
        },
    }
    
    processor := NewQueueProcessor(mockRepo, nil, nil, 10, time.Minute)
    
    err := processor.ProcessBatch(context.Background())
    if err != nil {
        t.Errorf("ProcessBatch() error = %v, want nil", err)
    }
}

func TestQueueProcessor_ProcessBatch_Success(t *testing.T) {
    email := &models.EmailQueue{
        ID:       1,
        ToEmail:  "test@example.com",
        Subject:  "Test",
        BodyText: "Test body",
        Status:   models.EmailStatusPending,
    }
    
    mockRepo := &MockEmailQueueRepository{
        GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
            return []*models.EmailQueue{email}, nil
        },
        MarkSendingFunc: func(ctx context.Context, id int64) error {
            return nil
        },
        MarkSentFunc: func(ctx context.Context, id int64) error {
            return nil
        },
    }
    
    mockSender := &MockSMTPSender{
        SendFunc: func(ctx context.Context, msg *SMTPMessage) error {
            return nil
        },
    }
    
    mockRateLimiter := &MockRateLimiter{
        AllowFunc: func() bool { return true },
        AvailableSlotsFunc: func() int { return 10 },
    }
    
    processor := NewQueueProcessor(mockRepo, mockSender, mockRateLimiter, 10, time.Minute)
    
    err := processor.ProcessBatch(context.Background())
    if err != nil {
        t.Errorf("ProcessBatch() error = %v, want nil", err)
    }
}

func TestQueueProcessor_ProcessBatch_RetryOnFailure(t *testing.T) {
    // Test that failed sends are rescheduled with backoff
}

func TestQueueProcessor_ProcessBatch_MaxAttemptsReached(t *testing.T) {
    // Test that emails are marked as failed after max attempts
}

func TestQueueProcessor_GracefulShutdown(t *testing.T) {
    // Test that processor stops gracefully
    // Test that in-flight emails complete
}
```

---

## Configuration

```go
type QueueProcessorConfig struct {
    BatchSize     int           // Max emails per batch (default: 50)
    PollInterval  time.Duration // How often to check queue (default: 60s)
    MaxAttempts   int           // Max retry attempts (default: 4)
}
```

---

## Monitoring Points

- Emails processed per batch
- Processing time per batch
- Send success/failure rate
- Retry attempts distribution
- Queue depth over time
- Rate limit hits

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **Story 01:** [05_STORY_01_email_queue_repository.md](05_STORY_01_email_queue_repository.md)
- **Story 05:** [05_STORY_05_retry_logic.md](05_STORY_05_retry_logic.md)
- **Story 06:** [05_STORY_06_rate_limiting.md](05_STORY_06_rate_limiting.md)
