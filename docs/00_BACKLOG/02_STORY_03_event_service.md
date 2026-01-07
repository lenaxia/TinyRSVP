# User Story: Event Service Layer

**Epic:** [02_EPIC_events.md](02_EPIC_events.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 8 hours

---

## User Story

As a **developer**, I want **a service layer for event business logic** so that **event operations include validation, authorization, and proper state management**.

---

## Acceptance Criteria

- [ ] EventService interface defined
- [ ] Create event with validation and authorization
- [ ] Get event by ID with authorization check
- [ ] Update event with optimistic locking
- [ ] List events with filtering
- [ ] Publish event with state transition validation
- [ ] Cancel event with state transition validation
- [ ] Archive event operation
- [ ] Permission checks enforced on all operations
- [ ] All tests pass with timeout
- [ ] Integration tests with repository pass

---

## Technical Details

### EventService Interface

```go
package events

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
    CreateEvent(ctx context.Context, event *models.Event) error
    GetEvent(ctx context.Context, id int64) (*models.Event, error)
    UpdateEvent(ctx context.Context, event *models.Event) error
    DeleteEvent(ctx context.Context, id int64) error
    ListEvents(ctx context.Context, filters ListFilters) ([]*models.Event, error)
    PublishEvent(ctx context.Context, id int64) error
    CancelEvent(ctx context.Context, id int64, reason string) error
    ArchiveEvent(ctx context.Context, id int64) error
    GetEventsToArchive(ctx context.Context) ([]*models.Event, error)
}

type ListFilters struct {
    CreatorID *int64
    Status    *models.EventStatus
    Limit     int
    Offset    int
}
```

### Service Implementation

```go
type service struct {
    repo      repositories.EventRepository
    validator Validator
    authz     auth.PermissionChecker
}

func NewService(
    repo repositories.EventRepository,
    validator Validator,
    authz auth.PermissionChecker,
) Service {
    return &service{
        repo:      repo,
        validator: validator,
        authz:     authz,
    }
}
```

### Permission Requirements

- **CreateEvent**: Requires `event:create` permission (EventManager or Admin)
- **GetEvent**: Requires `event:read` permission (EventManager or Admin)
- **UpdateEvent**: Requires `event:update` permission + ownership check
- **DeleteEvent**: Requires `event:delete` permission + ownership check
- **PublishEvent**: Requires `event:publish` permission + ownership check
- **CancelEvent**: Requires `event:cancel` permission + ownership check
- **ArchiveEvent**: Admin only

---

## Tasks

### Phase 1: Service Setup (TDD)
- [ ] Write test for service constructor
- [ ] Write test for dependency injection
- [ ] Implement NewService constructor
- [ ] Run tests (should pass)

### Phase 2: Create Event (TDD)
- [ ] Write test for creating event as EventManager
- [ ] Write test for creating event as Admin
- [ ] Write test for creating event without permission
- [ ] Write test for creating event with validation error
- [ ] Write test for creating event with repository error
- [ ] Write test for user context extraction
- [ ] Implement CreateEvent method
- [ ] Run tests (should pass)

### Phase 3: Get Event (TDD)
- [ ] Write test for getting existing event
- [ ] Write test for getting non-existent event
- [ ] Write test for getting event without permission
- [ ] Write test for getting own event
- [ ] Write test for admin getting any event
- [ ] Implement GetEvent method
- [ ] Run tests (should pass)

### Phase 4: Update Event (TDD)
- [ ] Write test for updating own event
- [ ] Write test for updating event as admin
- [ ] Write test for updating other's event (should fail)
- [ ] Write test for update with validation error
- [ ] Write test for update with version conflict
- [ ] Write test for updating published event
- [ ] Write test for updating cancelled event
- [ ] Implement UpdateEvent method
- [ ] Run tests (should pass)

### Phase 5: List Events (TDD)
- [ ] Write test for listing all events as admin
- [ ] Write test for listing own events as manager
- [ ] Write test for listing with status filter
- [ ] Write test for listing with pagination
- [ ] Write test for listing without permission
- [ ] Implement ListEvents method
- [ ] Run tests (should pass)

### Phase 6: State Transitions (TDD)
- [ ] Write test for publishing draft event
- [ ] Write test for publishing already published event
- [ ] Write test for publishing without permission
- [ ] Write test for cancelling published event
- [ ] Write test for cancelling with reason
- [ ] Write test for cancelling without permission
- [ ] Write test for archiving old event
- [ ] Write test for archiving recent event
- [ ] Implement PublishEvent method
- [ ] Implement CancelEvent method
- [ ] Implement ArchiveEvent method
- [ ] Run tests (should pass)

### Phase 7: Integration Tests
- [ ] Write integration test for full event lifecycle
- [ ] Write integration test for concurrent updates
- [ ] Write integration test for permission enforcement
- [ ] Run integration tests

---

## Testing Requirements

### Unit Tests

```go
func TestEventService_CreateEvent(t *testing.T) {
    tests := []struct {
        name    string
        user    *models.User
        event   *models.Event
        wantErr bool
        errMsg  string
    }{
        {
            name: "event manager creates event",
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
            wantErr: false,
        },
        {
            name: "admin creates event",
            user: &models.User{
                ID:   1,
                Role: models.RoleAdmin,
            },
            event: &models.Event{
                Title:       "Admin Event",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 0,
            },
            wantErr: false,
        },
        {
            name: "guest cannot create event",
            user: &models.User{
                ID:   1,
                Role: models.RoleGuest,
            },
            event: &models.Event{
                Title:       "Guest Event",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 0,
            },
            wantErr: true,
            errMsg:  "permission denied",
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
            wantErr: true,
            errMsg:  "validation error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &repositories.MockEventRepository{}
            mockValidator := &MockValidator{
                ValidateCreateFunc: func(ctx context.Context, e *models.Event) error {
                    if len(e.Title) < 3 {
                        return &models.ValidationError{Field: "title", Message: "too short"}
                    }
                    return nil
                },
            }
            mockAuthz := &auth.MockPermissionChecker{
                HasPermissionFunc: func(ctx context.Context, perm string) bool {
                    return tt.user.Role == models.RoleAdmin || 
                           tt.user.Role == models.RoleEventManager
                },
            }
            
            service := NewService(mockRepo, mockValidator, mockAuthz)
            
            ctx := auth.WithUser(context.Background(), tt.user)
            err := service.CreateEvent(ctx, tt.event)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateEvent() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && err != nil && tt.errMsg != "" {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
                }
            }
            
            if !tt.wantErr {
                if tt.event.CreatedBy != tt.user.ID {
                    t.Errorf("CreatedBy = %d, want %d", tt.event.CreatedBy, tt.user.ID)
                }
                if tt.event.Status != models.EventStatusDraft {
                    t.Errorf("Status = %q, want %q", tt.event.Status, models.EventStatusDraft)
                }
                if tt.event.Version != 1 {
                    t.Errorf("Version = %d, want 1", tt.event.Version)
                }
            }
        })
    }
}

func TestEventService_UpdateEvent(t *testing.T) {
    tests := []struct {
        name         string
        user         *models.User
        existingUser int64
        event        *models.Event
        wantErr      bool
        errMsg       string
    }{
        {
            name: "owner updates own event",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            existingUser: 1,
            event: &models.Event{
                ID:          1,
                Title:       "Updated Title",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 0,
                Version:     1,
            },
            wantErr: false,
        },
        {
            name: "admin updates any event",
            user: &models.User{
                ID:   2,
                Role: models.RoleAdmin,
            },
            existingUser: 1,
            event: &models.Event{
                ID:          1,
                Title:       "Admin Update",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 0,
                Version:     1,
            },
            wantErr: false,
        },
        {
            name: "non-owner cannot update",
            user: &models.User{
                ID:   2,
                Role: models.RoleEventManager,
            },
            existingUser: 1,
            event: &models.Event{
                ID:          1,
                Title:       "Unauthorized Update",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 0,
                Version:     1,
            },
            wantErr: true,
            errMsg:  "permission denied",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &repositories.MockEventRepository{
                GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
                    return &models.Event{
                        ID:        id,
                        CreatedBy: tt.existingUser,
                        Version:   1,
                    }, nil
                },
                UpdateWithVersionFunc: func(ctx context.Context, e *models.Event, v int) error {
                    return nil
                },
            }
            mockValidator := &MockValidator{}
            mockAuthz := &auth.MockPermissionChecker{
                HasPermissionFunc: func(ctx context.Context, perm string) bool {
                    return true
                },
                CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
                    return user.Role == models.RoleAdmin || user.ID == event.CreatedBy
                },
            }
            
            service := NewService(mockRepo, mockValidator, mockAuthz)
            
            ctx := auth.WithUser(context.Background(), tt.user)
            err := service.UpdateEvent(ctx, tt.event)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("UpdateEvent() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestEventService_PublishEvent(t *testing.T) {
    tests := []struct {
        name         string
        user         *models.User
        eventStatus  models.EventStatus
        eventOwner   int64
        wantErr      bool
        errMsg       string
    }{
        {
            name: "publish draft event",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            eventStatus: models.EventStatusDraft,
            eventOwner:  1,
            wantErr:     false,
        },
        {
            name: "cannot publish already published",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            eventStatus: models.EventStatusPublished,
            eventOwner:  1,
            wantErr:     true,
            errMsg:      "invalid state transition",
        },
        {
            name: "cannot publish cancelled",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            eventStatus: models.EventStatusCancelled,
            eventOwner:  1,
            wantErr:     true,
            errMsg:      "invalid state transition",
        },
        {
            name: "non-owner cannot publish",
            user: &models.User{
                ID:   2,
                Role: models.RoleEventManager,
            },
            eventStatus: models.EventStatusDraft,
            eventOwner:  1,
            wantErr:     true,
            errMsg:      "permission denied",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &repositories.MockEventRepository{
                GetByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
                    return &models.Event{
                        ID:        id,
                        Status:    tt.eventStatus,
                        CreatedBy: tt.eventOwner,
                    }, nil
                },
                UpdateStatusFunc: func(ctx context.Context, id int64, status models.EventStatus) error {
                    return nil
                },
            }
            mockValidator := &MockValidator{
                ValidateStateTransitionFunc: func(from, to models.EventStatus) error {
                    if from == models.EventStatusDraft && to == models.EventStatusPublished {
                        return nil
                    }
                    return fmt.Errorf("invalid state transition")
                },
            }
            mockAuthz := &auth.MockPermissionChecker{
                CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
                    return user.Role == models.RoleAdmin || user.ID == event.CreatedBy
                },
            }
            
            service := NewService(mockRepo, mockValidator, mockAuthz)
            
            ctx := auth.WithUser(context.Background(), tt.user)
            err := service.PublishEvent(ctx, 1)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("PublishEvent() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- 02_STORY_01_event_model.md - Event model and validation
- 02_STORY_02_event_repository.md - Event repository
- Epic 01 (Auth) - Permission checking

**Blocks:** 
- Event HTTP handlers
- Event lifecycle operations

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/events/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Permission checks verified
- [ ] State transitions validated
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### Permission Checking

The service layer enforces permissions before any operation:
1. Extract user from context
2. Check if user has required permission
3. For update/delete/publish/cancel, verify ownership or admin role
4. Proceed with operation if authorized

### State Management

State transitions are validated before updating:
1. Retrieve current event state
2. Validate transition using validator
3. Update status if valid
4. Return error if invalid

### Error Handling

Service layer returns domain-specific errors:
- `PermissionDeniedError` for authorization failures
- `ValidationError` for validation failures
- `VersionConflictError` for optimistic locking conflicts
- `NotFoundError` for missing resources

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **HLD:** Section 5 (Event Model), Section 6 (Authorization)
- **LLD:** [lld/02_EVENT_LLD.md](../lld/02_EVENT_LLD.md) - Section 4.1
- **Epic:** [02_EPIC_events.md](02_EPIC_events.md)
