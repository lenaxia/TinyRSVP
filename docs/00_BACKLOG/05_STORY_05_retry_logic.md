# User Story: Retry Logic with Exponential Backoff

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 0.5 days

---

## User Story

As a **system**, I want **retry logic with exponential backoff for failed email sends** so that **transient failures are handled gracefully without overwhelming the SMTP server**.

---

## Acceptance Criteria

- [ ] Exponential backoff calculation implemented
- [ ] Retry schedule: 1min, 5min, 15min, 30min
- [ ] Maximum 4 retry attempts (configurable)
- [ ] Distinguish between transient and permanent errors
- [ ] No retry for permanent errors (5xx SMTP codes)
- [ ] Retry for transient errors (4xx SMTP codes)
- [ ] Jitter added to prevent thundering herd
- [ ] All tests pass with timeout
- [ ] Unit tests for backoff calculation

---

## Technical Details

### Retry Policy Interface

```go
package email

import (
    "time"
)

type RetryPolicy interface {
    // CalculateBackoff returns the delay before next retry
    CalculateBackoff(attempt int) time.Duration
    
    // ShouldRetry determines if an error should be retried
    ShouldRetry(err error, attempt int) bool
    
    // MaxAttempts returns the maximum number of retry attempts
    MaxAttempts() int
}
```

### Implementation

```go
package email

import (
    "errors"
    "math/rand"
    "time"
)

type retryPolicy struct {
    maxAttempts int
    baseDelay   time.Duration
    maxDelay    time.Duration
    jitterPct   float64
}

func NewRetryPolicy(maxAttempts int) RetryPolicy {
    return &retryPolicy{
        maxAttempts: maxAttempts,
        baseDelay:   1 * time.Minute,
        maxDelay:    30 * time.Minute,
        jitterPct:   0.1, // 10% jitter
    }
}

func (p *retryPolicy) CalculateBackoff(attempt int) time.Duration {
    if attempt <= 0 {
        return 0
    }
    
    var delay time.Duration
    
    switch attempt {
    case 1:
        delay = 1 * time.Minute
    case 2:
        delay = 5 * time.Minute
    case 3:
        delay = 15 * time.Minute
    default:
        delay = 30 * time.Minute
    }
    
    // Add jitter to prevent thundering herd
    jitter := time.Duration(float64(delay) * p.jitterPct * (rand.Float64()*2 - 1))
    delay += jitter
    
    // Ensure delay doesn't exceed max
    if delay > p.maxDelay {
        delay = p.maxDelay
    }
    
    return delay
}

func (p *retryPolicy) ShouldRetry(err error, attempt int) bool {
    // Don't retry if max attempts reached
    if attempt >= p.maxAttempts {
        return false
    }
    
    // Don't retry permanent errors
    var permErr *PermanentError
    if errors.As(err, &permErr) {
        return false
    }
    
    // Retry transient errors
    var transErr *TransientError
    if errors.As(err, &transErr) {
        return true
    }
    
    // Retry unknown errors (network issues, timeouts, etc.)
    return true
}

func (p *retryPolicy) MaxAttempts() int {
    return p.maxAttempts
}
```

### Error Classification

```go
package email

import "fmt"

// PermanentError represents an error that should not be retried
type PermanentError struct {
    Err     error
    SMTPCode int
}

func (e *PermanentError) Error() string {
    return fmt.Sprintf("permanent error (SMTP %d): %v", e.SMTPCode, e.Err)
}

func (e *PermanentError) Unwrap() error {
    return e.Err
}

func NewPermanentError(err error, code int) error {
    return &PermanentError{
        Err:     err,
        SMTPCode: code,
    }
}

// TransientError represents an error that can be retried
type TransientError struct {
    Err     error
    SMTPCode int
}

func (e *TransientError) Error() string {
    return fmt.Sprintf("transient error (SMTP %d): %v", e.SMTPCode, e.Err)
}

func (e *TransientError) Unwrap() error {
    return e.Err
}

func NewTransientError(err error, code int) error {
    return &TransientError{
        Err:     err,
        SMTPCode: code,
    }
}

// IsPermanentError checks if an error is permanent
func IsPermanentError(err error) bool {
    var permErr *PermanentError
    return errors.As(err, &permErr)
}

// IsTransientError checks if an error is transient
func IsTransientError(err error) bool {
    var transErr *TransientError
    return errors.As(err, &transErr)
}
```

---

## Tasks

### Phase 1: Interface Definition (TDD)
- [ ] Define RetryPolicy interface
- [ ] Define error types (Permanent, Transient)
- [ ] Write test for interface contract
- [ ] Document retry behavior

### Phase 2: Backoff Calculation (TDD)
- [ ] Write test for attempt 1 (1 minute)
- [ ] Write test for attempt 2 (5 minutes)
- [ ] Write test for attempt 3 (15 minutes)
- [ ] Write test for attempt 4+ (30 minutes)
- [ ] Implement CalculateBackoff method
- [ ] Write test for jitter addition
- [ ] Implement jitter logic
- [ ] Write test for max delay cap
- [ ] Implement max delay cap

### Phase 3: Retry Decision (TDD)
- [ ] Write test for permanent error (no retry)
- [ ] Write test for transient error (retry)
- [ ] Write test for max attempts reached
- [ ] Implement ShouldRetry method
- [ ] Write test for unknown errors
- [ ] Handle unknown errors

### Phase 4: Error Classification (TDD)
- [ ] Write test for SMTP 5xx (permanent)
- [ ] Write test for SMTP 4xx (transient)
- [ ] Implement error classification
- [ ] Write test for error unwrapping
- [ ] Verify error chain works

### Phase 5: Integration Testing
- [ ] Test full retry cycle
- [ ] Test with queue processor
- [ ] Verify backoff timing
- [ ] Test jitter distribution
- [ ] Test max attempts enforcement

---

## Dependencies

**Depends on:**
- Story 03: SMTP Sender (error types)

**Blocks:**
- Story 02: Email Queue Processor (uses retry policy)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Interface defined and documented
- [ ] Implementation complete
- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] Jitter working correctly
- [ ] Code reviewed
- [ ] Documentation updated

---

## Test Requirements

### Unit Tests

```go
func TestRetryPolicy_CalculateBackoff(t *testing.T) {
    policy := NewRetryPolicy(4)
    
    tests := []struct {
        name    string
        attempt int
        want    time.Duration
        delta   time.Duration // Allow for jitter
    }{
        {"attempt 0", 0, 0, 0},
        {"attempt 1", 1, 1 * time.Minute, 6 * time.Second},
        {"attempt 2", 2, 5 * time.Minute, 30 * time.Second},
        {"attempt 3", 3, 15 * time.Minute, 90 * time.Second},
        {"attempt 4", 4, 30 * time.Minute, 3 * time.Minute},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := policy.CalculateBackoff(tt.attempt)
            
            if tt.want == 0 {
                if got != 0 {
                    t.Errorf("CalculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
                }
                return
            }
            
            // Check within delta (for jitter)
            diff := got - tt.want
            if diff < 0 {
                diff = -diff
            }
            
            if diff > tt.delta {
                t.Errorf("CalculateBackoff(%d) = %v, want ~%v (±%v)", tt.attempt, got, tt.want, tt.delta)
            }
        })
    }
}

func TestRetryPolicy_ShouldRetry_PermanentError(t *testing.T) {
    policy := NewRetryPolicy(4)
    
    err := NewPermanentError(fmt.Errorf("mailbox unavailable"), 550)
    
    if policy.ShouldRetry(err, 1) {
        t.Error("ShouldRetry() = true for permanent error, want false")
    }
}

func TestRetryPolicy_ShouldRetry_TransientError(t *testing.T) {
    policy := NewRetryPolicy(4)
    
    err := NewTransientError(fmt.Errorf("mailbox busy"), 450)
    
    if !policy.ShouldRetry(err, 1) {
        t.Error("ShouldRetry() = false for transient error, want true")
    }
}

func TestRetryPolicy_ShouldRetry_MaxAttempts(t *testing.T) {
    policy := NewRetryPolicy(4)
    
    err := NewTransientError(fmt.Errorf("temporary failure"), 450)
    
    if policy.ShouldRetry(err, 4) {
        t.Error("ShouldRetry() = true when max attempts reached, want false")
    }
}

func TestRetryPolicy_Jitter(t *testing.T) {
    policy := NewRetryPolicy(4)
    
    // Run multiple times to verify jitter varies
    delays := make(map[time.Duration]bool)
    
    for i := 0; i < 100; i++ {
        delay := policy.CalculateBackoff(1)
        delays[delay] = true
    }
    
    // Should have multiple different delays due to jitter
    if len(delays) < 10 {
        t.Errorf("Jitter not working, only %d unique delays in 100 attempts", len(delays))
    }
}
```

---

## Retry Schedule

| Attempt | Base Delay | With Jitter (±10%) | Total Time |
|---------|------------|-------------------|------------|
| 1       | 1 min      | 54s - 66s         | ~1 min     |
| 2       | 5 min      | 4.5m - 5.5m       | ~6 min     |
| 3       | 15 min     | 13.5m - 16.5m     | ~21 min    |
| 4       | 30 min     | 27m - 33m         | ~51 min    |

**Total retry window:** ~51 minutes before permanent failure

---

## Jitter Rationale

Jitter prevents the "thundering herd" problem where many emails fail simultaneously (e.g., SMTP server restart) and all retry at exactly the same time, overwhelming the server again.

**Benefits:**
- Spreads retry load over time
- Reduces server spike on recovery
- Improves overall success rate

**Implementation:**
- ±10% random variance
- Applied to each retry delay
- Bounded by max delay

---

## SMTP Error Code Reference

### Permanent Errors (5xx) - No Retry
- **550**: Mailbox unavailable (user doesn't exist)
- **551**: User not local; please try forward path
- **552**: Exceeded storage allocation
- **553**: Mailbox name not allowed
- **554**: Transaction failed

### Transient Errors (4xx) - Retry
- **421**: Service not available, closing transmission channel
- **450**: Mailbox unavailable (e.g., mailbox busy)
- **451**: Local error in processing
- **452**: Insufficient system storage

---

## Configuration

```go
type RetryConfig struct {
    MaxAttempts int           // Default: 4
    BaseDelay   time.Duration // Default: 1 minute
    MaxDelay    time.Duration // Default: 30 minutes
    JitterPct   float64       // Default: 0.1 (10%)
}
```

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **Story 02:** [05_STORY_02_email_queue_processor.md](05_STORY_02_email_queue_processor.md)
- **Story 03:** [05_STORY_03_smtp_sender.md](05_STORY_03_smtp_sender.md)
- **RFC 5321:** SMTP (error codes)
