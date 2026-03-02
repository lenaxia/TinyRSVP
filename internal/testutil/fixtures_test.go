package testutil_test

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

func TestLoadFixture_ReadsFile(t *testing.T) {
	data := testutil.LoadFixture(t, "events/sample_event.json")
	if len(data) == 0 {
		t.Error("LoadFixture returned empty data")
	}
}

func TestLoadFixtureString_ReturnsString(t *testing.T) {
	s := testutil.LoadFixtureString(t, "events/sample_event.json")
	if s == "" {
		t.Error("LoadFixtureString returned empty string")
	}
	if s[0] != '{' {
		t.Errorf("expected JSON object, got %q", s[:1])
	}
}

func TestLoadFixtureJSON_UnmarshalsEvent(t *testing.T) {
	var event models.Event
	testutil.LoadFixtureJSON(t, "events/sample_event.json", &event)

	if event.ID != 1 {
		t.Errorf("ID = %d, want 1", event.ID)
	}
	if event.Title != "Sample Event" {
		t.Errorf("Title = %q, want Sample Event", event.Title)
	}
	if event.Status != models.EventStatusDraft {
		t.Errorf("Status = %q, want draft", event.Status)
	}
	if event.Timezone != "UTC" {
		t.Errorf("Timezone = %q, want UTC", event.Timezone)
	}
}

func TestFixtureDir_ReturnsPath(t *testing.T) {
	dir := testutil.FixtureDir(t)
	if dir == "" {
		t.Error("FixtureDir returned empty string")
	}
}
