package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	Version(ctx context.Context) (uint, bool, error)
	Steps(ctx context.Context, n int) error
}

type migrator struct {
	migrate *migrate.Migrate
}

func NewMigrator(db *sql.DB, migrationsPath string) (Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &migrator{migrate: m}, nil
}

func (m *migrator) Up(ctx context.Context) error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}
	return nil
}

func (m *migrator) Down(ctx context.Context) error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down failed: %w", err)
	}
	return nil
}

func (m *migrator) Version(ctx context.Context) (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	if err == migrate.ErrNilVersion {
		return 0, false, err
	}
	return version, dirty, nil
}

func (m *migrator) Steps(ctx context.Context, n int) error {
	if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration steps failed: %w", err)
	}
	return nil
}
