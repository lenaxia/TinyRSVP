package models

import (
	"testing"
	"time"
)

func TestSession_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired session from past",
			expiresAt: now.Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired session from 1 second ago",
			expiresAt: now.Add(-1 * time.Second),
			want:      true,
		},
		{
			name:      "valid session expires in future",
			expiresAt: now.Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "valid session expires in 1 second",
			expiresAt: now.Add(1 * time.Second),
			want:      false,
		},
		{
			name:      "session expires far in future",
			expiresAt: now.Add(7 * 24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{ExpiresAt: tt.expiresAt}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("Session.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_StructFields(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	session := &Session{
		ID:             "session-abc-123",
		UserID:         456,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
		LastAccessedAt: now,
		IPAddress:      &ipAddress,
		UserAgent:      &userAgent,
	}

	t.Run("all fields are accessible", func(t *testing.T) {
		if session.ID != "session-abc-123" {
			t.Errorf("ID = %v, want 'session-abc-123'", session.ID)
		}
		if session.UserID != 456 {
			t.Errorf("UserID = %v, want 456", session.UserID)
		}
		if session.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
		if session.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should not be zero")
		}
		if session.LastAccessedAt.IsZero() {
			t.Error("LastAccessedAt should not be zero")
		}
		if session.IPAddress == nil || *session.IPAddress != ipAddress {
			t.Errorf("IPAddress = %v, want %v", session.IPAddress, ipAddress)
		}
		if session.UserAgent == nil || *session.UserAgent != userAgent {
			t.Errorf("UserAgent = %v, want %v", session.UserAgent, userAgent)
		}
	})

	t.Run("optional fields can be nil", func(t *testing.T) {
		minimalSession := &Session{
			ID:             "minimal-session",
			UserID:         1,
			CreatedAt:      now,
			ExpiresAt:      expiresAt,
			LastAccessedAt: now,
		}

		if minimalSession.IPAddress != nil {
			t.Error("IPAddress should be nil for minimal session")
		}
		if minimalSession.UserAgent != nil {
			t.Error("UserAgent should be nil for minimal session")
		}
	})
}

func TestSession_ExpirationEdgeCases(t *testing.T) {
	t.Run("session expiring exactly now", func(t *testing.T) {
		now := time.Now()
		session := &Session{ExpiresAt: now}
		time.Sleep(1 * time.Millisecond)
		if !session.IsExpired() {
			t.Error("Session expiring at current time should be expired after brief delay")
		}
	})

	t.Run("multiple calls to IsExpired are consistent", func(t *testing.T) {
		futureSession := &Session{ExpiresAt: time.Now().Add(1 * time.Hour)}
		pastSession := &Session{ExpiresAt: time.Now().Add(-1 * time.Hour)}

		for i := 0; i < 5; i++ {
			if futureSession.IsExpired() {
				t.Error("Future session should consistently return false")
			}
			if !pastSession.IsExpired() {
				t.Error("Past session should consistently return true")
			}
		}
	})
}

func TestSession_TimeFields(t *testing.T) {
	t.Run("time fields can be compared", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			CreatedAt:      now,
			ExpiresAt:      now.Add(24 * time.Hour),
			LastAccessedAt: now.Add(1 * time.Hour),
		}

		if !session.CreatedAt.Before(session.ExpiresAt) {
			t.Error("CreatedAt should be before ExpiresAt")
		}
		if !session.CreatedAt.Before(session.LastAccessedAt) {
			t.Error("CreatedAt should be before LastAccessedAt")
		}
		if !session.LastAccessedAt.Before(session.ExpiresAt) {
			t.Error("LastAccessedAt should be before ExpiresAt")
		}
	})
}
