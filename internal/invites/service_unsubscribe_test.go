package invites

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestUnsubscribeFromReminders_Success(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)

	updateCalled := false
	mockRepo := &mockInviteRepository{
		getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
			if tokenHash != validHash {
				t.Errorf("Expected tokenHash %s, got %s", validHash, tokenHash)
			}
			return &models.Invite{
				ID:           1,
				EventID:      1,
				TokenHash:    tokenHash,
				Status:       models.InviteStatusSent,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
		updateFunc: func(ctx context.Context, invite *models.Invite) error {
			updateCalled = true
			if !invite.Unsubscribed {
				t.Error("Expected Unsubscribed to be true")
			}
			return nil
		},
	}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return validHash, nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, validToken)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !updateCalled {
		t.Error("Expected Update to be called")
	}
}

func TestUnsubscribeFromReminders_InvalidToken(t *testing.T) {
	ctx := context.Background()
	validHash := strings.Repeat("b", 43)

	mockRepo := &mockInviteRepository{
		getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
			return nil, &models.NotFoundError{
				Resource: "invite",
				ID:       "token",
			}
		},
	}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return validHash, nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}

	var notFoundErr *models.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestUnsubscribeFromReminders_ExpiredToken(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)

	mockRepo := &mockInviteRepository{
		getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
			return &models.Invite{
				ID:           1,
				EventID:      1,
				TokenHash:    tokenHash,
				Status:       models.InviteStatusSent,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(-24 * time.Hour),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return validHash, nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, validToken)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}

	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected error to contain 'expired', got %v", err)
	}
}

func TestUnsubscribeFromReminders_AlreadyUnsubscribed(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)

	updateCalled := false
	mockRepo := &mockInviteRepository{
		getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
			return &models.Invite{
				ID:           1,
				EventID:      1,
				TokenHash:    tokenHash,
				Status:       models.InviteStatusSent,
				Unsubscribed: true,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
		updateFunc: func(ctx context.Context, invite *models.Invite) error {
			updateCalled = true
			return nil
		},
	}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return validHash, nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, validToken)
	if err != nil {
		t.Errorf("Expected no error for already unsubscribed invite, got %v", err)
	}

	if updateCalled {
		t.Error("Expected Update not to be called for already unsubscribed invite")
	}
}

func TestUnsubscribeFromReminders_RevokedInvite(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)

	mockRepo := &mockInviteRepository{
		getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
			return &models.Invite{
				ID:           1,
				EventID:      1,
				TokenHash:    tokenHash,
				Status:       models.InviteStatusRevoked,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return validHash, nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, validToken)
	if err == nil {
		t.Error("Expected error for revoked invite, got nil")
	}

	if !strings.Contains(err.Error(), "revoked") {
		t.Errorf("Expected error to contain 'revoked', got %v", err)
	}
}

func TestUnsubscribeFromReminders_UpdateError(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)

	mockRepo := &mockInviteRepository{
		getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
			return &models.Invite{
				ID:           1,
				EventID:      1,
				TokenHash:    tokenHash,
				Status:       models.InviteStatusSent,
				Unsubscribed: false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
		updateFunc: func(ctx context.Context, invite *models.Invite) error {
			return errors.New("database error")
		},
	}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return validHash, nil
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, validToken)
	if err == nil {
		t.Error("Expected error from update failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to unsubscribe") {
		t.Errorf("Expected error to contain 'failed to unsubscribe', got %v", err)
	}
}

func TestUnsubscribeFromReminders_HashError(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)

	mockRepo := &mockInviteRepository{}

	mockGen := &mockGenerator{
		hashFunc: func(t string) (string, error) {
			return "", errors.New("hash failure")
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.UnsubscribeFromReminders(ctx, validToken)
	if err == nil {
		t.Error("Expected error from hash failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to hash token") {
		t.Errorf("Expected error to contain 'failed to hash token', got %v", err)
	}
}
