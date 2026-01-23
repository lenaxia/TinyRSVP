# User Story: Create testutil Package Structure

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 30 minutes
**Phase:** 1 - Foundation

---

## User Story

As a **developer**, I want **a centralized testutil package with clear structure** so that **all test utilities are organized and discoverable in one location**.

---

## Acceptance Criteria

- [ ] `internal/testutil/` directory created
- [ ] Package documentation (doc.go) created
- [ ] Subdirectories created (mocks/, builders/)
- [ ] README.md created with package overview
- [ ] Package compiles without errors
- [ ] Package imports successfully in test files

---

## Technical Details

### Directory Structure

```
internal/testutil/
├── doc.go                    # Package documentation
├── README.md                 # Usage guide
├── pointers.go               # To be created in Story 02
├── database.go               # To be created in Story 03
├── context.go                # To be created in Story 04
├── http.go                   # To be created in Story 19
├── fixtures.go               # To be created in Story 20
├── mocks/                    # Generated mocks (Story 06-08)
│   └── .gitkeep
└── builders/                 # Test data builders (Story 18)
    └── .gitkeep
```

### Package Documentation (doc.go)

```go
// Package testutil provides centralized testing utilities for TinyRSVP.
//
// This package includes:
//   - Pointer helpers for creating optional field values
//   - Database setup and cleanup helpers
//   - Auth context creation utilities
//   - Generated mocks for all major interfaces
//   - Test data builders for common entities
//   - HTTP request/response test helpers
//   - Fixture file loaders
//
// Usage:
//
//	import "github.com/lenaxia/tinyrsvp/internal/testutil"
//	import "github.com/lenaxia/tinyrsvp/internal/testutil/mocks"
//
// See README.md for detailed examples and patterns.
package testutil
```

### README.md Outline

```markdown
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

This package is under active development as part of Epic 12.

See [Epic 12](../../12_EPIC_test_infrastructure.md) for details.
```

---

## Tasks (TDD Approach)

### Task 1: Create Directory Structure
- [ ] Create `internal/testutil/` directory
- [ ] Create `internal/testutil/mocks/` directory
- [ ] Create `internal/testutil/builders/` directory
- [ ] Add `.gitkeep` files to empty directories

### Task 2: Create Package Documentation
- [ ] Create `internal/testutil/doc.go`
- [ ] Add package comment with usage
- [ ] Verify package compiles: `go build ./internal/testutil`

### Task 3: Create README
- [ ] Create `internal/testutil/README.md`
- [ ] Add package overview
- [ ] Add structure outline
- [ ] Note work-in-progress status

### Task 4: Validation
- [ ] Test import in a sample test file
- [ ] Verify no compilation errors
- [ ] Check godoc renders correctly: `go doc github.com/lenaxia/tinyrsvp/internal/testutil`

---

## Dependencies

**Depends on:** None (first story in epic)  
**Blocks:** All other stories in Phase 1

---

## Validation

```bash
# Verify package structure
ls -la internal/testutil/
ls -la internal/testutil/mocks/
ls -la internal/testutil/builders/

# Verify package compiles
go build ./internal/testutil

# Verify godoc
go doc github.com/lenaxia/tinyrsvp/internal/testutil

# Verify can be imported
# In any test file:
import "github.com/lenaxia/tinyrsvp/internal/testutil"
```

---

## Notes

This story establishes the foundation for all test infrastructure work. Keep it simple - just create the structure and basic documentation. The actual utilities will be added in subsequent stories.
