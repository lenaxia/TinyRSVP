package handlers

import (
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/config"
)

func TestConfigToSettingsView_RedactsSecrets(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			BaseURL:      "https://rsvp.example.com",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: config.DatabaseConfig{
			Type:         "sqlite",
			Path:         "/data/tinyrsvp.db",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		OIDC: config.OIDCConfig{
			Enabled:      true,
			IssuerURL:    "https://idp.example.com",
			ClientID:     "my-client-id",
			ClientSecret: "super-secret-value",
			RedirectURL:  "https://rsvp.example.com/auth/oidc/callback",
		},
		Email: config.EmailConfig{
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUser:     "user@example.com",
			SMTPPassword: "smtp-password-secret",
			FromEmail:    "rsvp@example.com",
			FromName:     "TinyRSVP",
		},
		Security: config.SecurityConfig{
			SessionDuration: 7 * 24 * time.Hour,
			TokenExpiry:     24 * time.Hour,
			HMACSecretKey:   "hmac-secret-key",
		},
		Token: config.TokenConfig{
			Secret:         "token-secret",
			HashingEnabled: true,
		},
	}

	view := ConfigToSettingsView(cfg)

	if view.Email.SMTPPasswordSet != true {
		t.Error("SMTPPasswordSet should be true when password is set")
	}

	if view.Security.HMACKeySet != true {
		t.Error("HMACKeySet should be true when HMAC key is set")
	}

	if view.Token.SecretSet != true {
		t.Error("Token.SecretSet should be true when secret is set")
	}

	if view.Auth.Method != "OIDC" {
		t.Errorf("Auth.Method = %q, want OIDC", view.Auth.Method)
	}

	if view.Auth.OIDCClientID != "my-client-id" {
		t.Errorf("OIDCClientID = %q, want my-client-id", view.Auth.OIDCClientID)
	}

	nonSecretFields := []string{
		view.Server.Host,
		view.Database.Path,
		view.Email.SMTPHost,
		view.Storage.Type,
	}
	for _, v := range nonSecretFields {
		if v == "" && v != "" {
			t.Error("non-secret field should be populated")
		}
	}
}

func TestConfigToSettingsView_NoSecretsSet(t *testing.T) {
	cfg := &config.Config{}

	view := ConfigToSettingsView(cfg)

	if view.Email.SMTPPasswordSet != false {
		t.Error("SMTPPasswordSet should be false when password is empty")
	}

	if view.Security.HMACKeySet != false {
		t.Error("HMACKeySet should be false when key is empty")
	}

	if view.Token.SecretSet != false {
		t.Error("Token.SecretSet should be false when secret is empty")
	}

	if view.Auth.Method != "None" {
		t.Errorf("Auth.Method = %q, want None", view.Auth.Method)
	}
}

func TestConfigToSettingsView_ForwardAuthMethod(t *testing.T) {
	cfg := &config.Config{
		ForwardAuth: config.ForwardAuthConfig{
			Enabled: true,
		},
	}

	view := ConfigToSettingsView(cfg)

	if view.Auth.Method != "Forward Auth" {
		t.Errorf("Auth.Method = %q, want Forward Auth", view.Auth.Method)
	}
}

func TestConfigToSettingsView_OIDCClientSecretNeverLeaked(t *testing.T) {
	cfg := &config.Config{
		OIDC: config.OIDCConfig{
			Enabled:      true,
			ClientSecret: "must-not-appear-anywhere",
		},
	}

	view := ConfigToSettingsView(cfg)

	if view.Auth.OIDCClientID == "must-not-appear-anywhere" {
		t.Error("ClientSecret leaked into ClientID field")
	}
}
