package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	OIDC        OIDCConfig
	ForwardAuth ForwardAuthConfig
	Email       EmailConfig
	Storage     StorageConfig
	Security    SecurityConfig
	Token       TokenConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BaseURL      string
}

type DatabaseConfig struct {
	Type         string
	Path         string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type OIDCConfig struct {
	Enabled      bool
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type ForwardAuthConfig struct {
	Enabled     bool
	UserHeader  string
	EmailHeader string
	TrustedIPs  []string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type StorageConfig struct {
	Type       string
	LocalPath  string
	S3Bucket   string
	S3Region   string
	S3Endpoint string
}

type SecurityConfig struct {
	SessionDuration time.Duration
	TokenExpiry     time.Duration
	HMACSecretKey   string
}

type TokenConfig struct {
	Secret string
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := cfg.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg.setDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadFromEnv() error {
	var err error

	c.Server.Host = getEnvString("SERVER_HOST", "0.0.0.0")
	c.Server.Port, err = getEnvInt("SERVER_PORT", 8080)
	if err != nil {
		return fmt.Errorf("SERVER_PORT: %w", err)
	}

	c.Server.ReadTimeout, err = getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return fmt.Errorf("SERVER_READ_TIMEOUT: %w", err)
	}

	c.Server.WriteTimeout, err = getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return fmt.Errorf("SERVER_WRITE_TIMEOUT: %w", err)
	}

	c.Server.IdleTimeout, err = getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second)
	if err != nil {
		return fmt.Errorf("SERVER_IDLE_TIMEOUT: %w", err)
	}

	c.Server.BaseURL = getEnvString("SERVER_BASE_URL", "")
	if c.Server.BaseURL == "" {
		return fmt.Errorf("SERVER_BASE_URL is required")
	}

	c.Database.Type = getEnvString("DATABASE_TYPE", "sqlite")
	c.Database.Path = getEnvString("DATABASE_PATH", "")
	if c.Database.Path == "" {
		return fmt.Errorf("DATABASE_PATH is required")
	}

	c.Database.MaxOpenConns, err = getEnvInt("DATABASE_MAX_OPEN_CONNS", 25)
	if err != nil {
		return fmt.Errorf("DATABASE_MAX_OPEN_CONNS: %w", err)
	}

	c.Database.MaxIdleConns, err = getEnvInt("DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNS: %w", err)
	}

	c.Database.MaxLifetime, err = getEnvDuration("DATABASE_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return fmt.Errorf("DATABASE_MAX_LIFETIME: %w", err)
	}

	c.OIDC.Enabled, err = getEnvBool("OIDC_ENABLED", false)
	if err != nil {
		return fmt.Errorf("OIDC_ENABLED: %w", err)
	}

	c.OIDC.IssuerURL = getEnvString("OIDC_ISSUER_URL", "")
	c.OIDC.ClientID = getEnvString("OIDC_CLIENT_ID", "")
	c.OIDC.ClientSecret = getEnvString("OIDC_CLIENT_SECRET", "")
	c.OIDC.RedirectURL = getEnvString("OIDC_REDIRECT_URL", "")

	if err := c.loadForwardAuthFromEnv(); err != nil {
		return err
	}

	c.Email.SMTPHost = getEnvString("SMTP_HOST", "")
	if c.Email.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}

	c.Email.SMTPPort, err = getEnvInt("SMTP_PORT", 587)
	if err != nil {
		return fmt.Errorf("SMTP_PORT: %w", err)
	}

	c.Email.SMTPUser = getEnvString("SMTP_USER", "")
	c.Email.SMTPPassword = getEnvString("SMTP_PASSWORD", "")
	c.Email.FromEmail = getEnvString("EMAIL_FROM", "")
	if c.Email.FromEmail == "" {
		return fmt.Errorf("EMAIL_FROM is required")
	}

	c.Email.FromName = getEnvString("EMAIL_FROM_NAME", "TinyRSVP")

	c.Storage.Type = getEnvString("STORAGE_TYPE", "local")
	c.Storage.LocalPath = getEnvString("STORAGE_LOCAL_PATH", "/data/uploads")
	c.Storage.S3Bucket = getEnvString("STORAGE_S3_BUCKET", "")
	c.Storage.S3Region = getEnvString("STORAGE_S3_REGION", "")
	c.Storage.S3Endpoint = getEnvString("STORAGE_S3_ENDPOINT", "")

	c.Security.SessionDuration, err = getEnvDuration("SECURITY_SESSION_DURATION", 168*time.Hour)
	if err != nil {
		return fmt.Errorf("SECURITY_SESSION_DURATION: %w", err)
	}

	c.Security.TokenExpiry, err = getEnvDuration("SECURITY_TOKEN_EXPIRY", 720*time.Hour)
	if err != nil {
		return fmt.Errorf("SECURITY_TOKEN_EXPIRY: %w", err)
	}

	c.Security.HMACSecretKey = getEnvString("SECURITY_HMAC_SECRET", "")

	c.Token.Secret = getEnvString("TOKEN_SECRET", "")

	return nil
}

func (c *Config) setDefaults() {
	if c.Security.HMACSecretKey == "" {
		c.Security.HMACSecretKey = generateHMACSecret()
	}
	if c.Token.Secret == "" {
		c.Token.Secret = generateHMACSecret()
	}
}

func (c *Config) Validate() error {
	if err := c.validateServer(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := c.validateDatabase(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}

	if err := c.validateOIDC(); err != nil {
		return fmt.Errorf("OIDC config: %w", err)
	}

	if err := c.validateForwardAuth(); err != nil {
		return fmt.Errorf("forward auth config: %w", err)
	}

	if err := c.validateAuthModes(); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	if err := c.validateEmail(); err != nil {
		return fmt.Errorf("email config: %w", err)
	}

	if err := c.validateStorage(); err != nil {
		return fmt.Errorf("storage config: %w", err)
	}

	if err := c.validateSecurity(); err != nil {
		return fmt.Errorf("security config: %w", err)
	}

	if err := c.validateToken(); err != nil {
		return fmt.Errorf("token config: %w", err)
	}

	return nil
}

func (c *Config) validateServer() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be > 0")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be > 0")
	}

	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be > 0")
	}

	if c.Server.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if _, err := url.Parse(c.Server.BaseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	return nil
}

func (c *Config) validateDatabase() error {
	if c.Database.Type != "sqlite" && c.Database.Type != "postgres" {
		return fmt.Errorf("invalid database type: %s (must be sqlite or postgres)", c.Database.Type)
	}

	if c.Database.Type == "sqlite" && c.Database.Path == "" {
		return fmt.Errorf("database path is required for SQLite")
	}

	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be > 0")
	}

	if c.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be > 0")
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("max idle connections (%d) cannot exceed max open connections (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}

	return nil
}

func (c *Config) validateOIDC() error {
	if !c.OIDC.Enabled {
		return nil
	}

	if c.OIDC.IssuerURL == "" {
		return fmt.Errorf("issuer URL is required when OIDC is enabled")
	}

	if _, err := url.Parse(c.OIDC.IssuerURL); err != nil {
		return fmt.Errorf("invalid issuer URL: %w", err)
	}

	if !strings.HasPrefix(c.OIDC.IssuerURL, "https://") {
		return fmt.Errorf("issuer URL must use HTTPS")
	}

	if c.OIDC.ClientID == "" {
		return fmt.Errorf("client ID is required when OIDC is enabled")
	}

	if c.OIDC.ClientSecret == "" {
		return fmt.Errorf("client secret is required when OIDC is enabled")
	}

	if c.OIDC.RedirectURL == "" {
		return fmt.Errorf("redirect URL is required when OIDC is enabled")
	}

	if _, err := url.Parse(c.OIDC.RedirectURL); err != nil {
		return fmt.Errorf("invalid redirect URL: %w", err)
	}

	return nil
}

func (c *Config) validateEmail() error {
	if c.Email.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}

	if c.Email.SMTPPort < 1 || c.Email.SMTPPort > 65535 {
		return fmt.Errorf("invalid SMTP port: %d (must be 1-65535)", c.Email.SMTPPort)
	}

	if c.Email.FromEmail == "" {
		return fmt.Errorf("from email is required")
	}

	if !strings.Contains(c.Email.FromEmail, "@") {
		return fmt.Errorf("invalid from email format")
	}

	return nil
}

func (c *Config) validateStorage() error {
	if c.Storage.Type != "local" && c.Storage.Type != "s3" {
		return fmt.Errorf("invalid storage type: %s (must be local or s3)", c.Storage.Type)
	}

	if c.Storage.Type == "local" && c.Storage.LocalPath == "" {
		return fmt.Errorf("local path is required for local storage")
	}

	if c.Storage.Type == "s3" {
		if c.Storage.S3Bucket == "" {
			return fmt.Errorf("S3 bucket is required for S3 storage")
		}
		if c.Storage.S3Region == "" {
			return fmt.Errorf("S3 region is required for S3 storage")
		}
	}

	return nil
}

func (c *Config) validateSecurity() error {
	if c.Security.SessionDuration < 1*time.Hour {
		return fmt.Errorf("session duration must be >= 1h")
	}

	if c.Security.SessionDuration > 720*time.Hour {
		return fmt.Errorf("session duration must be <= 720h (30 days)")
	}

	if c.Security.TokenExpiry < 24*time.Hour {
		return fmt.Errorf("token expiry must be >= 24h")
	}

	if c.Security.TokenExpiry > 8760*time.Hour {
		return fmt.Errorf("token expiry must be <= 8760h (365 days)")
	}

	if len(c.Security.HMACSecretKey) < 32 {
		return fmt.Errorf("HMAC secret key must be at least 32 bytes")
	}

	return nil
}

func (c *Config) validateToken() error {
	if len(c.Token.Secret) < 32 {
		return fmt.Errorf("token secret must be at least 32 bytes")
	}

	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf(`Config{
  Server: {Host: %s, Port: %d, BaseURL: %s}
  Database: {Type: %s, Path: %s, MaxOpenConns: %d, MaxIdleConns: %d}
  OIDC: {Enabled: %t, IssuerURL: %s, ClientID: %s, ClientSecret: ***}
  Email: {SMTPHost: %s, SMTPPort: %d, SMTPUser: %s, SMTPPassword: ***, FromEmail: %s}
  Storage: {Type: %s, LocalPath: %s, S3Bucket: %s, S3Region: %s}
  Security: {SessionDuration: %s, TokenExpiry: %s, HMACSecretKey: ***}
  Token: {Secret: ***}
}`,
		c.Server.Host, c.Server.Port, c.Server.BaseURL,
		c.Database.Type, c.Database.Path, c.Database.MaxOpenConns, c.Database.MaxIdleConns,
		c.OIDC.Enabled, c.OIDC.IssuerURL, c.OIDC.ClientID,
		c.Email.SMTPHost, c.Email.SMTPPort, c.Email.SMTPUser, c.Email.FromEmail,
		c.Storage.Type, c.Storage.LocalPath, c.Storage.S3Bucket, c.Storage.S3Region,
		c.Security.SessionDuration, c.Security.TokenExpiry,
	)
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %w", err)
	}

	return intValue, nil
}

func getEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %w", err)
	}

	return duration, nil
}

func getEnvBool(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value: %w", err)
	}

	return boolValue, nil
}

func generateHMACSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate HMAC secret: %v", err))
	}
	return hex.EncodeToString(bytes)
}
