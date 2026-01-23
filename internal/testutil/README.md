# TinyRSVP Test Utilities

Centralized testing utilities for TinyRSVP.

## Quick Start

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

// Create pointers for optional fields
email := testutil.StringPtr("test@example.com")
capacity := testutil.IntPtr(100)
allowMaybe := testutil.BoolPtr(true)
deadline := testutil.TimePtr(time.Now().Add(24 * time.Hour))

// Use in test data
invite := &models.Invite{
    Email:       testutil.StringPtr("test@example.com"),
    MaxPlusOnes: 2,
}

event := &models.Event{
    Title:          "Test Event",
    AllowMaybeRSVP: testutil.BoolPtr(true),
    EventCapacity:  testutil.IntPtr(50),
}
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
- [ ] Database helpers (Story 03)
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
