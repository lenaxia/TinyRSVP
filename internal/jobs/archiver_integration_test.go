package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupIntegrationTestDB(t *testing.T) db.Database {
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

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (1, 'admin@example.com', 'Admin User', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return database
}

type mockAuthz struct{}

func (m *mockAuthz) IsAdmin(user *models.User) bool                                  { return true }
func (m *mockAuthz) IsEventManager(user *models.User) bool                           { return true }
func (m *mockAuthz) CanCreateEvent(ctx context.Context, user *models.User) bool      { return true }
func (m *mockAuthz) CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	return true
}
func (m *mockAuthz) CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	return true
}
func (m *mockAuthz) CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	return true
}
func (m *mockAuthz) CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool {
	return true
}
func (m *mockAuthz) CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool {
	return true
}
func (m *mockAuthz) CanManageUsers(ctx context.Context, user *models.User) bool { return true }
func (m *mockAuthz) CanConfigureSystem(ctx context.Context, user *models.User) bool {
	return true
}

func TestEventArchiver_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewEventRepository(database)
	tzValidator := events.NewTimezoneValidator()
	validator := events.NewValidator(tzValidator)
	authz := &mockAuthz{}
	service := events.NewService(repo, validator, authz)
	archiver := NewEventArchiver(service, 30)

	adminUser := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	ctx := auth.WithUser(context.Background(), adminUser)

	now := time.Now()

	oldPublishedEvent := &models.Event{
		Title:       "Old Published Event",
		StartTime:   now.Add(-40 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
		Version:     1,
	}

	oldCancelledEvent := &models.Event{
		Title:       "Old Cancelled Event",
		StartTime:   now.Add(-35 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
		Version:     1,
	}

	recentEvent := &models.Event{
		Title:       "Recent Event",
		StartTime:   now.Add(-10 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
		Version:     1,
	}

	oldDraftEvent := &models.Event{
		Title:       "Old Draft Event",
		StartTime:   now.Add(-40 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
		Version:     1,
	}

	if err := repo.Create(ctx, oldPublishedEvent); err != nil {
		t.Fatalf("Failed to create old published event: %v", err)
	}
	if err := repo.UpdateStatus(ctx, oldPublishedEvent.ID, models.EventStatusPublished); err != nil {
		t.Fatalf("Failed to publish old event: %v", err)
	}

	if err := repo.Create(ctx, oldCancelledEvent); err != nil {
		t.Fatalf("Failed to create old cancelled event: %v", err)
	}
	if err := repo.UpdateStatus(ctx, oldCancelledEvent.ID, models.EventStatusCancelled); err != nil {
		t.Fatalf("Failed to cancel old event: %v", err)
	}

	if err := repo.Create(ctx, recentEvent); err != nil {
		t.Fatalf("Failed to create recent event: %v", err)
	}
	if err := repo.UpdateStatus(ctx, recentEvent.ID, models.EventStatusPublished); err != nil {
		t.Fatalf("Failed to publish recent event: %v", err)
	}

	if err := repo.Create(ctx, oldDraftEvent); err != nil {
		t.Fatalf("Failed to create old draft event: %v", err)
	}

	if err := archiver.Run(ctx); err != nil {
		t.Fatalf("Archiver run failed: %v", err)
	}

	updatedOldPublished, err := repo.GetByID(ctx, oldPublishedEvent.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve old published event: %v", err)
	}
	if updatedOldPublished.Status != models.EventStatusArchived {
		t.Errorf("Old published event status = %v, want %v", updatedOldPublished.Status, models.EventStatusArchived)
	}

	updatedOldCancelled, err := repo.GetByID(ctx, oldCancelledEvent.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve old cancelled event: %v", err)
	}
	if updatedOldCancelled.Status != models.EventStatusArchived {
		t.Errorf("Old cancelled event status = %v, want %v", updatedOldCancelled.Status, models.EventStatusArchived)
	}

	updatedRecent, err := repo.GetByID(ctx, recentEvent.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve recent event: %v", err)
	}
	if updatedRecent.Status != models.EventStatusPublished {
		t.Errorf("Recent event status = %v, want %v", updatedRecent.Status, models.EventStatusPublished)
	}

	updatedOldDraft, err := repo.GetByID(ctx, oldDraftEvent.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve old draft event: %v", err)
	}
	if updatedOldDraft.Status != models.EventStatusDraft {
		t.Errorf("Old draft event status = %v, want %v", updatedOldDraft.Status, models.EventStatusDraft)
	}

	if err := archiver.Run(ctx); err != nil {
		t.Fatalf("Second archiver run failed: %v", err)
	}
}
