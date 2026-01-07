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

- [ ] Tokens expire 30 days after event date
- [ ] Expired tokens cannot be used for RSVP
- [ ] Cleanup job removes expired tokens
- [ ] Cleanup job runs daily
- [ ] Cleanup job logs statistics
- [ ] Manual cleanup endpoint for admins
- [ ] Expiration date calculated on invite creation
- [ ] Expiration check on token validation

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
- [ ] Add expiration check to token validation
- [ ] Implement `DeleteExpired()` in repository
- [ ] Implement `CleanupExpiredTokens()` in service
- [ ] Add scheduled cleanup job
- [ ] Add manual cleanup endpoint for admins
- [ ] Add logging for cleanup operations

### Testing
- [ ] Test expiration date calculation
- [ ] Test expired token rejection
- [ ] Test cleanup deletes expired tokens
- [ ] Test cleanup preserves valid tokens
- [ ] Test cleanup statistics
- [ ] Test scheduled job execution

### Documentation
- [ ] Document expiration policy
- [ ] Document cleanup schedule
- [ ] Document manual cleanup endpoint

---

## References

- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md)
- **Story 02:** [03_STORY_02_token_validation.md](03_STORY_02_token_validation.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Expiration logic implemented
- [ ] Cleanup job implemented
- [ ] Tests passing (>90% coverage)
- [ ] Documentation complete
- [ ] Code reviewed
