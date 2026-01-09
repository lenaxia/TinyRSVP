# User Story: Rate Limiting Middleware

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1 day

---

## User Story

As a **system administrator**, I want **per-IP rate limiting** so that **the application is protected against abuse and DoS attacks**.

---

## Acceptance Criteria

- [ ] Rate limiting per IP address
- [ ] Sliding window algorithm
- [ ] Configurable limits (requests/minute)
- [ ] Different limits for anonymous/authenticated/admin
- [ ] 429 status with Retry-After header
- [ ] Rate limit headers on all responses
- [ ] Whitelist for trusted IPs
- [ ] Blacklist for banned IPs
- [ ] In-memory storage (Redis-free for v0)
- [ ] Rate limit metrics

---

## Technical Details

### Package Location
- `internal/middleware/rate_limit.go` - Rate limiting middleware
- `internal/middleware/rate_limit_test.go` - Tests

### Rate Limiting Implementation

```go
type RateLimiter struct {
    limits map[string]*limit
    mu     sync.RWMutex
}

type limit struct {
    tokens    int
    lastReset time.Time
}

func (rl *RateLimiter) Allow(ip string, maxTokens int) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    l, exists := rl.limits[ip]
    if !exists {
        l = &limit{tokens: maxTokens, lastReset: time.Now()}
        rl.limits[ip] = l
    }
    
    // Reset if window expired
    if time.Since(l.lastReset) > time.Minute {
        l.tokens = maxTokens
        l.lastReset = time.Now()
    }
    
    if l.tokens > 0 {
        l.tokens--
        return true
    }
    
    return false
}
```

---

## Tasks

- [ ] Implement rate limiter
- [ ] Add sliding window algorithm
- [ ] Configure rate limits
- [ ] Add whitelist/blacklist
- [ ] Set rate limit headers
- [ ] Handle 429 responses
- [ ] Add metrics
- [ ] Test rate limiting

---

## Dependencies

**Depends on:** 08_STORY_01_middleware_chain.md

**Blocks:** None

---

## Rate Limits

- Anonymous: 100 requests/minute
- Authenticated: 300 requests/minute
- Admin: 1000 requests/minute

---

## Testing Strategy

```go
func TestRateLimit_Allow(t *testing.T)
func TestRateLimit_Deny(t *testing.T)
func TestRateLimit_Reset(t *testing.T)
func TestRateLimit_DifferentIPs(t *testing.T)
```

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Rate limiting implemented
- [ ] Tests passing
- [ ] Documentation complete
