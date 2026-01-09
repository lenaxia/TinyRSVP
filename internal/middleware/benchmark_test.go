package middleware

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkRecovery(b *testing.B) {
	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkRequestID(b *testing.B) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkRealIP(b *testing.B) {
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkLogging(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkTimeout(b *testing.B) {
	handler := Timeout(30 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkChain_Empty(b *testing.B) {
	handler := Chain()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkChain_Full(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(30*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkChain_WithExistingRequestID(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(30*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Request-ID", "existing-id-12345")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkChain_WithXForwardedFor(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		Logging(logger),
		Timeout(30*time.Second),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func TestPerformanceTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	logger := log.New(io.Discard, "", 0)

	tests := []struct {
		name      string
		handler   http.Handler
		maxMicros int64
	}{
		{
			name:      "Recovery",
			handler:   Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			maxMicros: 1,
		},
		{
			name:      "RequestID",
			handler:   RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			maxMicros: 5,
		},
		{
			name:      "RealIP",
			handler:   RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			maxMicros: 10,
		},
		{
			name:      "Logging",
			handler:   Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			maxMicros: 50,
		},
		{
			name:      "Timeout",
			handler:   Timeout(30 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			maxMicros: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"

			var totalDuration time.Duration
			iterations := 1000

			for i := 0; i < iterations; i++ {
				rec := httptest.NewRecorder()
				start := time.Now()
				tt.handler.ServeHTTP(rec, req)
				totalDuration += time.Since(start)
			}

			avgMicros := totalDuration.Microseconds() / int64(iterations)
			t.Logf("%s average: %dµs (target: <%dµs)", tt.name, avgMicros, tt.maxMicros)

			if avgMicros > tt.maxMicros*10 {
				t.Errorf("%s exceeded 10x target: %dµs > %dµs", tt.name, avgMicros, tt.maxMicros*10)
			}
		})
	}
}
