package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEventRepository_GetComponentOverrides(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")
	event, err := repo.GetByID(ctx, eventID)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#ff0000",
					},
				},
			},
		},
		Additions: []models.Component{},
		Removals:  []string{},
	}

	overridesJSON, err := json.Marshal(overrides)
	if err != nil {
		t.Fatalf("Failed to marshal overrides: %v", err)
	}
	overridesStr := string(overridesJSON)
	event.ComponentOverrides = &overridesStr

	if err := repo.Update(ctx, event); err != nil {
		t.Fatalf("Failed to update event: %v", err)
	}

	result, err := repo.GetComponentOverrides(ctx, event.ID)
	if err != nil {
		t.Fatalf("GetComponentOverrides failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", result.Version)
	}

	if len(result.Overrides) != 1 {
		t.Errorf("Expected 1 override, got %d", len(result.Overrides))
	}
}

func TestEventRepository_GetComponentOverrides_NoOverrides(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")

	result, err := repo.GetComponentOverrides(ctx, eventID)
	if err != nil {
		t.Fatalf("GetComponentOverrides failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for event with no overrides")
	}
}

func TestEventRepository_GetComponentOverrides_InvalidJSON(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")
	event, err := repo.GetByID(ctx, eventID)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	invalidJSON := "invalid json"
	event.ComponentOverrides = &invalidJSON

	if err := repo.Update(ctx, event); err != nil {
		t.Fatalf("Failed to update event: %v", err)
	}

	_, err = repo.GetComponentOverrides(ctx, event.ID)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestEventRepository_GetComponentOverrides_EventNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	ctx := context.Background()

	_, err := repo.GetComponentOverrides(ctx, 99999)
	if err == nil {
		t.Error("Expected error for non-existent event")
	}

	if _, ok := err.(*models.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestEventRepository_UpdateComponentOverrides(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#00ff00",
					},
				},
			},
		},
		Additions: []models.Component{
			{
				ID:   "custom-subtitle",
				Type: models.ComponentTypeTextBox,
				Position: models.Position{
					Mode: models.PositionModeAbsolute,
					X:    strPtr("50%"),
					Y:    strPtr("300px"),
				},
				Dimensions: models.Dimensions{
					Width:  "80%",
					Height: "auto",
				},
				ZIndex:  10,
				Visible: true,
			},
		},
		Removals: []string{"location-text"},
	}

	err := repo.UpdateComponentOverrides(ctx, eventID, overrides)
	if err != nil {
		t.Fatalf("UpdateComponentOverrides failed: %v", err)
	}

	result, err := repo.GetComponentOverrides(ctx, eventID)
	if err != nil {
		t.Fatalf("GetComponentOverrides failed: %v", err)
	}

	if result.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", result.Version)
	}

	if len(result.Overrides) != 1 {
		t.Errorf("Expected 1 override, got %d", len(result.Overrides))
	}

	if len(result.Additions) != 1 {
		t.Errorf("Expected 1 addition, got %d", len(result.Additions))
	}

	if len(result.Removals) != 1 {
		t.Errorf("Expected 1 removal, got %d", len(result.Removals))
	}
}

func TestEventRepository_UpdateComponentOverrides_EventNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	ctx := context.Background()

	overrides := &models.ComponentOverrides{
		Version:   "1.0",
		Overrides: []models.ComponentOverride{},
		Additions: []models.Component{},
		Removals:  []string{},
	}

	err := repo.UpdateComponentOverrides(ctx, 99999, overrides)
	if err == nil {
		t.Error("Expected error for non-existent event")
	}

	if _, ok := err.(*models.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestEventRepository_UpdateComponentOverrides_NilOverrides(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")

	err := repo.UpdateComponentOverrides(ctx, eventID, nil)
	if err == nil {
		t.Error("Expected error for nil overrides")
	}
}

func TestEventRepository_DeleteComponentOverrides(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")

	overrides := &models.ComponentOverrides{
		Version: "1.0",
		Overrides: []models.ComponentOverride{
			{
				ID: "title-text",
				Updates: map[string]interface{}{
					"content": map[string]interface{}{
						"color": "#ff0000",
					},
				},
			},
		},
		Additions: []models.Component{},
		Removals:  []string{},
	}

	err := repo.UpdateComponentOverrides(ctx, eventID, overrides)
	if err != nil {
		t.Fatalf("UpdateComponentOverrides failed: %v", err)
	}

	err = repo.DeleteComponentOverrides(ctx, eventID)
	if err != nil {
		t.Fatalf("DeleteComponentOverrides failed: %v", err)
	}

	result, err := repo.GetComponentOverrides(ctx, eventID)
	if err != nil {
		t.Fatalf("GetComponentOverrides failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result after deletion")
	}
}

func TestEventRepository_DeleteComponentOverrides_EventNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	ctx := context.Background()

	err := repo.DeleteComponentOverrides(ctx, 99999)
	if err == nil {
		t.Error("Expected error for non-existent event")
	}

	if _, ok := err.(*models.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestEventRepository_DeleteComponentOverrides_NoOverrides(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewEventRepository(database)
	userRepo := NewUserRepository(database)
	ctx := context.Background()

	user := createTestUser(t, userRepo)
	eventID := createTestEventForCustomization(t, database, user.ID, "Test Event")

	err := repo.DeleteComponentOverrides(ctx, eventID)
	if err != nil {
		t.Fatalf("DeleteComponentOverrides should succeed even with no overrides: %v", err)
	}
}

func createTestEventForCustomization(t *testing.T, database db.Database, creatorID int64, title string) int64 {
	t.Helper()

	repo := NewEventRepository(database)
	ctx := context.Background()

	publicID := "test-event-custom-" + time.Now().Format("20060102150405")
	event := &models.Event{
		PublicID:    &publicID,
		Title:       title,
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   creatorID,
		MaxPlusOnes: 0,
	}

	if err := repo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	return event.ID
}

func strPtr(s string) *string {
	return &s
}
