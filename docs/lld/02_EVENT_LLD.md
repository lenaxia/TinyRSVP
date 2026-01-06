# Domain 2: Event Management - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 5 - Event Model](../02_REVISED_HLD.md#5-event-model)

---

## 1. Overview

### 1.1 Purpose

Manages event lifecycle including creation, updates, state transitions, validation, and timezone handling.

### 1.2 Responsibilities

- Event CRUD operations
- Event lifecycle state machine (draft → published → cancelled → archived)
- Event validation (title, dates, timezone)
- Optimistic locking for concurrent updates
- Event ownership and permission enforcement
- ICS sequence management for calendar updates
- Auto-archiving of past events

### 1.3 Design Principles

- **State Machine** - Explicit state transitions
- **Optimistic Locking** - Version-based concurrency control
- **Timezone Aware** - IANA timezone validation
- **Permission Enforced** - Check ownership on all mutations
- **Audit Logged** - Track all changes

---

## 2. Package Structure

```
internal/
├── events/
│   ├── service.go              # Event service implementation
│   ├── service_test.go
│   ├── validator.go            # Event validation
│   ├── validator_test.go
│   ├── lifecycle.go            # State machine logic
│   ├── lifecycle_test.go
│   └── timezone.go             # Timezone utilities
│       └── timezone_test.go
```

---

## 3. Interfaces

### 3.1 Event Service Interface

```go
package events

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
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

### 3.2 Event Validator Interface

```go
package events

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type Validator interface {
    ValidateCreate(ctx context.Context, event *models.Event) error
    ValidateUpdate(ctx context.Context, event *models.Event) error
    ValidateStateTransition(from, to models.EventStatus) error
}
```

---

## 4. Implementation

### 4.1 Event Service

```go
package events

import (
    "context"
    "fmt"
    "time"
    
    "github.com/yourusername/tinyrsvp/internal/db/repositories"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type service struct {
    repo      repositories.EventRepository
    validator Validator
    authz     AuthorizationChecker
}

func NewService(repo repositories.EventRepository, validator Validator, authz AuthorizationChecker) Service {
    return &service{
        repo:      repo,
        validator: validator,
        authz:     authz,
    }
}

func (s *service) CreateEvent(ctx context.Context, event *models.Event) error {
    user, _ := auth.UserFromContext(ctx)
    
    if !s.authz.CanCreateEvent(ctx, user) {
        return fmt.Errorf("permission denied")
    }
    
    if err := s.validator.ValidateCreate(ctx, event); err != nil {
        return err
    }
    
    event.CreatedBy = user.ID
    event.Status = models.EventStatusDraft
    event.Version = 1
    event.ICSSequence = 0
    
    return s.repo.Create(ctx, event)
}

func (s *service) UpdateEvent(ctx context.Context, event *models.Event) error {
    user, _ := auth.UserFromContext(ctx)
    
    existing, err := s.repo.GetByID(ctx, event.ID)
    if err != nil {
        return err
    }
    
    if !s.authz.CanEditEvent(ctx, user, existing) {
        return fmt.Errorf("permission denied")
    }
    
    if err := s.validator.ValidateUpdate(ctx, event); err != nil {
        return err
    }
    
    return s.repo.UpdateWithVersion(ctx, event, existing.Version)
}

func (s *service) PublishEvent(ctx context.Context, id int64) error {
    event, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    if err := s.validator.ValidateStateTransition(event.Status, models.EventStatusPublished); err != nil {
        return err
    }
    
    return s.repo.UpdateStatus(ctx, id, models.EventStatusPublished)
}
```

### 4.2 Event Validator

```go
package events

import (
    "context"
    "fmt"
    "time"
    
    "github.com/yourusername/tinyrsvp/internal/models"
)

type validator struct {
    tzValidator TimezoneValidator
}

func NewValidator(tzValidator TimezoneValidator) Validator {
    return &validator{tzValidator: tzValidator}
}

func (v *validator) ValidateCreate(ctx context.Context, event *models.Event) error {
    if len(event.Title) < 3 || len(event.Title) > 200 {
        return &models.ValidationError{
            Field:   "title",
            Message: "Event title must be between 3 and 200 characters",
        }
    }
    
    if event.Description != nil && len(*event.Description) > 5000 {
        return &models.ValidationError{
            Field:   "description",
            Message: "Description cannot exceed 5000 characters",
        }
    }
    
    if event.StartTime.Before(time.Now()) {
        return &models.ValidationError{
            Field:   "start_time",
            Message: "Event start time must be in the future",
        }
    }
    
    if event.EndTime != nil && event.EndTime.Before(event.StartTime) {
        return &models.ValidationError{
            Field:   "end_time",
            Message: "Event end time must be after start time",
        }
    }
    
    if !v.tzValidator.IsValid(event.Timezone) {
        return &models.ValidationError{
            Field:   "timezone",
            Message: "Invalid timezone. Use IANA format like 'America/Los_Angeles'",
        }
    }
    
    if event.RSVPDeadline != nil && event.RSVPDeadline.After(event.StartTime) {
        return &models.ValidationError{
            Field:   "rsvp_deadline",
            Message: "RSVP deadline must be before event start time",
        }
    }
    
    if event.MaxPlusOnes < 0 || event.MaxPlusOnes > 10 {
        return &models.ValidationError{
            Field:   "max_plus_ones",
            Message: "Max plus ones must be between 0 and 10",
        }
    }
    
    return nil
}
```

---

## 5. State Machine

```
DRAFT → PUBLISHED → CANCELLED → ARCHIVED
  ↓                      ↓
CANCELLED            ARCHIVED
```

**Transitions:**
- DRAFT → PUBLISHED: Manual publish
- DRAFT → CANCELLED: Manual cancel
- PUBLISHED → CANCELLED: Manual cancel
- PUBLISHED → ARCHIVED: Auto (30 days after event)
- CANCELLED → ARCHIVED: Auto (30 days after event)

---

## 6. Dependencies

**Internal:**
- Domain 1 (Auth) - Permission checking
- Domain 7 (Database) - Event repository

**Dependents:**
- Domain 3 (Invite) - Invites belong to events
- Domain 4 (RSVP) - Questions belong to events
- Domain 5 (Email) - Event details in emails

---

## 7. Testing Strategy

```go
func TestEventService_CreateEvent(t *testing.T) {
    tests := []struct {
        name    string
        event   *models.Event
        user    *models.User
        wantErr bool
    }{
        {
            name: "valid event",
            event: &models.Event{
                Title:       "Birthday Party",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 2,
            },
            user:    &models.User{Role: models.RoleEventManager},
            wantErr: false,
        },
        {
            name: "past start time",
            event: &models.Event{
                Title:     "Past Event",
                StartTime: time.Now().Add(-24 * time.Hour),
                Timezone:  "America/Los_Angeles",
            },
            user:    &models.User{Role: models.RoleEventManager},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := auth.WithUser(context.Background(), tt.user)
            err := service.CreateEvent(ctx, tt.event)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateEvent() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

**Document Status:** ✅ Complete

**Next Domain:** [Domain 3: Invite & Token Management](03_INVITE_LLD.md)
