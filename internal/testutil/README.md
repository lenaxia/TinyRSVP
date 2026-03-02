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

- **Pointer Helpers**: `StringPtr()`, `IntPtr()`, `Int64Ptr()`, `BoolPtr()`, `TimePtr()`, `Float64Ptr()` ✅
- **Database Helpers**: `SetupTestDB()`, `SetupTestDBWithMigrations()`, `CreateTestUser()`, `CreateTestEvent()`, `CreateTestInvite()` ✅
- **Context Helpers**: `CreateAdminContext()`, `CreateEventManagerContext()`, `CreateAnonymousContext()`, `CreateTestContext()` ✅
- **Generated Mocks**: 30 mocks across `mocks/repositories/`, `mocks/services/`, `mocks/other/` ✅
- **Test Builders**: `builders/` — fluent test data builders ✅
- **HTTP Helpers**: `http.go` — request/response builders ✅

## Regenerating Mocks

After any interface change, regenerate all mocks:

```bash
./scripts/generate_mocks.sh
```

Commit the generated files alongside your interface changes.

## Usage

### Generated Mocks

TinyRSVP uses [mockgen](https://github.com/uber-go/mock) to generate mocks for all interfaces. Mocks are organized into subdirectories to avoid import cycles.

**Mock Package Structure:**

```
internal/testutil/mocks/
├── repositories/    - Repository mocks (no import cycles)
├── services/        - Service mocks (import their packages)
└── other/           - Database, validators, storage, auth
```

**Regenerating Mocks:**

When interface definitions change, regenerate mocks:

```bash
./scripts/generate_mocks.sh
```

**Using Generated Mocks:**

#### ✅ Handler Tests (Recommended Pattern)

```go
import (
    "testing"
    "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
    "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
    "go.uber.org/mock/gomock"
)

func TestMyHandler(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    // Use both repository and service mocks
    mockEventRepo := repositories.NewMockEventRepository(ctrl)
    mockInviteService := services.NewMockInviteService(ctrl)
    
    // Set expectations
    mockEventRepo.EXPECT().
        GetByID(gomock.Any(), int64(123)).
        Return(&models.Event{ID: 123}, nil)
    
    mockInviteService.EXPECT().
        GetInviteByID(gomock.Any(), int64(1)).
        Return(&models.Invite{ID: 1}, nil)
    
    // Use mocks in handler
    handler := NewMyHandler(mockInviteService, mockEventRepo)
    // ... test handler
}
```

#### ✅ Service Tests (Use Repository Mocks Only)

```go
// internal/events/service_test.go
import (
    "testing"
    "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
    // ✅ Don't import mocks/services - would create import cycle!
    "go.uber.org/mock/gomock"
)

func TestMyService(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    // Use repository mocks (no import cycle)
    mockEventRepo := repositories.NewMockEventRepository(ctrl)
    mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
    
    // Set expectations
    mockEventRepo.EXPECT().
        GetByID(gomock.Any(), int64(123)).
        Return(&models.Event{ID: 123}, nil)
    
    // Use mocks in service
    service := NewMyService(mockEventRepo, mockInviteRepo)
    // ... test service
}
```

**Import Cycle Warning:**

⚠️ **DO NOT** import `mocks/services` from the same package that defines the service interface:

```go
// ❌ WRONG: internal/events/service_test.go
import "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/services"
// This creates an import cycle: events → mocks/services → events

// ✅ CORRECT: Use repository mocks instead
import "github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
```

**Available Mocks (30 total):**

**Repositories (`mocks/repositories/`):**
- `MockEventRepository` — 17 methods
- `MockInviteRepository` — 13 methods
- `MockUserRepository` — 12 methods
- `MockRSVPRepository`
- `MockTemplateRepository`
- `MockAnswerRepository`
- `MockQuestionRepository`
- `MockConfigRepository`
- `MockSessionRepository`
- `MockEmailQueueRepository`

**Services (`mocks/services/`):**
- `MockEventService`
- `MockInviteService` — 16 methods
- `MockRSVPService`
- `MockTemplateService`
- `MockEmailService`
- `MockDashboardService`
- `MockUserService`
- `MockAdminDashboardService`
- `MockUserListService`

**Other (`mocks/other/`):**
- `MockDatabase`
- `MockUserService` (auth interface)
- `MockSessionManager`
- `MockAuthenticator`
- `MockAuthorizationChecker`
- `MockEventValidator`
- `MockTemplateValidator`
- `MockProvider` (storage)
- `MockUserCounter`, `MockEventCounter`, `MockInviteCounter` (admin counters)
- `MockSMTPSender`, `MockRateLimiter`, `MockTemplateRenderer`, `MockEmailMetrics`
- `MockJobsEventService`

**Key Features:**
- Auto-generated from interface definitions
- Type-safe method expectations
- Automatic verification via `ctrl.Finish()`
- Clear error messages on unexpected calls

**Mock Generation Workflow:**
1. Change an interface definition
2. Run `./scripts/generate_mocks.sh`
3. Commit the regenerated files alongside your changes

### Context Helpers

Context helpers provide easy authentication context creation for testing authenticated operations without duplicating context setup code.

**Available Functions:**
- `CreateTestContext(user *models.User) context.Context` - Creates context with given user
- `CreateAdminContext() context.Context` - Creates context with admin user (ID: 1, Role: admin)
- `CreateEventManagerContext() context.Context` - Creates context with event manager user (ID: 2, Role: event_manager)
- `CreateAnonymousContext() context.Context` - Creates context with no user (unauthenticated)

**Example:**

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

func TestAdminOperation(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    service := NewMyService(db)
    
    // Test admin-only operation
    ctx := testutil.CreateAdminContext()
    err := service.DeleteUser(ctx, userID)
    
    if err != nil {
        t.Fatalf("Expected admin to delete user, got error: %v", err)
    }
}

func TestEventManagerOperation(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    service := NewEventService(db)
    
    // Test event manager operation
    ctx := testutil.CreateEventManagerContext()
    event, err := service.CreateEvent(ctx, eventData)
    
    if err != nil {
        t.Fatalf("Expected event manager to create event, got error: %v", err)
    }
}

func TestAnonymousOperation(t *testing.T) {
    service := NewPublicService()
    
    // Test unauthenticated operation
    ctx := testutil.CreateAnonymousContext()
    _, err := service.GetPublicEvents(ctx)
    
    if err != nil {
        t.Fatalf("Expected public operation to work, got error: %v", err)
    }
}

func TestCustomUserOperation(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    
    // Create a custom user for specific test case
    user := &models.User{
        ID:    123,
        Email: "custom@example.com",
        Name:  "Custom User",
        Role:  models.RoleGuest,
    }
    
    ctx := testutil.CreateTestContext(user)
    
    // Test operation with custom user...
}
```

**Before (duplicated context setup):**
```go
func createAdminContext() context.Context {
    user := &models.User{ID: 1, Email: "admin@test.com", Role: models.RoleAdmin}
    return auth.WithUser(context.Background(), user)
}

func TestAdminOperation(t *testing.T) {
    ctx := createAdminContext()
    // ... test code
}
```

**After (centralized context helpers):**
```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

func TestAdminOperation(t *testing.T) {
    ctx := testutil.CreateAdminContext()
    // ... test code
}
```

**Key Features:**
- Consistent test users across all tests (admin has ID: 1, manager has ID: 2)
- Realistic test emails and names
- Uses `auth.WithUser()` internally for proper context attachment
- Supports custom users via `CreateTestContext()`
- Anonymous context for testing unauthenticated operations

**User Details:**
- **Admin**: ID: 1, Email: admin@test.example.com, Name: Test Admin, Role: admin
- **Event Manager**: ID: 2, Email: manager@test.example.com, Name: Test Event Manager, Role: event_manager
- **Anonymous**: No user attached (unauthenticated context)

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
