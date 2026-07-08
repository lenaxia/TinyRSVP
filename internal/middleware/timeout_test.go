package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout_FastRequest(t *testing.T) {
	handler := Timeout(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", rec.Body.String())
	}
}

func TestTimeout_SlowRequest(t *testing.T) {
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// http.TimeoutHandler returns 503 Service Unavailable (not 504) on
	// timeout. This matches the HLD's SERVICE_UNAVAILABLE (503) spec.
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "timeout") && !strings.Contains(body, "Timeout") {
		t.Errorf("expected timeout message in body, got %s", body)
	}
}

func TestTimeout_ContextCancellation(t *testing.T) {
	// http.TimeoutHandler does NOT cancel the handler's context on timeout.
	// Instead, it lets the handler continue running in the background while
	// returning 503 to the client. The handler can observe client disconnects
	// via r.Context().Done(), but NOT the timeout itself.
	//
	// This test verifies the timeout returns 503 and the handler eventually
	// completes (rather than hanging forever).
	handlerCompleted := make(chan struct{})
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		close(handlerCompleted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	// Wait for the handler to complete in the background.
	select {
	case <-handlerCompleted:
		// Handler completed as expected.
	case <-time.After(1 * time.Second):
		t.Error("handler did not complete within 1 second after timeout")
	}
}

func TestTimeout_ZeroTimeout(t *testing.T) {
	// http.TimeoutHandler with 0 duration triggers an immediate timeout
	// (the handler's context is already expired). This matches the
	// standard library's behavior where 0 means "already expired."
	handler := Timeout(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 with zero timeout (immediate), got %d", rec.Code)
	}
}

func TestTimeout_VeryLongTimeout(t *testing.T) {
	handler := Timeout(10 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTimeout_MultipleRequests(t *testing.T) {
	handler := Timeout(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, rec.Code)
		}
	}
}

func TestTimeout_PreservesHandlerBehavior(t *testing.T) {
	handler := Timeout(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	if rec.Body.String() != "created" {
		t.Errorf("expected body 'created', got %s", rec.Body.String())
	}

	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("expected custom header, got %s", rec.Header().Get("X-Custom"))
	}
}

// TestTimeout_NoConcurrentWriteRace verifies that the timeout middleware
// does not produce "superfluous WriteHeader" panics or corrupted responses
// when the handler completes near the timeout boundary. This is a
// regression test for the race condition where both the handler goroutine
// and the timeout goroutine write to the same ResponseWriter.
//
// The test runs many iterations with a handler that completes just before
// the timeout, which maximizes the probability of hitting the race.
func TestTimeout_NoConcurrentWriteRace(t *testing.T) {
	handler := Timeout(10*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep just under the timeout — this creates a tight race window.
		time.Sleep(9 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// The response should be either 200 (handler won the race) or 504
		// (timeout won the race). It should NEVER be a corrupted state
		// like 0 with a non-empty body, or a body containing both "ok"
		// and "Request timeout".
		body := rec.Body.String()
		if rec.Code == 0 {
			t.Errorf("iteration %d: status code 0 (no WriteHeader called), body=%q", i, body)
		}
		if strings.Contains(body, "ok") && strings.Contains(body, "timeout") {
			t.Errorf("iteration %d: body contains both handler output and timeout message: %q", i, body)
		}
		if rec.Code == http.StatusOK && body != "ok" {
			t.Errorf("iteration %d: status 200 but body=%q (expected 'ok')", i, body)
		}
		if rec.Code == http.StatusGatewayTimeout && body != "Request timeout" {
			t.Errorf("iteration %d: status 504 but body=%q (expected 'Request timeout')", i, body)
		}
	}
}

// TestTimeout_HandlerPanicPropagates verifies that panics in the wrapped
// handler are propagated to the caller (http.TimeoutHandler in Go 1.26+
// does not recover panics — it lets them propagate through the goroutine).
func TestTimeout_HandlerPanicPropagates(t *testing.T) {
	handler := Timeout(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic to propagate, but none was recovered")
		}
		if r != "test panic" {
			t.Errorf("expected panic 'test panic', got %v", r)
		}
	}()

	handler.ServeHTTP(rec, req)
}
