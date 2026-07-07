package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
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
	GetByPublicID(ctx context.Context, publicID string) (*models.Event, error)
	GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error
	UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filters ListFilters) ([]*models.Event, error)
	ListWithStats(ctx context.Context, filters ListFilters) ([]*models.EventWithStats, error)
	GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error)
	GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error)
	GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error)
	CountEvents(ctx context.Context) (int, error)
	GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error)
	UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error
	DeleteComponentOverrides(ctx context.Context, eventID int64) error
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
			public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id, custom_theme_image_url, custom_theme_color,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	version := 1
	icsSequence := 0

	result, err := r.db.Exec(ctx, query,
		event.PublicID,
		event.FriendlyName,
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
		event.AllowRSVPAfterDeadline,
		event.AllowMaybeRSVP,
		event.PrivateGuestList,
		event.FamilyHeadcount,
		event.EventCapacity,
		event.TemplateID,
		event.CustomThemeImageURL,
		event.CustomThemeColor,
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
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id, custom_theme_image_url, custom_theme_color, component_overrides,
			created_at, updated_at
		FROM events
		WHERE id = ?
	`

	event := &models.Event{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.PublicID,
		&event.FriendlyName,
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
		&event.AllowRSVPAfterDeadline,
		&event.AllowMaybeRSVP,
		&event.PrivateGuestList,
		&event.FamilyHeadcount,
		&event.EventCapacity,
		&event.TemplateID,
		&event.CustomThemeImageURL,
		&event.CustomThemeColor,
		&event.ComponentOverrides,
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

func (r *eventRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	query := `
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id,
			custom_theme_image_url, custom_theme_color, component_overrides,
			created_at, updated_at
		FROM events
		WHERE public_id = ?
	`

	event := &models.Event{}
	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&event.ID,
		&event.PublicID,
		&event.FriendlyName,
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
		&event.AllowRSVPAfterDeadline,
		&event.AllowMaybeRSVP,
		&event.PrivateGuestList,
		&event.FamilyHeadcount,
		&event.EventCapacity,
		&event.TemplateID,
		&event.CustomThemeImageURL,
		&event.CustomThemeColor,
		&event.ComponentOverrides,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Event",
				ID:       publicID,
			}
		}
		return nil, fmt.Errorf("failed to get event by public_id: %w", err)
	}

	return event, nil
}

func (r *eventRepository) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	query := `
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id,
			custom_theme_image_url, custom_theme_color, component_overrides,
			created_at, updated_at
		FROM events
		WHERE friendly_name = ?
	`

	event := &models.Event{}
	err := r.db.QueryRow(ctx, query, friendlyName).Scan(
		&event.ID,
		&event.PublicID,
		&event.FriendlyName,
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
		&event.AllowRSVPAfterDeadline,
		&event.AllowMaybeRSVP,
		&event.PrivateGuestList,
		&event.FamilyHeadcount,
		&event.EventCapacity,
		&event.TemplateID,
		&event.CustomThemeImageURL,
		&event.CustomThemeColor,
		&event.ComponentOverrides,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Event",
				ID:       friendlyName,
			}
		}
		return nil, fmt.Errorf("failed to get event by friendly_name: %w", err)
	}

	return event, nil
}

func (r *eventRepository) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events
		SET title = ?, description = ?, start_time = ?, end_time = ?,
			timezone = ?, location = ?, max_plus_ones = ?, rsvp_deadline = ?,
			allow_rsvp_after_deadline = ?, allow_maybe_rsvp = ?, private_guest_list = ?,
			family_headcount = ?, event_capacity = ?,
			template_id = ?, custom_theme_image_url = ?, custom_theme_color = ?,
			component_overrides = ?, updated_at = ?
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
		event.AllowRSVPAfterDeadline,
		event.AllowMaybeRSVP,
		event.PrivateGuestList,
		event.FamilyHeadcount,
		event.EventCapacity,
		event.TemplateID,
		event.CustomThemeImageURL,
		event.CustomThemeColor,
		event.ComponentOverrides,
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
			allow_rsvp_after_deadline = ?, allow_maybe_rsvp = ?, private_guest_list = ?,
			family_headcount = ?, event_capacity = ?,
			template_id = ?, custom_theme_image_url = ?, custom_theme_color = ?,
			component_overrides = ?, version = version + 1, updated_at = ?
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
		event.AllowRSVPAfterDeadline,
		event.AllowMaybeRSVP,
		event.PrivateGuestList,
		event.FamilyHeadcount,
		event.EventCapacity,
		event.TemplateID,
		event.CustomThemeImageURL,
		event.CustomThemeColor,
		event.ComponentOverrides,
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
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id,
			custom_theme_image_url, custom_theme_color, component_overrides,
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
			&event.PublicID,
			&event.FriendlyName,
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
			&event.AllowRSVPAfterDeadline,
			&event.AllowMaybeRSVP,
			&event.PrivateGuestList,
			&event.FamilyHeadcount,
			&event.EventCapacity,
			&event.TemplateID,
			&event.CustomThemeImageURL,
			&event.CustomThemeColor,
			&event.ComponentOverrides,
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

func (r *eventRepository) ListWithStats(ctx context.Context, filters ListFilters) ([]*models.EventWithStats, error) {
	baseSelect := `
		e.id, e.public_id, e.friendly_name, e.title, e.description, e.start_time, e.end_time, e.timezone, e.location,
		e.status, e.created_by, e.version, e.ics_sequence, e.max_plus_ones, e.rsvp_deadline,
		e.allow_rsvp_after_deadline, e.allow_maybe_rsvp, e.private_guest_list, e.family_headcount, e.event_capacity,
		e.template_id, e.custom_theme_image_url, e.custom_theme_color, e.component_overrides,
		e.created_at, e.updated_at`

	query := fmt.Sprintf(`
		SELECT %s,
			COUNT(DISTINCT i.id) AS invite_count,
			COUNT(DISTINCT r.id) AS rsvp_count,
			COUNT(DISTINCT CASE WHEN r.response = 'yes' THEN r.id END) AS accept_count
		FROM events e
		LEFT JOIN invites i ON e.id = i.event_id AND i.status != 'revoked'
		LEFT JOIN rsvps r ON i.id = r.invite_id
		WHERE 1=1
	`, baseSelect)

	args := []interface{}{}

	if filters.CreatorID != nil {
		query += " AND e.created_by = ?"
		args = append(args, *filters.CreatorID)
	}

	if filters.Status != nil {
		query += " AND e.status = ?"
		args = append(args, *filters.Status)
	}

	query += " GROUP BY e.id"
	query += " ORDER BY e.start_time DESC"

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
		return nil, fmt.Errorf("failed to list events with stats: %w", err)
	}
	defer rows.Close()

	var results []*models.EventWithStats
	for rows.Next() {
		ews := &models.EventWithStats{}
		err := rows.Scan(
			&ews.ID,
			&ews.PublicID,
			&ews.FriendlyName,
			&ews.Title,
			&ews.Description,
			&ews.StartTime,
			&ews.EndTime,
			&ews.Timezone,
			&ews.Location,
			&ews.Status,
			&ews.CreatedBy,
			&ews.Version,
			&ews.ICSSequence,
			&ews.MaxPlusOnes,
			&ews.RSVPDeadline,
			&ews.AllowRSVPAfterDeadline,
			&ews.AllowMaybeRSVP,
			&ews.PrivateGuestList,
			&ews.FamilyHeadcount,
			&ews.EventCapacity,
			&ews.TemplateID,
			&ews.CustomThemeImageURL,
			&ews.CustomThemeColor,
			&ews.ComponentOverrides,
			&ews.CreatedAt,
			&ews.UpdatedAt,
			&ews.InviteCount,
			&ews.RSVPCount,
			&ews.AcceptCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event with stats: %w", err)
		}
		results = append(results, ews)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events with stats: %w", err)
	}

	return results, nil
}

func (r *eventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id,
			custom_theme_image_url, custom_theme_color, component_overrides,
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
			&event.PublicID,
			&event.FriendlyName,
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
			&event.AllowRSVPAfterDeadline,
			&event.AllowMaybeRSVP,
			&event.PrivateGuestList,
			&event.FamilyHeadcount,
			&event.EventCapacity,
			&event.TemplateID,
			&event.CustomThemeImageURL,
			&event.CustomThemeColor,
			&event.ComponentOverrides,
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
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id,
			custom_theme_image_url, custom_theme_color, component_overrides,
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
			&event.PublicID,
			&event.FriendlyName,
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
			&event.AllowRSVPAfterDeadline,
			&event.AllowMaybeRSVP,
			&event.PrivateGuestList,
			&event.FamilyHeadcount,
			&event.EventCapacity,
			&event.TemplateID,
			&event.CustomThemeImageURL,
			&event.CustomThemeColor,
			&event.ComponentOverrides,
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
		SELECT id, public_id, friendly_name, title, description, start_time, end_time, timezone, location,
			status, created_by, version, ics_sequence, max_plus_ones, rsvp_deadline,
			allow_rsvp_after_deadline, allow_maybe_rsvp, private_guest_list, family_headcount, event_capacity,
			template_id,
			custom_theme_image_url, custom_theme_color, component_overrides,
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
			&event.PublicID,
			&event.FriendlyName,
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
			&event.AllowRSVPAfterDeadline,
			&event.AllowMaybeRSVP,
			&event.PrivateGuestList,
			&event.FamilyHeadcount,
			&event.EventCapacity,
			&event.TemplateID,
			&event.CustomThemeImageURL,
			&event.CustomThemeColor,
			&event.ComponentOverrides,
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

func (r *eventRepository) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {
	query := `
		SELECT component_overrides
		FROM events
		WHERE id = ?
	`

	var overridesJSON *string
	err := r.db.QueryRow(ctx, query, eventID).Scan(&overridesJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Event",
				ID:       eventID,
			}
		}
		return nil, fmt.Errorf("failed to get component overrides: %w", err)
	}

	if overridesJSON == nil || *overridesJSON == "" {
		return nil, nil
	}

	var overrides models.ComponentOverrides
	if err := json.Unmarshal([]byte(*overridesJSON), &overrides); err != nil {
		return nil, fmt.Errorf("failed to unmarshal component overrides: %w", err)
	}

	return &overrides, nil
}

func (r *eventRepository) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	if overrides == nil {
		return fmt.Errorf("overrides cannot be nil")
	}

	overridesJSON, err := json.Marshal(overrides)
	if err != nil {
		return fmt.Errorf("failed to marshal component overrides: %w", err)
	}

	overridesStr := string(overridesJSON)

	query := `
		UPDATE events
		SET component_overrides = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, overridesStr, now, eventID)
	if err != nil {
		return fmt.Errorf("failed to update component overrides: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Event",
			ID:       eventID,
		}
	}

	return nil
}

func (r *eventRepository) DeleteComponentOverrides(ctx context.Context, eventID int64) error {
	query := `
		UPDATE events
		SET component_overrides = NULL, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete component overrides: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Event",
			ID:       eventID,
		}
	}

	return nil
}
