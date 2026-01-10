package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
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
	GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error)
	CountEvents(ctx context.Context) (int, error)
}

type ListFilters struct {
	CreatorID *int64
	Status    *models.EventStatus
	Limit     int
	Offset    int
}

type eventRepository struct {
	db db.Database
}

func NewEventRepository(database db.Database) EventRepository {
	return &eventRepository{db: database}
}

func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	if event.Title == "" {
		return &models.ValidationError{
			Field:   "title",
			Message: "title is required",
		}
	}

	if event.Timezone == "" {
		return &models.ValidationError{
			Field:   "timezone",
			Message: "timezone is required",
		}
	}

	query := `
		INSERT INTO events (
			title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	version := 1
	icsSequence := 0

	result, err := r.db.Exec(ctx, query,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.Timezone,
		event.Location,
		event.Status,
		event.CreatedBy,
		version,
		icsSequence,
		event.MaxPlusOnes,
		event.RSVPDeadline,
		now,
		now,
	)

	if err != nil {
		if isForeignKeyConstraintError(err) {
			return fmt.Errorf("invalid creator: user %d does not exist", event.CreatedBy)
		}
		return fmt.Errorf("failed to create event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	event.ID = id
	event.Version = version
	event.ICSSequence = icsSequence
	event.CreatedAt = now
	event.UpdatedAt = now

	return nil
}

func (r *eventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			created_at, updated_at
		FROM events
		WHERE id = ?
	`

	event := &models.Event{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&event.EndTime,
		&event.Timezone,
		&event.Location,
		&event.Status,
		&event.CreatedBy,
		&event.Version,
		&event.ICSSequence,
		&event.MaxPlusOnes,
		&event.RSVPDeadline,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Event",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get event by id: %w", err)
	}

	return event, nil
}

func (r *eventRepository) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events
		SET title = ?, description = ?, start_time = ?, end_time = ?,
			timezone = ?, location = ?, max_plus_ones = ?, rsvp_deadline = ?,
			updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.Timezone,
		event.Location,
		event.MaxPlusOnes,
		event.RSVPDeadline,
		now,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Event",
			ID:       event.ID,
		}
	}

	event.UpdatedAt = now

	return nil
}

func (r *eventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	query := `
		UPDATE events
		SET title = ?, description = ?, start_time = ?, end_time = ?,
			timezone = ?, location = ?, max_plus_ones = ?, rsvp_deadline = ?,
			version = version + 1, updated_at = ?
		WHERE id = ? AND version = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.Timezone,
		event.Location,
		event.MaxPlusOnes,
		event.RSVPDeadline,
		now,
		event.ID,
		expectedVersion,
	)

	if err != nil {
		return fmt.Errorf("failed to update event with version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		existing, err := r.GetByID(ctx, event.ID)
		if err != nil {
			if _, ok := err.(*models.NotFoundError); ok {
				return err
			}
			return fmt.Errorf("failed to check if event exists: %w", err)
		}

		return &models.OptimisticLockError{
			Resource:        "Event",
			ID:              event.ID,
			ExpectedVersion: expectedVersion,
			ActualVersion:   existing.Version,
		}
	}

	event.Version = expectedVersion + 1
	event.UpdatedAt = now

	return nil
}

func (r *eventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	query := `
		UPDATE events
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Event",
			ID:       id,
		}
	}

	return nil
}

func (r *eventRepository) Delete(ctx context.Context, id int64) error {
	return r.UpdateStatus(ctx, id, models.EventStatusArchived)
}

func (r *eventRepository) List(ctx context.Context, filters ListFilters) ([]*models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			created_at, updated_at
		FROM events
		WHERE 1=1
	`

	args := []interface{}{}

	if filters.CreatorID != nil {
		query += " AND created_by = ?"
		args = append(args, *filters.CreatorID)
	}

	if filters.Status != nil {
		query += " AND status = ?"
		args = append(args, *filters.Status)
	}

	query += " ORDER BY start_time DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.Timezone,
			&event.Location,
			&event.Status,
			&event.CreatedBy,
			&event.Version,
			&event.ICSSequence,
			&event.MaxPlusOnes,
			&event.RSVPDeadline,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

func (r *eventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			created_at, updated_at
		FROM events
		WHERE status = ?
		ORDER BY start_time DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by status: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.Timezone,
			&event.Location,
			&event.Status,
			&event.CreatedBy,
			&event.Version,
			&event.ICSSequence,
			&event.MaxPlusOnes,
			&event.RSVPDeadline,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

func (r *eventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			created_at, updated_at
		FROM events
		WHERE status IN (?, ?)
			AND start_time < datetime('now', '-' || ? || ' days')
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, models.EventStatusPublished, models.EventStatusCancelled, daysAfterEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get events to archive: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.Timezone,
			&event.Location,
			&event.Status,
			&event.CreatedBy,
			&event.Version,
			&event.ICSSequence,
			&event.MaxPlusOnes,
			&event.RSVPDeadline,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

func (r *eventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			created_at, updated_at
		FROM events
		WHERE created_by = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by creator: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.Timezone,
			&event.Location,
			&event.Status,
			&event.CreatedBy,
			&event.Version,
			&event.ICSSequence,
			&event.MaxPlusOnes,
			&event.RSVPDeadline,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

func isForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "FOREIGN KEY constraint failed") ||
		strings.Contains(errMsg, "foreign key constraint")
}

func (r *eventRepository) CountEvents(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM events WHERE status != 'archived'`
	
	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}
	
	return count, nil
}
