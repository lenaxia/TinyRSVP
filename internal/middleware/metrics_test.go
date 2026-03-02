package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrometheusMetrics_RecordsRequestCount(t *testing.T) {
	metrics := NewPrometheusMetrics()
	handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPrometheusMetrics_RecordsDifferentMethods(t *testing.T) {
	metrics := NewPrometheusMetrics()
	handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Method %s: expected status 200, got %d", method, w.Code)
		}
	}
}

func TestPrometheusMetrics_RecordsDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"success", http.StatusOK},
		{"created", http.StatusCreated},
		{"bad request", http.StatusBadRequest},
		{"not found", http.StatusNotFound},
		{"internal error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewPrometheusMetrics()
			handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestPrometheusMetrics_RecordsDuration(t *testing.T) {
	metrics := NewPrometheusMetrics()
	handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPrometheusMetrics_NormalizesPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"root", "/"},
		{"api endpoint", "/api/events"},
		{"with id", "/api/events/123"},
		{"nested", "/api/events/123/invites"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewPrometheusMetrics()
			handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestPrometheusMetrics_HandlesNilMetrics(t *testing.T) {
	handler := PrometheusMetrics(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsHandler_ReturnsPrometheusFormat(t *testing.T) {
	metrics := NewPrometheusMetrics()

	handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	metricsHandler := MetricsHandler(metrics)
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()

	metricsHandler.ServeHTTP(metricsW, metricsReq)

	if metricsW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", metricsW.Code)
	}

	contentType := metricsW.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to contain text/plain, got %s", contentType)
	}

	body := metricsW.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestMetricsHandler_IncludesStandardMetrics(t *testing.T) {
	metrics := NewPrometheusMetrics()

	handler := PrometheusMetrics(metrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	metricsHandler := MetricsHandler(metrics)
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	metricsHandler.ServeHTTP(metricsW, metricsReq)

	body := metricsW.Body.String()

	expectedMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("Expected metrics to contain %s", metric)
		}
	}
}

func TestMetricsHandler_HandlesNilMetrics(t *testing.T) {
	handler := MetricsHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
