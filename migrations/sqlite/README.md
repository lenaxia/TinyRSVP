# SQLite Migrations

This directory contains database migration files for TinyRSVP using SQLite.

## Migration Tool

Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate).

## File Naming Convention

Migration files follow the pattern: `{version}_{description}.{direction}.sql`

- `version`: Sequential number (e.g., 000001, 000002)
- `description`: Brief description with underscores
- `direction`: Either `up` or `down`

Examples:
- `000001_initial_schema.up.sql`
- `000001_initial_schema.down.sql`

## Migration Files

### 000001_initial_schema

Creates all 11 core tables:
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

### 000002_add_indexes

Adds performance indexes for:
- Foreign key columns
- Frequently queried columns
- Composite indexes for common queries

## Running Migrations

Migrations are automatically run on application startup via [`cmd/server/main.go`](../../cmd/server/main.go).

### Manual Migration Commands

```bash
# Apply all pending migrations
go run cmd/server/main.go

# Or use the migrator directly in code:
migrator, _ := db.NewMigrator(database.DB(), "migrations/sqlite")
migrator.Up(context.Background())
```

### Rollback

```bash
# Rollback one migration
migrator.Steps(context.Background(), -1)

# Rollback all migrations
migrator.Down(context.Background())
```

## Schema Design

See [`docs/lld/07_DATABASE_LLD.md`](../../docs/lld/07_DATABASE_LLD.md) for complete schema documentation.

## Important Notes

- Foreign keys are enabled via connection string: `_foreign_keys=1`
- WAL mode is enabled for better concurrency: `_journal_mode=WAL`
- All timestamps use UTC
- All text fields use UTF-8 encoding
- Optimistic locking via version columns where applicable
