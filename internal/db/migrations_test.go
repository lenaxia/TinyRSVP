package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewMigrator(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(t *testing.T) *sql.DB
		migrationsPath string
		wantErr        bool
		errContains    string
	}{
		{
			name: "valid migrator creation",
			setupDB: func(t *testing.T) *sql.DB {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				return db
			},
			migrationsPath: createTestMigrations(t),
			wantErr:        false,
		},
		{
			name: "nil database",
			setupDB: func(t *testing.T) *sql.DB {
				return nil
			},
			migrationsPath: createTestMigrations(t),
			wantErr:        true,
			errContains:    "required",
		},
		{
			name: "invalid migrations path",
			setupDB: func(t *testing.T) *sql.DB {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				return db
			},
			migrationsPath: "/nonexistent/path",
			wantErr:        true,
			errContains:    "migrator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB(t)
			if db != nil {
				defer db.Close()
			}

			migrator, err := NewMigrator(db, tt.migrationsPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewMigrator() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewMigrator() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewMigrator() unexpected error = %v", err)
				return
			}

			if migrator == nil {
				t.Error("NewMigrator() returned nil migrator")
			}
		})
	}
}

func TestMigrator_Up(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(t *testing.T) (*sql.DB, string)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful migration up",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)
				return db, migrationsPath
			},
			wantErr: false,
		},
		{
			name: "migration up with no changes",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)

				migrator, err := NewMigrator(db, migrationsPath)
				if err != nil {
					t.Fatalf("failed to create migrator: %v", err)
				}

				if err := migrator.Up(context.Background()); err != nil {
					t.Fatalf("failed initial migration: %v", err)
				}

				return db, migrationsPath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, migrationsPath := tt.setupDB(t)
			defer db.Close()

			migrator, err := NewMigrator(db, migrationsPath)
			if err != nil {
				t.Fatalf("NewMigrator() error = %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = migrator.Up(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Up() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Up() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Up() unexpected error = %v", err)
			}
		})
	}
}

func TestMigrator_Down(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(t *testing.T) (*sql.DB, string)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful migration down",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)

				migrator, err := NewMigrator(db, migrationsPath)
				if err != nil {
					t.Fatalf("failed to create migrator: %v", err)
				}

				if err := migrator.Up(context.Background()); err != nil {
					t.Fatalf("failed migration up: %v", err)
				}

				return db, migrationsPath
			},
			wantErr: false,
		},
		{
			name: "migration down with no migrations applied",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)
				return db, migrationsPath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, migrationsPath := tt.setupDB(t)
			defer db.Close()

			migrator, err := NewMigrator(db, migrationsPath)
			if err != nil {
				t.Fatalf("NewMigrator() error = %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = migrator.Down(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Down() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Down() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Down() unexpected error = %v", err)
			}
		})
	}
}

func TestMigrator_Version(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(t *testing.T) (*sql.DB, string)
		wantVersion uint
		wantDirty   bool
		wantErr     bool
	}{
		{
			name: "no migrations applied",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)
				return db, migrationsPath
			},
			wantVersion: 0,
			wantDirty:   false,
			wantErr:     true,
		},
		{
			name: "migrations applied",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)

				migrator, err := NewMigrator(db, migrationsPath)
				if err != nil {
					t.Fatalf("failed to create migrator: %v", err)
				}

				if err := migrator.Up(context.Background()); err != nil {
					t.Fatalf("failed migration up: %v", err)
				}

				return db, migrationsPath
			},
			wantVersion: 1,
			wantDirty:   false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, migrationsPath := tt.setupDB(t)
			defer db.Close()

			migrator, err := NewMigrator(db, migrationsPath)
			if err != nil {
				t.Fatalf("NewMigrator() error = %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			version, dirty, err := migrator.Version(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Version() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Version() unexpected error = %v", err)
				return
			}

			if version != tt.wantVersion {
				t.Errorf("Version() version = %v, want %v", version, tt.wantVersion)
			}

			if dirty != tt.wantDirty {
				t.Errorf("Version() dirty = %v, want %v", dirty, tt.wantDirty)
			}
		})
	}
}

func TestMigrator_Steps(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(t *testing.T) (*sql.DB, string)
		steps       int
		wantErr     bool
		errContains string
	}{
		{
			name: "step forward one migration",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)
				return db, migrationsPath
			},
			steps:   1,
			wantErr: false,
		},
		{
			name: "step backward one migration",
			setupDB: func(t *testing.T) (*sql.DB, string) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatalf("failed to open database: %v", err)
				}
				migrationsPath := createTestMigrations(t)

				migrator, err := NewMigrator(db, migrationsPath)
				if err != nil {
					t.Fatalf("failed to create migrator: %v", err)
				}

				if err := migrator.Up(context.Background()); err != nil {
					t.Fatalf("failed migration up: %v", err)
				}

				return db, migrationsPath
			},
			steps:   -1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, migrationsPath := tt.setupDB(t)
			defer db.Close()

			migrator, err := NewMigrator(db, migrationsPath)
			if err != nil {
				t.Fatalf("NewMigrator() error = %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = migrator.Steps(ctx, tt.steps)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Steps() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Steps() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Steps() unexpected error = %v", err)
			}
		})
	}
}

func createTestMigrations(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	upSQL := `CREATE TABLE test_table (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`

	downSQL := `DROP TABLE IF EXISTS test_table;`

	upFile := filepath.Join(tmpDir, "000001_test.up.sql")
	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		t.Fatalf("failed to write up migration: %v", err)
	}

	downFile := filepath.Join(tmpDir, "000001_test.down.sql")
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		t.Fatalf("failed to write down migration: %v", err)
	}

	return tmpDir
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMigration_AllTablesCreated(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	expectedTables := []string{
		"users",
		"sessions",
		"events",
		"invites",
		"rsvps",
		"preference_questions",
		"rsvp_answers",
		"email_queue",
		"templates",
	}

	for _, tableName := range expectedTables {
		t.Run(tableName, func(t *testing.T) {
			var count int
			query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
			err := db.QueryRow(query, tableName).Scan(&count)
			if err != nil {
				t.Errorf("failed to query table %s: %v", tableName, err)
				return
			}
			if count != 1 {
				t.Errorf("table %s not found", tableName)
			}
		})
	}
}

func TestMigration_TableColumns(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	tests := []struct {
		table   string
		columns []string
	}{
		{
			table:   "users",
			columns: []string{"id", "email", "name", "role", "oidc_subject", "created_at", "updated_at", "last_login_at"},
		},
		{
			table:   "sessions",
			columns: []string{"id", "user_id", "created_at", "expires_at", "last_accessed_at", "ip_address", "user_agent"},
		},
		{
			table:   "events",
			columns: []string{"id", "title", "description", "location", "start_time", "end_time", "timezone", "rsvp_deadline", "max_plus_ones", "status", "template_id", "created_by", "created_at", "updated_at", "version", "ics_sequence"},
		},
		{
			table:   "invites",
			columns: []string{"id", "event_id", "name", "email", "token_hash", "max_plus_ones", "status", "sent_at", "viewed_at", "unsubscribed", "email_invalid", "created_at", "updated_at", "expires_at"},
		},
		{
			table:   "rsvps",
			columns: []string{"id", "invite_id", "response", "plus_ones", "created_at", "updated_at"},
		},
		{
			table:   "preference_questions",
			columns: []string{"id", "event_id", "question_text", "question_type", "options", "required", "display_order", "created_at"},
		},
		{
			table:   "rsvp_answers",
			columns: []string{"id", "rsvp_id", "question_id", "answer_text", "answer_option", "answer_boolean", "created_at", "updated_at"},
		},
		{
			table:   "email_queue",
			columns: []string{"id", "to_email", "to_name", "subject", "body_text", "body_html", "attachments", "status", "attempts", "max_attempts", "last_attempt_at", "last_error", "scheduled_for", "created_at"},
		},
		{
			table:   "templates",
			columns: []string{"id", "name", "type", "html_content", "text_content", "css_content", "is_default", "created_by", "created_at", "updated_at"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.table, func(t *testing.T) {
			rows, err := db.Query("PRAGMA table_info(" + tt.table + ")")
			if err != nil {
				t.Fatalf("failed to get table info: %v", err)
			}
			defer rows.Close()

			foundColumns := make(map[string]bool)
			for rows.Next() {
				var cid int
				var name, ctype string
				var notnull, pk int
				var dfltValue sql.NullString
				if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
					t.Fatalf("failed to scan row: %v", err)
				}
				foundColumns[name] = true
			}

			for _, col := range tt.columns {
				if !foundColumns[col] {
					t.Errorf("column %s not found in table %s", col, tt.table)
				}
			}
		})
	}
}

func TestMigration_ForeignKeys(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&_foreign_keys=1")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (email, name, role) VALUES ('test@example.com', 'Test User', 'admin')`)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO events (title, start_time, timezone, created_by) VALUES ('Test Event', datetime('now', '+1 day'), 'UTC', 1)`)
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	t.Run("foreign key violation", func(t *testing.T) {
		_, err = db.Exec(`INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (999, 'hash123', 2, datetime('now', '+30 days'))`)
		if err == nil {
			t.Error("expected foreign key constraint violation for non-existent event_id, got nil")
		}
	})

	t.Run("valid foreign key", func(t *testing.T) {
		_, err = db.Exec(`INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'hash123', 2, datetime('now', '+30 days'))`)
		if err != nil {
			t.Errorf("failed to insert invite with valid event_id: %v", err)
		}
	})

	t.Run("cascade delete", func(t *testing.T) {
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM invites WHERE event_id = 1`).Scan(&count)
		if err != nil {
			t.Fatalf("failed to count invites: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 invite, got %d", count)
		}

		_, err = db.Exec(`DELETE FROM events WHERE id = 1`)
		if err != nil {
			t.Fatalf("failed to delete event: %v", err)
		}

		err = db.QueryRow(`SELECT COUNT(*) FROM invites WHERE event_id = 1`).Scan(&count)
		if err != nil {
			t.Fatalf("failed to count invites after delete: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 invites after cascade delete, got %d", count)
		}
	})
}

func TestMigration_CheckConstraints(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "invalid user role",
			query:   `INSERT INTO users (email, name, role) VALUES ('test@example.com', 'Test', 'invalid_role')`,
			wantErr: true,
		},
		{
			name:    "valid user role admin",
			query:   `INSERT INTO users (email, name, role) VALUES ('admin@example.com', 'Admin', 'admin')`,
			wantErr: false,
		},
		{
			name:    "valid user role event_manager",
			query:   `INSERT INTO users (email, name, role) VALUES ('manager@example.com', 'Manager', 'event_manager')`,
			wantErr: false,
		},
		{
			name:    "invalid event status",
			query:   `INSERT INTO events (title, start_time, timezone, status, created_by) VALUES ('Test', datetime('now', '+1 day'), 'UTC', 'invalid', 1)`,
			wantErr: true,
		},
		{
			name:    "invalid max_plus_ones negative",
			query:   `INSERT INTO events (title, start_time, timezone, max_plus_ones, created_by) VALUES ('Test', datetime('now', '+1 day'), 'UTC', -1, 1)`,
			wantErr: true,
		},
		{
			name:    "invalid max_plus_ones too high",
			query:   `INSERT INTO events (title, start_time, timezone, max_plus_ones, created_by) VALUES ('Test', datetime('now', '+1 day'), 'UTC', 11, 1)`,
			wantErr: true,
		},
		{
			name:    "valid max_plus_ones boundary 0",
			query:   `INSERT INTO events (title, start_time, timezone, max_plus_ones, created_by) VALUES ('Test0', datetime('now', '+1 day'), 'UTC', 0, 1)`,
			wantErr: false,
		},
		{
			name:    "valid max_plus_ones boundary 10",
			query:   `INSERT INTO events (title, start_time, timezone, max_plus_ones, created_by) VALUES ('Test10', datetime('now', '+1 day'), 'UTC', 10, 1)`,
			wantErr: false,
		},
		{
			name:    "invalid rsvp response",
			query:   `INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'hash1', 0, datetime('now', '+30 days')); INSERT INTO rsvps (invite_id, response) VALUES (1, 'invalid')`,
			wantErr: true,
		},
		{
			name:    "valid rsvp response yes",
			query:   `INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'hash2', 0, datetime('now', '+30 days')); INSERT INTO rsvps (invite_id, response) VALUES (2, 'yes')`,
			wantErr: false,
		},
		{
			name:    "valid rsvp response no",
			query:   `INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'hash3', 0, datetime('now', '+30 days')); INSERT INTO rsvps (invite_id, response) VALUES (3, 'no')`,
			wantErr: false,
		},
		{
			name:    "valid rsvp response maybe",
			query:   `INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'hash4', 0, datetime('now', '+30 days')); INSERT INTO rsvps (invite_id, response) VALUES (4, 'maybe')`,
			wantErr: false,
		},
		{
			name:    "invalid invite status",
			query:   `INSERT INTO invites (event_id, token_hash, max_plus_ones, status, expires_at) VALUES (1, 'hash5', 0, 'invalid', datetime('now', '+30 days'))`,
			wantErr: true,
		},
		{
			name:    "invalid email status",
			query:   `INSERT INTO email_queue (to_email, subject, body_text, status) VALUES ('test@example.com', 'Test', 'Body', 'invalid')`,
			wantErr: true,
		},
		{
			name:    "invalid question type",
			query:   `INSERT INTO preference_questions (event_id, question_text, question_type) VALUES (1, 'Test Question', 'invalid')`,
			wantErr: true,
		},
		{
			name:    "invalid template type",
			query:   `INSERT INTO templates (name, type, html_content) VALUES ('Test', 'invalid', '<html></html>')`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec(tt.query)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMigration_UniqueConstraints(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	t.Run("duplicate user email", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO users (email, name, role) VALUES ('test@example.com', 'User 1', 'admin')`)
		if err != nil {
			t.Fatalf("failed to insert first user: %v", err)
		}

		_, err = db.Exec(`INSERT INTO users (email, name, role) VALUES ('test@example.com', 'User 2', 'event_manager')`)
		if err == nil {
			t.Error("expected unique constraint violation for duplicate email, got nil")
		}
	})

	t.Run("duplicate invite token_hash", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO events (title, start_time, timezone, created_by) VALUES ('Event', datetime('now', '+1 day'), 'UTC', 1)`)
		if err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}

		_, err = db.Exec(`INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'unique_hash', 0, datetime('now', '+30 days'))`)
		if err != nil {
			t.Fatalf("failed to insert first invite: %v", err)
		}

		_, err = db.Exec(`INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'unique_hash', 0, datetime('now', '+30 days'))`)
		if err == nil {
			t.Error("expected unique constraint violation for duplicate token_hash, got nil")
		}
	})

	t.Run("duplicate rsvp for invite", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO invites (event_id, token_hash, max_plus_ones, expires_at) VALUES (1, 'hash_for_rsvp', 0, datetime('now', '+30 days'))`)
		if err != nil {
			t.Fatalf("failed to insert invite: %v", err)
		}

		var inviteID int64
		err = db.QueryRow(`SELECT id FROM invites WHERE token_hash = 'hash_for_rsvp'`).Scan(&inviteID)
		if err != nil {
			t.Fatalf("failed to get invite ID: %v", err)
		}

		_, err = db.Exec(`INSERT INTO rsvps (invite_id, response) VALUES (?, 'yes')`, inviteID)
		if err != nil {
			t.Fatalf("failed to insert first rsvp: %v", err)
		}

		_, err = db.Exec(`INSERT INTO rsvps (invite_id, response) VALUES (?, 'no')`, inviteID)
		if err == nil {
			t.Error("expected unique constraint violation for duplicate invite_id in rsvps, got nil")
		}
	})
}

func TestMigration_DownRollback(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users table: %v", err)
	}
	if count != 1 {
		t.Error("users table should exist after migration up")
	}

	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users table: %v", err)
	}
	if count != 0 {
		t.Error("users table should not exist after migration down")
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'sessions', 'events', 'invites', 'rsvps', 'preference_questions', 'rsvp_answers', 'email_queue', 'templates')`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tables after down migration, got %d", count)
	}
}
