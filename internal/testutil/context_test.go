package testutil_test

import (
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

func TestCreateTestContext(t *testing.T) {
	user := &models.User{
		ID:    123,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}

	ctx := testutil.CreateTestContext(user)

	// Verify user is in context
	ctxUser, ok := auth.UserFromContext(ctx)
	if !ok {
		t.Fatal("Expected user in context")
	}
	if ctxUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, ctxUser.ID)
	}
	if ctxUser.Email != user.Email {
		t.Errorf("Expected email %q, got %q", user.Email, ctxUser.Email)
	}
	if ctxUser.Name != user.Name {
		t.Errorf("Expected name %q, got %q", user.Name, ctxUser.Name)
	}
	if ctxUser.Role != user.Role {
		t.Errorf("Expected role %q, got %q", user.Role, ctxUser.Role)
	}
}

func TestCreateTestContextWithNilUser(t *testing.T) {
	ctx := testutil.CreateTestContext(nil)

	// Verify nil user is handled correctly
	ctxUser, ok := auth.UserFromContext(ctx)
	if ok {
		t.Errorf("Expected no user in context for nil user, got %+v", ctxUser)
	}
}

func TestCreateAdminContext(t *testing.T) {
	ctx := testutil.CreateAdminContext()

	// Verify admin user is in context
	ctxUser, ok := auth.UserFromContext(ctx)
	if !ok {
		t.Fatal("Expected admin user in context")
	}

	// Verify user has admin role
	if ctxUser.Role != models.RoleAdmin {
		t.Errorf("Expected role %q, got %q", models.RoleAdmin, ctxUser.Role)
	}

	// Verify user has ID
	if ctxUser.ID == 0 {
		t.Error("Expected non-zero user ID")
	}

	// Verify user has email
	if ctxUser.Email == "" {
		t.Error("Expected non-empty email")
	}

	// Verify user has name
	if ctxUser.Name == "" {
		t.Error("Expected non-empty name")
	}

	// Verify IsAdmin returns true
	if !ctxUser.IsAdmin() {
		t.Error("Expected IsAdmin() to return true")
	}
}

func TestCreateEventManagerContext(t *testing.T) {
	ctx := testutil.CreateEventManagerContext()

	// Verify event manager user is in context
	ctxUser, ok := auth.UserFromContext(ctx)
	if !ok {
		t.Fatal("Expected event manager user in context")
	}

	// Verify user has event manager role
	if ctxUser.Role != models.RoleEventManager {
		t.Errorf("Expected role %q, got %q", models.RoleEventManager, ctxUser.Role)
	}

	// Verify user has ID
	if ctxUser.ID == 0 {
		t.Error("Expected non-zero user ID")
	}

	// Verify user has email
	if ctxUser.Email == "" {
		t.Error("Expected non-empty email")
	}

	// Verify user has name
	if ctxUser.Name == "" {
		t.Error("Expected non-empty name")
	}

	// Verify IsEventManager returns true
	if !ctxUser.IsEventManager() {
		t.Error("Expected IsEventManager() to return true")
	}

	// Verify is NOT admin
	if ctxUser.IsAdmin() {
		t.Error("Expected IsAdmin() to return false for event manager")
	}
}

func TestCreateAnonymousContext(t *testing.T) {
	ctx := testutil.CreateAnonymousContext()

	// Verify no user is in context
	ctxUser, ok := auth.UserFromContext(ctx)
	if ok {
		t.Errorf("Expected no user in anonymous context, got %+v", ctxUser)
	}
}

func TestCreateAdminContextConsistency(t *testing.T) {
	// Create multiple admin contexts and verify they're consistent
	ctx1 := testutil.CreateAdminContext()
	ctx2 := testutil.CreateAdminContext()

	user1, ok1 := auth.UserFromContext(ctx1)
	user2, ok2 := auth.UserFromContext(ctx2)

	if !ok1 || !ok2 {
		t.Fatal("Expected users in both contexts")
	}

	// Verify both contexts have the same admin user properties
	if user1.ID != user2.ID {
		t.Errorf("Expected consistent admin ID, got %d and %d", user1.ID, user2.ID)
	}
	if user1.Email != user2.Email {
		t.Errorf("Expected consistent admin email, got %q and %q", user1.Email, user2.Email)
	}
	if user1.Role != user2.Role {
		t.Errorf("Expected consistent admin role, got %q and %q", user1.Role, user2.Role)
	}
}

func TestCreateEventManagerContextConsistency(t *testing.T) {
	// Create multiple event manager contexts and verify they're consistent
	ctx1 := testutil.CreateEventManagerContext()
	ctx2 := testutil.CreateEventManagerContext()

	user1, ok1 := auth.UserFromContext(ctx1)
	user2, ok2 := auth.UserFromContext(ctx2)

	if !ok1 || !ok2 {
		t.Fatal("Expected users in both contexts")
	}

	// Verify both contexts have the same event manager user properties
	if user1.ID != user2.ID {
		t.Errorf("Expected consistent event manager ID, got %d and %d", user1.ID, user2.ID)
	}
	if user1.Email != user2.Email {
		t.Errorf("Expected consistent event manager email, got %q and %q", user1.Email, user2.Email)
	}
	if user1.Role != user2.Role {
		t.Errorf("Expected consistent event manager role, got %q and %q", user1.Role, user2.Role)
	}
}

func TestContextHelpersDifferentUsers(t *testing.T) {
	// Verify that admin and event manager contexts have different users
	adminCtx := testutil.CreateAdminContext()
	managerCtx := testutil.CreateEventManagerContext()

	adminUser, ok1 := auth.UserFromContext(adminCtx)
	managerUser, ok2 := auth.UserFromContext(managerCtx)

	if !ok1 || !ok2 {
		t.Fatal("Expected users in both contexts")
	}

	// Verify they have different IDs
	if adminUser.ID == managerUser.ID {
		t.Error("Expected admin and event manager to have different IDs")
	}

	// Verify they have different roles
	if adminUser.Role == managerUser.Role {
		t.Error("Expected admin and event manager to have different roles")
	}
}
