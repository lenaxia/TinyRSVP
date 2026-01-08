package invites

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestMarkInviteSent(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	tests := []struct {
		name           string
		initialStatus  models.InviteStatus
		mockGen        *mockGenerator
		mockRepo       *mockInviteRepository
		wantErr        bool
		errContains    string
		checkSentAt    bool
		checkStatus    bool
		expectedStatus models.InviteStatus
	}{
		{
			name:          "draft to sent",
			initialStatus: models.InviteStatusDraft,
			mockGen:       &mockGenerator{},
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
					if invite.Status != models.InviteStatusSent {
						t.Errorf("Status = %s, want %s", invite.Status, models.InviteStatusSent)
					}
					if invite.SentAt == nil {
						t.Error("SentAt is nil, want timestamp")
					}
					return nil
				},
			},
			wantErr:        false,
			checkSentAt:    true,
			checkStatus:    true,
			expectedStatus: models.InviteStatusSent,
		},
		{
			name:          "sent to sent idempotent",
			initialStatus: models.InviteStatusSent,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					sentAt := time.Now().Add(-1 * time.Hour)
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						SentAt:      &sentAt,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:        false,
			checkSentAt:    true,
			checkStatus:    true,
			expectedStatus: models.InviteStatusSent,
		},
		{
			name:          "viewed to sent invalid",
			initialStatus: models.InviteStatusViewed,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusViewed,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:     true,
			errContains: "cannot transition from viewed to sent",
		},
		{
			name:          "responded to sent invalid",
			initialStatus: models.InviteStatusResponded,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusResponded,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:     true,
			errContains: "cannot transition from responded to sent",
		},
		{
			name:          "revoked to sent invalid",
			initialStatus: models.InviteStatusRevoked,
			mockGen:       &mockGenerator{},
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
			errContains: "cannot transition from revoked to sent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			err := service.MarkInviteSent(ctx, 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("MarkInviteSent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("MarkInviteSent() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestMarkInviteViewed(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	tests := []struct {
		name           string
		initialStatus  models.InviteStatus
		mockGen        *mockGenerator
		mockRepo       *mockInviteRepository
		wantErr        bool
		errContains    string
		checkViewedAt  bool
		checkStatus    bool
		expectedStatus models.InviteStatus
	}{
		{
			name:          "sent to viewed",
			initialStatus: models.InviteStatusSent,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					sentAt := time.Now().Add(-1 * time.Hour)
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						SentAt:      &sentAt,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
				updateFunc: func(ctx context.Context, invite *models.Invite) error {
					if invite.Status != models.InviteStatusViewed {
						t.Errorf("Status = %s, want %s", invite.Status, models.InviteStatusViewed)
					}
					if invite.ViewedAt == nil {
						t.Error("ViewedAt is nil, want timestamp")
					}
					return nil
				},
			},
			wantErr:        false,
			checkViewedAt:  true,
			checkStatus:    true,
			expectedStatus: models.InviteStatusViewed,
		},
		{
			name:          "viewed to viewed idempotent",
			initialStatus: models.InviteStatusViewed,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					sentAt := time.Now().Add(-2 * time.Hour)
					viewedAt := time.Now().Add(-1 * time.Hour)
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusViewed,
						SentAt:      &sentAt,
						ViewedAt:    &viewedAt,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:        false,
			checkViewedAt:  true,
			checkStatus:    true,
			expectedStatus: models.InviteStatusViewed,
		},
		{
			name:          "draft to viewed invalid",
			initialStatus: models.InviteStatusDraft,
			mockGen:       &mockGenerator{},
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
			wantErr:     true,
			errContains: "cannot transition from draft to viewed",
		},
		{
			name:          "responded to viewed invalid",
			initialStatus: models.InviteStatusResponded,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusResponded,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:     true,
			errContains: "cannot transition from responded to viewed",
		},
		{
			name:          "revoked to viewed invalid",
			initialStatus: models.InviteStatusRevoked,
			mockGen:       &mockGenerator{},
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
			errContains: "cannot transition from revoked to viewed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			err := service.MarkInviteViewed(ctx, 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("MarkInviteViewed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("MarkInviteViewed() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestMarkInviteResponded(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	tests := []struct {
		name           string
		initialStatus  models.InviteStatus
		mockGen        *mockGenerator
		mockRepo       *mockInviteRepository
		wantErr        bool
		errContains    string
		checkStatus    bool
		expectedStatus models.InviteStatus
	}{
		{
			name:          "viewed to responded",
			initialStatus: models.InviteStatusViewed,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					sentAt := time.Now().Add(-2 * time.Hour)
					viewedAt := time.Now().Add(-1 * time.Hour)
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusViewed,
						SentAt:      &sentAt,
						ViewedAt:    &viewedAt,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
				updateFunc: func(ctx context.Context, invite *models.Invite) error {
					if invite.Status != models.InviteStatusResponded {
						t.Errorf("Status = %s, want %s", invite.Status, models.InviteStatusResponded)
					}
					return nil
				},
			},
			wantErr:        false,
			checkStatus:    true,
			expectedStatus: models.InviteStatusResponded,
		},
		{
			name:          "responded to responded idempotent",
			initialStatus: models.InviteStatusResponded,
			mockGen:       &mockGenerator{},
			mockRepo: &mockInviteRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("b", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusResponded,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			wantErr:        false,
			checkStatus:    true,
			expectedStatus: models.InviteStatusResponded,
		},
		{
			name:          "draft to responded invalid",
			initialStatus: models.InviteStatusDraft,
			mockGen:       &mockGenerator{},
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
			wantErr:     true,
			errContains: "cannot transition from draft to responded",
		},
		{
			name:          "sent to responded invalid",
			initialStatus: models.InviteStatusSent,
			mockGen:       &mockGenerator{},
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
			wantErr:     true,
			errContains: "cannot transition from sent to responded",
		},
		{
			name:          "revoked to responded invalid",
			initialStatus: models.InviteStatusRevoked,
			mockGen:       &mockGenerator{},
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
			errContains: "cannot transition from revoked to responded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInviteService(tt.mockGen, tt.mockRepo)
			err := service.MarkInviteResponded(ctx, 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("MarkInviteResponded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("MarkInviteResponded() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestMarkInviteNotFound(t *testing.T) {
	ctx := context.Background()

	mockGen := &mockGenerator{}
	mockRepo := &mockInviteRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
			return nil, &models.NotFoundError{Resource: "invite"}
		},
	}

	service := NewInviteService(mockGen, mockRepo)

	err := service.MarkInviteSent(ctx, 99999)
	if err == nil {
		t.Error("MarkInviteSent() expected error for non-existent invite, got nil")
	}

	err = service.MarkInviteViewed(ctx, 99999)
	if err == nil {
		t.Error("MarkInviteViewed() expected error for non-existent invite, got nil")
	}

	err = service.MarkInviteResponded(ctx, 99999)
	if err == nil {
		t.Error("MarkInviteResponded() expected error for non-existent invite, got nil")
	}
}
