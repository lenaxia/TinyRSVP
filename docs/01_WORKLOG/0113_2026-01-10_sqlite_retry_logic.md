# Worklog: SQLite Automatic Retry Logic

**Date:** 2026-01-10  
**Story:** 10_STORY_07 - SQLite Locking Error Mitigation  
**Status:** Complete

## Summary

Implemented automatic retry logic with exponential backoff to handle SQLite locking errors (`SQLITE_LOCKED`, `SQLITE_BUSY`). This prevents transient database lock errors from causing request failures.

## Problem

SQLite can experience locking errors in concurrent scenarios:
- `attempt to write a readonly database` (permissions issue)
- `SQLITE_LOCKED` - Database is locked by another connection
- `SQLITE_BUSY` - Database file is busy
- `SQLITE_LOCKED_SHAREDCACHE` - Shared cache lock contention

These errors are often transient and can be resolved by retrying the operation after a brief delay.

## Solution

### 1. Created Retry Wrapper (`internal/db/retry.go`)

**Key Components:**

#### RetryConfig
```go
type RetryConfig struct {
    MaxAttempts       int           // Maximum retry attempts
    InitialDelay      time.Duration // Starting delay
    MaxDelay          time.Duration // Maximum delay cap
    BackoffMultiplier float64       // Exponential multiplier
}
```

**Default Configuration:**
- Max Attempts: 5
- Initial Delay: 10ms
- Max Delay: 1 second
- Backoff Multiplier: 2.0 (exponential)

**Retry Schedule:**
1. Attempt 1: Immediate
2. Attempt 2: 10ms delay
3. Attempt 3: 20ms delay
4. Attempt 4: 40ms delay
5. Attempt 5: 80ms delay

#### RetryableDatabase
Wraps the base `Database` interface and automatically retries operations that fail with SQLite locking errors.

**Features:**
- Detects SQLite-specific lock errors
- Exponential backoff with configurable parameters
- Context-aware (respects cancellation)
- Only retries lock errors (other errors fail immediately)
- Transparent wrapper (implements same `Database` interface)

### 2. Updated Database Initialization (`cmd/server/main.go`)

Changed from:
```go
database, err := db.NewDatabase(cfg)
```

To:
```go
baseDatabase, err := db.NewDatabase(cfg)
database := db.NewRetryableDatabase(baseDatabase, db.DefaultRetryConfig)
```

### 3. Existing SQLite Optimizations (Already in Place)

The database DSN already includes important optimizations:
```go
file:path?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=1
```

- **WAL Mode** (`_journal_mode=WAL`): Write-Ahead Logging for better concurrency
- **Busy Timeout** (`_busy_timeout=5000`): Wait up to 5 seconds before returning SQLITE_BUSY
- **Shared Cache** (`cache=shared`): Allows multiple connections to share cache
- **Foreign Keys** (`_foreign_keys=1`): Enforce referential integrity

## How It Works

### Retry Flow
```
1. Execute operation
2. Success? → Return result
3. Lock error? → Calculate backoff delay
4. Wait (with context cancellation support)
5. Retry (up to MaxAttempts)
6. Max attempts exceeded? → Return error
```

### Error Detection
```go
func isSQLiteLockedError(err error) bool {
    var sqliteErr sqlite3.Error
    if errors.As(err, &sqliteErr) {
        return sqliteErr.Code == sqlite3.ErrLocked ||
               sqliteErr.Code == sqlite3.ErrBusy ||
               sqliteErr.ExtendedCode == sqlite3.ErrLockedSharedCache
    }
    return false
}
```

### Exponential Backoff
```go
delay = initialDelay * (multiplier ^ attempt)
delay = min(delay, maxDelay)
```

Example with defaults:
- Attempt 0: 10ms * (2^0) = 10ms
- Attempt 1: 10ms * (2^1) = 20ms
- Attempt 2: 10ms * (2^2) = 40ms
- Attempt 3: 10ms * (2^3) = 80ms
- Attempt 4: 10ms * (2^4) = 160ms

## Benefits

1. **Automatic Recovery**: Transient lock errors resolve themselves
2. **No Code Changes**: Transparent to existing code
3. **Configurable**: Retry parameters can be tuned
4. **Context-Aware**: Respects request timeouts and cancellations
5. **Efficient**: Exponential backoff prevents thundering herd
6. **Selective**: Only retries lock errors, not other failures

## Testing

### Manual Testing
1. Run server with concurrent requests
2. Verify no `SQLITE_LOCKED` errors in logs
3. Monitor retry attempts (add logging if needed)

### Load Testing
```bash
# Concurrent requests to trigger lock contention
for i in {1..50}; do
  curl http://localhost:8080/ &
done
wait
```

## Configuration Options

To customize retry behavior, modify `db.DefaultRetryConfig`:

```go
// More aggressive retries
customConfig := db.RetryConfig{
    MaxAttempts:       10,
    InitialDelay:      5 * time.Millisecond,
    MaxDelay:          2 * time.Second,
    BackoffMultiplier: 2.0,
}
database := db.NewRetryableDatabase(baseDatabase, customConfig)
```

## Limitations

1. **QueryRow**: Cannot be retried (error deferred until Scan())
   - Use `Query()` instead for retry support
2. **Transaction Retries**: Entire transaction is retried, not individual statements
3. **Max Delay Cap**: Prevents indefinite waiting

## Alternative Solutions Considered

### 1. Increase Busy Timeout Only
- **Pros**: Simple, already implemented
- **Cons**: Fixed timeout, no exponential backoff
- **Decision**: Use both (5s timeout + retry logic)

### 2. Connection Pooling
- **Pros**: Reduces contention
- **Cons**: Already implemented (`MaxOpenConns`)
- **Decision**: Keep existing pool settings

### 3. PostgreSQL Migration
- **Pros**: Better concurrency
- **Cons**: More complex deployment, not homelab-friendly
- **Decision**: Keep SQLite as default, support Postgres as option

## Files Created
- `internal/db/retry.go` - Retry logic implementation

## Files Modified
- `cmd/server/main.go` - Wrap database with retry logic

## Next Steps

1. **Monitor Production**: Watch for retry attempts in logs
2. **Tune Parameters**: Adjust retry config based on real-world usage
3. **Add Metrics**: Track retry attempts and success rates
4. **Consider Logging**: Add structured logging for retry attempts

## References
- SQLite WAL Mode: https://www.sqlite.org/wal.html
- SQLite Busy Handling: https://www.sqlite.org/c3ref/busy_handler.html
- Exponential Backoff: https://en.wikipedia.org/wiki/Exponential_backoff
