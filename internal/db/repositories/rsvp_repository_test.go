package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupRSVPTestDB(t *testing.T) db.Database {
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

	return database
}


func createTestEventForRSVP(t *testing.T, db db.Database) int64 {
	t.Helper()

	ctx := context.Background()
	result, err := db.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Test Event', ?, 'America/Los_Angeles', 'published', 1, 2)
	`, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get event ID: %v", err)
	}

	return id
}

func createTestInviteForRSVP(t *testing.T, db db.Database, eventID int64, maxPlusOnes int) int64 {
	t.Helper()

	ctx := context.Background()
	tokenHash := fmt.Sprintf("test_token_hash_%d_%d", eventID, time.Now().UnixNano())
	result, err := db.Exec(ctx, `
		INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
		VALUES (?, 'Test Guest', 'guest@example.com', ?, ?, 'sent', ?)
	`, eventID, tokenHash, maxPlusOnes, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get invite ID: %v", err)
	}

	return id
}

func TestNewRSVPRepository(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	if repo == nil {
		t.Fatal("NewRSVPRepository returned nil")
	}
}

func TestRSVPRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		rsvp    *models.RSVP
		wantErr bool
		errType string
	}{
		{
			name: "valid yes response with plus ones",
			rsvp: &models.RSVP{
				Response: models.RSVPResponseYes,
				PlusOnes: 2,
			},
			wantErr: false,
		},
		{
			name: "valid no response",
			rsvp: &models.RSVP{
				Response: models.RSVPResponseNo,
				PlusOnes: 0,
			},
			wantErr: false,
		},
		{
			name: "valid maybe response",
			rsvp: &models.RSVP{
				Response: models.RSVPResponseMaybe,
				PlusOnes: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := setupRSVPTestDB(t)
			defer database.Close()

			repo := NewRSVPRepository(database)
			eventID := createTestEventForRSVP(t, database)
			inviteID := createTestInviteForRSVP(t, database, eventID, 2)

			tt.rsvp.InviteID = inviteID

			ctx := context.Background()
			err := repo.Create(ctx, tt.rsvp)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.rsvp.ID == 0 {
					t.Error("Expected ID to be set after creation")
				}
				if tt.rsvp.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.rsvp.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestRSVPRepository_Create_DuplicateInvite(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)
	inviteID := createTestInviteForRSVP(t, database, eventID, 2)

	ctx := context.Background()

	first := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}

	err := repo.Create(ctx, first)
	if err != nil {
		t.Fatalf("First Create() failed: %v", err)
	}

	duplicate := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseNo,
		PlusOnes: 0,
	}

	err = repo.Create(ctx, duplicate)
	if err == nil {
		t.Error("Expected error for duplicate invite_id, got nil")
	}
}

func TestRSVPRepository_GetByID(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)
	inviteID := createTestInviteForRSVP(t, database, eventID, 2)

	ctx := context.Background()

	created := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseYes,
		PlusOnes: 2,
	}

	err := repo.Create(ctx, created)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("ID = %d, want %d", retrieved.ID, created.ID)
	}
	if retrieved.InviteID != created.InviteID {
		t.Errorf("InviteID = %d, want %d", retrieved.InviteID, created.InviteID)
	}
	if retrieved.Response != created.Response {
		t.Errorf("Response = %s, want %s", retrieved.Response, created.Response)
	}
	if retrieved.PlusOnes != created.PlusOnes {
		t.Errorf("PlusOnes = %d, want %d", retrieved.PlusOnes, created.PlusOnes)
	}
}

func TestRSVPRepository_GetByID_NotFound(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	var notFoundErr *models.NotFoundError
	if err == nil {
		t.Error("GetByID() expected error, got nil")
	} else if !errors.As(err, &notFoundErr) {
		t.Errorf("GetByID() error = %v, want NotFoundError", err)
	}
}

func TestRSVPRepository_GetByInviteID(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)
	inviteID := createTestInviteForRSVP(t, database, eventID, 2)

	ctx := context.Background()

	created := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseMaybe,
		PlusOnes: 1,
	}

	err := repo.Create(ctx, created)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	retrieved, err := repo.GetByInviteID(ctx, inviteID)
	if err != nil {
		t.Fatalf("GetByInviteID() error = %v", err)
	}

	if retrieved.InviteID != inviteID {
		t.Errorf("InviteID = %d, want %d", retrieved.InviteID, inviteID)
	}
	if retrieved.Response != models.RSVPResponseMaybe {
		t.Errorf("Response = %s, want %s", retrieved.Response, models.RSVPResponseMaybe)
	}
}

func TestRSVPRepository_GetByInviteID_NotFound(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	ctx := context.Background()

	_, err := repo.GetByInviteID(ctx, 99999)
	var notFoundErr *models.NotFoundError
	if err == nil {
		t.Error("GetByInviteID() expected error, got nil")
	} else if !errors.As(err, &notFoundErr) {
		t.Errorf("GetByInviteID() error = %v, want NotFoundError", err)
	}
}

func TestRSVPRepository_GetByEventID(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)

	invite1 := createTestInviteForRSVP(t, database, eventID, 2)
	invite2 := createTestInviteForRSVP(t, database, eventID, 1)
	invite3 := createTestInviteForRSVP(t, database, eventID, 0)

	ctx := context.Background()

	rsvps := []*models.RSVP{
		{InviteID: invite1, Response: models.RSVPResponseYes, PlusOnes: 2},
		{InviteID: invite2, Response: models.RSVPResponseNo, PlusOnes: 0},
		{InviteID: invite3, Response: models.RSVPResponseMaybe, PlusOnes: 0},
	}

	for _, rsvp := range rsvps {
		if err := repo.Create(ctx, rsvp); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}

	retrieved, err := repo.GetByEventID(ctx, eventID)
	if err != nil {
		t.Fatalf("GetByEventID() error = %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("GetByEventID() returned %d RSVPs, want 3", len(retrieved))
	}
}

func TestRSVPRepository_Update(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)
	inviteID := createTestInviteForRSVP(t, database, eventID, 2)

	ctx := context.Background()

	original := &models.RSVP{
		InviteID: inviteID,
		Response: models.RSVPResponseMaybe,
		PlusOnes: 1,
	}

	err := repo.Create(ctx, original)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	original.Response = models.RSVPResponseYes
	original.PlusOnes = 2

	err = repo.Update(ctx, original)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByID(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updated.Response != models.RSVPResponseYes {
		t.Errorf("Response = %s, want %s", updated.Response, models.RSVPResponseYes)
	}
	if updated.PlusOnes != 2 {
		t.Errorf("PlusOnes = %d, want 2", updated.PlusOnes)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestRSVPRepository_GetStats(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)

	invites := make([]int64, 10)
	for i := 0; i < 10; i++ {
		invites[i] = createTestInviteForRSVP(t, database, eventID, 2)
	}

	ctx := context.Background()

	rsvps := []struct {
		inviteIdx int
		response  models.RSVPResponse
		plusOnes  int
	}{
		{0, models.RSVPResponseYes, 2},
		{1, models.RSVPResponseYes, 1},
		{2, models.RSVPResponseYes, 0},
		{3, models.RSVPResponseNo, 0},
		{4, models.RSVPResponseNo, 0},
		{5, models.RSVPResponseMaybe, 1},
	}

	for _, r := range rsvps {
		rsvp := &models.RSVP{
			InviteID: invites[r.inviteIdx],
			Response: r.response,
			PlusOnes: r.plusOnes,
		}
		if err := repo.Create(ctx, rsvp); err != nil {
			t.Fatalf("Create() failed: %v", err)
		}
	}

	stats, err := repo.GetStats(ctx, eventID)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalInvites != 10 {
		t.Errorf("TotalInvites = %d, want 10", stats.TotalInvites)
	}
	if stats.YesCount != 3 {
		t.Errorf("YesCount = %d, want 3", stats.YesCount)
	}
	if stats.NoCount != 2 {
		t.Errorf("NoCount = %d, want 2", stats.NoCount)
	}
	if stats.MaybeCount != 1 {
		t.Errorf("MaybeCount = %d, want 1", stats.MaybeCount)
	}
	if stats.NoResponse != 4 {
		t.Errorf("NoResponse = %d, want 4", stats.NoResponse)
	}
	if stats.TotalGuests != 6 {
		t.Errorf("TotalGuests = %d, want 6 (3 yes RSVPs: 1+2, 1+1, 1+0)", stats.TotalGuests)
	}
}

func TestRSVPRepository_GetStats_EmptyEvent(t *testing.T) {
	database := setupRSVPTestDB(t)
	defer database.Close()

	repo := NewRSVPRepository(database)
	eventID := createTestEventForRSVP(t, database)

	ctx := context.Background()
	stats, err := repo.GetStats(ctx, eventID)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalInvites != 0 {
		t.Errorf("TotalInvites = %d, want 0", stats.TotalInvites)
	}
	if stats.YesCount != 0 {
		t.Errorf("YesCount = %d, want 0", stats.YesCount)
	}
	if stats.TotalGuests != 0 {
		t.Errorf("TotalGuests = %d, want 0", stats.TotalGuests)
	}
}
