package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/tinyrsvp/internal/db"
	"github.com/yourusername/tinyrsvp/internal/db/repositories"
	"github.com/yourusername/tinyrsvp/internal/models"
)

func setupIntegrationTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 10,
		MaxIdleConns: 10,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestUserService_Bootstrap_FirstUserIsAdmin_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	user, err := service.CreateUser(ctx, "admin@example.com", "Admin User", nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Role != models.RoleAdmin {
		t.Errorf("Expected first user to be admin, got %s", user.Role)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
}

func TestUserService_Bootstrap_SecondUserIsEventManager_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	_, err := service.CreateUser(ctx, "admin@example.com", "Admin User", nil)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	user, err := service.CreateUser(ctx, "user@example.com", "Regular User", nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Role != models.RoleEventManager {
		t.Errorf("Expected second user to be event manager, got %s", user.Role)
	}
}

func TestUserService_Bootstrap_ThirdUserIsEventManager_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	_, err := service.CreateUser(ctx, "admin@example.com", "Admin User", nil)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	_, err = service.CreateUser(ctx, "user1@example.com", "User One", nil)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	user, err := service.CreateUser(ctx, "user2@example.com", "User Two", nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Role != models.RoleEventManager {
		t.Errorf("Expected third user to be event manager, got %s", user.Role)
	}
}

func TestUserService_Bootstrap_ConcurrentFirstUser_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan *models.User, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			user, err := service.CreateUser(ctx,
				fmt.Sprintf("user%d@example.com", n),
				fmt.Sprintf("User %d", n),
				nil)
			if err == nil {
				results <- user
			}
		}(i)
	}

	wg.Wait()
	close(results)

	adminCount := 0
	eventManagerCount := 0
	successCount := 0

	for user := range results {
		successCount++
		if user.Role == models.RoleAdmin {
			adminCount++
		} else if user.Role == models.RoleEventManager {
			eventManagerCount++
		}
	}

	if adminCount != 1 {
		t.Errorf("Expected exactly 1 admin, got %d", adminCount)
	}

	if successCount < 1 {
		t.Errorf("Expected at least 1 successful user creation, got %d", successCount)
	}

	if adminCount == 1 && successCount > 1 {
		expectedManagers := successCount - 1
		if eventManagerCount != expectedManagers {
			t.Errorf("Expected %d event managers, got %d", expectedManagers, eventManagerCount)
		}
	}
}

func TestUserService_Bootstrap_ConcurrentFirstUserWithOIDC_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan *models.User, 3)
	errors := make(chan error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			subject := fmt.Sprintf("oidc-subject-%d", n)
			user, err := service.CreateUser(ctx,
				fmt.Sprintf("oidc%d@example.com", n),
				fmt.Sprintf("OIDC User %d", n),
				&subject)
			if err != nil {
				errors <- err
			} else {
				results <- user
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	adminCount := 0
	for user := range results {
		if user.Role == models.RoleAdmin {
			adminCount++
		}
	}

	if adminCount != 1 {
		t.Errorf("Expected exactly 1 admin with OIDC, got %d", adminCount)
	}
}

func TestUserService_GetOrCreateUser_FirstUserIsAdmin_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	subject := "oidc-first-user"
	user, err := service.GetOrCreateUser(ctx, "first@example.com", "First User", &subject)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	if user.Role != models.RoleAdmin {
		t.Errorf("Expected first user via GetOrCreateUser to be admin, got %s", user.Role)
	}
}

func TestUserService_GetOrCreateUser_SecondUserIsEventManager_Integration(t *testing.T) {
	database := setupIntegrationTestDB(t)
	defer database.Close()

	repo := repositories.NewUserRepository(database)
	service := NewUserService(repo)
	ctx := context.Background()

	subject1 := "oidc-first"
	_, err := service.GetOrCreateUser(ctx, "first@example.com", "First User", &subject1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	subject2 := "oidc-second"
	user, err := service.GetOrCreateUser(ctx, "second@example.com", "Second User", &subject2)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	if user.Role != models.RoleEventManager {
		t.Errorf("Expected second user via GetOrCreateUser to be event manager, got %s", user.Role)
	}
}
