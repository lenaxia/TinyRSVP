package invites

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

type mockSendInviteRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
	updateFunc  func(ctx context.Context, invite *models.Invite) error
}

func (m *mockSendInviteRepo) GetByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockSendInviteRepo) Update(ctx context.Context, invite *models.Invite) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, invite)
	}
	return nil
}

func (m *mockSendInviteRepo) Create(ctx context.Context, invite *models.Invite) error {
	return nil
}

func (m *mockSendInviteRepo) CreateBatch(ctx context.Context, invites []*models.Invite) error {
	return nil
}

func (m *mockSendInviteRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockSendInviteRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockSendInviteRepo) ListByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) CountByEventID(ctx context.Context, eventID int64) (int, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) CountByEventIDWithFilters(ctx context.Context, eventID int64, filters repositories.InviteFilters) (int, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) GetStats(ctx context.Context, eventID int64) (*repositories.InviteStats, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *mockSendInviteRepo) GetByEventIDs(ctx context.Context, eventIDs []int64) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockSendInviteRepo) CountInvites(ctx context.Context) (int, error) {
	return 0, nil
}

type mockEmailQueueRepo struct {
	createFunc func(ctx context.Context, email *models.EmailQueue) error
}

func (m *mockEmailQueueRepo) Create(ctx context.Context, email *models.EmailQueue) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, email)
	}
	return nil
}

func (m *mockEmailQueueRepo) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockEmailQueueRepo) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	return nil
}

func (m *mockEmailQueueRepo) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockEmailQueueRepo) MarkSending(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) MarkSent(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockEmailQueueRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *mockEmailQueueRepo) MarkCancelled(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEmailQueueRepo) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	return nil
}

func (m *mockEmailQueueRepo) GetStats(ctx context.Context) (*repositories.EmailQueueStats, error) {
	return &repositories.EmailQueueStats{}, nil
}

func TestSendInvite_Success(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
		Name:        stringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
		updateFunc: func(ctx context.Context, inv *models.Invite) error {
			return nil
		},
	}

	emailRepo := &mockEmailQueueRepo{
		createFunc: func(ctx context.Context, email *models.EmailQueue) error {
			return nil
		},
	}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendInvite_NoEmail(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       nil,
		Name:        stringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusDraft,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	emailRepo := &mockEmailQueueRepo{}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err == nil {
		t.Error("Expected error for invite without email, got nil")
	}
}

func TestSendInvite_RevokedInvite(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	email := "test@example.com"

	invite := &models.Invite{
		ID:          1,
		EventID:     100,
		Email:       &email,
		Name:        stringPtr("Test User"),
		TokenHash:   "dGVzdF90b2tlbl9oYXNoXzEyMzQ1Njc4OTBhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5eg==",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusRevoked,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockSendInviteRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return invite, nil
		},
	}

	emailRepo := &mockEmailQueueRepo{}

	generator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	service := &inviteService{
		generator: generator,
		repo:      repo,
	}

	req := &SendInviteRequest{
		InviteID: 1,
		BaseURL:  "https://rsvp.example.com",
	}

	err := service.SendInvite(context.Background(), req, emailRepo)
	if err == nil {
		t.Error("Expected error for revoked invite, got nil")
	}
}
