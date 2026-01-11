package events

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/eventid"
)

var testTime = time.Now().Add(24 * time.Hour)

func TestService_CreateEvent_GeneratesPublicID(t *testing.T) {
	mockRepo := &mockEventRepository{}
	mockValidator := &mockValidator{}
	mockAuthz := &mockAuthorizationChecker{
		CanCreateEventFunc: func(ctx context.Context, user *models.User) bool {
			return true
		},
	}

	mockRepo.CreateFunc = func(ctx context.Context, event *models.Event) error {
		event.ID = 1
		return nil
	}

	service := NewService(mockRepo, mockValidator, mockAuthz)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}

	ctx := auth.WithUser(context.Background(), user)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   testTime,
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 0,
	}

	err := service.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	if event.PublicID == nil {
		t.Fatal("Expected PublicID to be generated, got nil")
	}

	if err := eventid.ValidateEventID(*event.PublicID); err != nil {
		t.Errorf("Generated PublicID is invalid: %v", err)
	}

	if len(*event.PublicID) != eventid.IDLength {
		t.Errorf("Expected PublicID length %d, got %d", eventid.IDLength, len(*event.PublicID))
	}
}

func TestService_CreateEvent_PublicIDUniqueness(t *testing.T) {
	mockRepo := &mockEventRepository{}
	mockValidator := &mockValidator{}
	mockAuthz := &mockAuthorizationChecker{
		CanCreateEventFunc: func(ctx context.Context, user *models.User) bool {
			return true
		},
	}

	nextID := int64(1)
	mockRepo.CreateFunc = func(ctx context.Context, event *models.Event) error {
		event.ID = nextID
		nextID++
		return nil
	}

	service := NewService(mockRepo, mockValidator, mockAuthz)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}

	ctx := auth.WithUser(context.Background(), user)

	publicIDs := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		event := &models.Event{
			Title:       "Test Event",
			StartTime:   testTime,
			Timezone:    "America/Los_Angeles",
			MaxPlusOnes: 0,
		}

		err := service.CreateEvent(ctx, event)
		if err != nil {
			t.Fatalf("CreateEvent() error = %v", err)
		}

		if event.PublicID == nil {
			t.Fatal("Expected PublicID to be generated")
		}

		if publicIDs[*event.PublicID] {
			t.Errorf("Duplicate PublicID generated: %s", *event.PublicID)
		}
		publicIDs[*event.PublicID] = true
	}

	if len(publicIDs) != iterations {
		t.Errorf("Expected %d unique PublicIDs, got %d", iterations, len(publicIDs))
	}
}

func TestService_CreateEvent_PreservesUserProvidedFriendlyName(t *testing.T) {
	mockRepo := &mockEventRepository{}
	mockValidator := &mockValidator{}
	mockAuthz := &mockAuthorizationChecker{
		CanCreateEventFunc: func(ctx context.Context, user *models.User) bool {
			return true
		},
	}

	mockRepo.CreateFunc = func(ctx context.Context, event *models.Event) error {
		event.ID = 1
		return nil
	}

	service := NewService(mockRepo, mockValidator, mockAuthz)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}

	ctx := auth.WithUser(context.Background(), user)

	friendlyName := "my-custom-event"
	event := &models.Event{
		FriendlyName: &friendlyName,
		Title:        "Test Event",
		StartTime:    testTime,
		Timezone:     "America/Los_Angeles",
		MaxPlusOnes:  0,
	}

	err := service.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	if event.PublicID == nil {
		t.Fatal("Expected PublicID to be generated")
	}

	if event.FriendlyName == nil {
		t.Fatal("Expected FriendlyName to be preserved")
	}

	if *event.FriendlyName != friendlyName {
		t.Errorf("Expected FriendlyName %s, got %s", friendlyName, *event.FriendlyName)
	}
}

func TestService_CreateEvent_AllowsNilFriendlyName(t *testing.T) {
	mockRepo := &mockEventRepository{}
	mockValidator := &mockValidator{}
	mockAuthz := &mockAuthorizationChecker{
		CanCreateEventFunc: func(ctx context.Context, user *models.User) bool {
			return true
		},
	}

	mockRepo.CreateFunc = func(ctx context.Context, event *models.Event) error {
		event.ID = 1
		return nil
	}

	service := NewService(mockRepo, mockValidator, mockAuthz)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleEventManager,
	}

	ctx := auth.WithUser(context.Background(), user)

	event := &models.Event{
		FriendlyName: nil,
		Title:        "Test Event",
		StartTime:    testTime,
		Timezone:     "America/Los_Angeles",
		MaxPlusOnes:  0,
	}

	err := service.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	if event.PublicID == nil {
		t.Fatal("Expected PublicID to be generated")
	}

	if event.FriendlyName != nil {
		t.Errorf("Expected FriendlyName to remain nil, got %v", *event.FriendlyName)
	}
}
