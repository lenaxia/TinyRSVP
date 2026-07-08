package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
)

func TestIsSQLiteLockedError_NilError(t *testing.T) {
	if isSQLiteLockedError(nil) {
		t.Error("nil error should not be a locked error")
	}
}

func TestIsSQLiteLockedError_NonSQLiteError(t *testing.T) {
	if isSQLiteLockedError(errors.New("some other error")) {
		t.Error("non-SQLite error should not be a locked error")
	}
}

func TestIsSQLiteLockedError_SQLiteLockedError(t *testing.T) {
	err := sqlite3.Error{Code: sqlite3.ErrLocked}
	if !isSQLiteLockedError(err) {
		t.Error("ErrLocked should be a locked error")
	}
}

func TestIsSQLiteLockedError_SQLiteBusyError(t *testing.T) {
	err := sqlite3.Error{Code: sqlite3.ErrBusy}
	if !isSQLiteLockedError(err) {
		t.Error("ErrBusy should be a locked error")
	}
}

func TestIsSQLiteLockedError_SQLiteLockedSharedCache(t *testing.T) {
	err := sqlite3.Error{ExtendedCode: sqlite3.ErrLockedSharedCache}
	if !isSQLiteLockedError(err) {
		t.Error("ErrLockedSharedCache should be a locked error")
	}
}

func TestIsSQLiteLockedError_OtherSQLiteError(t *testing.T) {
	err := sqlite3.Error{Code: sqlite3.ErrConstraint}
	if isSQLiteLockedError(err) {
		t.Error("ErrConstraint should not be a locked error")
	}
}

func TestNewRetryableDatabase(t *testing.T) {
	mockDB := &mockDatabase{}
	config := RetryConfig{MaxAttempts: 3, InitialDelay: 1 * time.Millisecond}
	rdb := NewRetryableDatabase(mockDB, config)

	if rdb == nil {
		t.Fatal("expected non-nil RetryableDatabase")
	}
	if rdb.config.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", rdb.config.MaxAttempts)
	}
}

func TestRetryableDatabase_DB(t *testing.T) {
	mockDB := &mockDatabase{sqlDB: &sql.DB{}}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	if rdb.DB() == nil {
		t.Error("expected non-nil sql.DB")
	}
}

func TestRetryableDatabase_Close(t *testing.T) {
	mockDB := &mockDatabase{closed: false}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	if err := rdb.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	if !mockDB.closed {
		t.Error("expected underlying DB to be closed")
	}
}

func TestRetryableDatabase_Ping_Success(t *testing.T) {
	mockDB := &mockDatabase{pingErr: nil}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	if err := rdb.Ping(context.Background()); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestRetryableDatabase_Ping_RetriesOnLock(t *testing.T) {
	lockErr := sqlite3.Error{Code: sqlite3.ErrLocked}
	mockDB := &mockDatabase{
		pingErrs: []error{lockErr, lockErr, nil},
	}
	config := RetryConfig{MaxAttempts: 5, InitialDelay: 1 * time.Millisecond, MaxDelay: 5 * time.Millisecond, BackoffMultiplier: 2.0}
	rdb := NewRetryableDatabase(mockDB, config)

	if err := rdb.Ping(context.Background()); err != nil {
		t.Errorf("Ping() after retries error = %v", err)
	}
	if mockDB.pingCallCount != 3 {
		t.Errorf("expected 3 ping calls, got %d", mockDB.pingCallCount)
	}
}

func TestRetryableDatabase_Ping_NonRetryableError(t *testing.T) {
	nonRetryable := errors.New("connection refused")
	mockDB := &mockDatabase{pingErr: nonRetryable}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	err := rdb.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, nonRetryable) {
		t.Errorf("expected connection refused error, got %v", err)
	}
	if mockDB.pingCallCount != 1 {
		t.Errorf("expected 1 ping call (no retry for non-locked error), got %d", mockDB.pingCallCount)
	}
}

func TestRetryableDatabase_Ping_MaxRetriesExceeded(t *testing.T) {
	lockErr := sqlite3.Error{Code: sqlite3.ErrLocked}
	mockDB := &mockDatabase{pingErr: lockErr}
	config := RetryConfig{MaxAttempts: 3, InitialDelay: 1 * time.Millisecond, MaxDelay: 2 * time.Millisecond, BackoffMultiplier: 2.0}
	rdb := NewRetryableDatabase(mockDB, config)

	err := rdb.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}
	if mockDB.pingCallCount != 3 {
		t.Errorf("expected 3 ping calls, got %d", mockDB.pingCallCount)
	}
}

func TestRetryableDatabase_Ping_ContextCancelledBeforeAttempt(t *testing.T) {
	mockDB := &mockDatabase{}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rdb.Ping(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetryableDatabase_Ping_ContextCancelledDuringBackoff(t *testing.T) {
	lockErr := sqlite3.Error{Code: sqlite3.ErrLocked}
	mockDB := &mockDatabase{pingErr: lockErr}
	config := RetryConfig{MaxAttempts: 10, InitialDelay: 1 * time.Second, MaxDelay: 5 * time.Second, BackoffMultiplier: 2.0}
	rdb := NewRetryableDatabase(mockDB, config)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := rdb.Ping(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context during backoff")
	}
}

func TestRetryableDatabase_Exec_Success(t *testing.T) {
	mockDB := &mockDatabase{execResult: &mockResult{}}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	result, err := rdb.Exec(context.Background(), "INSERT INTO test VALUES (?)", 1)
	if err != nil {
		t.Errorf("Exec() error = %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestRetryableDatabase_Exec_RetriesOnLock(t *testing.T) {
	lockErr := sqlite3.Error{Code: sqlite3.ErrLocked}
	mockDB := &mockDatabase{
		execErrs: []error{lockErr, nil},
		execResult: &mockResult{},
	}
	config := RetryConfig{MaxAttempts: 5, InitialDelay: 1 * time.Millisecond, MaxDelay: 5 * time.Millisecond, BackoffMultiplier: 2.0}
	rdb := NewRetryableDatabase(mockDB, config)

	_, err := rdb.Exec(context.Background(), "INSERT INTO test VALUES (?)", 1)
	if err != nil {
		t.Errorf("Exec() after retry error = %v", err)
	}
	if mockDB.execCallCount != 2 {
		t.Errorf("expected 2 exec calls, got %d", mockDB.execCallCount)
	}
}

func TestRetryableDatabase_Query_Success(t *testing.T) {
	mockDB := &mockDatabase{queryErr: nil}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	_, err := rdb.Query(context.Background(), "SELECT 1")
	if err != nil {
		t.Errorf("Query() error = %v", err)
	}
}

func TestRetryableDatabase_Query_RetriesOnLock(t *testing.T) {
	lockErr := sqlite3.Error{Code: sqlite3.ErrLocked}
	mockDB := &mockDatabase{
		queryErrs: []error{lockErr, nil},
	}
	config := RetryConfig{MaxAttempts: 5, InitialDelay: 1 * time.Millisecond, MaxDelay: 5 * time.Millisecond, BackoffMultiplier: 2.0}
	rdb := NewRetryableDatabase(mockDB, config)

	_, err := rdb.Query(context.Background(), "SELECT 1")
	if err != nil {
		t.Errorf("Query() after retry error = %v", err)
	}
}

func TestRetryableDatabase_QueryRow(t *testing.T) {
	mockDB := &mockDatabase{}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	// QueryRow delegates to the underlying DB — just verify it doesn't panic
	_ = rdb.QueryRow(context.Background(), "SELECT 1")
	mockDB.queryRowCalled = true
}

func TestRetryableDatabase_WithTransaction_Success(t *testing.T) {
	mockDB := &mockDatabase{txErr: nil}
	rdb := NewRetryableDatabase(mockDB, DefaultRetryConfig)

	err := rdb.WithTransaction(context.Background(), func(tx *sql.Tx) error {
		return nil
	})
	if err != nil {
		t.Errorf("WithTransaction() error = %v", err)
	}
}

func TestRetryableDatabase_WithTransaction_RetriesOnLock(t *testing.T) {
	lockErr := sqlite3.Error{Code: sqlite3.ErrLocked}
	mockDB := &mockDatabase{
		txErrs: []error{lockErr, nil},
	}
	config := RetryConfig{MaxAttempts: 5, InitialDelay: 1 * time.Millisecond, MaxDelay: 5 * time.Millisecond, BackoffMultiplier: 2.0}
	rdb := NewRetryableDatabase(mockDB, config)

	err := rdb.WithTransaction(context.Background(), func(tx *sql.Tx) error {
		return nil
	})
	if err != nil {
		t.Errorf("WithTransaction() after retry error = %v", err)
	}
	if mockDB.txCallCount != 2 {
		t.Errorf("expected 2 tx calls, got %d", mockDB.txCallCount)
	}
}

func TestCalculateBackoff_ExponentialGrowth(t *testing.T) {
	rdb := &RetryableDatabase{
		config: RetryConfig{
			InitialDelay:      10 * time.Millisecond,
			MaxDelay:          1 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 10 * time.Millisecond},
		{1, 20 * time.Millisecond},
		{2, 40 * time.Millisecond},
		{3, 80 * time.Millisecond},
	}

	for _, tt := range tests {
		got := rdb.calculateBackoff(tt.attempt)
		if got != tt.want {
			t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestCalculateBackoff_CappedAtMaxDelay(t *testing.T) {
	rdb := &RetryableDatabase{
		config: RetryConfig{
			InitialDelay:      100 * time.Millisecond,
			MaxDelay:          500 * time.Millisecond,
			BackoffMultiplier: 10.0,
		},
	}

	got := rdb.calculateBackoff(10)
	if got != 500*time.Millisecond {
		t.Errorf("calculateBackoff(10) = %v, want 500ms (capped)", got)
	}
}

func TestCalculateBackoff_ZeroAttempt(t *testing.T) {
	rdb := &RetryableDatabase{
		config: RetryConfig{
			InitialDelay:      50 * time.Millisecond,
			MaxDelay:          1 * time.Second,
			BackoffMultiplier: 1.5,
		},
	}

	got := rdb.calculateBackoff(0)
	if got != 50*time.Millisecond {
		t.Errorf("calculateBackoff(0) = %v, want 50ms", got)
	}
}

// --- Mock implementations ---

type mockResult struct{}

func (m *mockResult) LastInsertId() (int64, error) { return 1, nil }
func (m *mockResult) RowsAffected() (int64, error) { return 1, nil }

type mockDatabase struct {
	sqlDB          *sql.DB
	closed         bool
	pingErr        error
	pingErrs       []error
	pingCallCount  int
	execResult     sql.Result
	execErr        error
	execErrs       []error
	execCallCount  int
	queryErr       error
	queryErrs      []error
	queryCallCount int
	txErr          error
	txErrs         []error
	txCallCount    int
	queryRowCalled bool
}

func (m *mockDatabase) DB() *sql.DB { return m.sqlDB }

func (m *mockDatabase) Close() error {
	m.closed = true
	return nil
}

func (m *mockDatabase) Ping(ctx context.Context) error {
	m.pingCallCount++
	if len(m.pingErrs) > 0 {
		err := m.pingErrs[0]
		m.pingErrs = m.pingErrs[1:]
		return err
	}
	return m.pingErr
}

func (m *mockDatabase) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	m.txCallCount++
	if len(m.txErrs) > 0 {
		err := m.txErrs[0]
		m.txErrs = m.txErrs[1:]
		return err
	}
	return m.txErr
}

func (m *mockDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	m.execCallCount++
	if len(m.execErrs) > 0 {
		err := m.execErrs[0]
		m.execErrs = m.execErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	if m.execErr != nil {
		return nil, m.execErr
	}
	return m.execResult, nil
}

func (m *mockDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	m.queryCallCount++
	if len(m.queryErrs) > 0 {
		err := m.queryErrs[0]
		m.queryErrs = m.queryErrs[1:]
		return nil, err
	}
	return nil, m.queryErr
}

func (m *mockDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
