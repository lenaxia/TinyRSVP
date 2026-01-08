package email

import (
	"context"
	"log"
)

type stubSMTPSender struct{}

func NewStubSMTPSender() SMTPSender {
	return &stubSMTPSender{}
}

func (s *stubSMTPSender) Send(ctx context.Context, msg *SMTPMessage) error {
	log.Printf("STUB: Would send email to %s with subject: %s", msg.To, msg.Subject)
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
