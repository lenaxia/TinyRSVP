package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventService interface {
	GetEventsToArchive(ctx context.Context) ([]*models.Event, error)
	ArchiveEvent(ctx context.Context, id int64) error
}

type EventArchiver struct {
	service        EventService
	daysAfterEvent int
}

func NewEventArchiver(service EventService, daysAfterEvent int) *EventArchiver {
	return &EventArchiver{
		service:        service,
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
		select {
		case <-ctx.Done():
			log.Printf("Event archiving interrupted: %d archived, %d failed", archived, failed)
			return ctx.Err()
		default:
		}

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
