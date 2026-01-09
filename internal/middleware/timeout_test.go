package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout_FastRequest(t *testing.T) {
	handler := Timeout(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", rec.Body.String())
	}
}

func TestTimeout_SlowRequest(t *testing.T) {
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "timeout") && !strings.Contains(body, "Timeout") {
		t.Errorf("expected timeout message in body, got %s", body)
	}
}

func TestTimeout_ContextCancellation(t *testing.T) {
	contextCancelled := false
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			contextCancelled = true
			return
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	time.Sleep(100 * time.Millisecond)

	if !contextCancelled {
		t.Error("expected context to be cancelled")
	}
}

func TestTimeout_ZeroTimeout(t *testing.T) {
	handler := Timeout(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected immediate timeout with 0 duration, got status %d", rec.Code)
	}
}

func TestTimeout_VeryLongTimeout(t *testing.T) {
	handler := Timeout(10 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTimeout_MultipleRequests(t *testing.T) {
	handler := Timeout(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, rec.Code)
		}
	}
}

func TestTimeout_PreservesHandlerBehavior(t *testing.T) {
	handler := Timeout(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
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

	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("expected custom header, got %s", rec.Header().Get("X-Custom"))
	}
}
