package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRF_SafeMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"GET request", http.MethodGet},
		{"HEAD request", http.MethodHead},
		{"OPTIONS request", http.MethodOptions},
		{"TRACE request", http.MethodTrace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status %d for safe method %s, got %d", http.StatusOK, tt.method, rec.Code)
			}
		})
	}
}

func TestCSRF_UnsafeMethods_MissingToken(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"POST without token", http.MethodPost},
		{"PUT without token", http.MethodPut},
		{"DELETE without token", http.MethodDelete},
		{"PATCH without token", http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("Expected status %d for %s without token, got %d", http.StatusForbidden, tt.method, rec.Code)
			}

			if !strings.Contains(rec.Body.String(), "CSRF") {
				t.Error("Expected error message to mention CSRF")
			}
		})
	}
}

func TestCSRF_TokenGeneration(t *testing.T) {
	tests := []struct {
		name       string
		tokenLen   int
		wantMinLen int
	}{
		{"default length", 32, 40},
		{"short length", 16, 20},
		{"long length", 64, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CSRF(tt.tokenLen)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := GetCSRFToken(r.Context())
				if token == "" {
					t.Error("Expected CSRF token in context, got empty string")
				}
				if len(token) < tt.wantMinLen {
					t.Errorf("Expected token length >= %d, got %d", tt.wantMinLen, len(token))
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

			if csrfCookie.Value == "" {
				t.Error("Expected non-empty CSRF cookie value")
			}

			if len(csrfCookie.Value) < tt.wantMinLen {
				t.Errorf("Expected cookie value length >= %d, got %d", tt.wantMinLen, len(csrfCookie.Value))
			}
		})
	}
}

func TestCSRF_CookieAttributes(t *testing.T) {
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

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

	if csrfCookie.Path != "/" {
		t.Errorf("Expected cookie path '/', got '%s'", csrfCookie.Path)
	}

	if csrfCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite=Strict, got %v", csrfCookie.SameSite)
	}

	if csrfCookie.HttpOnly {
		t.Error("Expected HttpOnly=false for CSRF cookie (needs to be readable by JavaScript)")
	}

	if csrfCookie.Secure {
		t.Error("Expected Secure=false in test environment")
	}
}

func TestCSRF_ValidToken_FormField(t *testing.T) {
	var capturedToken string
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedToken = GetCSRFToken(r.Context())
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

	if csrfCookie == nil {
		t.Fatal("Expected CSRF cookie to be set")
	}

	body := strings.NewReader("csrf_token=" + capturedToken)
	req2 := httptest.NewRequest(http.MethodPost, "/test", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status %d with valid token, got %d: %s", http.StatusOK, rec2.Code, rec2.Body.String())
	}
}

func TestCSRF_ValidToken_Header(t *testing.T) {
	var capturedToken string
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedToken = GetCSRFToken(r.Context())
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

	if csrfCookie == nil {
		t.Fatal("Expected CSRF cookie to be set")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.Header.Set(CSRFHeaderName, capturedToken)
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status %d with valid token in header, got %d: %s", http.StatusOK, rec2.Code, rec2.Body.String())
	}
}

func TestCSRF_InvalidToken(t *testing.T) {
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	if csrfCookie == nil {
		t.Fatal("Expected CSRF cookie to be set")
	}

	body := strings.NewReader("csrf_token=invalid_token_value")
	req2 := httptest.NewRequest(http.MethodPost, "/test", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Errorf("Expected status %d with invalid token, got %d", http.StatusForbidden, rec2.Code)
	}
}

func TestCSRF_MissingCookie(t *testing.T) {
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader("csrf_token=some_token")
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d without cookie, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestCSRF_TokenRotation(t *testing.T) {
	var token1, token2 string
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			token1 = GetCSRFToken(r.Context())
		} else {
			token2 = GetCSRFToken(r.Context())
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

	body := strings.NewReader("csrf_token=" + token1)
	req2 := httptest.NewRequest(http.MethodPost, "/test", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, rec2.Code)
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
		t.Fatal("Expected new CSRF cookie after token use")
	}

	if token2 == "" {
		t.Fatal("Expected token2 to be captured")
	}

	if token1 == token2 {
		t.Error("Expected token to rotate after use, but tokens are identical")
	}

	if csrfCookie.Value == newCSRFCookie.Value {
		t.Error("Expected cookie value to change after token use")
	}
}

func TestCSRF_GetCSRFToken_NoToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	token := GetCSRFToken(req.Context())

	if token != "" {
		t.Errorf("Expected empty token from context without CSRF middleware, got %s", token)
	}
}

func TestCSRF_ZeroLength(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with zero token length")
		}
	}()

	CSRF(0)
}

func TestCSRF_NegativeLength(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with negative token length")
		}
	}()

	CSRF(-1)
}

func TestCSRF_MultipleRequests_DifferentTokens(t *testing.T) {
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tokens := make(map[string]bool)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		cookies := rec.Result().Cookies()
		for _, c := range cookies {
			if c.Name == CSRFCookieName {
				if tokens[c.Value] {
					t.Errorf("Token collision detected: %s", c.Value)
				}
				tokens[c.Value] = true
			}
		}
	}

	if len(tokens) != 10 {
		t.Errorf("Expected 10 unique tokens, got %d", len(tokens))
	}
}

func TestCSRF_EmptyTokenValue(t *testing.T) {
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	body := strings.NewReader("csrf_token=")
	req2 := httptest.NewRequest(http.MethodPost, "/test", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Errorf("Expected status %d with empty token, got %d", http.StatusForbidden, rec2.Code)
	}
}

func TestCSRF_HeaderPrecedence(t *testing.T) {
	var capturedToken string
	handler := CSRF(32)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedToken = GetCSRFToken(r.Context())
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

	body := strings.NewReader("csrf_token=wrong_token")
	req2 := httptest.NewRequest(http.MethodPost, "/test", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set(CSRFHeaderName, capturedToken)
	req2.AddCookie(csrfCookie)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected header to take precedence, got status %d", rec2.Code)
	}
}
