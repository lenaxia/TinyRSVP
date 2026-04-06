package email

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mailhogSMTPPort / mailhogAPIBase target the MailHog container from
//
//	docker-compose.test.yml (docker run -d --name mailhog-test \
//	  -p 1025:1025 -p 8025:8025 mailhog/mailhog:latest MH_STORAGE=memory).
//
// Tests are skipped automatically when MailHog is not reachable, so they are
// safe to run in any environment.
const mailhogSMTPPort = 1025
const mailhogAPIBase = "http://localhost:8025/api/v2"
const mailhogDeleteAPI = "http://localhost:8025/api/v1/messages"

func mailhogConfig() *Config {
	return &Config{
		SMTPHost:         "localhost",
		SMTPPort:         mailhogSMTPPort,
		SMTPUsername:     "",
		SMTPPassword:     "",
		FromEmail:        "sender@tinyrsvp.test",
		FromName:         "TinyRSVP Test",
		UseTLS:           false,
		SkipVerify:       false,
		Timeout:          5 * time.Second,
		RateLimit:        50,
		MaxRetryAttempts: 4,
	}
}

// skipIfMailhogUnavailable skips the test when MailHog's HTTP API is unreachable.
func skipIfMailhogUnavailable(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(mailhogAPIBase + "/messages?limit=1")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("MailHog not reachable at localhost:8025 — skipping SMTP integration tests")
	}
	resp.Body.Close()
}

// deleteAllMailhogMessages purges MailHog's inbox so each test starts clean.
func deleteAllMailhogMessages(t *testing.T) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, mailhogDeleteAPI, nil)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to clear MailHog inbox: %v", err)
	}
	resp.Body.Close()
}

// mailhogMessage is a minimal subset of the MailHog message JSON.
type mailhogMessage struct {
	ID      string `json:"ID"`
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
	Raw struct {
		From string   `json:"From"`
		To   []string `json:"To"`
		Data string   `json:"Data"`
	} `json:"Raw"`
}

type mailhogMessages struct {
	Total int              `json:"total"`
	Items []mailhogMessage `json:"items"`
}

// fetchMailhogMessages returns all messages currently in MailHog.
func fetchMailhogMessages(t *testing.T) []mailhogMessage {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(mailhogAPIBase + "/messages?limit=50")
	if err != nil {
		t.Fatalf("failed to fetch MailHog messages: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var msgs mailhogMessages
	if err := json.Unmarshal(body, &msgs); err != nil {
		t.Fatalf("failed to parse MailHog response: %v", err)
	}
	return msgs.Items
}

// waitForMailhogMessages polls until at least n messages appear or the timeout
// expires, returning whatever was found.
func waitForMailhogMessages(t *testing.T, n int, timeout time.Duration) []mailhogMessage {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		msgs := fetchMailhogMessages(t)
		if len(msgs) >= n {
			return msgs
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fetchMailhogMessages(t)
}

// ---------------------------------------------------------------------------
// TestConnection
// ---------------------------------------------------------------------------

func TestSMTPSender_TestConnection_MailHog(t *testing.T) {
	skipIfMailhogUnavailable(t)

	sender, err := NewSMTPSender(mailhogConfig())
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sender.TestConnection(ctx); err != nil {
		t.Errorf("TestConnection() error = %v, want nil", err)
	}
}

func TestSMTPSender_TestConnection_RefusedPort(t *testing.T) {
	config := mailhogConfig()
	config.SMTPPort = 19999 // nothing listening here

	sender, err := NewSMTPSender(config)
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sender.TestConnection(ctx); err == nil {
		t.Error("TestConnection() error = nil, want connection refused error")
	}
}

// ---------------------------------------------------------------------------
// Send — plain text
// ---------------------------------------------------------------------------

func TestSMTPSender_Send_PlainText_MailHog(t *testing.T) {
	skipIfMailhogUnavailable(t)
	deleteAllMailhogMessages(t)

	sender, err := NewSMTPSender(mailhogConfig())
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := &SMTPMessage{
		To:       "recipient@tinyrsvp.test",
		Subject:  "Plain text integration test",
		BodyText: "Hello from TinyRSVP plain-text test.",
	}

	if err := sender.Send(ctx, msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	msgs := waitForMailhogMessages(t, 1, 3*time.Second)
	if len(msgs) == 0 {
		t.Fatal("expected 1 message in MailHog, got 0")
	}

	got := msgs[0]
	if got.Raw.From != "sender@tinyrsvp.test" {
		t.Errorf("From = %q, want %q", got.Raw.From, "sender@tinyrsvp.test")
	}
	if len(got.Raw.To) == 0 || got.Raw.To[0] != "recipient@tinyrsvp.test" {
		t.Errorf("To = %v, want [recipient@tinyrsvp.test]", got.Raw.To)
	}
	if !strings.Contains(got.Raw.Data, "Plain text integration test") {
		t.Errorf("subject not found in raw message data")
	}
	if !strings.Contains(got.Raw.Data, "Hello from TinyRSVP plain-text test.") {
		t.Errorf("body text not found in raw message data")
	}
}

// ---------------------------------------------------------------------------
// Send — HTML + text multipart
// ---------------------------------------------------------------------------

func TestSMTPSender_Send_HTMLAndText_MailHog(t *testing.T) {
	skipIfMailhogUnavailable(t)
	deleteAllMailhogMessages(t)

	sender, err := NewSMTPSender(mailhogConfig())
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	toName := "Test Guest"
	msg := &SMTPMessage{
		To:       "guest@tinyrsvp.test",
		ToName:   &toName,
		Subject:  "HTML multipart integration test",
		BodyText: "You are invited (plain text fallback).",
		BodyHTML: "<html><body><h1>You are invited!</h1></body></html>",
	}

	if err := sender.Send(ctx, msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	msgs := waitForMailhogMessages(t, 1, 3*time.Second)
	if len(msgs) == 0 {
		t.Fatal("expected 1 message in MailHog, got 0")
	}

	raw := msgs[0].Raw.Data
	if !strings.Contains(raw, "multipart/alternative") {
		t.Error("expected multipart/alternative content-type in raw message")
	}
	if !strings.Contains(raw, "You are invited (plain text fallback).") {
		t.Error("plain text part not found in raw message")
	}
	if !strings.Contains(raw, "You are invited!") {
		t.Error("HTML part not found in raw message")
	}
	if !strings.Contains(raw, "Test Guest") {
		t.Errorf("To name not found in raw message headers")
	}
}

// ---------------------------------------------------------------------------
// Send — with ICS attachment
// ---------------------------------------------------------------------------

func TestSMTPSender_Send_WithAttachment_MailHog(t *testing.T) {
	skipIfMailhogUnavailable(t)
	deleteAllMailhogMessages(t)

	sender, err := NewSMTPSender(mailhogConfig())
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	icsData := []byte("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nEND:VCALENDAR\r\n")
	msg := &SMTPMessage{
		To:       "guest@tinyrsvp.test",
		Subject:  "Attachment integration test",
		BodyText: "Please find the event calendar attached.",
		Attachments: []Attachment{
			{
				Filename:    "event.ics",
				ContentType: "text/calendar",
				Data:        icsData,
			},
		},
	}

	if err := sender.Send(ctx, msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	msgs := waitForMailhogMessages(t, 1, 3*time.Second)
	if len(msgs) == 0 {
		t.Fatal("expected 1 message in MailHog, got 0")
	}

	raw := msgs[0].Raw.Data
	if !strings.Contains(raw, "multipart/mixed") {
		t.Error("expected multipart/mixed content-type for message with attachment")
	}
	if !strings.Contains(raw, `filename="event.ics"`) {
		t.Error("attachment filename not found in raw message")
	}
	if !strings.Contains(raw, "Content-Transfer-Encoding: base64") {
		t.Error("base64 encoding header not found in raw message")
	}
	if !strings.Contains(raw, "Please find the event calendar attached.") {
		t.Error("body text not found in raw message")
	}
}

// ---------------------------------------------------------------------------
// Send — multiple recipients in sequence
// ---------------------------------------------------------------------------

func TestSMTPSender_Send_MultipleMessages_MailHog(t *testing.T) {
	skipIfMailhogUnavailable(t)
	deleteAllMailhogMessages(t)

	sender, err := NewSMTPSender(mailhogConfig())
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	recipients := []string{
		"alice@tinyrsvp.test",
		"bob@tinyrsvp.test",
		"charlie@tinyrsvp.test",
	}

	for i, to := range recipients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg := &SMTPMessage{
			To:       to,
			Subject:  fmt.Sprintf("Bulk test message %d", i+1),
			BodyText: fmt.Sprintf("Message number %d", i+1),
		}
		if err := sender.Send(ctx, msg); err != nil {
			cancel()
			t.Fatalf("Send() to %s error = %v", to, err)
		}
		cancel()
	}

	msgs := waitForMailhogMessages(t, 3, 5*time.Second)
	if len(msgs) != 3 {
		t.Errorf("MailHog message count = %d, want 3", len(msgs))
	}
}

// ---------------------------------------------------------------------------
// Send — context cancellation before dial
// ---------------------------------------------------------------------------

func TestSMTPSender_Send_CancelledContext(t *testing.T) {
	skipIfMailhogUnavailable(t)

	sender, err := NewSMTPSender(mailhogConfig())
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	msg := &SMTPMessage{
		To:       "nobody@tinyrsvp.test",
		Subject:  "Should not be sent",
		BodyText: "This message should never arrive.",
	}

	if err := sender.Send(ctx, msg); err == nil {
		t.Error("Send() with cancelled context returned nil error, want context error")
	}
}
