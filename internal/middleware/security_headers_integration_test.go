package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeaders_Integration(t *testing.T) {
	tests := []struct {
		name           string
		config         *SecurityHeadersConfig
		path           string
		method         string
		checkHeaders   map[string]string
	}{
		{
			name:   "GET request with default config",
			config: nil,
			path:   "/api/events",
			method: http.MethodGet,
			checkHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"X-Content-Type-Options":    "nosniff",
				"X-Frame-Options":           "DENY",
				"X-XSS-Protection":          "1; mode=block",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
			},
		},
		{
			name:   "POST request with default config",
			config: nil,
			path:   "/api/events",
			method: http.MethodPost,
			checkHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			},
		},
		{
			name: "custom security config",
			config: &SecurityHeadersConfig{
				HSTSMaxAge:            intPtr(63072000),
				HSTSIncludeSubDomains: true,
				HSTSPreload:           true,
				XFrameOptions:         "SAMEORIGIN",
			},
			path:   "/rsvp/token123",
			method: http.MethodGet,
			checkHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",
				"X-Frame-Options":           "SAMEORIGIN",
			},
		},
		{
			name: "static assets",
			config: nil,
			path:   "/static/css/main.css",
			method: http.MethodGet,
			checkHeaders: map[string]string{
				"X-Content-Type-Options": "nosniff",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := SecurityHeaders(tt.config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			for headerName, expectedValue := range tt.checkHeaders {
				if got := rec.Header().Get(headerName); got != expectedValue {
					t.Errorf("header %s mismatch:\nexpected: %q\ngot:      %q", headerName, expectedValue, got)
				}
			}
		})
	}
}

func TestSecurityHeaders_WithFullMiddlewareChain(t *testing.T) {
	config := &SecurityHeadersConfig{
		HSTSMaxAge:            intPtr(31536000),
		HSTSIncludeSubDomains: true,
	}

	handler := Chain(
		Recovery,
		RequestID,
		RealIP,
		SecurityHeaders(config),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	requiredHeaders := []string{
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
		"X-Request-ID",
	}

	for _, header := range requiredHeaders {
		if rec.Header().Get(header) == "" {
			t.Errorf("expected header %s to be set", header)
		}
	}
}

func TestSecurityHeaders_CSPViolationScenarios(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityHeadersConfig
		verify func(t *testing.T, csp string)
	}{
		{
			name: "allows custom script-src without unsafe-inline",
			config: &SecurityHeadersConfig{
				CSPScriptSrc: []string{"'self'"},
			},
			verify: func(t *testing.T, csp string) {
				if !contains(csp, "script-src 'self'") {
					t.Error("CSP should restrict scripts to self only")
				}
				parts := strings.Split(csp, "; ")
				for _, part := range parts {
					if strings.HasPrefix(part, "script-src ") {
						if strings.Contains(part, "'unsafe-inline'") {
							t.Error("CSP should not allow unsafe-inline when explicitly configured")
						}
						if part != "script-src 'self'" {
							t.Errorf("script-src should be exactly 'self', got: %s", part)
						}
					}
				}
			},
		},
		{
			name: "allows data URIs for images",
			config: nil,
			verify: func(t *testing.T, csp string) {
				if !contains(csp, "img-src 'self' data: https:") {
					t.Error("CSP should allow data URIs for images")
				}
			},
		},
		{
			name: "prevents framing",
			config: nil,
			verify: func(t *testing.T, csp string) {
				if !contains(csp, "frame-ancestors 'none'") {
					t.Error("CSP should prevent framing")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := SecurityHeaders(tt.config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			csp := rec.Header().Get("Content-Security-Policy")
			tt.verify(t, csp)
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
