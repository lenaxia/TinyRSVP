package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yourusername/tinyrsvp/internal/db/repositories"
	"github.com/yourusername/tinyrsvp/internal/models"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	user := &models.User{
		Email:       email,
		Name:        name,
		OIDCSubject: oidcSubject,
	}

	_, err := s.repo.CreateWithBootstrapCheck(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
	var user *models.User
	var err error

	if oidcSubject != nil {
		user, err = s.repo.GetByOIDCSubject(ctx, *oidcSubject)
		if err == nil {
			if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
				return nil, fmt.Errorf("failed to update last login: %w", err)
			}
			return user, nil
		}
	}

	user, err = s.repo.GetByEmail(ctx, email)
	if err == nil {
		if oidcSubject != nil && user.OIDCSubject == nil {
			user.OIDCSubject = oidcSubject
			if err := s.repo.Update(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
		if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("failed to update last login: %w", err)
		}
		return user, nil
	}

	return s.CreateUser(ctx, email, name, oidcSubject)
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	return s.repo.Update(ctx, user)
}

func (s *userService) UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Role = role
	return s.repo.Update(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return &models.ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	if !emailRegex.MatchString(email) {
		return &models.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	return nil
}

func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return &models.ValidationError{
			Field:   "name",
			Message: "name is required",
		}
	}

	return nil
}
