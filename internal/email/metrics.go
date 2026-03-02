package email

import (
	"time"
)

type Metrics interface {
	RecordQueueSize(size int)
	RecordEmailQueued()
	RecordEmailDequeued()
	RecordEmailSent(duration time.Duration)
	RecordEmailFailed(reason string)
	RecordRetryAttempt(attempt int)
	RecordRateLimitHit()
	RecordRateLimitWait(duration time.Duration)
	RecordBatchProcessed(count int, duration time.Duration)
	RecordProcessingError(err error)
}

type noOpMetrics struct{}

func NewNoOpMetrics() Metrics {
	return &noOpMetrics{}
}

func (m *noOpMetrics) RecordQueueSize(size int)                               {}
func (m *noOpMetrics) RecordEmailQueued()                                     {}
func (m *noOpMetrics) RecordEmailDequeued()                                   {}
func (m *noOpMetrics) RecordEmailSent(duration time.Duration)                 {}
func (m *noOpMetrics) RecordEmailFailed(reason string)                        {}
func (m *noOpMetrics) RecordRetryAttempt(attempt int)                         {}
func (m *noOpMetrics) RecordRateLimitHit()                                    {}
func (m *noOpMetrics) RecordRateLimitWait(duration time.Duration)             {}
func (m *noOpMetrics) RecordBatchProcessed(count int, duration time.Duration) {}
func (m *noOpMetrics) RecordProcessingError(err error)                        {}
