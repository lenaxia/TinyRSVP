package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsIPAllowlist(t *testing.T) {
	tests := []struct {
		name        string
		trustedIPs  []string
		realIP      string
		remoteAddr  string
		wantAllowed bool
		wantStatus  int
	}{
		{
			name:        "loopback allowed by default (empty allowlist, ipv4 loopback)",
			trustedIPs:  nil,
			realIP:      "",
			remoteAddr:  "127.0.0.1:40000",
			wantAllowed: true,
		},
		{
			name:        "loopback allowed by default (ipv6 loopback)",
			trustedIPs:  nil,
			realIP:      "",
			remoteAddr:  "[::1]:40000",
			wantAllowed: true,
		},
		{
			name:        "non-loopback denied by default",
			trustedIPs:  nil,
			realIP:      "",
			remoteAddr:  "203.0.113.5:40000",
			wantAllowed: false,
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "explicit trusted IP allowed",
			trustedIPs:  []string{"10.0.0.5"},
			realIP:      "10.0.0.5",
			remoteAddr:  "172.16.0.1:40000",
			wantAllowed: true,
		},
		{
			name:        "non-trusted IP denied when allowlist set",
			trustedIPs:  []string{"10.0.0.5"},
			realIP:      "10.0.0.9",
			remoteAddr:  "10.0.0.9:40000",
			wantAllowed: false,
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "trusted CIDR range allowed",
			trustedIPs:  []string{"172.19.0.0/16"},
			realIP:      "172.19.0.42",
			remoteAddr:  "172.19.0.42:40000",
			wantAllowed: true,
		},
		{
			name:        "IP outside trusted CIDR denied",
			trustedIPs:  []string{"172.19.0.0/16"},
			realIP:      "172.18.0.42",
			remoteAddr:  "172.18.0.42:40000",
			wantAllowed: false,
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "real IP from X-Forwarded-For honored when allowlist set",
			trustedIPs:  []string{"192.168.1.10"},
			realIP:      "192.168.1.10",
			remoteAddr:  "10.0.0.1:40000",
			wantAllowed: true,
		},
		{
			name:        "malformed remote addr denied (does not panic)",
			trustedIPs:  nil,
			realIP:      "",
			remoteAddr:  "not-an-ip",
			wantAllowed: false,
			wantStatus:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := MetricsIPAllowlist(tt.trustedIPs)(next)

			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.realIP != "" {
				ctx := context.WithValue(req.Context(), RealIPKey, tt.realIP)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if called != tt.wantAllowed {
				t.Fatalf("allowed=%v, want %v (status=%d)", called, tt.wantAllowed, rec.Code)
			}
			if !tt.wantAllowed && rec.Code != tt.wantStatus {
				t.Errorf("status=%d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestMetricsIPAllowlistDefaultIsLoopback(t *testing.T) {
	t.Run("loopback ipv4 default allowed", func(t *testing.T) {
		called := false
		handler := MetricsIPAllowlist(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !called {
			t.Error("loopback IPv4 must be allowed by default")
		}
	})

	t.Run("public IP default denied", func(t *testing.T) {
		called := false
		handler := MetricsIPAllowlist(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.RemoteAddr = "8.8.8.8:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if called {
			t.Error("public IP must be denied by default")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("status=%d, want 403", rec.Code)
		}
	})
}
