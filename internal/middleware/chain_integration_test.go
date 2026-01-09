package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMiddlewareChain_FullStack_Integration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(1*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		realIP := GetRealIP(r.Context())

		if requestID == "" {
			t.Error("expected request ID in context")
		}
		if realIP == "" {
			t.Error("expected real IP in context")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", rec.Body.String())
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "GET") {
		t.Error("expected request to be logged")
	}
	if !strings.Contains(logOutput, requestID) {
		t.Error("expected request ID in log output")
	}
}

func TestMiddlewareChain_PanicRecovery_Integration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(1*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic in full chain")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Internal Server Error") {
		t.Errorf("expected error message, got %s", body)
	}

	if strings.Contains(body, "test panic in full chain") {
		t.Error("panic message leaked to client")
	}
}

func TestMiddlewareChain_Timeout_Integration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(50*time.Millisecond),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504, got %d", rec.Code)
	}
}

func TestMiddlewareChain_OrderVerification_Integration(t *testing.T) {
	var order []string

	trackingMiddleware := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := Chain(
		trackingMiddleware("recovery"),
		trackingMiddleware("requestID"),
		trackingMiddleware("realIP"),
		trackingMiddleware("logging"),
		trackingMiddleware("timeout"),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expected := []string{
		"recovery-before",
		"requestID-before",
		"realIP-before",
		"logging-before",
		"timeout-before",
		"handler",
		"timeout-after",
		"logging-after",
		"realIP-after",
		"requestID-after",
		"recovery-after",
	}

	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}

	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

func TestMiddlewareChain_RealWorldScenario_Integration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(1*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		realIP := GetRealIP(r.Context())

		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Real-IP", realIP)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID in response")
	}

	realIP := rec.Header().Get("X-Real-IP")
	if realIP != "203.0.113.1" {
		t.Errorf("expected real IP 203.0.113.1, got %s", realIP)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "GET") {
		t.Error("expected GET in log")
	}
	if !strings.Contains(logOutput, "/api/events") {
		t.Error("expected path in log")
	}
	if !strings.Contains(logOutput, "200") {
		t.Error("expected status in log")
	}
	if !strings.Contains(logOutput, requestID) {
		t.Error("expected request ID in log")
	}
}

func TestMiddlewareChain_MultipleRequests_Integration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(1*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, rec.Code)
		}
	}

	logOutput := logBuf.String()
	count := strings.Count(logOutput, "GET")
	if count != 10 {
		t.Errorf("expected 10 log entries, got %d", count)
	}
}
