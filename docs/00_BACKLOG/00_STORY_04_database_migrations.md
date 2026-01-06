# User Story: Database Migrations

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 5 hours

---

## User Story

As a **developer**, I want **automated database migrations** so that **the database schema can be versioned and deployed reliably**.

---

## Acceptance Criteria

- [ ] Migration system using golang-migrate integrated
- [ ] All 11 database tables created via migrations
- [ ] Up and down migrations implemented
- [ ] Migrations run automatically on startup
- [ ] Migration version tracked in database
- [ ] Rollback capability functional
- [ ] All indexes and constraints defined
- [ ] All tests pass with timeout

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
- [ ] Write test for migrator creation
- [ ] Write test for migration up
- [ ] Write test for migration down
- [ ] Write test for version tracking
- [ ] Implement Migrator interface
- [ ] Implement NewMigrator() constructor
- [ ] Run tests (should pass)

### Phase 2: Create Migration Files
- [ ] Create migrations/sqlite/README.md
- [ ] Create 001_initial_schema.up.sql with all 11 tables
- [ ] Create 001_initial_schema.down.sql
- [ ] Create 002_add_indexes.up.sql
- [ ] Create 002_add_indexes.down.sql

### Phase 3: Migration Tests (TDD)
- [ ] Write test to verify all tables created
- [ ] Write test to verify all columns exist
- [ ] Write test to verify foreign keys work
- [ ] Write test to verify constraints work
- [ ] Write test for down migration
- [ ] Run tests (should pass)

### Phase 4: Integration
- [ ] Update cmd/server/main.go to run migrations on startup
- [ ] Add migration logging
- [ ] Test with real database file
- [ ] Document migration process

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

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout
- [ ] Test coverage >= 80%
- [ ] All 11 tables created
- [ ] All indexes created
- [ ] Foreign key constraints verified
- [ ] Up and down migrations tested
- [ ] Code formatted with go fmt
- [ ] No errors from go vet
- [ ] Documentation complete
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
