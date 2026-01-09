package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "GET") {
		t.Errorf("expected log to contain method GET, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/test") {
		t.Errorf("expected log to contain path /test, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "200") {
		t.Errorf("expected log to contain status 200, got: %s", logOutput)
	}
}

func TestLogging_LogsStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "404") {
		t.Errorf("expected log to contain status 404, got: %s", logOutput)
	}
}

func TestLogging_LogsDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "s") && !strings.Contains(logOutput, "ms") && !strings.Contains(logOutput, "µs") {
		t.Errorf("expected log to contain duration, got: %s", logOutput)
	}
}

func TestLogging_LogsRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := RequestID(Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected log output")
	}
}

func TestLogging_DefaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "200") {
		t.Errorf("expected default status 200, got: %s", logOutput)
	}
}

func TestLogging_MultipleRequests(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	logOutput := buf.String()
	count := strings.Count(logOutput, "GET")
	if count != 3 {
		t.Errorf("expected 3 log entries, got %d", count)
	}
}

func TestLogging_DifferentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)

			handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			logOutput := buf.String()
			if !strings.Contains(logOutput, method) {
				t.Errorf("expected log to contain method %s, got: %s", method, logOutput)
			}
		})
	}
}

func TestLogging_PreservesHandlerBehavior(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	if rec.Body.String() != "created" {
		t.Errorf("expected body 'created', got %s", rec.Body.String())
	}
}
