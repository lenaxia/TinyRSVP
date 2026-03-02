package builders

import (
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

// InviteBuilder constructs models.Invite instances with sensible defaults.
type InviteBuilder struct {
	t      *testing.T
	invite *models.Invite
}

// NewInviteBuilder returns a builder pre-populated with valid defaults.
func NewInviteBuilder(t *testing.T) *InviteBuilder {
	t.Helper()
	now := time.Now()
	email := fmt.Sprintf("guest-%d@example.com", now.UnixNano())
	name := "Test Guest"
	hash := fmt.Sprintf("hash-%d", now.UnixNano())
	return &InviteBuilder{
		t: t,
		invite: &models.Invite{
			EventID:   1,
			Email:     &email,
			Name:      &name,
			TokenHash: hash,
			Status:    models.InviteStatusDraft,
			ExpiresAt: now.Add(30 * 24 * time.Hour),
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func (b *InviteBuilder) WithEventID(id int64) *InviteBuilder {
	b.invite.EventID = id
	return b
}

func (b *InviteBuilder) WithEmail(email string) *InviteBuilder {
	b.invite.Email = &email
	return b
}

func (b *InviteBuilder) WithName(name string) *InviteBuilder {
	b.invite.Name = &name
	return b
}

func (b *InviteBuilder) WithStatus(status models.InviteStatus) *InviteBuilder {
	b.invite.Status = status
	return b
}

func (b *InviteBuilder) WithMaxPlusOnes(n int) *InviteBuilder {
	b.invite.MaxPlusOnes = n
	return b
}

func (b *InviteBuilder) WithTokenHash(hash string) *InviteBuilder {
	b.invite.TokenHash = hash
	return b
}

func (b *InviteBuilder) WithExpiresAt(t time.Time) *InviteBuilder {
	b.invite.ExpiresAt = t
	return b
}

// Expired sets ExpiresAt to the past.
func (b *InviteBuilder) Expired() *InviteBuilder {
	past := time.Now().Add(-24 * time.Hour)
	b.invite.ExpiresAt = past
	return b
}

// Sent sets the status to sent.
func (b *InviteBuilder) Sent() *InviteBuilder {
	return b.WithStatus(models.InviteStatusSent)
}

// Revoked sets the status to revoked.
func (b *InviteBuilder) Revoked() *InviteBuilder {
	reason := "manually revoked"
	b.invite.Status = models.InviteStatusRevoked
	b.invite.RevocationReason = &reason
	return b
}

// WithRevocationReason sets a custom revocation reason (also sets status to revoked).
func (b *InviteBuilder) WithRevocationReason(reason string) *InviteBuilder {
	b.invite.Status = models.InviteStatusRevoked
	b.invite.RevocationReason = &reason
	return b
}

// Unsubscribed marks the invite as unsubscribed.
func (b *InviteBuilder) Unsubscribed() *InviteBuilder {
	b.invite.Unsubscribed = true
	return b
}

// Build returns the constructed Invite.
func (b *InviteBuilder) Build() *models.Invite {
	b.t.Helper()
	inv := *b.invite
	return &inv
}
