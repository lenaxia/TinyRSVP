package invites

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestInviteService_ListInvites(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	for i := 0; i < 5; i++ {
		status := models.InviteStatusDraft
		if i == 3 {
			status = models.InviteStatusSent
		} else if i == 4 {
			status = models.InviteStatusViewed
		}
		
		invite := &models.Invite{
			ID:          int64(i + 1),
			EventID:     eventID,
			Email:       stringPtr("user" + string(rune('0'+i)) + "@example.com"),
			Name:        stringPtr("User " + string(rune('A'+i))),
			TokenHash:   "hash" + string(rune('0'+i)),
			MaxPlusOnes: 2,
			Status:      status,
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockRepo.invites[invite.ID] = invite
	}

	mockRepo.stats[eventID] = &repositories.InviteStats{
		Total:     5,
		Draft:     3,
		Sent:      1,
		Viewed:    1,
		Responded: 0,
		Revoked:   0,
	}

	tests := []struct {
		name          string
		req           *ListInvitesRequest
		expectedCount int
		wantErr       bool
	}{
		{
			name: "list all invites",
			req: &ListInvitesRequest{
				EventID: eventID,
				Limit:   50,
				Offset:  0,
			},
			expectedCount: 5,
			wantErr:       false,
		},
		{
			name: "list with pagination",
			req: &ListInvitesRequest{
				EventID: eventID,
				Limit:   2,
				Offset:  0,
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "list with status filter",
			req: &ListInvitesRequest{
				EventID: eventID,
				Status:  stringPtr("draft"),
				Limit:   50,
				Offset:  0,
			},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name: "list with search",
			req: &ListInvitesRequest{
				EventID: eventID,
				Search:  stringPtr("User A"),
				Limit:   50,
				Offset:  0,
			},
			expectedCount: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.ListInvites(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListInvites() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp == nil {
					t.Fatal("expected response, got nil")
				}

				if len(resp.Invites) != tt.expectedCount {
					t.Errorf("expected %d invites, got %d", tt.expectedCount, len(resp.Invites))
				}

				if resp.Total == 0 {
					t.Error("expected total count to be set")
				}

				if resp.Stats == nil {
					t.Error("expected stats to be set")
				} else {
					if resp.Stats.Total != 5 {
						t.Errorf("expected stats total 5, got %d", resp.Stats.Total)
					}
				}
			}
		})
	}
}

func TestInviteService_ListInvites_InvalidStatus(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()

	req := &ListInvitesRequest{
		EventID: 1,
		Status:  stringPtr("invalid_status"),
		Limit:   50,
		Offset:  0,
	}

	_, err := service.ListInvites(ctx, req)
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestInviteService_ListInvites_InvalidSortBy(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()

	req := &ListInvitesRequest{
		EventID: 1,
		SortBy:  stringPtr("invalid_field"),
		Limit:   50,
		Offset:  0,
	}

	_, err := service.ListInvites(ctx, req)
	if err == nil {
		t.Error("expected error for invalid sort field, got nil")
	}
}

func TestInviteService_ListInvites_InvalidSortOrder(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()

	req := &ListInvitesRequest{
		EventID:   1,
		SortOrder: stringPtr("invalid_order"),
		Limit:     50,
		Offset:    0,
	}

	_, err := service.ListInvites(ctx, req)
	if err == nil {
		t.Error("expected error for invalid sort order, got nil")
	}
}

func TestInviteService_ListInvites_InvalidLimit(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()

	tests := []struct {
		name  string
		limit int
	}{
		{"negative limit", -1},
		{"zero limit", 0},
		{"limit too large", 101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ListInvitesRequest{
				EventID: 1,
				Limit:   tt.limit,
				Offset:  0,
			}

			_, err := service.ListInvites(ctx, req)
			if err == nil {
				t.Error("expected error for invalid limit, got nil")
			}
		})
	}
}

func TestInviteService_ListInvites_InvalidOffset(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()

	req := &ListInvitesRequest{
		EventID: 1,
		Limit:   50,
		Offset:  -1,
	}

	_, err := service.ListInvites(ctx, req)
	if err == nil {
		t.Error("expected error for negative offset, got nil")
	}
}

func TestInviteService_ListInvites_DefaultValues(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	for i := 0; i < 3; i++ {
		invite := &models.Invite{
			ID:          int64(i + 1),
			EventID:     eventID,
			Email:       stringPtr("user" + string(rune('0'+i)) + "@example.com"),
			Name:        stringPtr("User " + string(rune('A'+i))),
			TokenHash:   "hash" + string(rune('0'+i)),
			MaxPlusOnes: 2,
			Status:      models.InviteStatusDraft,
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockRepo.invites[invite.ID] = invite
	}

	mockRepo.stats[eventID] = &repositories.InviteStats{
		Total: 3,
		Draft: 3,
	}

	req := &ListInvitesRequest{
		EventID: eventID,
		Limit:   50,
		Offset:  0,
	}

	resp, err := service.ListInvites(ctx, req)
	if err != nil {
		t.Fatalf("ListInvites() error = %v", err)
	}

	if len(resp.Invites) != 3 {
		t.Errorf("expected 3 invites with default limit, got %d", len(resp.Invites))
	}
}

