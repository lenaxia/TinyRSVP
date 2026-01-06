package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yourusername/tinyrsvp/internal/db"
	"github.com/yourusername/tinyrsvp/internal/models"
)

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id string) (*models.Session, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) (int64, error)
	UpdateLastAccessed(ctx context.Context, id string) error
}

type sessionRepository struct {
	db db.Database
}

func NewSessionRepository(database db.Database) SessionRepository {
	return &sessionRepository{db: database}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at, last_accessed_at, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		now,
		session.ExpiresAt,
		now,
		session.IPAddress,
		session.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	session.CreatedAt = now
	session.LastAccessedAt = now

	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at, last_accessed_at, ip_address, user_agent
		FROM sessions
		WHERE id = ?
	`

	session := &models.Session{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastAccessedAt,
		&session.IPAddress,
		&session.UserAgent,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Session",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get session by id: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at, last_accessed_at, ip_address, user_agent
		FROM sessions
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user id: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.LastAccessedAt,
			&session.IPAddress,
			&session.UserAgent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions
		SET expires_at = ?, last_accessed_at = ?, ip_address = ?, user_agent = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query,
		session.ExpiresAt,
		session.LastAccessedAt,
		session.IPAddress,
		session.UserAgent,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Session",
			ID:       session.ID,
		}
	}

	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Session",
			ID:       id,
		}
	}

	return nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = ?`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user id: %w", err)
	}

	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < ?`

	result, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *sessionRepository) UpdateLastAccessed(ctx context.Context, id string) error {
	query := `
		UPDATE sessions
		SET last_accessed_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last accessed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Session",
			ID:       id,
		}
	}

	return nil
}
