package invites

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

type mockUpdateInviteRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
	updateFunc  func(ctx context.Context, invite *models.Invite) error
}

func (m *mockUpdateInviteRepo) GetByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockUpdateInviteRepo) Update(ctx context.Context, invite *models.Invite) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, invite)
	}
	return nil
}

func (m *mockUpdateInviteRepo) Create(ctx context.Context, invite *models.Invite) error {
	return nil
}

func (m *mockUpdateInviteRepo) CreateBatch(ctx context.Context, invites []*models.Invite) error {
	return nil
}

func (m *mockUpdateInviteRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockUpdateInviteRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockUpdateInviteRepo) ListByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockUpdateInviteRepo) CountByEventID(ctx context.Context, eventID int64) (int, error) {
	return 0, nil
}

func (m *mockUpdateInviteRepo) CountByEventIDWithFilters(ctx context.Context, eventID int64, filters repositories.InviteFilters) (int, error) {
	return 0, nil
}

func (m *mockUpdateInviteRepo) GetStats(ctx context.Context, eventID int64) (*repositories.InviteStats, error) {
	return nil, nil
}

func (m *mockUpdateInviteRepo) GetByEventIDs(ctx context.Context, eventIDs []int64) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockUpdateInviteRepo) FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error) {
	return nil, nil
}

func (m *mockUpdateInviteRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *mockUpdateInviteRepo) CountInvites(ctx context.Context) (int, error) {
	return 0, nil
}

func TestUpdateInvite_Success(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockUpdateInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		updateFunc: func(ctx context.Context, inv *models.Invite) error {
			return nil
		},
	}

	service := &inviteService{
		repo: repo,
	}

	req := &UpdateInviteRequest{
		InviteID:    1,
		Name:        testutil.StringPtr("Updated Name"),
		MaxPlusOnes: testutil.IntPtr(3),
	}

	err := service.UpdateInvite(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUpdateInvite_NotFound(t *testing.T) {
	repo := &mockUpdateInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	service := &inviteService{
		repo: repo,
	}

	req := &UpdateInviteRequest{
		InviteID: 999,
		Name:     testutil.StringPtr("Updated Name"),
	}

	err := service.UpdateInvite(context.Background(), req)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateInvite_CannotUpdateResponded(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusResponded,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockUpdateInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	service := &inviteService{
		repo: repo,
	}

	req := &UpdateInviteRequest{
		InviteID: 1,
		Name:     testutil.StringPtr("Updated Name"),
	}

	err := service.UpdateInvite(context.Background(), req)
	if err == nil {
		t.Error("Expected error for responded invite, got nil")
	}
}

func TestUpdateInvite_CannotUpdateRevoked(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       testutil.StringPtr("test@example.com"),
		Name:        testutil.StringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusRevoked,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockUpdateInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	service := &inviteService{
		repo: repo,
	}

	req := &UpdateInviteRequest{
		InviteID: 1,
		Name:     testutil.StringPtr("Updated Name"),
	}

	err := service.UpdateInvite(context.Background(), req)
	if err == nil {
		t.Error("Expected error for revoked invite, got nil")
	}
}
