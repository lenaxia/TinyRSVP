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

func TestInviteService_ListInvites_EmptyResults(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	mockRepo.stats[eventID] = &repositories.InviteStats{
		Total:     0,
		Draft:     0,
		Sent:      0,
		Viewed:    0,
		Responded: 0,
		Revoked:   0,
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

	if len(resp.Invites) != 0 {
		t.Errorf("expected 0 invites for event with no invites, got %d", len(resp.Invites))
	}

	if resp.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Total)
	}

	if resp.Stats == nil {
		t.Fatal("expected stats to be set")
	}

	if resp.Stats.Total != 0 {
		t.Errorf("expected stats total 0, got %d", resp.Stats.Total)
	}
}

func TestInviteService_ListInvites_SortBySentAt(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	now := time.Now()
	sentAt1 := now.Add(-2 * time.Hour)
	sentAt2 := now.Add(-1 * time.Hour)

	invite1 := &models.Invite{
		ID:          1,
		EventID:     eventID,
		Email:       stringPtr("user1@example.com"),
		Name:        stringPtr("User 1"),
		TokenHash:   "hash1",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusSent,
		SentAt:      &sentAt1,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	invite2 := &models.Invite{
		ID:          2,
		EventID:     eventID,
		Email:       stringPtr("user2@example.com"),
		Name:        stringPtr("User 2"),
		TokenHash:   "hash2",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusSent,
		SentAt:      &sentAt2,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.invites[invite1.ID] = invite1
	mockRepo.invites[invite2.ID] = invite2

	mockRepo.stats[eventID] = &repositories.InviteStats{
		Total: 2,
		Sent:  2,
	}

	sortBy := "sent_at"
	sortOrder := "desc"
	req := &ListInvitesRequest{
		EventID:   eventID,
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
		Limit:     50,
		Offset:    0,
	}

	resp, err := service.ListInvites(ctx, req)
	if err != nil {
		t.Fatalf("ListInvites() error = %v", err)
	}

	if len(resp.Invites) != 2 {
		t.Errorf("expected 2 invites, got %d", len(resp.Invites))
	}
}

func TestInviteService_ListInvites_SortByViewedAt(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	now := time.Now()
	viewedAt1 := now.Add(-2 * time.Hour)
	viewedAt2 := now.Add(-1 * time.Hour)

	invite1 := &models.Invite{
		ID:          1,
		EventID:     eventID,
		Email:       stringPtr("user1@example.com"),
		Name:        stringPtr("User 1"),
		TokenHash:   "hash1",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusViewed,
		ViewedAt:    &viewedAt1,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	invite2 := &models.Invite{
		ID:          2,
		EventID:     eventID,
		Email:       stringPtr("user2@example.com"),
		Name:        stringPtr("User 2"),
		TokenHash:   "hash2",
		MaxPlusOnes: 2,
		Status:      models.InviteStatusViewed,
		ViewedAt:    &viewedAt2,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.invites[invite1.ID] = invite1
	mockRepo.invites[invite2.ID] = invite2

	mockRepo.stats[eventID] = &repositories.InviteStats{
		Total:  2,
		Viewed: 2,
	}

	sortBy := "viewed_at"
	sortOrder := "asc"
	req := &ListInvitesRequest{
		EventID:   eventID,
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
		Limit:     50,
		Offset:    0,
	}

	resp, err := service.ListInvites(ctx, req)
	if err != nil {
		t.Fatalf("ListInvites() error = %v", err)
	}

	if len(resp.Invites) != 2 {
		t.Errorf("expected 2 invites, got %d", len(resp.Invites))
	}
}

func TestInviteService_ListInvites_StatisticsAccuracy(t *testing.T) {
	mockRepo := &mockInviteRepository{
		invites: make(map[int64]*models.Invite),
		stats:   make(map[int64]*repositories.InviteStats),
	}
	mockGen := &mockGenerator{}
	service := NewInviteService(mockGen, mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	statuses := []models.InviteStatus{
		models.InviteStatusDraft,
		models.InviteStatusDraft,
		models.InviteStatusSent,
		models.InviteStatusSent,
		models.InviteStatusSent,
		models.InviteStatusViewed,
		models.InviteStatusResponded,
		models.InviteStatusRevoked,
	}

	for i, status := range statuses {
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
		Total:     8,
		Draft:     2,
		Sent:      3,
		Viewed:    1,
		Responded: 1,
		Revoked:   1,
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

	if resp.Stats == nil {
		t.Fatal("expected stats to be set")
	}

	if resp.Stats.Total != 8 {
		t.Errorf("Stats.Total = %d, want 8", resp.Stats.Total)
	}
	if resp.Stats.Draft != 2 {
		t.Errorf("Stats.Draft = %d, want 2", resp.Stats.Draft)
	}
	if resp.Stats.Sent != 3 {
		t.Errorf("Stats.Sent = %d, want 3", resp.Stats.Sent)
	}
	if resp.Stats.Viewed != 1 {
		t.Errorf("Stats.Viewed = %d, want 1", resp.Stats.Viewed)
	}
	if resp.Stats.Responded != 1 {
		t.Errorf("Stats.Responded = %d, want 1", resp.Stats.Responded)
	}
	if resp.Stats.Revoked != 1 {
		t.Errorf("Stats.Revoked = %d, want 1", resp.Stats.Revoked)
	}
}
