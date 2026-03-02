package handlers

import (
	"context"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type FullMockInviteService struct {
	CreateInviteFunc             func(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error)
	CreateManualInviteFunc       func(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error)
	GetInviteByTokenFunc         func(ctx context.Context, token string) (*models.Invite, error)
	GetInviteByIDFunc            func(ctx context.Context, id int64) (*models.Invite, error)
	UpdateInviteFunc             func(ctx context.Context, req *invites.UpdateInviteRequest) error
	DeleteInviteFunc             func(ctx context.Context, inviteID int64) error
	SendInviteFunc               func(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error
	RevokeInviteFunc             func(ctx context.Context, req *invites.RevokeInviteRequest) error
	RegenerateTokenFunc          func(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error)
	ListInvitesFunc              func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error)
	ListInvitesByEventIDFunc     func(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error)
	ImportCSVFunc                func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error)
	CleanupExpiredTokensFunc     func(ctx context.Context) (int64, error)
	MarkInviteSentFunc           func(ctx context.Context, inviteID int64) error
	MarkInviteViewedFunc         func(ctx context.Context, inviteID int64) error
	MarkInviteRespondedFunc      func(ctx context.Context, inviteID int64) error
	UnsubscribeFromRemindersFunc func(ctx context.Context, token string) error
}

func (m *FullMockInviteService) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	if m.CreateInviteFunc != nil {
		return m.CreateInviteFunc(ctx, eventID, name, email, maxPlusOnes, expiresAt)
	}
	return nil, "", nil
}

func (m *FullMockInviteService) CreateManualInvite(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
	if m.CreateManualInviteFunc != nil {
		return m.CreateManualInviteFunc(ctx, req, expiresAt)
	}
	return nil, nil
}

func (m *FullMockInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	if m.GetInviteByTokenFunc != nil {
		return m.GetInviteByTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *FullMockInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.GetInviteByIDFunc != nil {
		return m.GetInviteByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *FullMockInviteService) UpdateInvite(ctx context.Context, req *invites.UpdateInviteRequest) error {
	if m.UpdateInviteFunc != nil {
		return m.UpdateInviteFunc(ctx, req)
	}
	return nil
}

func (m *FullMockInviteService) DeleteInvite(ctx context.Context, inviteID int64) error {
	if m.DeleteInviteFunc != nil {
		return m.DeleteInviteFunc(ctx, inviteID)
	}
	return nil
}

func (m *FullMockInviteService) SendInvite(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
	if m.SendInviteFunc != nil {
		return m.SendInviteFunc(ctx, req, emailRepo)
	}
	return nil
}

func (m *FullMockInviteService) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	if m.RevokeInviteFunc != nil {
		return m.RevokeInviteFunc(ctx, req)
	}
	return nil
}

func (m *FullMockInviteService) RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
	if m.RegenerateTokenFunc != nil {
		return m.RegenerateTokenFunc(ctx, inviteID)
	}
	return nil, nil
}

func (m *FullMockInviteService) ListInvites(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
	if m.ListInvitesFunc != nil {
		return m.ListInvitesFunc(ctx, req)
	}
	return &invites.ListInvitesResponse{
		Invites: []*models.Invite{},
		Total:   0,
		Stats:   &repositories.InviteStats{},
	}, nil
}

func (m *FullMockInviteService) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	if m.ListInvitesByEventIDFunc != nil {
		return m.ListInvitesByEventIDFunc(ctx, eventID, filters)
	}
	return nil, nil
}

func (m *FullMockInviteService) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
	if m.ImportCSVFunc != nil {
		return m.ImportCSVFunc(ctx, eventID, csvData, defaultMaxPlusOnes, expiresAt)
	}
	return nil, nil
}

func (m *FullMockInviteService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	if m.CleanupExpiredTokensFunc != nil {
		return m.CleanupExpiredTokensFunc(ctx)
	}
	return 0, nil
}

func (m *FullMockInviteService) MarkInviteSent(ctx context.Context, inviteID int64) error {
	if m.MarkInviteSentFunc != nil {
		return m.MarkInviteSentFunc(ctx, inviteID)
	}
	return nil
}

func (m *FullMockInviteService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	if m.MarkInviteViewedFunc != nil {
		return m.MarkInviteViewedFunc(ctx, inviteID)
	}
	return nil
}

func (m *FullMockInviteService) MarkInviteResponded(ctx context.Context, inviteID int64) error {
	if m.MarkInviteRespondedFunc != nil {
		return m.MarkInviteRespondedFunc(ctx, inviteID)
	}
	return nil
}

func (m *FullMockInviteService) UnsubscribeFromReminders(ctx context.Context, token string) error {
	if m.UnsubscribeFromRemindersFunc != nil {
		return m.UnsubscribeFromRemindersFunc(ctx, token)
	}
	return nil
}
