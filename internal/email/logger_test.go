package email

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slogger := slog.New(handler)
	logger := NewLogger(slogger)

	t.Run("EmailQueued", func(t *testing.T) {
		buf.Reset()
		logger.EmailQueued(123, "test@example.com")

		output := buf.String()
		if !strings.Contains(output, "email queued") {
			t.Errorf("Expected 'email queued' in output, got: %s", output)
		}
		if !strings.Contains(output, "123") {
			t.Errorf("Expected email ID in output, got: %s", output)
		}
		if !strings.Contains(output, "test@example.com") {
			t.Errorf("Expected recipient in output, got: %s", output)
		}
	})

	t.Run("EmailSending", func(t *testing.T) {
		buf.Reset()
		logger.EmailSending(456, "user@example.com", 2)

		output := buf.String()
		if !strings.Contains(output, "sending email") {
			t.Errorf("Expected 'sending email' in output, got: %s", output)
		}
		if !strings.Contains(output, "456") {
			t.Errorf("Expected email ID in output, got: %s", output)
		}
		if !strings.Contains(output, "user@example.com") {
			t.Errorf("Expected recipient in output, got: %s", output)
		}
	})

	t.Run("EmailSent", func(t *testing.T) {
		buf.Reset()
		logger.EmailSent(789, "success@example.com", 2*time.Second)

		output := buf.String()
		if !strings.Contains(output, "email sent successfully") {
			t.Errorf("Expected 'email sent successfully' in output, got: %s", output)
		}
		if !strings.Contains(output, "789") {
			t.Errorf("Expected email ID in output, got: %s", output)
		}
	})

	t.Run("EmailFailed", func(t *testing.T) {
		buf.Reset()
		logger.EmailFailed(111, "fail@example.com", 1, context.DeadlineExceeded)

		output := buf.String()
		if !strings.Contains(output, "email send failed") {
			t.Errorf("Expected 'email send failed' in output, got: %s", output)
		}
		if !strings.Contains(output, "ERROR") {
			t.Errorf("Expected ERROR level in output, got: %s", output)
		}
	})

	t.Run("EmailRetrying", func(t *testing.T) {
		buf.Reset()
		logger.EmailRetrying(222, "retry@example.com", 2, 5*time.Minute)

		output := buf.String()
		if !strings.Contains(output, "retrying email") {
			t.Errorf("Expected 'retrying email' in output, got: %s", output)
		}
		if !strings.Contains(output, "WARN") {
			t.Errorf("Expected WARN level in output, got: %s", output)
		}
	})

	t.Run("EmailPermanentlyFailed", func(t *testing.T) {
		buf.Reset()
		logger.EmailPermanentlyFailed(333, "permfail@example.com", 4, context.DeadlineExceeded)

		output := buf.String()
		if !strings.Contains(output, "email permanently failed") {
			t.Errorf("Expected 'email permanently failed' in output, got: %s", output)
		}
		if !strings.Contains(output, "ERROR") {
			t.Errorf("Expected ERROR level in output, got: %s", output)
		}
	})

	t.Run("RateLimitHit", func(t *testing.T) {
		buf.Reset()
		logger.RateLimitHit(10, 30*time.Second)

		output := buf.String()
		if !strings.Contains(output, "rate limit hit") {
			t.Errorf("Expected 'rate limit hit' in output, got: %s", output)
		}
		if !strings.Contains(output, "WARN") {
			t.Errorf("Expected WARN level in output, got: %s", output)
		}
	})

	t.Run("BatchProcessed", func(t *testing.T) {
		buf.Reset()
		logger.BatchProcessed(25, 10*time.Second)

		output := buf.String()
		if !strings.Contains(output, "batch processed") {
			t.Errorf("Expected 'batch processed' in output, got: %s", output)
		}
		if !strings.Contains(output, "25") {
			t.Errorf("Expected batch count in output, got: %s", output)
		}
	})

	t.Run("QueueProcessorStarted", func(t *testing.T) {
		buf.Reset()
		logger.QueueProcessorStarted(60*time.Second, 50)

		output := buf.String()
		if !strings.Contains(output, "queue processor started") {
			t.Errorf("Expected 'queue processor started' in output, got: %s", output)
		}
	})

	t.Run("QueueProcessorStopped", func(t *testing.T) {
		buf.Reset()
		logger.QueueProcessorStopped()

		output := buf.String()
		if !strings.Contains(output, "queue processor stopped") {
			t.Errorf("Expected 'queue processor stopped' in output, got: %s", output)
		}
	})
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slogger := slog.New(handler)
	logger := NewLogger(slogger)

	logger.EmailSent(999, "json@example.com", 1500*time.Millisecond)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["msg"] != "email sent successfully" {
		t.Errorf("Expected msg 'email sent successfully', got: %v", logEntry["msg"])
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got: %v", logEntry["level"])
	}

	emailID, ok := logEntry["email_id"].(float64)
	if !ok || int64(emailID) != 999 {
		t.Errorf("Expected email_id 999, got: %v", logEntry["email_id"])
	}

	if logEntry["recipient"] != "json@example.com" {
		t.Errorf("Expected recipient 'json@example.com', got: %v", logEntry["recipient"])
	}
}

func TestLoggerNoSensitiveData(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slogger := slog.New(handler)
	logger := NewLogger(slogger)

	logger.EmailQueued(1, "user@example.com")
	logger.EmailSending(2, "user@example.com", 1)
	logger.EmailSent(3, "user@example.com", time.Second)

	output := buf.String()

	sensitivePatterns := []string{
		"password",
		"secret",
		"token",
		"api_key",
		"smtp_password",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(output), pattern) {
			t.Errorf("Found sensitive data pattern '%s' in logs: %s", pattern, output)
		}
	}
}
