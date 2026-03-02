package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EmailQueueRepository interface {
	Create(ctx context.Context, email *models.EmailQueue) error
	GetByID(ctx context.Context, id int64) (*models.EmailQueue, error)
	GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error)
	GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error)
	GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error)
	UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error
	IncrementAttempts(ctx context.Context, id int64, errorMsg string) error
	MarkSending(ctx context.Context, id int64) error
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, errorMsg string) error
	MarkCancelled(ctx context.Context, id int64) error
	Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error
	GetStats(ctx context.Context) (*EmailQueueStats, error)
}

type EmailQueueStats struct {
	PendingCount   int
	SendingCount   int
	SentCount      int
	FailedCount    int
	CancelledCount int
	TotalCount     int
}

type emailQueueRepository struct {
	db db.Database
}

func NewEmailQueueRepository(database db.Database) EmailQueueRepository {
	return &emailQueueRepository{db: database}
}

func (r *emailQueueRepository) Create(ctx context.Context, email *models.EmailQueue) error {
	if err := email.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO email_queue (
			to_email, to_name, subject, body_text, body_html, attachments,
			status, attempts, max_attempts, scheduled_for, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		email.ToEmail,
		email.ToName,
		email.Subject,
		email.BodyText,
		email.BodyHTML,
		email.Attachments,
		email.Status,
		email.Attempts,
		email.MaxAttempts,
		email.ScheduledFor,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create email queue entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	email.ID = id
	email.CreatedAt = now

	return nil
}

func (r *emailQueueRepository) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	query := `
		SELECT id, to_email, to_name, subject, body_text, body_html, attachments,
		       status, attempts, max_attempts, last_attempt_at, last_error,
		       scheduled_for, created_at
		FROM email_queue
		WHERE id = ?
	`

	email := &models.EmailQueue{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&email.ID,
		&email.ToEmail,
		&email.ToName,
		&email.Subject,
		&email.BodyText,
		&email.BodyHTML,
		&email.Attachments,
		&email.Status,
		&email.Attempts,
		&email.MaxAttempts,
		&email.LastAttemptAt,
		&email.LastError,
		&email.ScheduledFor,
		&email.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "EmailQueue",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get email queue entry by id: %w", err)
	}

	return email, nil
}

func (r *emailQueueRepository) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	if maxCount <= 0 {
		return []*models.EmailQueue{}, nil
	}

	query := `
		SELECT id, to_email, to_name, subject, body_text, body_html, attachments,
		       status, attempts, max_attempts, last_attempt_at, last_error,
		       scheduled_for, created_at
		FROM email_queue
		WHERE status = ? AND scheduled_for <= ?
		ORDER BY scheduled_for ASC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, models.EmailStatusPending, time.Now(), maxCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending emails: %w", err)
	}
	defer rows.Close()

	var emails []*models.EmailQueue
	for rows.Next() {
		email := &models.EmailQueue{}
		err := rows.Scan(
			&email.ID,
			&email.ToEmail,
			&email.ToName,
			&email.Subject,
			&email.BodyText,
			&email.BodyHTML,
			&email.Attachments,
			&email.Status,
			&email.Attempts,
			&email.MaxAttempts,
			&email.LastAttemptAt,
			&email.LastError,
			&email.ScheduledFor,
			&email.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email queue entry: %w", err)
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email queue entries: %w", err)
	}

	return emails, nil
}

func (r *emailQueueRepository) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	query := `
		SELECT id, to_email, to_name, subject, body_text, body_html, attachments,
		       status, attempts, max_attempts, last_attempt_at, last_error,
		       scheduled_for, created_at
		FROM email_queue
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get emails by status: %w", err)
	}
	defer rows.Close()

	var emails []*models.EmailQueue
	for rows.Next() {
		email := &models.EmailQueue{}
		err := rows.Scan(
			&email.ID,
			&email.ToEmail,
			&email.ToName,
			&email.Subject,
			&email.BodyText,
			&email.BodyHTML,
			&email.Attachments,
			&email.Status,
			&email.Attempts,
			&email.MaxAttempts,
			&email.LastAttemptAt,
			&email.LastError,
			&email.ScheduledFor,
			&email.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email queue entry: %w", err)
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email queue entries: %w", err)
	}

	return emails, nil
}

func (r *emailQueueRepository) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	query := `
		SELECT id, to_email, to_name, subject, body_text, body_html, attachments,
		       status, attempts, max_attempts, last_attempt_at, last_error,
		       scheduled_for, created_at
		FROM email_queue
		WHERE to_email = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, email, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get emails by recipient: %w", err)
	}
	defer rows.Close()

	var emails []*models.EmailQueue
	for rows.Next() {
		emailEntry := &models.EmailQueue{}
		err := rows.Scan(
			&emailEntry.ID,
			&emailEntry.ToEmail,
			&emailEntry.ToName,
			&emailEntry.Subject,
			&emailEntry.BodyText,
			&emailEntry.BodyHTML,
			&emailEntry.Attachments,
			&emailEntry.Status,
			&emailEntry.Attempts,
			&emailEntry.MaxAttempts,
			&emailEntry.LastAttemptAt,
			&emailEntry.LastError,
			&emailEntry.ScheduledFor,
			&emailEntry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email queue entry: %w", err)
		}
		emails = append(emails, emailEntry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email queue entries: %w", err)
	}

	return emails, nil
}

func (r *emailQueueRepository) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	query := `
		UPDATE email_queue
		SET status = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	query := `
		UPDATE email_queue
		SET attempts = attempts + 1,
		    last_error = ?,
		    last_attempt_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, errorMsg, now, id)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) MarkSending(ctx context.Context, id int64) error {
	query := `
		UPDATE email_queue
		SET status = ?
		WHERE id = ? AND status = ?
	`

	result, err := r.db.Exec(ctx, query, models.EmailStatusSending, id, models.EmailStatusPending)
	if err != nil {
		return fmt.Errorf("failed to mark email as sending: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) MarkSent(ctx context.Context, id int64) error {
	query := `
		UPDATE email_queue
		SET status = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, models.EmailStatusSent, id)
	if err != nil {
		return fmt.Errorf("failed to mark email as sent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	query := `
		UPDATE email_queue
		SET status = ?,
		    last_error = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, models.EmailStatusFailed, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to mark email as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) MarkCancelled(ctx context.Context, id int64) error {
	query := `
		UPDATE email_queue
		SET status = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, models.EmailStatusCancelled, id)
	if err != nil {
		return fmt.Errorf("failed to mark email as cancelled: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	query := `
		UPDATE email_queue
		SET scheduled_for = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query, scheduledFor, id)
	if err != nil {
		return fmt.Errorf("failed to reschedule email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "EmailQueue",
			ID:       id,
		}
	}

	return nil
}

func (r *emailQueueRepository) GetStats(ctx context.Context) (*EmailQueueStats, error) {
	query := `
		SELECT 
			status,
			COUNT(*) as count
		FROM email_queue
		GROUP BY status
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get email queue stats: %w", err)
	}
	defer rows.Close()

	stats := &EmailQueueStats{}
	for rows.Next() {
		var status models.EmailStatus
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		switch status {
		case models.EmailStatusPending:
			stats.PendingCount = count
		case models.EmailStatusSending:
			stats.SendingCount = count
		case models.EmailStatusSent:
			stats.SentCount = count
		case models.EmailStatusFailed:
			stats.FailedCount = count
		case models.EmailStatusCancelled:
			stats.CancelledCount = count
		}
		stats.TotalCount += count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stats: %w", err)
	}

	return stats, nil
}
