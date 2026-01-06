package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/yourusername/tinyrsvp/internal/db"
	"github.com/yourusername/tinyrsvp/internal/models"
)

const hmacSecretKey = "hmac_secret"

type ConfigRepository interface {
	Get(ctx context.Context, key string) (*models.Config, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	GetAll(ctx context.Context) ([]*models.Config, error)
	GetHMACSecret(ctx context.Context) ([]byte, error)
	SetHMACSecret(ctx context.Context, secret []byte) error
}

type configRepository struct {
	db db.Database
}

func NewConfigRepository(database db.Database) ConfigRepository {
	return &configRepository{db: database}
}

func (r *configRepository) Get(ctx context.Context, key string) (*models.Config, error) {
	query := `
		SELECT key, value, updated_at
		FROM config
		WHERE key = ?
	`

	config := &models.Config{}
	err := r.db.QueryRow(ctx, query, key).Scan(
		&config.Key,
		&config.Value,
		&config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &models.NotFoundError{
				Resource: "Config",
				ID:       key,
			}
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return config, nil
}

func (r *configRepository) Set(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, key, value, now)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	return nil
}

func (r *configRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM config WHERE key = ?`

	result, err := r.db.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "Config",
			ID:       key,
		}
	}

	return nil
}

func (r *configRepository) GetAll(ctx context.Context) ([]*models.Config, error) {
	query := `
		SELECT key, value, updated_at
		FROM config
		ORDER BY key
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.Config
	for rows.Next() {
		config := &models.Config{}
		err := rows.Scan(
			&config.Key,
			&config.Value,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating configs: %w", err)
	}

	return configs, nil
}

func (r *configRepository) GetHMACSecret(ctx context.Context) ([]byte, error) {
	config, err := r.Get(ctx, hmacSecretKey)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			secret, err := r.generateAndStoreHMACSecret(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate HMAC secret: %w", err)
			}
			return secret, nil
		}
		return nil, err
	}

	secret, err := base64.StdEncoding.DecodeString(config.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode HMAC secret: %w", err)
	}

	return secret, nil
}

func (r *configRepository) SetHMACSecret(ctx context.Context, secret []byte) error {
	encoded := base64.StdEncoding.EncodeToString(secret)
	return r.Set(ctx, hmacSecretKey, encoded)
}

func (r *configRepository) generateAndStoreHMACSecret(ctx context.Context) ([]byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate random secret: %w", err)
	}

	if err := r.SetHMACSecret(ctx, secret); err != nil {
		return nil, err
	}

	return secret, nil
}
