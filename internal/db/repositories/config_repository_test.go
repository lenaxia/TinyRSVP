package repositories

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestConfigRepository_Get(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	if err := repo.Set(ctx, "test_key", "test_value"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing key",
			key:     "test_key",
			wantErr: false,
		},
		{
			name:    "non-existing key",
			key:     "non_existing_key",
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := repo.Get(ctx, tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Expected error type %T, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if config.Key != tt.key {
					t.Errorf("Expected key %s, got %s", tt.key, config.Key)
				}
				if config.Value != "test_value" {
					t.Errorf("Expected value 'test_value', got %s", config.Value)
				}
			}
		})
	}
}

func TestConfigRepository_Set(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "set new key",
			key:     "new_key",
			value:   "new_value",
			wantErr: false,
		},
		{
			name:    "set another key",
			key:     "another_key",
			value:   "another_value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Set(ctx, tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				config, err := repo.Get(ctx, tt.key)
				if err != nil {
					t.Fatalf("Get() error = %v", err)
				}

				if config.Value != tt.value {
					t.Errorf("Expected value %s, got %s", tt.value, config.Value)
				}
			}
		})
	}

	t.Run("update existing key", func(t *testing.T) {
		if err := repo.Set(ctx, "new_key", "updated_value"); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		config, err := repo.Get(ctx, "new_key")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if config.Value != "updated_value" {
			t.Errorf("Expected value 'updated_value', got %s", config.Value)
		}
	})
}

func TestConfigRepository_Delete(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	if err := repo.Set(ctx, "delete_key", "delete_value"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	t.Run("delete existing key", func(t *testing.T) {
		if err := repo.Delete(ctx, "delete_key"); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err := repo.Get(ctx, "delete_key")
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Error("Expected NotFoundError after deletion")
		}
	})

	t.Run("delete non-existing key", func(t *testing.T) {
		err := repo.Delete(ctx, "non_existing_key")
		if err == nil {
			t.Error("Expected error for non-existing key")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestConfigRepository_GetAll(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	t.Run("get all when empty", func(t *testing.T) {
		configs, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll() error = %v", err)
		}

		if len(configs) != 0 {
			t.Errorf("Expected 0 configs, got %d", len(configs))
		}
	})

	configData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range configData {
		if err := repo.Set(ctx, k, v); err != nil {
			t.Fatalf("Failed to set config %s: %v", k, err)
		}
	}

	t.Run("get all configs", func(t *testing.T) {
		configs, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll() error = %v", err)
		}

		if len(configs) != 3 {
			t.Errorf("Expected 3 configs, got %d", len(configs))
		}

		configMap := make(map[string]string)
		for _, c := range configs {
			configMap[c.Key] = c.Value
		}

		for k, v := range configData {
			if configMap[k] != v {
				t.Errorf("Expected value %s for key %s, got %s", v, k, configMap[k])
			}
		}
	})
}

func TestConfigRepository_GetHMACSecret(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	t.Run("auto-generate HMAC secret on first call", func(t *testing.T) {
		secret, err := repo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		if len(secret) < 32 {
			t.Errorf("Expected secret >= 32 bytes, got %d", len(secret))
		}
	})

	t.Run("return same secret on subsequent calls", func(t *testing.T) {
		secret1, err := repo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		secret2, err := repo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		if !bytes.Equal(secret1, secret2) {
			t.Error("Expected same secret on subsequent calls")
		}
	})
}

func TestConfigRepository_SetHMACSecret(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	customSecret := []byte("my-custom-secret-that-is-32-bytes-long-exactly!")

	t.Run("set custom HMAC secret", func(t *testing.T) {
		if err := repo.SetHMACSecret(ctx, customSecret); err != nil {
			t.Fatalf("SetHMACSecret() error = %v", err)
		}

		secret, err := repo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		if !bytes.Equal(secret, customSecret) {
			t.Error("Expected custom secret to be returned")
		}
	})

	t.Run("overwrite existing HMAC secret", func(t *testing.T) {
		newSecret := []byte("another-secret-that-is-also-32-bytes-long-now!")

		if err := repo.SetHMACSecret(ctx, newSecret); err != nil {
			t.Fatalf("SetHMACSecret() error = %v", err)
		}

		secret, err := repo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		if !bytes.Equal(secret, newSecret) {
			t.Error("Expected new secret to be returned")
		}

		if bytes.Equal(secret, customSecret) {
			t.Error("Expected old secret to be overwritten")
		}
	})
}

func TestConfigRepository_HMACSecretPersistence(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	t.Run("HMAC secret persists across repository instances", func(t *testing.T) {
		secret1, err := repo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		newRepo := NewConfigRepository(database)
		secret2, err := newRepo.GetHMACSecret(ctx)
		if err != nil {
			t.Fatalf("GetHMACSecret() error = %v", err)
		}

		if !bytes.Equal(secret1, secret2) {
			t.Error("Expected HMAC secret to persist across repository instances")
		}
	})
}

func TestConfigRepository_SetAndGetMultipleKeys(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	testData := map[string]string{
		"smtp_host":     "smtp.example.com",
		"smtp_port":     "587",
		"smtp_username": "user@example.com",
		"app_name":      "TinyRSVP",
		"max_file_size": "10485760",
	}

	t.Run("set multiple keys", func(t *testing.T) {
		for k, v := range testData {
			if err := repo.Set(ctx, k, v); err != nil {
				t.Fatalf("Failed to set config %s: %v", k, err)
			}
		}
	})

	t.Run("get each key individually", func(t *testing.T) {
		for k, expectedValue := range testData {
			config, err := repo.Get(ctx, k)
			if err != nil {
				t.Fatalf("Get() error for key %s: %v", k, err)
			}

			if config.Value != expectedValue {
				t.Errorf("Expected value %s for key %s, got %s", expectedValue, k, config.Value)
			}
		}
	})

	t.Run("get all returns all keys", func(t *testing.T) {
		configs, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll() error = %v", err)
		}

		if len(configs) < len(testData) {
			t.Errorf("Expected at least %d configs, got %d", len(testData), len(configs))
		}
	})
}

func TestConfigRepository_UpdateExistingKey(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	key := "update_test"
	originalValue := "original"
	updatedValue := "updated"

	t.Run("set initial value", func(t *testing.T) {
		if err := repo.Set(ctx, key, originalValue); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		config, err := repo.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if config.Value != originalValue {
			t.Errorf("Expected value %s, got %s", originalValue, config.Value)
		}
	})

	t.Run("update value", func(t *testing.T) {
		if err := repo.Set(ctx, key, updatedValue); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		config, err := repo.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if config.Value != updatedValue {
			t.Errorf("Expected value %s, got %s", updatedValue, config.Value)
		}
	})

	t.Run("updated_at timestamp changes", func(t *testing.T) {
		config1, err := repo.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if err := repo.Set(ctx, key, "another_value"); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		config2, err := repo.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if !config2.UpdatedAt.After(config1.UpdatedAt) {
			t.Error("Expected UpdatedAt to be later after update")
		}
	})
}

func TestConfigRepository_EmptyValues(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	t.Run("can store empty string value", func(t *testing.T) {
		if err := repo.Set(ctx, "empty_key", ""); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		config, err := repo.Get(ctx, "empty_key")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if config.Value != "" {
			t.Errorf("Expected empty value, got %s", config.Value)
		}
	})
}

func TestConfigRepository_LargeValues(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	t.Run("can store large value", func(t *testing.T) {
		largeValue := string(make([]byte, 10000))
		for i := range largeValue {
			largeValue = largeValue[:i] + "a" + largeValue[i+1:]
		}

		if err := repo.Set(ctx, "large_key", largeValue); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		config, err := repo.Get(ctx, "large_key")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if len(config.Value) != 10000 {
			t.Errorf("Expected value length 10000, got %d", len(config.Value))
		}
	})
}

func TestConfigRepository_SpecialCharacters(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewConfigRepository(database)
	ctx := context.Background()

	specialValues := []string{
		"value with spaces",
		"value\nwith\nnewlines",
		"value\twith\ttabs",
		"value'with'quotes",
		`value"with"double"quotes`,
		"value;with;semicolons",
		"value=with=equals",
		"value&with&ampersands",
	}

	for i, value := range specialValues {
		t.Run(fmt.Sprintf("special_value_%d", i), func(t *testing.T) {
			key := fmt.Sprintf("special_key_%d", i)

			if err := repo.Set(ctx, key, value); err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			config, err := repo.Get(ctx, key)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			if config.Value != value {
				t.Errorf("Expected value %q, got %q", value, config.Value)
			}
		})
	}
}
