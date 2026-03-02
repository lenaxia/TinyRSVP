package builders_test

import (
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/builders"
)

func TestNewInviteBuilder_Defaults(t *testing.T) {
	inv := builders.NewInviteBuilder(t).Build()

	if inv.EventID != 1 {
		t.Errorf("default EventID = %d, want 1", inv.EventID)
	}
	if inv.Email == nil || *inv.Email == "" {
		t.Error("default Email should not be empty")
	}
	if inv.Name == nil || *inv.Name == "" {
		t.Error("default Name should not be empty")
	}
	if inv.TokenHash == "" {
		t.Error("default TokenHash should not be empty")
	}
	if inv.Status != models.InviteStatusDraft {
		t.Errorf("default Status = %q, want draft", inv.Status)
	}
	if inv.ExpiresAt.Before(time.Now()) {
		t.Error("default ExpiresAt should be in the future")
	}
}

func TestInviteBuilder_WithEventID(t *testing.T) {
	inv := builders.NewInviteBuilder(t).WithEventID(99).Build()
	if inv.EventID != 99 {
		t.Errorf("EventID = %d, want 99", inv.EventID)
	}
}

func TestInviteBuilder_WithEmail(t *testing.T) {
	inv := builders.NewInviteBuilder(t).WithEmail("custom@example.com").Build()
	if inv.Email == nil || *inv.Email != "custom@example.com" {
		t.Error("Email not set correctly")
	}
}

func TestInviteBuilder_WithName(t *testing.T) {
	inv := builders.NewInviteBuilder(t).WithName("Alice").Build()
	if inv.Name == nil || *inv.Name != "Alice" {
		t.Error("Name not set correctly")
	}
}

func TestInviteBuilder_WithStatus(t *testing.T) {
	inv := builders.NewInviteBuilder(t).WithStatus(models.InviteStatusViewed).Build()
	if inv.Status != models.InviteStatusViewed {
		t.Errorf("Status = %q, want viewed", inv.Status)
	}
}

func TestInviteBuilder_Sent(t *testing.T) {
	inv := builders.NewInviteBuilder(t).Sent().Build()
	if inv.Status != models.InviteStatusSent {
		t.Errorf("Status = %q, want sent", inv.Status)
	}
}

func TestInviteBuilder_Revoked(t *testing.T) {
	inv := builders.NewInviteBuilder(t).Revoked().Build()
	if inv.Status != models.InviteStatusRevoked {
		t.Errorf("Status = %q, want revoked", inv.Status)
	}
	if inv.RevocationReason == nil || *inv.RevocationReason == "" {
		t.Error("Revoked invite should have a RevocationReason")
	}
}

func TestInviteBuilder_WithRevocationReason(t *testing.T) {
	inv := builders.NewInviteBuilder(t).WithRevocationReason("no-show").Build()
	if inv.Status != models.InviteStatusRevoked {
		t.Errorf("Status = %q, want revoked", inv.Status)
	}
	if inv.RevocationReason == nil || *inv.RevocationReason != "no-show" {
		t.Error("RevocationReason not set correctly")
	}
}

func TestInviteBuilder_Unsubscribed(t *testing.T) {
	inv := builders.NewInviteBuilder(t).Unsubscribed().Build()
	if !inv.Unsubscribed {
		t.Error("Unsubscribed should be true")
	}
}

func TestInviteBuilder_WithMaxPlusOnes(t *testing.T) {
	inv := builders.NewInviteBuilder(t).WithMaxPlusOnes(2).Build()
	if inv.MaxPlusOnes != 2 {
		t.Errorf("MaxPlusOnes = %d, want 2", inv.MaxPlusOnes)
	}
}

func TestInviteBuilder_Expired(t *testing.T) {
	inv := builders.NewInviteBuilder(t).Expired().Build()
	if inv.ExpiresAt.After(time.Now()) {
		t.Error("Expired invite ExpiresAt should be in the past")
	}
}

func TestInviteBuilder_BuildReturnsIndependentCopies(t *testing.T) {
	b := builders.NewInviteBuilder(t).WithName("Original")
	i1 := b.Build()
	name := "Modified"
	i1.Name = &name
	i2 := b.Build()
	if i2.Name == nil || *i2.Name != "Original" {
		t.Error("Build() should return independent copies; modifying i1 affected i2")
	}
}
