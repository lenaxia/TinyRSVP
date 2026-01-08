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

type QuestionRepository interface {
	Create(ctx context.Context, question *models.PreferenceQuestion) error
	GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error)
	GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
	Update(ctx context.Context, question *models.PreferenceQuestion) error
	Delete(ctx context.Context, id int64) error
	Reorder(ctx context.Context, eventID int64, questionIDs []int64) error
}

type questionRepository struct {
	db db.Database
}

func NewQuestionRepository(database db.Database) QuestionRepository {
	return &questionRepository{db: database}
}

func (r *questionRepository) Create(ctx context.Context, question *models.PreferenceQuestion) error {
	query := `
		INSERT INTO preference_questions (
			event_id, question_text, question_type, options, required, display_order,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	result, err := r.db.Exec(ctx, query,
		question.EventID,
		question.QuestionText,
		question.QuestionType,
		question.Options,
		question.Required,
		question.DisplayOrder,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create question: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	question.ID = id
	question.CreatedAt = now
	question.UpdatedAt = now

	return nil
}

func (r *questionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	query := `
		SELECT id, event_id, question_text, question_type, options, required, display_order,
			created_at, updated_at
		FROM preference_questions
		WHERE id = ?
	`

	var question models.PreferenceQuestion
	err := r.db.QueryRow(ctx, query, id).Scan(
		&question.ID,
		&question.EventID,
		&question.QuestionText,
		&question.QuestionType,
		&question.Options,
		&question.Required,
		&question.DisplayOrder,
		&question.CreatedAt,
		&question.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "preference_question",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	return &question, nil
}

func (r *questionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	query := `
		SELECT id, event_id, question_text, question_type, options, required, display_order,
			created_at, updated_at
		FROM preference_questions
		WHERE event_id = ?
		ORDER BY display_order ASC, id ASC
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []*models.PreferenceQuestion
	for rows.Next() {
		var question models.PreferenceQuestion
		err := rows.Scan(
			&question.ID,
			&question.EventID,
			&question.QuestionText,
			&question.QuestionType,
			&question.Options,
			&question.Required,
			&question.DisplayOrder,
			&question.CreatedAt,
			&question.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}
		questions = append(questions, &question)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating questions: %w", err)
	}

	return questions, nil
}

func (r *questionRepository) Update(ctx context.Context, question *models.PreferenceQuestion) error {
	query := `
		UPDATE preference_questions
		SET question_text = ?, question_type = ?, options = ?, required = ?,
			display_order = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()

	result, err := r.db.Exec(ctx, query,
		question.QuestionText,
		question.QuestionType,
		question.Options,
		question.Required,
		question.DisplayOrder,
		now,
		question.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "preference_question",
			ID:       question.ID,
		}
	}

	question.UpdatedAt = now

	return nil
}

func (r *questionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM preference_questions WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "preference_question",
			ID:       id,
		}
	}

	return nil
}

func (r *questionRepository) Reorder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		verifyQuery := `
			SELECT COUNT(*) FROM preference_questions
			WHERE id = ? AND event_id = ?
		`

		for _, id := range questionIDs {
			var count int
			err := tx.QueryRowContext(ctx, verifyQuery, id, eventID).Scan(&count)
			if err != nil {
				return fmt.Errorf("failed to verify question %d: %w", id, err)
			}
			if count == 0 {
				return &models.ValidationError{
					Field:   "question_ids",
					Message: fmt.Sprintf("question %d does not belong to event %d", id, eventID),
				}
			}
		}

		updateQuery := `
			UPDATE preference_questions
			SET display_order = ?, updated_at = ?
			WHERE id = ?
		`

		now := time.Now()
		for i, id := range questionIDs {
			_, err := tx.ExecContext(ctx, updateQuery, i, now, id)
			if err != nil {
				return fmt.Errorf("failed to update display order for question %d: %w", id, err)
			}
		}

		return nil
	})
}
