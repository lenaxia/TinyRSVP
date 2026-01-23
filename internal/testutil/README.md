# TinyRSVP Test Utilities

Centralized testing utilities for TinyRSVP.

## Quick Start

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

// Setup test database with migrations
db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")

// Create test data
admin := testutil.CreateTestUser(t, db, models.RoleAdmin)
eventID := testutil.CreateTestEvent(t, db, admin.ID)
inviteID := testutil.CreateTestInvite(t, db, eventID, "unique-token-hash")

// Create pointers for optional fields
email := testutil.StringPtr("test@example.com")
capacity := testutil.IntPtr(100)
allowMaybe := testutil.BoolPtr(true)
deadline := testutil.TimePtr(time.Now().Add(24 * time.Hour))
```

## Contents

- **Pointer Helpers**: `StringPtr()`, `IntPtr()`, etc. (Story 02)
- **Database Helpers**: `SetupTestDB()`, `CreateTestUser()`, etc. (Story 03)
- **Context Helpers**: `CreateAdminContext()`, etc. (Story 04)
- **Generated Mocks**: `mocks/mock_*.go` (Stories 06-08)
- **Test Builders**: `builders/*_builder.go` (Story 18)
- **HTTP Helpers**: Request/response builders (Story 19)
- **Fixtures**: Load test data from JSON (Story 20)

## Package Status

This package is under active development as part of Epic 12: Test Infrastructure Modernization.

See [Epic 12](../../docs/00_BACKLOG/12_EPIC_test_infrastructure.md) for details.

## Utilities Available

### Phase 1: Foundation (In Progress)
- [x] Pointer helpers (Story 02) ✅
- [x] Database helpers (Story 03) ✅
- [ ] Context helpers (Story 04)

### Phase 2: Mock Generation (Planned)
- [ ] Generated mocks for repositories (Story 06)
- [ ] Generated mocks for services (Story 07)
- [ ] Generated mocks for utilities (Story 08)

### Phase 5: Advanced Features (Planned)
- [ ] Test data builders (Story 18)
- [ ] HTTP test helpers (Story 19)
- [ ] Fixture file loaders (Story 20)

## Usage

### Database Helpers

Database helpers make it easy to set up test databases and create test data without duplicating setup code.

**Available Functions:**
- `SetupTestDB(t *testing.T) db.Database` - Creates in-memory SQLite database
- `SetupTestDBWithMigrations(t *testing.T, migrationPath string) db.Database` - Creates DB and runs migrations
- `CreateTestUser(t *testing.T, database db.Database, role models.UserRole) *models.User` - Creates test user
- `CreateTestEvent(t *testing.T, database db.Database, creatorID int64) int64` - Creates test event
- `CreateTestInvite(t *testing.T, database db.Database, eventID int64, tokenHash string) int64` - Creates test invite

**Example:**

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

func TestMyFeature(t *testing.T) {
    // Setup database with migrations
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    // Database automatically closed via t.Cleanup()
    
    // Create test user
    admin := testutil.CreateTestUser(t, db, models.RoleAdmin)
    manager := testutil.CreateTestUser(t, db, models.RoleEventManager)
    
    // Create test event
    eventID := testutil.CreateTestEvent(t, db, admin.ID)
    
    // Create test invite
    inviteID := testutil.CreateTestInvite(t, db, eventID, "unique-token-123")
    
    // Use test data in your test...
}

func TestWithoutMigrations(t *testing.T) {
    // For tests that need custom schema
    db := testutil.SetupTestDB(t)
    
    // Create your own tables
    db.Exec(ctx, "CREATE TABLE custom_test (id INTEGER PRIMARY KEY)")
    
    // Use database...
}
```

**Before (duplicated across 23+ files):**
```go
func setupTestDB(t *testing.T) db.Database {
    database, _ := db.NewDatabase(db.Config{Type: "sqlite", Path: ":memory:"})
    migrator, _ := db.NewMigrator(database.DB(), "../../migrations/sqlite")
    migrator.Up(context.Background())
    return database
}

func createTestUser(t *testing.T, db db.Database) *models.User {
    repo := repositories.NewUserRepository(db)
    user := &models.User{Email: "test@example.com", Name: "Test", Role: models.RoleAdmin}
    repo.Create(context.Background(), user)
    return user
}
```

**After (single implementation):**
```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

func TestMyFeature(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    user := testutil.CreateTestUser(t, db, models.RoleAdmin)
    // ... test code
}
```

**Key Features:**
- Automatic cleanup via `t.Cleanup()` - no manual `defer db.Close()`
- Unique email/timestamp generation prevents test collisions
- Uses `t.Helper()` for better error reporting
- Migration path relative to test file location
- Sensible defaults for all test data

### Pointer Helpers

Pointer helpers make it easy to create pointers to values for optional fields in test data.

**Available Functions:**
- `StringPtr(s string) *string`
- `IntPtr(i int) *int`
- `Int64Ptr(i int64) *int64`
- `BoolPtr(b bool) *bool`
- `TimePtr(t time.Time) *time.Time`
- `Float64Ptr(f float64) *float64`

**Example:**

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

func TestCreateInvite(t *testing.T) {
    invite := &models.Invite{
        EventID:     1,
        Email:       testutil.StringPtr("guest@example.com"),
        Name:        testutil.StringPtr("John Doe"),
        MaxPlusOnes: 2,
        Status:      models.InviteStatusDraft,
    }
    
    // Use invite in test...
}

func TestCreateEvent(t *testing.T) {
    event := &models.Event{
        Title:          "Birthday Party",
        AllowMaybeRSVP: testutil.BoolPtr(true),
        EventCapacity:  testutil.IntPtr(50),
        RSVPDeadline:   testutil.TimePtr(time.Now().Add(7 * 24 * time.Hour)),
    }
    
    // Use event in test...
}
```

**Before (with duplicated helpers):**
```go
// Every test file had this
func stringPtr(s string) *string { return &s }

email := stringPtr("test@example.com")
```

**After (centralized):**
```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

email := testutil.StringPtr("test@example.com")
```

## Development

This package follows TinyRSVP's strict guidelines:
- ✅ Test-Driven Development (tests written first)
- ✅ Type safety (no `map[string]interface{}`)
- ✅ Zero technical debt
- ✅ Idiomatic Go patterns

## Contributing

When adding new utilities:
1. Write tests first
2. Implement utility
3. Update this README
4. Add usage examples
5. Update godoc comments
