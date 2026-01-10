package admin

import (
	"context"
	"fmt"
)

type UserCounter interface {
	CountUsers(ctx context.Context) (int, error)
}

type EventCounter interface {
	CountEvents(ctx context.Context) (int, error)
}

type InviteCounter interface {
	CountInvites(ctx context.Context) (int, error)
}

type AdminStats struct {
	TotalUsers   int
	TotalEvents  int
	TotalInvites int
}

type AdminService struct {
	userCounter   UserCounter
	eventCounter  EventCounter
	inviteCounter InviteCounter
}

func NewAdminService(userCounter UserCounter, eventCounter EventCounter, inviteCounter InviteCounter) *AdminService {
	return &AdminService{
		userCounter:   userCounter,
		eventCounter:  eventCounter,
		inviteCounter: inviteCounter,
	}
}

func (s *AdminService) GetAdminStats(ctx context.Context) (*AdminStats, error) {
	userCount, err := s.userCounter.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	eventCount, err := s.eventCounter.CountEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	inviteCount, err := s.inviteCounter.CountInvites(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count invites: %w", err)
	}

	return &AdminStats{
		TotalUsers:   userCount,
		TotalEvents:  eventCount,
		TotalInvites: inviteCount,
	}, nil
}
