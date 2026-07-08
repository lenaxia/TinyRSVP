package db

import (
	"testing"
)

func TestNewDatabaseWithRetry(t *testing.T) {
	cfg := Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	}

	db, err := NewDatabaseWithRetry(cfg)
	if err != nil {
		t.Fatalf("NewDatabaseWithRetry: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("expected non-nil database")
	}

	if _, ok := db.(*RetryableDatabase); !ok {
		t.Error("expected *RetryableDatabase")
	}
}

func TestNewDatabaseWithRetry_InvalidConfig(t *testing.T) {
	cfg := Config{
		Type: "invalid",
	}

	_, err := NewDatabaseWithRetry(cfg)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}
