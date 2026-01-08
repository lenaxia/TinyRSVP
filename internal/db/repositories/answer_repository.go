package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type AnswerRepository interface {
	Create(ctx context.Context, answer *models.RSVPAnswer) error
	GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error)
	GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error)
	Update(ctx context.Context, answer *models.RSVPAnswer) error
	DeleteByRSVPID(ctx context.Context, rsvpID int64) error
}

type answerRepository struct {
	db db.Database
}

func NewAnswerRepository(database db.Database) AnswerRepository {
	return &answerRepository{db: database}
}

func (r *answerRepository) Create(ctx context.Context, answer *models.RSVPAnswer) error {
	if err := answer.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO rsvp_answers (rsvp_id, question_id, answer_text, answer_option, answer_boolean, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(ctx, query,
		answer.RSVPID,
		answer.QuestionID,
		answer.AnswerText,
		answer.AnswerOption,
		answer.AnswerBoolean,
	)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	answer.ID = id

	created, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}

	*answer = *created
	return nil
}

func (r *answerRepository) getByID(ctx context.Context, id int64) (*models.RSVPAnswer, error) {
	query := `
		SELECT id, rsvp_id, question_id, answer_text, answer_option, answer_boolean, created_at, updated_at
		FROM rsvp_answers
		WHERE id = ?
	`

	var answer models.RSVPAnswer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&answer.ID,
		&answer.RSVPID,
		&answer.QuestionID,
		&answer.AnswerText,
		&answer.AnswerOption,
		&answer.AnswerBoolean,
		&answer.CreatedAt,
		&answer.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &answer, nil
}

func (r *answerRepository) GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
	query := `
		SELECT id, rsvp_id, question_id, answer_text, answer_option, answer_boolean, created_at, updated_at
		FROM rsvp_answers
		WHERE rsvp_id = ?
		ORDER BY question_id
	`

	rows, err := r.db.Query(ctx, query, rsvpID)
	if err != nil {
		return nil, fmt.Errorf("failed to query answers: %w", err)
	}
	defer rows.Close()

	var answers []*models.RSVPAnswer
	for rows.Next() {
		var answer models.RSVPAnswer
		err := rows.Scan(
			&answer.ID,
			&answer.RSVPID,
			&answer.QuestionID,
			&answer.AnswerText,
			&answer.AnswerOption,
			&answer.AnswerBoolean,
			&answer.CreatedAt,
			&answer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan answer: %w", err)
		}
		answers = append(answers, &answer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating answers: %w", err)
	}

	return answers, nil
}

func (r *answerRepository) GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error) {
	query := `
		SELECT id, rsvp_id, question_id, answer_text, answer_option, answer_boolean, created_at, updated_at
		FROM rsvp_answers
		WHERE question_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query answers: %w", err)
	}
	defer rows.Close()

	var answers []*models.RSVPAnswer
	for rows.Next() {
		var answer models.RSVPAnswer
		err := rows.Scan(
			&answer.ID,
			&answer.RSVPID,
			&answer.QuestionID,
			&answer.AnswerText,
			&answer.AnswerOption,
			&answer.AnswerBoolean,
			&answer.CreatedAt,
			&answer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan answer: %w", err)
		}
		answers = append(answers, &answer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating answers: %w", err)
	}

	return answers, nil
}

func (r *answerRepository) Update(ctx context.Context, answer *models.RSVPAnswer) error {
	if err := answer.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE rsvp_answers
		SET answer_text = ?, answer_option = ?, answer_boolean = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query,
		answer.AnswerText,
		answer.AnswerOption,
		answer.AnswerBoolean,
		answer.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update answer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	updated, err := r.getByID(ctx, answer.ID)
	if err != nil {
		return err
	}

	*answer = *updated
	return nil
}

func (r *answerRepository) DeleteByRSVPID(ctx context.Context, rsvpID int64) error {
	query := `DELETE FROM rsvp_answers WHERE rsvp_id = ?`

	_, err := r.db.Exec(ctx, query, rsvpID)
	if err != nil {
		return fmt.Errorf("failed to delete answers: %w", err)
	}

	return nil
}
