# Database Package

## Purpose

Provides database connection management, query execution, transaction support, and automated migrations for TinyRSVP.

## Key Components

- **Database Interface**: Abstraction for database operations
- **Migrator Interface**: Automated schema migrations
- **Connection Pooling**: Configured connection pool with limits
- **Transaction Support**: ACID transactions with rollback on error/panic
- **Context-Aware**: All operations support context cancellation
- **SQLite Optimizations**: WAL mode, foreign keys, busy timeout

## Usage

```go
import "github.com/lenaxia/tinyrsvp/internal/db"

// Create database connection
db, err := db.NewDatabase(db.Config{
    Type:         "sqlite",
    Path:         "/data/tinyrsvp.db",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    MaxLifetime:  5 * time.Minute,
})
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Execute query
result, err := db.Exec(ctx, "INSERT INTO users (email) VALUES (?)", "user@example.com")

// Query rows
rows, err := db.Query(ctx, "SELECT * FROM users")

// Query single row
var email string
err := db.QueryRow(ctx, "SELECT email FROM users WHERE id = ?", 1).Scan(&email)

// Transaction
err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
    _, err := tx.Exec("INSERT INTO users (email) VALUES (?)", "user@example.com")
    if err != nil {
        return err // Automatic rollback
    }
    return nil // Automatic commit
})
```

## SQLite Configuration

The database connection string includes:
- `cache=shared` - Shared cache mode for better concurrency
- `mode=rwc` - Read-write-create mode
- `_journal_mode=WAL` - Write-Ahead Logging for better performance
- `_busy_timeout=5000` - 5 second timeout for busy database
- `_foreign_keys=1` - Enable foreign key constraints

## Connection Pool Settings

Default settings:
- **MaxOpenConns**: 25 - Maximum concurrent connections
- **MaxIdleConns**: 5 - Idle connections kept ready
- **MaxLifetime**: 5 minutes - Connection recycling interval

## Transaction Behavior

- Automatic rollback on error
- Automatic rollback on panic (with re-panic)
- Context cancellation support
- No nested transactions

## Testing

All database operations are tested with:
- Happy path scenarios
- Error conditions
- Context cancellation
- Transaction rollback
- SQLite-specific features

Run tests:
```bash
go test -timeout 30s -v ./internal/db/...
```

Check coverage:
```bash
go test -timeout 30s -cover ./internal/db/...
```

## Migrations

Automated database migrations using golang-migrate:

```go
// Run migrations on startup
migrator, err := db.NewMigrator(database.DB(), "migrations/sqlite")
if err != nil {
    log.Fatal(err)
}

if err := migrator.Up(context.Background()); err != nil {
    log.Fatal(err)
}

// Check migration version
version, dirty, err := migrator.Version(context.Background())
```

**Migration Files:**
- Location: `migrations/sqlite/`
- Format: `{version}_{description}.{up|down}.sql`
- See: [`migrations/sqlite/README.md`](../../migrations/sqlite/README.md)

**Operations:**
- `Up()` - Apply all pending migrations
- `Down()` - Rollback all migrations
- `Version()` - Get current migration version
- `Steps(n)` - Apply/rollback n migrations

## Dependencies

- `database/sql` - Go standard library
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/golang-migrate/migrate/v4` - Migration tool
- `github.com/golang-migrate/migrate/v4/database/sqlite3` - SQLite driver
- `github.com/golang-migrate/migrate/v4/source/file` - File source

## Repository Pattern

The database package includes repository implementations for data access abstraction.

### Available Repositories

- **UserRepository** - User CRUD operations and queries
- **SessionRepository** - Session management and cleanup
- **ConfigRepository** - Key-value configuration storage

### Usage Example

```go
import (
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

// Create repositories
userRepo := repositories.NewUserRepository(database)
sessionRepo := repositories.NewSessionRepository(database)
configRepo := repositories.NewConfigRepository(database)

// User operations
user := &models.User{
    Email: "user@example.com",
    Name:  "John Doe",
    Role:  models.RoleEventManager,
}
err := userRepo.Create(ctx, user)

// Session operations
session := &models.Session{
    ID:        "session-id",
    UserID:    user.ID,
    ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
}
err := sessionRepo.Create(ctx, session)

// Config operations
err := configRepo.Set(ctx, "smtp_host", "smtp.example.com")
config, err := configRepo.Get(ctx, "smtp_host")

// HMAC secret (auto-generated on first call)
secret, err := configRepo.GetHMACSecret(ctx)
```

### Repository Features

- **Type Safety**: All repositories use strongly-typed models
- **Error Mapping**: Database errors mapped to domain errors
- **Context Support**: All methods accept context for cancellation
- **Validation**: Input validation before database operations
- **Testing**: Comprehensive test coverage with table-driven tests

### Domain Errors

Repositories return domain-specific errors:
- `NotFoundError` - Resource not found
- `ConflictError` - Unique constraint violation
- `ValidationError` - Input validation failure
- `OptimisticLockError` - Version mismatch (future)

## Related Packages

- `internal/config` - Database configuration
- `internal/db/repositories` - Repository implementations
- `internal/models` - Domain models and errors
- `migrations/sqlite` - Migration SQL files
