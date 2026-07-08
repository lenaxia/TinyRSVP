package email

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
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

// fakeSMTPServer is a minimal in-process SMTP server used to exercise
// (*smtpSender).Send without an external (MailHog) dependency. It speaks just
// enough of the SMTP protocol for the net/smtp client to authenticate, send a
// message and quit. Per-command response overrides let tests simulate
// permanent/transient failures at any step.
type fakeSMTPServer struct {
	listener net.Listener
	host     string
	port     int

	// Optional response overrides. When empty, a default success response is sent.
	authResp   string // response to AUTH (default "235 OK")
	mailResp   string // response to MAIL FROM (default "250 OK")
	rcptResp   string // response to RCPT TO (default "250 OK")
	dataResp   string // response to the final "." of DATA (default "250 OK")
	dataCmdResp string // response to the DATA command (default "354 ..."); when set, DATA is rejected and data mode is not entered
	quitResp   string // response to QUIT (default "221 Bye")

	dropOnConnect bool // close the connection immediately (simulates reset)
}

func newFakeSMTPServer(t *testing.T) *fakeSMTPServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake SMTP server: %v", err)
	}
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	s := &fakeSMTPServer{listener: ln, host: host, port: port}
	go s.serve()
	return s
}

func (s *fakeSMTPServer) close() {
	s.listener.Close()
}

func (s *fakeSMTPServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handle(conn)
	}
}

func (s *fakeSMTPServer) handle(conn net.Conn) {
	defer conn.Close()
	if s.dropOnConnect {
		return
	}
	r := bufio.NewReader(conn)
	writeLine := func(msg string) {
		conn.Write([]byte(msg))
	}
	orDefault := func(resp, def string) string {
		if resp != "" {
			return resp
		}
		return def
	}

	writeLine("220 fake.smtp ESMTP ready\r\n")

	inData := false
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		if inData {
			if strings.TrimSpace(line) == "." {
				inData = false
				writeLine(orDefault(s.dataResp, "250 OK\r\n"))
			}
			continue
		}
		upper := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			writeLine("250 fake.smtp\r\n")
		case strings.HasPrefix(upper, "AUTH"):
			writeLine(orDefault(s.authResp, "235 OK\r\n"))
		case strings.HasPrefix(upper, "MAIL"):
			writeLine(orDefault(s.mailResp, "250 OK\r\n"))
		case strings.HasPrefix(upper, "RCPT"):
			writeLine(orDefault(s.rcptResp, "250 OK\r\n"))
		case strings.HasPrefix(upper, "DATA"):
			if s.dataCmdResp != "" {
				writeLine(s.dataCmdResp)
			} else {
				writeLine("354 Start mail input; end with <CRLF>.<CRLF>\r\n")
				inData = true
			}
		case strings.HasPrefix(upper, "QUIT"):
			writeLine(orDefault(s.quitResp, "221 Bye\r\n"))
			return
		case strings.HasPrefix(upper, "RSET"), strings.HasPrefix(upper, "NOOP"):
			writeLine("250 OK\r\n")
		default:
			writeLine("500 Unrecognized command\r\n")
		}
	}
}

// TestSMTPSender_Send exercises (*smtpSender).Send against an in-process fake
// SMTP server, covering happy paths (with and without auth) and unhappy paths
// where the server rejects a command with a permanent (5xx) SMTP error.
func TestSMTPSender_Send(t *testing.T) {
	tests := []struct {
		name     string
		useAuth  bool
		setup    func(*fakeSMTPServer)
		wantErr  bool
		wantPerm bool // expect the error to be a *PermanentError
	}{
		{
			name:    "happy path without auth",
			useAuth: false,
			wantErr: false,
		},
		{
			name:    "happy path with auth",
			useAuth: true,
			wantErr: false,
		},
		{
			name:     "rcpt rejected yields permanent error",
			setup:    func(s *fakeSMTPServer) { s.rcptResp = "550 No such user\r\n" },
			wantErr:  true,
			wantPerm: true,
		},
		{
			name:     "mail from rejected yields permanent error",
			setup:    func(s *fakeSMTPServer) { s.mailResp = "550 Sender not allowed\r\n" },
			wantErr:  true,
			wantPerm: true,
		},
		{
			name:     "data rejected yields permanent error",
			setup:    func(s *fakeSMTPServer) { s.dataResp = "554 Message rejected\r\n" },
			wantErr:  true,
			wantPerm: true,
		},
		{
			name:     "data command rejected yields permanent error",
			setup:    func(s *fakeSMTPServer) { s.dataCmdResp = "554 No data accepted\r\n" },
			wantErr:  true,
			wantPerm: true,
		},
		{
			name:    "auth rejected yields non-classified error",
			useAuth: true,
			setup:   func(s *fakeSMTPServer) { s.authResp = "535 Authentication failed\r\n" },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newFakeSMTPServer(t)
			defer srv.close()
			if tt.setup != nil {
				tt.setup(srv)
			}

			cfg := testConfig()
			cfg.SMTPHost = srv.host
			cfg.SMTPPort = srv.port
			cfg.UseTLS = false
			cfg.SkipVerify = false
			cfg.Timeout = 5 * time.Second
			if !tt.useAuth {
				cfg.SMTPUsername = ""
				cfg.SMTPPassword = ""
			}

			sender, err := NewSMTPSender(cfg)
			if err != nil {
				t.Fatalf("NewSMTPSender() error = %v", err)
			}

			toName := "Recipient"
			msg := &SMTPMessage{
				To:       "recipient@example.com",
				ToName:   &toName,
				Subject:  "Test Subject",
				BodyText: "plain body",
				BodyHTML: "<p>html body</p>",
			}

			err = sender.Send(context.Background(), msg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Send() error = nil, want error")
				}
				if tt.wantPerm {
					if _, ok := err.(*PermanentError); !ok {
						t.Errorf("Send() error = %T (%v), want *PermanentError", err, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Send() error = %v, want nil", err)
			}
		})
	}
}

// TestSMTPSender_Send_DialFailure covers the unhappy path where the SMTP server
// is unreachable: Send must surface a "failed to connect to SMTP" error.
func TestSMTPSender_Send_DialFailure(t *testing.T) {
	cfg := testConfig()
	cfg.SMTPHost = "127.0.0.1"
	cfg.SMTPPort = 1 // nothing listening on port 1 -> connection refused
	cfg.UseTLS = false
	cfg.Timeout = 2 * time.Second

	sender, err := NewSMTPSender(cfg)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	msg := &SMTPMessage{
		To:       "recipient@example.com",
		Subject:  "Test",
		BodyText: "body",
	}

	err = sender.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("Send() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to connect to SMTP") {
		t.Errorf("Send() error = %q, want it to contain %q", err.Error(), "failed to connect to SMTP")
	}
}

// TestSMTPSender_PermanentError_Unwrap covers PermanentError.Unwrap (and the
// errors.Is integration that relies on it) across a wrapped error and nil.
func TestSMTPSender_PermanentError_Unwrap(t *testing.T) {
	tests := []struct {
		name    string
		wrapped error
		wantNil bool
	}{
		{"wrapped error", errors.New("550 mailbox unavailable"), false},
		{"nil wrapped", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := &PermanentError{Err: tt.wrapped}

			got := pe.Unwrap()
			if got != tt.wrapped {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wrapped)
			}
			if tt.wantNil && got != nil {
				t.Errorf("Unwrap() = %v, want nil", got)
			}

			// errors.Is exercises Unwrap on the non-nil path.
			if tt.wrapped != nil {
				if !errors.Is(pe, tt.wrapped) {
					t.Errorf("errors.Is(pe, wrapped) = false, want true")
				}
			}

			// Error() should embed the wrapped message when non-nil.
			if tt.wrapped != nil {
				if !strings.Contains(pe.Error(), tt.wrapped.Error()) {
					t.Errorf("Error() = %q, want it to contain %q", pe.Error(), tt.wrapped.Error())
				}
			}
		})
	}
}
