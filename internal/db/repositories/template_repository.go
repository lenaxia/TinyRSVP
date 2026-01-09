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

type TemplateRepository interface {
	Create(ctx context.Context, template *models.Template) error
	GetByID(ctx context.Context, id int64) (*models.Template, error)
	GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error)
	GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
	List(ctx context.Context, filters *TemplateFilters) ([]*models.Template, error)
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id int64) error
	SetActive(ctx context.Context, id int64, active bool) error
	IsTemplateInUse(ctx context.Context, id int64) (bool, error)
	SetDefault(ctx context.Context, id int64) error
}

type TemplateFilters struct {
	EventID   *int64
	Type      *models.TemplateType
	IsDefault *bool
	IsActive  *bool
	CreatedBy *int64
	Limit     int
	Offset    int
}

type templateRepository struct {
	db db.Database
}

func NewTemplateRepository(database db.Database) TemplateRepository {
	return &templateRepository{db: database}
}

func (r *templateRepository) Create(ctx context.Context, template *models.Template) error {
	if err := template.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO templates (
			event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	version := 1

	result, err := r.db.Exec(ctx, query,
		template.EventID,
		template.Name,
		template.Type,
		template.Description,
		template.HTMLContent,
		template.TextContent,
		template.CSSContent,
		template.IsDefault,
		template.IsActive,
		version,
		template.CreatedBy,
		now,
		now,
	)

	if err != nil {
		if isForeignKeyConstraintError(err) {
			return fmt.Errorf("invalid event_id or created_by: foreign key constraint failed")
		}
		return fmt.Errorf("failed to create template: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	template.ID = id
	template.Version = version
	template.CreatedAt = now
	template.UpdatedAt = now

	return nil
}

func (r *templateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at
		FROM templates
		WHERE id = ?
	`

	template := &models.Template{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.EventID,
		&template.Name,
		&template.Type,
		&template.Description,
		&template.HTMLContent,
		&template.TextContent,
		&template.CSSContent,
		&template.IsDefault,
		&template.IsActive,
		&template.Version,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Template",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get template by id: %w", err)
	}

	return template, nil
}

func (r *templateRepository) GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at
		FROM templates
		WHERE event_id = ? AND type = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	template := &models.Template{}
	err := r.db.QueryRow(ctx, query, eventID, templateType).Scan(
		&template.ID,
		&template.EventID,
		&template.Name,
		&template.Type,
		&template.Description,
		&template.HTMLContent,
		&template.TextContent,
		&template.CSSContent,
		&template.IsDefault,
		&template.IsActive,
		&template.Version,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Template",
				ID:       0,
			}
		}
		return nil, fmt.Errorf("failed to get template by event and type: %w", err)
	}

	return template, nil
}

func (r *templateRepository) GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at
		FROM templates
		WHERE type = ? AND is_default = 1 AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	template := &models.Template{}
	err := r.db.QueryRow(ctx, query, templateType).Scan(
		&template.ID,
		&template.EventID,
		&template.Name,
		&template.Type,
		&template.Description,
		&template.HTMLContent,
		&template.TextContent,
		&template.CSSContent,
		&template.IsDefault,
		&template.IsActive,
		&template.Version,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Template",
				ID:       0,
			}
		}
		return nil, fmt.Errorf("failed to get default template by type: %w", err)
	}

	return template, nil
}

func (r *templateRepository) List(ctx context.Context, filters *TemplateFilters) ([]*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at
		FROM templates
		WHERE 1=1
	`

	args := []interface{}{}

	if filters != nil {
		if filters.EventID != nil {
			query += " AND event_id = ?"
			args = append(args, *filters.EventID)
		}

		if filters.Type != nil {
			query += " AND type = ?"
			args = append(args, *filters.Type)
		}

		if filters.IsDefault != nil {
			query += " AND is_default = ?"
			args = append(args, *filters.IsDefault)
		}

		if filters.IsActive != nil {
			query += " AND is_active = ?"
			args = append(args, *filters.IsActive)
		}

		if filters.CreatedBy != nil {
			query += " AND created_by = ?"
			args = append(args, *filters.CreatedBy)
		}
	}

	query += " ORDER BY created_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters != nil && filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		template := &models.Template{}
		err := rows.Scan(
			&template.ID,
			&template.EventID,
			&template.Name,
			&template.Type,
			&template.Description,
			&template.HTMLContent,
			&template.TextContent,
			&template.CSSContent,
			&template.IsDefault,
			&template.IsActive,
			&template.Version,
			&template.CreatedBy,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

func (r *templateRepository) Update(ctx context.Context, template *models.Template) error {
	if err := template.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE templates
		SET name = ?, description = ?,
			html_content = ?, text_content = ?, css_content = ?,
			version = version + 1, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		template.Name,
		template.Description,
		template.HTMLContent,
		template.TextContent,
		template.CSSContent,
		now,
		template.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Template",
			ID:       template.ID,
		}
	}

	template.Version++
	template.UpdatedAt = now

	return nil
}

func (r *templateRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM templates WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Template",
			ID:       id,
		}
	}

	return nil
}

func (r *templateRepository) SetActive(ctx context.Context, id int64, active bool) error {
	query := `
		UPDATE templates
		SET is_active = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, active, now, id)
	if err != nil {
		return fmt.Errorf("failed to set template active status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Template",
			ID:       id,
		}
	}

	return nil
}

func (r *templateRepository) IsTemplateInUse(ctx context.Context, id int64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM events
		WHERE template_id = ?
	`

	var count int
	err := r.db.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if template is in use: %w", err)
	}

	return count > 0, nil
}

func (r *templateRepository) SetDefault(ctx context.Context, id int64) error {
	template, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	
	err = r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		unsetQuery := `
			UPDATE templates
			SET is_default = 0, updated_at = ?
			WHERE type = ? AND is_default = 1
		`

		_, err := tx.ExecContext(ctx, unsetQuery, now, template.Type)
		if err != nil {
			return fmt.Errorf("failed to unset previous default: %w", err)
		}

		setQuery := `
			UPDATE templates
			SET is_default = 1, updated_at = ?
			WHERE id = ?
		`

		result, err := tx.ExecContext(ctx, setQuery, now, id)
		if err != nil {
			return fmt.Errorf("failed to set new default: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return &models.NotFoundError{
				Resource: "Template",
				ID:       id,
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
