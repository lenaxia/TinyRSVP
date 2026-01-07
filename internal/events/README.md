# Events Package

## Purpose

This package provides event business logic for the TinyRSVP application, including validation, authorization, and state management.

## Rules

- All event operations MUST go through the Service layer
- All event data MUST be validated before database operations
- All operations MUST check permissions before execution
- State transitions follow a strict state machine model
- Use the Validator interface for all validation operations
- Timezone validation uses Go's standard IANA timezone database

## Structure

- `service.go` - Event service layer with business logic
- `service_test.go` - Comprehensive service tests
- `validator.go` - Main event validator implementation
- `validator_test.go` - Comprehensive validation tests
- `timezone_validator.go` - IANA timezone validation
- `timezone_validator_test.go` - Timezone validation tests

## Key Components

### Service Interface

```go
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
```

The service layer enforces:
- Permission checks on all operations
- Event validation before persistence
- State transition validation
- Optimistic locking for updates
- User context extraction and authorization

### Validator Interface

```go
type Validator interface {
    ValidateCreate(ctx context.Context, event *models.Event) error
    ValidateUpdate(ctx context.Context, event *models.Event) error
    ValidateStateTransition(from, to models.EventStatus) error
}
```

### TimezoneValidator Interface

```go
type TimezoneValidator interface {
    IsValid(timezone string) bool
    GetLocation(timezone string) (*time.Location, error)
}
```

## Validation Rules

### Title
- Required
- 3-200 characters
- No leading/trailing whitespace
- XSS Protection: Sanitization occurs at the template rendering layer using Go's html/template package, which automatically escapes HTML entities. Input validation focuses on length and format constraints.

### Description
- Optional
- Max 5000 characters

### Start Time
- Required
- Must be in future (at creation)

### End Time
- Optional
- Must be after start time
- Must be within 7 days of start time

### Timezone
- Required
- Must be valid IANA timezone (e.g., "America/Los_Angeles")

### RSVP Deadline
- Optional
- Must be before start time
- Must be in future (at creation)

### Max Plus Ones
- Integer 0-10
- Default: 0

### Location
- Optional
- Max 500 characters

## State Transition Rules

### Valid Transitions
- DRAFT → PUBLISHED
- DRAFT → CANCELLED
- PUBLISHED → CANCELLED
- PUBLISHED → ARCHIVED
- CANCELLED → ARCHIVED

### Invalid Transitions
- Any state → DRAFT (cannot revert to draft)
- ARCHIVED → Any state (archived is final)
- CANCELLED → PUBLISHED (cannot un-cancel)

## Update Validation Behavior

### Start Time Validation
ValidateUpdate intentionally does NOT validate that start time is in the future. This allows:
- Updating published events that have already started or passed
- Modifying event details after the event has occurred
- Correcting event information retroactively

The start time future validation only applies during event creation (ValidateCreate) to prevent creating events in the past.

## Usage

### Service Layer

```go
import (
    "github.com/lenaxia/tinyrsvp/internal/events"
    "github.com/lenaxia/tinyrsvp/internal/auth"
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
)

repo := repositories.NewEventRepository(db)
tzValidator := events.NewTimezoneValidator()
validator := events.NewValidator(tzValidator)
authz := auth.NewAuthorizationChecker()

service := events.NewService(repo, validator, authz)

ctx := auth.WithUser(context.Background(), user)

event := &models.Event{
    Title:       "Birthday Party",
    StartTime:   time.Now().Add(24 * time.Hour),
    Timezone:    "America/Los_Angeles",
    MaxPlusOnes: 2,
}

if err := service.CreateEvent(ctx, event); err != nil {
    return err
}

if err := service.PublishEvent(ctx, event.ID); err != nil {
    return err
}

events, err := service.ListEvents(ctx, events.ListFilters{
    Limit: 10,
})
if err != nil {
    return err
}
```

### Validator Only

```go
tzValidator := events.NewTimezoneValidator()
validator := events.NewValidator(tzValidator)

event := &models.Event{
    Title:       "Birthday Party",
    StartTime:   time.Now().Add(24 * time.Hour),
    Timezone:    "America/Los_Angeles",
    MaxPlusOnes: 2,
}

if err := validator.ValidateCreate(ctx, event); err != nil {
    return err
}

if err := validator.ValidateStateTransition(
    models.EventStatusDraft,
    models.EventStatusPublished,
); err != nil {
    return err
}
```

## Testing

Run tests with timeout:
```bash
go test -timeout 30s ./internal/events/...
```

Check coverage:
```bash
go test -timeout 30s -cover ./internal/events/...
```

Current coverage: 93.9%
