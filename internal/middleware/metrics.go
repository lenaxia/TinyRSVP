package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetricsCollector struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	registry        *prometheus.Registry
}

func NewPrometheusMetrics() *PrometheusMetricsCollector {
	registry := prometheus.NewRegistry()
	return NewPrometheusMetricsWithRegistry(registry)
}

func NewPrometheusMetricsWithRegistry(registerer prometheus.Registerer) *PrometheusMetricsCollector {
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	if registerer != nil {
		registerer.MustRegister(requestsTotal)
		registerer.MustRegister(requestDuration)
	}

	var registry *prometheus.Registry
	if reg, ok := registerer.(*prometheus.Registry); ok {
		registry = reg
	}

	return &PrometheusMetricsCollector{
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
		registry:        registry,
	}
}

func PrometheusMetrics(metrics *PrometheusMetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if metrics == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			
			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			path := normalizePath(r.URL.Path)
			status := strconv.Itoa(rw.status)

			metrics.requestsTotal.WithLabelValues(r.Method, path, status).Inc()
			metrics.requestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
		})
	}
}

var idPattern = regexp.MustCompile(`/\d+(/|$)`)
var tokenPattern = regexp.MustCompile(`/[a-f0-9]{64}(/|$)`)

func normalizePath(path string) string {
	path = idPattern.ReplaceAllString(path, "/{id}$1")
	path = tokenPattern.ReplaceAllString(path, "/{token}$1")
	return path
}

func MetricsHandler(metrics *PrometheusMetricsCollector) http.Handler {
	if metrics == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("# No metrics available\n"))
		})
	}
	
	if metrics.registry != nil {
		return promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{})
	}
	
	return promhttp.Handler()
}
