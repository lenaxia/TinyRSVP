# Default Template System Implementation

**Date:** 2026-01-09  
**Story:** [06_STORY_03_default_templates.md](../00_BACKLOG/06_STORY_03_default_templates.md)  
**Status:** Complete

---

## Summary

Implemented a complete default template system for TinyRSVP, providing out-of-the-box templates for invite emails, RSVP pages, and confirmation pages. The system uses Go's `embed` package to bundle templates in the binary and provides idempotent seeding functionality.

---

## What Was Implemented

### 1. Default Template Content Files

Created four template files in [`internal/templates/defaults/`](../../internal/templates/defaults/):

- **`invite_email.html`** - Mobile-responsive HTML email with inline CSS
- **`invite_email.txt`** - Plain text email alternative
- **`rsvp_page.html`** - Mobile-first RSVP form with dynamic question support
- **`confirmation_page.html`** - Confirmation page with response summary

**Key Features:**
- Mobile-responsive designs (320px-1024px+)
- Touch-friendly UI elements (44px minimum tap targets)
- Inline CSS for email compatibility
- All template variables utilized
- Conditional rendering for optional fields
- Professional styling with color-coded response statuses

### 2. Template Seeder (`seeder.go`)

Implemented [`TemplateSeeder`](../../internal/templates/seeder.go) with the following capabilities:

**Core Functionality:**
- Embeds default templates using `//go:embed` directive
- Creates system default templates on first run
- Idempotent operation (safe to run multiple times)
- Context-aware with cancellation support
- Proper error handling and reporting

**API:**
```go
seeder := NewSeeder(templateRepo, systemUserID)
err := seeder.SeedDefaults(ctx)
```

**Seeding Logic:**
1. Checks if default template already exists for each type
2. Reads embedded template files
3. Creates Template model with proper metadata
4. Inserts into database via repository
5. Skips if template already exists

### 3. Comprehensive Test Suite

#### Unit Tests (`seeder_test.go`) - 10 tests
- Template creation for all three types
- Idempotent seeding verification
- Template metadata validation (names, flags, version)
- Context cancellation handling
- Repository error handling
- Template content size validation
- Variable usage verification
- Zero createdBy validation

#### Integration Tests (`seeder_integration_test.go`) - 5 tests
- Real database seeding
- Idempotent seeding with database
- Template parseability verification
- Template renderability with test data
- Template validation with validator

**Test Results:** All 15 tests passing with timeout

### 4. Enhanced Validator

Updated [`validator.go`](../../internal/templates/validator.go) to support new template variables:

**RSVP Page Variables Added:**
- `Token` - RSVP token for form submission
- `MaxPlusOnes` - Maximum allowed guests
- Question fields: `ID`, `QuestionText`, `QuestionType`, `Required`, `HelpText`, `Options`, `Value`, `Label`

**Confirmation Page Variables Added:**
- `Token` - For calendar and update links
- `RSVP.Notes` - Guest notes
- Answer fields: `QuestionText`, `AnswerDisplay`

**Test Data Updated:**
- Enhanced `createTestData()` with complete question/answer structures
- Added proper nested field support for range operations

---

## Technical Decisions

### 1. Embed vs External Files

**Decision:** Use `go:embed` to bundle templates in binary

**Rationale:**
- Single binary deployment (no external file dependencies)
- Templates always available
- Version control with code
- No file path configuration needed

### 2. Idempotent Seeding

**Decision:** Check for existing templates before creating

**Rationale:**
- Safe to call on every startup
- Allows manual template updates to persist
- No duplicate template creation
- Graceful handling of partial seeding

### 3. System User ID

**Decision:** Require explicit `createdBy` parameter

**Rationale:**
- Maintains audit trail
- Follows existing patterns
- Allows tracking of system-created vs user-created templates
- Validates against zero value

### 4. Template Variable Expansion

**Decision:** Add nested field variables to allowed list

**Rationale:**
- Templates use `{{range}}` over Questions and Answers
- Nested fields accessed within range context
- Validator needs to allow these for proper validation
- Maintains security while enabling functionality

---

## Integration Points

### Database Schema
Uses existing `templates` table from migration `000007_add_templates.up.sql`:
- All fields properly populated
- Foreign key to users table (created_by)
- Supports NULL event_id for system defaults

### Template Repository
Leverages existing [`TemplateRepository`](../../internal/db/repositories/template_repository.go):
- `GetDefaultByType()` - Check for existing defaults
- `Create()` - Insert new templates
- Full CRUD support for future customization

### Template Engine
Works with existing [`Engine`](../../internal/templates/engine.go):
- All custom functions available
- XSS protection automatic
- Parse and execute validation

### Template Validator
Enhanced existing [`Validator`](../../internal/templates/validator.go):
- Added new allowed variables
- Updated test data structures
- Maintains security guarantees

---

## Usage Example

```go
import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/templates"
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
)

func initializeTemplates(db Database) error {
    repo := repositories.NewTemplateRepository(db)
    seeder := templates.NewSeeder(repo, 1) // 1 = system user ID
    
    ctx := context.Background()
    if err := seeder.SeedDefaults(ctx); err != nil {
        return fmt.Errorf("failed to seed templates: %w", err)
    }
    
    return nil
}
```

---

## Files Created

- `internal/templates/defaults/invite_email.html` - HTML email template
- `internal/templates/defaults/invite_email.txt` - Text email template
- `internal/templates/defaults/rsvp_page.html` - RSVP form template
- `internal/templates/defaults/confirmation_page.html` - Confirmation template
- `internal/templates/seeder.go` - Template seeding implementation
- `internal/templates/seeder_test.go` - Unit tests (10 tests)
- `internal/templates/seeder_integration_test.go` - Integration tests (5 tests)

## Files Modified

- `internal/templates/validator.go` - Added new allowed variables and updated test data
- `internal/templates/validator_test.go` - Updated expected variable lists
- `internal/templates/README.md` - Added seeder documentation
- `docs/00_BACKLOG/06_STORY_03_default_templates.md` - Updated status and checklists

---

## Test Results

```bash
$ go test -timeout 30s ./internal/templates/...
PASS
ok  	github.com/lenaxia/tinyrsvp/internal/templates	0.072s
```

**Total Tests:** 93 tests (78 existing + 15 new)
**Status:** All passing
**Coverage:** Comprehensive unit and integration coverage

---

## Next Steps

### Immediate
1. Call seeder from application startup in `cmd/server/main.go`
2. Create system user (ID=1) if not exists during bootstrap
3. Add seeder call after database initialization

### Future Enhancements
1. Template customization UI (Story 04)
2. Template preview functionality
3. Template versioning and rollback
4. Template export/import

---

## Notes

- Templates are mobile-responsive with breakpoints at 480px and 1024px
- Email templates use inline CSS for maximum compatibility
- All templates use semantic HTML5
- XSS protection verified through validator integration tests
- Templates demonstrate best practices for variable usage
- Seeder is idempotent and safe for production use

---

## References

- **Story:** [06_STORY_03_default_templates.md](../00_BACKLOG/06_STORY_03_default_templates.md)
- **Epic:** [06_EPIC_templates.md](../00_BACKLOG/06_EPIC_templates.md)
- **Templates Package:** [internal/templates/](../../internal/templates/)
- **Default Templates:** [internal/templates/defaults/](../../internal/templates/defaults/)
