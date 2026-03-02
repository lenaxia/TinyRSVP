package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRF_BodyConsumption(t *testing.T) {
	t.Run("ParseForm consumes body", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("csrf_token", "test-token")
		formData.Set("title", "Test Event")

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		if err := req.ParseForm(); err != nil {
			t.Fatalf("First ParseForm failed: %v", err)
		}

		t.Logf("First parse - csrf_token: %s", req.FormValue("csrf_token"))
		t.Logf("First parse - title: %s", req.FormValue("title"))

		body, err := io.ReadAll(req.Body)
		t.Logf("Body after first ParseForm: %q (len=%d)", string(body), len(body))
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}

		if err := req.ParseForm(); err != nil {
			t.Logf("Second ParseForm error: %v", err)
		}

		t.Logf("Second parse - csrf_token: %s", req.FormValue("csrf_token"))
		t.Logf("Second parse - title: %s", req.FormValue("title"))
	})

	t.Run("CSRF middleware with handler that parses form", func(t *testing.T) {
		csrfMiddleware := CSRF(32)

		var handlerParsedTitle string
		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				t.Logf("Handler ParseForm error: %v", err)
			}
			handlerParsedTitle = r.FormValue("title")
			t.Logf("Handler got title: %s", handlerParsedTitle)
			w.WriteHeader(http.StatusOK)
		}))

		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)

		var csrfCookie *http.Cookie
		for _, c := range rec1.Result().Cookies() {
			if c.Name == CSRFCookieName {
				csrfCookie = c
				break
			}
		}

		formData := url.Values{}
		formData.Set("csrf_token", csrfCookie.Value)
		formData.Set("title", "Test Event Title")

		req2 := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(formData.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.AddCookie(csrfCookie)

		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", rec2.Code, rec2.Body.String())
		}

		if handlerParsedTitle == "" {
			t.Error("Handler did not receive form data - body was consumed by CSRF middleware!")
		}
	})
}
