package models

import (
	"fmt"
	"net/mail"
	"time"
)

type InviteStatus string

const (
	InviteStatusDraft     InviteStatus = "draft"
	InviteStatusSent      InviteStatus = "sent"
	InviteStatusViewed    InviteStatus = "viewed"
	InviteStatusResponded InviteStatus = "responded"
	InviteStatusRevoked   InviteStatus = "revoked"
)

type Invite struct {
	ID               int64        `db:"id" json:"id"`
	EventID          int64        `db:"event_id" json:"event_id"`
	Name             *string      `db:"name" json:"name,omitempty"`
	Email            *string      `db:"email" json:"email,omitempty"`
	Token            *string      `db:"token" json:"token,omitempty"`
	TokenHash        string       `db:"token_hash" json:"-"`
	MaxPlusOnes      int          `db:"max_plus_ones" json:"max_plus_ones"`
	Status           InviteStatus `db:"status" json:"status"`
	SentAt           *time.Time   `db:"sent_at" json:"sent_at,omitempty"`
	ViewedAt         *time.Time   `db:"viewed_at" json:"viewed_at,omitempty"`
	RevocationReason *string      `db:"revocation_reason" json:"revocation_reason,omitempty"`
	Unsubscribed     bool         `db:"unsubscribed" json:"unsubscribed"`
	EmailInvalid     bool         `db:"email_invalid" json:"email_invalid"`
	CreatedAt        time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updated_at"`
	ExpiresAt        time.Time    `db:"expires_at" json:"expires_at"`
	// RSVPResponse is populated only when fetched via a JOIN with the rsvps table.
	RSVPResponse *RSVPResponse `db:"rsvp_response" json:"rsvp_response,omitempty"`
}

func (i *Invite) Validate() error {
	if i.EventID <= 0 {
		return &ValidationError{
			Field:   "event_id",
			Message: "event_id must be positive",
		}
	}

	if i.TokenHash == "" {
		return &ValidationError{
			Field:   "token_hash",
			Message: "token_hash is required",
		}
	}

	if len(i.TokenHash) != 43 {
		return &ValidationError{
			Field:   "token_hash",
			Message: "token_hash must be 43 characters",
		}
	}

	if i.MaxPlusOnes < 0 || i.MaxPlusOnes > 10 {
		return &ValidationError{
			Field:   "max_plus_ones",
			Message: "max_plus_ones must be between 0 and 10",
		}
	}

	if !isValidInviteStatus(i.Status) {
		return &ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("invalid status: %s", i.Status),
		}
	}

	if i.ExpiresAt.IsZero() {
		return &ValidationError{
			Field:   "expires_at",
			Message: "expires_at is required",
		}
	}

	if i.ExpiresAt.Before(time.Now()) {
		return &ValidationError{
			Field:   "expires_at",
			Message: "expires_at must be in the future",
		}
	}

	if i.Status == InviteStatusSent && i.Email == nil {
		return &ValidationError{
			Field:   "email",
			Message: "email is required for sent invites",
		}
	}

	if i.Name != nil && len(*i.Name) > 100 {
		return &ValidationError{
			Field:   "name",
			Message: "name must not exceed 100 characters",
		}
	}

	if i.Email != nil {
		if len(*i.Email) > 255 {
			return &ValidationError{
				Field:   "email",
				Message: "email must not exceed 255 characters",
			}
		}
		if _, err := mail.ParseAddress(*i.Email); err != nil {
			return &ValidationError{
				Field:   "email",
				Message: "email must be a valid email address",
			}
		}
	}

	return nil
}

func (i *Invite) CanTransitionTo(newStatus InviteStatus) error {
	validTransitions := map[InviteStatus][]InviteStatus{
		InviteStatusDraft:     {InviteStatusSent, InviteStatusRevoked},
		InviteStatusSent:      {InviteStatusViewed, InviteStatusRevoked},
		InviteStatusViewed:    {InviteStatusResponded, InviteStatusRevoked},
		InviteStatusResponded: {},
		InviteStatusRevoked:   {},
	}

	allowed, exists := validTransitions[i.Status]
	if !exists {
		return &ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("unknown current status: %s", i.Status),
		}
	}

	for _, validStatus := range allowed {
		if validStatus == newStatus {
			return nil
		}
	}

	return &ValidationError{
		Field:   "status",
		Message: fmt.Sprintf("cannot transition from %s to %s", i.Status, newStatus),
	}
}

func isValidInviteStatus(status InviteStatus) bool {
	switch status {
	case InviteStatusDraft, InviteStatusSent, InviteStatusViewed,
		InviteStatusResponded, InviteStatusRevoked:
		return true
	default:
		return false
	}
}
