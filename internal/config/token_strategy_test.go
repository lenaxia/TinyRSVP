package config

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestConfig_TokenStrategy_HMAC_WithSecret(t *testing.T) {
	env := map[string]string{
		"SERVER_PORT":     "8080",
		"DATABASE_PATH":   "/tmp/test.db",
		"SMTP_HOST":       "localhost",
		"EMAIL_FROM":      "test@example.com",
		"SERVER_BASE_URL": "http://localhost:8080",
		"TOKEN_SECRET":    "da8f152a3cc3d58054cb988a463344503ad1ad09fba718a8a5e6e9513d16040f",
	}
	setTestEnv(t, env)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if strings.Contains(output, "WARNING") {
		t.Errorf("Should not show warning when TOKEN_SECRET is set, got:\n%s", output)
	}

	if cfg.Token.Secret != "da8f152a3cc3d58054cb988a463344503ad1ad09fba718a8a5e6e9513d16040f" {
		t.Errorf("Token.Secret = %s, want provided secret", cfg.Token.Secret)
	}
}

func TestConfig_TokenStrategy_HMAC_NoSecret_FailsToStart(t *testing.T) {
	env := map[string]string{
		"SERVER_PORT":     "8080",
		"DATABASE_PATH":   "/tmp/test.db",
		"SMTP_HOST":       "localhost",
		"EMAIL_FROM":      "test@example.com",
		"SERVER_BASE_URL": "http://localhost:8080",
		// TOKEN_SECRET intentionally omitted
	}
	setTestEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail when TOKEN_SECRET is not set, got nil error")
	}
	if !strings.Contains(err.Error(), "TOKEN_SECRET is required") {
		t.Errorf("Expected error about TOKEN_SECRET being required, got: %v", err)
	}
}

func TestConfig_TokenStrategy_PlainTokenMode(t *testing.T) {
	env := map[string]string{
		"SERVER_PORT":           "8080",
		"DATABASE_PATH":         "/tmp/test.db",
		"SMTP_HOST":             "localhost",
		"EMAIL_FROM":            "test@example.com",
		"SERVER_BASE_URL":       "http://localhost:8080",
		"TOKEN_HASHING_ENABLED": "false",
	}
	setTestEnv(t, env)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "WARNING: Token hashing disabled") {
		t.Error("Expected warning about token hashing being disabled")
	}

	if !strings.Contains(output, "tokens will be stored in plain text") {
		t.Error("Expected warning about plain text storage")
	}

	if !cfg.Token.HashingEnabled {
		if cfg.Token.Secret != "" {
			t.Error("Token.Secret should be empty when hashing is disabled")
		}
	}
}

func TestConfig_TokenStrategy_HMACEnabledByDefault(t *testing.T) {
	env := map[string]string{
		"SERVER_PORT":     "8080",
		"DATABASE_PATH":   "/tmp/test.db",
		"SMTP_HOST":       "localhost",
		"EMAIL_FROM":      "test@example.com",
		"SERVER_BASE_URL": "http://localhost:8080",
		"TOKEN_SECRET":    "da8f152a3cc3d58054cb988a463344503ad1ad09fba718a8a5e6e9513d16040f",
	}
	setTestEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if !cfg.Token.HashingEnabled {
		t.Error("Token.HashingEnabled should default to true")
	}

	if cfg.Token.Secret == "" {
		t.Error("Token.Secret should be set when hashing is enabled")
	}
}
