package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1000,
		BurstSize:         1000,
	})
	defer limiter.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("192.168.1.1", 1000)
	}
}

func BenchmarkRateLimiter_Allow_DifferentIPs(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1000,
		BurstSize:         1000,
	})
	defer limiter.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := "192.168.1." + string(rune(i%255))
		limiter.Allow(ip, 1000)
	}
}

func BenchmarkRateLimiter_Allow_Whitelist(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1000,
		BurstSize:         1000,
		WhitelistedIPs:    []string{"10.0.0.1"},
	})
	defer limiter.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("10.0.0.1", 1000)
	}
}

func BenchmarkRateLimiter_Allow_Blacklist(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1000,
		BurstSize:         1000,
		BlacklistedIPs:    []string{"10.0.0.99"},
	})
	defer limiter.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("10.0.0.99", 1000)
	}
}

func BenchmarkRateLimit_Middleware_Anonymous(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10000,
		BurstSize:         10000,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:      10000,
		AuthenticatedLimit:  10000,
		AdminLimit:          10000,
	})

	wrappedHandler := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.1")
	r = r.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		wrappedHandler.ServeHTTP(w, r)
	}
}

func BenchmarkRateLimit_Middleware_Authenticated(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10000,
		BurstSize:         10000,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:      10000,
		AuthenticatedLimit:  10000,
		AdminLimit:          10000,
	})

	wrappedHandler := middleware(handler)

	user := &models.User{
		ID:   1,
		Role: models.RoleEventManager,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.1")
	ctx = auth.WithUser(ctx, user)
	r = r.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		wrappedHandler.ServeHTTP(w, r)
	}
}

func BenchmarkRateLimit_Middleware_Admin(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10000,
		BurstSize:         10000,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:      10000,
		AuthenticatedLimit:  10000,
		AdminLimit:          10000,
	})

	wrappedHandler := middleware(handler)

	admin := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.1")
	ctx = auth.WithUser(ctx, admin)
	r = r.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		wrappedHandler.ServeHTTP(w, r)
	}
}

func BenchmarkRateLimit_Middleware_FullChain(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10000,
		BurstSize:         10000,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chain := Chain(
		Recovery,
		RequestID,
		RealIP,
		RateLimit(limiter, RateLimitConfig{
			AnonymousLimit:      10000,
			AuthenticatedLimit:  10000,
			AdminLimit:          10000,
		}),
	)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		chain.ServeHTTP(w, r)
	}
}

func BenchmarkRateLimit_Middleware_RateLimited(b *testing.B) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:      1,
		AuthenticatedLimit:  10,
		AdminLimit:          100,
	})

	wrappedHandler := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.1")
	r = r.WithContext(ctx)

	wrappedHandler.ServeHTTP(w, r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		wrappedHandler.ServeHTTP(w, r)
	}
}
