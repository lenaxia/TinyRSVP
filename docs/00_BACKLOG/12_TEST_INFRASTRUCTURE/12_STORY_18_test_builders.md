# User Story: Create Test Data Builders

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 2-3 hours
**Phase:** 5 - Advanced Features

---

## User Story

As a **developer**, I want **fluent test data builders for complex objects** so that **I can create test data with readable, maintainable code**.

---

## Acceptance Criteria

- [ ] EventBuilder implemented
- [ ] InviteBuilder implemented
- [ ] UserBuilder implemented
- [ ] RSVPBuilder implemented
- [ ] All builders have tests
- [ ] Documentation with examples

---

## Implementation

### EventBuilder

```go
package builders

type EventBuilder struct {
    t     *testing.T
    event *models.Event
}

func NewEventBuilder(t *testing.T) *EventBuilder {
    // Returns builder with sensible defaults
}

func (b *EventBuilder) WithTitle(title string) *EventBuilder
func (b *EventBuilder) WithStatus(status models.EventStatus) *EventBuilder
func (b *EventBuilder) WithCreator(creatorID int64) *EventBuilder
func (b *EventBuilder) Published() *EventBuilder
func (b *EventBuilder) InFuture() *EventBuilder
func (b *EventBuilder) InPast() *EventBuilder
func (b *EventBuilder) WithCapacity(capacity int) *EventBuilder
func (b *EventBuilder) Build() *models.Event
func (b *EventBuilder) BuildAndCreate(repo repositories.EventRepository) *models.Event
```

### Usage Example

```go
event := builders.NewEventBuilder(t).
    WithTitle("Birthday Party").
    Published().
    InFuture().
    WithCapacity(50).
    Build()
```

---

## Builders to Create

1. **EventBuilder** - Fluent event construction
2. **InviteBuilder** - Fluent invite construction
3. **UserBuilder** - Fluent user construction
4. **RSVPBuilder** - Fluent RSVP construction

---

## Tasks

### Task 1: EventBuilder
- [ ] Write tests for EventBuilder
- [ ] Implement EventBuilder
- [ ] Add helper methods (Published, InFuture, etc.)

### Task 2: InviteBuilder
- [ ] Write tests for InviteBuilder
- [ ] Implement InviteBuilder
- [ ] Add helper methods

### Task 3: UserBuilder
- [ ] Write tests for UserBuilder
- [ ] Implement UserBuilder
- [ ] Add role helpers (Admin, EventManager)

### Task 4: RSVPBuilder
- [ ] Write tests for RSVPBuilder
- [ ] Implement RSVPBuilder
- [ ] Add response helpers (Yes, No, Maybe)

### Task 5: Documentation
- [ ] Update testutil/README.md
- [ ] Add usage examples
- [ ] Document builder patterns

---

## Dependencies

**Depends on:** Story 17 (Phase 4 complete)

---

## Benefits

**Before:**
```go
event := &models.Event{
    Title:       "Test Event",
    Slug:        "test-event-" + time.Now().Format("20060102150405"),
    Description: testutil.StringPtr("Test description"),
    Location:    testutil.StringPtr("Test location"),
    StartTime:   time.Now().Add(24 * time.Hour),
    EndTime:     time.Now().Add(26 * time.Hour),
    Timezone:    "UTC",
    Status:      models.EventStatusPublished,
    CreatedBy:   userID,
    MaxPlusOnes: 2,
    AllowMaybeRSVP: true,
}
```

**After:**
```go
event := builders.NewEventBuilder(t).
    WithTitle("Test Event").
    Published().
    InFuture().
    Build()
```

Much more readable and maintainable!
