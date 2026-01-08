package invites

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

type ImportResult struct {
	Total      int
	Created    int
	Failed     int
	Duplicates int
	Errors     []ImportError
}

type ImportError struct {
	Row     int
	Email   string
	Message string
}

type CreateManualInviteRequest struct {
	EventID     int64
	Name        *string
	MaxPlusOnes *int
}

type CreateManualInviteResponse struct {
	Invite  *models.Invite
	Token   string
	RSVPURL string
}

type RevokeInviteRequest struct {
	InviteID int64
	Reason   *string
}

type InviteService interface {
	CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error)
	CreateManualInvite(ctx context.Context, req *CreateManualInviteRequest, expiresAt time.Time) (*CreateManualInviteResponse, error)
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
	RevokeInvite(ctx context.Context, req *RevokeInviteRequest) error
	ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error)
	ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*ImportResult, error)
	CleanupExpiredTokens(ctx context.Context) (int64, error)
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

func (s *inviteService) CreateManualInvite(ctx context.Context, req *CreateManualInviteRequest, expiresAt time.Time) (*CreateManualInviteResponse, error) {
	maxPlusOnes := 0
	if req.MaxPlusOnes != nil {
		maxPlusOnes = *req.MaxPlusOnes
	}

	invite, plainToken, err := s.CreateInvite(ctx, req.EventID, req.Name, nil, maxPlusOnes, expiresAt)
	if err != nil {
		return nil, err
	}

	rsvpURL := fmt.Sprintf("/rsvp/%s", plainToken)

	return &CreateManualInviteResponse{
		Invite:  invite,
		Token:   plainToken,
		RSVPURL: rsvpURL,
	}, nil
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

	if invite.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invite has expired")
	}

	if invite.Status == models.InviteStatusRevoked {
		return nil, fmt.Errorf("invite has been revoked")
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

func (s *inviteService) RevokeInvite(ctx context.Context, req *RevokeInviteRequest) error {
	invite, err := s.repo.GetByID(ctx, req.InviteID)
	if err != nil {
		return fmt.Errorf("failed to get invite: %w", err)
	}

	if err := invite.CanTransitionTo(models.InviteStatusRevoked); err != nil {
		return err
	}

	invite.Status = models.InviteStatusRevoked
	invite.RevocationReason = req.Reason
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

func (s *inviteService) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*ImportResult, error) {
	rows, err := parseCSV(csvData)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{
		Total:  len(rows),
		Errors: []ImportError{},
	}

	emailsSeen := make(map[string]int)
	var validRows []struct {
		row    CSVRow
		rowNum int
	}
	var allEmails []string

	for i, row := range rows {
		rowNum := i + 2
		allEmails = append(allEmails, row.Email)

		emailLower := strings.ToLower(row.Email)
		if prevRow, exists := emailsSeen[emailLower]; exists {
			result.Duplicates++
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum,
				Email:   row.Email,
				Message: fmt.Sprintf("duplicate email (first seen on row %d)", prevRow),
			})
			continue
		}
		emailsSeen[emailLower] = rowNum

		maxPlusOnes := defaultMaxPlusOnes
		if row.MaxPlusOnes != nil {
			maxPlusOnes = *row.MaxPlusOnes
		}

		var email *string
		if row.Email != "" {
			email = &row.Email
		}

		tempInvite := &models.Invite{
			EventID:     eventID,
			Name:        row.Name,
			Email:       email,
			TokenHash:   strings.Repeat("x", 43),
			MaxPlusOnes: maxPlusOnes,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   expiresAt,
		}

		if err := tempInvite.Validate(); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum,
				Email:   row.Email,
				Message: err.Error(),
			})
			continue
		}

		validRows = append(validRows, struct {
			row    CSVRow
			rowNum int
		}{row, rowNum})
	}

	if len(validRows) == 0 {
		return result, nil
	}

	dbDuplicates, err := s.repo.FindDuplicateEmails(ctx, eventID, allEmails)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate emails: %w", err)
	}

	dbDuplicateMap := make(map[string]bool)
	for _, email := range dbDuplicates {
		dbDuplicateMap[strings.ToLower(email)] = true
	}

	var finalInvites []*models.Invite
	for _, vr := range validRows {
		emailLower := strings.ToLower(vr.row.Email)
		if dbDuplicateMap[emailLower] {
			result.Duplicates++
			result.Errors = append(result.Errors, ImportError{
				Row:     vr.rowNum,
				Email:   vr.row.Email,
				Message: "email already invited to this event",
			})
			continue
		}

		plainToken, err := s.generator.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}

		tokenHash, err := s.generator.Hash(plainToken)
		if err != nil {
			return nil, fmt.Errorf("failed to hash token: %w", err)
		}

		maxPlusOnes := defaultMaxPlusOnes
		if vr.row.MaxPlusOnes != nil {
			maxPlusOnes = *vr.row.MaxPlusOnes
		}

		var email *string
		if vr.row.Email != "" {
			email = &vr.row.Email
		}

		invite := &models.Invite{
			EventID:     eventID,
			Name:        vr.row.Name,
			Email:       email,
			TokenHash:   tokenHash,
			MaxPlusOnes: maxPlusOnes,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   expiresAt,
		}

		finalInvites = append(finalInvites, invite)
	}

	if len(finalInvites) == 0 {
		return result, nil
	}

	if err := s.repo.CreateBatch(ctx, finalInvites); err != nil {
		return nil, fmt.Errorf("failed to create invites: %w", err)
	}

	result.Created = len(finalInvites)

	return result, nil
}

func (s *inviteService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	count, err := s.repo.DeleteExpired(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return count, nil
}
