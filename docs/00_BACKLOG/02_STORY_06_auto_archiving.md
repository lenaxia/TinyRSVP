# User Story: Event Auto-Archiving

**Epic:** [02_EPIC_events.md](02_EPIC_events.md)
**Priority:** Low
**Status:** Not Started
**Estimated Effort:** 4 hours

---

## User Story

As a **system administrator**, I want **events to automatically archive 30 days after completion** so that **the system maintains a clean active event list without manual intervention**.

---

## Acceptance Criteria

- [ ] Scheduled job runs daily
- [ ] Job identifies events older than 30 days
- [ ] Job only archives published or cancelled events
- [ ] Job updates event status to archived
- [ ] Job is idempotent (safe to run multiple times)
- [ ] Job logs archiving activity
- [ ] Job handles errors gracefully
- [ ] All tests pass with timeout

---

## Technical Details

### Archiving Job

```go
package jobs

import (
    "context"
    "log"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/events"
)

type EventArchiver struct {
    service       events.Service
    daysAfterEvent int
}

func NewEventArchiver(service events.Service, daysAfterEvent int) *EventArchiver {
    return &EventArchiver{
        service:       service,
        daysAfterEvent: daysAfterEvent,
    }
}

func (a *EventArchiver) Run(ctx context.Context) error {
    log.Printf("Starting event archiving job (threshold: %d days)", a.daysAfterEvent)
    
    eventsToArchive, err := a.service.GetEventsToArchive(ctx)
    if err != nil {
        return fmt.Errorf("failed to get events to archive: %w", err)
    }
    
    if len(eventsToArchive) == 0 {
        log.Println("No events to archive")
        return nil
    }
    
    archived := 0
    failed := 0
    
    for _, event := range eventsToArchive {
        if err := a.service.ArchiveEvent(ctx, event.ID); err != nil {
            log.Printf("Failed to archive event %d: %v", event.ID, err)
            failed++
            continue
        }
        
        log.Printf("Archived event %d: %s", event.ID, event.Title)
        archived++
    }
    
    log.Printf("Event archiving complete: %d archived, %d failed", archived, failed)
    
    if failed > 0 {
        return fmt.Errorf("failed to archive %d events", failed)
    }
    
    return nil
}
```

### Scheduler Integration

```go
package main

import (
    "context"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/jobs"
)

func startScheduledJobs(archiver *jobs.EventArchiver) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    go func() {
        for range ticker.C {
            ctx := context.Background()
            if err := archiver.Run(ctx); err != nil {
                log.Printf("Event archiving job failed: %v", err)
            }
        }
    }()
    
    log.Println("Scheduled jobs started")
}
```

### Configuration

```yaml
jobs:
  event_archiving:
    enabled: true
    schedule: "0 2 * * *"  # 2 AM daily
    days_after_event: 30
```

---

## Tasks

### Phase 1: Job Implementation (TDD)
- [ ] Write test for identifying events to archive
- [ ] Write test for archiving single event
- [ ] Write test for archiving multiple events
- [ ] Write test for handling errors
- [ ] Write test for idempotency
- [ ] Write test for logging
- [ ] Implement EventArchiver
- [ ] Run tests (should pass)

### Phase 2: Service Integration (TDD)
- [ ] Write test for GetEventsToArchive
- [ ] Write test for ArchiveEvent
- [ ] Write test for date threshold calculation
- [ ] Write test for status filtering
- [ ] Implement service methods
- [ ] Run tests (should pass)

### Phase 3: Scheduler Setup (TDD)
- [ ] Write test for scheduler initialization
- [ ] Write test for job execution
- [ ] Write test for error handling
- [ ] Write test for graceful shutdown
- [ ] Implement scheduler
- [ ] Run tests (should pass)

### Phase 4: Integration Tests
- [ ] Write integration test for full archiving cycle
- [ ] Write integration test with real database
- [ ] Write integration test for concurrent execution
- [ ] Run integration tests

---

## Testing Requirements

### Unit Tests

```go
func TestEventArchiver_Run(t *testing.T) {
    tests := []struct {
        name          string
        eventsToFind  []*models.Event
        archiveErrors map[int64]error
        wantErr       bool
        wantArchived  int
    }{
        {
            name: "archive multiple events",
            eventsToFind: []*models.Event{
                {ID: 1, Title: "Old Event 1", StartTime: time.Now().Add(-40 * 24 * time.Hour)},
                {ID: 2, Title: "Old Event 2", StartTime: time.Now().Add(-35 * 24 * time.Hour)},
            },
            archiveErrors: map[int64]error{},
            wantErr:       false,
            wantArchived:  2,
        },
        {
            name:         "no events to archive",
            eventsToFind: []*models.Event{},
            wantErr:      false,
            wantArchived: 0,
        },
        {
            name: "partial failure",
            eventsToFind: []*models.Event{
                {ID: 1, Title: "Event 1", StartTime: time.Now().Add(-40 * 24 * time.Hour)},
                {ID: 2, Title: "Event 2", StartTime: time.Now().Add(-35 * 24 * time.Hour)},
            },
            archiveErrors: map[int64]error{
                2: fmt.Errorf("database error"),
            },
            wantErr:      true,
            wantArchived: 1,
        },
        {
            name: "idempotent - already archived",
            eventsToFind: []*models.Event{
                {ID: 1, Title: "Event", StartTime: time.Now().Add(-40 * 24 * time.Hour), Status: models.EventStatusArchived},
            },
            archiveErrors: map[int64]error{},
            wantErr:       false,
            wantArchived:  0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            archivedCount := 0
            
            mockService := &events.MockService{
                GetEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
                    return tt.eventsToFind, nil
                },
                ArchiveEventFunc: func(ctx context.Context, id int64) error {
                    if err, exists := tt.archiveErrors[id]; exists {
                        return err
                    }
                    archivedCount++
                    return nil
                },
            }
            
            archiver := jobs.NewEventArchiver(mockService, 30)
            
            err := archiver.Run(context.Background())
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if archivedCount != tt.wantArchived {
                t.Errorf("Archived %d events, want %d", archivedCount, tt.wantArchived)
            }
        })
    }
}

func TestEventService_GetEventsToArchive(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    service := events.NewService(repo, validator, authz)
    
    now := time.Now()
    
    testEvents := []*models.Event{
        {
            Title:     "Old Published",
            StartTime: now.Add(-40 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusPublished,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Old Cancelled",
            StartTime: now.Add(-35 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusCancelled,
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
            Title:     "Already Archived",
            StartTime: now.Add(-50 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusArchived,
            CreatedBy: 1,
            Version:   1,
        },
        {
            Title:     "Old Draft",
            StartTime: now.Add(-40 * 24 * time.Hour),
            Timezone:  "America/Los_Angeles",
            Status:    models.EventStatusDraft,
            CreatedBy: 1,
            Version:   1,
        },
    }
    
    for _, e := range testEvents {
        if err := repo.Create(context.Background(), e); err != nil {
            t.Fatalf("Failed to create test event: %v", err)
        }
    }
    
    results, err := service.GetEventsToArchive(context.Background())
    if err != nil {
        t.Fatalf("GetEventsToArchive() error = %v", err)
    }
    
    if len(results) != 2 {
        t.Errorf("GetEventsToArchive() returned %d events, want 2", len(results))
    }
    
    for _, event := range results {
        if event.Status == models.EventStatusArchived {
            t.Error("Returned already archived event")
        }
        
        if event.Status == models.EventStatusDraft {
            t.Error("Returned draft event")
        }
        
        daysSince := int(now.Sub(event.StartTime).Hours() / 24)
        if daysSince < 30 {
            t.Errorf("Returned event only %d days old", daysSince)
        }
    }
}

func TestEventArchiver_Idempotency(t *testing.T) {
    mockService := &events.MockService{
        GetEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
            return []*models.Event{
                {ID: 1, Title: "Event", StartTime: time.Now().Add(-40 * 24 * time.Hour)},
            }, nil
        },
        ArchiveEventFunc: func(ctx context.Context, id int64) error {
            return nil
        },
    }
    
    archiver := jobs.NewEventArchiver(mockService, 30)
    
    if err := archiver.Run(context.Background()); err != nil {
        t.Errorf("First run failed: %v", err)
    }
    
    mockService.GetEventsToArchiveFunc = func(ctx context.Context) ([]*models.Event, error) {
        return []*models.Event{}, nil
    }
    
    if err := archiver.Run(context.Background()); err != nil {
        t.Errorf("Second run failed: %v", err)
    }
}
```

### Integration Tests

```go
func TestEventArchiver_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewEventRepository(db)
    service := events.NewService(repo, validator, authz)
    archiver := jobs.NewEventArchiver(service, 30)
    
    now := time.Now()
    
    oldEvent := &models.Event{
        Title:     "Old Event",
        StartTime: now.Add(-40 * 24 * time.Hour),
        Timezone:  "America/Los_Angeles",
        Status:    models.EventStatusPublished,
        CreatedBy: 1,
        Version:   1,
    }
    
    if err := repo.Create(context.Background(), oldEvent); err != nil {
        t.Fatalf("Failed to create event: %v", err)
    }
    
    if err := archiver.Run(context.Background()); err != nil {
        t.Fatalf("Archiver run failed: %v", err)
    }
    
    updated, err := repo.GetByID(context.Background(), oldEvent.ID)
    if err != nil {
        t.Fatalf("Failed to retrieve event: %v", err)
    }
    
    if updated.Status != models.EventStatusArchived {
        t.Errorf("Event status = %v, want %v", updated.Status, models.EventStatusArchived)
    }
}
```

---

## Dependencies

**Depends on:** 
- 02_STORY_03_event_service.md - Event service
- 02_STORY_02_event_repository.md - Event repository

**Blocks:** 
- None (optional feature)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Job runs successfully
- [ ] Logging implemented
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### Scheduling Options

Consider using a cron library for more flexible scheduling:

```go
import "github.com/robfig/cron/v3"

c := cron.New()
c.AddFunc("0 2 * * *", func() {
    archiver.Run(context.Background())
})
c.Start()
```

### Error Handling

The job should:
1. Log all errors
2. Continue processing remaining events on individual failures
3. Return error if any events failed to archive
4. Not fail on empty result set

### Monitoring

Consider adding metrics:
- Number of events archived
- Number of failures
- Job execution time
- Last successful run timestamp

### Graceful Shutdown

Ensure job can be interrupted cleanly:

```go
func (a *EventArchiver) RunWithContext(ctx context.Context) error {
    events, err := a.service.GetEventsToArchive(ctx)
    if err != nil {
        return err
    }
    
    for _, event := range events {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            a.service.ArchiveEvent(ctx, event.ID)
        }
    }
    
    return nil
}
```

---

## References

- **HLD:** Section 5 (Event Model), Section 11 (Background Jobs)
- **LLD:** [lld/02_EVENT_LLD.md](../lld/02_EVENT_LLD.md)
- **Epic:** [02_EPIC_events.md](02_EPIC_events.md)
