package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/middleware"
)

func TestRouter_CSPReportEndpoint(t *testing.T) {
	router := NewRouter(nil)

	csrfToken, csrfCookie := getCSRFTokenFromRouter(router)

	report := map[string]interface{}{
		"csp-report": map[string]interface{}{
			"document-uri":       "https://example.com/page",
			"violated-directive": "script-src 'self'",
			"blocked-uri":        "https://evil.com/script.js",
		},
	}

	body, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/csp-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/csp-report")
	req.Header.Set(middleware.CSRFHeaderName, csrfToken)
	req.AddCookie(csrfCookie)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestRouter_SecurityHeadersOnAllRoutes(t *testing.T) {
	router := NewRouter(nil)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodGet, "/login"},
		{http.MethodGet, "/auth/callback"},
		{http.MethodGet, "/static/css/main.css"},
		{http.MethodGet, "/rsvp/token123"},
	}

	requiredHeaders := []string{
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			for _, header := range requiredHeaders {
				if rec.Header().Get(header) == "" {
					t.Errorf("expected header %s to be set on %s %s", header, route.method, route.path)
				}
			}

			csp := rec.Header().Get("Content-Security-Policy")
			if !strings.Contains(csp, "default-src 'self'") {
				t.Errorf("CSP should contain default-src 'self', got: %s", csp)
			}

			hsts := rec.Header().Get("Strict-Transport-Security")
			if !strings.Contains(hsts, "max-age=") {
				t.Errorf("HSTS should contain max-age, got: %s", hsts)
			}
		})
	}
}

func TestRouter_SecurityHeadersOnAPIRoutes(t *testing.T) {
	router := NewRouter(&RouterHandlers{
		EventHandlers: &mockEventHandlers{},
	})

	apiRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/events"},
		{http.MethodPost, "/api/events"},
	}

	for _, route := range apiRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Error("expected X-Content-Type-Options header on API route")
			}

			if rec.Header().Get("Content-Security-Policy") == "" {
				t.Error("expected Content-Security-Policy header on API route")
			}
		})
	}
}
