package middleware

import (
	"net/http"
	"time"
)

// Timeout wraps the next handler with a request timeout. If the handler
// does not complete within the given duration, a 504 Gateway Timeout is
// returned with the body "Request timeout".
//
// Uses http.TimeoutHandler from the standard library, which buffers the
// handler's writes in a separate goroutine and only commits them to the
// real ResponseWriter if the handler completes before the timeout. This
// avoids the race condition where both the handler goroutine and the
// timeout goroutine write to the same ResponseWriter concurrently.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request timeout")
	}
}
