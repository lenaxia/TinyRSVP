package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

// SetupTestDB creates an in-memory SQLite database for testing.
// The database is automatically closed when the test completes via t.Cleanup().
// Migrations are NOT run - use SetupTestDBWithMigrations for that.
//
// Example:
//
//	func TestMyFeature(t *testing.T) {
//	    db := testutil.SetupTestDB(t)
//	    // db is automatically closed after test
//	    // ... use db for testing
//	}
func SetupTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

// SetupTestDBWithMigrations creates a test database and runs all migrations.
// The database is automatically closed when the test completes.
// The migrationPath should be relative to the test file location.
//
// Example:
//
//	func TestMyFeature(t *testing.T) {
//	    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
//	    user := testutil.CreateTestUser(t, db, models.RoleAdmin)
//	    // ... use db and user for testing
//	}
func SetupTestDBWithMigrations(t *testing.T, migrationPath string) db.Database {
	t.Helper()

	database := SetupTestDB(t)

	migrator, err := db.NewMigrator(database.DB(), migrationPath)
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

// CreateTestUser creates a test user with the given role.
// The user is persisted to the database and returned with its ID set.
// Email addresses are automatically generated with timestamps to ensure uniqueness.
//
// Example:
//
//	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
//	admin := testutil.CreateTestUser(t, db, models.RoleAdmin)
//	manager := testutil.CreateTestUser(t, db, models.RoleEventManager)
func CreateTestUser(t *testing.T, database db.Database, role models.UserRole) *models.User {
	t.Helper()

	repo := repositories.NewUserRepository(database)

	// Generate unique email with nanosecond precision timestamp
	timestamp := time.Now().Format("20060102150405.000000")
	user := &models.User{
		Email:     fmt.Sprintf("test-%s@example.com", timestamp),
		Name:      "Test User",
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestEvent creates a test event and returns its ID.
// The event is created with sensible defaults:
// - Starts 7 days from now
// - Duration of 2 hours
// - Status: draft
// - Max plus ones: 2
// - Only uses columns from the base schema for maximum compatibility
//
// Example:
//
//	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
//	user := testutil.CreateTestUser(t, db, models.RoleEventManager)
//	eventID := testutil.CreateTestEvent(t, db, user.ID)
func CreateTestEvent(t *testing.T, database db.Database, creatorID int64) int64 {
	t.Helper()

	startTime := time.Now().Add(7 * 24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	result, err := database.Exec(context.Background(), `
		INSERT INTO events (
			title, description, location, 
			start_time, end_time, timezone, status, created_by,
			max_plus_ones
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "Test Event", "Test Description", "Test Location",
		startTime, endTime, "UTC", "draft", creatorID, 2)

	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get event ID: %v", err)
	}

	return id
}

// CreateTestInvite creates a test invite and returns its ID.
// The invite is created with sensible defaults:
// - Expires 30 days from now
// - Status: draft
// - Max plus ones: 2
// - Test email and name
//
// Example:
//
//	db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
//	user := testutil.CreateTestUser(t, db, models.RoleEventManager)
//	eventID := testutil.CreateTestEvent(t, db, user.ID)
//	inviteID := testutil.CreateTestInvite(t, db, eventID, "unique-token-hash")
func CreateTestInvite(t *testing.T, database db.Database, eventID int64, tokenHash string) int64 {
	t.Helper()

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	result, err := database.Exec(context.Background(), `
		INSERT INTO invites (
			event_id, token_hash, email, name,
			max_plus_ones, status, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, eventID, tokenHash, "test@example.com", "Test Guest",
		2, "draft", expiresAt)

	if err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get invite ID: %v", err)
	}

	return id
}
