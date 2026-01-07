package auth

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type AuthorizationChecker interface {
	CanCreateEvent(ctx context.Context, user *models.User) bool
	CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool
	CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool
	CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool
	CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool
	CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool
	CanManageUsers(ctx context.Context, user *models.User) bool
	CanConfigureSystem(ctx context.Context, user *models.User) bool
	IsAdmin(user *models.User) bool
	IsEventManager(user *models.User) bool
}

type authorizationChecker struct{}

func NewAuthorizationChecker() AuthorizationChecker {
	return &authorizationChecker{}
}

func (c *authorizationChecker) IsAdmin(user *models.User) bool {
	if user == nil {
		return false
	}
	return user.Role == models.RoleAdmin
}

func (c *authorizationChecker) IsEventManager(user *models.User) bool {
	if user == nil {
		return false
	}
	return user.Role == models.RoleEventManager || user.Role == models.RoleAdmin
}

func (c *authorizationChecker) CanCreateEvent(ctx context.Context, user *models.User) bool {
	return c.IsEventManager(user)
}

func (c *authorizationChecker) CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if user == nil || event == nil {
		return false
	}
	if c.IsAdmin(user) {
		return true
	}
	return user.ID == event.CreatedBy
}

func (c *authorizationChecker) CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if user == nil || event == nil {
		return false
	}
	if c.IsAdmin(user) {
		return true
	}
	if user.ID == event.CreatedBy {
		return event.Status == models.EventStatusDraft || event.Status == models.EventStatusPublished
	}
	return false
}

func (c *authorizationChecker) CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if user == nil || event == nil {
		return false
	}
	return c.IsEventManager(user)
}

func (c *authorizationChecker) CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool {
	return c.CanEditEvent(ctx, user, event)
}

func (c *authorizationChecker) CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool {
	return c.CanEditEvent(ctx, user, event)
}

func (c *authorizationChecker) CanManageUsers(ctx context.Context, user *models.User) bool {
	return c.IsAdmin(user)
}

func (c *authorizationChecker) CanConfigureSystem(ctx context.Context, user *models.User) bool {
	return c.IsAdmin(user)
}
