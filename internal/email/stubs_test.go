package email

import (
	"context"
	"testing"
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
