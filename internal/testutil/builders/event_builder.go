package builders

import (
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

// EventBuilder constructs models.Event instances with sensible defaults.
// Use Build() to get an in-memory struct, or BuildAndCreate() to persist it.
type EventBuilder struct {
	t     *testing.T
	event *models.Event
}

// NewEventBuilder returns a builder pre-populated with valid defaults.
func NewEventBuilder(t *testing.T) *EventBuilder {
	t.Helper()
	now := time.Now()
	title := fmt.Sprintf("Test Event %d", now.UnixNano())
	return &EventBuilder{
		t: t,
		event: &models.Event{
			Title:     title,
			StartTime: now.Add(24 * time.Hour),
			EndTime:   timePtr(now.Add(26 * time.Hour)),
			Timezone:  "UTC",
			Status:    models.EventStatusDraft,
			CreatedBy: 1,
		},
	}
}

func (b *EventBuilder) WithTitle(title string) *EventBuilder {
	b.event.Title = title
	return b
}

func (b *EventBuilder) WithStatus(status models.EventStatus) *EventBuilder {
	b.event.Status = status
	return b
}

func (b *EventBuilder) WithCreator(creatorID int64) *EventBuilder {
	b.event.CreatedBy = creatorID
	return b
}

func (b *EventBuilder) WithDescription(desc string) *EventBuilder {
	b.event.Description = &desc
	return b
}

func (b *EventBuilder) WithLocation(loc string) *EventBuilder {
	b.event.Location = &loc
	return b
}

func (b *EventBuilder) WithCapacity(cap int) *EventBuilder {
	b.event.EventCapacity = &cap
	return b
}

func (b *EventBuilder) WithMaxPlusOnes(n int) *EventBuilder {
	b.event.MaxPlusOnes = n
	return b
}

// Draft sets the status to draft.
func (b *EventBuilder) Draft() *EventBuilder {
	return b.WithStatus(models.EventStatusDraft)
}

// Published sets the status to published.
func (b *EventBuilder) Published() *EventBuilder {
	return b.WithStatus(models.EventStatusPublished)
}

// Cancelled sets the status to cancelled.
func (b *EventBuilder) Cancelled() *EventBuilder {
	return b.WithStatus(models.EventStatusCancelled)
}

// InFuture sets start/end times to be in the future (default: +24h/+26h).
func (b *EventBuilder) InFuture() *EventBuilder {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(2 * time.Hour)
	b.event.StartTime = start
	b.event.EndTime = &end
	return b
}

// InPast sets start/end times to be in the past.
func (b *EventBuilder) InPast() *EventBuilder {
	start := time.Now().Add(-48 * time.Hour)
	end := start.Add(2 * time.Hour)
	b.event.StartTime = start
	b.event.EndTime = &end
	return b
}

// WithRSVPDeadline sets the RSVP deadline.
func (b *EventBuilder) WithRSVPDeadline(d time.Time) *EventBuilder {
	b.event.RSVPDeadline = &d
	return b
}

// AllowMaybe enables maybe responses.
func (b *EventBuilder) AllowMaybe() *EventBuilder {
	b.event.AllowMaybeRSVP = true
	return b
}

// Build returns the constructed Event.
func (b *EventBuilder) Build() *models.Event {
	b.t.Helper()
	// Return a copy so callers don't share state
	e := *b.event
	return &e
}

func timePtr(t time.Time) *time.Time {
	return &t
}
