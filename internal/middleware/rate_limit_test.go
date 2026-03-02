package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRateLimiter_Allow_SingleRequest(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,
		BurstSize:         10,
	})

	allowed, remaining, resetTime := limiter.Allow("192.168.1.1", 10)

	if !allowed {
		t.Error("Expected first request to be allowed")
	}
	if remaining != 9 {
		t.Errorf("Expected 9 remaining requests, got %d", remaining)
	}
	if resetTime.IsZero() {
		t.Error("Expected non-zero reset time")
	}
}

func TestRateLimiter_Allow_MultipleRequests(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 5,
		BurstSize:         5,
	})

	ip := "192.168.1.2"

	for i := 0; i < 5; i++ {
		allowed, remaining, _ := limiter.Allow(ip, 5)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
		expectedRemaining := 4 - i
		if remaining != expectedRemaining {
			t.Errorf("Request %d: expected %d remaining, got %d", i+1, expectedRemaining, remaining)
		}
	}

	allowed, remaining, _ := limiter.Allow(ip, 5)
	if allowed {
		t.Error("Request 6 should be denied")
	}
	if remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", remaining)
	}
}

func TestRateLimiter_Allow_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 2,
		BurstSize:         2,
	})

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	for i := 0; i < 2; i++ {
		allowed1, _, _ := limiter.Allow(ip1, 2)
		allowed2, _, _ := limiter.Allow(ip2, 2)

		if !allowed1 {
			t.Errorf("IP1 request %d should be allowed", i+1)
		}
		if !allowed2 {
			t.Errorf("IP2 request %d should be allowed", i+1)
		}
	}

	allowed1, _, _ := limiter.Allow(ip1, 2)
	allowed2, _, _ := limiter.Allow(ip2, 2)

	if allowed1 {
		t.Error("IP1 request 3 should be denied")
	}
	if allowed2 {
		t.Error("IP2 request 3 should be denied")
	}
}

func TestRateLimiter_Allow_WindowReset(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 2,
		BurstSize:         2,
		WindowDuration:    100 * time.Millisecond,
	})

	ip := "192.168.1.3"

	allowed1, _, _ := limiter.Allow(ip, 2)
	allowed2, _, _ := limiter.Allow(ip, 2)

	if !allowed1 || !allowed2 {
		t.Error("First two requests should be allowed")
	}

	allowed3, _, _ := limiter.Allow(ip, 2)
	if allowed3 {
		t.Error("Third request should be denied")
	}

	time.Sleep(150 * time.Millisecond)

	allowed4, remaining, _ := limiter.Allow(ip, 2)
	if !allowed4 {
		t.Error("Request after window reset should be allowed")
	}
	if remaining != 1 {
		t.Errorf("Expected 1 remaining after reset, got %d", remaining)
	}
}

func TestRateLimiter_Whitelist(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		WhitelistedIPs:    []string{"10.0.0.1", "10.0.0.2"},
	})

	whitelistedIP := "10.0.0.1"
	normalIP := "192.168.1.1"

	for i := 0; i < 10; i++ {
		allowed, _, _ := limiter.Allow(whitelistedIP, 1)
		if !allowed {
			t.Errorf("Whitelisted IP request %d should always be allowed", i+1)
		}
	}

	allowed1, _, _ := limiter.Allow(normalIP, 1)
	if !allowed1 {
		t.Error("Normal IP first request should be allowed")
	}

	allowed2, _, _ := limiter.Allow(normalIP, 1)
	if allowed2 {
		t.Error("Normal IP second request should be denied")
	}
}

func TestRateLimiter_Blacklist(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
		BlacklistedIPs:    []string{"10.0.0.99", "10.0.0.100"},
	})

	blacklistedIP := "10.0.0.99"
	normalIP := "192.168.1.1"

	allowed1, _, _ := limiter.Allow(blacklistedIP, 100)
	if allowed1 {
		t.Error("Blacklisted IP should never be allowed")
	}

	allowed2, _, _ := limiter.Allow(normalIP, 100)
	if !allowed2 {
		t.Error("Normal IP should be allowed")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,
		BurstSize:         10,
		WindowDuration:    50 * time.Millisecond,
		CleanupInterval:   100 * time.Millisecond,
	})

	limiter.Allow("192.168.1.1", 10)
	limiter.Allow("192.168.1.2", 10)
	limiter.Allow("192.168.1.3", 10)

	limiter.mu.RLock()
	initialCount := len(limiter.limits)
	limiter.mu.RUnlock()

	if initialCount != 3 {
		t.Errorf("Expected 3 entries, got %d", initialCount)
	}

	time.Sleep(200 * time.Millisecond)

	limiter.mu.RLock()
	finalCount := len(limiter.limits)
	limiter.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", finalCount)
	}

	limiter.Stop()
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
	})

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ip := "192.168.1." + strconv.Itoa(id%10)
			allowed, _, _ := limiter.Allow(ip, 100)
			if allowed {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount == 0 {
		t.Error("Expected some requests to succeed")
	}
}

func TestRateLimit_Middleware_Anonymous(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 2,
		BurstSize:         2,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     2,
		AuthenticatedLimit: 10,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/test", nil)
	ctx1 := context.WithValue(r1.Context(), RealIPKey, "192.168.1.1")
	r1 = r1.WithContext(ctx1)

	wrappedHandler.ServeHTTP(w1, r1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request: expected status %d, got %d", http.StatusOK, w1.Code)
	}

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/test", nil)
	ctx2 := context.WithValue(r2.Context(), RealIPKey, "192.168.1.1")
	r2 = r2.WithContext(ctx2)

	wrappedHandler.ServeHTTP(w2, r2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second request: expected status %d, got %d", http.StatusOK, w2.Code)
	}

	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/test", nil)
	ctx3 := context.WithValue(r3.Context(), RealIPKey, "192.168.1.1")
	r3 = r3.WithContext(ctx3)

	wrappedHandler.ServeHTTP(w3, r3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Third request: expected status %d, got %d", http.StatusTooManyRequests, w3.Code)
	}
}

func TestRateLimit_Middleware_Authenticated(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,
		BurstSize:         10,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     2,
		AuthenticatedLimit: 5,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	user := &models.User{
		ID:   1,
		Role: models.RoleEventManager,
	}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.2")
		ctx = auth.WithUser(ctx, user)
		r = r.WithContext(ctx)

		wrappedHandler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.2")
	ctx = auth.WithUser(ctx, user)
	r = r.WithContext(ctx)

	wrappedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 6: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_Middleware_Admin(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     2,
		AuthenticatedLimit: 5,
		AdminLimit:         10,
	})

	wrappedHandler := middleware(handler)

	admin := &models.User{
		ID:   1,
		Role: models.RoleAdmin,
	}

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.3")
		ctx = auth.WithUser(ctx, admin)
		r = r.WithContext(ctx)

		wrappedHandler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.3")
	ctx = auth.WithUser(ctx, admin)
	r = r.WithContext(ctx)

	wrappedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 11: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_Middleware_Headers(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 5,
		BurstSize:         5,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     5,
		AuthenticatedLimit: 10,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "192.168.1.4")
	r = r.WithContext(ctx)

	wrappedHandler.ServeHTTP(w, r)

	if w.Header().Get("X-RateLimit-Limit") != "5" {
		t.Errorf("Expected X-RateLimit-Limit: 5, got %s", w.Header().Get("X-RateLimit-Limit"))
	}

	if w.Header().Get("X-RateLimit-Remaining") != "4" {
		t.Errorf("Expected X-RateLimit-Remaining: 4, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}

	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header to be set")
	}
}

func TestRateLimit_Middleware_RetryAfter(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		WindowDuration:    time.Minute,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     1,
		AuthenticatedLimit: 10,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/test", nil)
	ctx1 := context.WithValue(r1.Context(), RealIPKey, "192.168.1.5")
	r1 = r1.WithContext(ctx1)

	wrappedHandler.ServeHTTP(w1, r1)

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/test", nil)
	ctx2 := context.WithValue(r2.Context(), RealIPKey, "192.168.1.5")
	r2 = r2.WithContext(ctx2)

	wrappedHandler.ServeHTTP(w2, r2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w2.Code)
	}

	retryAfter := w2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("Expected Retry-After header to be set")
	}

	retrySeconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		t.Errorf("Retry-After should be a number: %v", err)
	}

	if retrySeconds <= 0 || retrySeconds > 60 {
		t.Errorf("Retry-After should be between 1 and 60 seconds, got %d", retrySeconds)
	}
}

func TestRateLimit_Middleware_NoRealIP(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,
		BurstSize:         10,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     10,
		AuthenticatedLimit: 20,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	wrappedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d when no real IP, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimit_Middleware_WhitelistedIP(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		WhitelistedIPs:    []string{"10.0.0.1"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     1,
		AuthenticatedLimit: 10,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(r.Context(), RealIPKey, "10.0.0.1")
		r = r.WithContext(ctx)

		wrappedHandler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: whitelisted IP should always get status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}
}

func TestRateLimit_Middleware_BlacklistedIP(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
		BlacklistedIPs:    []string{"10.0.0.99"},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:     100,
		AuthenticatedLimit: 100,
		AdminLimit:         100,
	})

	wrappedHandler := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), RealIPKey, "10.0.0.99")
	r = r.WithContext(ctx)

	wrappedHandler.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Blacklisted IP should get status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimiter_Metrics(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 2,
		BurstSize:         2,
	})

	limiter.Allow("192.168.1.1", 2)
	limiter.Allow("192.168.1.1", 2)
	limiter.Allow("192.168.1.1", 2)

	metrics := limiter.GetMetrics()

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.AllowedRequests != 2 {
		t.Errorf("Expected 2 allowed requests, got %d", metrics.AllowedRequests)
	}

	if metrics.DeniedRequests != 1 {
		t.Errorf("Expected 1 denied request, got %d", metrics.DeniedRequests)
	}

	if metrics.ActiveIPs != 1 {
		t.Errorf("Expected 1 active IP, got %d", metrics.ActiveIPs)
	}
}
