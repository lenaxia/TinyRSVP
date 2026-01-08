package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type RSVPRepository interface {
	Create(ctx context.Context, rsvp *models.RSVP) error
	GetByID(ctx context.Context, id int64) (*models.RSVP, error)
	GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error)
	GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error)
	Update(ctx context.Context, rsvp *models.RSVP) error
	GetStats(ctx context.Context, eventID int64) (*RSVPStats, error)
}

type RSVPStats struct {
	TotalInvites int `json:"total_invites"`
	YesCount     int `json:"yes_count"`
	NoCount      int `json:"no_count"`
	MaybeCount   int `json:"maybe_count"`
	NoResponse   int `json:"no_response"`
	TotalGuests  int `json:"total_guests"`
}

type rsvpRepository struct {
	db db.Database
}

func NewRSVPRepository(database db.Database) RSVPRepository {
	return &rsvpRepository{db: database}
}

func (r *rsvpRepository) Create(ctx context.Context, rsvp *models.RSVP) error {
	if err := rsvp.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO rsvps (invite_id, response, plus_ones, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(ctx, query, rsvp.InviteID, rsvp.Response, rsvp.PlusOnes)
	if err != nil {
		return fmt.Errorf("failed to create RSVP: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	rsvp.ID = id

	created, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	*rsvp = *created
	return nil
}

func (r *rsvpRepository) GetByID(ctx context.Context, id int64) (*models.RSVP, error) {
	query := `
		SELECT id, invite_id, response, plus_ones, created_at, updated_at
		FROM rsvps
		WHERE id = ?
	`

	var rsvp models.RSVP
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rsvp.ID,
		&rsvp.InviteID,
		&rsvp.Response,
		&rsvp.PlusOnes,
		&rsvp.CreatedAt,
		&rsvp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &rsvp, nil
}

func (r *rsvpRepository) GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error) {
	query := `
		SELECT id, invite_id, response, plus_ones, created_at, updated_at
		FROM rsvps
		WHERE invite_id = ?
	`

	var rsvp models.RSVP
	err := r.db.QueryRow(ctx, query, inviteID).Scan(
		&rsvp.ID,
		&rsvp.InviteID,
		&rsvp.Response,
		&rsvp.PlusOnes,
		&rsvp.CreatedAt,
		&rsvp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &rsvp, nil
}

func (r *rsvpRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error) {
	query := `
		SELECT r.id, r.invite_id, r.response, r.plus_ones, r.created_at, r.updated_at
		FROM rsvps r
		INNER JOIN invites i ON r.invite_id = i.id
		WHERE i.event_id = ?
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query RSVPs: %w", err)
	}
	defer rows.Close()

	var rsvps []*models.RSVP
	for rows.Next() {
		var rsvp models.RSVP
		err := rows.Scan(
			&rsvp.ID,
			&rsvp.InviteID,
			&rsvp.Response,
			&rsvp.PlusOnes,
			&rsvp.CreatedAt,
			&rsvp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan RSVP: %w", err)
		}
		rsvps = append(rsvps, &rsvp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating RSVPs: %w", err)
	}

	return rsvps, nil
}

func (r *rsvpRepository) Update(ctx context.Context, rsvp *models.RSVP) error {
	if err := rsvp.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE rsvps
		SET response = ?, plus_ones = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, rsvp.Response, rsvp.PlusOnes, rsvp.ID)
	if err != nil {
		return fmt.Errorf("failed to update RSVP: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	updated, err := r.GetByID(ctx, rsvp.ID)
	if err != nil {
		return err
	}

	*rsvp = *updated
	return nil
}

func (r *rsvpRepository) GetStats(ctx context.Context, eventID int64) (*RSVPStats, error) {
	query := `
		SELECT 
			COUNT(DISTINCT i.id) as total_invites,
			COUNT(CASE WHEN r.response = 'yes' THEN 1 END) as yes_count,
			COUNT(CASE WHEN r.response = 'no' THEN 1 END) as no_count,
			COUNT(CASE WHEN r.response = 'maybe' THEN 1 END) as maybe_count,
			COALESCE(SUM(CASE WHEN r.response = 'yes' THEN 1 + r.plus_ones ELSE 0 END), 0) as total_guests
		FROM invites i
		LEFT JOIN rsvps r ON i.id = r.invite_id
		WHERE i.event_id = ?
	`

	var stats RSVPStats
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&stats.TotalInvites,
		&stats.YesCount,
		&stats.NoCount,
		&stats.MaybeCount,
		&stats.TotalGuests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get RSVP stats: %w", err)
	}

	stats.NoResponse = stats.TotalInvites - stats.YesCount - stats.NoCount - stats.MaybeCount

	return &stats, nil
}
