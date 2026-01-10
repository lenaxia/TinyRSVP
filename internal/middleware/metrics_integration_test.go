package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsEndpoint_Integration(t *testing.T) {
	metrics := NewPrometheusMetrics()
	
	mux := http.NewServeMux()
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	mux.Handle("/test", PrometheusMetrics(metrics)(testHandler))
	mux.Handle("/metrics", MetricsHandler(metrics))
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	mux.ServeHTTP(metricsW, metricsReq)
	
	if metricsW.Code != http.StatusOK {
		t.Fatalf("Expected metrics status 200, got %d", metricsW.Code)
	}
	
	body := metricsW.Body.String()
	
	if !strings.Contains(body, "http_requests_total") {
		t.Error("Expected metrics to contain http_requests_total")
	}
	
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Error("Expected metrics to contain http_request_duration_seconds")
	}
	
	if !strings.Contains(body, `method="GET"`) {
		t.Error("Expected metrics to contain method label")
	}
	
	if !strings.Contains(body, `status="200"`) {
		t.Error("Expected metrics to contain status label")
	}
}

func TestMetricsEndpoint_MultipleRequests(t *testing.T) {
	metrics := NewPrometheusMetrics()
	
	mux := http.NewServeMux()
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	mux.Handle("/test", PrometheusMetrics(metrics)(testHandler))
	mux.Handle("/metrics", MetricsHandler(metrics))
	
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Fatalf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}
	
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	mux.ServeHTTP(metricsW, metricsReq)
	
	body := metricsW.Body.String()
	
	if !strings.Contains(body, "http_requests_total") {
		t.Error("Expected metrics to contain http_requests_total")
	}
}

func TestMetricsEndpoint_DifferentStatusCodes(t *testing.T) {
	metrics := NewPrometheusMetrics()
	
	mux := http.NewServeMux()
	
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	
	mux.Handle("/success", PrometheusMetrics(metrics)(successHandler))
	mux.Handle("/error", PrometheusMetrics(metrics)(errorHandler))
	mux.Handle("/metrics", MetricsHandler(metrics))
	
	successReq := httptest.NewRequest(http.MethodGet, "/success", nil)
	successW := httptest.NewRecorder()
	mux.ServeHTTP(successW, successReq)
	
	errorReq := httptest.NewRequest(http.MethodGet, "/error", nil)
	errorW := httptest.NewRecorder()
	mux.ServeHTTP(errorW, errorReq)
	
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	mux.ServeHTTP(metricsW, metricsReq)
	
	body := metricsW.Body.String()
	
	if !strings.Contains(body, `status="200"`) {
		t.Error("Expected metrics to contain status 200")
	}
	
	if !strings.Contains(body, `status="500"`) {
		t.Error("Expected metrics to contain status 500")
	}
}

func TestMetricsEndpoint_PathNormalization(t *testing.T) {
	metrics := NewPrometheusMetrics()
	
	mux := http.NewServeMux()
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	mux.Handle("/api/events/123", PrometheusMetrics(metrics)(testHandler))
	mux.Handle("/api/events/456", PrometheusMetrics(metrics)(testHandler))
	mux.Handle("/metrics", MetricsHandler(metrics))
	
	req1 := httptest.NewRequest(http.MethodGet, "/api/events/123", nil)
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, req1)
	
	req2 := httptest.NewRequest(http.MethodGet, "/api/events/456", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsW := httptest.NewRecorder()
	mux.ServeHTTP(metricsW, metricsReq)
	
	body := metricsW.Body.String()
	
	if !strings.Contains(body, `path="/api/events/{id}"`) {
		t.Error("Expected metrics to normalize path with {id}")
	}
}
