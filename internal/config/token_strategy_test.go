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

func TestConfig_TokenStrategy_HMAC_WithHardcodedFallback(t *testing.T) {
	env := map[string]string{
		"SERVER_PORT":     "8080",
		"DATABASE_PATH":   "/tmp/test.db",
		"SMTP_HOST":       "localhost",
		"EMAIL_FROM":      "test@example.com",
		"SERVER_BASE_URL": "http://localhost:8080",
	}
	setTestEnv(t, env)

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg1, err := Load()
	if err != nil {
		t.Fatalf("First Load() error = %v, want nil", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "WARNING: TOKEN_SECRET not set") {
		t.Error("Expected warning about TOKEN_SECRET not set")
	}

	if !strings.Contains(output, "using hardcoded fallback") {
		t.Error("Expected warning about hardcoded fallback")
	}

	r2, w2, _ := os.Pipe()
	os.Stderr = w2

	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	w2.Close()
	os.Stderr = oldStderr

	var buf2 bytes.Buffer
	buf2.ReadFrom(r2)

	if cfg1.Token.Secret != cfg2.Token.Secret {
		t.Error("Hardcoded fallback secret should be consistent across loads")
	}

	if len(cfg1.Token.Secret) < 32 {
		t.Errorf("Hardcoded fallback secret length = %d, want >= 32", len(cfg1.Token.Secret))
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
	}
	setTestEnv(t, env)

	oldStderr := os.Stderr
	devNull, _ := os.Open(os.DevNull)
	os.Stderr = devNull
	defer func() {
		os.Stderr = oldStderr
		devNull.Close()
	}()

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
