package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

func TestHandleCallback_ValidCallback(t *testing.T) {
	mockProvider, privateKey := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	idToken := createTestIDToken(t, privateKey, mockProvider.URL, cfg.ClientID, "user123", "user@example.com", "Test User")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=test-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	mockProvider.setTokenResponse(idToken)

	result, err := auth.HandleCallback(w, r)
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected AuthResult, got nil")
	}

	if result.Email != "user@example.com" {
		t.Errorf("Expected email user@example.com, got %s", result.Email)
	}

	if result.Name != "Test User" {
		t.Errorf("Expected name Test User, got %s", result.Name)
	}

	if result.OIDCSubject == nil {
		t.Fatal("Expected OIDC subject, got nil")
	}

	if *result.OIDCSubject != "user123" {
		t.Errorf("Expected subject user123, got %s", *result.OIDCSubject)
	}
}

func TestHandleCallback_MissingStateCookie(t *testing.T) {
	mockProvider, _ := setupMockOIDCProviderWithJWKS(t)
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
	r := httptest.NewRequest("GET", "/auth/callback?code=test-code&state=test-state", nil)

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for missing state cookie, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func TestHandleCallback_StateMismatch(t *testing.T) {
	mockProvider, _ := setupMockOIDCProviderWithJWKS(t)
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
	r := httptest.NewRequest("GET", "/auth/callback?code=test-code&state=state-from-query", nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: "different-state-in-cookie",
	})

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for state mismatch, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func TestHandleCallback_MissingAuthorizationCode(t *testing.T) {
	mockProvider, _ := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for missing code, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func TestHandleCallback_MissingEmailClaim(t *testing.T) {
	mockProvider, privateKey := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	idToken := createTestIDTokenWithoutEmail(t, privateKey, mockProvider.URL, cfg.ClientID, "user123", "Test User")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=test-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	mockProvider.setTokenResponse(idToken)

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for missing email claim, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func TestHandleCallback_InvalidAuthCode(t *testing.T) {
	mockProvider, _ := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	mockProvider.setTokenError()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=invalid-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for invalid auth code, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

type mockOIDCServer struct {
	*httptest.Server
	tokenResponse string
	tokenError    bool
}

func setupMockOIDCProviderWithJWKS(t *testing.T) (*mockOIDCServer, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	jwk := jose.JSONWebKey{
		Key:       &privateKey.PublicKey,
		KeyID:     "test-key-id",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}

	mock := &mockOIDCServer{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			discoveryDoc := fmt.Sprintf(`{
				"issuer": "%s",
				"authorization_endpoint": "%s/authorize",
				"token_endpoint": "%s/token",
				"jwks_uri": "%s/jwks",
				"response_types_supported": ["code"],
				"subject_types_supported": ["public"],
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, mock.URL, mock.URL, mock.URL, mock.URL)
			w.Write([]byte(discoveryDoc))

		case "/jwks":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jwks)

		case "/token":
			if mock.tokenError {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "invalid_grant"}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := fmt.Sprintf(`{
				"access_token": "test-access-token",
				"token_type": "Bearer",
				"expires_in": 3600,
				"id_token": "%s"
			}`, mock.tokenResponse)
			w.Write([]byte(response))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	mock.Server = server
	return mock, privateKey
}

func (m *mockOIDCServer) setTokenResponse(idToken string) {
	m.tokenResponse = idToken
	m.tokenError = false
}

func (m *mockOIDCServer) setTokenError() {
	m.tokenError = true
}

func createTestIDToken(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, subject, email, name string) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: privateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key-id"),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	claims := jwt.Claims{
		Issuer:   issuer,
		Subject:  subject,
		Audience: jwt.Audience{audience},
		Expiry:   jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}

	customClaims := map[string]interface{}{
		"email": email,
		"name":  name,
	}

	builder := jwt.Signed(signer).Claims(claims).Claims(customClaims)
	token, err := builder.Serialize()
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	return token
}

func createTestIDTokenWithoutEmail(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, subject, name string) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: privateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key-id"),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	claims := jwt.Claims{
		Issuer:   issuer,
		Subject:  subject,
		Audience: jwt.Audience{audience},
		Expiry:   jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}

	customClaims := map[string]interface{}{
		"name": name,
	}

	builder := jwt.Signed(signer).Claims(claims).Claims(customClaims)
	token, err := builder.Serialize()
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	return token
}

func TestHandleCallback_ExpiredIDToken(t *testing.T) {
	mockProvider, privateKey := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	idToken := createExpiredIDToken(t, privateKey, mockProvider.URL, cfg.ClientID, "user123", "user@example.com", "Test User")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=test-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	mockProvider.setTokenResponse(idToken)

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func createExpiredIDToken(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience, subject, email, name string) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: privateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key-id"),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	claims := jwt.Claims{
		Issuer:   issuer,
		Subject:  subject,
		Audience: jwt.Audience{audience},
		Expiry:   jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		IssuedAt: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}

	customClaims := map[string]interface{}{
		"email": email,
		"name":  name,
	}

	builder := jwt.Signed(signer).Claims(claims).Claims(customClaims)
	token, err := builder.Serialize()
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	return token
}

func TestHandleCallback_WrongIssuer(t *testing.T) {
	mockProvider, privateKey := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	idToken := createTestIDToken(t, privateKey, "https://evil.com", cfg.ClientID, "user123", "user@example.com", "Test User")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=test-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	mockProvider.setTokenResponse(idToken)

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for wrong issuer, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func TestHandleCallback_WrongAudience(t *testing.T) {
	mockProvider, privateKey := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	idToken := createTestIDToken(t, privateKey, mockProvider.URL, "wrong-client-id", "user123", "user@example.com", "Test User")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=test-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	mockProvider.setTokenResponse(idToken)

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for wrong audience, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}

func TestHandleCallback_EmptyEmailClaim(t *testing.T) {
	mockProvider, privateKey := setupMockOIDCProviderWithJWKS(t)
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

	stateValue := "test-state-123"

	idToken := createTestIDToken(t, privateKey, mockProvider.URL, cfg.ClientID, "user123", "", "Test User")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?code=test-code&state=%s", stateValue), nil)
	r.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: stateValue,
	})

	mockProvider.setTokenResponse(idToken)

	result, err := auth.HandleCallback(w, r)
	if err == nil {
		t.Fatal("Expected error for empty email claim, got nil")
	}

	if result != nil {
		t.Error("Expected nil result for error case")
	}
}
