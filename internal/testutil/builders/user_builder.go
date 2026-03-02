package builders

import (
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

// UserBuilder constructs models.User instances with sensible defaults.
type UserBuilder struct {
	t    *testing.T
	user *models.User
}

// NewUserBuilder returns a builder pre-populated with valid defaults.
func NewUserBuilder(t *testing.T) *UserBuilder {
	t.Helper()
	now := time.Now()
	return &UserBuilder{
		t: t,
		user: &models.User{
			Email:     fmt.Sprintf("user-%d@example.com", now.UnixNano()),
			Name:      "Test User",
			Role:      models.RoleGuest,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func (b *UserBuilder) WithID(id int64) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
	b.user.Name = name
	return b
}

func (b *UserBuilder) WithRole(role models.UserRole) *UserBuilder {
	b.user.Role = role
	return b
}

func (b *UserBuilder) WithOIDCSubject(subject string) *UserBuilder {
	b.user.OIDCSubject = &subject
	return b
}

// Admin sets the role to admin.
func (b *UserBuilder) Admin() *UserBuilder {
	return b.WithRole(models.RoleAdmin)
}

// EventManager sets the role to event_manager.
func (b *UserBuilder) EventManager() *UserBuilder {
	return b.WithRole(models.RoleEventManager)
}

// Guest sets the role to guest.
func (b *UserBuilder) Guest() *UserBuilder {
	return b.WithRole(models.RoleGuest)
}

// Build returns the constructed User.
func (b *UserBuilder) Build() *models.User {
	b.t.Helper()
	u := *b.user
	return &u
}
