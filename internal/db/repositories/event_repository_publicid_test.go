package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEventRepository_GetByPublicID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db)
	repo := NewEventRepository(db)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	publicID := "aBcD123456"
	event := &models.Event{
		PublicID:    &publicID,
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		CreatedBy:   user.ID,
		Status:      models.EventStatusDraft,
		MaxPlusOnes: 2,
	}

	err := repo.Create(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name      string
		publicID  string
		wantErr   bool
		wantEvent bool
	}{
		{
			name:      "valid public_id",
			publicID:  publicID,
			wantErr:   false,
			wantEvent: true,
		},
		{
			name:      "non-existent public_id",
			publicID:  "nonexist1",
			wantErr:   true,
			wantEvent: false,
		},
		{
			name:      "empty public_id",
			publicID:  "",
			wantErr:   true,
			wantEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByPublicID(ctx, tt.publicID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPublicID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantEvent && result == nil {
				t.Error("Expected event, got nil")
				return
			}

			if !tt.wantEvent && result != nil {
				t.Error("Expected nil, got event")
				return
			}

			if tt.wantEvent {
				if result.PublicID == nil || *result.PublicID != publicID {
					t.Errorf("Expected public_id %s, got %v", publicID, result.PublicID)
				}
				if result.Title != event.Title {
					t.Errorf("Expected title %s, got %s", event.Title, result.Title)
				}
			}
		})
	}
}

func TestEventRepository_GetByFriendlyName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db)
	repo := NewEventRepository(db)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	publicID := "xyz9876543"
	friendlyName := "summer-party-2026"
	event := &models.Event{
		PublicID:     &publicID,
		FriendlyName: &friendlyName,
		Title:        "Summer Party 2026",
		StartTime:    time.Now().Add(24 * time.Hour),
		Timezone:     "America/Los_Angeles",
		CreatedBy:    user.ID,
		Status:       models.EventStatusDraft,
		MaxPlusOnes:  2,
	}

	err := repo.Create(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	tests := []struct {
		name         string
		friendlyName string
		wantErr      bool
		wantEvent    bool
	}{
		{
			name:         "valid friendly_name",
			friendlyName: friendlyName,
			wantErr:      false,
			wantEvent:    true,
		},
		{
			name:         "non-existent friendly_name",
			friendlyName: "nonexistent-event",
			wantErr:      true,
			wantEvent:    false,
		},
		{
			name:         "empty friendly_name",
			friendlyName: "",
			wantErr:      true,
			wantEvent:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByFriendlyName(ctx, tt.friendlyName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByFriendlyName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantEvent && result == nil {
				t.Error("Expected event, got nil")
				return
			}

			if !tt.wantEvent && result != nil {
				t.Error("Expected nil, got event")
				return
			}

			if tt.wantEvent {
				if result.FriendlyName == nil || *result.FriendlyName != friendlyName {
					t.Errorf("Expected friendly_name %s, got %v", friendlyName, result.FriendlyName)
				}
				if result.Title != event.Title {
					t.Errorf("Expected title %s, got %s", event.Title, result.Title)
				}
			}
		})
	}
}

func TestEventRepository_Create_WithPublicID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db)
	repo := NewEventRepository(db)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	tests := []struct {
		name      string
		publicID  *string
		wantErr   bool
		checkFunc func(*testing.T, *models.Event)
	}{
		{
			name:     "with public_id",
			publicID: stringPtr("test123456"),
			wantErr:  false,
			checkFunc: func(t *testing.T, e *models.Event) {
				if e.PublicID == nil {
					t.Error("Expected public_id to be set")
				} else if *e.PublicID != "test123456" {
					t.Errorf("Expected public_id test123456, got %s", *e.PublicID)
				}
			},
		},
		{
			name:     "without public_id",
			publicID: nil,
			wantErr:  false,
			checkFunc: func(t *testing.T, e *models.Event) {
				if e.PublicID != nil {
					t.Errorf("Expected public_id to be nil, got %v", *e.PublicID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &models.Event{
				PublicID:    tt.publicID,
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				CreatedBy:   user.ID,
				Status:      models.EventStatusDraft,
				MaxPlusOnes: 0,
			}

			err := repo.Create(ctx, event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, event)
			}
		})
	}
}

func TestEventRepository_Create_WithFriendlyName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db)
	repo := NewEventRepository(db)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	tests := []struct {
		name         string
		friendlyName *string
		wantErr      bool
		checkFunc    func(*testing.T, *models.Event)
	}{
		{
			name:         "with friendly_name",
			friendlyName: stringPtr("test-event-2026"),
			wantErr:      false,
			checkFunc: func(t *testing.T, e *models.Event) {
				if e.FriendlyName == nil {
					t.Error("Expected friendly_name to be set")
				} else if *e.FriendlyName != "test-event-2026" {
					t.Errorf("Expected friendly_name test-event-2026, got %s", *e.FriendlyName)
				}
			},
		},
		{
			name:         "without friendly_name",
			friendlyName: nil,
			wantErr:      false,
			checkFunc: func(t *testing.T, e *models.Event) {
				if e.FriendlyName != nil {
					t.Errorf("Expected friendly_name to be nil, got %v", *e.FriendlyName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &models.Event{
				FriendlyName: tt.friendlyName,
				Title:        "Test Event",
				StartTime:    time.Now().Add(24 * time.Hour),
				Timezone:     "America/Los_Angeles",
				CreatedBy:    user.ID,
				Status:       models.EventStatusDraft,
				MaxPlusOnes:  0,
			}

			err := repo.Create(ctx, event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, event)
			}
		})
	}
}

func TestEventRepository_UniquenessConstraints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db)
	repo := NewEventRepository(db)
	ctx := context.Background()

	user := createTestUser(t, userRepo)

	t.Run("duplicate public_id should fail", func(t *testing.T) {
		publicID := "duplicate12"

		event1 := &models.Event{
			PublicID:    &publicID,
			Title:       "Event 1",
			StartTime:   time.Now().Add(24 * time.Hour),
			Timezone:    "America/Los_Angeles",
			CreatedBy:   user.ID,
			Status:      models.EventStatusDraft,
			MaxPlusOnes: 0,
		}

		err := repo.Create(ctx, event1)
		if err != nil {
			t.Fatalf("Failed to create first event: %v", err)
		}

		event2 := &models.Event{
			PublicID:    &publicID,
			Title:       "Event 2",
			StartTime:   time.Now().Add(48 * time.Hour),
			Timezone:    "America/Los_Angeles",
			CreatedBy:   user.ID,
			Status:      models.EventStatusDraft,
			MaxPlusOnes: 0,
		}

		err = repo.Create(ctx, event2)
		if err == nil {
			t.Error("Expected error for duplicate public_id, got nil")
		}
	})

	t.Run("duplicate friendly_name should fail", func(t *testing.T) {
		friendlyName := "duplicate-event"

		event1 := &models.Event{
			FriendlyName: &friendlyName,
			Title:        "Event 1",
			StartTime:    time.Now().Add(24 * time.Hour),
			Timezone:     "America/Los_Angeles",
			CreatedBy:    user.ID,
			Status:       models.EventStatusDraft,
			MaxPlusOnes:  0,
		}

		err := repo.Create(ctx, event1)
		if err != nil {
			t.Fatalf("Failed to create first event: %v", err)
		}

		event2 := &models.Event{
			FriendlyName: &friendlyName,
			Title:        "Event 2",
			StartTime:    time.Now().Add(48 * time.Hour),
			Timezone:     "America/Los_Angeles",
			CreatedBy:    user.ID,
			Status:       models.EventStatusDraft,
			MaxPlusOnes:  0,
		}

		err = repo.Create(ctx, event2)
		if err == nil {
			t.Error("Expected error for duplicate friendly_name, got nil")
		}
	})
}
