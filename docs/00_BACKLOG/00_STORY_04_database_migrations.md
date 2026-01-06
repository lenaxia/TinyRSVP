# User Story: Database Migrations

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 5 hours
**Actual Effort:** 2 hours
**Completed:** 2026-01-06

---

## User Story

As a **developer**, I want **automated database migrations** so that **the database schema can be versioned and deployed reliably**.

---

## Acceptance Criteria

- [x] Migration system using golang-migrate integrated
- [x] All 11 database tables created via migrations
- [x] Up and down migrations implemented
- [x] Migrations run automatically on startup
- [x] Migration version tracked in database
- [x] Rollback capability functional
- [x] All indexes and constraints defined
- [x] All tests pass with timeout

---

## Technical Details

### Migration Tool

**Library:** `github.com/golang-migrate/migrate/v4`

### Migrator Interface

```go
package db

import "context"

type Migrator interface {
    Up(ctx context.Context) error
    Down(ctx context.Context) error
    Version(ctx context.Context) (uint, bool, error)
    Steps(ctx context.Context, n int) error
}
```

### Tables to Create (11 total)

1. users
2. sessions  
3. events
4. invites
5. rsvps
6. preference_questions
7. rsvp_answers
8. email_queue
9. templates
10. audit_log
11. config

---

## Tasks

### Phase 1: Migration System Setup (TDD)
- [x] Write test for migrator creation
- [x] Write test for migration up
- [x] Write test for migration down
- [x] Write test for version tracking
- [x] Implement Migrator interface
- [x] Implement NewMigrator() constructor
- [x] Run tests (should pass)

### Phase 2: Create Migration Files
- [x] Create migrations/sqlite/README.md
- [x] Create 001_initial_schema.up.sql with all 11 tables
- [x] Create 001_initial_schema.down.sql
- [x] Create 002_add_indexes.up.sql
- [x] Create 002_add_indexes.down.sql

### Phase 3: Migration Tests (TDD)
- [x] Write test to verify all tables created
- [x] Write test to verify all columns exist
- [x] Write test to verify foreign keys work
- [x] Write test to verify constraints work
- [x] Write test for down migration
- [x] Run tests (should pass)

### Phase 4: Integration
- [x] Update cmd/server/main.go to run migrations on startup
- [x] Add migration logging
- [x] Test with real database file
- [x] Document migration process

---

## Testing Requirements

See Database LLD Section 8 for complete test examples.

Key tests:
- Migration up creates all tables
- Migration down removes all tables
- Foreign key constraints enforced
- Check constraints validated
- Version tracking works

---

## Dependencies

**Depends on:** 
- [00_STORY_go_module_setup.md](00_STORY_go_module_setup.md)
- [00_STORY_database_connection.md](00_STORY_database_connection.md)

**Blocks:** 
- [00_STORY_repository_pattern.md](00_STORY_repository_pattern.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout
- [x] Test coverage >= 80% (83.1%)
- [x] All 11 tables created
- [x] All indexes created
- [x] Foreign key constraints verified
- [x] Up and down migrations tested
- [x] Code formatted with go fmt
- [x] No errors from go vet
- [x] Documentation complete
- [ ] Changes committed to git

---

## References

- **README-LLM.md:** TDD Requirements
- **HLD:** Section 13 (Database Schema)
- **LLD:** [lld/07_DATABASE_LLD.md](../lld/07_DATABASE_LLD.md) - Complete schema with all SQL
- **golang-migrate:** https://github.com/golang-migrate/migrate

---

## Notes

For complete SQL migration files, see Database LLD Section 5.2 which contains:
- Full CREATE TABLE statements for all 11 tables
- All CHECK constraints
- All foreign key relationships
- All indexes
- Up and down migrations

The implementation should copy these SQL statements into migration files.
