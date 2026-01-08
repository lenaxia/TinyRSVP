package email

import (
	"sync"
	"time"
)

type rateLimiter struct {
	maxPerMinute int
	windowSize   time.Duration
	timestamps   []time.Time
	mu           sync.Mutex
}

func NewRateLimiter(maxPerMinute int) RateLimiter {
	return &rateLimiter{
		maxPerMinute: maxPerMinute,
		windowSize:   1 * time.Minute,
		timestamps:   make([]time.Time, 0, maxPerMinute),
	}
}

func (r *rateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanup()

	return len(r.timestamps) < r.maxPerMinute
}

func (r *rateLimiter) AvailableSlots() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanup()

	available := r.maxPerMinute - len(r.timestamps)
	if available < 0 {
		return 0
	}
	return available
}

func (r *rateLimiter) WaitTime() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanup()

	if len(r.timestamps) < r.maxPerMinute {
		return 0
	}

	oldest := r.timestamps[0]
	expiresAt := oldest.Add(r.windowSize)
	waitTime := time.Until(expiresAt)

	if waitTime < 0 {
		return 0
	}

	return waitTime
}

func (r *rateLimiter) Record() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanup()
	r.timestamps = append(r.timestamps, time.Now())
}

func (r *rateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.timestamps = r.timestamps[:0]
}

func (r *rateLimiter) cleanup() {
	now := time.Now()
	cutoff := now.Add(-r.windowSize)

	i := 0
	for i < len(r.timestamps) && r.timestamps[i].Before(cutoff) {
		i++
	}

	if i > 0 {
		r.timestamps = r.timestamps[i:]
	}
}
