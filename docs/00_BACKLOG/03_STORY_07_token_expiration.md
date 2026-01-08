# User Story: Token Expiration & Cleanup

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As a **system administrator**, I want **automatic token expiration and cleanup** so that **old invites don't accumulate and security is maintained**.

---

## Acceptance Criteria

- [x] Tokens expire 30 days after event date
- [x] Expired tokens cannot be used for RSVP
- [x] Cleanup job removes expired tokens
- [x] Cleanup job runs daily
- [x] Cleanup job logs statistics
- [x] Manual cleanup endpoint for admins
- [x] Expiration date calculated on invite creation
- [x] Expiration check on token validation

---

## Technical Details

### Expiration Calculation

```go
expiresAt := event.StartTime.Add(30 * 24 * time.Hour)
```

### Cleanup Job

```go
func (s *service) CleanupExpiredTokens(ctx context.Context) error {
    before := time.Now()
    count, err := s.repo.DeleteExpired(ctx, before)
    if err != nil {
        return err
    }
    log.Info("Cleaned up expired tokens", "count", count)
    return nil
}
```

### Scheduled Job

```go
func startCleanupJob(service invites.Service) {
    ticker := time.NewTicker(24 * time.Hour)
    go func() {
        for range ticker.C {
            ctx := context.Background()
            if err := service.CleanupExpiredTokens(ctx); err != nil {
                log.Error("Cleanup failed", "error", err)
            }
        }
    }()
}
```

---

## Subtasks

### Implementation
- [x] Add expiration check to token validation
- [x] Implement `DeleteExpired()` in repository
- [x] Implement `CleanupExpiredTokens()` in service
- [x] Add scheduled cleanup job
- [x] Add manual cleanup endpoint for admins
- [x] Add logging for cleanup operations

### Testing
- [x] Test expiration date calculation
- [x] Test expired token rejection
- [x] Test cleanup deletes expired tokens
- [x] Test cleanup preserves valid tokens
- [x] Test cleanup statistics
- [x] Test scheduled job execution

### Documentation
- [x] Document expiration policy
- [x] Document cleanup schedule
- [x] Document manual cleanup endpoint

---

## References

- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md)
- **Story 02:** [03_STORY_02_token_validation.md](03_STORY_02_token_validation.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Expiration logic implemented
- [x] Cleanup job implemented
- [x] Tests passing (>90% coverage)
- [x] Documentation complete
- [x] Code reviewed

---

## Status

**Status:** Complete
**Completed:** 2026-01-07
**Coverage:** 92.0% (invites package)
