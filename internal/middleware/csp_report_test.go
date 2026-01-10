package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCSPReportHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
		expectLog      bool
	}{
		{
			name:           "valid CSP report",
			method:         http.MethodPost,
			contentType:    "application/csp-report",
			body:           `{"csp-report":{"document-uri":"https://example.com/page","violated-directive":"script-src 'self'","blocked-uri":"https://evil.com/script.js"}}`,
			expectedStatus: http.StatusNoContent,
			expectLog:      true,
		},
		{
			name:           "valid JSON content type",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"csp-report":{"document-uri":"https://example.com/page","violated-directive":"style-src 'self'"}}`,
			expectedStatus: http.StatusNoContent,
			expectLog:      true,
		},
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			contentType:    "application/csp-report",
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectLog:      false,
		},
		{
			name:           "invalid content type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "invalid",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectLog:      false,
		},
		{
			name:           "empty body",
			method:         http.MethodPost,
			contentType:    "application/csp-report",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectLog:      false,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			contentType:    "application/csp-report",
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectLog:      false,
		},
		{
			name:           "body too large",
			method:         http.MethodPost,
			contentType:    "application/csp-report",
			body:           strings.Repeat("x", 11000),
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectLog:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logger := log.New(&logBuf, "", 0)

			handler := CSPReportHandler(logger)

			req := httptest.NewRequest(tt.method, "/api/csp-report", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			logOutput := logBuf.String()
			if tt.expectLog && logOutput == "" {
				t.Error("expected CSP violation to be logged")
			}
			if !tt.expectLog && strings.Contains(logOutput, "CSP violation") {
				t.Error("did not expect CSP violation to be logged")
			}

			if tt.expectLog && !strings.Contains(logOutput, "CSP violation") {
				t.Errorf("expected log to contain 'CSP violation', got: %s", logOutput)
			}
		})
	}
}

func TestCSPReportHandler_ParsesReport(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := CSPReportHandler(logger)

	report := map[string]interface{}{
		"csp-report": map[string]interface{}{
			"document-uri":       "https://example.com/page",
			"violated-directive": "script-src 'self'",
			"blocked-uri":        "https://evil.com/script.js",
			"source-file":        "https://example.com/app.js",
			"line-number":        42,
			"column-number":      15,
		},
	}

	body, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/csp-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/csp-report")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	logOutput := logBuf.String()
	expectedFields := []string{
		"document-uri",
		"violated-directive",
		"blocked-uri",
		"https://example.com/page",
		"script-src 'self'",
		"https://evil.com/script.js",
	}

	for _, field := range expectedFields {
		if !strings.Contains(logOutput, field) {
			t.Errorf("expected log to contain %q, got: %s", field, logOutput)
		}
	}
}

func TestCSPReportHandler_NilLogger(t *testing.T) {
	handler := CSPReportHandler(nil)

	report := map[string]interface{}{
		"csp-report": map[string]interface{}{
			"violated-directive": "script-src 'self'",
		},
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest(http.MethodPost, "/api/csp-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/csp-report")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestCSPReportHandler_Integration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
	)(CSPReportHandler(logger))

	report := map[string]interface{}{
		"csp-report": map[string]interface{}{
			"document-uri":       "https://example.com/page",
			"violated-directive": "img-src 'self'",
			"blocked-uri":        "https://untrusted.com/image.jpg",
		},
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest(http.MethodPost, "/api/csp-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/csp-report")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "CSP violation") {
		t.Errorf("expected CSP violation to be logged, got: %s", logOutput)
	}

	if !strings.Contains(logOutput, "img-src 'self'") {
		t.Errorf("expected violated directive in log, got: %s", logOutput)
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestCSPReportHandler_ConcurrentRequests(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	handler := CSPReportHandler(logger)

	report := map[string]interface{}{
		"csp-report": map[string]interface{}{
			"violated-directive": "script-src 'self'",
		},
	}

	body, _ := json.Marshal(report)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodPost, "/api/csp-report", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/csp-report")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCSPReportHandler_DefaultLogger(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handler := CSPReportHandler(nil)

	report := map[string]interface{}{
		"csp-report": map[string]interface{}{
			"violated-directive": "default-src 'self'",
		},
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest(http.MethodPost, "/api/csp-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/csp-report")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
