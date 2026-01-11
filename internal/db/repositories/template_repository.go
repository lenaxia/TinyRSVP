package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
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
	GetByNameAndType(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error)
	List(ctx context.Context, filters *TemplateFilters) ([]*models.Template, error)
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id int64) error
	SetActive(ctx context.Context, id int64, active bool) error
	IsTemplateInUse(ctx context.Context, id int64) (bool, error)
	SetDefault(ctx context.Context, id int64) error
	GetTemplatesByCategory(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error)
	ListThemes(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error)
	GetComponentConfig(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error)
	UpdateComponentConfig(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error
	ValidateComponentConfig(ctx context.Context, config *models.ComponentConfiguration) error
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
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	version := 1

	tagsJSON, err := serializeTags(template.Tags)
	if err != nil {
		return fmt.Errorf("failed to serialize tags: %w", err)
	}

	var createdBy interface{}
	if template.CreatedBy == 0 {
		createdBy = nil
	} else {
		createdBy = template.CreatedBy
	}

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
		createdBy,
		now,
		now,
		template.Category,
		template.ThumbnailURL,
		template.ImageURL,
		tagsJSON,
		template.SortOrder,
		template.ComponentConfig,
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
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		FROM templates
		WHERE id = ?
	`

	template := &models.Template{}
	var tagsJSON *string
	var createdBy sql.NullInt64

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
		&createdBy,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.Category,
		&template.ThumbnailURL,
		&template.ImageURL,
		&tagsJSON,
		&template.SortOrder,
		&template.ComponentConfig,
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

	if createdBy.Valid {
		template.CreatedBy = createdBy.Int64
	} else {
		template.CreatedBy = 0
	}

	if err := deserializeTags(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return template, nil
}

func (r *templateRepository) GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		FROM templates
		WHERE event_id = ? AND type = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	template := &models.Template{}
	var tagsJSON *string
	var createdBy sql.NullInt64

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
		&createdBy,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.Category,
		&template.ThumbnailURL,
		&template.ImageURL,
		&tagsJSON,
		&template.SortOrder,
		&template.ComponentConfig,
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

	if createdBy.Valid {
		template.CreatedBy = createdBy.Int64
	} else {
		template.CreatedBy = 0
	}

	if err := deserializeTags(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return template, nil
}

func (r *templateRepository) GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		FROM templates
		WHERE type = ? AND is_default = 1 AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	template := &models.Template{}
	var tagsJSON *string
	var createdBy sql.NullInt64

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
		&createdBy,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.Category,
		&template.ThumbnailURL,
		&template.ImageURL,
		&tagsJSON,
		&template.SortOrder,
		&template.ComponentConfig,
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

	if createdBy.Valid {
		template.CreatedBy = createdBy.Int64
	} else {
		template.CreatedBy = 0
	}

	if err := deserializeTags(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return template, nil
}

func (r *templateRepository) GetByNameAndType(ctx context.Context, name string, templateType models.TemplateType) (*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		FROM templates
		WHERE name = ? AND type = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	template := &models.Template{}
	var tagsJSON *string
	var createdBy sql.NullInt64

	err := r.db.QueryRow(ctx, query, name, templateType).Scan(
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
		&createdBy,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.Category,
		&template.ThumbnailURL,
		&template.ImageURL,
		&tagsJSON,
		&template.SortOrder,
		&template.ComponentConfig,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Template",
				ID:       0,
			}
		}
		return nil, fmt.Errorf("failed to get template by name and type: %w", err)
	}

	if createdBy.Valid {
		template.CreatedBy = createdBy.Int64
	} else {
		template.CreatedBy = 0
	}

	if err := deserializeTags(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return template, nil
}

func (r *templateRepository) List(ctx context.Context, filters *TemplateFilters) ([]*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
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
		template, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
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
			category = ?, thumbnail_url = ?, image_url = ?, tags = ?, sort_order = ?,
			version = version + 1, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()

	tagsJSON, err := serializeTags(template.Tags)
	if err != nil {
		return fmt.Errorf("failed to serialize tags: %w", err)
	}

	result, err := r.db.Exec(ctx, query,
		template.Name,
		template.Description,
		template.HTMLContent,
		template.TextContent,
		template.CSSContent,
		template.Category,
		template.ThumbnailURL,
		template.ImageURL,
		tagsJSON,
		template.SortOrder,
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

func (r *templateRepository) GetTemplatesByCategory(ctx context.Context, category models.TemplateCategory) ([]*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		FROM templates
		WHERE category = ?
		ORDER BY sort_order ASC, name ASC
	`

	rows, err := r.db.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates by category: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		template, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

func (r *templateRepository) ListThemes(ctx context.Context, templateType models.TemplateType, category *models.TemplateCategory) ([]*models.Template, error) {
	query := `
		SELECT id, event_id, name, type, description,
			html_content, text_content, css_content,
			is_default, is_active, version,
			created_by, created_at, updated_at,
			category, thumbnail_url, image_url, tags, sort_order, component_config
		FROM templates
		WHERE type = ?
	`

	args := []interface{}{templateType}

	if category != nil {
		query += " AND category = ?"
		args = append(args, *category)
	}

	query += " ORDER BY sort_order ASC, name ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list themes: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		template, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

func (r *templateRepository) scanTemplate(scanner interface{ Scan(...interface{}) error }) (*models.Template, error) {
	template := &models.Template{}
	var tagsJSON *string
	var createdBy sql.NullInt64

	err := scanner.Scan(
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
		&createdBy,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.Category,
		&template.ThumbnailURL,
		&template.ImageURL,
		&tagsJSON,
		&template.SortOrder,
		&template.ComponentConfig,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan template: %w", err)
	}

	if createdBy.Valid {
		template.CreatedBy = createdBy.Int64
	} else {
		template.CreatedBy = 0
	}

	if err := deserializeTags(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to deserialize tags: %w", err)
	}

	return template, nil
}

func serializeTags(tags []string) (*string, error) {
	if tags == nil || len(tags) == 0 {
		empty := "[]"
		return &empty, nil
	}

	data, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	result := string(data)
	return &result, nil
}

func deserializeTags(tagsJSON *string, tags *[]string) error {
	if tagsJSON == nil || *tagsJSON == "" {
		*tags = []string{}
		return nil
	}

	if err := json.Unmarshal([]byte(*tagsJSON), tags); err != nil {
		return err
	}

	if *tags == nil {
		*tags = []string{}
	}

	return nil
}

func (r *templateRepository) GetComponentConfig(ctx context.Context, templateID int64) (*models.ComponentConfiguration, error) {
	query := `
		SELECT component_config
		FROM templates
		WHERE id = ?
	`

	var componentConfig *string
	err := r.db.QueryRow(ctx, query, templateID).Scan(&componentConfig)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Template",
				ID:       templateID,
			}
		}
		return nil, fmt.Errorf("failed to get component config: %w", err)
	}

	if componentConfig == nil || *componentConfig == "" {
		return nil, nil
	}

	var config models.ComponentConfiguration
	if err := json.Unmarshal([]byte(*componentConfig), &config); err != nil {
		return nil, fmt.Errorf("failed to parse component configuration: %w", err)
	}

	return &config, nil
}

func (r *templateRepository) UpdateComponentConfig(ctx context.Context, templateID int64, config *models.ComponentConfiguration) error {
	if config != nil {
		if err := r.ValidateComponentConfig(ctx, config); err != nil {
			return err
		}
	}

	var componentConfigJSON *string
	if config != nil {
		data, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal component configuration: %w", err)
		}
		str := string(data)
		componentConfigJSON = &str
	}

	query := `
		UPDATE templates
		SET component_config = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, componentConfigJSON, now, templateID)
	if err != nil {
		return fmt.Errorf("failed to update component config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Template",
			ID:       templateID,
		}
	}

	return nil
}

func (r *templateRepository) ValidateComponentConfig(ctx context.Context, config *models.ComponentConfiguration) error {
	if config == nil {
		return nil
	}

	if config.Version == "" {
		return &models.ValidationError{
			Field:   "version",
			Message: "version is required",
		}
	}

	if len(config.Components) > 50 {
		return &models.ValidationError{
			Field:   "components",
			Message: "maximum 50 components allowed",
		}
	}

	componentIDs := make(map[string]bool)
	for i, component := range config.Components {
		if component.ID == "" {
			return &models.ValidationError{
				Field:   fmt.Sprintf("components[%d].id", i),
				Message: "component ID is required",
			}
		}

		if componentIDs[component.ID] {
			return &models.ValidationError{
				Field:   fmt.Sprintf("components[%d].id", i),
				Message: fmt.Sprintf("duplicate component ID: %s", component.ID),
			}
		}
		componentIDs[component.ID] = true

		if !component.Type.IsValid() {
			return &models.ValidationError{
				Field:   fmt.Sprintf("components[%d].type", i),
				Message: fmt.Sprintf("invalid component type: %s", component.Type),
			}
		}

		if !component.Position.Mode.IsValid() {
			return &models.ValidationError{
				Field:   fmt.Sprintf("components[%d].position.mode", i),
				Message: fmt.Sprintf("invalid position mode: %s", component.Position.Mode),
			}
		}
	}

	return nil
}
