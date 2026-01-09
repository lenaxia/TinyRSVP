package email

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string

	FromEmail string
	FromName  string

	UseTLS     bool
	SkipVerify bool

	Timeout        time.Duration
	MaxConnections int

	RateLimit int

	QueuePollInterval time.Duration
	QueueBatchSize    int

	MaxRetryAttempts int

	TestOnStartup bool
}

func LoadConfig() (*Config, error) {
	config := &Config{
		SMTPPort:          587,
		UseTLS:            true,
		SkipVerify:        false,
		Timeout:           30 * time.Second,
		MaxConnections:    10,
		RateLimit:         50,
		QueuePollInterval: 60 * time.Second,
		QueueBatchSize:    50,
		MaxRetryAttempts:  4,
		TestOnStartup:     true,
	}

	if host := os.Getenv("SMTP_HOST"); host != "" {
		config.SMTPHost = host
	}

	if port := os.Getenv("SMTP_PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
		}
		config.SMTPPort = p
	}

	if username := os.Getenv("SMTP_USERNAME"); username != "" {
		config.SMTPUsername = username
	}

	if password := os.Getenv("SMTP_PASSWORD"); password != "" {
		config.SMTPPassword = password
	}

	if from := os.Getenv("SMTP_FROM_EMAIL"); from != "" {
		config.FromEmail = from
	}

	if name := os.Getenv("SMTP_FROM_NAME"); name != "" {
		config.FromName = name
	}

	if tls := os.Getenv("SMTP_TLS"); tls != "" {
		config.UseTLS = tls == "true"
	}

	if skip := os.Getenv("SMTP_SKIP_VERIFY"); skip != "" {
		config.SkipVerify = skip == "true"
	}

	if timeout := os.Getenv("SMTP_TIMEOUT"); timeout != "" {
		t, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_TIMEOUT: %w", err)
		}
		config.Timeout = t
	}

	if limit := os.Getenv("EMAIL_RATE_LIMIT"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, fmt.Errorf("invalid EMAIL_RATE_LIMIT: %w", err)
		}
		config.RateLimit = l
	}

	if test := os.Getenv("EMAIL_TEST_ON_STARTUP"); test != "" {
		config.TestOnStartup = test == "true"
	}

	if maxRetry := os.Getenv("MAX_RETRY_ATTEMPTS"); maxRetry != "" {
		m, err := strconv.Atoi(maxRetry)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_RETRY_ATTEMPTS: %w", err)
		}
		config.MaxRetryAttempts = m
	}

	if pollInterval := os.Getenv("QUEUE_POLL_INTERVAL"); pollInterval != "" {
		p, err := time.ParseDuration(pollInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid QUEUE_POLL_INTERVAL: %w", err)
		}
		config.QueuePollInterval = p
	}

	if batchSize := os.Getenv("QUEUE_BATCH_SIZE"); batchSize != "" {
		b, err := strconv.Atoi(batchSize)
		if err != nil {
			return nil, fmt.Errorf("invalid QUEUE_BATCH_SIZE: %w", err)
		}
		config.QueueBatchSize = b
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid email configuration: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}

	if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
		return fmt.Errorf("SMTP_PORT must be between 1 and 65535")
	}

	if c.FromEmail == "" {
		return fmt.Errorf("SMTP_FROM_EMAIL is required")
	}

	if !isValidEmail(c.FromEmail) {
		return fmt.Errorf("SMTP_FROM_EMAIL is not a valid email address")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("SMTP_TIMEOUT must be positive")
	}

	if c.RateLimit <= 0 {
		return fmt.Errorf("EMAIL_RATE_LIMIT must be positive")
	}

	if c.MaxRetryAttempts < 1 || c.MaxRetryAttempts > 10 {
		return fmt.Errorf("MAX_RETRY_ATTEMPTS must be between 1 and 10")
	}

	return nil
}

func (c *Config) Sanitized() *Config {
	sanitized := *c
	if sanitized.SMTPPassword != "" {
		sanitized.SMTPPassword = "***REDACTED***"
	}
	return &sanitized
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
