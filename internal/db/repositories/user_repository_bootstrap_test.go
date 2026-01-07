package repositories

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestUserRepository_CreateWithBootstrapCheck(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	t.Run("first user gets admin role", func(t *testing.T) {
		user := &models.User{
			Email: "first@example.com",
			Name:  "First User",
		}

		isFirst, err := repo.CreateWithBootstrapCheck(ctx, user)
		if err != nil {
			t.Fatalf("CreateWithBootstrapCheck failed: %v", err)
		}

		if !isFirst {
			t.Error("Expected isFirst to be true for first user")
		}

		if user.ID == 0 {
			t.Error("Expected user ID to be set")
		}

		if user.Role != models.RoleAdmin {
			t.Errorf("Expected first user to have admin role, got %s", user.Role)
		}
	})

	t.Run("second user is not first", func(t *testing.T) {
		user := &models.User{
			Email: "second@example.com",
			Name:  "Second User",
		}

		isFirst, err := repo.CreateWithBootstrapCheck(ctx, user)
		if err != nil {
			t.Fatalf("CreateWithBootstrapCheck failed: %v", err)
		}

		if isFirst {
			t.Error("Expected isFirst to be false for second user")
		}

		if user.Role != models.RoleEventManager {
			t.Errorf("Expected second user to have event manager role, got %s", user.Role)
		}
	})
}

func TestUserRepository_CreateWithBootstrapCheck_Concurrent(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			user := &models.User{
				Email: fmt.Sprintf("user%d@example.com", n),
				Name:  fmt.Sprintf("User %d", n),
			}
			isFirst, err := repo.CreateWithBootstrapCheck(ctx, user)
			if err == nil {
				results <- isFirst
			}
		}(i)
	}

	wg.Wait()
	close(results)

	firstCount := 0
	for isFirst := range results {
		if isFirst {
			firstCount++
		}
	}

	if firstCount != 1 {
		t.Errorf("Expected exactly 1 first user, got %d", firstCount)
	}
}

func TestUserRepository_CreateWithBootstrapCheck_ConcurrentWithDelay(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	repo := NewUserRepository(database)
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(time.Millisecond * time.Duration(n))
			user := &models.User{
				Email: fmt.Sprintf("delayed%d@example.com", n),
				Name:  fmt.Sprintf("Delayed User %d", n),
			}
			isFirst, err := repo.CreateWithBootstrapCheck(ctx, user)
			if err == nil {
				results <- isFirst
			}
		}(i)
	}

	wg.Wait()
	close(results)

	firstCount := 0
	totalSuccess := 0
	for isFirst := range results {
		totalSuccess++
		if isFirst {
			firstCount++
		}
	}

	if firstCount != 1 {
		t.Errorf("Expected exactly 1 first user, got %d (total success: %d)", firstCount, totalSuccess)
	}
}
