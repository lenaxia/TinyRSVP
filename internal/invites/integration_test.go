package invites

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

type testDatabase struct {
	db *sql.DB
}

func (d *testDatabase) DB() *sql.DB {
	return d.db
}

func (d *testDatabase) Close() error {
	return d.db.Close()
}

func (d *testDatabase) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *testDatabase) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *testDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *testDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *testDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func setupTestDB(t *testing.T) db.Database {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	schema := `
	CREATE TABLE events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		description TEXT,
		location TEXT,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP NOT NULL,
		timezone TEXT NOT NULL,
		max_attendees INTEGER,
		require_approval BOOLEAN NOT NULL DEFAULT FALSE,
		allow_plus_ones BOOLEAN NOT NULL DEFAULT FALSE,
		max_plus_ones INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		archived_at TIMESTAMP,
		CHECK (status IN ('draft', 'published', 'archived')),
		CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10)
	);

	CREATE TABLE invites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
		name TEXT,
		email TEXT,
		token_hash TEXT NOT NULL UNIQUE,
		max_plus_ones INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		sent_at TIMESTAMP,
		viewed_at TIMESTAMP,
		unsubscribed BOOLEAN NOT NULL DEFAULT FALSE,
		email_invalid BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		CHECK (status IN ('draft', 'sent', 'viewed', 'responded', 'revoked')),
		CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10)
	);

	CREATE INDEX idx_invites_event_id ON invites(event_id);
	CREATE INDEX idx_invites_token_hash ON invites(token_hash);
	CREATE INDEX idx_invites_email ON invites(email);
	CREATE INDEX idx_invites_status ON invites(status);
	CREATE INDEX idx_invites_expires_at ON invites(expires_at);
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return &testDatabase{db: sqlDB}
}

func createTestEvent(t *testing.T, database db.Database) int64 {
	t.Helper()

	result, err := database.Exec(context.Background(), `
		INSERT INTO events (title, slug, description, location, start_time, end_time, timezone, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "Test Event", "test-event", "Test Description", "Test Location",
		time.Now().Add(7*24*time.Hour), time.Now().Add(8*24*time.Hour), "America/Los_Angeles", "published")

	if err != nil {
		t.Fatalf("failed to create test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get event ID: %v", err)
	}

	return id
}

func TestIntegration_FullInviteWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eventID := createTestEvent(t, db)

	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	repo := repositories.NewInviteRepository(db)
	service := NewInviteService(generator, repo)

	ctx := context.Background()
	email := "test@example.com"
	name := "Test User"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	t.Run("create invite with token generation", func(t *testing.T) {
		invite, plainToken, err := service.CreateInvite(ctx, eventID, &name, &email, 2, expiresAt)
		if err != nil {
			t.Fatalf("CreateInvite() error = %v", err)
		}

		if invite == nil {
			t.Fatal("invite is nil")
		}

		if invite.ID == 0 {
			t.Error("invite ID not set")
		}

		if invite.TokenHash == "" {
			t.Error("token hash not set")
		}

		if len(invite.TokenHash) != 43 {
			t.Errorf("token hash length = %d, want 43", len(invite.TokenHash))
		}

		if plainToken == "" {
			t.Error("plain token not returned")
		}

		if len(plainToken) != 43 {
			t.Errorf("plain token length = %d, want 43", len(plainToken))
		}

		if invite.Status != models.InviteStatusDraft {
			t.Errorf("status = %s, want %s", invite.Status, models.InviteStatusDraft)
		}

		t.Run("retrieve invite by token", func(t *testing.T) {
			retrieved, err := service.GetInviteByToken(ctx, plainToken)
			if err != nil {
				t.Fatalf("GetInviteByToken() error = %v", err)
			}

			if retrieved.ID != invite.ID {
				t.Errorf("retrieved ID = %d, want %d", retrieved.ID, invite.ID)
			}

			if retrieved.TokenHash != invite.TokenHash {
				t.Errorf("retrieved token hash = %s, want %s", retrieved.TokenHash, invite.TokenHash)
			}

			if retrieved.Email == nil || *retrieved.Email != email {
				t.Errorf("retrieved email = %v, want %s", retrieved.Email, email)
			}
		})

		t.Run("retrieve invite by ID", func(t *testing.T) {
			retrieved, err := service.GetInviteByID(ctx, invite.ID)
			if err != nil {
				t.Fatalf("GetInviteByID() error = %v", err)
			}

			if retrieved.ID != invite.ID {
				t.Errorf("retrieved ID = %d, want %d", retrieved.ID, invite.ID)
			}
		})

		t.Run("revoke invite", func(t *testing.T) {
			err := service.RevokeInvite(ctx, invite.ID)
			if err != nil {
				t.Fatalf("RevokeInvite() error = %v", err)
			}

			retrieved, err := service.GetInviteByID(ctx, invite.ID)
			if err != nil {
				t.Fatalf("GetInviteByID() error = %v", err)
			}

			if retrieved.Status != models.InviteStatusRevoked {
				t.Errorf("status = %s, want %s", retrieved.Status, models.InviteStatusRevoked)
			}
		})

		t.Run("cannot revoke already revoked invite", func(t *testing.T) {
			err := service.RevokeInvite(ctx, invite.ID)
			if err == nil {
				t.Error("expected error when revoking already revoked invite")
			}

			if !strings.Contains(err.Error(), "cannot transition from revoked") {
				t.Errorf("error = %v, want error containing 'cannot transition from revoked'", err)
			}
		})
	})
}

func TestIntegration_TokenHashingConsistency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eventID := createTestEvent(t, db)

	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	repo := repositories.NewInviteRepository(db)
	service := NewInviteService(generator, repo)

	ctx := context.Background()
	email := "test@example.com"
	name := "Test User"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	invite, plainToken, err := service.CreateInvite(ctx, eventID, &name, &email, 2, expiresAt)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	t.Run("same token produces same hash", func(t *testing.T) {
		hash1, err := generator.Hash(plainToken)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		hash2, err := generator.Hash(plainToken)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("hash1 = %s, hash2 = %s, want same", hash1, hash2)
		}

		if hash1 != invite.TokenHash {
			t.Errorf("hash = %s, stored hash = %s, want same", hash1, invite.TokenHash)
		}
	})

	t.Run("different token produces different hash", func(t *testing.T) {
		otherToken, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		otherHash, err := generator.Hash(otherToken)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}

		if otherHash == invite.TokenHash {
			t.Error("different tokens produced same hash")
		}
	})

	t.Run("wrong token cannot retrieve invite", func(t *testing.T) {
		wrongToken, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		_, err = service.GetInviteByToken(ctx, wrongToken)
		if err == nil {
			t.Error("expected error when using wrong token")
		}

		if !strings.Contains(err.Error(), "Invite not found") {
			t.Errorf("error = %v, want error containing 'Invite not found'", err)
		}
	})
}

func TestIntegration_MultipleInvites(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eventID := createTestEvent(t, db)

	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	repo := repositories.NewInviteRepository(db)
	service := NewInviteService(generator, repo)

	ctx := context.Background()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	invites := []struct {
		name  string
		email string
	}{
		{"User 1", "user1@example.com"},
		{"User 2", "user2@example.com"},
		{"User 3", "user3@example.com"},
	}

	tokens := make(map[int64]string)

	t.Run("create multiple invites", func(t *testing.T) {
		for _, inv := range invites {
			name := inv.name
			email := inv.email
			invite, plainToken, err := service.CreateInvite(ctx, eventID, &name, &email, 1, expiresAt)
			if err != nil {
				t.Fatalf("CreateInvite() error = %v", err)
			}

			tokens[invite.ID] = plainToken

			if len(invite.TokenHash) != 43 {
				t.Errorf("token hash length = %d, want 43", len(invite.TokenHash))
			}
		}

		if len(tokens) != len(invites) {
			t.Errorf("created %d invites, want %d", len(tokens), len(invites))
		}
	})

	t.Run("each token retrieves correct invite", func(t *testing.T) {
		for inviteID, plainToken := range tokens {
			retrieved, err := service.GetInviteByToken(ctx, plainToken)
			if err != nil {
				t.Fatalf("GetInviteByToken() error = %v", err)
			}

			if retrieved.ID != inviteID {
				t.Errorf("retrieved ID = %d, want %d", retrieved.ID, inviteID)
			}
		}
	})

	t.Run("list invites by event", func(t *testing.T) {
		filters := repositories.InviteFilters{
			Limit: 100,
		}

		list, err := service.ListInvitesByEventID(ctx, eventID, filters)
		if err != nil {
			t.Fatalf("ListInvitesByEventID() error = %v", err)
		}

		if len(list) != len(invites) {
			t.Errorf("listed %d invites, want %d", len(list), len(invites))
		}
	})
}

func TestIntegration_EmailValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eventID := createTestEvent(t, db)

	secret := []byte("test-secret-key-32-bytes-long!")
	generator := token.NewGenerator(secret)
	repo := repositories.NewInviteRepository(db)
	service := NewInviteService(generator, repo)

	ctx := context.Background()
	name := "Test User"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "test@mail.example.com", false},
		{"valid email with plus", "test+tag@example.com", false},
		{"invalid email no @", "notanemail", true},
		{"invalid email no domain", "user@", true},
		{"invalid email no local", "@example.com", true},
		{"invalid email spaces", "user name@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := tt.email
			_, _, err := service.CreateInvite(ctx, eventID, &name, &email, 2, expiresAt)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateInvite() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "email must be a valid email address") {
					t.Errorf("error = %v, want error containing 'email must be a valid email address'", err)
				}
			}
		})
	}
}
