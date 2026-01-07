package invites

import (
	"context"
	"fmt"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

type InviteService interface {
	CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error)
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
	RevokeInvite(ctx context.Context, id int64) error
	ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error)
}

type inviteService struct {
	generator token.Generator
	repo      repositories.InviteRepository
}

func NewInviteService(generator token.Generator, repo repositories.InviteRepository) InviteService {
	return &inviteService{
		generator: generator,
		repo:      repo,
	}
}

func (s *inviteService) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	plainToken, err := s.generator.Generate()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash, err := s.generator.Hash(plainToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash token: %w", err)
	}

	invite := &models.Invite{
		EventID:     eventID,
		Name:        name,
		Email:       email,
		TokenHash:   tokenHash,
		MaxPlusOnes: maxPlusOnes,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
	}

	if err := invite.Validate(); err != nil {
		return nil, "", err
	}

	if err := s.repo.Create(ctx, invite); err != nil {
		return nil, "", fmt.Errorf("failed to create invite: %w", err)
	}

	return invite, plainToken, nil
}

func (s *inviteService) GetInviteByToken(ctx context.Context, plainToken string) (*models.Invite, error) {
	tokenHash, err := s.generator.Hash(plainToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	invite, err := s.repo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}

	return invite, nil
}

func (s *inviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	invite, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}

	return invite, nil
}

func (s *inviteService) RevokeInvite(ctx context.Context, id int64) error {
	invite, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invite: %w", err)
	}

	if err := invite.CanTransitionTo(models.InviteStatusRevoked); err != nil {
		return err
	}

	invite.Status = models.InviteStatusRevoked
	invite.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, invite); err != nil {
		return fmt.Errorf("failed to update invite: %w", err)
	}

	return nil
}

func (s *inviteService) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	invites, err := s.repo.ListByEventID(ctx, eventID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list invites: %w", err)
	}

	return invites, nil
}
