package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/repositories"
	"go.uber.org/mock/gomock"
)

func TestDashboardService_GetDashboardStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedStats := &models.DashboardStats{
		TotalEvents:     3,
		DraftEvents:     1,
		PublishedEvents: 2,
		TotalInvites:    4,
		PendingInvites:  2,
		TotalRSVPs:      3,
		AcceptedRSVPs:   2,
		DeclinedRSVPs:   1,
	}

	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
	mockRSVPRepo := repositories.NewMockRSVPRepository(ctrl)

	mockEventRepo.EXPECT().
		GetDashboardStatsByCreator(gomock.Any(), int64(1)).
		Return(expectedStats, nil)

	service := NewDashboardService(mockEventRepo, mockInviteRepo, mockRSVPRepo)

	user := &models.User{ID: 1, Role: models.RoleEventManager}
	ctx := auth.WithUser(context.Background(), user)

	stats, err := service.GetDashboardStats(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TotalEvents != 3 {
		t.Errorf("expected TotalEvents 3, got %d", stats.TotalEvents)
	}
	if stats.DraftEvents != 1 {
		t.Errorf("expected DraftEvents 1, got %d", stats.DraftEvents)
	}
	if stats.PublishedEvents != 2 {
		t.Errorf("expected PublishedEvents 2, got %d", stats.PublishedEvents)
	}
	if stats.TotalInvites != 4 {
		t.Errorf("expected TotalInvites 4, got %d", stats.TotalInvites)
	}
	if stats.PendingInvites != 2 {
		t.Errorf("expected PendingInvites 2, got %d", stats.PendingInvites)
	}
	if stats.TotalRSVPs != 3 {
		t.Errorf("expected TotalRSVPs 3, got %d", stats.TotalRSVPs)
	}
	if stats.AcceptedRSVPs != 2 {
		t.Errorf("expected AcceptedRSVPs 2, got %d", stats.AcceptedRSVPs)
	}
	if stats.DeclinedRSVPs != 1 {
		t.Errorf("expected DeclinedRSVPs 1, got %d", stats.DeclinedRSVPs)
	}
	if stats.ResponseRate != 75 {
		t.Errorf("expected ResponseRate 75, got %d", stats.ResponseRate)
	}
}

func TestDashboardService_GetDashboardStats_NoEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
	mockRSVPRepo := repositories.NewMockRSVPRepository(ctrl)

	// Set expectation
	mockEventRepo.EXPECT().
		GetDashboardStatsByCreator(gomock.Any(), int64(1)).
		Return(&models.DashboardStats{}, nil)

	service := NewDashboardService(mockEventRepo, mockInviteRepo, mockRSVPRepo)

	user := &models.User{ID: 1, Role: models.RoleEventManager}
	ctx := auth.WithUser(context.Background(), user)

	stats, err := service.GetDashboardStats(ctx, user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TotalEvents != 0 {
		t.Errorf("expected TotalEvents 0, got %d", stats.TotalEvents)
	}
	if stats.ResponseRate != 0 {
		t.Errorf("expected ResponseRate 0, got %d", stats.ResponseRate)
	}
}

func TestDashboardService_GetDashboardStats_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
	mockRSVPRepo := repositories.NewMockRSVPRepository(ctrl)

	service := NewDashboardService(mockEventRepo, mockInviteRepo, mockRSVPRepo)

	ctx := context.Background()

	_, err := service.GetDashboardStats(ctx, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var permErr *models.PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Errorf("expected PermissionDeniedError, got %T", err)
	}
}

func TestDashboardService_GetRecentActivity_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()
	email := "guest1@example.com"
	events := []*models.Event{
		{ID: 1, Title: "Event 1", Status: models.EventStatusPublished, CreatedBy: 1, CreatedAt: now.Add(-1 * time.Hour)},
		{ID: 2, Title: "Event 2", Status: models.EventStatusDraft, CreatedBy: 1, CreatedAt: now.Add(-2 * time.Hour)},
	}

	invites := []*models.Invite{
		{ID: 1, EventID: 1, Email: &email, Status: models.InviteStatusViewed, CreatedAt: now.Add(-30 * time.Minute)},
	}

	rsvps := []*models.RSVP{
		{ID: 1, InviteID: 1, Response: models.RSVPResponseYes, CreatedAt: now.Add(-15 * time.Minute)},
	}

	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
	mockRSVPRepo := repositories.NewMockRSVPRepository(ctrl)

	// Set expectations
	mockEventRepo.EXPECT().
		GetByCreatorID(gomock.Any(), int64(1)).
		Return(events, nil)

	mockInviteRepo.EXPECT().
		GetByEventIDs(gomock.Any(), []int64{1, 2}).
		Return(invites, nil)

	mockRSVPRepo.EXPECT().
		GetByInviteIDs(gomock.Any(), []int64{1}).
		Return(rsvps, nil)

	service := NewDashboardService(mockEventRepo, mockInviteRepo, mockRSVPRepo)

	user := &models.User{ID: 1, Role: models.RoleEventManager}
	ctx := auth.WithUser(context.Background(), user)

	activity, err := service.GetRecentActivity(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activity) == 0 {
		t.Fatal("expected activity items, got none")
	}

	if len(activity) > 10 {
		t.Errorf("expected max 10 activity items, got %d", len(activity))
	}

	for _, item := range activity {
		if item.EventID == nil {
			t.Errorf("activity item %q has nil EventID; all items should link to an event", item.Title)
		}
	}
}

func TestDashboardService_GetRecentActivity_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
	mockRSVPRepo := repositories.NewMockRSVPRepository(ctrl)

	service := NewDashboardService(mockEventRepo, mockInviteRepo, mockRSVPRepo)

	ctx := context.Background()

	_, err := service.GetRecentActivity(ctx, 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var permErr *models.PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Errorf("expected PermissionDeniedError, got %T", err)
	}
}

func TestDashboardService_GetRecentActivity_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositories.NewMockEventRepository(ctrl)
	mockInviteRepo := repositories.NewMockInviteRepository(ctrl)
	mockRSVPRepo := repositories.NewMockRSVPRepository(ctrl)

	// Set expectation
	mockEventRepo.EXPECT().
		GetByCreatorID(gomock.Any(), int64(1)).
		Return([]*models.Event{}, nil)

	service := NewDashboardService(mockEventRepo, mockInviteRepo, mockRSVPRepo)

	user := &models.User{ID: 1, Role: models.RoleEventManager}
	ctx := auth.WithUser(context.Background(), user)

	activity, err := service.GetRecentActivity(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activity) != 0 {
		t.Errorf("expected 0 activity items, got %d", len(activity))
	}
}
