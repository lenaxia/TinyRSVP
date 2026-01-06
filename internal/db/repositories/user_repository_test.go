package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/tinyrsvp/internal/db"
	"github.com/yourusername/tinyrsvp/internal/models"
)

func setupTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestUserRepository_Create(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
		errType interface{}
	}{
		{
			name: "valid user",
			user: &models.User{
				Email: "test@example.com",
				Name:  "Test User",
				Role:  models.RoleEventManager,
			},
			wantErr: false,
		},
		{
			name: "valid admin user",
			user: &models.User{
				Email: "admin@example.com",
				Name:  "Admin User",
				Role:  models.RoleAdmin,
			},
			wantErr: false,
		},
		{
			name: "user with OIDC subject",
			user: &models.User{
				Email:       "oidc@example.com",
				Name:        "OIDC User",
				Role:        models.RoleEventManager,
				OIDCSubject: stringPtr("google-oauth2|123456"),
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			user: &models.User{
				Email: "test@example.com",
				Name:  "Another User",
				Role:  models.RoleAdmin,
			},
			wantErr: true,
			errType: &models.ConflictError{},
		},
		{
			name: "empty email",
			user: &models.User{
				Email: "",
				Name:  "No Email",
				Role:  models.RoleEventManager,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Expected error type %T, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if tt.user.ID == 0 {
					t.Error("Expected ID to be set")
				}
				if tt.user.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.user.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	user := &models.User{
		Email: "getbyid@example.com",
		Name:  "Get By ID User",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      999999,
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Expected error type %T, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if found.ID != tt.id {
					t.Errorf("Expected ID %d, got %d", tt.id, found.ID)
				}
				if found.Email != user.Email {
					t.Errorf("Expected email %s, got %s", user.Email, found.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	user := &models.User{
		Email: "getbyemail@example.com",
		Name:  "Get By Email User",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing email",
			email:   "getbyemail@example.com",
			wantErr: false,
		},
		{
			name:    "non-existing email",
			email:   "notfound@example.com",
			wantErr: true,
			errType: &models.NotFoundError{},
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.GetByEmail(ctx, tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Expected error type %T, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if found.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, found.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetByOIDCSubject(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	oidcSubject := "google-oauth2|test123"
	user := &models.User{
		Email:       "oidc@example.com",
		Name:        "OIDC User",
		Role:        models.RoleEventManager,
		OIDCSubject: &oidcSubject,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name    string
		subject string
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing OIDC subject",
			subject: oidcSubject,
			wantErr: false,
		},
		{
			name:    "non-existing OIDC subject",
			subject: "google-oauth2|notfound",
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.GetByOIDCSubject(ctx, tt.subject)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByOIDCSubject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.As(err, &tt.errType) {
					t.Errorf("Expected error type %T, got %T", tt.errType, err)
				}
			}

			if !tt.wantErr {
				if found.OIDCSubject == nil || *found.OIDCSubject != tt.subject {
					t.Errorf("Expected OIDC subject %s, got %v", tt.subject, found.OIDCSubject)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	user := &models.User{
		Email: "update@example.com",
		Name:  "Original Name",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("update name", func(t *testing.T) {
		user.Name = "Updated Name"
		if err := repo.Update(ctx, user); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		found, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if found.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", found.Name)
		}
	})

	t.Run("update role", func(t *testing.T) {
		user.Role = models.RoleAdmin
		if err := repo.Update(ctx, user); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		found, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if found.Role != models.RoleAdmin {
			t.Errorf("Expected role %s, got %s", models.RoleAdmin, found.Role)
		}
	})

	t.Run("update non-existing user", func(t *testing.T) {
		nonExisting := &models.User{
			ID:    999999,
			Email: "nonexisting@example.com",
			Name:  "Non Existing",
			Role:  models.RoleEventManager,
		}

		err := repo.Update(ctx, nonExisting)
		if err == nil {
			t.Error("Expected error for non-existing user")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	user := &models.User{
		Email: "delete@example.com",
		Name:  "Delete User",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("delete existing user", func(t *testing.T) {
		if err := repo.Delete(ctx, user.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err := repo.GetByID(ctx, user.ID)
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Error("Expected NotFoundError after deletion")
		}
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		err := repo.Delete(ctx, 999999)
		if err == nil {
			t.Error("Expected error for non-existing user")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func TestUserRepository_List(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		user := &models.User{
			Email: fmt.Sprintf("list%d@example.com", i),
			Name:  fmt.Sprintf("List User %d", i),
			Role:  models.RoleEventManager,
		}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("list all users", func(t *testing.T) {
		users, err := repo.List(ctx, 10, 0)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(users) != 5 {
			t.Errorf("Expected 5 users, got %d", len(users))
		}
	})

	t.Run("list with limit", func(t *testing.T) {
		users, err := repo.List(ctx, 2, 0)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}
	})

	t.Run("list with offset", func(t *testing.T) {
		users, err := repo.List(ctx, 10, 3)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 users (5 total - 3 offset), got %d", len(users))
		}
	})
}

func TestUserRepository_Count(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	t.Run("count empty", func(t *testing.T) {
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}

		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})

	for i := 0; i < 3; i++ {
		user := &models.User{
			Email: fmt.Sprintf("count%d@example.com", i),
			Name:  fmt.Sprintf("Count User %d", i),
			Role:  models.RoleEventManager,
		}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	t.Run("count with users", func(t *testing.T) {
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}

		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
	})
}

func TestUserRepository_IsFirstUser(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	t.Run("is first user when empty", func(t *testing.T) {
		isFirst, err := repo.IsFirstUser(ctx)
		if err != nil {
			t.Fatalf("IsFirstUser() error = %v", err)
		}

		if !isFirst {
			t.Error("Expected true for first user check")
		}
	})

	user := &models.User{
		Email: "first@example.com",
		Name:  "First User",
		Role:  models.RoleAdmin,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("is not first user after creating one", func(t *testing.T) {
		isFirst, err := repo.IsFirstUser(ctx)
		if err != nil {
			t.Fatalf("IsFirstUser() error = %v", err)
		}

		if isFirst {
			t.Error("Expected false after creating first user")
		}
	})
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	user := &models.User{
		Email: "lastlogin@example.com",
		Name:  "Last Login User",
		Role:  models.RoleEventManager,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.LastLoginAt != nil {
		t.Error("Expected LastLoginAt to be nil initially")
	}

	t.Run("update last login", func(t *testing.T) {
		if err := repo.UpdateLastLogin(ctx, user.ID); err != nil {
			t.Fatalf("UpdateLastLogin() error = %v", err)
		}

		found, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if found.LastLoginAt == nil {
			t.Error("Expected LastLoginAt to be set")
		}

		if time.Since(*found.LastLoginAt) > time.Minute {
			t.Error("LastLoginAt should be recent")
		}
	})

	t.Run("update last login for non-existing user", func(t *testing.T) {
		err := repo.UpdateLastLogin(ctx, 999999)
		if err == nil {
			t.Error("Expected error for non-existing user")
		}

		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("Expected NotFoundError, got %T", err)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
