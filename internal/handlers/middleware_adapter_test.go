package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAdapter_RequireAuth(t *testing.T) {
	authCalled := false
	requireAuthFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			next.ServeHTTP(w, r)
		})
	}

	adapter := NewMiddlewareAdapter(requireAuthFunc, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := adapter.RequireAuth(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if !authCalled {
		t.Error("Expected auth middleware to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddlewareAdapter_RequireAdmin(t *testing.T) {
	adminCalled := false
	requireAdminFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			adminCalled = true
			next.ServeHTTP(w, r)
		})
	}

	adapter := NewMiddlewareAdapter(nil, requireAdminFunc)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := adapter.RequireAdmin(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if !adminCalled {
		t.Error("Expected admin middleware to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddlewareAdapter_NilMiddleware(t *testing.T) {
	adapter := NewMiddlewareAdapter(nil, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authWrapped := adapter.RequireAuth(handler)
	adminWrapped := adapter.RequireAdmin(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	w := httptest.NewRecorder()
	authWrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("RequireAuth with nil func: expected status 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	adminWrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("RequireAdmin with nil func: expected status 200, got %d", w.Code)
	}
}

func TestMiddlewareAdapter_ChainedMiddleware(t *testing.T) {
	authCalled := false
	adminCalled := false

	requireAuthFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			next.ServeHTTP(w, r)
		})
	}

	requireAdminFunc := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			adminCalled = true
			next.ServeHTTP(w, r)
		})
	}

	adapter := NewMiddlewareAdapter(requireAuthFunc, requireAdminFunc)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := adapter.RequireAuth(adapter.RequireAdmin(handler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if !authCalled {
		t.Error("Expected auth middleware to be called")
	}

	if !adminCalled {
		t.Error("Expected admin middleware to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
