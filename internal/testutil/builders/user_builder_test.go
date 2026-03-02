package builders_test

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/builders"
)

func TestNewUserBuilder_Defaults(t *testing.T) {
	u := builders.NewUserBuilder(t).Build()

	if u.Email == "" {
		t.Error("default Email should not be empty")
	}
	if u.Name == "" {
		t.Error("default Name should not be empty")
	}
	if u.Role != models.RoleGuest {
		t.Errorf("default Role = %q, want guest", u.Role)
	}
}

func TestUserBuilder_WithID(t *testing.T) {
	u := builders.NewUserBuilder(t).WithID(77).Build()
	if u.ID != 77 {
		t.Errorf("ID = %d, want 77", u.ID)
	}
}

func TestUserBuilder_WithEmail(t *testing.T) {
	u := builders.NewUserBuilder(t).WithEmail("custom@test.com").Build()
	if u.Email != "custom@test.com" {
		t.Errorf("Email = %q, want custom@test.com", u.Email)
	}
}

func TestUserBuilder_WithName(t *testing.T) {
	u := builders.NewUserBuilder(t).WithName("Bob").Build()
	if u.Name != "Bob" {
		t.Errorf("Name = %q, want Bob", u.Name)
	}
}

func TestUserBuilder_Admin(t *testing.T) {
	u := builders.NewUserBuilder(t).Admin().Build()
	if u.Role != models.RoleAdmin {
		t.Errorf("Role = %q, want admin", u.Role)
	}
	if !u.IsAdmin() {
		t.Error("IsAdmin() should return true for admin user")
	}
}

func TestUserBuilder_EventManager(t *testing.T) {
	u := builders.NewUserBuilder(t).EventManager().Build()
	if u.Role != models.RoleEventManager {
		t.Errorf("Role = %q, want event_manager", u.Role)
	}
	if !u.IsEventManager() {
		t.Error("IsEventManager() should return true for event_manager user")
	}
}

func TestUserBuilder_Guest(t *testing.T) {
	u := builders.NewUserBuilder(t).Guest().Build()
	if u.Role != models.RoleGuest {
		t.Errorf("Role = %q, want guest", u.Role)
	}
}

func TestUserBuilder_WithOIDCSubject(t *testing.T) {
	u := builders.NewUserBuilder(t).WithOIDCSubject("sub-123").Build()
	if u.OIDCSubject == nil || *u.OIDCSubject != "sub-123" {
		t.Error("OIDCSubject not set correctly")
	}
}

func TestUserBuilder_BuildReturnsIndependentCopies(t *testing.T) {
	b := builders.NewUserBuilder(t).WithName("Original")
	u1 := b.Build()
	u1.Name = "Modified"
	u2 := b.Build()
	if u2.Name != "Original" {
		t.Error("Build() should return independent copies; modifying u1 affected u2")
	}
}
