package email

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type smtpSender struct {
	config *Config
	auth   smtp.Auth
}

func NewSMTPSender(config *Config) (SMTPSender, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SMTP config: %w", err)
	}

	var auth smtp.Auth
	if config.SMTPUsername != "" && config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	}

	return &smtpSender{
		config: config,
		auth:   auth,
	}, nil
}

func (s *smtpSender) Send(ctx context.Context, msg *SMTPMessage) error {
	mimeMsg, err := s.buildMIMEMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to build MIME message: %w", err)
	}

	client, err := s.connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP: %w", err)
	}
	defer client.Close()

	if s.auth != nil {
		if err := client.Auth(s.auth); err != nil {
			return classifyError(err)
		}
	}

	if err := client.Mail(s.config.FromEmail); err != nil {
		return classifyError(err)
	}

	if err := client.Rcpt(msg.To); err != nil {
		return classifyError(err)
	}

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
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	dialer := &net.Dialer{
		Timeout: s.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}

	if s.config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName:         s.config.SMTPHost,
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

	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))

	toHeader := msg.To
	if msg.ToName != nil && *msg.ToName != "" {
		toHeader = fmt.Sprintf("%s <%s>", *msg.ToName, msg.To)
	}
	buf.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	if len(msg.Attachments) > 0 {
		boundary := generateBoundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

		if err := s.writeBodyParts(&buf, msg, boundary); err != nil {
			return nil, err
		}

		if err := s.writeAttachments(&buf, msg.Attachments, boundary); err != nil {
			return nil, err
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.BodyHTML != "" {
		boundary := generateBoundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))

		if err := s.writeBodyParts(&buf, msg, boundary); err != nil {
			return nil, err
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		buf.WriteString(msg.BodyText)
		buf.WriteString("\r\n")
	}

	return []byte(buf.String()), nil
}

func (s *smtpSender) writeBodyParts(buf *strings.Builder, msg *SMTPMessage, boundary string) error {
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	buf.WriteString(msg.BodyText)
	buf.WriteString("\r\n\r\n")

	if msg.BodyHTML != "" {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
		buf.WriteString(msg.BodyHTML)
		buf.WriteString("\r\n\r\n")
	}

	return nil
}

func (s *smtpSender) writeAttachments(buf *strings.Builder, attachments []Attachment, boundary string) error {
	for _, att := range attachments {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.ContentType))
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

		encoded := base64.StdEncoding.EncodeToString(att.Data)
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


func classifyError(err error) error {
	errStr := err.Error()

	if strings.Contains(errStr, "550") ||
		strings.Contains(errStr, "551") ||
		strings.Contains(errStr, "552") ||
		strings.Contains(errStr, "553") ||
		strings.Contains(errStr, "554") {
		return &PermanentError{Err: err}
	}

	if strings.Contains(errStr, "421") ||
		strings.Contains(errStr, "450") ||
		strings.Contains(errStr, "451") ||
		strings.Contains(errStr, "452") {
		return &TransientError{Err: err}
	}

	return err
}

func generateBoundary() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("boundary_%x", b)
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
