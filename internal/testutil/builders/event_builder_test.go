package builders_test

import (
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/builders"
)

func TestNewEventBuilder_Defaults(t *testing.T) {
	e := builders.NewEventBuilder(t).Build()

	if e.Title == "" {
		t.Error("default Title should not be empty")
	}
	if e.Timezone != "UTC" {
		t.Errorf("default Timezone = %q, want UTC", e.Timezone)
	}
	if e.Status != models.EventStatusDraft {
		t.Errorf("default Status = %q, want draft", e.Status)
	}
	if e.CreatedBy != 1 {
		t.Errorf("default CreatedBy = %d, want 1", e.CreatedBy)
	}
	if e.StartTime.IsZero() {
		t.Error("default StartTime should not be zero")
	}
	if e.EndTime == nil {
		t.Error("default EndTime should not be nil")
	}
}

func TestEventBuilder_WithTitle(t *testing.T) {
	e := builders.NewEventBuilder(t).WithTitle("My Event").Build()
	if e.Title != "My Event" {
		t.Errorf("Title = %q, want %q", e.Title, "My Event")
	}
}

func TestEventBuilder_WithStatus(t *testing.T) {
	e := builders.NewEventBuilder(t).Published().Build()
	if e.Status != models.EventStatusPublished {
		t.Errorf("Status = %q, want published", e.Status)
	}
}

func TestEventBuilder_Cancelled(t *testing.T) {
	e := builders.NewEventBuilder(t).Cancelled().Build()
	if e.Status != models.EventStatusCancelled {
		t.Errorf("Status = %q, want cancelled", e.Status)
	}
}

func TestEventBuilder_WithCreator(t *testing.T) {
	e := builders.NewEventBuilder(t).WithCreator(42).Build()
	if e.CreatedBy != 42 {
		t.Errorf("CreatedBy = %d, want 42", e.CreatedBy)
	}
}

func TestEventBuilder_WithDescription(t *testing.T) {
	e := builders.NewEventBuilder(t).WithDescription("a description").Build()
	if e.Description == nil || *e.Description != "a description" {
		t.Error("Description not set correctly")
	}
}

func TestEventBuilder_WithLocation(t *testing.T) {
	e := builders.NewEventBuilder(t).WithLocation("Berlin").Build()
	if e.Location == nil || *e.Location != "Berlin" {
		t.Error("Location not set correctly")
	}
}

func TestEventBuilder_WithCapacity(t *testing.T) {
	e := builders.NewEventBuilder(t).WithCapacity(50).Build()
	if e.EventCapacity == nil || *e.EventCapacity != 50 {
		t.Error("EventCapacity not set correctly")
	}
}

func TestEventBuilder_WithMaxPlusOnes(t *testing.T) {
	e := builders.NewEventBuilder(t).WithMaxPlusOnes(3).Build()
	if e.MaxPlusOnes != 3 {
		t.Errorf("MaxPlusOnes = %d, want 3", e.MaxPlusOnes)
	}
}

func TestEventBuilder_InFuture(t *testing.T) {
	e := builders.NewEventBuilder(t).InFuture().Build()
	if !e.StartTime.After(time.Now()) {
		t.Error("InFuture: StartTime should be in the future")
	}
}

func TestEventBuilder_InPast(t *testing.T) {
	e := builders.NewEventBuilder(t).InPast().Build()
	if !e.StartTime.Before(time.Now()) {
		t.Error("InPast: StartTime should be in the past")
	}
}

func TestEventBuilder_WithRSVPDeadline(t *testing.T) {
	deadline := time.Now().Add(12 * time.Hour)
	e := builders.NewEventBuilder(t).WithRSVPDeadline(deadline).Build()
	if e.RSVPDeadline == nil {
		t.Fatal("RSVPDeadline should not be nil")
	}
	if !e.RSVPDeadline.Equal(deadline) {
		t.Errorf("RSVPDeadline = %v, want %v", e.RSVPDeadline, deadline)
	}
}

func TestEventBuilder_AllowMaybe(t *testing.T) {
	e := builders.NewEventBuilder(t).AllowMaybe().Build()
	if !e.AllowMaybeRSVP {
		t.Error("AllowMaybeRSVP should be true")
	}
}

func TestEventBuilder_BuildReturnsIndependentCopies(t *testing.T) {
	b := builders.NewEventBuilder(t).WithTitle("Original")
	e1 := b.Build()
	e1.Title = "Modified"
	e2 := b.Build()
	if e2.Title != "Original" {
		t.Error("Build() should return independent copies; modifying e1 affected e2")
	}
}
