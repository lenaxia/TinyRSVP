package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

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
		{
			name: "minimal valid config",
			config: Config{
				Type: "sqlite",
				Path: ":memory:",
			},
			wantErr: false,
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
		{
			name:    "timeout context",
			ctx:     timeoutContext(),
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

func TestDatabase_Close(t *testing.T) {
	db := setupTestDB(t)

	err := db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	err = db.Ping(context.Background())
	if err == nil {
		t.Error("Expected error after Close(), got nil")
	}
}

func TestDatabase_ConnectionPooling(t *testing.T) {
	config := Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 10,
		MaxIdleConns: 3,
		MaxLifetime:  2 * time.Minute,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("NewDatabase() error = %v", err)
	}
	defer db.Close()

	stats := db.DB().Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections = 10, got %d", stats.MaxOpenConnections)
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

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
			_, err := tx.Exec("INSERT INTO test (name) VALUES (?)", "test4")
			return err
		})

		if err == nil {
			t.Error("Expected error with cancelled context, got nil")
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

	t.Run("Exec with cancelled context", func(t *testing.T) {
		ctx := cancelledContext()
		_, err := db.Exec(ctx, "INSERT INTO test (name) VALUES (?)", "test2")
		if err == nil {
			t.Error("Expected error with cancelled context, got nil")
		}
	})

	t.Run("Query with cancelled context", func(t *testing.T) {
		ctx := cancelledContext()
		_, err := db.Query(ctx, "SELECT name FROM test")
		if err == nil {
			t.Error("Expected error with cancelled context, got nil")
		}
	})
}

func TestSQLiteWALMode(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	db, err := NewDatabase(Config{
		Type:         "sqlite",
		Path:         tmpFile,
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Minute,
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	var journalMode string
	err = db.QueryRow(context.Background(), "PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got %q", journalMode)
	}
}

func TestSQLiteForeignKeys(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var foreignKeys int
	err := db.QueryRow(context.Background(), "PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("Failed to query foreign_keys: %v", err)
	}

	if foreignKeys != 1 {
		t.Errorf("Expected foreign_keys enabled (1), got %d", foreignKeys)
	}
}

func TestSQLiteBusyTimeout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var timeout int
	err := db.QueryRow(context.Background(), "PRAGMA busy_timeout").Scan(&timeout)
	if err != nil {
		t.Fatalf("Failed to query busy_timeout: %v", err)
	}

	if timeout < 5000 {
		t.Errorf("Expected busy_timeout >= 5000ms, got %d", timeout)
	}
}

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

func timeoutContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Nanosecond)
	return ctx
}
