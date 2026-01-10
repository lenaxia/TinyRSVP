package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRateLimit_Integration_FullChain(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 5,
		BurstSize:         5,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	chain := Chain(
		Recovery,
		RequestID,
		RealIP,
		RateLimit(limiter, RateLimitConfig{
			AnonymousLimit:      5,
			AuthenticatedLimit:  10,
			AdminLimit:          100,
		}),
	)(handler)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "192.168.1.100")

		chain.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}

		if w.Header().Get("X-Request-ID") == "" {
			t.Error("Expected X-Request-ID header")
		}

		if w.Header().Get("X-RateLimit-Limit") == "" {
			t.Error("Expected X-RateLimit-Limit header")
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "192.168.1.100")

	chain.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 6: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	if w.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header on rate limited response")
	}
}

func TestRateLimit_Integration_WithAuth(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			t.Error("Expected user in context")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated: " + user.Email))
	})

	rateLimitMiddleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:      2,
		AuthenticatedLimit:  5,
		AdminLimit:          100,
	})

	chain := Chain(
		Recovery,
		RequestID,
		RealIP,
		rateLimitMiddleware,
	)(handler)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "192.168.1.101")

		ctx := auth.WithUser(r.Context(), user)
		r = r.WithContext(ctx)

		chain.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}

		if w.Header().Get("X-RateLimit-Limit") != "5" {
			t.Errorf("Request %d: expected limit 5 for authenticated user, got %s", i+1, w.Header().Get("X-RateLimit-Limit"))
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "192.168.1.101")
	ctx := auth.WithUser(r.Context(), user)
	r = r.WithContext(ctx)

	chain.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 6: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_Integration_AdminHigherLimit(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         100,
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rateLimitMiddleware := RateLimit(limiter, RateLimitConfig{
		AnonymousLimit:      2,
		AuthenticatedLimit:  5,
		AdminLimit:          20,
	})

	chain := Chain(
		Recovery,
		RequestID,
		RealIP,
		rateLimitMiddleware,
	)(handler)

	admin := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  models.RoleAdmin,
	}

	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "192.168.1.102")

		ctx := auth.WithUser(r.Context(), admin)
		r = r.WithContext(ctx)

		chain.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}

		if w.Header().Get("X-RateLimit-Limit") != "20" {
			t.Errorf("Request %d: expected limit 20 for admin, got %s", i+1, w.Header().Get("X-RateLimit-Limit"))
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "192.168.1.102")
	ctx := auth.WithUser(r.Context(), admin)
	r = r.WithContext(ctx)

	chain.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 21: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_Integration_MultipleIPs(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,
		BurstSize:         10,
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
			AnonymousLimit:      3,
			AuthenticatedLimit:  10,
			AdminLimit:          100,
		}),
	)(handler)

	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	for _, ip := range ips {
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.Header.Set("X-Real-IP", ip)

			chain.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("IP %s request %d: expected status %d, got %d", ip, i+1, http.StatusOK, w.Code)
			}
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", ip)

		chain.ServeHTTP(w, r)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("IP %s request 4: expected status %d, got %d", ip, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestRateLimit_Integration_WhitelistBypass(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		WhitelistedIPs:    []string{"10.0.0.1", "10.0.0.2"},
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
			AnonymousLimit:      1,
			AuthenticatedLimit:  10,
			AdminLimit:          100,
		}),
	)(handler)

	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "10.0.0.1")

		chain.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Whitelisted IP request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}
}

func TestRateLimit_Integration_BlacklistBlock(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 1000,
		BurstSize:         1000,
		BlacklistedIPs:    []string{"10.0.0.99"},
	})
	defer limiter.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for blacklisted IP")
		w.WriteHeader(http.StatusOK)
	})

	chain := Chain(
		Recovery,
		RequestID,
		RealIP,
		RateLimit(limiter, RateLimitConfig{
			AnonymousLimit:      1000,
			AuthenticatedLimit:  1000,
			AdminLimit:          1000,
		}),
	)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "10.0.0.99")

	chain.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Blacklisted IP: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_Integration_WindowReset(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 2,
		BurstSize:         2,
		WindowDuration:    200 * time.Millisecond,
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
			AnonymousLimit:      2,
			AuthenticatedLimit:  10,
			AdminLimit:          100,
		}),
	)(handler)

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "192.168.1.200")

		chain.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "192.168.1.200")

	chain.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 3: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	time.Sleep(250 * time.Millisecond)

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Real-IP", "192.168.1.200")

	chain.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Request after window reset: expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimit_Integration_Metrics(t *testing.T) {
	limiter := NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 3,
		BurstSize:         3,
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
			AnonymousLimit:      3,
			AuthenticatedLimit:  10,
			AdminLimit:          100,
		}),
	)(handler)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("X-Real-IP", "192.168.1.250")

		chain.ServeHTTP(w, r)
	}

	metrics := limiter.GetMetrics()

	if metrics.TotalRequests != 5 {
		t.Errorf("Expected 5 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.AllowedRequests != 3 {
		t.Errorf("Expected 3 allowed requests, got %d", metrics.AllowedRequests)
	}

	if metrics.DeniedRequests != 2 {
		t.Errorf("Expected 2 denied requests, got %d", metrics.DeniedRequests)
	}

	if metrics.ActiveIPs != 1 {
		t.Errorf("Expected 1 active IP, got %d", metrics.ActiveIPs)
	}
}
