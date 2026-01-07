package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func createTestUser(t *testing.T, repo UserRepository) *models.User {
	t.Helper()

	user := &models.User{
		Email: fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Name:  "Test User",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func TestSessionRepository_Create(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	tests := []struct {
		name    string
		session *models.Session
		wantErr bool
	}{
		{
			name: "valid session",
			session: &models.Session{
				ID:        "session-abc-123",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "session with IP and user agent",
			session: &models.Session{
				ID:        "session-with-metadata",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				IPAddress: stringPtr("192.168.1.1"),
				UserAgent: stringPtr("Mozilla/5.0"),
			},
			wantErr: false,
		},
		{
			name: "duplicate session ID",
			session: &models.Session{
				ID:        "session-abc-123",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionRepo.Create(ctx, tt.session)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.session.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.session.LastAccessedAt.IsZero() {
					t.Error("Expected LastAccessedAt to be set")
				}
			}
		})
	}
}

func TestSessionRepository_GetByID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	session := &models.Session{
		ID:        "get-by-id-session",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing session",
			id:      "get-by-id-session",
			wantErr: false,
		},
		{
			name:    "non-existing session",
			id:      "non-existing-session",
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := sessionRepo.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Expected error type %T, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if found.ID != tt.id {
					t.Errorf("Expected ID %s, got %s", tt.id, found.ID)
				}
				if found.UserID != user.ID {
					t.Errorf("Expected UserID %d, got %d", user.ID, found.UserID)
				}
			}
		})
	}
}

func TestSessionRepository_GetByUserID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	for i := 0; i < 3; i++ {
		session := &models.Session{
			ID:        fmt.Sprintf("user-session-%d", i),
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := sessionRepo.Create(ctx, session); err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	t.Run("get sessions for user with sessions", func(t *testing.T) {
		sessions, err := sessionRepo.GetByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(sessions) != 3 {
			t.Errorf("Expected 3 sessions, got %d", len(sessions))
		}
	})

	t.Run("get sessions for user without sessions", func(t *testing.T) {
		anotherUser := createTestUser(t, userRepo)
		sessions, err := sessionRepo.GetByUserID(ctx, anotherUser.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(sessions))
		}
	})
}

func TestSessionRepository_Update(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	session := &models.Session{
		ID:        "update-session",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	t.Run("update expires at", func(t *testing.T) {
		newExpiry := time.Now().Add(48 * time.Hour)
		session.ExpiresAt = newExpiry

		if err := sessionRepo.Update(ctx, session); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		found, err := sessionRepo.GetByID(ctx, session.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if found.ExpiresAt.Unix() != newExpiry.Unix() {
			t.Errorf("Expected ExpiresAt %v, got %v", newExpiry, found.ExpiresAt)
		}
	})

	t.Run("update non-existing session", func(t *testing.T) {
		nonExisting := &models.Session{
			ID:        "non-existing",
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := sessionRepo.Update(ctx, nonExisting)
		if err == nil {
			t.Error("Expected error for non-existing session")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	session := &models.Session{
		ID:        "delete-session",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	t.Run("delete existing session", func(t *testing.T) {
		if err := sessionRepo.Delete(ctx, session.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err := sessionRepo.GetByID(ctx, session.ID)
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Error("Expected NotFoundError after deletion")
		}
	})

	t.Run("delete non-existing session", func(t *testing.T) {
		err := sessionRepo.Delete(ctx, "non-existing")
		if err == nil {
			t.Error("Expected error for non-existing session")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestSessionRepository_DeleteByUserID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	for i := 0; i < 3; i++ {
		session := &models.Session{
			ID:        fmt.Sprintf("delete-by-user-%d", i),
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := sessionRepo.Create(ctx, session); err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	t.Run("delete all sessions for user", func(t *testing.T) {
		if err := sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
			t.Fatalf("DeleteByUserID() error = %v", err)
		}

		sessions, err := sessionRepo.GetByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after deletion, got %d", len(sessions))
		}
	})

	t.Run("delete sessions for user with no sessions", func(t *testing.T) {
		anotherUser := createTestUser(t, userRepo)
		if err := sessionRepo.DeleteByUserID(ctx, anotherUser.ID); err != nil {
			t.Fatalf("DeleteByUserID() error = %v", err)
		}
	})
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	expiredSession1 := &models.Session{
		ID:        "expired-session-1",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(-2 * time.Hour),
	}

	expiredSession2 := &models.Session{
		ID:        "expired-session-2",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	validSession := &models.Session{
		ID:        "valid-session",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := sessionRepo.Create(ctx, expiredSession1); err != nil {
		t.Fatalf("Failed to create expired session 1: %v", err)
	}

	if err := sessionRepo.Create(ctx, expiredSession2); err != nil {
		t.Fatalf("Failed to create expired session 2: %v", err)
	}

	if err := sessionRepo.Create(ctx, validSession); err != nil {
		t.Fatalf("Failed to create valid session: %v", err)
	}

	t.Run("delete expired sessions", func(t *testing.T) {
		deleted, err := sessionRepo.DeleteExpired(ctx)
		if err != nil {
			t.Fatalf("DeleteExpired() error = %v", err)
		}

		if deleted != 2 {
			t.Errorf("Expected 2 deleted sessions, got %d", deleted)
		}

		_, err = sessionRepo.GetByID(ctx, "expired-session-1")
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Error("Expected NotFoundError for expired session 1")
		}

		_, err = sessionRepo.GetByID(ctx, "expired-session-2")
		if !errors.As(err, &notFoundErr) {
			t.Error("Expected NotFoundError for expired session 2")
		}

		_, err = sessionRepo.GetByID(ctx, "valid-session")
		if err != nil {
			t.Error("Valid session should still exist")
		}
	})

	t.Run("delete expired when none exist", func(t *testing.T) {
		deleted, err := sessionRepo.DeleteExpired(ctx)
		if err != nil {
			t.Fatalf("DeleteExpired() error = %v", err)
		}

		if deleted != 0 {
			t.Errorf("Expected 0 deleted sessions, got %d", deleted)
		}
	})
}

func TestSessionRepository_UpdateLastAccessed(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	session := &models.Session{
		ID:        "update-last-accessed",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	originalLastAccessed := session.LastAccessedAt

	time.Sleep(10 * time.Millisecond)

	t.Run("update last accessed", func(t *testing.T) {
		if err := sessionRepo.UpdateLastAccessed(ctx, session.ID); err != nil {
			t.Fatalf("UpdateLastAccessed() error = %v", err)
		}

		found, err := sessionRepo.GetByID(ctx, session.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if !found.LastAccessedAt.After(originalLastAccessed) {
			t.Error("LastAccessedAt should be updated to a later time")
		}
	})

	t.Run("update last accessed for non-existing session", func(t *testing.T) {
		err := sessionRepo.UpdateLastAccessed(ctx, "non-existing")
		if err == nil {
			t.Error("Expected error for non-existing session")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestSessionRepository_CascadeDelete(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	session := &models.Session{
		ID:        "cascade-session",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	t.Run("sessions deleted when user deleted", func(t *testing.T) {
		if err := userRepo.Delete(ctx, user.ID); err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		_, err := sessionRepo.GetByID(ctx, session.ID)
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Error("Expected session to be deleted via cascade")
		}
	})
}

func TestSessionRepository_MultipleUsersMultipleSessions(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user1 := createTestUser(t, userRepo)
	user2 := createTestUser(t, userRepo)

	for i := 0; i < 2; i++ {
		session := &models.Session{
			ID:        fmt.Sprintf("user1-session-%d", i),
			UserID:    user1.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := sessionRepo.Create(ctx, session); err != nil {
			t.Fatalf("Failed to create session for user1: %v", err)
		}
	}

	for i := 0; i < 3; i++ {
		session := &models.Session{
			ID:        fmt.Sprintf("user2-session-%d", i),
			UserID:    user2.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := sessionRepo.Create(ctx, session); err != nil {
			t.Fatalf("Failed to create session for user2: %v", err)
		}
	}

	t.Run("each user has correct number of sessions", func(t *testing.T) {
		user1Sessions, err := sessionRepo.GetByUserID(ctx, user1.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(user1Sessions) != 2 {
			t.Errorf("Expected 2 sessions for user1, got %d", len(user1Sessions))
		}

		user2Sessions, err := sessionRepo.GetByUserID(ctx, user2.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(user2Sessions) != 3 {
			t.Errorf("Expected 3 sessions for user2, got %d", len(user2Sessions))
		}
	})

	t.Run("delete sessions for one user doesn't affect other", func(t *testing.T) {
		if err := sessionRepo.DeleteByUserID(ctx, user1.ID); err != nil {
			t.Fatalf("DeleteByUserID() error = %v", err)
		}

		user1Sessions, err := sessionRepo.GetByUserID(ctx, user1.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(user1Sessions) != 0 {
			t.Errorf("Expected 0 sessions for user1, got %d", len(user1Sessions))
		}

		user2Sessions, err := sessionRepo.GetByUserID(ctx, user2.ID)
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}

		if len(user2Sessions) != 3 {
			t.Errorf("Expected 3 sessions for user2, got %d", len(user2Sessions))
		}
	})
}

func TestSessionRepository_ExpiredSessionHandling(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewUserRepository(database)
	sessionRepo := NewSessionRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	t.Run("can retrieve expired session", func(t *testing.T) {
		expiredSession := &models.Session{
			ID:        "already-expired",
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		if err := sessionRepo.Create(ctx, expiredSession); err != nil {
			t.Fatalf("Failed to create expired session: %v", err)
		}

		found, err := sessionRepo.GetByID(ctx, "already-expired")
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if !found.IsExpired() {
			t.Error("Expected session to be expired")
		}
	})
}
