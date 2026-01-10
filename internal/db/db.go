package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Type         string
	Path         string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type Database interface {
	DB() *sql.DB
	Close() error
	Ping(ctx context.Context) error
	WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type database struct {
	db *sql.DB
}

func NewDatabase(cfg Config) (Database, error) {
	if cfg.Type == "" {
		return nil, fmt.Errorf("database type is required")
	}

	if cfg.Path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	var dsn string

	switch cfg.Type {
	case "sqlite":
		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=1", cfg.Path)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.MaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxLifetime)
	}

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &database{db: db}, nil
}

// NewDatabaseWithRetry creates a new database connection with automatic retry logic
func NewDatabaseWithRetry(cfg Config) (Database, error) {
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return NewRetryableDatabase(db, DefaultRetryConfig), nil
}

func (d *database) DB() *sql.DB {
	return d.db
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *database) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}
