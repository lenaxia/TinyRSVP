package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func intPtr(i int) *int {
	return &i
}

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name         string
		config       *SecurityHeadersConfig
		expectedHSTS string
		expectedCSP  string
		expectedXCTO string
		expectedXFO  string
		expectedXXSS string
		expectedRP   string
		expectedPP   string
	}{
		{
			name:         "default headers",
			config:       nil,
			expectedHSTS: "max-age=31536000; includeSubDomains",
			expectedCSP:  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			expectedXCTO: "nosniff",
			expectedXFO:  "DENY",
			expectedXXSS: "1; mode=block",
			expectedRP:   "strict-origin-when-cross-origin",
			expectedPP:   "geolocation=(), microphone=(), camera=()",
		},
		{
			name: "custom HSTS",
			config: &SecurityHeadersConfig{
				HSTSMaxAge:            intPtr(63072000),
				HSTSIncludeSubDomains: true,
				HSTSPreload:           true,
			},
			expectedHSTS: "max-age=63072000; includeSubDomains; preload",
			expectedCSP:  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			expectedXCTO: "nosniff",
			expectedXFO:  "DENY",
			expectedXXSS: "1; mode=block",
			expectedRP:   "strict-origin-when-cross-origin",
			expectedPP:   "geolocation=(), microphone=(), camera=()",
		},
		{
			name: "custom CSP",
			config: &SecurityHeadersConfig{
				CSPDefaultSrc:     []string{"'self'", "https://trusted.com"},
				CSPScriptSrc:      []string{"'self'"},
				CSPStyleSrc:       []string{"'self'"},
				CSPImgSrc:         []string{"'self'", "data:", "https:"},
				CSPFontSrc:        []string{"'self'"},
				CSPConnectSrc:     []string{"'self'"},
				CSPFrameAncestors: []string{"'none'"},
				CSPBaseURI:        []string{"'self'"},
				CSPFormAction:     []string{"'self'"},
			},
			expectedHSTS: "max-age=31536000; includeSubDomains",
			expectedCSP:  "default-src 'self' https://trusted.com; script-src 'self'; style-src 'self'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			expectedXCTO: "nosniff",
			expectedXFO:  "DENY",
			expectedXXSS: "1; mode=block",
			expectedRP:   "strict-origin-when-cross-origin",
			expectedPP:   "geolocation=(), microphone=(), camera=()",
		},
		{
			name: "custom all headers",
			config: &SecurityHeadersConfig{
				HSTSMaxAge:            intPtr(7776000),
				HSTSIncludeSubDomains: false,
				HSTSPreload:           false,
				XFrameOptions:         "SAMEORIGIN",
				XContentTypeOptions:   "nosniff",
				XXSSProtection:        "0",
				ReferrerPolicy:        "no-referrer",
				PermissionsPolicy:     "geolocation=(self)",
			},
			expectedHSTS: "max-age=7776000",
			expectedCSP:  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			expectedXCTO: "nosniff",
			expectedXFO:  "SAMEORIGIN",
			expectedXXSS: "0",
			expectedRP:   "no-referrer",
			expectedPP:   "geolocation=(self)",
		},
		{
			name: "disabled HSTS",
			config: &SecurityHeadersConfig{
				HSTSMaxAge: intPtr(0),
			},
			expectedHSTS: "",
			expectedCSP:  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
			expectedXCTO: "nosniff",
			expectedXFO:  "DENY",
			expectedXXSS: "1; mode=block",
			expectedRP:   "strict-origin-when-cross-origin",
			expectedPP:   "geolocation=(), microphone=(), camera=()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := SecurityHeaders(tt.config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			if tt.expectedHSTS != "" {
				if got := rec.Header().Get("Strict-Transport-Security"); got != tt.expectedHSTS {
					t.Errorf("HSTS header mismatch:\nexpected: %q\ngot:      %q", tt.expectedHSTS, got)
				}
			} else {
				if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
					t.Errorf("expected no HSTS header, got: %q", got)
				}
			}

			if got := rec.Header().Get("Content-Security-Policy"); got != tt.expectedCSP {
				t.Errorf("CSP header mismatch:\nexpected: %q\ngot:      %q", tt.expectedCSP, got)
			}

			if got := rec.Header().Get("X-Content-Type-Options"); got != tt.expectedXCTO {
				t.Errorf("X-Content-Type-Options mismatch: expected %q, got %q", tt.expectedXCTO, got)
			}

			if got := rec.Header().Get("X-Frame-Options"); got != tt.expectedXFO {
				t.Errorf("X-Frame-Options mismatch: expected %q, got %q", tt.expectedXFO, got)
			}

			if got := rec.Header().Get("X-XSS-Protection"); got != tt.expectedXXSS {
				t.Errorf("X-XSS-Protection mismatch: expected %q, got %q", tt.expectedXXSS, got)
			}

			if got := rec.Header().Get("Referrer-Policy"); got != tt.expectedRP {
				t.Errorf("Referrer-Policy mismatch: expected %q, got %q", tt.expectedRP, got)
			}

			if got := rec.Header().Get("Permissions-Policy"); got != tt.expectedPP {
				t.Errorf("Permissions-Policy mismatch: expected %q, got %q", tt.expectedPP, got)
			}
		})
	}
}

func TestSecurityHeaders_OverwritesExisting(t *testing.T) {
	handler := SecurityHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "ALLOW-FROM https://example.com")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Frame-Options"); got != "ALLOW-FROM https://example.com" {
		t.Errorf("expected handler to set X-Frame-Options to ALLOW-FROM https://example.com, got %q", got)
	}
}

func TestSecurityHeaders_AllMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handler := SecurityHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("security headers not set for method %s", method)
			}
		})
	}
}

func TestBuildCSP(t *testing.T) {
	tests := []struct {
		name     string
		config   *SecurityHeadersConfig
		expected string
	}{
		{
			name:     "nil config uses defaults",
			config:   nil,
			expected: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		},
		{
			name:     "empty config uses defaults",
			config:   &SecurityHeadersConfig{},
			expected: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		},
		{
			name: "custom directives",
			config: &SecurityHeadersConfig{
				CSPDefaultSrc:     []string{"'self'", "https://cdn.example.com"},
				CSPScriptSrc:      []string{"'self'", "'nonce-abc123'"},
				CSPStyleSrc:       []string{"'self'", "https://fonts.googleapis.com"},
				CSPImgSrc:         []string{"'self'", "data:", "https:", "blob:"},
				CSPFontSrc:        []string{"'self'", "https://fonts.gstatic.com"},
				CSPConnectSrc:     []string{"'self'", "https://api.example.com"},
				CSPFrameAncestors: []string{"'self'"},
				CSPBaseURI:        []string{"'self'"},
				CSPFormAction:     []string{"'self'", "https://form.example.com"},
			},
			expected: "default-src 'self' https://cdn.example.com; script-src 'self' 'nonce-abc123'; style-src 'self' https://fonts.googleapis.com; img-src 'self' data: https: blob:; font-src 'self' https://fonts.gstatic.com; connect-src 'self' https://api.example.com; frame-ancestors 'self'; base-uri 'self'; form-action 'self' https://form.example.com",
		},
		{
			name: "single values",
			config: &SecurityHeadersConfig{
				CSPDefaultSrc: []string{"'none'"},
				CSPScriptSrc:  []string{"'self'"},
			},
			expected: "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCSP(tt.config)
			if got != tt.expected {
				t.Errorf("CSP mismatch:\nexpected: %q\ngot:      %q", tt.expected, got)
			}
		})
	}
}

func TestBuildHSTS(t *testing.T) {
	tests := []struct {
		name     string
		config   *SecurityHeadersConfig
		expected string
	}{
		{
			name:     "nil config uses defaults",
			config:   nil,
			expected: "max-age=31536000; includeSubDomains",
		},
		{
			name:     "empty config uses defaults",
			config:   &SecurityHeadersConfig{},
			expected: "max-age=31536000; includeSubDomains",
		},
		{
			name: "custom max age only",
			config: &SecurityHeadersConfig{
				HSTSMaxAge: intPtr(7776000),
			},
			expected: "max-age=7776000",
		},
		{
			name: "with includeSubDomains",
			config: &SecurityHeadersConfig{
				HSTSMaxAge:            intPtr(31536000),
				HSTSIncludeSubDomains: true,
			},
			expected: "max-age=31536000; includeSubDomains",
		},
		{
			name: "with preload",
			config: &SecurityHeadersConfig{
				HSTSMaxAge:            intPtr(31536000),
				HSTSIncludeSubDomains: true,
				HSTSPreload:           true,
			},
			expected: "max-age=31536000; includeSubDomains; preload",
		},
		{
			name: "preload without includeSubDomains",
			config: &SecurityHeadersConfig{
				HSTSMaxAge:  intPtr(31536000),
				HSTSPreload: true,
			},
			expected: "max-age=31536000; preload",
		},
		{
			name: "zero max age disables HSTS",
			config: &SecurityHeadersConfig{
				HSTSMaxAge: intPtr(0),
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHSTS(tt.config)
			if got != tt.expected {
				t.Errorf("HSTS mismatch:\nexpected: %q\ngot:      %q", tt.expected, got)
			}
		})
	}
}

func TestSecurityHeaders_CSPReportOnly(t *testing.T) {
	config := &SecurityHeadersConfig{
		CSPReportOnly: true,
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Security-Policy") != "" {
		t.Error("Content-Security-Policy should not be set when CSPReportOnly is true")
	}

	expectedCSP := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
	if got := rec.Header().Get("Content-Security-Policy-Report-Only"); got != expectedCSP {
		t.Errorf("Content-Security-Policy-Report-Only mismatch:\nexpected: %q\ngot:      %q", expectedCSP, got)
	}
}

func TestSecurityHeaders_CSPWithReportURI(t *testing.T) {
	config := &SecurityHeadersConfig{
		CSPReportURI: "/api/csp-report",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "report-uri /api/csp-report") {
		t.Errorf("CSP should contain report-uri directive, got: %q", csp)
	}
}

func TestSecurityHeaders_Benchmark(t *testing.T) {
	handler := SecurityHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	for i := 0; i < 1000; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
