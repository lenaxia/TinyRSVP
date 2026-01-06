package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestConfig_Load_ValidMinimalConfig(t *testing.T) {
	env := map[string]string{
		"SERVER_PORT":     "8080",
		"DATABASE_PATH":   "/tmp/test.db",
		"SMTP_HOST":       "localhost",
		"EMAIL_FROM":      "test@example.com",
		"SERVER_BASE_URL": "http://localhost:8080",
	}
	setTestEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("Database.Path = %s, want /tmp/test.db", cfg.Database.Path)
	}

	if cfg.Email.SMTPHost != "localhost" {
		t.Errorf("Email.SMTPHost = %s, want localhost", cfg.Email.SMTPHost)
	}

	if cfg.Email.FromEmail != "test@example.com" {
		t.Errorf("Email.FromEmail = %s, want test@example.com", cfg.Email.FromEmail)
	}
}

func TestConfig_Load_MissingRequiredField(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		errMsg string
	}{
		{
			name: "missing DATABASE_PATH",
			env: map[string]string{
				"SERVER_PORT":     "8080",
				"SERVER_BASE_URL": "http://localhost:8080",
				"SMTP_HOST":       "localhost",
				"EMAIL_FROM":      "test@example.com",
			},
			errMsg: "DATABASE_PATH",
		},
		{
			name: "missing SMTP_HOST",
			env: map[string]string{
				"SERVER_PORT":     "8080",
				"SERVER_BASE_URL": "http://localhost:8080",
				"DATABASE_PATH":   "/tmp/test.db",
				"EMAIL_FROM":      "test@example.com",
			},
			errMsg: "SMTP_HOST",
		},
		{
			name: "missing EMAIL_FROM",
			env: map[string]string{
				"SERVER_PORT":     "8080",
				"SERVER_BASE_URL": "http://localhost:8080",
				"DATABASE_PATH":   "/tmp/test.db",
				"SMTP_HOST":       "localhost",
			},
			errMsg: "EMAIL_FROM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestEnv(t, tt.env)

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Load_InvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		errMsg string
	}{
		{
			name: "invalid port - too high",
			env: map[string]string{
				"SERVER_PORT":     "99999",
				"SERVER_BASE_URL": "http://localhost:8080",
				"DATABASE_PATH":   "/tmp/test.db",
				"SMTP_HOST":       "localhost",
				"EMAIL_FROM":      "test@example.com",
			},
			errMsg: "port",
		},
		{
			name: "invalid port - zero",
			env: map[string]string{
				"SERVER_PORT":     "0",
				"SERVER_BASE_URL": "http://localhost:8080",
				"DATABASE_PATH":   "/tmp/test.db",
				"SMTP_HOST":       "localhost",
				"EMAIL_FROM":      "test@example.com",
			},
			errMsg: "port",
		},
		{
			name: "invalid port - negative",
			env: map[string]string{
				"SERVER_PORT":     "-1",
				"SERVER_BASE_URL": "http://localhost:8080",
				"DATABASE_PATH":   "/tmp/test.db",
				"SMTP_HOST":       "localhost",
				"EMAIL_FROM":      "test@example.com",
			},
			errMsg: "port",
		},
		{
			name: "invalid duration",
			env: map[string]string{
				"SERVER_PORT":         "8080",
				"SERVER_BASE_URL":     "http://localhost:8080",
				"DATABASE_PATH":       "/tmp/test.db",
				"SMTP_HOST":           "localhost",
				"EMAIL_FROM":          "test@example.com",
				"SERVER_READ_TIMEOUT": "invalid",
			},
			errMsg: "duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestEnv(t, tt.env)

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Load_WithAllFields(t *testing.T) {
	env := map[string]string{
		"SERVER_HOST":               "127.0.0.1",
		"SERVER_PORT":               "9090",
		"SERVER_READ_TIMEOUT":       "15s",
		"SERVER_WRITE_TIMEOUT":      "20s",
		"SERVER_IDLE_TIMEOUT":       "180s",
		"SERVER_BASE_URL":           "https://rsvp.example.com",
		"DATABASE_TYPE":             "sqlite",
		"DATABASE_PATH":             "/data/app.db",
		"DATABASE_MAX_OPEN_CONNS":   "50",
		"DATABASE_MAX_IDLE_CONNS":   "10",
		"DATABASE_MAX_LIFETIME":     "10m",
		"OIDC_ENABLED":              "true",
		"OIDC_ISSUER_URL":           "https://auth.example.com",
		"OIDC_CLIENT_ID":            "client123",
		"OIDC_CLIENT_SECRET":        "secret456",
		"OIDC_REDIRECT_URL":         "https://rsvp.example.com/callback",
		"SMTP_HOST":                 "smtp.gmail.com",
		"SMTP_PORT":                 "587",
		"SMTP_USER":                 "user@example.com",
		"SMTP_PASSWORD":             "pass123",
		"EMAIL_FROM":                "noreply@example.com",
		"EMAIL_FROM_NAME":           "MyApp",
		"STORAGE_TYPE":              "s3",
		"STORAGE_LOCAL_PATH":        "/data/uploads",
		"STORAGE_S3_BUCKET":         "my-bucket",
		"STORAGE_S3_REGION":         "us-west-2",
		"STORAGE_S3_ENDPOINT":       "https://s3.amazonaws.com",
		"SECURITY_SESSION_DURATION": "240h",
		"SECURITY_TOKEN_EXPIRY":     "168h",
		"SECURITY_HMAC_SECRET":      "my-secret-key-exactly-32-bytes!!",
	}
	setTestEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %s, want 127.0.0.1", cfg.Server.Host)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}

	if cfg.Server.ReadTimeout != 15*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want 15s", cfg.Server.ReadTimeout)
	}

	if cfg.Database.Type != "sqlite" {
		t.Errorf("Database.Type = %s, want sqlite", cfg.Database.Type)
	}

	if cfg.Database.MaxOpenConns != 50 {
		t.Errorf("Database.MaxOpenConns = %d, want 50", cfg.Database.MaxOpenConns)
	}

	if !cfg.OIDC.Enabled {
		t.Error("OIDC.Enabled = false, want true")
	}

	if cfg.OIDC.IssuerURL != "https://auth.example.com" {
		t.Errorf("OIDC.IssuerURL = %s, want https://auth.example.com", cfg.OIDC.IssuerURL)
	}

	if cfg.Email.SMTPPort != 587 {
		t.Errorf("Email.SMTPPort = %d, want 587", cfg.Email.SMTPPort)
	}

	if cfg.Storage.Type != "s3" {
		t.Errorf("Storage.Type = %s, want s3", cfg.Storage.Type)
	}

	if cfg.Security.SessionDuration != 240*time.Hour {
		t.Errorf("Security.SessionDuration = %v, want 240h", cfg.Security.SessionDuration)
	}
}

func setTestEnv(t *testing.T, env map[string]string) {
	t.Helper()

	oldEnv := make(map[string]string)
	for k := range env {
		if v, exists := os.LookupEnv(k); exists {
			oldEnv[k] = v
		}
	}

	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("Failed to set env var %s: %v", k, err)
		}
	}

	t.Cleanup(func() {
		for k := range env {
			if oldVal, exists := oldEnv[k]; exists {
				os.Setenv(k, oldVal)
			} else {
				os.Unsetenv(k)
			}
		}
	})
}
