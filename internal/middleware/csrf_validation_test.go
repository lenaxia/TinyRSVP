package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCSRF_ValidationLogic(t *testing.T) {
	t.Run("exact reproduction of user's scenario", func(t *testing.T) {
		userToken := "94sNayaLR2RyTkeWcUSXHH6T9hwxtUmJZAjD-6UGLYQ="
		
		cookie := &http.Cookie{
			Name:  CSRFCookieName,
			Value: userToken,
		}

		formData := url.Values{}
		formData.Set("csrf_token", userToken)
		formData.Set("title", "sdfsadf")
		formData.Set("description", "")
		formData.Set("location", "")
		formData.Set("start_time", "0001-01-01T00:00")
		formData.Set("end_time", "")
		formData.Set("timezone", "America/Los_Angeles")
		formData.Set("rsvp_deadline", "")
		formData.Set("max_plus_ones", "0")
		formData.Set("action", "publish")

		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)

		t.Logf("Testing with exact user data")
		t.Logf("Cookie token: %s", userToken)
		t.Logf("Form data length: %d", len(formData.Encode()))

		submittedToken := getSubmittedToken(req)
		t.Logf("Submitted token from getSubmittedToken: %s", submittedToken)

		cookieFromReq, _ := req.Cookie(CSRFCookieName)
		t.Logf("Cookie from request: %s", cookieFromReq.Value)

		result := validateCSRFToken(req, userToken)
		t.Logf("Validation result: %v", result)

		if !result {
			t.Error("Validation failed with user's exact data!")
		}
	})

	t.Run("test getSubmittedToken with form data", func(t *testing.T) {
		token := "test-token-abc123"
		
		formData := url.Values{}
		formData.Set("csrf_token", token)
		formData.Set("other_field", "value")

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		submittedToken := getSubmittedToken(req)
		t.Logf("Submitted token: %s", submittedToken)
		t.Logf("Expected token: %s", token)

		if submittedToken != token {
			t.Errorf("getSubmittedToken returned wrong value: got %s, want %s", submittedToken, token)
		}

		formValue := req.FormValue("csrf_token")
		t.Logf("FormValue after getSubmittedToken: %s", formValue)

		if formValue != token {
			t.Error("Form data not accessible after getSubmittedToken")
		}
	})
}
