package config

import (
	"strings"
	"testing"
	"time"
)

func TestConfig_Validate_Server(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid server config",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
				BaseURL:      "http://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "invalid port - zero",
			config: ServerConfig{
				Port:         0,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
				BaseURL:      "http://localhost:8080",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "invalid port - too high",
			config: ServerConfig{
				Port:         99999,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
				BaseURL:      "http://localhost:8080",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "invalid read timeout",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  0,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
				BaseURL:      "http://localhost:8080",
			},
			wantErr: true,
			errMsg:  "read timeout",
		},
		{
			name: "empty base URL",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
				BaseURL:      "",
			},
			wantErr: true,
			errMsg:  "base URL",
		},
		{
			name: "invalid base URL",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
				BaseURL:      "ht!tp://invalid",
			},
			wantErr: true,
			errMsg:  "base URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Server: tt.config}
			err := cfg.validateServer()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Validate_Database(t *testing.T) {
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sqlite config",
			config: DatabaseConfig{
				Type:         "sqlite",
				Path:         "/tmp/test.db",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: false,
		},
		{
			name: "valid postgres config",
			config: DatabaseConfig{
				Type:         "postgres",
				Path:         "postgres://localhost/db",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: false,
		},
		{
			name: "invalid database type",
			config: DatabaseConfig{
				Type:         "mysql",
				Path:         "/tmp/test.db",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
			errMsg:  "database type",
		},
		{
			name: "sqlite missing path",
			config: DatabaseConfig{
				Type:         "sqlite",
				Path:         "",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
			errMsg:  "path",
		},
		{
			name: "invalid max open conns",
			config: DatabaseConfig{
				Type:         "sqlite",
				Path:         "/tmp/test.db",
				MaxOpenConns: 0,
				MaxIdleConns: 5,
			},
			wantErr: true,
			errMsg:  "max open",
		},
		{
			name: "max idle exceeds max open",
			config: DatabaseConfig{
				Type:         "sqlite",
				Path:         "/tmp/test.db",
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantErr: true,
			errMsg:  "idle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Database: tt.config}
			err := cfg.validateDatabase()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Validate_OIDC(t *testing.T) {
	tests := []struct {
		name    string
		config  OIDCConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "disabled OIDC",
			config: OIDCConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid OIDC config",
			config: OIDCConfig{
				Enabled:      true,
				IssuerURL:    "https://auth.example.com",
				ClientID:     "client123",
				ClientSecret: "secret456",
				RedirectURL:  "https://app.example.com/callback",
			},
			wantErr: false,
		},
		{
			name: "missing issuer URL",
			config: OIDCConfig{
				Enabled:      true,
				IssuerURL:    "",
				ClientID:     "client123",
				ClientSecret: "secret456",
				RedirectURL:  "https://app.example.com/callback",
			},
			wantErr: true,
			errMsg:  "issuer URL",
		},
		{
			name: "non-HTTPS issuer URL",
			config: OIDCConfig{
				Enabled:      true,
				IssuerURL:    "http://auth.example.com",
				ClientID:     "client123",
				ClientSecret: "secret456",
				RedirectURL:  "https://app.example.com/callback",
			},
			wantErr: true,
			errMsg:  "HTTPS",
		},
		{
			name: "missing client ID",
			config: OIDCConfig{
				Enabled:      true,
				IssuerURL:    "https://auth.example.com",
				ClientID:     "",
				ClientSecret: "secret456",
				RedirectURL:  "https://app.example.com/callback",
			},
			wantErr: true,
			errMsg:  "client ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{OIDC: tt.config}
			err := cfg.validateOIDC()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateOIDC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Validate_Email(t *testing.T) {
	tests := []struct {
		name    string
		config  EmailConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid email config",
			config: EmailConfig{
				SMTPHost:              "smtp.example.com",
				SMTPPort:              587,
				FromEmail:             "test@example.com",
				ProcessorBatchSize:    50,
				ProcessorPollInterval: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing SMTP host",
			config: EmailConfig{
				SMTPHost:  "",
				SMTPPort:  587,
				FromEmail: "test@example.com",
			},
			wantErr: true,
			errMsg:  "SMTP host",
		},
		{
			name: "invalid SMTP port",
			config: EmailConfig{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  0,
				FromEmail: "test@example.com",
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "invalid from email",
			config: EmailConfig{
				SMTPHost:  "smtp.example.com",
				SMTPPort:  587,
				FromEmail: "notanemail",
			},
			wantErr: true,
			errMsg:  "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Email: tt.config}
			err := cfg.validateEmail()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Validate_Storage(t *testing.T) {
	tests := []struct {
		name    string
		config  StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid local storage",
			config: StorageConfig{
				Type:      "local",
				LocalPath: "/data/uploads",
			},
			wantErr: false,
		},
		{
			name: "valid s3 storage",
			config: StorageConfig{
				Type:     "s3",
				S3Bucket: "my-bucket",
				S3Region: "us-west-2",
			},
			wantErr: false,
		},
		{
			name: "invalid storage type",
			config: StorageConfig{
				Type: "ftp",
			},
			wantErr: true,
			errMsg:  "storage type",
		},
		{
			name: "local storage missing path",
			config: StorageConfig{
				Type:      "local",
				LocalPath: "",
			},
			wantErr: true,
			errMsg:  "local path",
		},
		{
			name: "s3 storage missing bucket",
			config: StorageConfig{
				Type:     "s3",
				S3Bucket: "",
				S3Region: "us-west-2",
			},
			wantErr: true,
			errMsg:  "bucket",
		},
		{
			name: "s3 storage missing region",
			config: StorageConfig{
				Type:     "s3",
				S3Bucket: "my-bucket",
				S3Region: "",
			},
			wantErr: true,
			errMsg:  "region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Storage: tt.config}
			err := cfg.validateStorage()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_Validate_Security(t *testing.T) {
	tests := []struct {
		name    string
		config  SecurityConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid security config",
			config: SecurityConfig{
				SessionDuration: 168 * time.Hour,
				TokenExpiry:     720 * time.Hour,
				HMACSecretKey:   "my-secret-key-exactly-32-bytes!!",
			},
			wantErr: false,
		},
		{
			name: "session duration too short",
			config: SecurityConfig{
				SessionDuration: 30 * time.Minute,
				TokenExpiry:     720 * time.Hour,
				HMACSecretKey:   "my-secret-key-exactly-32-bytes!!",
			},
			wantErr: true,
			errMsg:  "session duration",
		},
		{
			name: "session duration too long",
			config: SecurityConfig{
				SessionDuration: 800 * time.Hour,
				TokenExpiry:     720 * time.Hour,
				HMACSecretKey:   "my-secret-key-exactly-32-bytes!!",
			},
			wantErr: true,
			errMsg:  "session duration",
		},
		{
			name: "token expiry too short",
			config: SecurityConfig{
				SessionDuration: 168 * time.Hour,
				TokenExpiry:     12 * time.Hour,
				HMACSecretKey:   "my-secret-key-exactly-32-bytes!!",
			},
			wantErr: true,
			errMsg:  "token expiry",
		},
		{
			name: "HMAC secret too short",
			config: SecurityConfig{
				SessionDuration: 168 * time.Hour,
				TokenExpiry:     720 * time.Hour,
				HMACSecretKey:   "short",
			},
			wantErr: true,
			errMsg:  "HMAC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Security: tt.config}
			err := cfg.validateSecurity()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %v", tt.errMsg, err)
			}
		})
	}
}

func TestConfig_String_MasksSensitiveData(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host:    "localhost",
			Port:    8080,
			BaseURL: "http://localhost:8080",
		},
		OIDC: OIDCConfig{
			Enabled:      true,
			ClientSecret: "super-secret-client-secret",
		},
		Email: EmailConfig{
			SMTPPassword: "super-secret-password",
			SMTPUser:     "user@example.com",
		},
		Security: SecurityConfig{
			HMACSecretKey: "super-secret-hmac-key-32-bytes!",
		},
	}

	result := cfg.String()

	if strings.Contains(result, "super-secret-client-secret") {
		t.Error("ClientSecret should be masked")
	}

	if strings.Contains(result, "super-secret-password") {
		t.Error("SMTPPassword should be masked")
	}

	if strings.Contains(result, "super-secret-hmac-key-32-bytes!") {
		t.Error("HMACSecretKey should be masked")
	}

	if !strings.Contains(result, "***") {
		t.Error("Expected masked values to contain ***")
	}

	if !strings.Contains(result, "localhost") {
		t.Error("Non-sensitive data should be visible")
	}

	if !strings.Contains(result, "user@example.com") {
		t.Error("Non-sensitive data should be visible")
	}
}

func TestConfig_SetDefaults_GeneratesHMACSecret(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			HMACSecretKey: "",
		},
	}

	cfg.setDefaults()

	if cfg.Security.HMACSecretKey == "" {
		t.Error("Expected HMAC secret to be generated")
	}

	if len(cfg.Security.HMACSecretKey) < 32 {
		t.Errorf("Generated HMAC secret too short: %d bytes", len(cfg.Security.HMACSecretKey))
	}
}

func TestConfig_SetDefaults_PreservesExistingHMACSecret(t *testing.T) {
	originalSecret := "my-existing-secret-key-32-bytes!"
	cfg := &Config{
		Security: SecurityConfig{
			HMACSecretKey: originalSecret,
		},
	}

	cfg.setDefaults()

	if cfg.Security.HMACSecretKey != originalSecret {
		t.Error("Existing HMAC secret should not be overwritten")
	}
}
