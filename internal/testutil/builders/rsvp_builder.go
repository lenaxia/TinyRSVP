package builders

import (
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

// RSVPBuilder constructs models.RSVP instances with sensible defaults.
type RSVPBuilder struct {
	t    *testing.T
	rsvp *models.RSVP
}

// NewRSVPBuilder returns a builder pre-populated with valid defaults.
func NewRSVPBuilder(t *testing.T) *RSVPBuilder {
	t.Helper()
	now := time.Now()
	return &RSVPBuilder{
		t: t,
		rsvp: &models.RSVP{
			InviteID:  1,
			Response:  models.RSVPResponseYes,
			PlusOnes:  0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func (b *RSVPBuilder) WithInviteID(id int64) *RSVPBuilder {
	b.rsvp.InviteID = id
	return b
}

func (b *RSVPBuilder) WithResponse(response models.RSVPResponse) *RSVPBuilder {
	b.rsvp.Response = response
	return b
}

func (b *RSVPBuilder) WithPlusOnes(n int) *RSVPBuilder {
	b.rsvp.PlusOnes = n
	return b
}

func (b *RSVPBuilder) WithAdultsCount(n int) *RSVPBuilder {
	b.rsvp.AdultsCount = &n
	return b
}

func (b *RSVPBuilder) WithKidsCount(n int) *RSVPBuilder {
	b.rsvp.KidsCount = &n
	return b
}

// Yes sets the response to yes.
func (b *RSVPBuilder) Yes() *RSVPBuilder {
	return b.WithResponse(models.RSVPResponseYes)
}

// No sets the response to no.
func (b *RSVPBuilder) No() *RSVPBuilder {
	return b.WithResponse(models.RSVPResponseNo)
}

// Maybe sets the response to maybe.
func (b *RSVPBuilder) Maybe() *RSVPBuilder {
	return b.WithResponse(models.RSVPResponseMaybe)
}

// Build returns the constructed RSVP.
func (b *RSVPBuilder) Build() *models.RSVP {
	b.t.Helper()
	r := *b.rsvp
	return &r
}
