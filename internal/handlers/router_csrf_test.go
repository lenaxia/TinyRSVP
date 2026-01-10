package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/middleware"
)

func TestRouter_CSRF_Integration(t *testing.T) {
	router := NewRouter(nil)

	t.Run("GET request sets CSRF cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		cookies := rec.Result().Cookies()
		var csrfCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.CSRFCookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("Expected CSRF cookie to be set on GET request")
		}

		if csrfCookie.Value == "" {
			t.Error("Expected non-empty CSRF cookie value")
		}
	})

	t.Run("POST without CSRF token returns 403", func(t *testing.T) {
		body := strings.NewReader(`{"test":"data"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/events", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for POST without CSRF token, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("POST with valid CSRF token succeeds", func(t *testing.T) {
		req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)

		cookies := rec1.Result().Cookies()
		var csrfCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.CSRFCookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("Expected CSRF cookie from GET request")
		}

		body := strings.NewReader(`{"test":"data"}`)
		req2 := httptest.NewRequest(http.MethodPost, "/api/events", body)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set(middleware.CSRFHeaderName, csrfCookie.Value)
		req2.AddCookie(csrfCookie)
		rec2 := httptest.NewRecorder()

		router.ServeHTTP(rec2, req2)

		if rec2.Code == http.StatusForbidden {
			t.Errorf("Expected POST with valid CSRF token to not return 403, got %d: %s", rec2.Code, rec2.Body.String())
		}
	})

	t.Run("PUT without CSRF token returns 403", func(t *testing.T) {
		body := strings.NewReader(`{"test":"data"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/events/1", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for PUT without CSRF token, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("DELETE without CSRF token returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/events/1", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for DELETE without CSRF token, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("PATCH without CSRF token returns 403", func(t *testing.T) {
		body := strings.NewReader(`{"test":"data"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/events/1", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for PATCH without CSRF token, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("Safe methods work without CSRF token", func(t *testing.T) {
		safeMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}

		for _, method := range safeMethods {
			req := httptest.NewRequest(method, "/health", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusForbidden {
				t.Errorf("Expected %s request to work without CSRF token, got %d", method, rec.Code)
			}
		}
	})

	t.Run("CSRF token rotates after POST", func(t *testing.T) {
		req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)

		cookies1 := rec1.Result().Cookies()
		var csrfCookie1 *http.Cookie
		for _, c := range cookies1 {
			if c.Name == middleware.CSRFCookieName {
				csrfCookie1 = c
				break
			}
		}

		body := strings.NewReader(`{"test":"data"}`)
		req2 := httptest.NewRequest(http.MethodPost, "/api/events", body)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set(middleware.CSRFHeaderName, csrfCookie1.Value)
		req2.AddCookie(csrfCookie1)
		rec2 := httptest.NewRecorder()

		router.ServeHTTP(rec2, req2)

		cookies2 := rec2.Result().Cookies()
		var csrfCookie2 *http.Cookie
		for _, c := range cookies2 {
			if c.Name == middleware.CSRFCookieName {
				csrfCookie2 = c
				break
			}
		}

		if csrfCookie2 == nil {
			t.Fatal("Expected new CSRF cookie after POST")
		}

		if csrfCookie1.Value == csrfCookie2.Value {
			t.Error("Expected CSRF token to rotate after POST")
		}
	})
}

func TestRouter_CSRF_WithAuth(t *testing.T) {
	router := NewRouter(nil)

	t.Run("Authenticated POST requires CSRF token", func(t *testing.T) {
		body := strings.NewReader(`{"name":"Test Event"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/events", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for authenticated POST without CSRF, got %d", http.StatusForbidden, rec.Code)
		}
	})
}

func TestRouter_CSRF_FormSubmission(t *testing.T) {
	router := NewRouter(nil)

	t.Run("Form submission with CSRF token in body", func(t *testing.T) {
		req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)

		cookies := rec1.Result().Cookies()
		var csrfCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == middleware.CSRFCookieName {
				csrfCookie = c
				break
			}
		}

		body := strings.NewReader("csrf_token=" + csrfCookie.Value + "&name=Test")
		req2 := httptest.NewRequest(http.MethodPost, "/api/events", body)
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.AddCookie(csrfCookie)
		rec2 := httptest.NewRecorder()

		router.ServeHTTP(rec2, req2)

		if rec2.Code == http.StatusForbidden {
			t.Errorf("Expected form submission with CSRF token to not return 403, got %d", rec2.Code)
		}
	})
}
