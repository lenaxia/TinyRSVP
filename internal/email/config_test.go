package email

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestConfig_LoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid configuration with all fields",
			envVars: map[string]string{
				"SMTP_HOST":             "smtp.gmail.com",
				"SMTP_PORT":             "587",
				"SMTP_USERNAME":         "user@gmail.com",
				"SMTP_PASSWORD":         "secret123",
				"EMAIL_FROM":            "noreply@example.com",
				"EMAIL_FROM_NAME":       "TinyRSVP",
				"SMTP_TLS":              "true",
				"SMTP_SKIP_VERIFY":      "false",
				"SMTP_TIMEOUT":          "45s",
				"EMAIL_RATE_LIMIT":      "100",
				"EMAIL_TEST_ON_STARTUP": "false",
				"MAX_RETRY_ATTEMPTS":    "5",
				"QUEUE_POLL_INTERVAL":   "30s",
				"QUEUE_BATCH_SIZE":      "25",
			},
			want: &Config{
				SMTPHost:          "smtp.gmail.com",
				SMTPPort:          587,
				SMTPUsername:      "user@gmail.com",
				SMTPPassword:      "secret123",
				FromEmail:         "noreply@example.com",
				FromName:          "TinyRSVP",
				UseTLS:            true,
				SkipVerify:        false,
				Timeout:           45 * time.Second,
				MaxConnections:    10,
				RateLimit:         100,
				QueuePollInterval: 30 * time.Second,
				QueueBatchSize:    25,
				MaxRetryAttempts:  5,
				TestOnStartup:     false,
			},
			wantErr: false,
		},
		{
			name: "minimal valid configuration with defaults",
			envVars: map[string]string{
				"SMTP_HOST":  "smtp.example.com",
				"EMAIL_FROM": "test@example.com",
			},
			want: &Config{
				SMTPHost:          "smtp.example.com",
				SMTPPort:          587,
				FromEmail:         "test@example.com",
				UseTLS:            true,
				SkipVerify:        false,
				Timeout:           30 * time.Second,
				MaxConnections:    10,
				RateLimit:         50,
				QueuePollInterval: 60 * time.Second,
				QueueBatchSize:    50,
				MaxRetryAttempts:  4,
				TestOnStartup:     true,
			},
			wantErr: false,
		},
		{
			name: "missing required host",
			envVars: map[string]string{
				"EMAIL_FROM": "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing required from email",
			envVars: map[string]string{
				"SMTP_HOST": "smtp.example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"SMTP_HOST":  "smtp.example.com",
				"SMTP_PORT":  "invalid",
				"EMAIL_FROM": "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			envVars: map[string]string{
				"SMTP_HOST":    "smtp.example.com",
				"EMAIL_FROM":   "test@example.com",
				"SMTP_TIMEOUT": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit",
			envVars: map[string]string{
				"SMTP_HOST":        "smtp.example.com",
				"EMAIL_FROM":       "test@example.com",
				"EMAIL_RATE_LIMIT": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEmailEnv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEmailEnv()

			got, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.SMTPHost != tt.want.SMTPHost {
				t.Errorf("SMTPHost = %v, want %v", got.SMTPHost, tt.want.SMTPHost)
			}
			if got.SMTPPort != tt.want.SMTPPort {
				t.Errorf("SMTPPort = %v, want %v", got.SMTPPort, tt.want.SMTPPort)
			}
			if got.FromEmail != tt.want.FromEmail {
				t.Errorf("FromEmail = %v, want %v", got.FromEmail, tt.want.FromEmail)
			}
			if got.RateLimit != tt.want.RateLimit {
				t.Errorf("RateLimit = %v, want %v", got.RateLimit, tt.want.RateLimit)
			}
			if got.Timeout != tt.want.Timeout {
				t.Errorf("Timeout = %v, want %v", got.Timeout, tt.want.Timeout)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr string
	}{
		{
			name: "valid config",
			config: &Config{
				SMTPHost:         "smtp.example.com",
				SMTPPort:         587,
				FromEmail:        "test@example.com",
				Timeout:          30 * time.Second,
				RateLimit:        50,
				MaxRetryAttempts: 4,
			},
			wantErr: "",
		},
		{
			name: "missing host",
			config: &Config{
				SMTPPort:  587,
				FromEmail: "test@example.com",
			},
			wantErr: "SMTP_HOST is required",
		},
		{
			name: "invalid port - zero",
			config: &Config{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  0,
				FromEmail: "test@example.com",
			},
			wantErr: "SMTP_PORT must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: &Config{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  99999,
				FromEmail: "test@example.com",
			},
			wantErr: "SMTP_PORT must be between 1 and 65535",
		},
		{
			name: "missing from email",
			config: &Config{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
			},
			wantErr: "EMAIL_FROM is required",
		},
		{
			name: "invalid email format",
			config: &Config{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  587,
				FromEmail: "not-an-email",
			},
			wantErr: "EMAIL_FROM is not a valid email address",
		},
		{
			name: "negative timeout",
			config: &Config{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  587,
				FromEmail: "test@example.com",
				Timeout:   -1 * time.Second,
			},
			wantErr: "SMTP_TIMEOUT must be positive",
		},
		{
			name: "zero rate limit",
			config: &Config{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  587,
				FromEmail: "test@example.com",
				Timeout:   30 * time.Second,
				RateLimit: 0,
			},
			wantErr: "EMAIL_RATE_LIMIT must be positive",
		},
		{
			name: "retry attempts too low",
			config: &Config{
				SMTPHost:         "smtp.example.com",
				SMTPPort:         587,
				FromEmail:        "test@example.com",
				Timeout:          30 * time.Second,
				RateLimit:        50,
				MaxRetryAttempts: 0,
			},
			wantErr: "MAX_RETRY_ATTEMPTS must be between 1 and 10",
		},
		{
			name: "retry attempts too high",
			config: &Config{
				SMTPHost:         "smtp.example.com",
				SMTPPort:         587,
				FromEmail:        "test@example.com",
				Timeout:          30 * time.Second,
				RateLimit:        50,
				MaxRetryAttempts: 11,
			},
			wantErr: "MAX_RETRY_ATTEMPTS must be between 1 and 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, want %v", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Validate() error = %v, want to contain %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestConfig_Sanitized(t *testing.T) {
	config := &Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "secret123",
		FromEmail:    "test@example.com",
		FromName:     "Test",
	}

	sanitized := config.Sanitized()

	if sanitized.SMTPPassword != "***REDACTED***" {
		t.Errorf("Sanitized password = %v, want ***REDACTED***", sanitized.SMTPPassword)
	}

	if config.SMTPPassword != "secret123" {
		t.Error("Original config was modified")
	}

	if sanitized.SMTPHost != config.SMTPHost {
		t.Error("Other fields should not be modified")
	}
	if sanitized.SMTPUsername != config.SMTPUsername {
		t.Error("Username should not be modified")
	}
}

func TestConfig_Sanitized_EmptyPassword(t *testing.T) {
	config := &Config{
		SMTPHost:  "smtp.example.com",
		SMTPPort:  587,
		FromEmail: "test@example.com",
	}

	sanitized := config.Sanitized()

	if sanitized.SMTPPassword != "" {
		t.Errorf("Empty password should remain empty, got %v", sanitized.SMTPPassword)
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid simple", "test@example.com", true},
		{"valid with subdomain", "user@mail.example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with dash", "user-name@example.com", true},
		{"valid with dot", "user.name@example.com", true},
		{"invalid no @", "notanemail", false},
		{"invalid no domain", "user@", false},
		{"invalid no local", "@example.com", false},
		{"invalid spaces", "user @example.com", false},
		{"invalid multiple @", "user@@example.com", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func clearEmailEnv() {
	envVars := []string{
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"EMAIL_FROM",
		"EMAIL_FROM_NAME",
		"SMTP_TLS",
		"SMTP_SKIP_VERIFY",
		"SMTP_TIMEOUT",
		"EMAIL_RATE_LIMIT",
		"EMAIL_TEST_ON_STARTUP",
		"MAX_RETRY_ATTEMPTS",
		"QUEUE_POLL_INTERVAL",
		"QUEUE_BATCH_SIZE",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
