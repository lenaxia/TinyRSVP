package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRF_DetailedDebug(t *testing.T) {
	t.Run("trace exact flow with existing cookie", func(t *testing.T) {
		existingToken := "94sNayaLR2RyTkeWcUSXHH6T9hwxtUmJZAjD-6UGLYQ="

		csrfMiddleware := CSRF(32)

		var contextToken string
		handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextToken = GetCSRFToken(r.Context())
			t.Logf("Token in context during handler: %s", contextToken)
			w.WriteHeader(http.StatusOK)
		}))

		cookie := &http.Cookie{
			Name:  CSRFCookieName,
			Value: existingToken,
		}

		formData := url.Values{}
		formData.Set("csrf_token", existingToken)
		formData.Set("title", "Test")

		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)

		t.Logf("Cookie value: %s", existingToken)
		t.Logf("Form csrf_token: %s", formData.Get("csrf_token"))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("Response status: %d", rec.Code)
		t.Logf("Response body: %s", rec.Body.String())
		t.Logf("Context token: %s", contextToken)

		if rec.Code == http.StatusForbidden {
			t.Errorf("Got 403 with matching tokens - this is the bug!")
		}
	})

	t.Run("manual validation test", func(t *testing.T) {
		token := "94sNayaLR2RyTkeWcUSXHH6T9hwxtUmJZAjD-6UGLYQ="

		cookie := &http.Cookie{
			Name:  CSRFCookieName,
			Value: token,
		}

		formData := url.Values{}
		formData.Set("csrf_token", token)

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)

		if err := req.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		submittedToken := getSubmittedToken(req)
		t.Logf("Submitted token: %s", submittedToken)
		t.Logf("Expected token: %s", token)
		t.Logf("Cookie token: %s", cookie.Value)

		result := validateCSRFToken(req, token)
		t.Logf("Validation result: %v", result)

		if !result {
			t.Error("Manual validation failed with matching tokens")
		}
	})

	t.Run("test with context value", func(t *testing.T) {
		token := "test-token-12345"

		ctx := context.WithValue(context.Background(), csrfTokenKey, token)

		retrievedToken := GetCSRFToken(ctx)

		if retrievedToken != token {
			t.Errorf("Expected token %s, got %s", token, retrievedToken)
		}
	})
}
