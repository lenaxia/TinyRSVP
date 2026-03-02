package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkSecurityHeaders(b *testing.B) {
	handler := SecurityHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkSecurityHeaders_CustomConfig(b *testing.B) {
	config := &SecurityHeadersConfig{
		HSTSMaxAge:            intPtr(63072000),
		HSTSIncludeSubDomains: true,
		HSTSPreload:           true,
		CSPDefaultSrc:         []string{"'self'", "https://cdn.example.com"},
		CSPScriptSrc:          []string{"'self'"},
		CSPStyleSrc:           []string{"'self'"},
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkCSPReportHandler(b *testing.B) {
	handler := CSPReportHandler(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		body := `{"csp-report":{"document-uri":"https://example.com/page","violated-directive":"script-src 'self'","blocked-uri":"https://evil.com/script.js"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/csp-report", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/csp-report")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkBuildCSP(b *testing.B) {
	config := &SecurityHeadersConfig{
		CSPDefaultSrc:     []string{"'self'", "https://cdn.example.com"},
		CSPScriptSrc:      []string{"'self'", "'nonce-abc123'"},
		CSPStyleSrc:       []string{"'self'", "https://fonts.googleapis.com"},
		CSPImgSrc:         []string{"'self'", "data:", "https:", "blob:"},
		CSPFontSrc:        []string{"'self'", "https://fonts.gstatic.com"},
		CSPConnectSrc:     []string{"'self'", "https://api.example.com"},
		CSPFrameAncestors: []string{"'self'"},
		CSPBaseURI:        []string{"'self'"},
		CSPFormAction:     []string{"'self'", "https://form.example.com"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = buildCSP(config)
	}
}

func BenchmarkBuildHSTS(b *testing.B) {
	config := &SecurityHeadersConfig{
		HSTSMaxAge:            intPtr(31536000),
		HSTSIncludeSubDomains: true,
		HSTSPreload:           true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = buildHSTS(config)
	}
}
