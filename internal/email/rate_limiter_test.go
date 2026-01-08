package email

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiter_Allow_WithinLimit(t *testing.T) {
	limiter := NewRateLimiter(3)

	for i := 0; i < 3; i++ {
		if !limiter.Allow() {
			t.Errorf("Allow() = false on attempt %d, want true", i+1)
		}
		limiter.Record()
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	limiter := NewRateLimiter(3)

	for i := 0; i < 3; i++ {
		limiter.Allow()
		limiter.Record()
	}

	if limiter.Allow() {
		t.Error("Allow() = true when limit reached, want false")
	}
}

func TestRateLimiter_AvailableSlots_Initial(t *testing.T) {
	limiter := NewRateLimiter(5)

	if got := limiter.AvailableSlots(); got != 5 {
		t.Errorf("AvailableSlots() = %d, want 5", got)
	}
}

func TestRateLimiter_AvailableSlots_AfterRecording(t *testing.T) {
	limiter := NewRateLimiter(5)

	limiter.Record()
	limiter.Record()

	if got := limiter.AvailableSlots(); got != 3 {
		t.Errorf("AvailableSlots() = %d, want 3", got)
	}
}

func TestRateLimiter_AvailableSlots_AtLimit(t *testing.T) {
	limiter := NewRateLimiter(2)

	limiter.Record()
	limiter.Record()

	if got := limiter.AvailableSlots(); got != 0 {
		t.Errorf("AvailableSlots() = %d, want 0", got)
	}
}

func TestRateLimiter_WaitTime_BelowLimit(t *testing.T) {
	limiter := NewRateLimiter(5)

	limiter.Record()

	waitTime := limiter.WaitTime()
	if waitTime != 0 {
		t.Errorf("WaitTime() = %v, want 0", waitTime)
	}
}

func TestRateLimiter_WaitTime_AtLimit(t *testing.T) {
	limiter := NewRateLimiter(1)

	limiter.Record()

	waitTime := limiter.WaitTime()
	if waitTime <= 0 || waitTime > 60*time.Second {
		t.Errorf("WaitTime() = %v, want between 0 and 60s", waitTime)
	}
}

func TestRateLimiter_WindowSliding(t *testing.T) {
	limiter := &rateLimiter{
		maxPerMinute: 2,
		windowSize:   2 * time.Second,
		timestamps:   make([]time.Time, 0, 2),
	}

	limiter.Record()
	limiter.Record()

	if limiter.Allow() {
		t.Error("Allow() = true when limit reached, want false")
	}

	time.Sleep(2100 * time.Millisecond)

	if !limiter.Allow() {
		t.Error("Allow() = false after window expired, want true")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter(2)

	limiter.Record()
	limiter.Record()

	if limiter.Allow() {
		t.Error("Allow() = true when limit reached, want false")
	}

	limiter.Reset()

	if !limiter.Allow() {
		t.Error("Allow() = false after reset, want true")
	}
}

func TestRateLimiter_Concurrent_ExactLimit(t *testing.T) {
	limiter := NewRateLimiter(100)

	var wg sync.WaitGroup
	var allowed atomic.Int32
	var mu sync.Mutex

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			if limiter.Allow() {
				limiter.Record()
				allowed.Add(1)
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if got := allowed.Load(); got != 100 {
		t.Errorf("Allowed %d concurrent sends, want 100", got)
	}
}

func TestRateLimiter_Concurrent_AvailableSlots(t *testing.T) {
	limiter := NewRateLimiter(50)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slots := limiter.AvailableSlots()
			if slots < 0 {
				t.Errorf("AvailableSlots() = %d, want >= 0", slots)
			}
		}()
	}

	wg.Wait()
}

func TestRateLimiter_Concurrent_Record(t *testing.T) {
	limiter := NewRateLimiter(100)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Record()
		}()
	}

	wg.Wait()

	if got := limiter.AvailableSlots(); got != 50 {
		t.Errorf("AvailableSlots() = %d, want 50", got)
	}
}

func TestRateLimiter_Concurrent_WaitTime(t *testing.T) {
	limiter := NewRateLimiter(10)

	for i := 0; i < 10; i++ {
		limiter.Record()
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			waitTime := limiter.WaitTime()
			if waitTime < 0 {
				t.Errorf("WaitTime() = %v, want >= 0", waitTime)
			}
		}()
	}

	wg.Wait()
}

func TestRateLimiter_MultipleRecords_CorrectCount(t *testing.T) {
	limiter := NewRateLimiter(10)

	for i := 0; i < 5; i++ {
		limiter.Record()
	}

	if got := limiter.AvailableSlots(); got != 5 {
		t.Errorf("AvailableSlots() = %d, want 5", got)
	}

	for i := 0; i < 5; i++ {
		limiter.Record()
	}

	if got := limiter.AvailableSlots(); got != 0 {
		t.Errorf("AvailableSlots() = %d, want 0", got)
	}
}

func TestRateLimiter_AllowWithoutRecord_DoesNotConsume(t *testing.T) {
	limiter := NewRateLimiter(3)

	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			t.Errorf("Allow() = false on attempt %d without Record(), want true", i+1)
		}
	}

	if got := limiter.AvailableSlots(); got != 3 {
		t.Errorf("AvailableSlots() = %d after Allow() without Record(), want 3", got)
	}
}

func TestRateLimiter_RecordWithoutAllow_ConsumesSlot(t *testing.T) {
	limiter := NewRateLimiter(5)

	limiter.Record()

	if got := limiter.AvailableSlots(); got != 4 {
		t.Errorf("AvailableSlots() = %d after Record() without Allow(), want 4", got)
	}
}

func TestRateLimiter_ZeroLimit(t *testing.T) {
	limiter := NewRateLimiter(0)

	if limiter.Allow() {
		t.Error("Allow() = true with zero limit, want false")
	}

	if got := limiter.AvailableSlots(); got != 0 {
		t.Errorf("AvailableSlots() = %d with zero limit, want 0", got)
	}
}

func TestRateLimiter_LargeLimit(t *testing.T) {
	limiter := NewRateLimiter(10000)

	if got := limiter.AvailableSlots(); got != 10000 {
		t.Errorf("AvailableSlots() = %d, want 10000", got)
	}

	for i := 0; i < 1000; i++ {
		limiter.Record()
	}

	if got := limiter.AvailableSlots(); got != 9000 {
		t.Errorf("AvailableSlots() = %d, want 9000", got)
	}
}

func TestRateLimiter_ResetClearsAllTimestamps(t *testing.T) {
	limiter := NewRateLimiter(5)

	for i := 0; i < 5; i++ {
		limiter.Record()
	}

	limiter.Reset()

	if got := limiter.AvailableSlots(); got != 5 {
		t.Errorf("AvailableSlots() = %d after reset, want 5", got)
	}

	if !limiter.Allow() {
		t.Error("Allow() = false after reset, want true")
	}
}

func TestRateLimiter_WaitTime_MultipleTimestamps(t *testing.T) {
	limiter := NewRateLimiter(3)

	limiter.Record()
	time.Sleep(100 * time.Millisecond)
	limiter.Record()
	time.Sleep(100 * time.Millisecond)
	limiter.Record()

	waitTime := limiter.WaitTime()
	if waitTime <= 0 || waitTime > 60*time.Second {
		t.Errorf("WaitTime() = %v, want between 0 and 60s", waitTime)
	}
}
