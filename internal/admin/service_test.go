package admin

import (
	"context"
	"errors"
	"testing"
)

type mockUserCounter struct {
	count int
	err   error
}

func (m *mockUserCounter) CountUsers(ctx context.Context) (int, error) {
	return m.count, m.err
}

type mockEventCounter struct {
	count int
	err   error
}

func (m *mockEventCounter) CountEvents(ctx context.Context) (int, error) {
	return m.count, m.err
}

type mockInviteCounter struct {
	count int
	err   error
}

func (m *mockInviteCounter) CountInvites(ctx context.Context) (int, error) {
	return m.count, m.err
}

func TestAdminService_GetStats_Success(t *testing.T) {
	userCounter := &mockUserCounter{count: 10}
	eventCounter := &mockEventCounter{count: 5}
	inviteCounter := &mockInviteCounter{count: 50}

	service := NewAdminService(userCounter, eventCounter, inviteCounter)

	stats, err := service.GetAdminStats(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats.TotalUsers != 10 {
		t.Errorf("Expected TotalUsers=10, got %d", stats.TotalUsers)
	}
	if stats.TotalEvents != 5 {
		t.Errorf("Expected TotalEvents=5, got %d", stats.TotalEvents)
	}
	if stats.TotalInvites != 50 {
		t.Errorf("Expected TotalInvites=50, got %d", stats.TotalInvites)
	}
}

func TestAdminService_GetStats_UserCountError(t *testing.T) {
	userCounter := &mockUserCounter{err: errors.New("user count error")}
	eventCounter := &mockEventCounter{count: 5}
	inviteCounter := &mockInviteCounter{count: 50}

	service := NewAdminService(userCounter, eventCounter, inviteCounter)

	_, err := service.GetAdminStats(context.Background())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAdminService_GetStats_EventCountError(t *testing.T) {
	userCounter := &mockUserCounter{count: 10}
	eventCounter := &mockEventCounter{err: errors.New("event count error")}
	inviteCounter := &mockInviteCounter{count: 50}

	service := NewAdminService(userCounter, eventCounter, inviteCounter)

	_, err := service.GetAdminStats(context.Background())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAdminService_GetStats_InviteCountError(t *testing.T) {
	userCounter := &mockUserCounter{count: 10}
	eventCounter := &mockEventCounter{count: 5}
	inviteCounter := &mockInviteCounter{err: errors.New("invite count error")}

	service := NewAdminService(userCounter, eventCounter, inviteCounter)

	_, err := service.GetAdminStats(context.Background())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAdminService_GetStats_ZeroCounts(t *testing.T) {
	userCounter := &mockUserCounter{count: 0}
	eventCounter := &mockEventCounter{count: 0}
	inviteCounter := &mockInviteCounter{count: 0}

	service := NewAdminService(userCounter, eventCounter, inviteCounter)

	stats, err := service.GetAdminStats(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats.TotalUsers != 0 {
		t.Errorf("Expected TotalUsers=0, got %d", stats.TotalUsers)
	}
	if stats.TotalEvents != 0 {
		t.Errorf("Expected TotalEvents=0, got %d", stats.TotalEvents)
	}
	if stats.TotalInvites != 0 {
		t.Errorf("Expected TotalInvites=0, got %d", stats.TotalInvites)
	}
}
