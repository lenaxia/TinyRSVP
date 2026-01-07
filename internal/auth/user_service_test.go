package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/yourusername/tinyrsvp/internal/models"
)

func TestUserService_CreateUser_FirstUserIsAdmin(t *testing.T) {
	mockRepo := &MockUserRepository{
		IsFirstUserFunc: func(ctx context.Context) (bool, error) {
			return true, nil
		},
		CreateFunc: func(ctx context.Context, user *models.User) error {
			user.ID = 1
			return nil
		},
	}

	service := NewUserService(mockRepo)

	user, err := service.CreateUser(context.Background(), "first@example.com", "First User", nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Role != models.RoleAdmin {
		t.Errorf("Expected first user to be admin, got %s", user.Role)
	}
}

func TestUserService_CreateUser_SecondUserIsEventManager(t *testing.T) {
	mockRepo := &MockUserRepository{
		IsFirstUserFunc: func(ctx context.Context) (bool, error) {
			return false, nil
		},
		CreateFunc: func(ctx context.Context, user *models.User) error {
			user.ID = 2
			return nil
		},
	}

	service := NewUserService(mockRepo)

	user, err := service.CreateUser(context.Background(), "second@example.com", "Second User", nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Role != models.RoleEventManager {
		t.Errorf("Expected second user to be event manager, got %s", user.Role)
	}
}

func TestUserService_GetOrCreateUser_ExistingByOIDCSubject(t *testing.T) {
	subject := "oidc-subject-123"
	mockRepo := &MockUserRepository{
		GetByOIDCSubjectFunc: func(ctx context.Context, sub string) (*models.User, error) {
			if sub == subject {
				return &models.User{
					ID:          1,
					Email:       "existing@example.com",
					Name:        "Existing User",
					Role:        models.RoleAdmin,
					OIDCSubject: &subject,
				}, nil
			}
			return nil, &models.NotFoundError{Resource: "User", ID: sub}
		},
		UpdateLastLoginFunc: func(ctx context.Context, userID int64) error {
			if userID != 1 {
				return fmt.Errorf("unexpected user ID: %d", userID)
			}
			return nil
		},
	}

	service := NewUserService(mockRepo)

	user, err := service.GetOrCreateUser(context.Background(), "existing@example.com", "Existing User", &subject)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}

	if user.Email != "existing@example.com" {
		t.Errorf("Expected email existing@example.com, got %s", user.Email)
	}
}

func TestUserService_GetOrCreateUser_ExistingByEmail(t *testing.T) {
	subject := "oidc-subject-123"
	mockRepo := &MockUserRepository{
		GetByOIDCSubjectFunc: func(ctx context.Context, sub string) (*models.User, error) {
			return nil, &models.NotFoundError{Resource: "User", ID: sub}
		},
		GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			if email == "existing@example.com" {
				return &models.User{
					ID:          2,
					Email:       email,
					Name:        "Existing User",
					Role:        models.RoleEventManager,
					OIDCSubject: nil,
				}, nil
			}
			return nil, &models.NotFoundError{Resource: "User", ID: email}
		},
		UpdateFunc: func(ctx context.Context, user *models.User) error {
			return nil
		},
		UpdateLastLoginFunc: func(ctx context.Context, userID int64) error {
			return nil
		},
	}

	service := NewUserService(mockRepo)

	user, err := service.GetOrCreateUser(context.Background(), "existing@example.com", "Existing User", &subject)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	if user.ID != 2 {
		t.Errorf("Expected user ID 2, got %d", user.ID)
	}

	if user.OIDCSubject == nil {
		t.Error("Expected OIDC subject to be updated")
	}
}

func TestUserService_GetOrCreateUser_NewUser(t *testing.T) {
	subject := "oidc-subject-new"
	mockRepo := &MockUserRepository{
		GetByOIDCSubjectFunc: func(ctx context.Context, sub string) (*models.User, error) {
			return nil, &models.NotFoundError{Resource: "User", ID: sub}
		},
		GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, &models.NotFoundError{Resource: "User", ID: email}
		},
		IsFirstUserFunc: func(ctx context.Context) (bool, error) {
			return false, nil
		},
		CreateFunc: func(ctx context.Context, user *models.User) error {
			user.ID = 3
			return nil
		},
	}

	service := NewUserService(mockRepo)

	user, err := service.GetOrCreateUser(context.Background(), "new@example.com", "New User", &subject)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	if user.ID != 3 {
		t.Errorf("Expected user ID 3, got %d", user.ID)
	}

	if user.Email != "new@example.com" {
		t.Errorf("Expected email new@example.com, got %s", user.Email)
	}

	if user.Role != models.RoleEventManager {
		t.Errorf("Expected role event_manager, got %s", user.Role)
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			if id == 1 {
				return &models.User{
					ID:    1,
					Email: "user@example.com",
					Name:  "Test User",
					Role:  models.RoleAdmin,
				}, nil
			}
			return nil, &models.NotFoundError{Resource: "User", ID: id}
		},
	}

	service := NewUserService(mockRepo)

	user, err := service.GetUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return nil, &models.NotFoundError{Resource: "User", ID: id}
		},
	}

	service := NewUserService(mockRepo)

	_, err := service.GetUserByID(context.Background(), 999)
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

func TestUserService_UpdateUserRole(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{
				ID:    id,
				Email: "user@example.com",
				Name:  "Test User",
				Role:  models.RoleEventManager,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, user *models.User) error {
			if user.Role != models.RoleAdmin {
				return fmt.Errorf("unexpected role: %s", user.Role)
			}
			return nil
		},
	}

	service := NewUserService(mockRepo)

	err := service.UpdateUserRole(context.Background(), 1, models.RoleAdmin)
	if err != nil {
		t.Fatalf("UpdateUserRole failed: %v", err)
	}
}

func TestUserService_CreateUser_InvalidEmail(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)

	_, err := service.CreateUser(context.Background(), "not-an-email", "Test User", nil)
	if err == nil {
		t.Fatal("Expected error for invalid email, got nil")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestUserService_CreateUser_EmptyEmail(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)

	_, err := service.CreateUser(context.Background(), "", "Test User", nil)
	if err == nil {
		t.Fatal("Expected error for empty email, got nil")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestUserService_CreateUser_EmptyName(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)

	_, err := service.CreateUser(context.Background(), "test@example.com", "", nil)
	if err == nil {
		t.Fatal("Expected error for empty name, got nil")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestUserService_GetUserByEmail_NotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, &models.NotFoundError{Resource: "User", ID: email}
		},
	}

	service := NewUserService(mockRepo)

	_, err := service.GetUserByEmail(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}

	var notFoundErr *models.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestUserService_ListUsers(t *testing.T) {
	mockRepo := &MockUserRepository{
		ListFunc: func(ctx context.Context, limit, offset int) ([]*models.User, error) {
			users := []*models.User{
				{ID: 1, Email: "user1@example.com", Name: "User 1", Role: models.RoleAdmin},
				{ID: 2, Email: "user2@example.com", Name: "User 2", Role: models.RoleEventManager},
			}
			if offset >= len(users) {
				return []*models.User{}, nil
			}
			end := offset + limit
			if end > len(users) {
				end = len(users)
			}
			return users[offset:end], nil
		},
	}

	service := NewUserService(mockRepo)

	users, err := service.ListUsers(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	usersPage2, err := service.ListUsers(context.Background(), 10, 2)
	if err != nil {
		t.Fatalf("ListUsers page 2 failed: %v", err)
	}

	if len(usersPage2) != 0 {
		t.Errorf("Expected 0 users on page 2, got %d", len(usersPage2))
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	mockRepo := &MockUserRepository{
		DeleteFunc: func(ctx context.Context, id int64) error {
			if id == 1 {
				return nil
			}
			return &models.NotFoundError{Resource: "User", ID: id}
		},
	}

	service := NewUserService(mockRepo)

	err := service.DeleteUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	err = service.DeleteUser(context.Background(), 999)
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	mockRepo := &MockUserRepository{
		UpdateFunc: func(ctx context.Context, user *models.User) error {
			if user.ID == 1 {
				return nil
			}
			return &models.NotFoundError{Resource: "User", ID: user.ID}
		},
	}

	service := NewUserService(mockRepo)

	user := &models.User{
		ID:    1,
		Email: "updated@example.com",
		Name:  "Updated User",
		Role:  models.RoleAdmin,
	}

	err := service.UpdateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	nonExistentUser := &models.User{
		ID:    999,
		Email: "nonexistent@example.com",
		Name:  "Non-existent User",
		Role:  models.RoleAdmin,
	}

	err = service.UpdateUser(context.Background(), nonExistentUser)
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

type MockUserRepository struct {
	CreateFunc           func(ctx context.Context, user *models.User) error
	GetByIDFunc          func(ctx context.Context, id int64) (*models.User, error)
	GetByEmailFunc       func(ctx context.Context, email string) (*models.User, error)
	GetByOIDCSubjectFunc func(ctx context.Context, subject string) (*models.User, error)
	UpdateFunc           func(ctx context.Context, user *models.User) error
	DeleteFunc           func(ctx context.Context, id int64) error
	ListFunc             func(ctx context.Context, limit, offset int) ([]*models.User, error)
	CountFunc            func(ctx context.Context) (int, error)
	IsFirstUserFunc      func(ctx context.Context) (bool, error)
	UpdateLastLoginFunc  func(ctx context.Context, userID int64) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &models.User{ID: id}, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return &models.User{Email: email}, nil
}

func (m *MockUserRepository) GetByOIDCSubject(ctx context.Context, subject string) (*models.User, error) {
	if m.GetByOIDCSubjectFunc != nil {
		return m.GetByOIDCSubjectFunc(ctx, subject)
	}
	return &models.User{OIDCSubject: &subject}, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return []*models.User{}, nil
}

func (m *MockUserRepository) Count(ctx context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}

func (m *MockUserRepository) IsFirstUser(ctx context.Context) (bool, error) {
	if m.IsFirstUserFunc != nil {
		return m.IsFirstUserFunc(ctx)
	}
	return false, nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(ctx, userID)
	}
	return nil
}
