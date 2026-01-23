# User Story: Install and Configure mockgen

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 30 minutes
**Phase:** 2 - Mock Generation Setup

---

## User Story

As a **developer**, I want **mockgen installed and configured** so that **I can generate mocks for all interfaces automatically**.

---

## Acceptance Criteria

- [ ] mockgen installed via go.mod
- [ ] `tools.go` created for tool dependencies
- [ ] `scripts/generate_mocks.sh` script created
- [ ] Script is executable
- [ ] Documentation on running mock generation
- [ ] Verified mockgen works with test interface

---

## Technical Details

### Tools.go Pattern

```go
//go:build tools
// +build tools

package tools

import (
    _ "go.uber.org/mock/mockgen"
)
```

### Installation

```bash
go get go.uber.org/mock/mockgen@latest
go mod tidy
```

### Mock Generation Script

Create `scripts/generate_mocks.sh`:

```bash
#!/bin/bash
set -e

echo "Generating mocks for TinyRSVP..."

# Repository interfaces
mockgen -source=internal/db/repositories/event_repository.go \
    -destination=internal/testutil/mocks/mock_event_repository.go \
    -package=mocks

# Add more as we implement them in Stories 06-08

echo "Mock generation complete!"
```

---

## Tasks

### Task 1: Install mockgen
- [ ] Run `go get go.uber.org/mock/mockgen@latest`
- [ ] Run `go mod tidy`
- [ ] Verify in go.mod

### Task 2: Create tools.go
- [ ] Create `tools.go` in project root
- [ ] Add build constraint
- [ ] Import mockgen

### Task 3: Create Generation Script
- [ ] Create `scripts/generate_mocks.sh`
- [ ] Add shebang and error handling
- [ ] Make executable: `chmod +x scripts/generate_mocks.sh`
- [ ] Test with simple interface

### Task 4: Documentation
- [ ] Document how to run script
- [ ] Add to README.md
- [ ] Note when to regenerate

---

## Dependencies

**Depends on:** Story 01 (testutil structure)  
**Blocks:** Stories 06, 07, 08 (all mock generation)

---

## Validation

```bash
# Verify mockgen installed
which mockgen
mockgen -version

# Test generation script
./scripts/generate_mocks.sh

# Verify generated mocks compile
go build ./internal/testutil/mocks/...
```

---

## Notes

mockgen is the official Go mock generation tool. It's maintained by Google and has zero runtime dependencies.
