package invites

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type IndividualInviteService interface {
	CreateIndividualInvite(ctx context.Context, user *models.User, req *CreateIndividualInviteRequest) (*CreateIndividualInviteResponse, error)
}

type CreateIndividualInviteRequest struct {
	EventID     int64
	Name        *string
	Email       string
	MaxPlusOnes *int
}

type CreateIndividualInviteResponse struct {
	Invite *models.Invite
	Token  string
}

type individualInviteService struct {
	generator  token.Generator
	inviteRepo repositories.InviteRepository
	eventRepo  repositories.EventRepository
}

func NewIndividualInviteService(
	generator token.Generator,
	inviteRepo repositories.InviteRepository,
	eventRepo repositories.EventRepository,
) IndividualInviteService {
	return &individualInviteService{
		generator:  generator,
		inviteRepo: inviteRepo,
		eventRepo:  eventRepo,
	}
}

func (s *individualInviteService) CreateIndividualInvite(
	ctx context.Context,
	user *models.User,
	req *CreateIndividualInviteRequest,
) (*CreateIndividualInviteResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	event, err := s.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		return nil, err
	}

	if err := s.validateEvent(event); err != nil {
		return nil, err
	}

	if err := s.checkPermission(user, event); err != nil {
		return nil, err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	duplicates, err := s.inviteRepo.FindDuplicateEmails(ctx, req.EventID, []string{normalizedEmail})
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate emails: %w", err)
	}
	if len(duplicates) > 0 {
		return nil, &models.ConflictError{
			Resource: "Invite",
			Field:    "email",
			Value:    req.Email,
		}
	}

	maxPlusOnes := event.MaxPlusOnes
	if req.MaxPlusOnes != nil {
		if *req.MaxPlusOnes > event.MaxPlusOnes {
			return nil, &models.ValidationError{
				Field:   "max_plus_ones",
				Message: fmt.Sprintf("max_plus_ones (%d) exceeds event limit (%d)", *req.MaxPlusOnes, event.MaxPlusOnes),
			}
		}
		maxPlusOnes = *req.MaxPlusOnes
	}

	plainToken, err := s.generator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash, err := s.generator.Hash(plainToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	expiresAt := event.StartTime.Add(30 * 24 * time.Hour)

	emailPtr := &normalizedEmail
	invite := &models.Invite{
		EventID:     req.EventID,
		Name:        req.Name,
		Email:       emailPtr,
		Token:       &plainToken,
		TokenHash:   tokenHash,
		MaxPlusOnes: maxPlusOnes,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	return &CreateIndividualInviteResponse{
		Invite: invite,
		Token:  plainToken,
	}, nil
}

func (s *individualInviteService) validateRequest(req *CreateIndividualInviteRequest) error {
	if req.Email == "" {
		return &models.ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	email := strings.TrimSpace(req.Email)
	if len(email) > 255 {
		return &models.ValidationError{
			Field:   "email",
			Message: "email must not exceed 255 characters",
		}
	}

	if !emailRegex.MatchString(email) {
		return &models.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	if req.Name != nil && len(*req.Name) > 100 {
		return &models.ValidationError{
			Field:   "name",
			Message: "name must not exceed 100 characters",
		}
	}

	if req.MaxPlusOnes != nil {
		if *req.MaxPlusOnes < 0 || *req.MaxPlusOnes > 10 {
			return &models.ValidationError{
				Field:   "max_plus_ones",
				Message: "max_plus_ones must be between 0 and 10",
			}
		}
	}

	return nil
}

func (s *individualInviteService) validateEvent(event *models.Event) error {
	if event.Status == models.EventStatusCancelled {
		return &models.ValidationError{
			Field:   "event",
			Message: "cannot create invite for cancelled event",
		}
	}

	if event.Status == models.EventStatusArchived {
		return &models.ValidationError{
			Field:   "event",
			Message: "cannot create invite for archived event",
		}
	}

	return nil
}

func (s *individualInviteService) checkPermission(user *models.User, event *models.Event) error {
	if user.IsAdmin() {
		return nil
	}

	if event.CreatedBy == user.ID {
		return nil
	}

	return &models.PermissionDeniedError{
		Action:   "create invite for",
		Resource: "Event",
		ID:       event.ID,
	}
}
