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

type InviteRepository interface {
	Create(ctx context.Context, invite *models.Invite) error
	CreateBatch(ctx context.Context, invites []*models.Invite) error
	GetByID(ctx context.Context, id int64) (*models.Invite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error)
	Update(ctx context.Context, invite *models.Invite) error
	Delete(ctx context.Context, id int64) error
	ListByEventID(ctx context.Context, eventID int64, filters InviteFilters) ([]*models.Invite, error)
	CountByEventID(ctx context.Context, eventID int64) (int, error)
	GetStats(ctx context.Context, eventID int64) (*InviteStats, error)
	FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error)
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type InviteFilters struct {
	Status       *models.InviteStatus
	Unsubscribed *bool
	EmailInvalid *bool
	Limit        int
	Offset       int
}

type InviteStats struct {
	Total     int
	Draft     int
	Sent      int
	Viewed    int
	Responded int
	Revoked   int
}

type inviteRepository struct {
	db db.Database
}

func NewInviteRepository(database db.Database) InviteRepository {
	return &inviteRepository{db: database}
}

func (r *inviteRepository) Create(ctx context.Context, invite *models.Invite) error {
	if err := invite.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO invites (
			event_id, name, email, token_hash, max_plus_ones, status,
			sent_at, viewed_at, unsubscribed, email_invalid,
			created_at, updated_at, expires_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	result, err := r.db.Exec(ctx, query,
		invite.EventID,
		invite.Name,
		invite.Email,
		invite.TokenHash,
		invite.MaxPlusOnes,
		invite.Status,
		invite.SentAt,
		invite.ViewedAt,
		invite.Unsubscribed,
		invite.EmailInvalid,
		now,
		now,
		invite.ExpiresAt,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return &models.ConflictError{
				Resource: "Invite",
				Field:    "token_hash",
				Value:    invite.TokenHash,
			}
		}
		if isForeignKeyConstraintError(err) {
			return fmt.Errorf("invalid event: event %d does not exist", invite.EventID)
		}
		return fmt.Errorf("failed to create invite: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	invite.ID = id
	invite.CreatedAt = now
	invite.UpdatedAt = now

	return nil
}

func (r *inviteRepository) CreateBatch(ctx context.Context, invites []*models.Invite) error {
	if len(invites) == 0 {
		return nil
	}

	if len(invites) > 500 {
		return &models.ValidationError{
			Field:   "invites",
			Message: "batch size cannot exceed 500",
		}
	}

	for _, invite := range invites {
		if err := invite.Validate(); err != nil {
			return err
		}
	}

	now := time.Now()

	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		query := `
			INSERT INTO invites (
				event_id, name, email, token_hash, max_plus_ones, status,
				sent_at, viewed_at, unsubscribed, email_invalid,
				created_at, updated_at, expires_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		for _, invite := range invites {
			result, err := tx.ExecContext(ctx, query,
				invite.EventID,
				invite.Name,
				invite.Email,
				invite.TokenHash,
				invite.MaxPlusOnes,
				invite.Status,
				invite.SentAt,
				invite.ViewedAt,
				invite.Unsubscribed,
				invite.EmailInvalid,
				now,
				now,
				invite.ExpiresAt,
			)

			if err != nil {
				if isUniqueConstraintError(err) {
					return &models.ConflictError{
						Resource: "Invite",
						Field:    "token_hash",
						Value:    invite.TokenHash,
					}
				}
				if isForeignKeyConstraintError(err) {
					return fmt.Errorf("invalid event: event %d does not exist", invite.EventID)
				}
				return fmt.Errorf("failed to create invite in batch: %w", err)
			}

			id, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert id: %w", err)
			}

			invite.ID = id
			invite.CreatedAt = now
			invite.UpdatedAt = now
		}

		return nil
	})
}

func (r *inviteRepository) GetByID(ctx context.Context, id int64) (*models.Invite, error) {
	query := `
		SELECT id, event_id, name, email, token_hash, max_plus_ones, status,
			sent_at, viewed_at, unsubscribed, email_invalid,
			created_at, updated_at, expires_at
		FROM invites
		WHERE id = ?
	`

	invite := &models.Invite{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&invite.ID,
		&invite.EventID,
		&invite.Name,
		&invite.Email,
		&invite.TokenHash,
		&invite.MaxPlusOnes,
		&invite.Status,
		&invite.SentAt,
		&invite.ViewedAt,
		&invite.Unsubscribed,
		&invite.EmailInvalid,
		&invite.CreatedAt,
		&invite.UpdatedAt,
		&invite.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Invite",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get invite by id: %w", err)
	}

	return invite, nil
}

func (r *inviteRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	query := `
		SELECT id, event_id, name, email, token_hash, max_plus_ones, status,
			sent_at, viewed_at, unsubscribed, email_invalid,
			created_at, updated_at, expires_at
		FROM invites
		WHERE token_hash = ?
	`

	invite := &models.Invite{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&invite.ID,
		&invite.EventID,
		&invite.Name,
		&invite.Email,
		&invite.TokenHash,
		&invite.MaxPlusOnes,
		&invite.Status,
		&invite.SentAt,
		&invite.ViewedAt,
		&invite.Unsubscribed,
		&invite.EmailInvalid,
		&invite.CreatedAt,
		&invite.UpdatedAt,
		&invite.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Invite",
				ID:       tokenHash,
			}
		}
		return nil, fmt.Errorf("failed to get invite by token hash: %w", err)
	}

	return invite, nil
}

func (r *inviteRepository) Update(ctx context.Context, invite *models.Invite) error {
	if err := invite.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE invites
		SET name = ?, email = ?, token_hash = ?, max_plus_ones = ?, status = ?,
			sent_at = ?, viewed_at = ?, revocation_reason = ?, unsubscribed = ?, email_invalid = ?,
			updated_at = ?, expires_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		invite.Name,
		invite.Email,
		invite.TokenHash,
		invite.MaxPlusOnes,
		invite.Status,
		invite.SentAt,
		invite.ViewedAt,
		invite.RevocationReason,
		invite.Unsubscribed,
		invite.EmailInvalid,
		now,
		invite.ExpiresAt,
		invite.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update invite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Invite",
			ID:       invite.ID,
		}
	}

	invite.UpdatedAt = now

	return nil
}

func (r *inviteRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM invites WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Invite",
			ID:       id,
		}
	}

	return nil
}

func (r *inviteRepository) ListByEventID(ctx context.Context, eventID int64, filters InviteFilters) ([]*models.Invite, error) {
	query := `
		SELECT id, event_id, name, email, token_hash, max_plus_ones, status,
			sent_at, viewed_at, unsubscribed, email_invalid,
			created_at, updated_at, expires_at
		FROM invites
		WHERE event_id = ?
	`

	args := []interface{}{eventID}

	if filters.Status != nil {
		query += " AND status = ?"
		args = append(args, *filters.Status)
	}

	if filters.Unsubscribed != nil {
		query += " AND unsubscribed = ?"
		args = append(args, *filters.Unsubscribed)
	}

	if filters.EmailInvalid != nil {
		query += " AND email_invalid = ?"
		args = append(args, *filters.EmailInvalid)
	}

	query += " ORDER BY created_at DESC"

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
		return nil, fmt.Errorf("failed to list invites: %w", err)
	}
	defer rows.Close()

	var invites []*models.Invite
	for rows.Next() {
		invite := &models.Invite{}
		err := rows.Scan(
			&invite.ID,
			&invite.EventID,
			&invite.Name,
			&invite.Email,
			&invite.TokenHash,
			&invite.MaxPlusOnes,
			&invite.Status,
			&invite.SentAt,
			&invite.ViewedAt,
			&invite.Unsubscribed,
			&invite.EmailInvalid,
			&invite.CreatedAt,
			&invite.UpdatedAt,
			&invite.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invite: %w", err)
		}
		invites = append(invites, invite)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invites: %w", err)
	}

	return invites, nil
}

func (r *inviteRepository) CountByEventID(ctx context.Context, eventID int64) (int, error) {
	query := `SELECT COUNT(*) FROM invites WHERE event_id = ?`

	var count int
	err := r.db.QueryRow(ctx, query, eventID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count invites: %w", err)
	}

	return count, nil
}

func (r *inviteRepository) GetStats(ctx context.Context, eventID int64) (*InviteStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END) as draft,
			SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
			SUM(CASE WHEN status = 'viewed' THEN 1 ELSE 0 END) as viewed,
			SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded,
			SUM(CASE WHEN status = 'revoked' THEN 1 ELSE 0 END) as revoked
		FROM invites
		WHERE event_id = ?
	`

	stats := &InviteStats{}
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&stats.Total,
		&stats.Draft,
		&stats.Sent,
		&stats.Viewed,
		&stats.Responded,
		&stats.Revoked,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get invite stats: %w", err)
	}

	return stats, nil
}

func (r *inviteRepository) FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error) {
	if len(emails) == 0 {
		return []string{}, nil
	}

	placeholders := make([]string, len(emails))
	args := []interface{}{eventID}
	for i, email := range emails {
		placeholders[i] = "?"
		args = append(args, email)
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT email
		FROM invites
		WHERE event_id = ? AND email IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate emails: %w", err)
	}
	defer rows.Close()

	var duplicates []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		duplicates = append(duplicates, email)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating emails: %w", err)
	}

	return duplicates, nil
}

func (r *inviteRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM invites WHERE expires_at < ?`

	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired invites: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
