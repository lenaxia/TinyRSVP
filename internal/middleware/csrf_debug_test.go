package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRF_FormSubmissionDebug(t *testing.T) {
	t.Run("reproduce 403 with valid token in form", func(t *testing.T) {
		csrfMiddleware := CSRF(32)

		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		}))

		req1 := httptest.NewRequest(http.MethodGet, "/events/new", nil)
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)

		var csrfCookie *http.Cookie
		for _, c := range rec1.Result().Cookies() {
			if c.Name == CSRFCookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("Expected CSRF cookie from GET request")
		}

		t.Logf("CSRF token from cookie: %s", csrfCookie.Value)

		formData := url.Values{}
		formData.Set("csrf_token", csrfCookie.Value)
		formData.Set("title", "Test Event")
		formData.Set("start_time", "2026-01-15T10:00")
		formData.Set("timezone", "America/Los_Angeles")

		req2 := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(formData.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.AddCookie(csrfCookie)

		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		t.Logf("Response status: %d", rec2.Code)
		t.Logf("Response body: %s", rec2.Body.String())

		if rec2.Code == http.StatusForbidden {
			t.Errorf("Expected success with valid CSRF token, got 403")

			if err := req2.ParseForm(); err == nil {
				t.Logf("Submitted csrf_token: %s", req2.FormValue("csrf_token"))
			}
		}
	})

	t.Run("test token comparison logic", func(t *testing.T) {
		token := "94sNayaLR2RyTkeWcUSXHH6T9hwxtUmJZAjD-6UGLYQ="

		csrfMiddleware := CSRF(32)

		var capturedExpectedToken string
		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedExpectedToken = GetCSRFToken(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		cookie := &http.Cookie{
			Name:  CSRFCookieName,
			Value: token,
		}

		formData := url.Values{}
		formData.Set("csrf_token", token)

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("Cookie token: %s", token)
		t.Logf("Expected token from context: %s", capturedExpectedToken)
		t.Logf("Response status: %d", rec.Code)

		if rec.Code == http.StatusForbidden {
			t.Errorf("Token comparison failed even though tokens match")
		}
	})
}
