# User Story: Rate Limiting

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **system**, I want **rate limiting for email sending** so that **SMTP provider limits are not exceeded and the account is not throttled or blocked**.

---

## Acceptance Criteria

- [ ] Rate limiter tracks sends in rolling time window
- [ ] Default limit: 50 emails per minute (configurable)
- [ ] Sliding window algorithm for accurate tracking
- [ ] Thread-safe for concurrent access
- [ ] Blocks when limit reached
- [ ] Returns available slots for batch sizing
- [ ] Resets automatically as window slides
- [ ] All tests pass with timeout
- [ ] Integration tests with queue processor

---

## Technical Details

### Rate Limiter Interface

```go
package email

import (
    "time"
)

type RateLimiter interface {
    // Allow checks if an email can be sent now
    Allow() bool
    
    // AvailableSlots returns how many emails can be sent now
    AvailableSlots() int
    
    // WaitTime returns how long to wait before next send
    WaitTime() time.Duration
    
    // Record records a successful send
    Record()
    
    // Reset clears the rate limiter (for testing)
    Reset()
}

type RateLimiterConfig struct {
    MaxPerMinute int           // Max emails per minute
    WindowSize   time.Duration // Rolling window size
}
```

### Implementation (Sliding Window)

```go
package email

import (
    "sync"
    "time"
)

type rateLimiter struct {
    maxPerMinute int
    windowSize   time.Duration
    timestamps   []time.Time
    mu           sync.Mutex
}

func NewRateLimiter(maxPerMinute int) RateLimiter {
    return &rateLimiter{
        maxPerMinute: maxPerMinute,
        windowSize:   1 * time.Minute,
        timestamps:   make([]time.Time, 0, maxPerMinute),
    }
}

func (r *rateLimiter) Allow() bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.cleanup()
    
    return len(r.timestamps) < r.maxPerMinute
}

func (r *rateLimiter) AvailableSlots() int {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.cleanup()
    
    available := r.maxPerMinute - len(r.timestamps)
    if available < 0 {
        return 0
    }
    return available
}

func (r *rateLimiter) WaitTime() time.Duration {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.cleanup()
    
    if len(r.timestamps) < r.maxPerMinute {
        return 0
    }
    
    // Wait until oldest timestamp expires
    oldest := r.timestamps[0]
    expiresAt := oldest.Add(r.windowSize)
    waitTime := time.Until(expiresAt)
    
    if waitTime < 0 {
        return 0
    }
    
    return waitTime
}

func (r *rateLimiter) Record() {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.cleanup()
    r.timestamps = append(r.timestamps, time.Now())
}

func (r *rateLimiter) Reset() {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.timestamps = r.timestamps[:0]
}

// cleanup removes timestamps outside the rolling window
func (r *rateLimiter) cleanup() {
    now := time.Now()
    cutoff := now.Add(-r.windowSize)
    
    // Find first timestamp within window
    i := 0
    for i < len(r.timestamps) && r.timestamps[i].Before(cutoff) {
        i++
    }
    
    // Remove old timestamps
    if i > 0 {
        r.timestamps = r.timestamps[i:]
    }
}
```

### Alternative: Token Bucket Implementation

```go
package email

import (
    "sync"
    "time"
)

type tokenBucketLimiter struct {
    capacity     int
    tokens       int
    refillRate   int           // tokens per second
    lastRefill   time.Time
    mu           sync.Mutex
}

func NewTokenBucketLimiter(maxPerMinute int) RateLimiter {
    return &tokenBucketLimiter{
        capacity:   maxPerMinute,
        tokens:     maxPerMinute,
        refillRate: maxPerMinute / 60, // per second
        lastRefill: time.Now(),
    }
}

func (t *tokenBucketLimiter) Allow() bool {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    t.refill()
    
    if t.tokens > 0 {
        t.tokens--
        return true
    }
    
    return false
}

func (t *tokenBucketLimiter) AvailableSlots() int {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    t.refill()
    return t.tokens
}

func (t *tokenBucketLimiter) refill() {
    now := time.Now()
    elapsed := now.Sub(t.lastRefill)
    
    tokensToAdd := int(elapsed.Seconds()) * t.refillRate
    if tokensToAdd > 0 {
        t.tokens += tokensToAdd
        if t.tokens > t.capacity {
            t.tokens = t.capacity
        }
        t.lastRefill = now
    }
}
```

---

## Tasks

### Phase 1: Interface Definition (TDD)
- [ ] Define RateLimiter interface
- [ ] Define RateLimiterConfig struct
- [ ] Write test for interface contract
- [ ] Document rate limiting behavior

### Phase 2: Sliding Window Implementation (TDD)
- [ ] Write test for Allow() method
- [ ] Implement Allow() method
- [ ] Write test for AvailableSlots()
- [ ] Implement AvailableSlots()
- [ ] Write test for WaitTime()
- [ ] Implement WaitTime()
- [ ] Write test for Record()
- [ ] Implement Record()

### Phase 3: Window Cleanup (TDD)
- [ ] Write test for timestamp cleanup
- [ ] Implement cleanup() method
- [ ] Write test for window sliding
- [ ] Verify old timestamps removed
- [ ] Write test for memory efficiency
- [ ] Optimize timestamp storage

### Phase 4: Concurrency Safety (TDD)
- [ ] Write test for concurrent Allow()
- [ ] Verify thread safety
- [ ] Write test for concurrent Record()
- [ ] Test race conditions
- [ ] Write test for high concurrency
- [ ] Verify no data races

### Phase 5: Integration Testing
- [ ] Test with queue processor
- [ ] Test rate limit enforcement
- [ ] Test window sliding behavior
- [ ] Test burst handling
- [ ] Test long-running scenarios
- [ ] Verify memory usage

---

## Dependencies

**Depends on:**
- None (standalone component)

**Blocks:**
- Story 02: Email Queue Processor

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Interface defined and documented
- [ ] Implementation complete
- [ ] All unit tests passing
- [ ] Concurrency tests passing
- [ ] Integration tests passing
- [ ] No race conditions
- [ ] Code reviewed
- [ ] Documentation updated

---

## Test Requirements

### Unit Tests

```go
func TestRateLimiter_Allow(t *testing.T) {
    limiter := NewRateLimiter(3) // 3 per minute for testing
    
    // Should allow first 3
    for i := 0; i < 3; i++ {
        if !limiter.Allow() {
            t.Errorf("Allow() = false on attempt %d, want true", i+1)
        }
        limiter.Record()
    }
    
    // Should block 4th
    if limiter.Allow() {
        t.Error("Allow() = true when limit reached, want false")
    }
}

func TestRateLimiter_AvailableSlots(t *testing.T) {
    limiter := NewRateLimiter(5)
    
    if got := limiter.AvailableSlots(); got != 5 {
        t.Errorf("AvailableSlots() = %d, want 5", got)
    }
    
    limiter.Record()
    limiter.Record()
    
    if got := limiter.AvailableSlots(); got != 3 {
        t.Errorf("AvailableSlots() = %d, want 3", got)
    }
}

func TestRateLimiter_WindowSliding(t *testing.T) {
    limiter := NewRateLimiter(2)
    
    // Fill the limit
    limiter.Record()
    limiter.Record()
    
    if limiter.Allow() {
        t.Error("Allow() = true when limit reached")
    }
    
    // Wait for window to slide
    time.Sleep(61 * time.Second)
    
    // Should allow again
    if !limiter.Allow() {
        t.Error("Allow() = false after window expired, want true")
    }
}

func TestRateLimiter_WaitTime(t *testing.T) {
    limiter := NewRateLimiter(1)
    
    limiter.Record()
    
    waitTime := limiter.WaitTime()
    if waitTime <= 0 || waitTime > 60*time.Second {
        t.Errorf("WaitTime() = %v, want between 0 and 60s", waitTime)
    }
}

func TestRateLimiter_Concurrent(t *testing.T) {
    limiter := NewRateLimiter(100)
    
    var wg sync.WaitGroup
    allowed := atomic.NewInt32(0)
    
    // Spawn 200 goroutines trying to send
    for i := 0; i < 200; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if limiter.Allow() {
                limiter.Record()
                allowed.Add(1)
            }
        }()
    }
    
    wg.Wait()
    
    // Should only allow 100
    if got := allowed.Load(); got != 100 {
        t.Errorf("Allowed %d concurrent sends, want 100", got)
    }
}

func TestRateLimiter_Reset(t *testing.T) {
    limiter := NewRateLimiter(2)
    
    limiter.Record()
    limiter.Record()
    
    if limiter.Allow() {
        t.Error("Allow() = true when limit reached")
    }
    
    limiter.Reset()
    
    if !limiter.Allow() {
        t.Error("Allow() = false after reset, want true")
    }
}
```

---

## Rate Limiting Strategies

### Sliding Window (Chosen)
**Pros:**
- Accurate rate limiting
- Smooth distribution
- No burst allowance

**Cons:**
- Memory overhead (stores timestamps)
- Cleanup required

### Token Bucket
**Pros:**
- Constant memory
- Allows bursts
- Simple implementation

**Cons:**
- Less accurate
- Can allow bursts above limit

### Fixed Window
**Pros:**
- Very simple
- Low memory

**Cons:**
- Boundary issues
- Can allow 2x limit at boundaries

---

## Configuration Examples

### Conservative (Gmail)
```go
config := &RateLimiterConfig{
    MaxPerMinute: 50,
    WindowSize:   1 * time.Minute,
}
```

### Aggressive (SendGrid)
```go
config := &RateLimiterConfig{
    MaxPerMinute: 500,
    WindowSize:   1 * time.Minute,
}
```

### Testing
```go
config := &RateLimiterConfig{
    MaxPerMinute: 10,
    WindowSize:   1 * time.Minute,
}
```

---

## SMTP Provider Limits

| Provider   | Limit          | Notes                    |
|------------|----------------|--------------------------|
| Gmail      | 100/day        | Free accounts            |
| Gmail      | 2000/day       | Workspace accounts       |
| SendGrid   | 100/day        | Free tier                |
| SendGrid   | Unlimited      | Paid (rate limited)      |
| Mailgun    | 300/day        | Free tier                |
| AWS SES    | 1/sec          | Sandbox                  |
| AWS SES    | Variable       | Production (request)     |

**Recommendation:** Default to 50/minute (3000/hour) as safe baseline

---

## Monitoring

### Metrics to Track
- Current rate (emails/minute)
- Available slots
- Rate limit hits
- Wait time distribution
- Queue backlog due to rate limiting

### Alerts
- Rate consistently at limit (may need increase)
- Long wait times (queue backing up)
- Frequent rate limit hits

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **Story 02:** [05_STORY_02_email_queue_processor.md](05_STORY_02_email_queue_processor.md)
- **Algorithm:** Sliding Window Rate Limiting
