package config

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestConfig_TokenSecretWarning_WhenNotSet(t *testing.T) {
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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedWarnings := []string{
		"WARNING: TOKEN_SECRET not set",
		"WARNING: Invite tokens will become invalid after server restart",
		"WARNING: Set TOKEN_SECRET environment variable",
		"WARNING: Generate with: openssl rand -hex 32",
	}

	for _, warning := range expectedWarnings {
		if !strings.Contains(output, warning) {
			t.Errorf("Expected warning containing %q, got:\n%s", warning, output)
		}
	}

	if cfg.Token.Secret == "" {
		t.Error("Token.Secret should be auto-generated when not set")
	}

	if len(cfg.Token.Secret) < 32 {
		t.Errorf("Auto-generated Token.Secret length = %d, want >= 32", len(cfg.Token.Secret))
	}
}

func TestConfig_TokenSecretWarning_WhenSet(t *testing.T) {
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

	if strings.Contains(output, "WARNING: TOKEN_SECRET not set") {
		t.Errorf("Should not show warning when TOKEN_SECRET is set, got:\n%s", output)
	}

	if cfg.Token.Secret != "da8f152a3cc3d58054cb988a463344503ad1ad09fba718a8a5e6e9513d16040f" {
		t.Errorf("Token.Secret = %s, want da8f152a3cc3d58054cb988a463344503ad1ad09fba718a8a5e6e9513d16040f", cfg.Token.Secret)
	}
}

func TestConfig_TokenSecretPersistence(t *testing.T) {
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

	cfg1, err := Load()
	if err != nil {
		t.Fatalf("First Load() error = %v, want nil", err)
	}

	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	if cfg1.Token.Secret == cfg2.Token.Secret {
		t.Error("Auto-generated secrets should be different on each Load() call, simulating server restart")
	}
}
