package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain_EmptyChain(t *testing.T) {
	handler := Chain()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", rec.Body.String())
	}
}

func TestChain_SingleMiddleware(t *testing.T) {
	called := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	handler := Chain(middleware)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected middleware to be called")
	}
}

func TestChain_MultipleMiddleware(t *testing.T) {
	var order []string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-before")
			next.ServeHTTP(w, r)
			order = append(order, "m1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-before")
			next.ServeHTTP(w, r)
			order = append(order, "m2-after")
		})
	}

	middleware3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m3-before")
			next.ServeHTTP(w, r)
			order = append(order, "m3-after")
		})
	}

	handler := Chain(middleware1, middleware2, middleware3)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expected := []string{"m1-before", "m2-before", "m3-before", "handler", "m3-after", "m2-after", "m1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}

	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, order[i])
		}
	}
}

func TestChain_MiddlewareCanShortCircuit(t *testing.T) {
	handlerCalled := false

	shortCircuit := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
	}

	handler := Chain(shortCircuit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("expected handler not to be called when middleware short-circuits")
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestChain_PreservesResponseWriter(t *testing.T) {
	handler := Chain(RequestID, RealIP)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		t.Errorf("expected custom header")
	}
}

func TestChain_ContextPropagation(t *testing.T) {
	handler := Chain(RequestID, RealIP)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		realIP := GetRealIP(r.Context())

		if requestID == "" {
			t.Error("expected request ID in context")
		}
		if realIP == "" {
			t.Error("expected real IP in context")
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}
