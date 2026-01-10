package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRF_Integration_WithChain(t *testing.T) {
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := GetCSRFToken(r.Context())
		if token == "" {
			t.Error("Expected CSRF token in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("Expected CSRF cookie to be set")
	}
}

func TestCSRF_Integration_MultipleRequests(t *testing.T) {
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/form", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	cookies := rec1.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("Expected CSRF cookie after GET")
	}

	body := strings.NewReader("csrf_token=" + csrfCookie.Value + "&data=test")
	req2 := httptest.NewRequest(http.MethodPost, "/submit", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status %d for valid CSRF token, got %d: %s", http.StatusOK, rec2.Code, rec2.Body.String())
	}

	newCookies := rec2.Result().Cookies()
	var newCSRFCookie *http.Cookie
	for _, c := range newCookies {
		if c.Name == CSRFCookieName {
			newCSRFCookie = c
			break
		}
	}

	if newCSRFCookie == nil {
		t.Fatal("Expected rotated CSRF cookie after POST")
	}

	if csrfCookie.Value == newCSRFCookie.Value {
		t.Error("Expected token to rotate after POST")
	}
}

func TestCSRF_Integration_AJAXRequest(t *testing.T) {
	var capturedToken string
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			capturedToken = GetCSRFToken(r.Context())
		}
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	cookies := rec1.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/data", strings.NewReader(`{"data":"test"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(CSRFHeaderName, capturedToken)
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status %d for AJAX request with CSRF header, got %d", http.StatusOK, rec2.Code)
	}
}

func TestCSRF_Integration_FailedValidation(t *testing.T) {
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader("data=test")
	req := httptest.NewRequest(http.MethodPost, "/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for missing CSRF token, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestCSRF_Integration_RecoveryAfterPanic(t *testing.T) {
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d after panic, got %d", http.StatusInternalServerError, rec.Code)
	}

	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil {
		t.Error("Expected CSRF cookie to be set even after panic")
	}
}

func TestCSRF_Integration_ConcurrentRequests(t *testing.T) {
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				errors <- http.ErrAbortHandler
				return
			}

			cookies := rec.Result().Cookies()
			var csrfCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == CSRFCookieName {
					csrfCookie = c
					break
				}
			}

			if csrfCookie == nil {
				errors <- http.ErrNoCookie
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
	}
}

func TestCSRF_Integration_FormAndHeaderPrecedence(t *testing.T) {
	var capturedToken string
	handler := Chain(
		Recovery,
		RequestID,
		CSRF(32),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			capturedToken = GetCSRFToken(r.Context())
		}
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	cookies := rec1.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CSRFCookieName {
			csrfCookie = c
			break
		}
	}

	body := strings.NewReader("csrf_token=wrong_token&data=test")
	req2 := httptest.NewRequest(http.MethodPost, "/submit", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set(CSRFHeaderName, capturedToken)
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected header to take precedence over form field, got status %d", rec2.Code)
	}
}
