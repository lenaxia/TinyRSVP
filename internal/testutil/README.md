# TinyRSVP Test Utilities

Centralized testing utilities for TinyRSVP.

## Quick Start

[Examples to be added as utilities are implemented]

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
- [ ] Pointer helpers (Story 02)
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

```go
import "github.com/lenaxia/tinyrsvp/internal/testutil"

// Examples will be added as utilities are implemented
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
