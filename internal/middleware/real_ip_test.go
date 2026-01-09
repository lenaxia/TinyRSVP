package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealIP_FromRemoteAddr(t *testing.T) {
	var capturedIP string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIP = GetRealIP(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "192.168.1.100:12345" {
		t.Errorf("expected IP 192.168.1.100:12345, got %s", capturedIP)
	}
}

func TestRealIP_FromXRealIP(t *testing.T) {
	var capturedIP string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIP = GetRealIP(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Real-IP", "203.0.113.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "203.0.113.1" {
		t.Errorf("expected IP 203.0.113.1, got %s", capturedIP)
	}
}

func TestRealIP_FromXForwardedFor(t *testing.T) {
	var capturedIP string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIP = GetRealIP(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1, 192.0.2.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "203.0.113.1" {
		t.Errorf("expected first IP 203.0.113.1, got %s", capturedIP)
	}
}

func TestRealIP_XRealIPTakesPrecedence(t *testing.T) {
	var capturedIP string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIP = GetRealIP(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.Header.Set("X-Forwarded-For", "198.51.100.1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "203.0.113.1" {
		t.Errorf("expected X-Real-IP to take precedence, got %s", capturedIP)
	}
}

func TestRealIP_EmptyXForwardedFor(t *testing.T) {
	var capturedIP string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIP = GetRealIP(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Forwarded-For", "")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "192.168.1.100:12345" {
		t.Errorf("expected RemoteAddr fallback, got %s", capturedIP)
	}
}

func TestRealIP_WhitespaceInXForwardedFor(t *testing.T) {
	var capturedIP string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIP = GetRealIP(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Forwarded-For", " 203.0.113.1 , 198.51.100.1 ")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedIP != "203.0.113.1" {
		t.Errorf("expected trimmed IP 203.0.113.1, got %s", capturedIP)
	}
}

func TestRealIP_EmptyContext(t *testing.T) {
	ctx := context.Background()
	ip := GetRealIP(ctx)

	if ip != "" {
		t.Errorf("expected empty string for context without IP, got %s", ip)
	}
}

func TestRealIP_ContextInjection(t *testing.T) {
	testIP := "203.0.113.1"
	ctx := context.WithValue(context.Background(), RealIPKey, testIP)

	ip := GetRealIP(ctx)
	if ip != testIP {
		t.Errorf("expected %s, got %s", testIP, ip)
	}
}

func TestRealIP_InvalidContextValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), RealIPKey, 12345)

	ip := GetRealIP(ctx)
	if ip != "" {
		t.Errorf("expected empty string for invalid type, got %s", ip)
	}
}

func TestRealIP_PreservesHandlerBehavior(t *testing.T) {
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", rec.Body.String())
	}
}
