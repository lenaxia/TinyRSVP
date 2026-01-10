package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
	WindowDuration    time.Duration
	CleanupInterval   time.Duration
	WhitelistedIPs    []string
	BlacklistedIPs    []string
}

type RateLimitConfig struct {
	AnonymousLimit     int
	AuthenticatedLimit int
	AdminLimit         int
}

type RateLimiter struct {
	config          RateLimiterConfig
	limits          map[string]*limit
	mu              sync.RWMutex
	whitelistMap    map[string]bool
	blacklistMap    map[string]bool
	stopCleanup     chan struct{}
	totalRequests   int64
	allowedRequests int64
	deniedRequests  int64
}

type limit struct {
	tokens    int
	resetTime time.Time
}

type RateLimitMetrics struct {
	TotalRequests   int64
	AllowedRequests int64
	DeniedRequests  int64
	ActiveIPs       int
}

func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.WindowDuration == 0 {
		config.WindowDuration = time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	whitelistMap := make(map[string]bool)
	for _, ip := range config.WhitelistedIPs {
		whitelistMap[ip] = true
	}

	blacklistMap := make(map[string]bool)
	for _, ip := range config.BlacklistedIPs {
		blacklistMap[ip] = true
	}

	rl := &RateLimiter{
		config:       config,
		limits:       make(map[string]*limit),
		whitelistMap: whitelistMap,
		blacklistMap: blacklistMap,
		stopCleanup:  make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Allow(ip string, maxTokens int) (allowed bool, remaining int, resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.totalRequests++

	if rl.blacklistMap[ip] {
		rl.deniedRequests++
		return false, 0, time.Now().Add(rl.config.WindowDuration)
	}

	if rl.whitelistMap[ip] {
		rl.allowedRequests++
		return true, maxTokens, time.Now().Add(rl.config.WindowDuration)
	}

	l, exists := rl.limits[ip]
	if !exists {
		l = &limit{
			tokens:    maxTokens,
			resetTime: time.Now().Add(rl.config.WindowDuration),
		}
		rl.limits[ip] = l
	}

	now := time.Now()
	if now.After(l.resetTime) {
		l.tokens = maxTokens
		l.resetTime = now.Add(rl.config.WindowDuration)
	}

	if l.tokens > 0 {
		l.tokens--
		rl.allowedRequests++
		return true, l.tokens, l.resetTime
	}

	rl.deniedRequests++
	return false, 0, l.resetTime
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, l := range rl.limits {
		if now.After(l.resetTime) {
			delete(rl.limits, ip)
		}
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

func (rl *RateLimiter) GetMetrics() RateLimitMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return RateLimitMetrics{
		TotalRequests:   rl.totalRequests,
		AllowedRequests: rl.allowedRequests,
		DeniedRequests:  rl.deniedRequests,
		ActiveIPs:       len(rl.limits),
	}
}

func RateLimit(limiter *RateLimiter, config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetRealIP(r.Context())
			if ip == "" {
				next.ServeHTTP(w, r)
				return
			}

			limit := config.AnonymousLimit

			user, ok := auth.UserFromContext(r.Context())
			if ok {
				if user.Role == models.RoleAdmin {
					limit = config.AdminLimit
				} else {
					limit = config.AuthenticatedLimit
				}
			}

			allowed, remaining, resetTime := limiter.Allow(ip, limit)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				retryAfter := int(time.Until(resetTime).Seconds())
				if retryAfter < 0 {
					retryAfter = 0
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
