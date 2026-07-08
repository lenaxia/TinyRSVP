package email

import (
	"context"
	"testing"
	"time"
)

func TestStubSMTPSender_Send(t *testing.T) {
	sender := NewStubSMTPSender()

	toName := "Test User"
	msg := &SMTPMessage{
		To:       "test@example.com",
		ToName:   &toName,
		Subject:  "Test Subject",
		BodyText: "Test body",
		BodyHTML: "<p>Test body</p>",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Errorf("Send() error = %v, want nil", err)
	}
}

func TestStubSMTPSender_TestConnection(t *testing.T) {
	sender := NewStubSMTPSender()

	err := sender.TestConnection(context.Background())
	if err != nil {
		t.Errorf("TestConnection() error = %v, want nil", err)
	}
}

func TestStubSMTPSender_Close(t *testing.T) {
	sender := NewStubSMTPSender()

	err := sender.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestStubSMTPSender_SendMultiple(t *testing.T) {
	sender := NewStubSMTPSender()

	for i := 0; i < 10; i++ {
		msg := &SMTPMessage{
			To:       "test@example.com",
			Subject:  "Test",
			BodyText: "Test",
		}

		if err := sender.Send(context.Background(), msg); err != nil {
			t.Errorf("Send() error = %v, want nil", err)
		}
	}
}

func TestStubRateLimiter_Allow(t *testing.T) {
	limiter := NewStubRateLimiter()

	if !limiter.Allow() {
		t.Error("Allow() = false, want true")
	}
}

func TestStubRateLimiter_AllowMultiple(t *testing.T) {
	limiter := NewStubRateLimiter()

	for i := 0; i < 100; i++ {
		if !limiter.Allow() {
			t.Errorf("Allow() call %d = false, want true", i)
		}
	}
}

func TestStubRateLimiter_AvailableSlots(t *testing.T) {
	limiter := NewStubRateLimiter()

	slots := limiter.AvailableSlots()
	if slots != 1000 {
		t.Errorf("AvailableSlots() = %d, want 1000", slots)
	}
}

func TestStubRateLimiter_AvailableSlotsAfterAllow(t *testing.T) {
	limiter := NewStubRateLimiter()

	limiter.Allow()

	slots := limiter.AvailableSlots()
	if slots != 1000 {
		t.Errorf("AvailableSlots() = %d, want 1000", slots)
	}
}

// TestStubRateLimiter_WaitTime covers stubRateLimiter.WaitTime, which must
// always report zero wait regardless of prior calls (table-driven).
func TestStubRateLimiter_WaitTime(t *testing.T) {
	tests := []struct {
		name  string
		setup func(RateLimiter)
		want  time.Duration
	}{
		{"fresh limiter", nil, 0},
		{"after allow", func(r RateLimiter) { r.Allow() }, 0},
		{"after record", func(r RateLimiter) { r.Record() }, 0},
		{"after many records", func(r RateLimiter) {
			for i := 0; i < 100; i++ {
				r.Record()
			}
		}, 0},
		{"after reset", func(r RateLimiter) { r.Reset() }, 0},
		{"after allow, record and reset", func(r RateLimiter) {
			r.Allow()
			r.Record()
			r.Reset()
		}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewStubRateLimiter()
			if tt.setup != nil {
				tt.setup(limiter)
			}
			if got := limiter.WaitTime(); got != tt.want {
				t.Errorf("WaitTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStubRateLimiter_Record covers stubRateLimiter.Record: it is a stateless
// no-op, so many calls must not panic or alter Allow/AvailableSlots/WaitTime.
func TestStubRateLimiter_Record(t *testing.T) {
	limiter := NewStubRateLimiter()

	for i := 0; i < 50; i++ {
		limiter.Record()
	}

	if !limiter.Allow() {
		t.Error("Allow() = false, want true after Record calls")
	}
	if got := limiter.AvailableSlots(); got != 1000 {
		t.Errorf("AvailableSlots() = %d, want 1000 after Record calls", got)
	}
	if got := limiter.WaitTime(); got != 0 {
		t.Errorf("WaitTime() = %v, want 0 after Record calls", got)
	}
}

// TestStubRateLimiter_Reset covers stubRateLimiter.Reset: it is a stateless
// no-op, so behaviour must be unchanged after a reset.
func TestStubRateLimiter_Reset(t *testing.T) {
	limiter := NewStubRateLimiter()

	limiter.Allow()
	limiter.Record()
	limiter.Reset()

	if !limiter.Allow() {
		t.Error("Allow() = false, want true after Reset")
	}
	if got := limiter.AvailableSlots(); got != 1000 {
		t.Errorf("AvailableSlots() = %d, want 1000 after Reset", got)
	}
	if got := limiter.WaitTime(); got != 0 {
		t.Errorf("WaitTime() = %v, want 0 after Reset", got)
	}
}
