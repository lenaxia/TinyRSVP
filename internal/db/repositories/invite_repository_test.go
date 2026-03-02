package repositories

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupInviteTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (1, 'test@example.com', 'Test User', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	_, err = database.Exec(ctx, `
		INSERT INTO events (id, title, start_time, timezone, status, created_by, max_plus_ones, created_at, updated_at)
		VALUES (1, 'Test Event', datetime('now', '+7 days'), 'America/Los_Angeles', 'draft', 1, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	return database
}

func inviteStatusPtr(s models.InviteStatus) *models.InviteStatus {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestNewInviteRepository(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)
	if repo == nil {
		t.Fatal("NewInviteRepository returned nil")
	}
}

func TestInviteRepository_Create(t *testing.T) {
	validTokenHash := strings.Repeat("a", 43)
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email := "guest@example.com"
	name := "Guest User"

	tests := []struct {
		name    string
		invite  *models.Invite
		wantErr bool
		errType string
	}{
		{
			name: "valid invite with required fields",
			invite: &models.Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid invite with all fields",
			invite: &models.Invite{
				EventID:     1,
				Name:        &name,
				Email:       &email,
				TokenHash:   strings.Repeat("b", 43),
				MaxPlusOnes: 1,
				Status:      models.InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "invalid - duplicate token hash",
			invite: &models.Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errType: "ConflictError",
		},
		{
			name: "invalid - missing event_id",
			invite: &models.Invite{
				EventID:     0,
				TokenHash:   strings.Repeat("c", 43),
				MaxPlusOnes: 2,
				Status:      models.InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "invalid - non-existent event",
			invite: &models.Invite{
				EventID:     999,
				TokenHash:   strings.Repeat("d", 43),
				MaxPlusOnes: 2,
				Status:      models.InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
		},
	}

	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.invite)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "ValidationError":
					if _, ok := err.(*models.ValidationError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				case "ConflictError":
					if _, ok := err.(*models.ConflictError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr {
				if tt.invite.ID == 0 {
					t.Error("Expected invite ID to be set after creation")
				}
				if tt.invite.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.invite.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestInviteRepository_CreateBatch(t *testing.T) {
	futureTime := time.Now().Add(30 * 24 * time.Hour)

	tests := []struct {
		name    string
		invites []*models.Invite
		wantErr bool
		errType string
	}{
		{
			name: "successful batch create",
			invites: []*models.Invite{
				{
					EventID:     1,
					TokenHash:   strings.Repeat("a", 43),
					MaxPlusOnes: 2,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   futureTime,
				},
				{
					EventID:     1,
					TokenHash:   strings.Repeat("b", 43),
					MaxPlusOnes: 1,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   futureTime,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty batch",
			invites: []*models.Invite{},
			wantErr: false,
		},
		{
			name: "batch with duplicate token hash - should rollback",
			invites: []*models.Invite{
				{
					EventID:     1,
					TokenHash:   strings.Repeat("c", 43),
					MaxPlusOnes: 2,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   futureTime,
				},
				{
					EventID:     1,
					TokenHash:   strings.Repeat("c", 43),
					MaxPlusOnes: 1,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   futureTime,
				},
			},
			wantErr: true,
			errType: "ConflictError",
		},
		{
			name: "batch with invalid invite - should rollback",
			invites: []*models.Invite{
				{
					EventID:     1,
					TokenHash:   strings.Repeat("d", 43),
					MaxPlusOnes: 2,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   futureTime,
				},
				{
					EventID:     0,
					TokenHash:   strings.Repeat("e", 43),
					MaxPlusOnes: 1,
					Status:      models.InviteStatusDraft,
					ExpiresAt:   futureTime,
				},
			},
			wantErr: true,
			errType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := setupInviteTestDB(t)
			defer database.Close()

			repo := NewInviteRepository(database)

			err := repo.CreateBatch(context.Background(), tt.invites)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "ValidationError":
					if _, ok := err.(*models.ValidationError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				case "ConflictError":
					if _, ok := err.(*models.ConflictError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr && len(tt.invites) > 0 {
				for _, invite := range tt.invites {
					if invite.ID == 0 {
						t.Error("Expected invite ID to be set after batch creation")
					}
				}

				count, err := repo.CountByEventID(context.Background(), 1)
				if err != nil {
					t.Fatalf("Failed to count invites: %v", err)
				}
				if count != len(tt.invites) {
					t.Errorf("Expected %d invites, got %d", len(tt.invites), count)
				}
			}
		})
	}
}

func TestInviteRepository_CreateBatch_TooLarge(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	invites := make([]*models.Invite, 501)
	for i := range invites {
		invites[i] = &models.Invite{
			EventID:     1,
			TokenHash:   strings.Repeat("a", 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		}
	}

	err := repo.CreateBatch(context.Background(), invites)
	if err == nil {
		t.Fatal("Expected error for batch size > 500")
	}

	if _, ok := err.(*models.ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestInviteRepository_GetByID(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	invite := &models.Invite{
		EventID:     1,
		TokenHash:   strings.Repeat("a", 43),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}

	if err := repo.Create(context.Background(), invite); err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType string
	}{
		{
			name:    "existing invite",
			id:      invite.ID,
			wantErr: false,
		},
		{
			name:    "non-existent invite",
			id:      999,
			wantErr: true,
			errType: "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				if _, ok := err.(*models.NotFoundError); !ok {
					t.Errorf("Expected error type %s, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if result.ID != tt.id {
					t.Errorf("GetByID() ID = %d, want %d", result.ID, tt.id)
				}
				if result.TokenHash != invite.TokenHash {
					t.Errorf("GetByID() TokenHash = %s, want %s", result.TokenHash, invite.TokenHash)
				}
			}
		})
	}
}

func TestInviteRepository_GetByTokenHash(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	tokenHash := strings.Repeat("a", 43)
	invite := &models.Invite{
		EventID:     1,
		TokenHash:   tokenHash,
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}

	if err := repo.Create(context.Background(), invite); err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	tests := []struct {
		name      string
		tokenHash string
		wantErr   bool
		errType   string
	}{
		{
			name:      "existing token hash",
			tokenHash: tokenHash,
			wantErr:   false,
		},
		{
			name:      "non-existent token hash",
			tokenHash: strings.Repeat("z", 43),
			wantErr:   true,
			errType:   "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByTokenHash(context.Background(), tt.tokenHash)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByTokenHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				if _, ok := err.(*models.NotFoundError); !ok {
					t.Errorf("Expected error type %s, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if result.TokenHash != tt.tokenHash {
					t.Errorf("GetByTokenHash() TokenHash = %s, want %s", result.TokenHash, tt.tokenHash)
				}
			}
		})
	}
}

func TestInviteRepository_Update(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	email := "original@example.com"
	invite := &models.Invite{
		EventID:     1,
		Email:       &email,
		TokenHash:   strings.Repeat("a", 43),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}

	if err := repo.Create(context.Background(), invite); err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	tests := []struct {
		name         string
		updateInvite func() *models.Invite
		wantErr      bool
		errType      string
	}{
		{
			name: "successful update",
			updateInvite: func() *models.Invite {
				i := *invite
				newEmail := "updated@example.com"
				i.Email = &newEmail
				i.Status = models.InviteStatusSent
				now := time.Now()
				i.SentAt = &now
				return &i
			},
			wantErr: false,
		},
		{
			name: "update non-existent invite",
			updateInvite: func() *models.Invite {
				i := *invite
				i.ID = 999
				return &i
			},
			wantErr: true,
			errType: "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateInvite := tt.updateInvite()
			err := repo.Update(context.Background(), updateInvite)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				if _, ok := err.(*models.NotFoundError); !ok {
					t.Errorf("Expected error type %s, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				retrieved, err := repo.GetByID(context.Background(), updateInvite.ID)
				if err != nil {
					t.Fatalf("Failed to retrieve updated invite: %v", err)
				}

				if retrieved.Status != updateInvite.Status {
					t.Errorf("Status = %s, want %s", retrieved.Status, updateInvite.Status)
				}
			}
		})
	}
}

func TestInviteRepository_Delete(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	invite := &models.Invite{
		EventID:     1,
		TokenHash:   strings.Repeat("a", 43),
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}

	if err := repo.Create(context.Background(), invite); err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType string
	}{
		{
			name:    "delete existing invite",
			id:      invite.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent invite",
			id:      999,
			wantErr: true,
			errType: "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				if _, ok := err.(*models.NotFoundError); !ok {
					t.Errorf("Expected error type %s, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				_, err := repo.GetByID(context.Background(), tt.id)
				if err == nil {
					t.Error("Expected invite to be deleted")
				}
			}
		})
	}
}

func TestInviteRepository_ListByEventID(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email1 := "user1@example.com"
	invites := []*models.Invite{
		{
			EventID:      1,
			TokenHash:    strings.Repeat("a", 43),
			MaxPlusOnes:  2,
			Status:       models.InviteStatusDraft,
			ExpiresAt:    futureTime,
			Unsubscribed: false,
			EmailInvalid: false,
		},
		{
			EventID:      1,
			Email:        &email1,
			TokenHash:    strings.Repeat("b", 43),
			MaxPlusOnes:  1,
			Status:       models.InviteStatusSent,
			ExpiresAt:    futureTime,
			Unsubscribed: false,
			EmailInvalid: false,
		},
		{
			EventID:      1,
			TokenHash:    strings.Repeat("c", 43),
			MaxPlusOnes:  0,
			Status:       models.InviteStatusDraft,
			ExpiresAt:    futureTime,
			Unsubscribed: true,
			EmailInvalid: false,
		},
	}

	for _, inv := range invites {
		if err := repo.Create(context.Background(), inv); err != nil {
			t.Fatalf("Failed to create invite: %v", err)
		}
	}

	tests := []struct {
		name      string
		eventID   int64
		filters   InviteFilters
		wantCount int
	}{
		{
			name:      "no filters",
			eventID:   1,
			filters:   InviteFilters{},
			wantCount: 3,
		},
		{
			name:    "filter by status",
			eventID: 1,
			filters: InviteFilters{
				Status: inviteStatusPtr(models.InviteStatusDraft),
			},
			wantCount: 2,
		},
		{
			name:    "filter by unsubscribed",
			eventID: 1,
			filters: InviteFilters{
				Unsubscribed: boolPtr(true),
			},
			wantCount: 1,
		},
		{
			name:    "with pagination",
			eventID: 1,
			filters: InviteFilters{
				Limit:  2,
				Offset: 0,
			},
			wantCount: 2,
		},
		{
			name:    "with pagination offset",
			eventID: 1,
			filters: InviteFilters{
				Limit:  2,
				Offset: 2,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.ListByEventID(context.Background(), tt.eventID, tt.filters)
			if err != nil {
				t.Errorf("ListByEventID() error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("ListByEventID() returned %d invites, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestInviteRepository_CountByEventID(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	for i := 0; i < 5; i++ {
		invite := &models.Invite{
			EventID:     1,
			TokenHash:   strings.Repeat(string(rune('a'+i)), 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		}
		if err := repo.Create(context.Background(), invite); err != nil {
			t.Fatalf("Failed to create invite: %v", err)
		}
	}

	count, err := repo.CountByEventID(context.Background(), 1)
	if err != nil {
		t.Fatalf("CountByEventID() error = %v", err)
	}

	if count != 5 {
		t.Errorf("CountByEventID() = %d, want 5", count)
	}
}

func TestInviteRepository_GetStats(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	statuses := []models.InviteStatus{
		models.InviteStatusDraft,
		models.InviteStatusDraft,
		models.InviteStatusSent,
		models.InviteStatusViewed,
		models.InviteStatusResponded,
		models.InviteStatusRevoked,
	}

	for i, status := range statuses {
		email := "user" + string(rune('a'+i)) + "@example.com"
		invite := &models.Invite{
			EventID:     1,
			TokenHash:   strings.Repeat(string(rune('a'+i)), 43),
			MaxPlusOnes: 2,
			Status:      status,
			ExpiresAt:   futureTime,
		}
		if status == models.InviteStatusSent {
			invite.Email = &email
		}
		if err := repo.Create(context.Background(), invite); err != nil {
			t.Fatalf("Failed to create invite: %v", err)
		}
	}

	stats, err := repo.GetStats(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.Total != 6 {
		t.Errorf("Total = %d, want 6", stats.Total)
	}
	if stats.Draft != 2 {
		t.Errorf("Draft = %d, want 2", stats.Draft)
	}
	if stats.Sent != 1 {
		t.Errorf("Sent = %d, want 1", stats.Sent)
	}
	if stats.Viewed != 1 {
		t.Errorf("Viewed = %d, want 1", stats.Viewed)
	}
	if stats.Responded != 1 {
		t.Errorf("Responded = %d, want 1", stats.Responded)
	}
	if stats.Revoked != 1 {
		t.Errorf("Revoked = %d, want 1", stats.Revoked)
	}
}

func TestInviteRepository_FindDuplicateEmails(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	existingEmails := []string{"user1@example.com", "user2@example.com"}

	for i, email := range existingEmails {
		e := email
		invite := &models.Invite{
			EventID:     1,
			Email:       &e,
			TokenHash:   strings.Repeat(string(rune('a'+i)), 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		}
		if err := repo.Create(context.Background(), invite); err != nil {
			t.Fatalf("Failed to create invite: %v", err)
		}
	}

	tests := []struct {
		name      string
		eventID   int64
		emails    []string
		wantCount int
	}{
		{
			name:      "no duplicates",
			eventID:   1,
			emails:    []string{"new1@example.com", "new2@example.com"},
			wantCount: 0,
		},
		{
			name:      "one duplicate",
			eventID:   1,
			emails:    []string{"user1@example.com", "new@example.com"},
			wantCount: 1,
		},
		{
			name:      "all duplicates",
			eventID:   1,
			emails:    []string{"user1@example.com", "user2@example.com"},
			wantCount: 2,
		},
		{
			name:      "empty list",
			eventID:   1,
			emails:    []string{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duplicates, err := repo.FindDuplicateEmails(context.Background(), tt.eventID, tt.emails)
			if err != nil {
				t.Errorf("FindDuplicateEmails() error = %v", err)
				return
			}

			if len(duplicates) != tt.wantCount {
				t.Errorf("FindDuplicateEmails() returned %d duplicates, want %d", len(duplicates), tt.wantCount)
			}
		})
	}
}

func TestInviteRepository_DeleteExpired(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	now := time.Now()
	ctx := context.Background()

	_, err := database.Exec(ctx, `
		INSERT INTO invites (event_id, token_hash, max_plus_ones, status, expires_at, created_at, updated_at)
		VALUES
			(1, ?, 2, 'draft', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			(1, ?, 2, 'draft', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			(1, ?, 2, 'draft', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		strings.Repeat("a", 43), now.Add(-24*time.Hour),
		strings.Repeat("b", 43), now.Add(-48*time.Hour),
		strings.Repeat("c", 43), now.Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("Failed to insert test invites: %v", err)
	}

	deleted, err := repo.DeleteExpired(context.Background(), now)
	if err != nil {
		t.Fatalf("DeleteExpired() error = %v", err)
	}

	if deleted != 2 {
		t.Errorf("DeleteExpired() deleted %d invites, want 2", deleted)
	}

	remaining, err := repo.CountByEventID(context.Background(), 1)
	if err != nil {
		t.Fatalf("CountByEventID() error = %v", err)
	}

	if remaining != 1 {
		t.Errorf("CountByEventID() = %d, want 1", remaining)
	}
}

func TestInviteRepository_CountInvites(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email1 := "test1@example.com"
	name1 := "Test User 1"
	email2 := "test2@example.com"
	name2 := "Test User 2"
	email3 := "test3@example.com"
	name3 := "Test User 3"

	invites := []*models.Invite{
		{
			EventID:     1,
			Email:       &email1,
			Name:        &name1,
			TokenHash:   strings.Repeat("a", 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   futureTime,
		},
		{
			EventID:     1,
			Email:       &email2,
			Name:        &name2,
			TokenHash:   strings.Repeat("b", 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusSent,
			ExpiresAt:   futureTime,
		},
		{
			EventID:     1,
			Email:       &email3,
			Name:        &name3,
			TokenHash:   strings.Repeat("c", 43),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusRevoked,
			ExpiresAt:   futureTime,
		},
	}

	for _, inv := range invites {
		if err := repo.Create(context.Background(), inv); err != nil {
			t.Fatalf("Failed to create invite: %v", err)
		}
	}

	count, err := repo.CountInvites(context.Background())
	if err != nil {
		t.Fatalf("CountInvites() error = %v", err)
	}

	if count != 3 {
		t.Errorf("CountInvites() = %d, want 3", count)
	}
}

func TestInviteRepository_CountInvites_Empty(t *testing.T) {
	database := setupInviteTestDB(t)
	defer database.Close()

	repo := NewInviteRepository(database)

	count, err := repo.CountInvites(context.Background())
	if err != nil {
		t.Fatalf("CountInvites() error = %v", err)
	}

	if count != 0 {
		t.Errorf("CountInvites() = %d, want 0", count)
	}
}
