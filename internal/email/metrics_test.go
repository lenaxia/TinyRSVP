package email

import (
	"errors"
	"testing"
	"time"
)

func TestNoOpMetrics(t *testing.T) {
	metrics := NewNoOpMetrics()

	t.Run("RecordQueueSize", func(t *testing.T) {
		metrics.RecordQueueSize(100)
	})

	t.Run("RecordEmailQueued", func(t *testing.T) {
		metrics.RecordEmailQueued()
	})

	t.Run("RecordEmailDequeued", func(t *testing.T) {
		metrics.RecordEmailDequeued()
	})

	t.Run("RecordEmailSent", func(t *testing.T) {
		metrics.RecordEmailSent(2 * time.Second)
	})

	t.Run("RecordEmailFailed", func(t *testing.T) {
		metrics.RecordEmailFailed("smtp_error")
	})

	t.Run("RecordRetryAttempt", func(t *testing.T) {
		metrics.RecordRetryAttempt(3)
	})

	t.Run("RecordRateLimitHit", func(t *testing.T) {
		metrics.RecordRateLimitHit()
	})

	t.Run("RecordRateLimitWait", func(t *testing.T) {
		metrics.RecordRateLimitWait(5 * time.Second)
	})

	t.Run("RecordBatchProcessed", func(t *testing.T) {
		metrics.RecordBatchProcessed(10, 3*time.Second)
	})

	t.Run("RecordProcessingError", func(t *testing.T) {
		metrics.RecordProcessingError(nil)
	})
}

type mockMetrics struct {
	queueSize        int
	emailsQueued     int
	emailsDequeued   int
	emailsSent       int
	emailsFailed     map[string]int
	retryAttempts    []int
	rateLimitHits    int
	rateLimitWaits   []time.Duration
	batchesProcessed int
	batchSizes       []int
	batchDurations   []time.Duration
	processingErrors int
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		emailsFailed:   make(map[string]int),
		retryAttempts:  []int{},
		rateLimitWaits: []time.Duration{},
		batchSizes:     []int{},
		batchDurations: []time.Duration{},
	}
}

func (m *mockMetrics) RecordQueueSize(size int) {
	m.queueSize = size
}

func (m *mockMetrics) RecordEmailQueued() {
	m.emailsQueued++
}

func (m *mockMetrics) RecordEmailDequeued() {
	m.emailsDequeued++
}

func (m *mockMetrics) RecordEmailSent(duration time.Duration) {
	m.emailsSent++
}

func (m *mockMetrics) RecordEmailFailed(reason string) {
	m.emailsFailed[reason]++
}

func (m *mockMetrics) RecordRetryAttempt(attempt int) {
	m.retryAttempts = append(m.retryAttempts, attempt)
}

func (m *mockMetrics) RecordRateLimitHit() {
	m.rateLimitHits++
}

func (m *mockMetrics) RecordRateLimitWait(duration time.Duration) {
	m.rateLimitWaits = append(m.rateLimitWaits, duration)
}

func (m *mockMetrics) RecordBatchProcessed(count int, duration time.Duration) {
	m.batchesProcessed++
	m.batchSizes = append(m.batchSizes, count)
	m.batchDurations = append(m.batchDurations, duration)
}

func (m *mockMetrics) RecordProcessingError(err error) {
	m.processingErrors++
}

func TestMockMetrics(t *testing.T) {
	metrics := newMockMetrics()

	t.Run("RecordQueueSize", func(t *testing.T) {
		metrics.RecordQueueSize(50)
		if metrics.queueSize != 50 {
			t.Errorf("Expected queue size 50, got %d", metrics.queueSize)
		}
	})

	t.Run("RecordEmailQueued", func(t *testing.T) {
		metrics.RecordEmailQueued()
		metrics.RecordEmailQueued()
		if metrics.emailsQueued != 2 {
			t.Errorf("Expected 2 emails queued, got %d", metrics.emailsQueued)
		}
	})

	t.Run("RecordEmailSent", func(t *testing.T) {
		metrics.RecordEmailSent(2 * time.Second)
		if metrics.emailsSent != 1 {
			t.Errorf("Expected 1 email sent, got %d", metrics.emailsSent)
		}
	})

	t.Run("RecordEmailFailed", func(t *testing.T) {
		metrics.RecordEmailFailed("smtp_error")
		metrics.RecordEmailFailed("smtp_error")
		metrics.RecordEmailFailed("timeout")
		if metrics.emailsFailed["smtp_error"] != 2 {
			t.Errorf("Expected 2 smtp_error failures, got %d", metrics.emailsFailed["smtp_error"])
		}
		if metrics.emailsFailed["timeout"] != 1 {
			t.Errorf("Expected 1 timeout failure, got %d", metrics.emailsFailed["timeout"])
		}
	})

	t.Run("RecordRetryAttempt", func(t *testing.T) {
		metrics.RecordRetryAttempt(1)
		metrics.RecordRetryAttempt(2)
		if len(metrics.retryAttempts) != 2 {
			t.Errorf("Expected 2 retry attempts, got %d", len(metrics.retryAttempts))
		}
	})

	t.Run("RecordRateLimitHit", func(t *testing.T) {
		metrics.RecordRateLimitHit()
		if metrics.rateLimitHits != 1 {
			t.Errorf("Expected 1 rate limit hit, got %d", metrics.rateLimitHits)
		}
	})

	t.Run("RecordBatchProcessed", func(t *testing.T) {
		metrics.RecordBatchProcessed(10, 3*time.Second)
		if metrics.batchesProcessed != 1 {
			t.Errorf("Expected 1 batch processed, got %d", metrics.batchesProcessed)
		}
		if len(metrics.batchSizes) != 1 || metrics.batchSizes[0] != 10 {
			t.Errorf("Expected batch size 10, got %v", metrics.batchSizes)
		}
	})
}

// assertNoPanic runs fn and fails the test (with the given label) if fn panics.
// The noOpMetrics methods are empty no-ops; their contract is "safe to call,
// never panic, no side effects". This helper turns that contract into an
// explicit assertion.
func assertNoPanic(t *testing.T, label string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", label, r)
		}
	}()
	fn()
}

// TestMetricsNoOp exercises every noOpMetrics method across happy, edge and
// unhappy inputs (table-driven). Each call must be a safe no-op.
func TestMetricsNoOp(t *testing.T) {
	metrics := NewNoOpMetrics()
	if metrics == nil {
		t.Fatal("NewNoOpMetrics() returned nil")
	}

	tests := []struct {
		name string
		fn   func()
	}{
		// Happy paths
		{"RecordQueueSize positive", func() { metrics.RecordQueueSize(100) }},
		{"RecordEmailQueued", func() { metrics.RecordEmailQueued() }},
		{"RecordEmailDequeued", func() { metrics.RecordEmailDequeued() }},
		{"RecordEmailSent", func() { metrics.RecordEmailSent(2 * time.Second) }},
		{"RecordEmailFailed", func() { metrics.RecordEmailFailed("smtp_error") }},
		{"RecordRetryAttempt", func() { metrics.RecordRetryAttempt(3) }},
		{"RecordRateLimitHit", func() { metrics.RecordRateLimitHit() }},
		{"RecordRateLimitWait", func() { metrics.RecordRateLimitWait(5 * time.Second) }},
		{"RecordBatchProcessed", func() { metrics.RecordBatchProcessed(10, 3*time.Second) }},
		{"RecordProcessingError non-nil", func() { metrics.RecordProcessingError(errors.New("boom")) }},

		// Edge / unhappy paths
		{"RecordQueueSize zero", func() { metrics.RecordQueueSize(0) }},
		{"RecordQueueSize negative", func() { metrics.RecordQueueSize(-5) }},
		{"RecordEmailSent zero duration", func() { metrics.RecordEmailSent(0) }},
		{"RecordEmailFailed empty reason", func() { metrics.RecordEmailFailed("") }},
		{"RecordRetryAttempt zero", func() { metrics.RecordRetryAttempt(0) }},
		{"RecordRateLimitWait zero", func() { metrics.RecordRateLimitWait(0) }},
		{"RecordBatchProcessed zero", func() { metrics.RecordBatchProcessed(0, 0) }},
		{"RecordProcessingError nil", func() { metrics.RecordProcessingError(nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertNoPanic(t, tt.name, tt.fn)
		})
	}
}

// TestMetricsNoOpImplementsInterface guarantees the no-op satisfies the
// Metrics interface at compile time and that repeated calls are idempotent.
func TestMetricsNoOpImplementsInterface(t *testing.T) {
	var _ Metrics = NewNoOpMetrics()

	metrics := NewNoOpMetrics()

	// Repeated calls must remain safe (idempotent no-op).
	for i := 0; i < 5; i++ {
		metrics.RecordEmailQueued()
		metrics.RecordEmailDequeued()
		metrics.RecordEmailSent(time.Second)
		metrics.RecordRetryAttempt(i)
		metrics.RecordBatchProcessed(i, time.Second)
	}

	assertNoPanic(t, "repeated mixed calls", func() {
		metrics.RecordQueueSize(0)
		metrics.RecordEmailFailed("anything")
		metrics.RecordRateLimitHit()
		metrics.RecordRateLimitWait(0)
		metrics.RecordProcessingError(nil)
	})
}
