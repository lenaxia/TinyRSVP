# Worklog 0168: Fix Timeout Middleware Race Condition

**Date:** 2026-07-08  
**Epic:** 10 (Technical Debt)  
**Branch:** `fix/timeout-middleware-race`

---

## Summary

Fixed a data race in the custom `Timeout` middleware that caused intermittent 504 responses and corrupted response bodies. Replaced the hand-rolled goroutine-based implementation with the standard library's `http.TimeoutHandler`, which uses a buffered ResponseWriter to eliminate concurrent writes.

## Root Cause

The original `Timeout` middleware (`internal/middleware/timeout.go`) spawned a goroutine to run the handler, while the main goroutine waited on either `done` or `ctx.Done()`. Both goroutines wrote to the same `http.ResponseWriter`:

```go
go func() {
    // ...
    next.ServeHTTP(w, r.WithContext(ctx))  // handler writes to w
}()

select {
case <-done:
    // ...
case <-ctx.Done():
    w.WriteHeader(http.StatusGatewayTimeout)  // timeout writes to w
    w.Write([]byte("Request timeout"))
}
```

When the handler completed near the timeout boundary, both goroutines wrote to `w` concurrently, causing:
- **Data race** (detected by `go test -race`)
- **Corrupted response bodies** (e.g., `"okRequest timeout"` — both writes concatenated)
- **"superfluous WriteHeader" log warnings** observed during integration testing

## Fix

Replaced the 40-line custom implementation with a one-line delegation to `http.TimeoutHandler`:

```go
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.TimeoutHandler(next, timeout, "Request timeout")
    }
}
```

`http.TimeoutHandler` uses a `bufferedResponseWriter` that buffers the handler's writes in the goroutine and only commits them to the real ResponseWriter if the handler completes before the timeout. This eliminates the concurrent-write race entirely.

## Behavior Changes

| Aspect | Before | After |
|---|---|---|
| Timeout status code | 504 Gateway Timeout | 503 Service Unavailable |
| Zero timeout | Immediate 504 | Immediate 503 (context already expired) |
| Handler context on timeout | Cancelled (handler could observe `ctx.Done()`) | NOT cancelled (handler runs to completion in background) |
| Panic propagation | Propagated to caller | Propagated to caller (Go 1.26+) |

The 503 status code matches the HLD's `SERVICE_UNAVAILABLE (503)` spec (`docs/02_DESIGN/02_REVISED_HLD.md:1729`).

## Tests

- **`TestTimeout_NoConcurrentWriteRace`** (new): runs 100 iterations with a handler that completes just under the timeout, verifying no corrupted responses. Fails with the old implementation, passes with the new one.
- **`TestTimeout_HandlerPanicPropagates`** (new): verifies panics propagate to the caller.
- Updated 4 existing tests to match the new behavior (503 instead of 504, zero-timeout, context cancellation, panic).
- Updated `TestMiddlewareChain_Timeout_Integration` to expect 503.
- All tests pass with `-race` detector.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** 9/9 timeout tests pass, full suite green with `-race`  
**Confidence:** HIGH — race condition eliminated, behavior matches HLD spec
