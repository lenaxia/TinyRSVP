package models

import "time"

type Session struct {
	ID             string    `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	ExpiresAt      time.Time `db:"expires_at" json:"expires_at"`
	LastAccessedAt time.Time `db:"last_accessed_at" json:"last_accessed_at"`
	IPAddress      *string   `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent      *string   `db:"user_agent" json:"user_agent,omitempty"`
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
