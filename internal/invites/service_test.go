package invites

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockGenerator struct {
	generateFunc func() (string, error)
	hashFunc     func(token string) (string, error)
}

func (m *mockGenerator) Generate() (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc()
	}
	return strings.Repeat("a", 43), nil
}

func (m *mockGenerator) Hash(t string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(t)
	}
	return strings.Repeat("b", 43), nil
}

type mockInviteRepository struct {
	createFunc              func(ctx context.Context, invite *models.Invite) error
	getByIDFunc             func(ctx context.Context, id int64) (*models.Invite, error)
	getByTokenHashFunc      func(ctx context.Context, tokenHash string) (*models.Invite, error)
	updateFunc              func(ctx context.Context, invite *models.Invite) error
	listByEventIDFunc       func(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error)
	findDuplicateEmailsFunc func(ctx context.Context, eventID int64, emails []string) ([]string, error)
}

func (m *mockInviteRepository) Create(ctx context.Context, invite *models.Invite) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, invite)
	}
	invite.ID = 1
	invite.CreatedAt = time.Now()
	invite.UpdatedAt = time.Now()
	return nil
}

func (m *mockInviteRepository) GetByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockInviteRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	if m.getByTokenHashFunc != nil {
		return m.getByTokenHashFunc(ctx, tokenHash)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockInviteRepository) Update(ctx context.Context, invite *models.Invite) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, invite)
	}
	invite.UpdatedAt = time.Now()
	return nil
}

func (m *mockInviteRepository) CreateBatch(ctx context.Context, invites []*models.Invite) error {
	return nil
}

func (m *mockInviteRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockInviteRepository) ListByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	if m.listByEventIDFunc != nil {
		return m.listByEventIDFunc(ctx, eventID, filters)
	}
	return []*models.Invite{}, nil
}

func (m *mockInviteRepository) CountByEventID(ctx context.Context, eventID int64) (int, error) {
	return 0, nil
}

func (m *mockInviteRepository) GetStats(ctx context.Context, eventID int64) (*repositories.InviteStats, error) {
	return &repositories.InviteStats{}, nil
}

func (m *mockInviteRepository) FindDuplicateEmails(ctx context.Context, eventID int64, emails []string) ([]string, error) {
	if m.findDuplicateEmailsFunc != nil {
		return m.findDuplicateEmailsFunc(ctx, eventID, emails)
	}
	return []string{}, nil
}

func (m *mockInviteRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

type mockInviteRepositoryWithDeleteExpired struct {
	mockInviteRepository
	deleteExpiredFunc func(ctx context.Context, before time.Time) (int64, error)
}

func (m *mockInviteRepositoryWithDeleteExpired) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	if m.deleteExpiredFunc != nil {
		return m.deleteExpiredFunc(ctx, before)
	}
	return 0, nil
}

func TestInviteService_CreateInvite(t *testing.T) {
	ctx := context.Background()
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email := "test@example.com"
	name := "Test User"

	tests := []struct {
		name           string
		eventID        int64
		inviteName     *string
		inviteEmail    *string
		maxPlusOnes    int
		expiresAt      time.Time
		mockGen        *mockGenerator
		mockRepo       *mockInviteRepository
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, invite *models.Invite, plainToken string)
	}{
		{
			name:        "successful invite creation with email",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: 2,
			expiresAt:   futureTime,
			mockGen:     &mockGenerator{},
			mockRepo:    &mockInviteRepository{},
			wantErr:     false,
			validateResult: func(t *testing.T, invite *models.Invite, plainToken string) {
				if invite == nil {
					t.Fatal("invite is nil")
				}
				if invite.EventID != 1 {
					t.Errorf("EventID = %d, want 1", invite.EventID)
				}
				if invite.Email == nil || *invite.Email != email {
					t.Errorf("Email = %v, want %s", invite.Email, email)
				}
				if invite.Name == nil || *invite.Name != name {
					t.Errorf("Name = %v, want %s", invite.Name, name)
				}
				if invite.MaxPlusOnes != 2 {
					t.Errorf("MaxPlusOnes = %d, want 2", invite.MaxPlusOnes)
				}
				if invite.Status != models.InviteStatusDraft {
					t.Errorf("Status = %s, want %s", invite.Status, models.InviteStatusDraft)
				}
				if invite.TokenHash == "" {
					t.Error("TokenHash is empty")
				}
				if len(invite.TokenHash) != 43 {
					t.Errorf("TokenHash length = %d, want 43", len(invite.TokenHash))
				}
				if plainToken == "" {
					t.Error("plainToken is empty")
				}
				if len(plainToken) != 43 {
					t.Errorf("plainToken length = %d, want 43", len(plainToken))
				}
			},
		},
		{
			name:        "successful invite creation without email",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: nil,
			maxPlusOnes: 0,
			expiresAt:   futureTime,
			mockGen:     &mockGenerator{},
			mockRepo:    &mockInviteRepository{},
			wantErr:     false,
			validateResult: func(t *testing.T, invite *models.Invite, plainToken string) {
				if invite == nil {
					t.Fatal("invite is nil")
				}
				if invite.Email != nil {
					t.Errorf("Email = %v, want nil", invite.Email)
				}
				if invite.Status != models.InviteStatusDraft {
					t.Errorf("Status = %s, want %s", invite.Status, models.InviteStatusDraft)
				}
			},
		},
		{
			name:        "token generation failure",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: 2,
			expiresAt:   futureTime,
			mockGen: &mockGenerator{
				generateFunc: func() (string, error) {
					return "", errors.New("entropy exhausted")
				},
			},
			mockRepo:    &mockInviteRepository{},
			wantErr:     true,
			errContains: "failed to generate token",
		},
		{
			name:        "token hashing failure",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: 2,
			expiresAt:   futureTime,
			mockGen: &mockGenerator{
				hashFunc: func(token string) (string, error) {
					return "", errors.New("hash failure")
				},
			},
			mockRepo:    &mockInviteRepository{},
			wantErr:     true,
			errContains: "failed to hash token",
		},
		{
			name:        "repository creation failure",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: 2,
			expiresAt:   futureTime,
			mockGen:     &mockGenerator{},
			mockRepo: &mockInviteRepository{
				createFunc: func(ctx context.Context, invite *models.Invite) error {
					return errors.New("database error")
				},
			},
			wantErr:     true,
			errContains: "failed to create invite",
		},
		{
			name:        "invalid event ID",
			eventID:     0,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: 2,
			expiresAt:   futureTime,
			mockGen:     &mockGenerator{},
			mockRepo:    &mockInviteRepository{},
			wantErr:     true,
			errContains: "event_id must be positive",
		},
		{
			name:        "invalid max plus ones",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: -1,
			expiresAt:   futureTime,
			mockGen:     &mockGenerator{},
			mockRepo:    &mockInviteRepository{},
			wantErr:     true,
			errContains: "max_plus_ones must be between 0 and 10",
		},
		{
			name:        "expired expiration time",
			eventID:     1,
			inviteName:  &name,
			inviteEmail: &email,
			maxPlusOnes: 2,
			expiresAt:   time.Now().Add(-24 * time.Hour),
			mockGen:     &mockGenerator{},
			mockRepo:    &mockInviteRepository{},
			wantErr:     true,
			errContains: "expires_at must be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			invite, plainToken, err := service.CreateInvite(ctx, tt.eventID, tt.inviteName, tt.inviteEmail, tt.maxPlusOnes, tt.expiresAt)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateInvite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CreateInvite() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if tt.validateResult != nil {
				tt.validateResult(t, invite, plainToken)
			}
		})
	}
}

func TestInviteService_GetInviteByToken(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)
	email := "test@example.com"

	tests := []struct {
		name           string
		token          string
		mockGen        *mockGenerator
		mockRepo       *mockInviteRepository
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, invite *models.Invite)
	}{
		{
			name:  "successful retrieval",
			token: validToken,
			mockGen: &mockGenerator{
				hashFunc: func(t string) (string, error) {
					return validHash, nil
				},
			},
			mockRepo: &mockInviteRepository{
				getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
					if tokenHash != validHash {
						t.Errorf("getByTokenHash called with %s, want %s", tokenHash, validHash)
					}
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   validHash,
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, invite *models.Invite) {
				if invite == nil {
					t.Fatal("invite is nil")
				}
				if invite.ID != 1 {
					t.Errorf("ID = %d, want 1", invite.ID)
				}
				if invite.TokenHash != validHash {
					t.Errorf("TokenHash = %s, want %s", invite.TokenHash, validHash)
				}
			},
		},
		{
			name:  "token hashing failure",
			token: validToken,
			mockGen: &mockGenerator{
				hashFunc: func(t string) (string, error) {
					return "", errors.New("hash failure")
				},
			},
			mockRepo:    &mockInviteRepository{},
			wantErr:     true,
			errContains: "failed to hash token",
		},
		{
			name:  "invite not found",
			token: validToken,
			mockGen: &mockGenerator{
				hashFunc: func(t string) (string, error) {
					return validHash, nil
				},
			},
			mockRepo: &mockInviteRepository{
				getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
					return nil, &models.NotFoundError{Resource: "invite"}
				},
			},
			wantErr:     true,
			errContains: "invite not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			invite, err := service.GetInviteByToken(ctx, tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInviteByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetInviteByToken() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if tt.validateResult != nil {
				tt.validateResult(t, invite)
			}
		})
	}
}

func TestInviteService_RevokeInvite(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	tests := []struct {
		name        string
		inviteID    int64
		mockGen     *mockGenerator
		mockRepo    *mockInviteRepository
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful revocation from draft",
			inviteID: 1,
			mockGen:  &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusDraft,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr: false,
		},
		{
			name:     "successful revocation from sent",
			inviteID: 1,
			mockGen:  &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr: false,
		},
		{
			name:     "invite not found",
			inviteID: 999,
			mockGen:  &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return nil, &models.NotFoundError{Resource: "invite"}
				},
			},
			wantErr:     true,
			errContains: "invite not found",
		},
		{
			name:     "cannot revoke already revoked",
			inviteID: 1,
			mockGen:  &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusRevoked,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:     true,
			errContains: "cannot transition from revoked",
		},
		{
			name:     "update failure",
			inviteID: 1,
			mockGen:  &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusDraft,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
				updateFunc: func(ctx context.Context, invite *models.Invite) error {
					return errors.New("database error")
				},
			},
			wantErr:     true,
			errContains: "failed to update invite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			err := service.RevokeInvite(ctx, tt.inviteID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeInvite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("RevokeInvite() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestInviteService_GetInviteByToken_ExpiredToken(t *testing.T) {
	ctx := context.Background()
	validToken := strings.Repeat("a", 43)
	validHash := strings.Repeat("b", 43)
	email := "test@example.com"

	tests := []struct {
		name        string
		token       string
		mockGen     *mockGenerator
		mockRepo    *mockInviteRepository
		wantErr     bool
		errContains string
	}{
		{
			name:  "expired token rejected",
			token: validToken,
			mockGen: &mockGenerator{
				hashFunc: func(t string) (string, error) {
					return validHash, nil
				},
			},
			mockRepo: &mockInviteRepository{
				getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   validHash,
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   time.Now().Add(-24 * time.Hour),
					}, nil
				},
			},
			wantErr:     true,
			errContains: "invite has expired",
		},
		{
			name:  "valid token not expired",
			token: validToken,
			mockGen: &mockGenerator{
				hashFunc: func(t string) (string, error) {
					return validHash, nil
				},
			},
			mockRepo: &mockInviteRepository{
				getByTokenHashFunc: func(ctx context.Context, tokenHash string) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   validHash,
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   time.Now().Add(24 * time.Hour),
					}, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			invite, err := service.GetInviteByToken(ctx, tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInviteByToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetInviteByToken() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if invite == nil {
				t.Fatal("invite is nil")
			}
		})
	}
}

func TestInviteService_CleanupExpiredTokens(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		mockGen        *mockGenerator
		mockRepo       *mockInviteRepositoryWithDeleteExpired
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, count int64)
	}{
		{
			name:    "successful cleanup with deleted tokens",
			mockGen: &mockGenerator{},
			mockRepo: &mockInviteRepositoryWithDeleteExpired{
				deleteExpiredFunc: func(ctx context.Context, before time.Time) (int64, error) {
					if before.After(time.Now()) {
						t.Error("before time should not be in the future")
					}
					return 5, nil
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, count int64) {
				if count != 5 {
					t.Errorf("CleanupExpiredTokens() count = %d, want 5", count)
				}
			},
		},
		{
			name:    "successful cleanup with no expired tokens",
			mockGen: &mockGenerator{},
			mockRepo: &mockInviteRepositoryWithDeleteExpired{
				deleteExpiredFunc: func(ctx context.Context, before time.Time) (int64, error) {
					return 0, nil
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, count int64) {
				if count != 0 {
					t.Errorf("CleanupExpiredTokens() count = %d, want 0", count)
				}
			},
		},
		{
			name:    "repository error",
			mockGen: &mockGenerator{},
			mockRepo: &mockInviteRepositoryWithDeleteExpired{
				deleteExpiredFunc: func(ctx context.Context, before time.Time) (int64, error) {
					return 0, errors.New("database error")
				},
			},
			wantErr:     true,
			errContains: "failed to cleanup expired tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			count, err := service.CleanupExpiredTokens(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("CleanupExpiredTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CleanupExpiredTokens() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if tt.validateResult != nil {
				tt.validateResult(t, count)
			}
		})
	}
}
