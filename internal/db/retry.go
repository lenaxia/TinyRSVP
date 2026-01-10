package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/mattn/go-sqlite3"
)

// RetryConfig defines retry behavior for database operations
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig provides sensible defaults for SQLite retry logic
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:       5,
	InitialDelay:      10 * time.Millisecond,
	MaxDelay:          1 * time.Second,
	BackoffMultiplier: 2.0,
}

// isSQLiteLockedError checks if an error is a SQLite locked/busy error
func isSQLiteLockedError(err error) bool {
	if err == nil {
		return false
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrLocked ||
			sqliteErr.Code == sqlite3.ErrBusy ||
			sqliteErr.ExtendedCode == sqlite3.ErrLockedSharedCache
	}

	return false
}

// RetryableDatabase wraps a Database with automatic retry logic
type RetryableDatabase struct {
	db     Database
	config RetryConfig
}

// NewRetryableDatabase creates a new database wrapper with retry logic
func NewRetryableDatabase(db Database, config RetryConfig) *RetryableDatabase {
	return &RetryableDatabase{
		db:     db,
		config: config,
	}
}

// DB returns the underlying sql.DB
func (r *RetryableDatabase) DB() *sql.DB {
	return r.db.DB()
}

// Close closes the database connection
func (r *RetryableDatabase) Close() error {
	return r.db.Close()
}

// Ping pings the database with retry logic
func (r *RetryableDatabase) Ping(ctx context.Context) error {
	return r.retryOperation(ctx, func() error {
		return r.db.Ping(ctx)
	})
}

// WithTransaction executes a function within a transaction with retry logic
func (r *RetryableDatabase) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	return r.retryOperation(ctx, func() error {
		return r.db.WithTransaction(ctx, fn)
	})
}

// Exec executes a query with retry logic
func (r *RetryableDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	err := r.retryOperation(ctx, func() error {
		var execErr error
		result, execErr = r.db.Exec(ctx, query, args...)
		return execErr
	})
	return result, err
}

// Query executes a query with retry logic
func (r *RetryableDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	err := r.retryOperation(ctx, func() error {
		var queryErr error
		rows, queryErr = r.db.Query(ctx, query, args...)
		return queryErr
	})
	return rows, err
}

// QueryRow executes a query that returns a single row with retry logic
func (r *RetryableDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Note: QueryRow doesn't return an error directly, so we can't retry it
	// The error is deferred until Scan() is called
	// For proper retry support, use Query() instead
	return r.db.QueryRow(ctx, query, args...)
}

// retryOperation executes an operation with exponential backoff retry logic
func (r *RetryableDatabase) retryOperation(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled before attempt %d: %w", attempt+1, err)
		}

		// Execute the operation
		lastErr = operation()

		// Success - return immediately
		if lastErr == nil {
			return nil
		}

		// Not a retryable error - return immediately
		if !isSQLiteLockedError(lastErr) {
			return lastErr
		}

		// Last attempt - don't sleep, just return the error
		if attempt == r.config.MaxAttempts-1 {
			return fmt.Errorf("max retry attempts (%d) exceeded: %w", r.config.MaxAttempts, lastErr)
		}

		// Calculate backoff delay with exponential increase
		delay := r.calculateBackoff(attempt)

		// Log retry attempt (in production, use proper logging)
		// For now, we'll just sleep
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
		}
	}

	return fmt.Errorf("max retry attempts exceeded: %w", lastErr)
}

// calculateBackoff calculates the backoff delay for a given attempt
func (r *RetryableDatabase) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: initialDelay * (multiplier ^ attempt)
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffMultiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	return time.Duration(delay)
}
