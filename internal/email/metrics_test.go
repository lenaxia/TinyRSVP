package email

import (
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
