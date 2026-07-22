package email

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

type prometheusMetrics struct {
	queueSize        prometheus.Gauge
	emailsTotal      *prometheus.CounterVec
	emailsSent       *prometheus.HistogramVec
	retryAttempts    *prometheus.CounterVec
	rateLimitHits    prometheus.Counter
	rateLimitWait    *prometheus.HistogramVec
	batchProcessed   *prometheus.CounterVec
	batchDuration    *prometheus.HistogramVec
	processingErrors prometheus.Counter
}

func NewPrometheusMetrics(reg prometheus.Registerer) Metrics {
	m := &prometheusMetrics{
		queueSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tinyrsvp_email_queue_size",
			Help: "Current number of emails in the queue",
		}),
		emailsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tinyrsvp_emails_total",
			Help: "Total number of emails by status",
		}, []string{"status"}),
		emailsSent: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tinyrsvp_email_send_duration_seconds",
			Help:    "Time spent sending an email",
			Buckets: prometheus.DefBuckets,
		}, []string{}),
		retryAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tinyrsvp_email_retry_attempts_total",
			Help: "Total email retry attempts",
		}, []string{"attempt"}),
		rateLimitHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "tinyrsvp_email_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		}),
		rateLimitWait: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tinyrsvp_email_rate_limit_wait_seconds",
			Help:    "Time spent waiting due to rate limiting",
			Buckets: prometheus.DefBuckets,
		}, []string{}),
		batchProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tinyrsvp_email_batch_processed_total",
			Help: "Total emails processed in batches",
		}, []string{"result"}),
		batchDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tinyrsvp_email_batch_duration_seconds",
			Help:    "Time spent processing a batch",
			Buckets: prometheus.DefBuckets,
		}, []string{}),
		processingErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "tinyrsvp_email_processing_errors_total",
			Help: "Total email processing errors",
		}),
	}

	if reg != nil {
		reg.MustRegister(
			m.queueSize,
			m.emailsTotal,
			m.emailsSent,
			m.retryAttempts,
			m.rateLimitHits,
			m.rateLimitWait,
			m.batchProcessed,
			m.batchDuration,
			m.processingErrors,
		)
	}

	return m
}

func (m *prometheusMetrics) RecordQueueSize(size int) {
	m.queueSize.Set(float64(size))
}

func (m *prometheusMetrics) RecordEmailQueued() {
	m.emailsTotal.WithLabelValues("queued").Inc()
}

func (m *prometheusMetrics) RecordEmailDequeued() {
	m.emailsTotal.WithLabelValues("dequeued").Inc()
}

func (m *prometheusMetrics) RecordEmailSent(duration time.Duration) {
	m.emailsTotal.WithLabelValues("sent").Inc()
	m.emailsSent.WithLabelValues().Observe(duration.Seconds())
}

func (m *prometheusMetrics) RecordEmailFailed(reason string) {
	m.emailsTotal.WithLabelValues("failed_" + reason).Inc()
}

func (m *prometheusMetrics) RecordRetryAttempt(attempt int) {
	m.retryAttempts.WithLabelValues(time.Now().Format("2006-01-02")).Inc()
}

func (m *prometheusMetrics) RecordRateLimitHit() {
	m.rateLimitHits.Inc()
}

func (m *prometheusMetrics) RecordRateLimitWait(duration time.Duration) {
	m.rateLimitWait.WithLabelValues().Observe(duration.Seconds())
}

func (m *prometheusMetrics) RecordBatchProcessed(count int, duration time.Duration) {
	m.batchProcessed.WithLabelValues("processed").Add(float64(count))
	m.batchDuration.WithLabelValues().Observe(duration.Seconds())
}

func (m *prometheusMetrics) RecordProcessingError(err error) {
	m.processingErrors.Inc()
}
