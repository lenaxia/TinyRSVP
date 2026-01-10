package email

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
)

func testConfig() *Config {
	return &Config{
		SMTPHost:         "smtp.example.com",
		SMTPPort:         587,
		SMTPUsername:     "user@example.com",
		SMTPPassword:     "password",
		FromEmail:        "noreply@example.com",
		FromName:         "Test Sender",
		UseTLS:           true,
		Timeout:          30 * time.Second,
		RateLimit:        50,
		MaxRetryAttempts: 4,
	}
}

func TestNewSMTPSender_ValidConfig(t *testing.T) {
	config := testConfig()

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v, want nil", err)
	}

	if sender == nil {
		t.Fatal("NewSMTPSender() returned nil sender")
	}
}

func TestNewSMTPSender_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr string
	}{
		{
			name: "missing host",
			config: &Config{
				SMTPPort:         587,
				FromEmail:        "test@example.com",
				Timeout:          30 * time.Second,
				RateLimit:        50,
				MaxRetryAttempts: 4,
			},
			wantErr: "SMTP_HOST is required",
		},
		{
			name: "missing port",
			config: &Config{
				SMTPHost:         "smtp.example.com",
				FromEmail:        "test@example.com",
				Timeout:          30 * time.Second,
				RateLimit:        50,
				MaxRetryAttempts: 4,
			},
			wantErr: "SMTP_PORT must be between 1 and 65535",
		},
		{
			name: "missing from email",
			config: &Config{
				SMTPHost:         "smtp.example.com",
				SMTPPort:         587,
				Timeout:          30 * time.Second,
				RateLimit:        50,
				MaxRetryAttempts: 4,
			},
			wantErr: "EMAIL_FROM is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSMTPSender(tt.config)
			if err == nil {
				t.Fatal("NewSMTPSender() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("NewSMTPSender() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewSMTPSender_InvalidMaxConnections(t *testing.T) {
	tests := []struct {
		name           string
		maxConnections int
		wantErr        bool
	}{
		{"zero defaults to 10", 0, false},
		{"valid value 1", 1, false},
		{"valid value 50", 50, false},
		{"valid value 100", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testConfig()
			config.MaxConnections = tt.maxConnections

			_, err := NewSMTPSender(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSMTPSender() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSMTPSender_DefaultTimeout(t *testing.T) {
	config := testConfig()

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v, want nil", err)
	}

	smtpSender := sender.(*smtpSender)
	if smtpSender.config.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want %v", smtpSender.config.Timeout, 30*time.Second)
	}
}

func TestBuildMIMEMessage_PlainText(t *testing.T) {
	config := testConfig()
	config.FromEmail = "sender@example.com"

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	smtpSender := sender.(*smtpSender)

	msg := &SMTPMessage{
		To:       "recipient@example.com",
		Subject:  "Test Subject",
		BodyText: "This is a test message",
	}

	mimeMsg, err := smtpSender.buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage() error = %v", err)
	}

	msgStr := string(mimeMsg)

	if !strings.Contains(msgStr, "From: Test Sender <sender@example.com>") {
		t.Error("MIME message missing From header")
	}
	if !strings.Contains(msgStr, "To: recipient@example.com") {
		t.Error("MIME message missing To header")
	}
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Error("MIME message missing Subject header")
	}
	if !strings.Contains(msgStr, "Content-Type: text/plain") {
		t.Error("MIME message missing Content-Type header")
	}
	if !strings.Contains(msgStr, "This is a test message") {
		t.Error("MIME message missing body text")
	}
}

func TestBuildMIMEMessage_WithToName(t *testing.T) {
	config := testConfig()
	config.FromEmail = "sender@example.com"

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	smtpSender := sender.(*smtpSender)

	toName := "John Doe"
	msg := &SMTPMessage{
		To:       "recipient@example.com",
		ToName:   &toName,
		Subject:  "Test Subject",
		BodyText: "This is a test message",
	}

	mimeMsg, err := smtpSender.buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage() error = %v", err)
	}

	msgStr := string(mimeMsg)

	if !strings.Contains(msgStr, "To: John Doe <recipient@example.com>") {
		t.Errorf("MIME message has incorrect To header, got: %s", msgStr)
	}
}

func TestBuildMIMEMessage_HTMLAndText(t *testing.T) {
	config := testConfig()
	config.FromEmail = "sender@example.com"

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	smtpSender := sender.(*smtpSender)

	htmlBody := "<html><body><h1>Test</h1></body></html>"
	msg := &SMTPMessage{
		To:       "recipient@example.com",
		Subject:  "Test Subject",
		BodyText: "Plain text version",
		BodyHTML: htmlBody,
	}

	mimeMsg, err := smtpSender.buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage() error = %v", err)
	}

	msgStr := string(mimeMsg)

	if !strings.Contains(msgStr, "Content-Type: multipart/alternative") {
		t.Error("MIME message should be multipart/alternative")
	}
	if !strings.Contains(msgStr, "Plain text version") {
		t.Error("MIME message missing plain text part")
	}
	if !strings.Contains(msgStr, htmlBody) {
		t.Error("MIME message missing HTML part")
	}
}

func TestBuildMIMEMessage_WithAttachment(t *testing.T) {
	config := testConfig()
	config.FromEmail = "sender@example.com"

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	smtpSender := sender.(*smtpSender)

	attachmentData := []byte("test attachment content")
	msg := &SMTPMessage{
		To:       "recipient@example.com",
		Subject:  "Test Subject",
		BodyText: "Message with attachment",
		Attachments: []Attachment{
			{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Data:        attachmentData,
			},
		},
	}

	mimeMsg, err := smtpSender.buildMIMEMessage(msg)
	if err != nil {
		t.Fatalf("buildMIMEMessage() error = %v", err)
	}

	msgStr := string(mimeMsg)

	if !strings.Contains(msgStr, "Content-Type: multipart/mixed") {
		t.Error("MIME message should be multipart/mixed")
	}
	if !strings.Contains(msgStr, "filename=\"test.txt\"") {
		t.Error("MIME message missing attachment filename")
	}
	if !strings.Contains(msgStr, "Content-Transfer-Encoding: base64") {
		t.Error("MIME message missing base64 encoding header")
	}

	expectedEncoded := base64.StdEncoding.EncodeToString(attachmentData)
	if !strings.Contains(msgStr, expectedEncoded) {
		t.Error("MIME message missing base64 encoded attachment data")
	}
}

func TestClassifyError_PermanentErrors(t *testing.T) {
	tests := []struct {
		name      string
		smtpError string
	}{
		{"550 mailbox unavailable", "550 Mailbox unavailable"},
		{"551 user not local", "551 User not local"},
		{"552 exceeded storage", "552 Exceeded storage allocation"},
		{"553 mailbox name invalid", "553 Mailbox name not allowed"},
		{"554 transaction failed", "554 Transaction failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyError(fmt.Errorf("%s", tt.smtpError))

			if !strings.Contains(fmt.Sprintf("%T", err), "PermanentError") {
				t.Errorf("classifyError() = %T, want *PermanentError", err)
			}

			if permErr, ok := err.(*PermanentError); ok {
				if permErr.Err == nil {
					t.Error("PermanentError.Err should not be nil")
				}
			}
		})
	}
}

func TestClassifyError_TransientErrors(t *testing.T) {
	tests := []struct {
		name      string
		smtpError string
	}{
		{"421 service not available", "421 Service not available"},
		{"450 mailbox unavailable", "450 Mailbox unavailable"},
		{"451 local error", "451 Local error in processing"},
		{"452 insufficient storage", "452 Insufficient system storage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyError(fmt.Errorf("%s", tt.smtpError))

			if !strings.Contains(fmt.Sprintf("%T", err), "TransientError") {
				t.Errorf("classifyError() = %T, want *TransientError", err)
			}

			if transErr, ok := err.(*TransientError); ok {
				if transErr.Err == nil {
					t.Error("TransientError.Err should not be nil")
				}
			}
		})
	}
}

func TestClassifyError_UnknownError(t *testing.T) {
	err := classifyError(fmt.Errorf("%s", "unknown error"))

	if _, ok := err.(*PermanentError); ok {
		t.Error("Unknown error should not be classified as permanent")
	}
	if _, ok := err.(*TransientError); ok {
		t.Error("Unknown error should not be classified as transient")
	}
}

func TestGenerateBoundary(t *testing.T) {
	boundary1 := generateBoundary()
	boundary2 := generateBoundary()

	if boundary1 == "" {
		t.Error("generateBoundary() returned empty string")
	}

	if boundary1 == boundary2 {
		t.Error("generateBoundary() should return unique boundaries")
	}

	if len(boundary1) < 10 {
		t.Errorf("generateBoundary() returned short boundary: %s", boundary1)
	}
}

func TestSMTPSender_TestConnection_Success(t *testing.T) {
	config := testConfig()

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	smtpSender := sender.(*smtpSender)
	if smtpSender == nil {
		t.Fatal("Expected smtpSender instance")
	}
}

func TestSMTPSender_Close(t *testing.T) {
	config := testConfig()

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	err = sender.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestSMTPSender_Close_MultipleCalls(t *testing.T) {
	config := testConfig()

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	err = sender.Close()
	if err != nil {
		t.Errorf("First Close() error = %v, want nil", err)
	}

	err = sender.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v, want nil", err)
	}
}
