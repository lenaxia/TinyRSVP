package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockDashboardEventRepo struct {
	events []*models.Event
	err    error
}

func (m *mockDashboardEventRepo) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.events, nil
}

type mockDashboardInviteRepo struct {
	invites []*models.Invite
	err     error
}

func (m *mockDashboardInviteRepo) GetByEventIDs(ctx context.Context, eventIDs []int64) ([]*models.Invite, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.invites, nil
}

type mockDashboardRSVPRepo struct {
	rsvps []*models.RSVP
	err   error
}

func (m *mockDashboardRSVPRepo) GetByInviteIDs(ctx context.Context, inviteIDs []int64) ([]*models.RSVP, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rsvps, nil
}

func TestDashboardService_GetDashboardStats_Success(t *testing.T) {
	now := time.Now()
	events := []*models.Event{
		{ID: 1, Status: models.EventStatusDraft, CreatedBy: 1, CreatedAt: now},
		{ID: 2, Status: models.EventStatusPublished, CreatedBy: 1, CreatedAt: now},
		{ID: 3, Status: models.EventStatusPublished, CreatedBy: 1, CreatedAt: now},
	}

	invites := []*models.Invite{
		{ID: 1, EventID: 1, Status: models.InviteStatusDraft},
		{ID: 2, EventID: 2, Status: models.InviteStatusSent},
		{ID: 3, EventID: 2, Status: models.InviteStatusViewed},
		{ID: 4, EventID: 3, Status: models.InviteStatusViewed},
	}

	rsvps := []*models.RSVP{
		{ID: 1, InviteID: 2, Response: models.RSVPResponseYes},
		{ID: 2, InviteID: 3, Response: models.RSVPResponseYes},
		{ID: 3, InviteID: 4, Response: models.RSVPResponseNo},
	}

	eventRepo := &mockDashboardEventRepo{events: events}
	inviteRepo := &mockDashboardInviteRepo{invites: invites}
	rsvpRepo := &mockDashboardRSVPRepo{rsvps: rsvps}

	service := NewDashboardService(eventRepo, inviteRepo, rsvpRepo)

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
	eventRepo := &mockDashboardEventRepo{events: []*models.Event{}}
	inviteRepo := &mockDashboardInviteRepo{invites: []*models.Invite{}}
	rsvpRepo := &mockDashboardRSVPRepo{rsvps: []*models.RSVP{}}

	service := NewDashboardService(eventRepo, inviteRepo, rsvpRepo)

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
	eventRepo := &mockDashboardEventRepo{}
	inviteRepo := &mockDashboardInviteRepo{}
	rsvpRepo := &mockDashboardRSVPRepo{}

	service := NewDashboardService(eventRepo, inviteRepo, rsvpRepo)

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

	eventRepo := &mockDashboardEventRepo{events: events}
	inviteRepo := &mockDashboardInviteRepo{invites: invites}
	rsvpRepo := &mockDashboardRSVPRepo{rsvps: rsvps}

	service := NewDashboardService(eventRepo, inviteRepo, rsvpRepo)

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
}

func TestDashboardService_GetRecentActivity_NoUser(t *testing.T) {
	eventRepo := &mockDashboardEventRepo{}
	inviteRepo := &mockDashboardInviteRepo{}
	rsvpRepo := &mockDashboardRSVPRepo{}

	service := NewDashboardService(eventRepo, inviteRepo, rsvpRepo)

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
	eventRepo := &mockDashboardEventRepo{events: []*models.Event{}}
	inviteRepo := &mockDashboardInviteRepo{invites: []*models.Invite{}}
	rsvpRepo := &mockDashboardRSVPRepo{rsvps: []*models.RSVP{}}

	service := NewDashboardService(eventRepo, inviteRepo, rsvpRepo)

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
