# User Story: Event Model and Validation

**Epic:** [02_EPIC_events.md](02_EPIC_events.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 4 hours
**Completed:** 2026-01-07

---

## User Story

As a **developer**, I want **a complete Event model with validation logic** so that **event data is properly structured and validated before persistence**.

---

## Acceptance Criteria

- [x] Event struct matches database schema
- [x] EventStatus enum with all lifecycle states defined
- [x] Event validation rules implemented
- [x] Timezone validation using IANA database
- [x] Title validation (3-200 characters)
- [x] Date validation (start before end, future dates)
- [x] RSVP deadline validation (before start time)
- [x] Max plus ones validation (0-10)
- [x] All validation tests pass with timeout
- [x] Validation error messages are clear and actionable

---

## Technical Details

### Event Model Structure

The Event model already exists in [`internal/models/event.go`](../../internal/models/event.go) with the following structure:

```go
type Event struct {
    ID           int64       `db:"id" json:"id"`
    Title        string      `db:"title" json:"title"`
    Description  *string     `db:"description" json:"description,omitempty"`
    StartTime    time.Time   `db:"start_time" json:"start_time"`
    EndTime      *time.Time  `db:"end_time" json:"end_time,omitempty"`
    Timezone     string      `db:"timezone" json:"timezone"`
    Location     *string     `db:"location" json:"location,omitempty"`
    Status       EventStatus `db:"status" json:"status"`
    CreatedBy    int64       `db:"created_by" json:"created_by"`
    Version      int         `db:"version" json:"version"`
    ICSSequence  int         `db:"ics_sequence" json:"ics_sequence"`
    MaxPlusOnes  int         `db:"max_plus_ones" json:"max_plus_ones"`
    RSVPDeadline *time.Time  `db:"rsvp_deadline" json:"rsvp_deadline,omitempty"`
    CreatedAt    time.Time   `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time   `db:"updated_at" json:"updated_at"`
}
```

### Event Status Enum

```go
type EventStatus string

const (
    EventStatusDraft     EventStatus = "draft"
    EventStatusPublished EventStatus = "published"
    EventStatusCancelled EventStatus = "cancelled"
    EventStatusArchived  EventStatus = "archived"
)
```

### Validator Interface

```go
package events

type Validator interface {
    ValidateCreate(ctx context.Context, event *models.Event) error
    ValidateUpdate(ctx context.Context, event *models.Event) error
    ValidateStateTransition(from, to models.EventStatus) error
}
```

### Timezone Validator Interface

```go
package events

type TimezoneValidator interface {
    IsValid(timezone string) bool
    GetLocation(timezone string) (*time.Location, error)
}
```

### Validation Rules

**Title:**
- Required
- 3-200 characters
- No leading/trailing whitespace
- Sanitized for XSS

**Description:**
- Optional
- Max 5000 characters

**Start Time:**
- Required
- Must be in future (at creation)
- Stored with timezone

**End Time:**
- Optional
- Must be after start time if provided
- Must be within 7 days of start time

**Timezone:**
- Required
- Must be valid IANA timezone (e.g., "America/Los_Angeles")
- Validated against Go's timezone database

**RSVP Deadline:**
- Optional
- Must be before start time
- Must be in future (at creation)

**Max Plus Ones:**
- Integer 0-10
- Default: 0

**Location:**
- Optional
- Max 500 characters

---

## Tasks

### Phase 1: Timezone Validator (TDD)
- [x] Write test for valid IANA timezones
- [x] Write test for invalid timezones
- [x] Write test for GetLocation with valid timezone
- [x] Write test for GetLocation with invalid timezone
- [x] Implement TimezoneValidator using time.LoadLocation
- [x] Run tests (should pass)

### Phase 2: Event Validator - Create (TDD)
- [x] Write test for valid event creation
- [x] Write test for title too short
- [x] Write test for title too long
- [x] Write test for missing title
- [x] Write test for description too long
- [x] Write test for start time in past
- [x] Write test for end time before start time
- [x] Write test for end time more than 7 days after start
- [x] Write test for invalid timezone
- [x] Write test for RSVP deadline after start time
- [x] Write test for RSVP deadline in past
- [x] Write test for max plus ones negative
- [x] Write test for max plus ones over 10
- [x] Implement ValidateCreate method
- [x] Run tests (should pass)

### Phase 3: Event Validator - Update (TDD)
- [x] Write test for valid event update
- [x] Write test for updating published event dates
- [x] Write test for updating cancelled event
- [x] Write test for updating archived event
- [x] Write test for version mismatch
- [x] Implement ValidateUpdate method
- [x] Run tests (should pass)

### Phase 4: State Transition Validator (TDD)
- [x] Write test for draft → published
- [x] Write test for draft → cancelled
- [x] Write test for published → cancelled
- [x] Write test for published → archived
- [x] Write test for cancelled → archived
- [x] Write test for invalid transitions
- [x] Write test for archived → any (should fail)
- [x] Implement ValidateStateTransition method
- [x] Run tests (should pass)

### Phase 5: Integration
- [x] Create validator constructor
- [x] Wire timezone validator into event validator
- [x] Add validation error types to models package
- [x] Document validation rules in README
- [x] Run all tests with timeout

---

## Testing Requirements

### Unit Tests

```go
func TestTimezoneValidator_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        timezone string
        want     bool
    }{
        {"valid US timezone", "America/Los_Angeles", true},
        {"valid EU timezone", "Europe/London", true},
        {"valid Asia timezone", "Asia/Tokyo", true},
        {"invalid timezone", "Invalid/Timezone", false},
        {"empty timezone", "", false},
        {"UTC", "UTC", true},
    }
    
    validator := NewTimezoneValidator()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := validator.IsValid(tt.timezone)
            if got != tt.want {
                t.Errorf("IsValid(%q) = %v, want %v", tt.timezone, got, tt.want)
            }
        })
    }
}

func TestEventValidator_ValidateCreate(t *testing.T) {
    tests := []struct {
        name    string
        event   *models.Event
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid event",
            event: &models.Event{
                Title:       "Birthday Party",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 2,
            },
            wantErr: false,
        },
        {
            name: "title too short",
            event: &models.Event{
                Title:     "AB",
                StartTime: time.Now().Add(24 * time.Hour),
                Timezone:  "America/Los_Angeles",
            },
            wantErr: true,
            errMsg:  "title must be between 3 and 200 characters",
        },
        {
            name: "start time in past",
            event: &models.Event{
                Title:     "Past Event",
                StartTime: time.Now().Add(-24 * time.Hour),
                Timezone:  "America/Los_Angeles",
            },
            wantErr: true,
            errMsg:  "start time must be in the future",
        },
        {
            name: "invalid timezone",
            event: &models.Event{
                Title:     "Event",
                StartTime: time.Now().Add(24 * time.Hour),
                Timezone:  "Invalid/Zone",
            },
            wantErr: true,
            errMsg:  "invalid timezone",
        },
        {
            name: "end time before start time",
            event: &models.Event{
                Title:     "Event",
                StartTime: time.Now().Add(24 * time.Hour),
                EndTime:   timePtr(time.Now().Add(12 * time.Hour)),
                Timezone:  "America/Los_Angeles",
            },
            wantErr: true,
            errMsg:  "end time must be after start time",
        },
        {
            name: "RSVP deadline after start",
            event: &models.Event{
                Title:        "Event",
                StartTime:    time.Now().Add(24 * time.Hour),
                RSVPDeadline: timePtr(time.Now().Add(48 * time.Hour)),
                Timezone:     "America/Los_Angeles",
            },
            wantErr: true,
            errMsg:  "RSVP deadline must be before event start time",
        },
        {
            name: "max plus ones negative",
            event: &models.Event{
                Title:       "Event",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: -1,
            },
            wantErr: true,
            errMsg:  "max plus ones must be between 0 and 10",
        },
        {
            name: "max plus ones over limit",
            event: &models.Event{
                Title:       "Event",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                MaxPlusOnes: 11,
            },
            wantErr: true,
            errMsg:  "max plus ones must be between 0 and 10",
        },
    }
    
    validator := NewValidator(NewTimezoneValidator())
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateCreate(context.Background(), tt.event)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && err != nil {
                if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
                    t.Errorf("ValidateCreate() error message = %q, want to contain %q", err.Error(), tt.errMsg)
                }
            }
        })
    }
}

func TestEventValidator_ValidateStateTransition(t *testing.T) {
    tests := []struct {
        name    string
        from    models.EventStatus
        to      models.EventStatus
        wantErr bool
    }{
        {"draft to published", models.EventStatusDraft, models.EventStatusPublished, false},
        {"draft to cancelled", models.EventStatusDraft, models.EventStatusCancelled, false},
        {"published to cancelled", models.EventStatusPublished, models.EventStatusCancelled, false},
        {"published to archived", models.EventStatusPublished, models.EventStatusArchived, false},
        {"cancelled to archived", models.EventStatusCancelled, models.EventStatusArchived, false},
        {"archived to published", models.EventStatusArchived, models.EventStatusPublished, true},
        {"archived to cancelled", models.EventStatusArchived, models.EventStatusCancelled, true},
        {"published to draft", models.EventStatusPublished, models.EventStatusDraft, true},
        {"cancelled to published", models.EventStatusCancelled, models.EventStatusPublished, true},
    }
    
    validator := NewValidator(NewTimezoneValidator())
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateStateTransition(tt.from, tt.to)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateStateTransition(%v, %v) error = %v, wantErr %v", 
                    tt.from, tt.to, err, tt.wantErr)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- Epic 00 (Foundation) - Complete
- Epic 01 (Auth) - Complete

**Blocks:** 
- All other Event Management stories

**External Dependencies:**
- Go standard library `time` package for timezone validation

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/events/...`)
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Validation rules documented
- [x] Changes committed to git

---

## Implementation Notes

### Timezone Validation

Use Go's standard library for timezone validation:

```go
func (v *timezoneValidator) IsValid(timezone string) bool {
    _, err := time.LoadLocation(timezone)
    return err == nil
}

func (v *timezoneValidator) GetLocation(timezone string) (*time.Location, error) {
    return time.LoadLocation(timezone)
}
```

### Validation Error Types

Create custom error types in [`internal/models/errors.go`](../../internal/models/errors.go):

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}
```

### State Transition Rules

Valid transitions:
- DRAFT → PUBLISHED (manual publish)
- DRAFT → CANCELLED (manual cancel)
- PUBLISHED → CANCELLED (manual cancel)
- PUBLISHED → ARCHIVED (auto after 30 days)
- CANCELLED → ARCHIVED (auto after 30 days)

Invalid transitions:
- Any state → DRAFT (cannot revert to draft)
- ARCHIVED → Any state (archived is final)
- CANCELLED → PUBLISHED (cannot un-cancel)

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **HLD:** Section 5 (Event Model)
- **LLD:** [lld/02_EVENT_LLD.md](../lld/02_EVENT_LLD.md) - Section 4.2
- **Epic:** [02_EPIC_events.md](02_EPIC_events.md)
