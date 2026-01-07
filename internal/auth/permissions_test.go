package auth

import (
	"context"
	"testing"

	"github.com/yourusername/tinyrsvp/internal/models"
)

func TestAuthorizationChecker_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		user *models.User
		want bool
	}{
		{
			name: "admin user",
			user: &models.User{Role: models.RoleAdmin},
			want: true,
		},
		{
			name: "event manager user",
			user: &models.User{Role: models.RoleEventManager},
			want: false,
		},
		{
			name: "nil user",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.IsAdmin(tt.user)
			if got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_IsEventManager(t *testing.T) {
	tests := []struct {
		name string
		user *models.User
		want bool
	}{
		{
			name: "admin user is event manager",
			user: &models.User{Role: models.RoleAdmin},
			want: true,
		},
		{
			name: "event manager user",
			user: &models.User{Role: models.RoleEventManager},
			want: true,
		},
		{
			name: "nil user",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.IsEventManager(tt.user)
			if got != tt.want {
				t.Errorf("IsEventManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanCreateEvent(t *testing.T) {
	tests := []struct {
		name string
		user *models.User
		want bool
	}{
		{
			name: "admin can create event",
			user: &models.User{ID: 1, Role: models.RoleAdmin},
			want: true,
		},
		{
			name: "event manager can create event",
			user: &models.User{ID: 1, Role: models.RoleEventManager},
			want: true,
		},
		{
			name: "nil user cannot create event",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanCreateEvent(context.Background(), tt.user)
			if got != tt.want {
				t.Errorf("CanCreateEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanEditEvent(t *testing.T) {
	tests := []struct {
		name  string
		user  *models.User
		event *models.Event
		want  bool
	}{
		{
			name:  "admin can edit any event",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  true,
		},
		{
			name:  "owner can edit own event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1},
			want:  true,
		},
		{
			name:  "non-owner cannot edit event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil user cannot edit event",
			user:  nil,
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil event cannot be edited",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanEditEvent(context.Background(), tt.user, tt.event)
			if got != tt.want {
				t.Errorf("CanEditEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanDeleteEvent(t *testing.T) {
	tests := []struct {
		name  string
		user  *models.User
		event *models.Event
		want  bool
	}{
		{
			name:  "admin can delete any event",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: &models.Event{ID: 100, CreatedBy: 2, Status: models.EventStatusPublished},
			want:  true,
		},
		{
			name:  "owner can delete draft event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusDraft},
			want:  true,
		},
		{
			name:  "owner can delete published event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusPublished},
			want:  true,
		},
		{
			name:  "owner cannot delete completed event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusCompleted},
			want:  false,
		},
		{
			name:  "owner cannot delete cancelled event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusCancelled},
			want:  false,
		},
		{
			name:  "owner cannot delete archived event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusArchived},
			want:  false,
		},
		{
			name:  "non-owner cannot delete event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 2, Status: models.EventStatusDraft},
			want:  false,
		},
		{
			name:  "nil user cannot delete event",
			user:  nil,
			event: &models.Event{ID: 100, CreatedBy: 2, Status: models.EventStatusDraft},
			want:  false,
		},
		{
			name:  "nil event cannot be deleted",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanDeleteEvent(context.Background(), tt.user, tt.event)
			if got != tt.want {
				t.Errorf("CanDeleteEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanViewEvent(t *testing.T) {
	tests := []struct {
		name  string
		user  *models.User
		event *models.Event
		want  bool
	}{
		{
			name:  "admin can view any event",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  true,
		},
		{
			name:  "event manager can view any event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  true,
		},
		{
			name:  "nil user cannot view event",
			user:  nil,
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil event cannot be viewed",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanViewEvent(context.Background(), tt.user, tt.event)
			if got != tt.want {
				t.Errorf("CanViewEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanManageInvites(t *testing.T) {
	tests := []struct {
		name  string
		user  *models.User
		event *models.Event
		want  bool
	}{
		{
			name:  "admin can manage invites for any event",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  true,
		},
		{
			name:  "owner can manage invites for own event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1},
			want:  true,
		},
		{
			name:  "non-owner cannot manage invites",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil user cannot manage invites",
			user:  nil,
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil event invites cannot be managed",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanManageInvites(context.Background(), tt.user, tt.event)
			if got != tt.want {
				t.Errorf("CanManageInvites() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanViewRSVPs(t *testing.T) {
	tests := []struct {
		name  string
		user  *models.User
		event *models.Event
		want  bool
	}{
		{
			name:  "admin can view RSVPs for any event",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  true,
		},
		{
			name:  "owner can view RSVPs for own event",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 1},
			want:  true,
		},
		{
			name:  "non-owner cannot view RSVPs",
			user:  &models.User{ID: 1, Role: models.RoleEventManager},
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil user cannot view RSVPs",
			user:  nil,
			event: &models.Event{ID: 100, CreatedBy: 2},
			want:  false,
		},
		{
			name:  "nil event RSVPs cannot be viewed",
			user:  &models.User{ID: 1, Role: models.RoleAdmin},
			event: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanViewRSVPs(context.Background(), tt.user, tt.event)
			if got != tt.want {
				t.Errorf("CanViewRSVPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanManageUsers(t *testing.T) {
	tests := []struct {
		name string
		user *models.User
		want bool
	}{
		{
			name: "admin can manage users",
			user: &models.User{Role: models.RoleAdmin},
			want: true,
		},
		{
			name: "event manager cannot manage users",
			user: &models.User{Role: models.RoleEventManager},
			want: false,
		},
		{
			name: "nil user cannot manage users",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanManageUsers(context.Background(), tt.user)
			if got != tt.want {
				t.Errorf("CanManageUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationChecker_CanConfigureSystem(t *testing.T) {
	tests := []struct {
		name string
		user *models.User
		want bool
	}{
		{
			name: "admin can configure system",
			user: &models.User{Role: models.RoleAdmin},
			want: true,
		},
		{
			name: "event manager cannot configure system",
			user: &models.User{Role: models.RoleEventManager},
			want: false,
		},
		{
			name: "nil user cannot configure system",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewAuthorizationChecker()

			got := checker.CanConfigureSystem(context.Background(), tt.user)
			if got != tt.want {
				t.Errorf("CanConfigureSystem() = %v, want %v", got, tt.want)
			}
		})
	}
}
