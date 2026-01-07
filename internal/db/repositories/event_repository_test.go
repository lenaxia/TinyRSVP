package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupEventTestDB(t *testing.T) db.Database {
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

func timePtr(t time.Time) *time.Time {
	return &t
}

func int64Ptr(i int64) *int64 {
	return &i
}

func statusPtr(s models.EventStatus) *models.EventStatus {
	return &s
}

func TestNewEventRepository(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	if repo == nil {
		t.Fatal("NewEventRepository returned nil")
	}
}

func TestEventRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		event   *models.Event
		wantErr bool
		errType string
	}{
		{
			name: "valid event with required fields only",
			event: &models.Event{
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				Status:      models.EventStatusDraft,
				CreatedBy:   1,
				MaxPlusOnes: 0,
			},
			wantErr: false,
		},
		{
			name: "valid event with all fields",
			event: &models.Event{
				Title:        "Complete Event",
				Description:  stringPtr("Full description"),
				StartTime:    time.Now().Add(24 * time.Hour),
				EndTime:      timePtr(time.Now().Add(26 * time.Hour)),
				Timezone:     "America/Los_Angeles",
				Location:     stringPtr("123 Main St"),
				Status:       models.EventStatusDraft,
				CreatedBy:    1,
				MaxPlusOnes:  2,
				RSVPDeadline: timePtr(time.Now().Add(12 * time.Hour)),
			},
			wantErr: false,
		},
		{
			name: "missing title",
			event: &models.Event{
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				Status:      models.EventStatusDraft,
				CreatedBy:   1,
				MaxPlusOnes: 0,
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "missing timezone",
			event: &models.Event{
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Status:      models.EventStatusDraft,
				CreatedBy:   1,
				MaxPlusOnes: 0,
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "invalid creator (foreign key)",
			event: &models.Event{
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				Status:      models.EventStatusDraft,
				CreatedBy:   999,
				MaxPlusOnes: 0,
			},
			wantErr: true,
		},
	}

	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.event)

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
				}
			}

			if !tt.wantErr {
				if tt.event.ID == 0 {
					t.Error("Expected event ID to be set after creation")
				}
				if tt.event.Version != 1 {
					t.Errorf("Expected version to be 1, got %d", tt.event.Version)
				}
				if tt.event.ICSSequence != 0 {
					t.Errorf("Expected ICSSequence to be 0, got %d", tt.event.ICSSequence)
				}
				if tt.event.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.event.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestEventRepository_GetByID(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType string
	}{
		{
			name:    "existing event",
			id:      event.ID,
			wantErr: false,
		},
		{
			name:    "non-existent event",
			id:      999,
			wantErr: true,
			errType: "NotFoundError",
		},
		{
			name:    "invalid ID",
			id:      0,
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
				switch tt.errType {
				case "NotFoundError":
					if _, ok := err.(*models.NotFoundError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr {
				if result.ID != tt.id {
					t.Errorf("GetByID() ID = %d, want %d", result.ID, tt.id)
				}
				if result.Title != event.Title {
					t.Errorf("GetByID() Title = %s, want %s", result.Title, event.Title)
				}
			}
		})
	}
}

func TestEventRepository_Update(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Original Title",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name        string
		updateEvent func() *models.Event
		wantErr     bool
		errType     string
	}{
		{
			name: "successful update",
			updateEvent: func() *models.Event {
				e := *event
				e.Title = "Updated Title"
				e.Description = stringPtr("New description")
				return &e
			},
			wantErr: false,
		},
		{
			name: "update non-existent event",
			updateEvent: func() *models.Event {
				e := *event
				e.ID = 999
				e.Title = "Should Fail"
				return &e
			},
			wantErr: true,
			errType: "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateEvent := tt.updateEvent()
			err := repo.Update(context.Background(), updateEvent)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "NotFoundError":
					if _, ok := err.(*models.NotFoundError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr {
				retrieved, err := repo.GetByID(context.Background(), updateEvent.ID)
				if err != nil {
					t.Fatalf("Failed to retrieve updated event: %v", err)
				}

				if retrieved.Title != updateEvent.Title {
					t.Errorf("Title = %q, want %q", retrieved.Title, updateEvent.Title)
				}
			}
		})
	}
}

func TestEventRepository_UpdateWithVersion(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Original Title",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name            string
		updateTitle     string
		expectedVersion int
		wantErr         bool
		errType         string
	}{
		{
			name:            "successful update",
			updateTitle:     "Updated Title",
			expectedVersion: 1,
			wantErr:         false,
		},
		{
			name:            "version conflict",
			updateTitle:     "Another Update",
			expectedVersion: 1,
			wantErr:         true,
			errType:         "OptimisticLockError",
		},
		{
			name:            "successful second update",
			updateTitle:     "Third Title",
			expectedVersion: 2,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event.Title = tt.updateTitle
			err := repo.UpdateWithVersion(context.Background(), event, tt.expectedVersion)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateWithVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "OptimisticLockError":
					if _, ok := err.(*models.OptimisticLockError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr {
				retrieved, err := repo.GetByID(context.Background(), event.ID)
				if err != nil {
					t.Fatalf("Failed to retrieve updated event: %v", err)
				}

				if retrieved.Title != tt.updateTitle {
					t.Errorf("Title = %q, want %q", retrieved.Title, tt.updateTitle)
				}

				if retrieved.Version != tt.expectedVersion+1 {
					t.Errorf("Version = %d, want %d", retrieved.Version, tt.expectedVersion+1)
				}
			}
		})
	}
}

func TestEventRepository_UpdateWithVersion_NonExistentEvent(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	nonExistentEvent := &models.Event{
		ID:          999,
		Title:       "Non-existent Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	err := repo.UpdateWithVersion(context.Background(), nonExistentEvent, 1)

	if err == nil {
		t.Fatal("UpdateWithVersion() expected error for non-existent event, got nil")
	}

	if _, ok := err.(*models.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError for non-existent event, got %T: %v", err, err)
	}
}


func TestEventRepository_Create_LastInsertIdError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	err := repo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	if event.ID == 0 {
		t.Error("Expected event ID to be set")
	}
}

func TestEventRepository_Update_RowsAffectedError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	event.Title = "Updated"
	err := repo.Update(context.Background(), event)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestEventRepository_UpdateWithVersion_RowsAffectedError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	event.Title = "Updated"
	err := repo.UpdateWithVersion(context.Background(), event, 1)
	if err != nil {
		t.Fatalf("UpdateWithVersion failed: %v", err)
	}

	if event.Version != 2 {
		t.Errorf("Expected version 2, got %d", event.Version)
	}
}

func TestEventRepository_UpdateStatus_RowsAffectedError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	err := repo.UpdateStatus(context.Background(), event.ID, models.EventStatusPublished)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
}

func TestEventRepository_List_RowsError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	results, err := repo.List(context.Background(), ListFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestEventRepository_GetByStatus_RowsError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	results, err := repo.GetByStatus(context.Background(), models.EventStatusDraft)
	if err != nil {
		t.Fatalf("GetByStatus failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestEventRepository_GetEventsToArchive_RowsError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	now := time.Now()
	event := &models.Event{
		Title:       "Old Event",
		StartTime:   now.Add(-40 * 24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusPublished,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	results, err := repo.GetEventsToArchive(context.Background(), 30)
	if err != nil {
		t.Fatalf("GetEventsToArchive failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestEventRepository_UpdateStatus(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name      string
		id        int64
		newStatus models.EventStatus
		wantErr   bool
		errType   string
	}{
		{
			name:      "update to published",
			id:        event.ID,
			newStatus: models.EventStatusPublished,
			wantErr:   false,
		},
		{
			name:      "update non-existent event",
			id:        999,
			newStatus: models.EventStatusCancelled,
			wantErr:   true,
			errType:   "NotFoundError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateStatus(context.Background(), tt.id, tt.newStatus)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "NotFoundError":
					if _, ok := err.(*models.NotFoundError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr {
				retrieved, err := repo.GetByID(context.Background(), tt.id)
				if err != nil {
					t.Fatalf("Failed to retrieve updated event: %v", err)
				}

				if retrieved.Status != tt.newStatus {
					t.Errorf("Status = %s, want %s", retrieved.Status, tt.newStatus)
				}
			}
		})
	}
}

func TestEventRepository_Delete(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType string
	}{
		{
			name:    "delete existing event",
			id:      event.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent event",
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
				switch tt.errType {
				case "NotFoundError":
					if _, ok := err.(*models.NotFoundError); !ok {
						t.Errorf("Expected error type %s, got %T", tt.errType, err)
					}
				}
			}

			if !tt.wantErr {
				retrieved, err := repo.GetByID(context.Background(), tt.id)
				if err != nil {
					t.Fatalf("Failed to retrieve deleted event: %v", err)
				}

				if retrieved.Status != models.EventStatusArchived {
					t.Errorf("Status = %s, want %s", retrieved.Status, models.EventStatusArchived)
				}
			}
		})
	}
}

func TestEventRepository_List(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	ctx := context.Background()
	_, err := database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (2, 'test2@example.com', 'Test User 2', 'event_manager', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create second test user: %v", err)
	}

	events := []*models.Event{
		{
			Title:       "Event 1",
			StartTime:   time.Now().Add(24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusDraft,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Event 2",
			StartTime:   time.Now().Add(48 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusPublished,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Event 3",
			StartTime:   time.Now().Add(72 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusDraft,
			CreatedBy:   2,
			MaxPlusOnes: 0,
		},
	}

	for _, e := range events {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	tests := []struct {
		name      string
		filters   ListFilters
		wantCount int
	}{
		{
			name:      "no filters",
			filters:   ListFilters{},
			wantCount: 3,
		},
		{
			name: "filter by creator",
			filters: ListFilters{
				CreatorID: int64Ptr(1),
			},
			wantCount: 2,
		},
		{
			name: "filter by status",
			filters: ListFilters{
				Status: statusPtr(models.EventStatusDraft),
			},
			wantCount: 2,
		},
		{
			name: "filter by creator and status",
			filters: ListFilters{
				CreatorID: int64Ptr(1),
				Status:    statusPtr(models.EventStatusDraft),
			},
			wantCount: 1,
		},
		{
			name: "with pagination",
			filters: ListFilters{
				Limit:  2,
				Offset: 0,
			},
			wantCount: 2,
		},
		{
			name: "with pagination offset",
			filters: ListFilters{
				Limit:  2,
				Offset: 2,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.List(context.Background(), tt.filters)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("List() returned %d events, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestEventRepository_GetByStatus(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	events := []*models.Event{
		{
			Title:       "Draft Event 1",
			StartTime:   time.Now().Add(24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusDraft,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Published Event",
			StartTime:   time.Now().Add(48 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusPublished,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Draft Event 2",
			StartTime:   time.Now().Add(72 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusDraft,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
	}

	for _, e := range events {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	tests := []struct {
		name      string
		status    models.EventStatus
		wantCount int
	}{
		{
			name:      "draft events",
			status:    models.EventStatusDraft,
			wantCount: 2,
		},
		{
			name:      "published events",
			status:    models.EventStatusPublished,
			wantCount: 1,
		},
		{
			name:      "cancelled events",
			status:    models.EventStatusCancelled,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetByStatus(context.Background(), tt.status)
			if err != nil {
				t.Errorf("GetByStatus() error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("GetByStatus() returned %d events, want %d", len(results), tt.wantCount)
			}

			for _, event := range results {
				if event.Status != tt.status {
					t.Errorf("Event status = %s, want %s", event.Status, tt.status)
				}
			}
		})
	}
}

func TestEventRepository_GetEventsToArchive(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	now := time.Now()

	events := []*models.Event{
		{
			Title:       "Old Completed Event",
			StartTime:   now.Add(-40 * 24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusPublished,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Recent Event",
			StartTime:   now.Add(-10 * 24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusPublished,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Old Cancelled Event",
			StartTime:   now.Add(-35 * 24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusCancelled,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
		{
			Title:       "Already Archived",
			StartTime:   now.Add(-50 * 24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			Status:      models.EventStatusArchived,
			CreatedBy:   1,
			MaxPlusOnes: 0,
		},
	}

	for _, e := range events {
		if err := repo.Create(context.Background(), e); err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}

	results, err := repo.GetEventsToArchive(context.Background(), 30)
	if err != nil {
		t.Fatalf("GetEventsToArchive() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("GetEventsToArchive() returned %d events, want 2", len(results))
	}

	for _, event := range results {
		if event.Status == models.EventStatusArchived {
			t.Error("GetEventsToArchive() returned already archived event")
		}

		daysSinceEvent := int(now.Sub(event.StartTime).Hours() / 24)
		if daysSinceEvent < 30 {
			t.Errorf("GetEventsToArchive() returned event only %d days old", daysSinceEvent)
		}
	}
}

func TestEventRepository_Integration_ConcurrentUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Concurrent Test",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			e := *event
			e.Title = fmt.Sprintf("Update %d", n)
			err := repo.UpdateWithVersion(context.Background(), &e, 1)
			errors <- err
		}(i)
	}

	wg.Wait()
	close(errors)

	successCount := 0
	conflictCount := 0

	for err := range errors {
		if err == nil {
			successCount++
		} else if _, ok := err.(*models.OptimisticLockError); ok {
			conflictCount++
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if successCount != 1 {
		t.Errorf("Expected 1 successful update, got %d", successCount)
	}

	if conflictCount != 1 {
		t.Errorf("Expected 1 version conflict, got %d", conflictCount)
	}
}

func TestEventRepository_Integration_TransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(context.Background(), event); err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	originalTitle := event.Title

	ctx := context.Background()
	err := database.WithTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE events SET title = ? WHERE id = ?
		`, "Updated in Transaction", event.ID)
		if err != nil {
			return err
		}

		return fmt.Errorf("simulated error to trigger rollback")
	})

	if err == nil {
		t.Fatal("Expected transaction to fail")
	}

	retrieved, err := repo.GetByID(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve event: %v", err)
	}

	if retrieved.Title != originalTitle {
		t.Errorf("Title = %q, want %q (transaction should have rolled back)", retrieved.Title, originalTitle)
	}
}

func TestIsForeignKeyConstraintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "SQLite foreign key error",
			err:      fmt.Errorf("FOREIGN KEY constraint failed"),
			expected: true,
		},
		{
			name:     "Postgres foreign key error",
			err:      fmt.Errorf("violates foreign key constraint"),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isForeignKeyConstraintError(tt.err)
			if result != tt.expected {
				t.Errorf("isForeignKeyConstraintError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEventRepository_GetByStatus_ScanError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	ctx := context.Background()
	_, err := database.Exec(ctx, `
		INSERT INTO events (id, title, description, start_time, end_time, timezone, location, status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline, created_at, updated_at)
		VALUES (100, 'Test', NULL, '2026-01-01 10:00:00', NULL, 'UTC', NULL, 'draft', 1, 1, 0, 0, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test event: %v", err)
	}

	results, err := repo.GetByStatus(context.Background(), models.EventStatusDraft)
	if err != nil {
		t.Fatalf("GetByStatus() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestEventRepository_GetEventsToArchive_ScanError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	now := time.Now()
	ctx := context.Background()
	_, err := database.Exec(ctx, `
		INSERT INTO events (id, title, description, start_time, end_time, timezone, location, status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline, created_at, updated_at)
		VALUES (100, 'Old Event', NULL, ?, NULL, 'UTC', NULL, 'published', 1, 1, 0, 0, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, now.Add(-40*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert test event: %v", err)
	}

	results, err := repo.GetEventsToArchive(context.Background(), 30)
	if err != nil {
		t.Fatalf("GetEventsToArchive() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestEventRepository_List_ScanError(t *testing.T) {
	database := setupEventTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)

	ctx := context.Background()
	_, err := database.Exec(ctx, `
		INSERT INTO events (id, title, description, start_time, end_time, timezone, location, status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline, created_at, updated_at)
		VALUES (100, 'Test', NULL, '2026-01-01 10:00:00', NULL, 'UTC', NULL, 'draft', 1, 1, 0, 0, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test event: %v", err)
	}

	results, err := repo.List(context.Background(), ListFilters{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}
