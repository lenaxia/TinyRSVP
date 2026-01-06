# User Story: Go Module Setup

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 2 hours
**Actual Effort:** 1 hour

---

## User Story

As a **developer**, I want **a properly initialized Go module with project structure** so that **I can start implementing features with a solid foundation**.

---

## Acceptance Criteria

- [x] Go module initialized with `go mod init`
- [x] Module name follows convention: `github.com/yourusername/tinyrsvp`
- [x] All required dependencies added to `go.mod`
- [x] Directory structure created per README-LLM.md specification
- [ ] All directories have README.md files (deferred - created as needed)
- [x] `.gitignore` configured for Go projects
- [x] `go mod tidy` runs without errors
- [x] `go build ./...` compiles successfully

---

## Technical Details

### Directory Structure to Create

```
TinyRSVP/
├── cmd/
│   └── server/
│       ├── main.go
│       └── README.md
├── internal/
│   ├── auth/
│   │   └── README.md
│   ├── config/
│   │   └── README.md
│   ├── db/
│   │   ├── repositories/
│   │   │   └── README.md
│   │   └── README.md
│   ├── email/
│   │   └── README.md
│   ├── events/
│   │   └── README.md
│   ├── handlers/
│   │   └── README.md
│   ├── invites/
│   │   └── README.md
│   ├── middleware/
│   │   └── README.md
│   ├── models/
│   │   └── README.md
│   ├── rsvp/
│   │   └── README.md
│   ├── storage/
│   │   └── README.md
│   └── templates/
│       └── README.md
├── pkg/
│   ├── ics/
│   │   └── README.md
│   └── token/
│       └── README.md
├── migrations/
│   ├── sqlite/
│   │   └── README.md
│   └── postgres/
│       └── README.md
├── templates/
│   ├── web/
│   │   └── README.md
│   └── email/
│       └── README.md
├── static/
│   ├── css/
│   ├── js/
│   └── images/
└── tests/
    └── e2e/
        └── README.md
```

### Required Dependencies

```go
require (
    github.com/mattn/go-sqlite3 v1.14.18
    github.com/golang-migrate/migrate/v4 v4.17.0
    github.com/coreos/go-oidc/v3 v3.9.0
    golang.org/x/oauth2 v0.15.0
    github.com/go-chi/chi/v5 v5.0.11
)
```

**Note:** Using chi router for HTTP routing - lightweight, idiomatic, stdlib-compatible.

### Initial main.go

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("TinyRSVP Server")
    os.Exit(0)
}
```

### .gitignore Contents

```
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# Database files
*.db
*.db-shm
*.db-wal

# Environment files
.env
.env.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Data directory
data/

# Temporary files
tmp/
temp/
```

---

## Tasks

### Phase 1: Module Initialization
- [ ] Run `go mod init github.com/yourusername/tinyrsvp`
- [ ] Create `.gitignore` file with Go-specific ignores
- [ ] Verify module initialization with `go mod tidy`

### Phase 2: Directory Structure
- [ ] Create all directories per structure above
- [ ] Create placeholder README.md in each major directory
- [ ] Create initial `cmd/server/main.go` with basic structure

### Phase 3: Dependency Setup
- [ ] Add SQLite driver: `go get github.com/mattn/go-sqlite3`
- [ ] Add migration tool: `go get github.com/golang-migrate/migrate/v4`
- [ ] Add OIDC library: `go get github.com/coreos/go-oidc/v3`
- [ ] Add OAuth2 library: `go get golang.org/x/oauth2`
- [ ] Run `go mod tidy` to clean up dependencies

### Phase 4: Verification
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `go fmt ./...` to format code
- [ ] Run `go vet ./...` to check for issues
- [ ] Commit initial structure

---

## Testing Requirements

### Verification Tests

```bash
# Test 1: Module initialization
go mod verify

# Test 2: Build succeeds
go build -o bin/tinyrsvp cmd/server/main.go

# Test 3: Binary runs
./bin/tinyrsvp

# Test 4: All directories exist
test -d cmd/server && \
test -d internal/config && \
test -d internal/db && \
test -d pkg/token && \
echo "Directory structure verified"
```

---

## Dependencies

**Depends on:** None (first story)  
**Blocks:** All other stories in Epic 00

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All verification tests pass
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Changes committed to git
- [ ] README.md files created in all directories

---

## Notes

- Use Go 1.21 or later for best compatibility
- Module name should match your actual GitHub repository
- All README.md files should follow the template in README-LLM.md
- Ensure CGO is enabled for SQLite support (CGO_ENABLED=1)

---

## References

- **README-LLM.md:** Repository Structure section
- **HLD:** Section 20 (Deployment)
- **LLD:** [lld/07_DATABASE_LLD.md](../lld/07_DATABASE_LLD.md) - Package structure
