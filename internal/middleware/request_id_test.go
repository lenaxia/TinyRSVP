package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_GeneratesID(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected request ID to be generated")
	}

	responseID := rec.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	if capturedID != responseID {
		t.Errorf("context ID %s != response header ID %s", capturedID, responseID)
	}
}

func TestRequestID_UsesExistingID(t *testing.T) {
	existingID := "existing-request-id-12345"
	var capturedID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedID != existingID {
		t.Errorf("expected ID %s, got %s", existingID, capturedID)
	}

	responseID := rec.Header().Get("X-Request-ID")
	if responseID != existingID {
		t.Errorf("expected response header %s, got %s", existingID, responseID)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	ids := make(map[string]bool)
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		ids[id] = true
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if len(ids) != 100 {
		t.Errorf("expected 100 unique IDs, got %d", len(ids))
	}
}

func TestRequestID_ValidFormat(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected request ID")
	}

	if len(id) < 16 {
		t.Errorf("request ID too short: %s", id)
	}

	if strings.Contains(id, " ") {
		t.Errorf("request ID contains spaces: %s", id)
	}
}

func TestRequestID_EmptyContext(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)

	if id != "" {
		t.Errorf("expected empty string for context without ID, got %s", id)
	}
}

func TestRequestID_ContextInjection(t *testing.T) {
	testID := "test-context-id"
	ctx := context.WithValue(context.Background(), RequestIDKey, testID)

	id := GetRequestID(ctx)
	if id != testID {
		t.Errorf("expected %s, got %s", testID, id)
	}
}

func TestRequestID_InvalidContextValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, 12345)

	id := GetRequestID(ctx)
	if id != "" {
		t.Errorf("expected empty string for invalid type, got %s", id)
	}
}

func TestRequestID_PreservesHandlerBehavior(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}
