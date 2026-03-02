package builders_test

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/builders"
)

func TestNewRSVPBuilder_Defaults(t *testing.T) {
	r := builders.NewRSVPBuilder(t).Build()

	if r.InviteID != 1 {
		t.Errorf("default InviteID = %d, want 1", r.InviteID)
	}
	if r.Response != models.RSVPResponseYes {
		t.Errorf("default Response = %q, want yes", r.Response)
	}
	if r.PlusOnes != 0 {
		t.Errorf("default PlusOnes = %d, want 0", r.PlusOnes)
	}
	if r.AdultsCount != nil {
		t.Error("default AdultsCount should be nil")
	}
	if r.KidsCount != nil {
		t.Error("default KidsCount should be nil")
	}
}

func TestRSVPBuilder_WithInviteID(t *testing.T) {
	r := builders.NewRSVPBuilder(t).WithInviteID(55).Build()
	if r.InviteID != 55 {
		t.Errorf("InviteID = %d, want 55", r.InviteID)
	}
}

func TestRSVPBuilder_Yes(t *testing.T) {
	r := builders.NewRSVPBuilder(t).Yes().Build()
	if r.Response != models.RSVPResponseYes {
		t.Errorf("Response = %q, want yes", r.Response)
	}
}

func TestRSVPBuilder_No(t *testing.T) {
	r := builders.NewRSVPBuilder(t).No().Build()
	if r.Response != models.RSVPResponseNo {
		t.Errorf("Response = %q, want no", r.Response)
	}
}

func TestRSVPBuilder_Maybe(t *testing.T) {
	r := builders.NewRSVPBuilder(t).Maybe().Build()
	if r.Response != models.RSVPResponseMaybe {
		t.Errorf("Response = %q, want maybe", r.Response)
	}
}

func TestRSVPBuilder_WithPlusOnes(t *testing.T) {
	r := builders.NewRSVPBuilder(t).WithPlusOnes(3).Build()
	if r.PlusOnes != 3 {
		t.Errorf("PlusOnes = %d, want 3", r.PlusOnes)
	}
}

func TestRSVPBuilder_WithAdultsCount(t *testing.T) {
	r := builders.NewRSVPBuilder(t).WithAdultsCount(2).Build()
	if r.AdultsCount == nil || *r.AdultsCount != 2 {
		t.Error("AdultsCount not set correctly")
	}
}

func TestRSVPBuilder_WithKidsCount(t *testing.T) {
	r := builders.NewRSVPBuilder(t).WithKidsCount(1).Build()
	if r.KidsCount == nil || *r.KidsCount != 1 {
		t.Error("KidsCount not set correctly")
	}
}

func TestRSVPBuilder_BuildReturnsIndependentCopies(t *testing.T) {
	b := builders.NewRSVPBuilder(t).WithPlusOnes(0)
	r1 := b.Build()
	r1.PlusOnes = 99
	r2 := b.Build()
	if r2.PlusOnes != 0 {
		t.Error("Build() should return independent copies; modifying r1 affected r2")
	}
}

func TestRSVPBuilder_Validate(t *testing.T) {
	r := builders.NewRSVPBuilder(t).Build()
	if err := r.Validate(); err != nil {
		t.Errorf("default RSVP should be valid, got error: %v", err)
	}
}
