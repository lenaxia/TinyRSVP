package jobs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockEventService struct {
	getEventsToArchiveFunc func(ctx context.Context) ([]*models.Event, error)
	archiveEventFunc       func(ctx context.Context, id int64) error
}

func (m *mockEventService) GetEventsToArchive(ctx context.Context) ([]*models.Event, error) {
	if m.getEventsToArchiveFunc != nil {
		return m.getEventsToArchiveFunc(ctx)
	}
	return nil, nil
}

func (m *mockEventService) ArchiveEvent(ctx context.Context, id int64) error {
	if m.archiveEventFunc != nil {
		return m.archiveEventFunc(ctx, id)
	}
	return nil
}

func TestNewEventArchiver(t *testing.T) {
	mockService := &mockEventService{}
	archiver := NewEventArchiver(mockService, 30)

	if archiver == nil {
		t.Fatal("NewEventArchiver returned nil")
	}

	if archiver.daysAfterEvent != 30 {
		t.Errorf("daysAfterEvent = %d, want 30", archiver.daysAfterEvent)
	}
}

func TestEventArchiver_Run_NoEvents(t *testing.T) {
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			return []*models.Event{}, nil
		},
	}

	archiver := NewEventArchiver(mockService, 30)
	err := archiver.Run(context.Background())

	if err != nil {
		t.Errorf("Run() error = %v, want nil", err)
	}
}

func TestEventArchiver_Run_SingleEvent(t *testing.T) {
	archivedCount := 0
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			return []*models.Event{
				{
					ID:        1,
					Title:     "Old Event",
					StartTime: time.Now().Add(-40 * 24 * time.Hour),
					Status:    models.EventStatusPublished,
				},
			}, nil
		},
		archiveEventFunc: func(ctx context.Context, id int64) error {
			if id != 1 {
				t.Errorf("ArchiveEvent called with id = %d, want 1", id)
			}
			archivedCount++
			return nil
		},
	}

	archiver := NewEventArchiver(mockService, 30)
	err := archiver.Run(context.Background())

	if err != nil {
		t.Errorf("Run() error = %v, want nil", err)
	}

	if archivedCount != 1 {
		t.Errorf("Archived %d events, want 1", archivedCount)
	}
}

func TestEventArchiver_Run_MultipleEvents(t *testing.T) {
	archivedIDs := []int64{}
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			return []*models.Event{
				{
					ID:        1,
					Title:     "Old Event 1",
					StartTime: time.Now().Add(-40 * 24 * time.Hour),
					Status:    models.EventStatusPublished,
				},
				{
					ID:        2,
					Title:     "Old Event 2",
					StartTime: time.Now().Add(-35 * 24 * time.Hour),
					Status:    models.EventStatusCancelled,
				},
				{
					ID:        3,
					Title:     "Old Event 3",
					StartTime: time.Now().Add(-50 * 24 * time.Hour),
					Status:    models.EventStatusPublished,
				},
			}, nil
		},
		archiveEventFunc: func(ctx context.Context, id int64) error {
			archivedIDs = append(archivedIDs, id)
			return nil
		},
	}

	archiver := NewEventArchiver(mockService, 30)
	err := archiver.Run(context.Background())

	if err != nil {
		t.Errorf("Run() error = %v, want nil", err)
	}

	if len(archivedIDs) != 3 {
		t.Errorf("Archived %d events, want 3", len(archivedIDs))
	}

	expectedIDs := []int64{1, 2, 3}
	for i, id := range expectedIDs {
		if archivedIDs[i] != id {
			t.Errorf("archivedIDs[%d] = %d, want %d", i, archivedIDs[i], id)
		}
	}
}

func TestEventArchiver_Run_PartialFailure(t *testing.T) {
	archivedIDs := []int64{}
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			return []*models.Event{
				{
					ID:        1,
					Title:     "Event 1",
					StartTime: time.Now().Add(-40 * 24 * time.Hour),
					Status:    models.EventStatusPublished,
				},
				{
					ID:        2,
					Title:     "Event 2",
					StartTime: time.Now().Add(-35 * 24 * time.Hour),
					Status:    models.EventStatusPublished,
				},
				{
					ID:        3,
					Title:     "Event 3",
					StartTime: time.Now().Add(-45 * 24 * time.Hour),
					Status:    models.EventStatusPublished,
				},
			}, nil
		},
		archiveEventFunc: func(ctx context.Context, id int64) error {
			if id == 2 {
				return fmt.Errorf("database error")
			}
			archivedIDs = append(archivedIDs, id)
			return nil
		},
	}

	archiver := NewEventArchiver(mockService, 30)
	err := archiver.Run(context.Background())

	if err == nil {
		t.Error("Run() error = nil, want error")
	}

	if len(archivedIDs) != 2 {
		t.Errorf("Archived %d events, want 2", len(archivedIDs))
	}

	if archivedIDs[0] != 1 || archivedIDs[1] != 3 {
		t.Errorf("archivedIDs = %v, want [1, 3]", archivedIDs)
	}
}

func TestEventArchiver_Run_GetEventsError(t *testing.T) {
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			return nil, fmt.Errorf("database connection error")
		},
	}

	archiver := NewEventArchiver(mockService, 30)
	err := archiver.Run(context.Background())

	if err == nil {
		t.Error("Run() error = nil, want error")
	}
}

func TestEventArchiver_Run_Idempotency(t *testing.T) {
	callCount := 0
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			callCount++
			if callCount == 1 {
				return []*models.Event{
					{
						ID:        1,
						Title:     "Event",
						StartTime: time.Now().Add(-40 * 24 * time.Hour),
						Status:    models.EventStatusPublished,
					},
				}, nil
			}
			return []*models.Event{}, nil
		},
		archiveEventFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	archiver := NewEventArchiver(mockService, 30)

	if err := archiver.Run(context.Background()); err != nil {
		t.Errorf("First run failed: %v", err)
	}

	if err := archiver.Run(context.Background()); err != nil {
		t.Errorf("Second run failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("GetEventsToArchive called %d times, want 2", callCount)
	}
}

func TestEventArchiver_Run_ContextCancellation(t *testing.T) {
	archivedCount := 0
	mockService := &mockEventService{
		getEventsToArchiveFunc: func(ctx context.Context) ([]*models.Event, error) {
			return []*models.Event{
				{ID: 1, Title: "Event 1", StartTime: time.Now().Add(-40 * 24 * time.Hour)},
				{ID: 2, Title: "Event 2", StartTime: time.Now().Add(-35 * 24 * time.Hour)},
				{ID: 3, Title: "Event 3", StartTime: time.Now().Add(-45 * 24 * time.Hour)},
			}, nil
		},
		archiveEventFunc: func(ctx context.Context, id int64) error {
			archivedCount++
			return nil
		},
	}

	archiver := NewEventArchiver(mockService, 30)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := archiver.Run(ctx)
	if err == nil {
		t.Error("Run() error = nil, want context cancellation error")
	}

	if archivedCount > 0 {
		t.Errorf("Archived %d events after cancellation, want 0", archivedCount)
	}
}
