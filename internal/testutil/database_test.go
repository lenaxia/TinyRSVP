package testutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

func TestSetupTestDB(t *testing.T) {
	db := testutil.SetupTestDB(t)

	if db == nil {
		t.Fatal("Expected non-nil database")
	}

	// Verify database is functional
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Verify we can execute queries
	_, err := db.Exec(ctx, "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Errorf("Failed to create test table: %v", err)
	}

	// Database should be automatically closed by t.Cleanup()
	// We don't need to manually close it
}

func TestSetupTestDBWithMigrations(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")

	if db == nil {
		t.Fatal("Expected non-nil database")
	}

	ctx := context.Background()

	// Verify migrations ran by checking if users table exists
	row := db.QueryRow(ctx, "SELECT COUNT(*) FROM users")
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("Failed to query users table (migrations may not have run): %v", err)
	}

	// Verify other core tables exist
	tables := []string{"events", "invites", "rsvps", "sessions", "templates"}
	for _, table := range tables {
		row := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+table)
		var c int
		if err := row.Scan(&c); err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}
}

func TestCreateTestUser(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
	ctx := context.Background()

	tests := []struct {
		name string
		role models.UserRole
	}{
		{"admin user", models.RoleAdmin},
		{"event manager user", models.RoleEventManager},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := testutil.CreateTestUser(t, db, tt.role)

			if user == nil {
				t.Fatal("Expected non-nil user")
			}

			if user.ID == 0 {
				t.Error("Expected user ID to be set")
			}

			if user.Email == "" {
				t.Error("Expected user email to be set")
			}

			if user.Role != tt.role {
				t.Errorf("Expected role %s, got %s", tt.role, user.Role)
			}

			if user.Name == "" {
				t.Error("Expected user name to be set")
			}

			// Verify user was actually created in database
			var dbEmail string
			row := db.QueryRow(ctx, "SELECT email FROM users WHERE id = ?", user.ID)
			if err := row.Scan(&dbEmail); err != nil {
				t.Errorf("Failed to query created user: %v", err)
			}

			if dbEmail != user.Email {
				t.Errorf("Expected email %s in DB, got %s", user.Email, dbEmail)
			}
		})
	}
}

func TestCreateTestUser_UniqueEmails(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")

	// Create multiple users to ensure unique emails
	user1 := testutil.CreateTestUser(t, db, models.RoleAdmin)
	user2 := testutil.CreateTestUser(t, db, models.RoleEventManager)

	if user1.Email == user2.Email {
		t.Error("Expected unique emails for different users")
	}

	if user1.ID == user2.ID {
		t.Error("Expected different user IDs")
	}
}

func TestCreateTestEvent(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
	ctx := context.Background()

	// Create a user first (events need a creator)
	user := testutil.CreateTestUser(t, db, models.RoleEventManager)

	eventID := testutil.CreateTestEvent(t, db, user.ID)

	if eventID == 0 {
		t.Fatal("Expected non-zero event ID")
	}

	// Verify event was created in database
	var title string
	var createdBy int64
	row := db.QueryRow(ctx, "SELECT title, created_by FROM events WHERE id = ?", eventID)
	if err := row.Scan(&title, &createdBy); err != nil {
		t.Fatalf("Failed to query created event: %v", err)
	}

	if title == "" {
		t.Error("Expected event title to be set")
	}

	if createdBy != user.ID {
		t.Errorf("Expected created_by %d, got %d", user.ID, createdBy)
	}
}

func TestCreateTestEvent_MultipleEvents(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
	ctx := context.Background()

	user := testutil.CreateTestUser(t, db, models.RoleEventManager)

	// Create multiple events
	eventID1 := testutil.CreateTestEvent(t, db, user.ID)
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	eventID2 := testutil.CreateTestEvent(t, db, user.ID)

	if eventID1 == eventID2 {
		t.Error("Expected different event IDs")
	}

	// Verify both events exist
	var title1, title2 string
	db.QueryRow(ctx, "SELECT title FROM events WHERE id = ?", eventID1).Scan(&title1)
	db.QueryRow(ctx, "SELECT title FROM events WHERE id = ?", eventID2).Scan(&title2)

	if title1 == "" || title2 == "" {
		t.Error("Expected both events to have titles")
	}
}

func TestCreateTestInvite(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
	ctx := context.Background()

	// Setup prerequisites
	user := testutil.CreateTestUser(t, db, models.RoleEventManager)
	eventID := testutil.CreateTestEvent(t, db, user.ID)

	tokenHash := "test-token-hash-123"
	inviteID := testutil.CreateTestInvite(t, db, eventID, tokenHash)

	if inviteID == 0 {
		t.Fatal("Expected non-zero invite ID")
	}

	// Verify invite was created in database
	var dbEventID int64
	var dbTokenHash string
	row := db.QueryRow(ctx, "SELECT event_id, token_hash FROM invites WHERE id = ?", inviteID)
	if err := row.Scan(&dbEventID, &dbTokenHash); err != nil {
		t.Fatalf("Failed to query created invite: %v", err)
	}

	if dbEventID != eventID {
		t.Errorf("Expected event_id %d, got %d", eventID, dbEventID)
	}

	if dbTokenHash != tokenHash {
		t.Errorf("Expected token_hash %s, got %s", tokenHash, dbTokenHash)
	}
}

func TestCreateTestInvite_MultipleInvites(t *testing.T) {
	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")

	user := testutil.CreateTestUser(t, db, models.RoleEventManager)
	eventID := testutil.CreateTestEvent(t, db, user.ID)

	// Create multiple invites with unique token hashes
	inviteID1 := testutil.CreateTestInvite(t, db, eventID, "token-hash-1")
	inviteID2 := testutil.CreateTestInvite(t, db, eventID, "token-hash-2")

	if inviteID1 == inviteID2 {
		t.Error("Expected different invite IDs")
	}

	if inviteID1 == 0 || inviteID2 == 0 {
		t.Error("Expected both invites to be created")
	}
}

// TestDatabaseCleanup verifies that t.Cleanup() properly closes the database
func TestDatabaseCleanup(t *testing.T) {
	// This test verifies cleanup behavior
	// We can't directly test cleanup, but we can verify the pattern works
	db := testutil.SetupTestDB(t)

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Database should be accessible during test: %v", err)
	}

	// Database will be automatically closed after this test completes
	// via t.Cleanup(). If there are issues, other tests will fail.
}
