package auth

import (
	"context"
	"testing"

	"github.com/yourusername/tinyrsvp/internal/models"
)

func TestWithUser(t *testing.T) {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}

	ctx := WithUser(context.Background(), user)

	retrievedUser, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("Expected user in context")
	}

	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}

	if retrievedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrievedUser.Email)
	}
}

func TestUserFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	_, ok := UserFromContext(ctx)
	if ok {
		t.Error("Expected no user in context")
	}
}

func TestWithSession(t *testing.T) {
	session := &models.Session{
		ID:     "session123",
		UserID: 1,
	}

	ctx := WithSession(context.Background(), session)

	retrievedSession, ok := SessionFromContext(ctx)
	if !ok {
		t.Fatal("Expected session in context")
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrievedSession.ID)
	}

	if retrievedSession.UserID != session.UserID {
		t.Errorf("Expected user ID %d, got %d", session.UserID, retrievedSession.UserID)
	}
}

func TestSessionFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	_, ok := SessionFromContext(ctx)
	if ok {
		t.Error("Expected no session in context")
	}
}

func TestContextChaining(t *testing.T) {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleAdmin,
	}

	session := &models.Session{
		ID:     "session123",
		UserID: 1,
	}

	ctx := WithUser(context.Background(), user)
	ctx = WithSession(ctx, session)

	retrievedUser, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("Expected user in context")
	}

	retrievedSession, ok := SessionFromContext(ctx)
	if !ok {
		t.Fatal("Expected session in context")
	}

	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, retrievedUser.ID)
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrievedSession.ID)
	}
}
