package testutil

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

// CreateTestContext creates a context with the given user attached.
// Useful for testing operations that require authentication.
//
// If user is nil, returns a context with no user attached.
//
// Example:
//
//	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleAdmin}
//	ctx := testutil.CreateTestContext(user)
//	result, err := service.DoSomething(ctx, eventID)
func CreateTestContext(user *models.User) context.Context {
	if user == nil {
		return context.Background()
	}
	return auth.WithUser(context.Background(), user)
}

// CreateAdminContext creates a context with a test admin user.
// Useful for testing admin-only operations.
//
// The admin user has:
//   - ID: 1
//   - Email: admin@test.example.com
//   - Name: Test Admin
//   - Role: RoleAdmin
//
// Example:
//
//	ctx := testutil.CreateAdminContext()
//	err := adminService.DeleteUser(ctx, userID)
func CreateAdminContext() context.Context {
	user := &models.User{
		ID:    1,
		Email: "admin@test.example.com",
		Name:  "Test Admin",
		Role:  models.RoleAdmin,
	}
	return auth.WithUser(context.Background(), user)
}

// CreateEventManagerContext creates a context with a test event manager user.
// Useful for testing event manager operations.
//
// The event manager user has:
//   - ID: 2
//   - Email: manager@test.example.com
//   - Name: Test Event Manager
//   - Role: RoleEventManager
//
// Example:
//
//	ctx := testutil.CreateEventManagerContext()
//	event, err := eventService.CreateEvent(ctx, eventData)
func CreateEventManagerContext() context.Context {
	user := &models.User{
		ID:    2,
		Email: "manager@test.example.com",
		Name:  "Test Event Manager",
		Role:  models.RoleEventManager,
	}
	return auth.WithUser(context.Background(), user)
}

// CreateAnonymousContext creates a context with no user attached.
// Useful for testing unauthenticated operations.
//
// Example:
//
//	ctx := testutil.CreateAnonymousContext()
//	_, err := service.PublicOperation(ctx)
func CreateAnonymousContext() context.Context {
	return context.Background()
}
