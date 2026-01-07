# Events Package

## Purpose

This package provides event validation logic for the TinyRSVP application, ensuring event data integrity before persistence.

## Rules

- All event data MUST be validated before database operations
- Use the Validator interface for all validation operations
- Timezone validation uses Go's standard IANA timezone database
- State transitions follow a strict state machine model

## Structure

- `validator.go` - Main event validator implementation
- `validator_test.go` - Comprehensive validation tests
- `timezone_validator.go` - IANA timezone validation
- `timezone_validator_test.go` - Timezone validation tests

## Key Components

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
- COMPLETED → ARCHIVED

### Invalid Transitions
- Any state → DRAFT (cannot revert to draft)
- ARCHIVED → Any state (archived is final)
- CANCELLED → PUBLISHED (cannot un-cancel)
- COMPLETED → PUBLISHED (cannot un-complete)

## Usage

```go
import (
    "github.com/lenaxia/tinyrsvp/internal/events"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

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
