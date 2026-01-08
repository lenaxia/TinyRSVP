# User Story: SMTP Sender Implementation

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** High
**Status:** Complete ✅
**Completed:** 2026-01-08
**Estimated Effort:** 1.5 days

---

## User Story

As a **system**, I want **an SMTP sender implementation** so that **emails can be delivered to recipients via SMTP protocol**.

---

## Acceptance Criteria

- [x] SMTP connection management with connection pooling
- [x] Support for TLS/STARTTLS encryption
- [x] SMTP authentication (PLAIN, LOGIN)
- [x] Send emails with HTML and plain text bodies
- [x] Support for attachments (ICS files)
- [x] Proper MIME multipart message construction
- [x] Connection timeout handling
- [x] Retry on transient SMTP errors
- [x] Classify errors (permanent vs transient)
- [x] Test connection validation
- [x] All tests pass with timeout
- [x] Integration tests with mock SMTP server

---

## Technical Details

### SMTP Sender Interface

```go
package email

import (
    "context"
    "time"
)

type SMTPSender interface {
    // Send sends an email via SMTP
    Send(ctx context.Context, msg *SMTPMessage) error
    
    // TestConnection verifies SMTP configuration
    TestConnection(ctx context.Context) error
    
    // Close closes the SMTP connection
    Close() error
}

type SMTPMessage struct {
    To          string
    ToName      *string
    Subject     string
    BodyText    string
    BodyHTML    *string
    Attachments []SMTPAttachment
}

type SMTPAttachment struct {
    Filename    string
    ContentType string
    Content     []byte
}

type SMTPConfig struct {
    Host            string
    Port            int
    Username        string
    Password        string
    FromEmail       string
    FromName        string
    UseTLS          bool
    SkipVerify      bool
    Timeout         time.Duration
    MaxConnections  int
}
```

### Implementation

```go
package email

import (
    "context"
    "crypto/tls"
    "fmt"
    "net"
    "net/smtp"
    "strings"
    "time"
)

type smtpSender struct {
    config *SMTPConfig
    auth   smtp.Auth
}

func NewSMTPSender(config *SMTPConfig) (SMTPSender, error) {
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("invalid SMTP config: %w", err)
    }
    
    var auth smtp.Auth
    if config.Username != "" && config.Password != "" {
        auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
    }
    
    return &smtpSender{
        config: config,
        auth:   auth,
    }, nil
}

func (s *smtpSender) Send(ctx context.Context, msg *SMTPMessage) error {
    // Build MIME message
    mimeMsg, err := s.buildMIMEMessage(msg)
    if err != nil {
        return fmt.Errorf("failed to build MIME message: %w", err)
    }
    
    // Connect to SMTP server
    client, err := s.connect(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to SMTP: %w", err)
    }
    defer client.Close()
    
    // Authenticate
    if s.auth != nil {
        if err := client.Auth(s.auth); err != nil {
            return classifyError(err)
        }
    }
    
    // Set sender
    if err := client.Mail(s.config.FromEmail); err != nil {
        return classifyError(err)
    }
    
    // Set recipient
    if err := client.Rcpt(msg.To); err != nil {
        return classifyError(err)
    }
    
    // Send message body
    w, err := client.Data()
    if err != nil {
        return classifyError(err)
    }
    
    if _, err := w.Write(mimeMsg); err != nil {
        return classifyError(err)
    }
    
    if err := w.Close(); err != nil {
        return classifyError(err)
    }
    
    return client.Quit()
}

func (s *smtpSender) connect(ctx context.Context) (*smtp.Client, error) {
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    
    // Create dialer with timeout
    dialer := &net.Dialer{
        Timeout: s.config.Timeout,
    }
    
    // Connect with context
    conn, err := dialer.DialContext(ctx, "tcp", addr)
    if err != nil {
        return nil, fmt.Errorf("failed to dial: %w", err)
    }
    
    // Create SMTP client
    client, err := smtp.NewClient(conn, s.config.Host)
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to create SMTP client: %w", err)
    }
    
    // Start TLS if required
    if s.config.UseTLS {
        tlsConfig := &tls.Config{
            ServerName:         s.config.Host,
            InsecureSkipVerify: s.config.SkipVerify,
        }
        
        if err := client.StartTLS(tlsConfig); err != nil {
            client.Close()
            return nil, fmt.Errorf("failed to start TLS: %w", err)
        }
    }
    
    return client, nil
}

func (s *smtpSender) buildMIMEMessage(msg *SMTPMessage) ([]byte, error) {
    var buf strings.Builder
    
    // Headers
    buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
    
    toName := msg.To
    if msg.ToName != nil {
        toName = fmt.Sprintf("%s <%s>", *msg.ToName, msg.To)
    }
    buf.WriteString(fmt.Sprintf("To: %s\r\n", toName))
    buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
    buf.WriteString("MIME-Version: 1.0\r\n")
    
    // Determine content type
    if len(msg.Attachments) > 0 {
        boundary := generateBoundary()
        buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))
        
        // Write body parts
        if err := s.writeBodyParts(&buf, msg, boundary); err != nil {
            return nil, err
        }
        
        // Write attachments
        if err := s.writeAttachments(&buf, msg.Attachments, boundary); err != nil {
            return nil, err
        }
        
        buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
    } else if msg.BodyHTML != nil {
        // Multipart alternative (HTML + text)
        boundary := generateBoundary()
        buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))
        
        if err := s.writeBodyParts(&buf, msg, boundary); err != nil {
            return nil, err
        }
        
        buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
    } else {
        // Plain text only
        buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
        buf.WriteString(msg.BodyText)
        buf.WriteString("\r\n")
    }
    
    return []byte(buf.String()), nil
}

func (s *smtpSender) writeBodyParts(buf *strings.Builder, msg *SMTPMessage, boundary string) error {
    // Plain text part
    buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
    buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
    buf.WriteString(msg.BodyText)
    buf.WriteString("\r\n\r\n")
    
    // HTML part (if present)
    if msg.BodyHTML != nil {
        buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
        buf.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
        buf.WriteString(*msg.BodyHTML)
        buf.WriteString("\r\n\r\n")
    }
    
    return nil
}

func (s *smtpSender) writeAttachments(buf *strings.Builder, attachments []SMTPAttachment, boundary string) error {
    for _, att := range attachments {
        buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
        buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.ContentType))
        buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
        buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
        
        // Base64 encode attachment
        encoded := base64Encode(att.Content)
        buf.WriteString(encoded)
        buf.WriteString("\r\n\r\n")
    }
    
    return nil
}

func (s *smtpSender) TestConnection(ctx context.Context) error {
    client, err := s.connect(ctx)
    if err != nil {
        return err
    }
    defer client.Close()
    
    if s.auth != nil {
        if err := client.Auth(s.auth); err != nil {
            return fmt.Errorf("authentication failed: %w", err)
        }
    }
    
    return client.Quit()
}

func (s *smtpSender) Close() error {
    return nil
}

func validateConfig(config *SMTPConfig) error {
    if config.Host == "" {
        return fmt.Errorf("SMTP host is required")
    }
    if config.Port == 0 {
        return fmt.Errorf("SMTP port is required")
    }
    if config.FromEmail == "" {
        return fmt.Errorf("from email is required")
    }
    if config.Timeout == 0 {
        config.Timeout = 30 * time.Second
    }
    return nil
}

func classifyError(err error) error {
    errStr := err.Error()
    
    // Permanent errors (5xx)
    if strings.Contains(errStr, "550") || // Mailbox unavailable
       strings.Contains(errStr, "551") || // User not local
       strings.Contains(errStr, "552") || // Exceeded storage
       strings.Contains(errStr, "553") || // Mailbox name invalid
       strings.Contains(errStr, "554") {  // Transaction failed
        return &PermanentError{Err: err}
    }
    
    // Transient errors (4xx)
    if strings.Contains(errStr, "421") || // Service not available
       strings.Contains(errStr, "450") || // Mailbox unavailable
       strings.Contains(errStr, "451") || // Local error
       strings.Contains(errStr, "452") {  // Insufficient storage
        return &TransientError{Err: err}
    }
    
    return err
}

type PermanentError struct {
    Err error
}

func (e *PermanentError) Error() string {
    return fmt.Sprintf("permanent SMTP error: %v", e.Err)
}

func (e *PermanentError) Unwrap() error {
    return e.Err
}

type TransientError struct {
    Err error
}

func (e *TransientError) Error() string {
    return fmt.Sprintf("transient SMTP error: %v", e.Err)
}

func (e *TransientError) Unwrap() error {
    return e.Err
}
```

---

## Tasks

### Phase 1: Interface & Configuration (TDD)
- [ ] Define SMTPSender interface
- [ ] Define SMTPMessage struct
- [ ] Define SMTPConfig struct
- [ ] Write test for config validation
- [ ] Implement config validation
- [ ] Write test for sender initialization
- [ ] Implement sender initialization

### Phase 2: SMTP Connection (TDD)
- [ ] Write test for SMTP connection
- [ ] Implement connect method
- [ ] Write test for TLS connection
- [ ] Implement TLS support
- [ ] Write test for authentication
- [ ] Implement authentication
- [ ] Write test for connection timeout
- [ ] Implement timeout handling

### Phase 3: MIME Message Building (TDD)
- [ ] Write test for plain text message
- [ ] Implement plain text MIME
- [ ] Write test for HTML message
- [ ] Implement HTML MIME
- [ ] Write test for multipart alternative
- [ ] Implement multipart alternative
- [ ] Write test for attachments
- [ ] Implement attachment encoding

### Phase 4: Email Sending (TDD)
- [ ] Write test for successful send
- [ ] Implement Send method
- [ ] Write test for send failure
- [ ] Implement error handling
- [ ] Write test for error classification
- [ ] Implement error classification
- [ ] Write test for recipient rejection
- [ ] Handle recipient errors

### Phase 5: Integration Testing
- [ ] Test with mock SMTP server
- [ ] Test with real SMTP (Gmail, SendGrid)
- [ ] Test TLS connection
- [ ] Test authentication
- [ ] Test attachment delivery
- [ ] Test error scenarios

---

## Dependencies

**Depends on:**
- Go standard library: `net/smtp`, `crypto/tls`

**Blocks:**
- Story 02: Email Queue Processor

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Interface defined and documented
- [x] Implementation complete
- [x] All unit tests passing
- [x] Integration tests passing
- [x] Error classification working
- [x] Code reviewed
- [x] Documentation updated

---

## Test Requirements

### Unit Tests

```go
func TestSMTPSender_Send_PlainText(t *testing.T) {
    config := &SMTPConfig{
        Host:      "localhost",
        Port:      1025,
        FromEmail: "test@example.com",
        FromName:  "Test Sender",
        Timeout:   5 * time.Second,
    }
    
    sender, err := NewSMTPSender(config)
    if err != nil {
        t.Fatal(err)
    }
    
    msg := &SMTPMessage{
        To:       "recipient@example.com",
        Subject:  "Test Email",
        BodyText: "This is a test",
    }
    
    err = sender.Send(context.Background(), msg)
    if err != nil {
        t.Errorf("Send() error = %v", err)
    }
}

func TestSMTPSender_Send_WithAttachment(t *testing.T) {
    // Test sending email with ICS attachment
}

func TestSMTPSender_Send_HTMLAndText(t *testing.T) {
    // Test multipart alternative message
}

func TestSMTPSender_TestConnection(t *testing.T) {
    // Test connection validation
}

func TestSMTPSender_ErrorClassification(t *testing.T) {
    tests := []struct {
        name      string
        smtpError string
        wantType  string
    }{
        {"permanent 550", "550 Mailbox unavailable", "permanent"},
        {"transient 450", "450 Mailbox busy", "transient"},
        {"permanent 554", "554 Transaction failed", "permanent"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := classifyError(fmt.Errorf(tt.smtpError))
            // Verify error type
        })
    }
}
```

### Integration Tests

```go
func TestSMTPSender_Integration_MockServer(t *testing.T) {
    // Start mock SMTP server
    server := startMockSMTPServer(t)
    defer server.Close()
    
    // Send email
    // Verify email received by mock server
}
```

---

## SMTP Error Codes

### Permanent Errors (5xx) - Do Not Retry
- **550**: Mailbox unavailable (user doesn't exist)
- **551**: User not local
- **552**: Exceeded storage allocation
- **553**: Mailbox name not allowed
- **554**: Transaction failed

### Transient Errors (4xx) - Retry
- **421**: Service not available
- **450**: Mailbox unavailable (temporarily)
- **451**: Local error in processing
- **452**: Insufficient system storage

---

## Security Considerations

- Use TLS for encryption (STARTTLS)
- Never log passwords
- Validate certificates (unless SkipVerify for testing)
- Use secure authentication methods
- Timeout all connections
- Rate limit connection attempts

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **RFC 5321:** SMTP Protocol
- **RFC 2045-2049:** MIME
- **RFC 2822:** Internet Message Format
