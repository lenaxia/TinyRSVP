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

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	CreateWithBootstrapCheck(ctx context.Context, user *models.User) (isFirst bool, err error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByOIDCSubject(ctx context.Context, subject string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context) (int, error)
	CountByRole(ctx context.Context, role models.UserRole) (int, error)
	IsFirstUser(ctx context.Context) (bool, error)
	UpdateLastLogin(ctx context.Context, userID int64) error
}

type userRepository struct {
	db db.Database
}

func NewUserRepository(database db.Database) UserRepository {
	return &userRepository{db: database}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user.Email == "" {
		return &models.ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	query := `
		INSERT INTO users (email, name, role, oidc_subject, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		user.Email,
		user.Name,
		user.Role,
		user.OIDCSubject,
		now,
		now,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			if strings.Contains(err.Error(), "email") {
				return &models.ConflictError{
					Resource: "User",
					Field:    "email",
					Value:    user.Email,
				}
			}
			if strings.Contains(err.Error(), "oidc_subject") {
				return &models.ConflictError{
					Resource: "User",
					Field:    "oidc_subject",
					Value:    user.OIDCSubject,
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now

	return nil
}

func (r *userRepository) CreateWithBootstrapCheck(ctx context.Context, user *models.User) (bool, error) {
	if user.Email == "" {
		return false, &models.ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	var isFirst bool
	var insertErr error

	txErr := r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		var count int
		err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email != 'system@tinyrsvp.local'").Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count users: %w", err)
		}

		isFirst = (count == 0)

		role := models.RoleEventManager
		if isFirst {
			role = models.RoleAdmin
		}

		query := `
			INSERT INTO users (email, name, role, oidc_subject, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`

		now := time.Now()
		result, err := tx.ExecContext(ctx, query,
			user.Email,
			user.Name,
			role,
			user.OIDCSubject,
			now,
			now,
		)

		if err != nil {
			if isUniqueConstraintError(err) {
				if strings.Contains(err.Error(), "email") {
					insertErr = &models.ConflictError{
						Resource: "User",
						Field:    "email",
						Value:    user.Email,
					}
					return insertErr
				}
				if strings.Contains(err.Error(), "oidc_subject") {
					insertErr = &models.ConflictError{
						Resource: "User",
						Field:    "oidc_subject",
						Value:    user.OIDCSubject,
					}
					return insertErr
				}
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}

		user.ID = id
		user.Role = role
		user.CreatedAt = now
		user.UpdatedAt = now

		return nil
	})

	if txErr != nil {
		if insertErr != nil {
			return false, insertErr
		}
		return false, txErr
	}

	return isFirst, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, name, role, oidc_subject, created_at, updated_at, last_login_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.OIDCSubject,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "User",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, oidc_subject, created_at, updated_at, last_login_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.OIDCSubject,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "User",
				ID:       email,
			}
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByOIDCSubject(ctx context.Context, subject string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, oidc_subject, created_at, updated_at, last_login_at
		FROM users
		WHERE oidc_subject = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, subject).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.OIDCSubject,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "User",
				ID:       subject,
			}
		}
		return nil, fmt.Errorf("failed to get user by oidc subject: %w", err)
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = ?, name = ?, role = ?, oidc_subject = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query,
		user.Email,
		user.Name,
		user.Role,
		user.OIDCSubject,
		now,
		user.ID,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			if strings.Contains(err.Error(), "email") {
				return &models.ConflictError{
					Resource: "User",
					Field:    "email",
					Value:    user.Email,
				}
			}
			if strings.Contains(err.Error(), "oidc_subject") {
				return &models.ConflictError{
					Resource: "User",
					Field:    "oidc_subject",
					Value:    user.OIDCSubject,
				}
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "User",
			ID:       user.ID,
		}
	}

	user.UpdatedAt = now

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "User",
			ID:       id,
		}
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, name, role, oidc_subject, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.OIDCSubject,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func (r *userRepository) CountByRole(ctx context.Context, role models.UserRole) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = ?`

	var count int
	err := r.db.QueryRow(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	return count, nil
}

func (r *userRepository) IsFirstUser(ctx context.Context) (bool, error) {
	count, err := r.Count(ctx)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET last_login_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "User",
			ID:       userID,
		}
	}

	return nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "duplicate key")
}
