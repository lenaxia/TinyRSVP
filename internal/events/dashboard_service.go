package events

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type DashboardEventRepository interface {
	GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error)
}

type DashboardInviteRepository interface {
	GetByEventIDs(ctx context.Context, eventIDs []int64) ([]*models.Invite, error)
}

type DashboardRSVPRepository interface {
	GetByInviteIDs(ctx context.Context, inviteIDs []int64) ([]*models.RSVP, error)
}

type DashboardStats struct {
	TotalEvents     int
	DraftEvents     int
	PublishedEvents int
	TotalInvites    int
	PendingInvites  int
	TotalRSVPs      int
	AcceptedRSVPs   int
	DeclinedRSVPs   int
	ResponseRate    int
}

func (s *DashboardStats) CalculateResponseRate() {
	if s.TotalInvites == 0 {
		s.ResponseRate = 0
		return
	}
	s.ResponseRate = (s.TotalRSVPs * 100) / s.TotalInvites
}

type ActivityItem struct {
	Icon        string
	Title       string
	Description string
	Time        string
	EventID     *int64
}

type DashboardService interface {
	GetDashboardStats(ctx context.Context, userID int64) (*DashboardStats, error)
	GetRecentActivity(ctx context.Context, userID int64, limit int) ([]*ActivityItem, error)
}

type dashboardService struct {
	eventRepo  DashboardEventRepository
	inviteRepo DashboardInviteRepository
	rsvpRepo   DashboardRSVPRepository
}

func NewDashboardService(
	eventRepo DashboardEventRepository,
	inviteRepo DashboardInviteRepository,
	rsvpRepo DashboardRSVPRepository,
) DashboardService {
	return &dashboardService{
		eventRepo:  eventRepo,
		inviteRepo: inviteRepo,
		rsvpRepo:   rsvpRepo,
	}
}

func FormatTimeAgo(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	return t.Format("Jan 2, 2006")
}

func (s *dashboardService) GetDashboardStats(ctx context.Context, userID int64) (*DashboardStats, error) {
	_, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "get dashboard stats",
			Resource: "Dashboard",
		}
	}

	events, err := s.eventRepo.GetByCreatorID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	stats := &DashboardStats{}

	stats.TotalEvents = len(events)
	for _, event := range events {
		switch event.Status {
		case models.EventStatusDraft:
			stats.DraftEvents++
		case models.EventStatusPublished:
			stats.PublishedEvents++
		}
	}

	if len(events) == 0 {
		stats.CalculateResponseRate()
		return stats, nil
	}

	eventIDs := make([]int64, len(events))
	for i, event := range events {
		eventIDs[i] = event.ID
	}

	invites, err := s.inviteRepo.GetByEventIDs(ctx, eventIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get invites: %w", err)
	}

	stats.TotalInvites = len(invites)
	for _, invite := range invites {
		if invite.Status == models.InviteStatusDraft || invite.Status == models.InviteStatusSent {
			stats.PendingInvites++
		}
	}

	if len(invites) == 0 {
		stats.CalculateResponseRate()
		return stats, nil
	}

	inviteIDs := make([]int64, len(invites))
	for i, invite := range invites {
		inviteIDs[i] = invite.ID
	}

	rsvps, err := s.rsvpRepo.GetByInviteIDs(ctx, inviteIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get rsvps: %w", err)
	}

	stats.TotalRSVPs = len(rsvps)
	for _, rsvp := range rsvps {
		switch rsvp.Response {
		case models.RSVPResponseYes:
			stats.AcceptedRSVPs++
		case models.RSVPResponseNo:
			stats.DeclinedRSVPs++
		}
	}

	stats.CalculateResponseRate()

	return stats, nil
}

type activityEvent struct {
	timestamp time.Time
	item      *ActivityItem
}

func (s *dashboardService) GetRecentActivity(ctx context.Context, userID int64, limit int) ([]*ActivityItem, error) {
	_, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "get recent activity",
			Resource: "Dashboard",
		}
	}

	events, err := s.eventRepo.GetByCreatorID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return []*ActivityItem{}, nil
	}

	eventIDs := make([]int64, len(events))
	eventMap := make(map[int64]*models.Event)
	for i, event := range events {
		eventIDs[i] = event.ID
		eventMap[event.ID] = event
	}

	invites, err := s.inviteRepo.GetByEventIDs(ctx, eventIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get invites: %w", err)
	}

	inviteMap := make(map[int64]*models.Invite)
	for _, invite := range invites {
		inviteMap[invite.ID] = invite
	}

	var inviteIDs []int64
	if len(invites) > 0 {
		inviteIDs = make([]int64, len(invites))
		for i, invite := range invites {
			inviteIDs[i] = invite.ID
		}
	}

	var rsvps []*models.RSVP
	if len(inviteIDs) > 0 {
		rsvps, err = s.rsvpRepo.GetByInviteIDs(ctx, inviteIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get rsvps: %w", err)
		}
	}

	var activityEvents []activityEvent

	for _, event := range events {
		eventID := event.ID
		activityEvents = append(activityEvents, activityEvent{
			timestamp: event.CreatedAt,
			item: &ActivityItem{
				Icon:        "📅",
				Title:       "Event Created",
				Description: fmt.Sprintf("%s created", event.Title),
				Time:        FormatTimeAgo(event.CreatedAt),
				EventID:     &eventID,
			},
		})
	}

	for _, invite := range invites {
		if event, ok := eventMap[invite.EventID]; ok {
			email := "unknown"
			if invite.Email != nil {
				email = *invite.Email
			}
			eventID := event.ID
			activityEvents = append(activityEvents, activityEvent{
				timestamp: invite.CreatedAt,
				item: &ActivityItem{
					Icon:        "✉️",
					Title:       "Invite Sent",
					Description: fmt.Sprintf("Invite sent to %s for %s", email, event.Title),
					Time:        FormatTimeAgo(invite.CreatedAt),
					EventID:     &eventID,
				},
			})
		}
	}

	for _, rsvp := range rsvps {
		if invite, ok := inviteMap[rsvp.InviteID]; ok {
			if event, ok := eventMap[invite.EventID]; ok {
				icon := "✅"
				response := "accepted"
				if rsvp.Response == models.RSVPResponseNo {
					icon = "❌"
					response = "declined"
				} else if rsvp.Response == models.RSVPResponseMaybe {
					icon = "❓"
					response = "maybe"
				}

				email := "unknown"
				if invite.Email != nil {
					email = *invite.Email
				}

			activityEvents = append(activityEvents, activityEvent{
				timestamp: rsvp.CreatedAt,
				item: &ActivityItem{
					Icon:        icon,
					Title:       "RSVP Received",
					Description: fmt.Sprintf("%s %s for %s", email, response, event.Title),
					Time:        FormatTimeAgo(rsvp.CreatedAt),
					EventID:     &event.ID,
				},
			})
			}
		}
	}

	sort.Slice(activityEvents, func(i, j int) bool {
		return activityEvents[i].timestamp.After(activityEvents[j].timestamp)
	})

	result := make([]*ActivityItem, 0, limit)
	for i := 0; i < len(activityEvents) && i < limit; i++ {
		result = append(result, activityEvents[i].item)
	}

	return result, nil
}
