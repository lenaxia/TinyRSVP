package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockEventRepository struct {
	CreateFunc                   func(ctx context.Context, event *models.Event) error
	GetByIDFunc                  func(ctx context.Context, id int64) (*models.Event, error)
	GetByPublicIDFunc            func(ctx context.Context, publicID string) (*models.Event, error)
	GetByFriendlyNameFunc        func(ctx context.Context, friendlyName string) (*models.Event, error)
	UpdateFunc                   func(ctx context.Context, event *models.Event) error
	UpdateWithVersionFunc        func(ctx context.Context, event *models.Event, expectedVersion int) error
	UpdateStatusFunc             func(ctx context.Context, id int64, status models.EventStatus) error
	DeleteFunc                   func(ctx context.Context, id int64) error
	ListFunc                     func(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error)
	GetByStatusFunc              func(ctx context.Context, status models.EventStatus) ([]*models.Event, error)
	GetEventsToArchiveFunc       func(ctx context.Context, daysAfterEvent int) ([]*models.Event, error)
	GetByCreatorIDFunc           func(ctx context.Context, creatorID int64) ([]*models.Event, error)
	GetComponentOverridesFunc    func(ctx context.Context, eventID int64) (*models.ComponentOverrides, error)
	UpdateComponentOverridesFunc func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error
	DeleteComponentOverridesFunc func(ctx context.Context, eventID int64) error
}

func (m *mockEventRepository) Create(ctx context.Context, event *models.Event) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	return nil
}

func (m *mockEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockEventRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	if m.GetByPublicIDFunc != nil {
		return m.GetByPublicIDFunc(ctx, publicID)
	}
	return nil, nil
}

func (m *mockEventRepository) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	if m.GetByFriendlyNameFunc != nil {
		return m.GetByFriendlyNameFunc(ctx, friendlyName)
	}
	return nil, nil
}

func (m *mockEventRepository) Update(ctx context.Context, event *models.Event) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, event)
	}
	return nil
}

func (m *mockEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	if m.UpdateWithVersionFunc != nil {
		return m.UpdateWithVersionFunc(ctx, event, expectedVersion)
	}
	return nil
}

func (m *mockEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockEventRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return nil, nil
}

func (m *mockEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	if m.GetByStatusFunc != nil {
		return m.GetByStatusFunc(ctx, status)
	}
	return nil, nil
}

func (m *mockEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	if m.GetEventsToArchiveFunc != nil {
		return m.GetEventsToArchiveFunc(ctx, daysAfterEvent)
	}
	return nil, nil
}

func (m *mockEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	if m.GetByCreatorIDFunc != nil {
		return m.GetByCreatorIDFunc(ctx, creatorID)
	}
	return nil, nil
}

func (m *mockEventRepository) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockEventRepository) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {
	if m.GetComponentOverridesFunc != nil {
		return m.GetComponentOverridesFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *mockEventRepository) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	if m.UpdateComponentOverridesFunc != nil {
		return m.UpdateComponentOverridesFunc(ctx, eventID, overrides)
	}
	return nil
}

func (m *mockEventRepository) DeleteComponentOverrides(ctx context.Context, eventID int64) error {
	if m.DeleteComponentOverridesFunc != nil {
		return m.DeleteComponentOverridesFunc(ctx, eventID)
	}
	return nil
}

type mockValidator struct {
	ValidateCreateFunc          func(ctx context.Context, event *models.Event) error
	ValidateUpdateFunc          func(ctx context.Context, event *models.Event) error
	ValidateStateTransitionFunc func(from, to models.EventStatus) error
}

func (m *mockValidator) ValidateCreate(ctx context.Context, event *models.Event) error {
	if m.ValidateCreateFunc != nil {
		return m.ValidateCreateFunc(ctx, event)
	}
	return nil
}

func (m *mockValidator) ValidateUpdate(ctx context.Context, event *models.Event) error {
	if m.ValidateUpdateFunc != nil {
		return m.ValidateUpdateFunc(ctx, event)
	}
	return nil
}

func (m *mockValidator) ValidateStateTransition(from, to models.EventStatus) error {
	if m.ValidateStateTransitionFunc != nil {
		return m.ValidateStateTransitionFunc(from, to)
	}
	return nil
}

type mockAuthorizationChecker struct {
	CanCreateEventFunc     func(ctx context.Context, user *models.User) bool
	CanEditEventFunc       func(ctx context.Context, user *models.User, event *models.Event) bool
	CanDeleteEventFunc     func(ctx context.Context, user *models.User, event *models.Event) bool
	CanViewEventFunc       func(ctx context.Context, user *models.User, event *models.Event) bool
	CanManageInvitesFunc   func(ctx context.Context, user *models.User, event *models.Event) bool
	CanViewRSVPsFunc       func(ctx context.Context, user *models.User, event *models.Event) bool
	CanManageUsersFunc     func(ctx context.Context, user *models.User) bool
	CanConfigureSystemFunc func(ctx context.Context, user *models.User) bool
	IsAdminFunc            func(user *models.User) bool
	IsEventManagerFunc     func(user *models.User) bool
}

func (m *mockAuthorizationChecker) CanCreateEvent(ctx context.Context, user *models.User) bool {
	if m.CanCreateEventFunc != nil {
		return m.CanCreateEventFunc(ctx, user)
	}
	return false
}

func (m *mockAuthorizationChecker) CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if m.CanEditEventFunc != nil {
		return m.CanEditEventFunc(ctx, user, event)
	}
	return false
}

func (m *mockAuthorizationChecker) CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if m.CanDeleteEventFunc != nil {
		return m.CanDeleteEventFunc(ctx, user, event)
	}
	return false
}

func (m *mockAuthorizationChecker) CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if m.CanViewEventFunc != nil {
		return m.CanViewEventFunc(ctx, user, event)
	}
	return false
}

func (m *mockAuthorizationChecker) CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool {
	if m.CanManageInvitesFunc != nil {
		return m.CanManageInvitesFunc(ctx, user, event)
	}
	return false
}

func (m *mockAuthorizationChecker) CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool {
	if m.CanViewRSVPsFunc != nil {
		return m.CanViewRSVPsFunc(ctx, user, event)
	}
	return false
}

func (m *mockAuthorizationChecker) CanManageUsers(ctx context.Context, user *models.User) bool {
	if m.CanManageUsersFunc != nil {
		return m.CanManageUsersFunc(ctx, user)
	}
	return false
}

func (m *mockAuthorizationChecker) CanConfigureSystem(ctx context.Context, user *models.User) bool {
	if m.CanConfigureSystemFunc != nil {
		return m.CanConfigureSystemFunc(ctx, user)
	}
	return false
}

func (m *mockAuthorizationChecker) IsAdmin(user *models.User) bool {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(user)
	}
	return false
}

func (m *mockAuthorizationChecker) IsEventManager(user *models.User) bool {
	if m.IsEventManagerFunc != nil {
		return m.IsEventManagerFunc(user)
	}
	return false
}

func TestNewService(t *testing.T) {
	repo := &mockEventRepository{}
	validator := &mockValidator{}
	authz := &mockAuthorizationChecker{}

	service := NewService(repo, validator, authz)

	if service == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestService_CreateEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		event      *models.Event
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
		validate   func(*testing.T, *models.Event)
	}{
		{
			name: "event manager creates event successfully",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.CanCreateEventFunc = func(ctx context.Context, user *models.User) bool {
					return true
				}
				val.ValidateCreateFunc = func(ctx context.Context, event *models.Event) error {
					return nil
				}
				repo.CreateFunc = func(ctx context.Context, event *models.Event) error {
					event.ID = 1
					return nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, event *models.Event) {
				if event.CreatedBy != 1 {
					t.Errorf("CreatedBy = %d, want 1", event.CreatedBy)
				}
				if event.Status != models.EventStatusDraft {
					t.Errorf("Status = %q, want %q", event.Status, models.EventStatusDraft)
				}
				if event.Version != 1 {
					t.Errorf("Version = %d, want 1", event.Version)
				}
			},
		},
		{
			name: "admin creates event successfully",
			user: &models.User{
				ID:   2,
				Role: models.RoleAdmin,
			},
			event: &models.Event{
				Title:       "Admin Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.CanCreateEventFunc = func(ctx context.Context, user *models.User) bool {
					return true
				}
				val.ValidateCreateFunc = func(ctx context.Context, event *models.Event) error {
					return nil
				}
				repo.CreateFunc = func(ctx context.Context, event *models.Event) error {
					event.ID = 2
					return nil
				}
			},
			wantErr: false,
			validate: func(t *testing.T, event *models.Event) {
				if event.CreatedBy != 2 {
					t.Errorf("CreatedBy = %d, want 2", event.CreatedBy)
				}
			},
		},
		{
			name: "guest cannot create event",
			user: &models.User{
				ID:   3,
				Role: models.RoleGuest,
			},
			event: &models.Event{
				Title:       "Guest Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.CanCreateEventFunc = func(ctx context.Context, user *models.User) bool {
					return false
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
		{
			name: "validation error",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				Title:       "AB",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.CanCreateEventFunc = func(ctx context.Context, user *models.User) bool {
					return true
				}
				val.ValidateCreateFunc = func(ctx context.Context, event *models.Event) error {
					return &models.ValidationError{Field: "title", Message: "too short"}
				}
			},
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "no user in context",
			user: nil,
			event: &models.Event{
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {},
			wantErr:    true,
			errType:    &models.PermissionDeniedError{},
		},
		{
			name: "repository error",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				Title:       "Test Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.CanCreateEventFunc = func(ctx context.Context, user *models.User) bool {
					return true
				}
				val.ValidateCreateFunc = func(ctx context.Context, event *models.Event) error {
					return nil
				}
				repo.CreateFunc = func(ctx context.Context, event *models.Event) error {
					return errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			err := service.CreateEvent(ctx, tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *models.PermissionDeniedError:
					var permErr *models.PermissionDeniedError
					if !errors.As(err, &permErr) {
						t.Errorf("CreateEvent() error type = %T, want *models.PermissionDeniedError", err)
					}
				case *models.ValidationError:
					var valErr *models.ValidationError
					if !errors.As(err, &valErr) {
						t.Errorf("CreateEvent() error type = %T, want *models.ValidationError", err)
					}
				}
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, tt.event)
			}
		})
	}
}

func TestService_GetEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		eventID    int64
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
	}{
		{
			name: "event manager gets event successfully",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
					}, nil
				}
				authz.CanViewEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return true
				}
			},
			wantErr: false,
		},
		{
			name: "admin gets any event",
			user: &models.User{
				ID:   2,
				Role: models.RoleAdmin,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
					}, nil
				}
				authz.CanViewEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return true
				}
			},
			wantErr: false,
		},
		{
			name: "guest cannot view event",
			user: &models.User{
				ID:   3,
				Role: models.RoleGuest,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
					}, nil
				}
				authz.CanViewEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return false
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
		{
			name:    "event not found",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			eventID: 999,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.NotFoundError{Resource: "Event", ID: id}
				}
			},
			wantErr: true,
			errType: &models.NotFoundError{},
		},
		{
			name:    "no user in context",
			user:    nil,
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{ID: 1}, nil
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			event, err := service.GetEvent(ctx, tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *models.PermissionDeniedError:
					var permErr *models.PermissionDeniedError
					if !errors.As(err, &permErr) {
						t.Errorf("GetEvent() error type = %T, want *models.PermissionDeniedError", err)
					}
				case *models.NotFoundError:
					var notFoundErr *models.NotFoundError
					if !errors.As(err, &notFoundErr) {
						t.Errorf("GetEvent() error type = %T, want *models.NotFoundError", err)
					}
				}
			}

			if !tt.wantErr && event == nil {
				t.Error("GetEvent() returned nil event")
			}
		})
	}
}

func TestService_UpdateEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		event      *models.Event
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
	}{
		{
			name: "owner updates own event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				ID:          1,
				Title:       "Updated Title",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
				Version:     1,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Version:   1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
				val.ValidateUpdateFunc = func(ctx context.Context, event *models.Event) error {
					return nil
				}
				repo.UpdateWithVersionFunc = func(ctx context.Context, event *models.Event, expectedVersion int) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "admin updates any event",
			user: &models.User{
				ID:   2,
				Role: models.RoleAdmin,
			},
			event: &models.Event{
				ID:          1,
				Title:       "Admin Update",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
				Version:     1,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Version:   1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.Role == models.RoleAdmin
				}
				val.ValidateUpdateFunc = func(ctx context.Context, event *models.Event) error {
					return nil
				}
				repo.UpdateWithVersionFunc = func(ctx context.Context, event *models.Event, expectedVersion int) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "non-owner cannot update",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				ID:          1,
				Title:       "Unauthorized Update",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
				Version:     1,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Version:   1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
		{
			name: "validation error",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				ID:          1,
				Title:       "AB",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
				Version:     1,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Version:   1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return true
				}
				val.ValidateUpdateFunc = func(ctx context.Context, event *models.Event) error {
					return &models.ValidationError{Field: "title", Message: "too short"}
				}
			},
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "version conflict",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			event: &models.Event{
				ID:          1,
				Title:       "Updated",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
				Version:     1,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Version:   1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return true
				}
				val.ValidateUpdateFunc = func(ctx context.Context, event *models.Event) error {
					return nil
				}
				repo.UpdateWithVersionFunc = func(ctx context.Context, event *models.Event, expectedVersion int) error {
					return &models.OptimisticLockError{
						Resource:        "Event",
						ID:              1,
						ExpectedVersion: 1,
						ActualVersion:   2,
					}
				}
			},
			wantErr: true,
			errType: &models.OptimisticLockError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			err := service.UpdateEvent(ctx, tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *models.PermissionDeniedError:
					var permErr *models.PermissionDeniedError
					if !errors.As(err, &permErr) {
						t.Errorf("UpdateEvent() error type = %T, want *models.PermissionDeniedError", err)
					}
				case *models.ValidationError:
					var valErr *models.ValidationError
					if !errors.As(err, &valErr) {
						t.Errorf("UpdateEvent() error type = %T, want *models.ValidationError", err)
					}
				case *models.OptimisticLockError:
					var lockErr *models.OptimisticLockError
					if !errors.As(err, &lockErr) {
						t.Errorf("UpdateEvent() error type = %T, want *models.OptimisticLockError", err)
					}
				}
			}
		})
	}
}

func TestService_DeleteEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		eventID    int64
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
	}{
		{
			name: "owner deletes own event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanDeleteEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
				repo.DeleteFunc = func(ctx context.Context, id int64) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "admin deletes any event",
			user: &models.User{
				ID:   2,
				Role: models.RoleAdmin,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanDeleteEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.Role == models.RoleAdmin
				}
				repo.DeleteFunc = func(ctx context.Context, id int64) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "non-owner cannot delete",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				authz.CanDeleteEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
		{
			name:    "event not found",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			eventID: 999,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.NotFoundError{Resource: "Event", ID: id}
				}
			},
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			err := service.DeleteEvent(ctx, tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *models.PermissionDeniedError:
					var permErr *models.PermissionDeniedError
					if !errors.As(err, &permErr) {
						t.Errorf("DeleteEvent() error type = %T, want *models.PermissionDeniedError", err)
					}
				case *models.NotFoundError:
					var notFoundErr *models.NotFoundError
					if !errors.As(err, &notFoundErr) {
						t.Errorf("DeleteEvent() error type = %T, want *models.NotFoundError", err)
					}
				}
			}
		})
	}
}

func TestService_ListEvents(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		filters    ListFilters
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		wantCount  int
	}{
		{
			name: "admin lists all events",
			user: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			filters: ListFilters{
				Limit: 10,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsAdminFunc = func(user *models.User) bool {
					return user.Role == models.RoleAdmin
				}
				authz.IsEventManagerFunc = func(user *models.User) bool {
					return user.Role == models.RoleAdmin || user.Role == models.RoleEventManager
				}
				repo.ListFunc = func(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
					return []*models.Event{
						{ID: 1, Title: "Event 1"},
						{ID: 2, Title: "Event 2"},
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "event manager lists own events",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			filters: ListFilters{
				Limit: 10,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsAdminFunc = func(user *models.User) bool {
					return false
				}
				authz.IsEventManagerFunc = func(user *models.User) bool {
					return user.Role == models.RoleAdmin || user.Role == models.RoleEventManager
				}
				repo.ListFunc = func(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
					if filters.CreatorID != nil && *filters.CreatorID == 1 {
						return []*models.Event{
							{ID: 1, Title: "My Event", CreatedBy: 1},
						}, nil
					}
					return nil, nil
				}
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "guest cannot list events",
			user: &models.User{
				ID:   3,
				Role: models.RoleGuest,
			},
			filters: ListFilters{
				Limit: 10,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsEventManagerFunc = func(user *models.User) bool {
					return false
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			events, err := service.ListEvents(ctx, tt.filters)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(events) != tt.wantCount {
				t.Errorf("ListEvents() returned %d events, want %d", len(events), tt.wantCount)
			}
		})
	}
}

func TestService_PublishEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		eventID    int64
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
	}{
		{
			name: "owner publishes draft event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
				val.ValidateStateTransitionFunc = func(from, to models.EventStatus) error {
					if from == models.EventStatusDraft && to == models.EventStatusPublished {
						return nil
					}
					return &models.ValidationError{Field: "status", Message: "invalid transition"}
				}
				repo.UpdateStatusFunc = func(ctx context.Context, id int64, status models.EventStatus) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "cannot publish already published event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Status:    models.EventStatusPublished,
						CreatedBy: 1,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return true
				}
				val.ValidateStateTransitionFunc = func(from, to models.EventStatus) error {
					return &models.ValidationError{Field: "status", Message: "already published"}
				}
			},
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "non-owner cannot publish",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			err := service.PublishEvent(ctx, tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("PublishEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *models.PermissionDeniedError:
					var permErr *models.PermissionDeniedError
					if !errors.As(err, &permErr) {
						t.Errorf("PublishEvent() error type = %T, want *models.PermissionDeniedError", err)
					}
				case *models.ValidationError:
					var valErr *models.ValidationError
					if !errors.As(err, &valErr) {
						t.Errorf("PublishEvent() error type = %T, want *models.ValidationError", err)
					}
				}
			}
		})
	}
}

func TestService_CancelEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		eventID    int64
		reason     string
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
	}{
		{
			name: "owner cancels published event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			reason:  "Weather conditions",
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Status:    models.EventStatusPublished,
						CreatedBy: 1,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
				val.ValidateStateTransitionFunc = func(from, to models.EventStatus) error {
					if from == models.EventStatusPublished && to == models.EventStatusCancelled {
						return nil
					}
					return &models.ValidationError{Field: "status", Message: "invalid transition"}
				}
				repo.UpdateStatusFunc = func(ctx context.Context, id int64, status models.EventStatus) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "non-owner cannot cancel",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			reason:  "Test",
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Status:    models.EventStatusPublished,
						CreatedBy: 1,
					}, nil
				}
				authz.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			err := service.CancelEvent(ctx, tt.eventID, tt.reason)

			if (err != nil) != tt.wantErr {
				t.Errorf("CancelEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *models.PermissionDeniedError:
					var permErr *models.PermissionDeniedError
					if !errors.As(err, &permErr) {
						t.Errorf("CancelEvent() error type = %T, want *models.PermissionDeniedError", err)
					}
				case *models.ValidationError:
					var valErr *models.ValidationError
					if !errors.As(err, &valErr) {
						t.Errorf("CancelEvent() error type = %T, want *models.ValidationError", err)
					}
				}
			}
		})
	}
}

func TestService_ArchiveEvent(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		eventID    int64
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		errType    error
	}{
		{
			name: "admin archives event",
			user: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsAdminFunc = func(user *models.User) bool {
					return user.Role == models.RoleAdmin
				}
				repo.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:     1,
						Status: models.EventStatusPublished,
					}, nil
				}
				val.ValidateStateTransitionFunc = func(from, to models.EventStatus) error {
					return nil
				}
				repo.UpdateStatusFunc = func(ctx context.Context, id int64, status models.EventStatus) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "non-admin cannot archive",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsAdminFunc = func(user *models.User) bool {
					return false
				}
			},
			wantErr: true,
			errType: &models.PermissionDeniedError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			err := service.ArchiveEvent(ctx, tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArchiveEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				var permErr *models.PermissionDeniedError
				if !errors.As(err, &permErr) {
					t.Errorf("ArchiveEvent() error type = %T, want *models.PermissionDeniedError", err)
				}
			}
		})
	}
}

func TestService_GetEventsToArchive(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		setupMocks func(*mockEventRepository, *mockValidator, *mockAuthorizationChecker)
		wantErr    bool
		wantCount  int
	}{
		{
			name: "admin gets events to archive",
			user: &models.User{
				ID:   1,
				Role: models.RoleAdmin,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsAdminFunc = func(user *models.User) bool {
					return user.Role == models.RoleAdmin
				}
				repo.GetEventsToArchiveFunc = func(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
					return []*models.Event{
						{ID: 1, Status: models.EventStatusPublished},
						{ID: 2, Status: models.EventStatusCancelled},
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "non-admin cannot get events to archive",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMocks: func(repo *mockEventRepository, val *mockValidator, authz *mockAuthorizationChecker) {
				authz.IsAdminFunc = func(user *models.User) bool {
					return false
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{}
			validator := &mockValidator{}
			authz := &mockAuthorizationChecker{}

			tt.setupMocks(repo, validator, authz)

			service := NewService(repo, validator, authz)

			ctx := context.Background()
			if tt.user != nil {
				ctx = auth.WithUser(ctx, tt.user)
			}

			events, err := service.GetEventsToArchive(ctx, 30)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventsToArchive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(events) != tt.wantCount {
				t.Errorf("GetEventsToArchive() returned %d events, want %d", len(events), tt.wantCount)
			}
		})
	}
}
