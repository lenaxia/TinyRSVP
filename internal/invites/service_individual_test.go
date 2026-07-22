package invites

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

type mockEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) ListWithStats(ctx context.Context, filters repositories.ListFilters) ([]*models.EventWithStats, error) {
	return nil, nil
}

func (m *mockEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockEventRepository) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {
	return nil, nil
}

func (m *mockEventRepository) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	return nil
}

func (m *mockEventRepository) DeleteComponentOverrides(ctx context.Context, eventID int64) error {
	return nil
}

func TestCreateIndividualInvite_Success(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	eventTime := time.Now().Add(30 * 24 * time.Hour)
	event := &models.Event{
		ID:          1,
		Title:       "Test Event",
		StartTime:   eventTime,
		Status:      models.EventStatusDraft,
		CreatedBy:   100,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			if id == 1 {
				return event, nil
			}
			return nil, &models.NotFoundError{Resource: "Event", ID: id}
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{
		ID:   100,
		Role: models.RoleEventManager,
	}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "guest@example.com",
		Name:    testutil.StringPtr("John Doe"),
	}

	resp, err := service.CreateIndividualInvite(ctx, user, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Invite == nil {
		t.Fatal("Expected invite to be non-nil")
	}

	if resp.Token == "" {
		t.Fatal("Expected token to be non-empty")
	}

	if resp.Invite.Email == nil || *resp.Invite.Email != "guest@example.com" {
		t.Errorf("Expected email 'guest@example.com', got %v", resp.Invite.Email)
	}

	if resp.Invite.Status != models.InviteStatusDraft {
		t.Errorf("Expected status 'draft', got %s", resp.Invite.Status)
	}

	expectedExpiry := eventTime.Add(30 * 24 * time.Hour)
	if resp.Invite.ExpiresAt.Before(expectedExpiry.Add(-1*time.Minute)) ||
		resp.Invite.ExpiresAt.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("Expected expiry around %v, got %v", expectedExpiry, resp.Invite.ExpiresAt)
	}
}

func TestCreateIndividualInvite_MissingEmail(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}
	eventRepo := &mockEventRepository{}
	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected error for missing email")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestCreateIndividualInvite_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}
	eventRepo := &mockEventRepository{}
	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	tests := []struct {
		name  string
		email string
	}{
		{"no at sign", "notanemail"},
		{"no domain", "test@"},
		{"no local part", "@example.com"},
		{"spaces", "test @example.com"},
		{"multiple at signs", "test@@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateIndividualInviteRequest{
				EventID: 1,
				Email:   tt.email,
			}

			_, err := service.CreateIndividualInvite(ctx, user, req)
			if err == nil {
				t.Fatal("Expected error for invalid email")
			}

			var validationErr *models.ValidationError
			if !errors.As(err, &validationErr) {
				t.Errorf("Expected ValidationError, got %T", err)
			}
		})
	}
}

func TestCreateIndividualInvite_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusDraft,
		CreatedBy:   100,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{
		findDuplicateEmailsFunc: func(ctx context.Context, eventID int64, emails []string) ([]string, error) {
			return emails, nil
		},
	}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "duplicate@example.com",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}

	var conflictErr *models.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Errorf("Expected ConflictError, got %T", err)
	}
}

func TestCreateIndividualInvite_EventNotFound(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return nil, &models.NotFoundError{Resource: "Event", ID: id}
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 999,
		Email:   "guest@example.com",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected error for non-existent event")
	}

	var notFoundErr *models.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestCreateIndividualInvite_CancelledEvent(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusCancelled,
		CreatedBy:   100,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "guest@example.com",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected error for cancelled event")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestCreateIndividualInvite_ArchivedEvent(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusArchived,
		CreatedBy:   100,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "guest@example.com",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected error for archived event")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestCreateIndividualInvite_PermissionDenied_NotCreator(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusDraft,
		CreatedBy:   200,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "guest@example.com",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected permission denied error")
	}

	var permErr *models.PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Errorf("Expected PermissionDeniedError, got %T", err)
	}
}

func TestCreateIndividualInvite_PermissionGranted_Admin(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusDraft,
		CreatedBy:   200,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleAdmin}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "guest@example.com",
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err != nil {
		t.Fatalf("Expected no error for admin, got %v", err)
	}
}

func TestCreateIndividualInvite_MaxPlusOnesDefault(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusDraft,
		CreatedBy:   100,
		MaxPlusOnes: 3,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	req := &CreateIndividualInviteRequest{
		EventID: 1,
		Email:   "guest@example.com",
	}

	resp, err := service.CreateIndividualInvite(ctx, user, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Invite.MaxPlusOnes != 3 {
		t.Errorf("Expected max_plus_ones to default to 3, got %d", resp.Invite.MaxPlusOnes)
	}
}

func TestCreateIndividualInvite_MaxPlusOnesCustom(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusDraft,
		CreatedBy:   100,
		MaxPlusOnes: 5,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	maxPlusOnes := 2
	req := &CreateIndividualInviteRequest{
		EventID:     1,
		Email:       "guest@example.com",
		MaxPlusOnes: &maxPlusOnes,
	}

	resp, err := service.CreateIndividualInvite(ctx, user, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Invite.MaxPlusOnes != 2 {
		t.Errorf("Expected max_plus_ones to be 2, got %d", resp.Invite.MaxPlusOnes)
	}
}

func TestCreateIndividualInvite_MaxPlusOnesExceeded(t *testing.T) {
	ctx := context.Background()
	generator := &mockGenerator{}

	event := &models.Event{
		ID:          1,
		StartTime:   time.Now().Add(30 * 24 * time.Hour),
		Status:      models.EventStatusDraft,
		CreatedBy:   100,
		MaxPlusOnes: 3,
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return event, nil
		},
	}

	inviteRepo := &mockInviteRepository{}

	service := NewIndividualInviteService(generator, inviteRepo, eventRepo)

	user := &models.User{ID: 100, Role: models.RoleEventManager}

	maxPlusOnes := 5
	req := &CreateIndividualInviteRequest{
		EventID:     1,
		Email:       "guest@example.com",
		MaxPlusOnes: &maxPlusOnes,
	}

	_, err := service.CreateIndividualInvite(ctx, user, req)
	if err == nil {
		t.Fatal("Expected error for max_plus_ones exceeding event limit")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func (m *mockEventRepository) GetDashboardStatsByCreator(ctx context.Context, creatorID int64) (*models.DashboardStats, error) {
	return &models.DashboardStats{}, nil
}

func (m * mockEventRepository) CountEventsByCreator(ctx context.Context, creatorID int64) (int, error) {
	return 0, nil
}
