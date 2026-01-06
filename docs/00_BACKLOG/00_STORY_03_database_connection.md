# User Story: Database Connection Management

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 4 hours

---

## User Story

As a **developer**, I want **reliable database connection management with pooling** so that **the application can efficiently interact with the database**.

---

## Acceptance Criteria

- [ ] Database connection established to SQLite
- [ ] Connection pooling configured with appropriate limits
- [ ] WAL mode enabled for SQLite
- [ ] Transaction support implemented
- [ ] Context-aware query execution
- [ ] Graceful connection shutdown
- [ ] Connection health checks functional
- [ ] All tests pass with timeout

---

## Technical Details

### Database Interface

```go
package db

import (
    "context"
    "database/sql"
)

type Database interface {
    DB() *sql.DB
    Close() error
    Ping(ctx context.Context) error
    WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}
```

### Configuration

```go
type Config struct {
    Type         string
    Path         string
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  time.Duration
}
```

### SQLite Connection String

```
file:/data/tinyrsvp.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000
```

**Parameters:**
- `cache=shared` - Enable shared cache mode
- `mode=rwc` - Read-write-create mode
- `_journal_mode=WAL` - Write-Ahead Logging for better concurrency
- `_busy_timeout=5000` - 5 second timeout for busy database

### Connection Pool Settings

```go
db.SetMaxOpenConns(25)      // Maximum open connections
db.SetMaxIdleConns(5)       // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
```

---

## Tasks

### Phase 1: Database Interface (TDD)
- [ ] Write test for creating database connection
- [ ] Write test for connection failure scenarios
- [ ] Write test for ping functionality
- [ ] Implement `Database` interface
- [ ] Implement `NewDatabase()` constructor
- [ ] Run tests (should pass)

### Phase 2: Connection Management (TDD)
- [ ] Write test for connection pooling configuration
- [ ] Write test for connection lifecycle
- [ ] Write test for graceful shutdown
- [ ] Implement connection pool setup
- [ ] Implement `Close()` method
- [ ] Run tests (should pass)

### Phase 3: Query Execution (TDD)
- [ ] Write test for `Exec()` with context
- [ ] Write test for `Query()` with context
- [ ] Write test for `QueryRow()` with context
- [ ] Write test for context cancellation
- [ ] Implement query execution methods
- [ ] Run tests (should pass)

### Phase 4: Transaction Support (TDD)
- [ ] Write test for successful transaction
- [ ] Write test for transaction rollback on error
- [ ] Write test for transaction rollback on panic
- [ ] Write test for nested transaction prevention
- [ ] Implement `WithTransaction()` method
- [ ] Run tests (should pass)

### Phase 5: SQLite Specific (TDD)
- [ ] Write test for WAL mode verification
- [ ] Write test for busy timeout handling
- [ ] Write test for foreign key enforcement
- [ ] Implement SQLite-specific setup
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Update `cmd/server/main.go` to initialize database
- [ ] Add database health logging
- [ ] Test connection with real SQLite file
- [ ] Document connection parameters

---

## Testing Requirements

### Unit Tests

```go
func TestNewDatabase(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name: "valid sqlite config",
            config: Config{
                Type:         "sqlite",
                Path:         ":memory:",
                MaxOpenConns: 25,
                MaxIdleConns: 5,
                MaxLifetime:  5 * time.Minute,
            },
            wantErr: false,
        },
        {
            name: "invalid database type",
            config: Config{
                Type: "invalid",
                Path: ":memory:",
            },
            wantErr: true,
        },
        {
            name: "empty path",
            config: Config{
                Type: "sqlite",
                Path: "",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db, err := NewDatabase(tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewDatabase() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                if db == nil {
                    t.Error("Expected database, got nil")
                }
                defer db.Close()
                
                if err := db.Ping(context.Background()); err != nil {
                    t.Errorf("Ping() failed: %v", err)
                }
            }
        })
    }
}

func TestDatabase_Ping(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    tests := []struct {
        name    string
        ctx     context.Context
        wantErr bool
    }{
        {
            name:    "valid context",
            ctx:     context.Background(),
            wantErr: false,
        },
        {
            name:    "cancelled context",
            ctx:     cancelledContext(),
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := db.Ping(tt.ctx)
            if (err != nil) != tt.wantErr {
                t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestDatabase_WithTransaction(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    createTable(t, db)
    
    t.Run("successful transaction", func(t *testing.T) {
        err := db.WithTransaction(context.Background(), func(tx *sql.Tx) error {
            _, err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test1")
            return err
        })
        
        if err != nil {
            t.Errorf("WithTransaction() error = %v", err)
        }
        
        var count int
        err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM test").Scan(&count)
        if err != nil {
            t.Fatalf("Failed to query count: %v", err)
        }
        
        if count != 1 {
            t.Errorf("Expected 1 row, got %d", count)
        }
    })
    
    t.Run("rollback on error", func(t *testing.T) {
        initialCount := getCount(t, db)
        
        err := db.WithTransaction(context.Background(), func(tx *sql.Tx) error {
            _, err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test2")
            if err != nil {
                return err
            }
            return fmt.Errorf("intentional error")
        })
        
        if err == nil {
            t.Error("Expected error, got nil")
        }
        
        finalCount := getCount(t, db)
        if finalCount != initialCount {
            t.Errorf("Expected count %d, got %d (transaction not rolled back)", initialCount, finalCount)
        }
    })
    
    t.Run("rollback on panic", func(t *testing.T) {
        initialCount := getCount(t, db)
        
        func() {
            defer func() {
                if r := recover(); r == nil {
                    t.Error("Expected panic")
                }
            }()
            
            db.WithTransaction(context.Background(), func(tx *sql.Tx) error {
                _, err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test3")
                if err != nil {
                    return err
                }
                panic("intentional panic")
            })
        }()
        
        finalCount := getCount(t, db)
        if finalCount != initialCount {
            t.Errorf("Expected count %d, got %d (transaction not rolled back)", initialCount, finalCount)
        }
    })
}

func TestDatabase_QueryExecution(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    createTable(t, db)
    
    t.Run("Exec", func(t *testing.T) {
        result, err := db.Exec(context.Background(), 
            "INSERT INTO test (name) VALUES (?)", "test1")
        if err != nil {
            t.Fatalf("Exec() error = %v", err)
        }
        
        rows, err := result.RowsAffected()
        if err != nil {
            t.Fatalf("RowsAffected() error = %v", err)
        }
        
        if rows != 1 {
            t.Errorf("Expected 1 row affected, got %d", rows)
        }
    })
    
    t.Run("Query", func(t *testing.T) {
        rows, err := db.Query(context.Background(), "SELECT name FROM test")
        if err != nil {
            t.Fatalf("Query() error = %v", err)
        }
        defer rows.Close()
        
        count := 0
        for rows.Next() {
            var name string
            if err := rows.Scan(&name); err != nil {
                t.Fatalf("Scan() error = %v", err)
            }
            count++
        }
        
        if count == 0 {
            t.Error("Expected at least one row")
        }
    })
    
    t.Run("QueryRow", func(t *testing.T) {
        var name string
        err := db.QueryRow(context.Background(), 
            "SELECT name FROM test WHERE name = ?", "test1").Scan(&name)
        if err != nil {
            t.Fatalf("QueryRow() error = %v", err)
        }
        
        if name != "test1" {
            t.Errorf("Expected 'test1', got %q", name)
        }
    })
}

func TestSQLiteWALMode(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    var journalMode string
    err := db.QueryRow(context.Background(), "PRAGMA journal_mode").Scan(&journalMode)
    if err != nil {
        t.Fatalf("Failed to query journal_mode: %v", err)
    }
    
    if journalMode != "wal" {
        t.Errorf("Expected WAL mode, got %q", journalMode)
    }
}
```

### Test Helpers

```go
func setupTestDB(t *testing.T) Database {
    t.Helper()
    
    db, err := NewDatabase(Config{
        Type:         "sqlite",
        Path:         ":memory:",
        MaxOpenConns: 1,
        MaxIdleConns: 1,
        MaxLifetime:  time.Minute,
    })
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    
    return db
}

func createTable(t *testing.T, db Database) {
    t.Helper()
    
    _, err := db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS test (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )
    `)
    if err != nil {
        t.Fatalf("Failed to create test table: %v", err)
    }
}

func getCount(t *testing.T, db Database) int {
    t.Helper()
    
    var count int
    err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM test").Scan(&count)
    if err != nil {
        t.Fatalf("Failed to get count: %v", err)
    }
    return count
}

func cancelledContext() context.Context {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    return ctx
}
```

---

## Dependencies

**Depends on:** 
- [00_STORY_go_module_setup.md](00_STORY_go_module_setup.md)
- [00_STORY_config_management.md](00_STORY_config_management.md)

**Blocks:** 
- [00_STORY_database_migrations.md](00_STORY_database_migrations.md)
- [00_STORY_repository_pattern.md](00_STORY_repository_pattern.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/db/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] WAL mode verified
- [ ] Connection pooling verified
- [ ] Transaction rollback verified
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Notes

### SQLite WAL Mode Benefits
- Better concurrency (readers don't block writers)
- Faster in most scenarios
- More robust crash recovery

### Connection Pool Tuning
- `MaxOpenConns`: Limit total connections (prevents resource exhaustion)
- `MaxIdleConns`: Keep connections ready (reduces latency)
- `MaxLifetime`: Recycle connections (prevents stale connections)

### Transaction Best Practices
- Always use context for cancellation
- Handle panics to ensure rollback
- Keep transactions short
- Don't nest transactions

### Security Considerations
- Always use parameterized queries (SQL injection prevention)
- Never log query parameters (may contain sensitive data)
- Enable foreign key constraints
- Use prepared statements for repeated queries

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/07_DATABASE_LLD.md](../lld/07_DATABASE_LLD.md) - Section 5.1
- **SQLite Documentation:** https://www.sqlite.org/wal.html
- **Go database/sql:** https://pkg.go.dev/database/sql
