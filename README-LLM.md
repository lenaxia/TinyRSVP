# TinyRSVP - LLM Implementation Guide

**Version:** 2.0  
**Last Updated:** 2026-02-04  
**Project Status:** Active Development (75% Complete)

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Critical Guidelines & Hard Rules](#critical-guidelines--hard-rules)
3. [Repository Structure](#repository-structure)
4. [Architecture Overview](#architecture-overview)
5. [Technology Stack](#technology-stack)
6. [Development Workflow](#development-workflow)
7. [Common Commands](#common-commands)
8. [Branch Management](#branch-management)
9. [Documentation Standards](#documentation-standards)
10. [Testing Requirements](#testing-requirements)

---

## Project Overview

TinyRSVP is a **self-hosted, small-scale RSVP and invitation platform**, designed for homelab environments, family events, clubs, and private gatherings. This project is **100% LLM-implemented** with significant human-in-the-loop oversight.

**Core Principles:**
- Guests never required to create accounts
- Token-based guest access
- Self-hosted first
- Single-node homelab compatible
- Feature completeness over feature breadth
- Docker first experience, this application should only be available in docker

**Primary Source Documents:**
- [`docs/02_DESIGN/02_REVISED_HLD.md`](docs/02_DESIGN/02_REVISED_HLD.md) - ⭐ AUTHORITATIVE specification
- [`docs/02_DESIGN/lld/`](docs/02_DESIGN/lld/) - Low-level designs (8 modules)
- [`docs/00_BACKLOG/`](docs/00_BACKLOG/) - Sprint stories and epics (150+ stories in 13 epics)
- [`docs/01_WORKLOG/`](docs/01_WORKLOG/) - Progress updates (141 entries: 0000-0140)
- [`docs/04_SUMMARIES/PROJECT_STATUS_ASSESSMENT.md`](docs/04_SUMMARIES/PROJECT_STATUS_ASSESSMENT.md) - Latest status

---

## Critical Guidelines & Hard Rules

### 0. Test Driven Development (TDD)

**MANDATORY:** Write tests BEFORE writing functional code. ALWAYS.

```go
// ✅ Correct workflow:
// 1. Write test
// 2. Run test (should fail)
// 3. Write minimal code to pass
// 4. Run test (should pass)
// 5. Refactor if needed
```

**Test Requirements:**
- Multiple happy path tests
- Multiple unhappy path tests
- Edge case coverage
- ALWAYS use timeouts when running tests to detect hangs
- Tests must pass before marking work complete

### 1. Type Safety First

**ALWAYS DO ✅:**
- Define strongly-typed structs for ALL data structures
- Create domain types for related fields
- Export structs from packages for reuse

**NEVER DO ❌:**
- NEVER use `map[string]interface{}` for structured data
- NEVER use `interface{}` when you know the type
- NEVER use type assertions when you can use generics
- NEVER pass untyped data between functions

**When Maps Are Acceptable:**

ONLY use `map[string]interface{}` when:
1. Parsing external JSON/YAML with unknown structure (convert to struct immediately)
2. Interacting with reflection-based libraries (convert to struct ASAP)
3. Truly dynamic configuration where structure is unknowable

**Example:**
```go
// ✅ Parse to map, convert to struct immediately
func parseConfig(data []byte) (*Config, error) {
    var raw map[string]interface{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, err
    }

    config := &Config{
        Field1: raw["field1"].(string),
        Field2: raw["field2"].(int),
    }

    return config, nil  // Return typed struct, not map
}
```

**Justification:**
1. Compile-time verification - Catch errors before runtime
2. IDE autocomplete - Developer productivity
3. Refactoring safety - Rename propagates everywhere
4. Self-documenting - Struct shows what's available
5. Performance - No reflection, no runtime type checking

### 2. Idiomatic Go

- Follow Go conventions, not Perl patterns
- Use Go's multiple return values (value, error) pattern
- Avoid global state and exceptions
- Create custom error types for domain-specific errors
- Prefer minimal or no concurrency when possible

### 3. Explicit Over Implicit

- Go favors explicit error handling
- Explicit type declarations
- No magic or hidden behavior

### 4. Concurrency Guidelines

- Design for goroutines/channels from the start
- Prefer minimal or no concurrency when possible
- Only add concurrency when there's clear benefit
- Always consider synchronization and race conditions



### 6. Communication Tone

**MANDATORY TONE RULES:**

- Always be neutral, factual, objective
- Do NOT be sensational, overly agreeable, or a sycophant
- Don't be a cheerleader, be a critical collaborator
- Never agree with something just because the user stated it
- Always validate statements and ensure you have meaningful and solid proof before making claims
- Provide honest and objective feedback
- If you agree, do so based on evidence or sound reasoning, not to please

### 7. Code Quality

- No comments unless its necessary and makes sense
- If you are going to leave comments make sure they are timeless in that they will not be out dated
- Comments get outdated and mislead LLMs 
- Code should be self-documenting through clear naming
- If you do see comments that are incorrect, either remove them or update them

### 8. Technical Debt

**ZERO TOLERANCE:**
- Do NOT create adapters for backwards compatibility
- Always remove legacy code
- Implement full final implementation
- Do NOT accrue technical debt
- Do NOT use weird hacks to get tests to pass
- Do the work needed to fix code or tests properly

### 9. Uncertainty Protocol

**If uncertain about proper behavior: ASK THE USER**

Do not guess, assume, or implement workarounds.

### 10. Tools are production code

If you create a script or tool, you MUST create it using TDD and
ensure proper and high test coverage. Never assume a script works
unless you have demonstrable proof with passing tests. If a script
or tool lacks tests, assume it is broken until proven otherwise.

### 11. Make sure you understand the entire architecture

Make sure you understand the entire architecture and how the changes
you are about to make fit in within that architecture. Understand the 
goals of what you are trying to achieve and how to go about it. 

ALWAYS review the HLD and relevant LLD(s).
- docs/02_DESIGN/02_REVISED_HLD.md (AUTHORITATIVE)
- docs/02_DESIGN/lld/ (Low-level designs)

### 12. Status Documentation Requirements

**MANDATORY:** When marking stories/epics complete:
- Run all tests
- Document test pass rate
- Document known issues
- Document confidence level
- Document production readiness

**Status Levels:**
- ✅ Complete - All tests pass, production ready
- ⚠️ Complete (with issues) - Tests mostly pass, known issues documented
- ⚠️ BROKEN - Tests failing, functionality broken
- ❌ Not Started - No implementation

**Example:**
```markdown
**Status:** ⚠️ Complete (with issues)
**Test Pass Rate:** 90% (3 failures)
**Confidence:** MEDIUM (75%)
**Production Ready:** Mostly (requires fixes)
**Known Issues:**
- Security headers test failing
- Template integration broken
```

---

## Repository Structure

```
TinyRSVP/
├── README.md                    # User-facing project README
├── README-LLM.md               # This file - LLM implementation guide
├── go.mod                      # Go module definition
├── go.sum                      # Go dependency checksums
│
├── cmd/                        # Application entrypoints
│   └── server/                 # Main server application
│       └── main.go
│
├── internal/                   # Private application code
│   ├── auth/                   # Authentication & authorization
│   ├── config/                 # Configuration management
│   ├── db/                     # Database layer
│   ├── email/                  # Email sending & queue
│   ├── events/                 # Event management
│   ├── handlers/               # HTTP handlers
│   ├── invites/                # Invite management
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Domain models (structs)
│   ├── rsvp/                   # RSVP logic
│   ├── storage/                # Asset storage abstraction
│   └── templates/              # Template rendering
│
├── pkg/                        # Public/reusable packages
│   ├── ics/                    # ICS calendar generation
│   └── token/                  # Token generation/validation
│
├── migrations/                 # Database migrations
│   ├── sqlite/
│   └── postgres/
│
├── templates/                  # HTML templates
│   ├── web/                    # Web page templates
│   └── email/                  # Email templates
│
├── static/                     # Static assets
│   ├── css/
│   ├── js/
│   └── images/
│
├── docs/                       # Documentation
│   ├── README.md               # Documentation index
│   ├── 00_BACKLOG/             # Sprint stories and epics (150+ stories)
│   │   ├── README.md
│   │   ├── 00_FOUNDATION/     # Epic folders with user stories
│   │   ├── 01_AUTH/
│   │   └── ...
│   ├── 01_WORKLOG/             # Progress updates (0000-0140)
│   │   └── README.md
│   ├── 02_DESIGN/              # Active design documents
│   │   ├── 02_REVISED_HLD.md  # ⭐ AUTHORITATIVE specification
│   │   ├── 04_LLD_INDEX.md
│   │   └── lld/               # Low-level designs (8 files)
│   ├── 03_REFERENCE/           # Reference documentation
│   ├── 04_SUMMARIES/           # Project status
│   └── 99_ARCHIVE/             # Historical documents
│
├── llm-workflows/              # LLM workflow templates
│   └── README.md               # Workflow index
│
└── tests/                      # Integration tests
    └── e2e/
```

**Key Principles:**
- Every major folder has a README.md
- READMEs are the first thing to read when entering a folder
- READMEs are short but lay out rules for reading/editing

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Reverse Proxy Layer                      │
│              (Traefik/Nginx + Authelia/Authentik)           │
│                   - TLS Termination                          │
│                   - Forward Auth (Optional)                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      TinyRSVP Application                    │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Auth       │  │   Events     │  │   Invites    │     │
│  │  Middleware  │  │   Handler    │  │   Handler    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│         │                  │                  │             │
│         └──────────────────┴──────────────────┘             │
│                            │                                │
│                            ▼                                │
│                   ┌─────────────────┐                       │
│                   │  Database Layer │                       │
│                   │  (SQLite/PG)    │                       │
│                   └─────────────────┘                       │
│                            │                                │
│         ┌──────────────────┼──────────────────┐            │
│         ▼                  ▼                  ▼            │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐        │
│  │  Email   │      │ Storage  │      │   ICS    │        │
│  │  Queue   │      │ Provider │      │Generator │        │
│  └──────────┘      └──────────┘      └──────────┘        │
└─────────────────────────────────────────────────────────────┘
         │                  │
         ▼                  ▼
  ┌──────────┐      ┌──────────────┐
  │   SMTP   │      │ Local FS/S3  │
  │  Server  │      │   Storage    │
  └──────────┘      └──────────────┘
```

**Key Components:**

1. **Auth Layer**: OIDC or Forward Auth for admins/managers
2. **Event Management**: CRUD operations for events
3. **Invite System**: Token-based guest access
4. **RSVP Handler**: Guest response processing
5. **Email Queue**: Async email sending with retries
6. **Storage Provider**: Pluggable (Local FS or S3-compatible)
7. **Template Engine**: Go `html/template` for web and email

**Data Flow:**

1. Admin creates event → Database
2. Admin creates invites → Generate tokens → Database
3. Admin sends invites → Email queue → SMTP
4. Guest clicks link → Token validation → RSVP page
5. Guest submits RSVP → Database → Confirmation email

---

## Technology Stack

### Implementation Language: Go

**Justification:**
- Single static binary
- Excellent standard library (SMTP, HTML templates)
- Strong OIDC/OAuth2 libraries
- SQLite and Postgres support
- Low memory footprint
- Easy cross-compilation
- Homelab-friendly

### Template Engine: Go `html/template`

**Why:**
- Automatic HTML escaping (XSS-safe)
- No build step required
- Built into Go standard library
- Works for both web pages and emails

### Frontend: Plain CSS (Mobile-First) + Vanilla JavaScript

**Why:**
- **Minimal footprint**: ~10-25KB total (minified)
- **No build step**: Serve static files directly
- **Fast loading**: Critical for user experience
- **Mobile-first**: Progressive enhancement for desktop
- **No framework overhead**: Pure browser APIs
- **Timeless**: No framework churn or deprecation

**CSS Approach:**
- Mobile-first responsive design
- CSS Grid + Flexbox for layouts
- Custom properties (CSS variables) for theming
- Media queries for tablet/desktop enhancements
- Critical CSS inlined in `<head>` for fast first paint

**JavaScript Approach:**
- Vanilla ES6+ (no transpilation needed)
- Progressive enhancement (works without JS)
- Module pattern for organization
- Event delegation for performance
- Minimal DOM manipulation

**Performance Targets:**
- First Contentful Paint: <1s
- Time to Interactive: <2s
- Total page weight: <100KB (including images)

**Mobile Experience:**
- Touch-friendly tap targets (44px minimum)
- Single-column layouts
- Stacked forms
- Hamburger navigation
- Full-width buttons
- Optimized for 320px-767px screens

**Desktop Experience:**
- Multi-column layouts (CSS Grid)
- Wider content areas (max 1200px centered)
- Inline form fields where appropriate
- Full navigation bars
- Hover states and interactions
- Optimized for 1024px+ screens

**Responsive Breakpoints:**
```css
/* Mobile: 320px-767px (base styles) */
/* Tablet: 768px-1023px */
@media (min-width: 768px) { ... }
/* Desktop: 1024px+ */
@media (min-width: 1024px) { ... }
```

**Optional Enhancement:**
- HTMX (14KB) for dynamic updates without full page reloads
- Only add if specific use case requires it (defer to v1+)

### OIDC Library: `github.com/coreos/go-oidc`

**Why:**
- De facto standard OIDC library in Go
- OIDC Discovery support
- ID Token validation
- JWKS rotation
- PKCE support
- Standards-compliant

### OAuth2 Helper: `golang.org/x/oauth2`

**Why:**
- Canonical Go OAuth2 library
- Authorization code flow
- Token exchange
- Refresh handling

### Database: SQLite (default) / PostgreSQL (optional)

**Why:**
- SQLite: Zero-config, single-file, perfect for homelab
- PostgreSQL: Optional for larger deployments

### Storage: Local Filesystem (default) / S3-compatible (optional)

**Why:**
- Local FS: Simple, no dependencies
- S3: Scalable, works with MinIO/Ceph

---

## Development Workflow

### 1. Before Starting Work

**ALWAYS:**
1. Read the relevant folder's README.md
2. Check [`docs/00_BACKLOG/`](docs/00_BACKLOG/) for current sprint stories (150+ stories in 13 epics)
3. Check [`docs/01_WORKLOG/`](docs/01_WORKLOG/) for recent progress (entries 0000-0140)
4. Review [`docs/02_DESIGN/02_REVISED_HLD.md`](docs/02_DESIGN/02_REVISED_HLD.md) for authoritative spec
5. Check [`docs/04_SUMMARIES/PROJECT_STATUS_ASSESSMENT.md`](docs/04_SUMMARIES/PROJECT_STATUS_ASSESSMENT.md) for latest status

### 2. During Work

**ALWAYS:**
1. Write tests FIRST (TDD)
2. Use strongly-typed structs (NO `map[string]interface{}`)
3. Update relevant README.md files as you work
4. Update backlog story checklists `[ ]` → `[x]`
5. Commit regularly (every logical unit of work)
6. Update architecture diagram in this file if structure changes

**Pre-Commit Hooks:**
- Pre-commit hooks automatically run before each commit
- Hooks enforce: go fmt, go vet, tests with -timeout 30s
- Hooks warn about: debug prints, TODOs, map[string]interface{} usage
- See `.git/hooks/README.md` for details on bypassing hooks
- Emergency bypass: `git commit --no-verify` (use sparingly)

### 3. After Completing Work

**ALWAYS:**
1. Run all tests with timeout
2. Verify tests pass
3. Update backlog story status
4. Create handoff document in [`docs/01_WORKLOG/`](docs/01_WORKLOG/)
5. Update this README-LLM.md if needed
6. Commit all changes with descriptive message

### 4. Major Features

**ALWAYS:**
1. Create feature branch (document in [Branch Management](#branch-management))
2. Update branch list in this README
3. Work in branch
4. Merge to main when complete and tested
5. Update branch list (mark as merged)

---

## Common Commands

### Setup

```bash
# Initialize Go module (if not done)
go mod init github.com/lenaxia/tinyrsvp

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Development

```bash
# Run application
go run cmd/server/main.go

# Build binary
go build -o bin/tinyrsvp cmd/server/main.go

# Run with race detector
go run -race cmd/server/main.go
```

### Testing

```bash
# Run all tests with timeout
go test -timeout 30s ./...

# Run tests with coverage
go test -timeout 30s -cover ./...

# Run tests with verbose output
go test -timeout 30s -v ./...

# Run specific package tests
go test -timeout 30s ./internal/auth/...

# Run tests with race detector
go test -timeout 30s -race ./...

# Generate coverage report
go test -timeout 30s -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (install: go install golang.org/x/lint/golint@latest)
golint ./...

# Run static analysis
go vet ./...

# Run all quality checks
go fmt ./... && go vet ./... && golint ./...
```

### Database Migrations

```bash
# Apply migrations (example - actual command TBD)
go run cmd/migrate/main.go up

# Rollback migration
go run cmd/migrate/main.go down

# Create new migration
go run cmd/migrate/main.go create migration_name
```

### Docker

```bash
# Build Docker image
docker build -t tinyrsvp:latest .

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e DATABASE_PATH=/data/tinyrsvp.db \
  tinyrsvp:latest

# Run with docker-compose
docker-compose up -d
```

---

## Branch Management

**Active Branches:**

| Branch Name | Purpose | Status | Created | Owner |
|-------------|---------|--------|---------|-------|
| `main` | Stable code | Active | 2026-01-06 | - |

**Merged Branches:**

| Branch Name | Purpose | Merged Date | PR/Commit |
|-------------|---------|-------------|-----------|
| _(none yet)_ | - | - | - |

**Branch Naming Convention:**
- Feature: `feature/short-description`
- Bugfix: `bugfix/issue-description`
- Hotfix: `hotfix/critical-issue`
- Docs: `docs/what-changed`

**Branch Workflow:**
1. Create branch from `main`
2. Document in table above
3. Work in branch
4. Regular commits
5. Merge to `main` when complete
6. Update table (move to "Merged Branches")
7. Delete branch

---

## Documentation Standards

### Design Documents

**Location:** [`docs/02_DESIGN/`](docs/02_DESIGN/)

**Naming:** `XX_DOCUMENT_NAME.md` where XX is creation order (00, 01, 02, etc.)

**Current Active Docs:**
- `02_REVISED_HLD.md` - ⭐ AUTHORITATIVE high-level design
- `04_LLD_INDEX.md` - Index of low-level designs
- `lld/` - Detailed low-level designs (8 modules)

**Purpose:**
- High-level designs
- Architecture decisions
- Technical specifications
- API designs

**When to Create:**
- Before implementing major features
- When making architectural decisions
- When defining new subsystems

**Historical/Superseded Docs:** Moved to [`docs/99_ARCHIVE/design/`](docs/99_ARCHIVE/design/)

### Worklog Documents

**Location:** [`docs/01_WORKLOG/`](docs/01_WORKLOG/)

**Naming:** `NNNN_YYYY-MM-DD_description.md` (continuous numbering from 0000)

**Current Count:** 141 entries (0000-0140)

**Purpose:**
- Progress updates
- Handoff documents
- Session summaries
- Blockers and decisions

**When to Create:**
- After completing significant work
- Before context switch
- When handing off to another session
- When documenting blockers

**Next Entry:** Use `0141_YYYY-MM-DD_description.md`

### Backlog Stories

**Location:** [`docs/00_BACKLOG/`](docs/00_BACKLOG/)

**Structure:**
- **13 Epic Folders:** Each epic has its own directory (00_FOUNDATION through 12_TEST_INFRASTRUCTURE)
- **Epic READMEs:** Each folder contains README.md with epic overview
- **User Stories:** Individual story files within each epic folder
- **150+ Stories:** Organized across all epics

**Epic Structure:**
```
00_BACKLOG/
├── 00_FOUNDATION/
│   ├── README.md                           # Epic overview
│   ├── 00_STORY_01_go_module_setup.md
│   ├── 00_STORY_02_config_management.md
│   └── ...
├── 01_AUTH/
│   ├── README.md
│   └── (all 01_STORY_* files)
└── ...
```

**Epic 10 - Technical Debt & Improvements:**
- Epic 10 is reserved for issues, improvements, and technical debt that don't fit into other epics
- When validation identifies gaps or improvements needed in existing functionality, add them to Epic 10
- Examples: Return URL preservation in auth, performance optimizations, refactoring opportunities
- Format: `10_STORY_XX_description.md` following the same user story template

**Format:**
```markdown
# Epic: Feature Name

## User Story: As a [role], I want [goal] so that [benefit]

### Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

### Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

### Status
- Status: Not Started | In Progress | Complete
- Assigned: LLM Session ID or Human
- Started: YYYY-MM-DD
- Completed: YYYY-MM-DD
```

**When to Add to Epic 10:**
- Validation identifies gaps in completed stories that aren't critical blockers
- Refactoring opportunities discovered during implementation
- Performance improvements identified
- Technical debt that should be addressed but doesn't block current work
- Cross-cutting concerns that span multiple epics

### Reference Documentation

**Location:** [`docs/03_REFERENCE/`](docs/03_REFERENCE/)

**Active References:**
- `PERMISSION_REFERENCE.md` - RBAC permission system
- `THEME_DESIGN_SYSTEM.md` - Theme and styling guidelines
- `TEMPLATE_EDITOR_API.md` - Template API reference
- `XSS_PREVENTION.md` - Security guidelines

**Purpose:** Technical references for active features

### Status & Summaries

**Location:** [`docs/04_SUMMARIES/`](docs/04_SUMMARIES/)

**Current Status:** `PROJECT_STATUS_ASSESSMENT.md` (Updated: 2026-02-03)

**Purpose:** Track overall project status, epic completion, and blockers

### Archive

**Location:** [`docs/99_ARCHIVE/`](docs/99_ARCHIVE/)

**Contents:** Historical documents no longer actively used
- `design/` - Superseded design documents
- `reference/` - Completed setup guides and plans
- `summaries/` - Old phase summaries

**Policy:** Documents archived when superseded, completed, or no longer relevant

### README Files

**Every major folder MUST have a README.md:**

**Template:**
```markdown
# Folder Name

## Purpose
Brief description of what this folder contains.

## Rules
- Rule 1 for reading/editing files here
- Rule 2 for reading/editing files here

## Structure
- `subfolder/` - Description
- `file.go` - Description

## Key Files
- `important.go` - Why it's important
```

**Update Frequency:**
- When folder structure changes
- When new files are added
- When rules change

---

## Testing Requirements

### Test-Driven Development (TDD)

**MANDATORY WORKFLOW:**

1. **Write Test First**
   ```go
   func TestUserCreation(t *testing.T) {
       // Arrange
       user := &User{Email: "test@example.com"}
       
       // Act
       err := CreateUser(user)
       
       // Assert
       if err != nil {
           t.Errorf("Expected no error, got %v", err)
       }
   }
   ```

2. **Run Test (Should Fail)**
   ```bash
   go test -timeout 30s ./internal/auth/
   ```

3. **Write Minimal Code**
   ```go
   func CreateUser(user *User) error {
       // Minimal implementation
       return nil
   }
   ```

4. **Run Test (Should Pass)**
   ```bash
   go test -timeout 30s ./internal/auth/
   ```

5. **Refactor If Needed**

### Test Coverage Requirements

**MANDATORY:**
- Multiple happy path tests
- Multiple unhappy path tests
- Edge case coverage
- Error condition tests

**Example:**
```go
func TestInviteTokenGeneration(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "test@example.com", false},
        {"empty email", "", true},
        {"invalid email", "notanemail", true},
        {"very long email", strings.Repeat("a", 300) + "@example.com", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            token, err := GenerateInviteToken(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateInviteToken() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && token == "" {
                t.Error("Expected non-empty token")
            }
        })
    }
}
```

### Test Organization

```
internal/
├── auth/
│   ├── auth.go
│   └── auth_test.go
├── events/
│   ├── events.go
│   └── events_test.go
```

**Rules:**
- Test files in same package as code
- Use `_test.go` suffix
- Use table-driven tests for multiple cases
- Use subtests with `t.Run()`

### Running Tests

**ALWAYS use timeout:**
```bash
# Good ✅
go test -timeout 30s ./...

# Bad ❌ (can hang forever)
go test ./...
```

**Before marking work complete:**
```bash
# Run all tests with race detector and timeout
go test -timeout 30s -race ./...

# Check coverage
go test -timeout 30s -cover ./...
```

---

## LLM Workflow Integration

### Using LLM Workflows

**Location:** [`llm-workflows/`](llm-workflows/)

**Purpose:** Standardized prompts and sequences for common tasks

**When to Use:**
- Writing design documents
- Creating test suites
- Updating documentation
- Performing code reviews
- Refactoring code

**How to Use:**
1. Read [`llm-workflows/README.md`](llm-workflows/README.md)
2. Select appropriate workflow
3. Follow workflow steps
4. Update workflow if improvements found

### Regular Maintenance Tasks

**After Every Work Session:**
1. Update backlog story checklists
2. Update relevant README.md files
3. Commit changes with descriptive message
4. Create worklog entry if significant progress

**Weekly (or every ~5 sessions):**
1. Review and update this README-LLM.md
2. Review and update architecture diagram
3. Review and update backlog priorities
4. Clean up completed branches

---

## Quick Reference

### Type Safety Checklist

- [ ] Using structs instead of `map[string]interface{}`?
- [ ] All types explicitly declared?
- [ ] Using generics instead of `interface{}`?
- [ ] Exported structs for reuse?

### TDD Checklist

- [ ] Tests written BEFORE code?
- [ ] Tests run with timeout?
- [ ] Multiple happy paths tested?
- [ ] Multiple unhappy paths tested?
- [ ] Edge cases covered?
- [ ] All tests passing?

### Error Handling Checklist

- [ ] Using `HandleError(w, r, err)` for all error responses?
- [ ] NOT using legacy helper functions that bypass HandleError?
- [ ] Error types properly defined (NotFoundError, ValidationError, etc.)?
- [ ] Tests set `Accept: application/json` header when expecting JSON?
- [ ] Content negotiation tested (JSON vs HTML)?
- [ ] Request ID propagation verified?

### Documentation Checklist

- [ ] README.md updated in affected folders?
- [ ] Backlog story checklist updated?
- [ ] Worklog entry created (if significant)?
- [ ] Architecture diagram updated (if changed)?
- [ ] Branch list updated (if applicable)?

### Commit Checklist

- [ ] All tests passing?
- [ ] Code formatted (`go fmt`)?
- [ ] No linter errors (`go vet`)?
- [ ] Documentation updated?
- [ ] Commit message descriptive?

---

## Getting Help

### When Uncertain

**ASK THE USER** - Do not guess or assume.

### Common Questions

**Q: Should I use a map here?**
A: No, unless parsing external data. Define a struct.

**Q: Should I add concurrency?**
A: Only if there's clear benefit. Prefer simplicity.

**Q: Should I add a comment?**
A: No, unless ABSOLUTELY necessary. Make code self-documenting.

**Q: Should I maintain backwards compatibility?**
A: No, implement the full final solution. No technical debt.

**Q: Tests are failing, should I hack around it?**
A: No, fix the code or tests properly. Ask user if uncertain.

**Q: Should I create a wrapper function for error handling?**
A: No, use `HandleError(w, r, err)` directly. Wrapper functions that bypass centralized error handling create technical debt and inconsistency.

**Q: How do I handle errors in HTTP handlers?**
A: Always use `HandleError(w, r, err)` which provides content negotiation, request ID logging, and proper error type mapping. Never create custom error response functions.

**Q: My tests are getting HTML instead of JSON responses?**
A: Set the `Accept: application/json` header in your test requests. Content negotiation is working correctly.

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-06 | Initial creation |
| 2.0 | 2026-02-04 | Major documentation reorganization: Added 02_DESIGN/, 03_REFERENCE/, 04_SUMMARIES/, 99_ARCHIVE/ folders. Updated all references to reflect new structure. Changed worklog naming to continuous numbering (0000-NNNN). Organized backlog into epic folders. |

---

**Remember:** This is a living document. Update it as the project evolves.
