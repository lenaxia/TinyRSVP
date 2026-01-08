package email

import (
	"context"
	"log"
	"time"
)

type stubSMTPSender struct{}

func NewStubSMTPSender() SMTPSender {
	return &stubSMTPSender{}
}

func (s *stubSMTPSender) Send(ctx context.Context, msg *SMTPMessage) error {
	log.Printf("STUB: Would send email to %s with subject: %s", msg.To, msg.Subject)
	return nil
}

func (s *stubSMTPSender) TestConnection(ctx context.Context) error {
	log.Printf("STUB: Would test SMTP connection")
	return nil
}

func (s *stubSMTPSender) Close() error {
	log.Printf("STUB: Would close SMTP connection")
	return nil
}

type stubRateLimiter struct{}

func NewStubRateLimiter() RateLimiter {
	return &stubRateLimiter{}
}

func (r *stubRateLimiter) Allow() bool {
	return true
}

func (r *stubRateLimiter) AvailableSlots() int {
	return 1000
}

func (r *stubRateLimiter) WaitTime() time.Duration {
	return 0
}

func (r *stubRateLimiter) Record() {
}

func (r *stubRateLimiter) Reset() {
}
