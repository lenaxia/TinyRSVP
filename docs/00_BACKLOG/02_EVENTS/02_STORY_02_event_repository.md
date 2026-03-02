# User Story: Event Repository

**Epic:** [02_EPIC_events.md](02_EPIC_events.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 6 hours
**Completed:** 2026-01-07

---

## User Story

As a **developer**, I want **a repository layer for event persistence** so that **events can be stored, retrieved, and updated in the database with proper concurrency control**.

---

## Acceptance Criteria

- [x] EventRepository interface defined
- [x] Create event operation implemented
- [x] Get event by ID operation implemented
- [x] Update event with optimistic locking implemented
- [x] Delete event (soft delete) operation implemented
- [x] List events with filters implemented
- [x] Get events by status implemented
- [x] Get events to archive implemented
- [x] Version conflict detection working
- [x] All database operations use transactions
- [x] All tests pass with timeout
- [x] Integration tests with real database pass

---

## Technical Details

### EventRepository Interface

```go
package repositories

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type EventRepository interface {
    Create(ctx context.Context, event *models.Event) error
    GetByID(ctx context.Context, id int64) (*models.Event, error)
    Update(ctx context.Context, event *models.Event) error
    UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error
    UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, filters ListFilters) ([]*models.Event, error)
    GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error)
    GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error)
}

type ListFilters struct {
    CreatorID *int64
    Status    *models.EventStatus
    Limit     int
    Offset    int
}
```

### Database Schema

Events table already exists from migrations:

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    timezone TEXT NOT NULL,
    location TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    created_by INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    ics_sequence INTEGER NOT NULL DEFAULT 0,
    max_plus_ones INTEGER NOT NULL DEFAULT 0,
    rsvp_deadline DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_start_time ON events(start_time);
```

### Optimistic Locking Implementation

```go
func (r *eventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
    query := `
        UPDATE events 
        SET title = ?, description = ?, start_time = ?, end_time = ?,
            timezone = ?, location = ?, max_plus_ones = ?, rsvp_deadline = ?,
            version = version + 1, updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND version = ?
    `
    
    result, err := r.db.ExecContext(ctx, query,
        event.Title, event.Description, event.StartTime, event.EndTime,
        event.Timezone, event.Location, event.MaxPlusOnes, event.RSVPDeadline,
        event.ID, expectedVersion,
    )
    
    if err != nil {
        return err
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rows == 0 {
        return &models.VersionConflictError{
            ResourceType: "event",
            ResourceID:   event.ID,
            Expected:     expectedVersion,
        }
    }
    
    return nil
}
```

---

## Tasks

### Phase 1: Repository Setup (TDD)
- [x] Write test for repository constructor
- [x] Write test for database connection validation
- [x] Implement NewEventRepository constructor
- [x] Run tests (should pass)

### Phase 2: Create Operation (TDD)
- [x] Write test for creating valid event
- [x] Write test for creating event with all optional fields
- [x] Write test for creating event with missing required fields
- [x] Write test for database constraint violations
- [x] Write test for transaction rollback on error
- [x] Implement Create method
- [x] Run tests (should pass)

### Phase 3: Read Operations (TDD)
- [x] Write test for GetByID with existing event
- [x] Write test for GetByID with non-existent event
- [x] Write test for GetByID with invalid ID
- [x] Write test for List with no filters
- [x] Write test for List with creator filter
- [x] Write test for List with status filter
- [x] Write test for List with pagination
- [x] Write test for GetByStatus
- [x] Implement GetByID method
- [x] Implement List method
- [x] Implement GetByStatus method
- [x] Run tests (should pass)

### Phase 4: Update Operations (TDD)
- [x] Write test for Update with valid changes
- [x] Write test for UpdateWithVersion success
- [x] Write test for UpdateWithVersion conflict
- [x] Write test for UpdateWithVersion non-existent event
- [x] Write test for UpdateStatus
- [x] Write test for concurrent updates
- [x] Implement Update method
- [x] Implement UpdateWithVersion method
- [x] Implement UpdateStatus method
- [x] Run tests (should pass)

### Phase 5: Delete and Archive Operations (TDD)
- [x] Write test for Delete (soft delete)
- [x] Write test for Delete non-existent event
- [x] Write test for GetEventsToArchive
- [x] Write test for GetEventsToArchive with no events
- [x] Write test for GetEventsToArchive date filtering
- [x] Implement Delete method
- [x] Implement GetEventsToArchive method
- [x] Run tests (should pass)

### Phase 6: Integration Tests
- [x] Write integration test for full CRUD cycle
- [x] Write integration test for concurrent updates
- [x] Write integration test for transaction rollback
- [x] Write integration test for foreign key constraints
- [x] Run integration tests with real database

---

## Testing Requirements

### Unit Tests

```go
func TestEventRepository_Create(t *testing.T) {
    tests := []struct {
        name    string
        event   *models.Event
        wantErr bool
    }{
        {
            name: "valid event",
            event: &models.Event{
                Title:       "Test Event",
                StartTime:   time.Now().Add(24 * time.Hour),
                Timezone:    "America/Los_Angeles",
                Status:      models.EventStatusDraft,
                CreatedBy:   1,
                Version:     1,
                MaxPlusOnes: 0,
            },
            wantErr: false,
        },
        {
            name: "event with all fields",
            event: &models.Event{
                Title:        "Complete Event",
                Description:  stringPtr("Full description"),
                StartTime:    time.Now().Add(24 * time.Hour),
                EndTime:      timePtr(time.Now().Add(26 * time.Hour)),
                Timezone:     "America/Los_Angeles",
                Location:     stringPtr("123 Main St"),
                Status:       models.EventStatusDraft,
                CreatedBy:    1,
                Version:      1,
                MaxPlusOnes:  2,
                RSVPDeadline: timePtr(time.Now().Add(12 * time.Hour)),
            },
            wantErr: false,
        },
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(context.Background(), tt.event)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && tt.event.ID == 0 {
                t.Error("Expected event ID to be set after creation")
            }
        })
    }
}

func TestEventRepository_UpdateWithVersion(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    
    event := &models.Event{
        Title:       "Original Title",
        StartTime:   time.Now().Add(24 * time.Hour),
        Timezone:    "America/Los_Angeles",
        Status:      models.EventStatusDraft,
        CreatedBy:   1,
        Version:     1,
        MaxPlusOnes: 0,
    }
    
    if err := repo.Create(context.Background(), event); err != nil {
        t.Fatalf("Failed to create event: %v", err)
    }
    
    tests := []struct {
        name            string
        updateTitle     string
        expectedVersion int
        wantErr         bool
        errType         string
    }{
        {
            name:            "successful update",
            updateTitle:     "Updated Title",
            expectedVersion: 1,
            wantErr:         false,
        },
        {
            name:            "version conflict",
            updateTitle:     "Another Update",
            expectedVersion: 1,
            wantErr:         true,
            errType:         "VersionConflictError",
        },
        {
            name:            "successful second update",
            updateTitle:     "Third Title",
            expectedVersion: 2,
            wantErr:         false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event.Title = tt.updateTitle
            err := repo.UpdateWithVersion(context.Background(), event, tt.expectedVersion)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("UpdateWithVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && tt.errType != "" {
                if _, ok := err.(*models.VersionConflictError); !ok {
                    t.Errorf("Expected error type %s, got %T", tt.errType, err)
                }
            }
            
            if !tt.wantErr {
                retrieved, err := repo.GetByID(context.Background(), event.ID)
                if err != nil {
                    t.Fatalf("Failed to retrieve updated event: %v", err)
                }
                
                if retrieved.Title != tt.updateTitle {
                    t.Errorf("Title = %q, want %q", retrieved.Title, tt.updateTitle)
                }
                
                if retrieved.Version != tt.expectedVersion+1 {
                    t.Errorf("Version = %d, want %d", retrieved.Version, tt.expectedVersion+1)
                }
            }
        })
    }
}

func TestEventRepository_List(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    
    events := []*models.Event{
        {
            Title:     "Event 1",
            StartTime: time.Now().Add(24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusDraft,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Event 2",
            StartTime: time.Now().Add(48 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusPublished,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Event 3",
            StartTime: time.Now().Add(72 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusDraft,
            CreatedBy: 2,
            Version:   1,
        },
    }
    
    for _, e := range events {
        if err := repo.Create(context.Background(), e); err != nil {
            t.Fatalf("Failed to create event: %v", err)
        }
    }
    
    tests := []struct {
        name      string
        filters   repositories.ListFilters
        wantCount int
    }{
        {
            name:      "no filters",
            filters:   repositories.ListFilters{},
            wantCount: 3,
        },
        {
            name: "filter by creator",
            filters: repositories.ListFilters{
                CreatorID: int64Ptr(1),
            },
            wantCount: 2,
        },
        {
            name: "filter by status",
            filters: repositories.ListFilters{
                Status: statusPtr(models.EventStatusDraft),
            },
            wantCount: 2,
        },
        {
            name: "filter by creator and status",
            filters: repositories.ListFilters{
                CreatorID: int64Ptr(1),
                Status:    statusPtr(models.EventStatusDraft),
            },
            wantCount: 1,
        },
        {
            name: "with pagination",
            filters: repositories.ListFilters{
                Limit:  2,
                Offset: 0,
            },
            wantCount: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            results, err := repo.List(context.Background(), tt.filters)
            if err != nil {
                t.Errorf("List() error = %v", err)
                return
            }
            
            if len(results) != tt.wantCount {
                t.Errorf("List() returned %d events, want %d", len(results), tt.wantCount)
            }
        })
    }
}

func TestEventRepository_GetEventsToArchive(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    
    now := time.Now()
    
    events := []*models.Event{
        {
            Title:     "Old Completed Event",
            StartTime: now.Add(-40 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusPublished,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Recent Event",
            StartTime: now.Add(-10 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusPublished,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Old Cancelled Event",
            StartTime: now.Add(-35 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusCancelled,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Already Archived",
            StartTime: now.Add(-50 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusArchived,
            CreatedBy: 1,
            Version:   1,
        },
    }
    
    for _, e := range events {
        if err := repo.Create(context.Background(), e); err != nil {
            t.Fatalf("Failed to create event: %v", err)
        }
    }
    
    results, err := repo.GetEventsToArchive(context.Background(), 30)
    if err != nil {
        t.Fatalf("GetEventsToArchive() error = %v", err)
    }
    
    if len(results) != 2 {
        t.Errorf("GetEventsToArchive() returned %d events, want 2", len(results))
    }
    
    for _, event := range results {
        if event.Status == models.EventStatusArchived {
            t.Error("GetEventsToArchive() returned already archived event")
        }
        
        daysSinceEvent := int(now.Sub(event.StartTime).Hours() / 24)
        if daysSinceEvent < 30 {
            t.Errorf("GetEventsToArchive() returned event only %d days old", daysSinceEvent)
        }
    }
}
```

### Integration Tests

```go
func TestEventRepository_Integration_ConcurrentUpdates(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    
    event := &models.Event{
        Title:       "Concurrent Test",
        StartTime:   time.Now().Add(24 * time.Hour),
        Timezone:    "America/Los_Angeles",
        Status:      models.EventStatusDraft,
        CreatedBy:   1,
        Version:     1,
        MaxPlusOnes: 0,
    }
    
    if err := repo.Create(context.Background(), event); err != nil {
        t.Fatalf("Failed to create event: %v", err)
    }
    
    var wg sync.WaitGroup
    errors := make(chan error, 2)
    
    for i := 0; i < 2; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            
            e := *event
            e.Title = fmt.Sprintf("Update %d", n)
            err := repo.UpdateWithVersion(context.Background(), &e, 1)
            errors <- err
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    successCount := 0
    conflictCount := 0
    
    for err := range errors {
        if err == nil {
            successCount++
        } else if _, ok := err.(*models.VersionConflictError); ok {
            conflictCount++
        } else {
            t.Errorf("Unexpected error: %v", err)
        }
    }
    
    if successCount != 1 {
        t.Errorf("Expected 1 successful update, got %d", successCount)
    }
    
    if conflictCount != 1 {
        t.Errorf("Expected 1 version conflict, got %d", conflictCount)
    }
}
```

---

## Dependencies

**Depends on:** 
- 02_STORY_01_event_model.md - Event model and validation

**Blocks:** 
- All event service and handler stories

**External Dependencies:**
- Database driver (sqlite3 or postgres)
- Database migrations

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/db/repositories/...`)
- [x] Integration tests pass with real database
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Optimistic locking verified with concurrent tests
- [x] Documentation complete
- [x] Changes committed to git

---

## Implementation Notes

### Optimistic Locking

The version field is incremented atomically in the UPDATE statement. If the WHERE clause matches zero rows, it means either:
1. The event doesn't exist, or
2. The version has changed (conflict)

The repository distinguishes between these cases by first checking if the event exists.

### Soft Delete

The Delete operation sets the status to ARCHIVED rather than removing the row. This preserves history and maintains referential integrity with related tables (invites, RSVPs, etc.).

### Transaction Handling

All write operations should be wrapped in transactions at the service layer. The repository provides the atomic operations, but transaction management is handled by the caller.

### Error Types

Create custom error types in [`internal/models/errors.go`](../../internal/models/errors.go):

```go
type VersionConflictError struct {
    ResourceType string
    ResourceID   int64
    Expected     int
}

func (e *VersionConflictError) Error() string {
    return fmt.Sprintf("version conflict for %s %d: expected version %d", 
        e.ResourceType, e.ResourceID, e.Expected)
}
```

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **HLD:** Section 5 (Event Model), Section 9 (Optimistic Locking)
- **LLD:** [lld/02_EVENT_LLD.md](../lld/02_EVENT_LLD.md) - Section 4.1
- **LLD:** [lld/07_DATABASE_LLD.md](../lld/07_DATABASE_LLD.md) - Repository Pattern
- **Epic:** [02_EPIC_events.md](02_EPIC_events.md)
