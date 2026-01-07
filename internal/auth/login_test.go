package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLogin_Success(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)

	err = auth.HandleLogin(w, r)
	if err != nil {
		t.Fatalf("HandleLogin failed: %v", err)
	}

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("Expected Location header, got empty")
	}

	if !strings.Contains(location, "/authorize") {
		t.Errorf("Expected authorization URL, got %s", location)
	}

	if !strings.Contains(location, "client_id=test-client-id") {
		t.Errorf("Expected client_id in URL, got %s", location)
	}

	if !strings.Contains(location, "redirect_uri=") {
		t.Errorf("Expected redirect_uri in URL, got %s", location)
	}

	if !strings.Contains(location, "state=") {
		t.Errorf("Expected state parameter in URL, got %s", location)
	}

	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			stateCookie = c
			break
		}
	}

	if stateCookie == nil {
		t.Fatal("Expected oauth_state cookie, got none")
	}

	if stateCookie.Value == "" {
		t.Error("Expected non-empty state cookie value")
	}

	if stateCookie.MaxAge != 300 {
		t.Errorf("Expected MaxAge 300, got %d", stateCookie.MaxAge)
	}

	if !stateCookie.HttpOnly {
		t.Error("Expected HttpOnly cookie")
	}

	if stateCookie.Path != "/" {
		t.Errorf("Expected Path /, got %s", stateCookie.Path)
	}

	if stateCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected SameSite Lax, got %v", stateCookie.SameSite)
	}
}

func TestHandleLogin_StateGeneration(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/login", nil)
	err = auth.HandleLogin(w1, r1)
	if err != nil {
		t.Fatalf("HandleLogin failed: %v", err)
	}

	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/login", nil)
	err = auth.HandleLogin(w2, r2)
	if err != nil {
		t.Fatalf("HandleLogin failed: %v", err)
	}

	var state1, state2 string
	for _, c := range w1.Result().Cookies() {
		if c.Name == "oauth_state" {
			state1 = c.Value
		}
	}
	for _, c := range w2.Result().Cookies() {
		if c.Name == "oauth_state" {
			state2 = c.Value
		}
	}

	if state1 == "" || state2 == "" {
		t.Fatal("Expected state cookies in both requests")
	}

	if state1 == state2 {
		t.Error("Expected different state values for each request")
	}

	if len(state1) < 32 {
		t.Errorf("Expected state length >= 32, got %d", len(state1))
	}
}

func TestHandleLogin_ScopesIncluded(t *testing.T) {
	mockProvider := setupMockOIDCProvider(t)
	defer mockProvider.Close()

	cfg := &OIDCConfig{
		IssuerURL:    mockProvider.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}

	mockUserService := &MockUserService{}
	mockSessionMgr := &MockSessionManager{}

	auth, err := NewOIDCAuthenticator(cfg, mockUserService, mockSessionMgr)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)

	err = auth.HandleLogin(w, r)
	if err != nil {
		t.Fatalf("HandleLogin failed: %v", err)
	}

	location := w.Header().Get("Location")

	if !strings.Contains(location, "scope=") {
		t.Error("Expected scope parameter in authorization URL")
	}

	if !strings.Contains(location, "openid") {
		t.Error("Expected openid scope in URL")
	}
}
