# User Story: Database Test Helpers

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 hour
**Phase:** 1 - Foundation

---

## User Story

As a **developer**, I want **centralized database test helpers** so that **I can easily set up test databases and create test data without duplicating setup code across 23+ test files**.

---

## Acceptance Criteria

- [ ] `internal/testutil/database.go` created
- [ ] `SetupTestDB()` function implemented
- [ ] `SetupTestDBWithMigrations()` function implemented
- [ ] `CreateTestUser()` helper implemented
- [ ] `CreateTestEvent()` helper implemented
- [ ] `CreateTestInvite()` helper implemented
- [ ] All helpers have comprehensive tests
- [ ] Documentation with usage examples
- [ ] Can replace 23+ duplicate setupTestDB implementations

---

## Technical Details

### Problem Statement

**Current State:**
- 23+ files define their own `setupTestDB()` function
- Each slightly different (some run migrations, some don't)
- Test user creation duplicated across files
- Test event creation duplicated across files
- ~500 lines of duplicate database setup code

### Implementation

```go
package testutil

import (
    "context"
    "testing"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/db"
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

// SetupTestDB creates an in-memory SQLite database for testing.
// The database is automatically closed when the test completes.
// Migrations are NOT run - use SetupTestDBWithMigrations for that.
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
func CreateTestUser(t *testing.T, database db.Database, role models.Role) *models.User {
    t.Helper()
    
    repo := repositories.NewUserRepository(database)
    user := &models.User{
        Email:     "test-" + time.Now().Format("20060102150405.000000") + "@example.com",
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
func CreateTestEvent(t *testing.T, database db.Database, creatorID int64) int64 {
    t.Helper()
    
    slug := "test-event-" + time.Now().Format("20060102150405.000000")
    startTime := time.Now().Add(7 * 24 * time.Hour)
    endTime := startTime.Add(2 * time.Hour)
    
    result, err := database.Exec(context.Background(), `
        INSERT INTO events (
            title, slug, description, location, 
            start_time, end_time, timezone, status, created_by,
            max_plus_ones, allow_maybe_rsvp
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, "Test Event", slug, "Test Description", "Test Location",
        startTime, endTime, "UTC", "draft", creatorID, 2, true)
    
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
```

---

## Tasks (TDD Approach)

### Task 1: Write Tests First
- [ ] Create `internal/testutil/database_test.go`
- [ ] Write test for SetupTestDB
- [ ] Write test for SetupTestDBWithMigrations
- [ ] Write test for CreateTestUser
- [ ] Write test for CreateTestEvent
- [ ] Write test for CreateTestInvite
- [ ] Run tests: `go test ./internal/testutil` (should FAIL)

### Task 2: Implement Database Helpers
- [ ] Create `internal/testutil/database.go`
- [ ] Implement SetupTestDB with cleanup
- [ ] Implement SetupTestDBWithMigrations
- [ ] Implement CreateTestUser
- [ ] Implement CreateTestEvent
- [ ] Implement CreateTestInvite
- [ ] Run tests: `go test ./internal/testutil` (should PASS)

### Task 3: Documentation
- [ ] Add godoc comments to all functions
- [ ] Update README.md with usage examples
- [ ] Document migration path conventions

---

## Dependencies

**Depends on:** Story 01 (testutil package structure)  
**Blocks:** Story 09 (validate pattern - needs DB helpers)

---

## Usage Examples

**Before (duplicated across 23+ files):**
```go
func setupTestDB(t *testing.T) db.Database {
    database, _ := db.NewDatabase(db.Config{Type: "sqlite", Path: ":memory:"})
    migrator, _ := db.NewMigrator(database.DB(), "../../migrations/sqlite")
    migrator.Up(context.Background())
    return database
}
```

**After (single implementation):**
```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

func TestMyFeature(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    user := testutil.CreateTestUser(t, db, models.RoleAdmin)
    eventID := testutil.CreateTestEvent(t, db, user.ID)
    // ... test code
}
```

---

## Validation

```bash
go test ./internal/testutil -v -run Database
```

---

## Notes

- SetupTestDB uses t.Cleanup() for automatic cleanup
- Unique timestamps prevent test collisions
- Migration path is relative to test file location
- All helpers use t.Helper() for better error reporting
