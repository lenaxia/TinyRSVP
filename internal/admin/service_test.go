package admin

import (
	"errors"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/testutil/mocks/other"
	"go.uber.org/mock/gomock"
)

func TestAdminService_GetStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCounter := other.NewMockUserCounter(ctrl)
	mockEventCounter := other.NewMockEventCounter(ctrl)
	mockInviteCounter := other.NewMockInviteCounter(ctrl)

	mockUserCounter.EXPECT().CountUsers(gomock.Any()).Return(10, nil)
	mockEventCounter.EXPECT().CountEvents(gomock.Any()).Return(5, nil)
	mockInviteCounter.EXPECT().CountInvites(gomock.Any()).Return(50, nil)

	service := NewAdminService(mockUserCounter, mockEventCounter, mockInviteCounter)

	stats, err := service.GetAdminStats(t.Context())
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCounter := other.NewMockUserCounter(ctrl)
	mockEventCounter := other.NewMockEventCounter(ctrl)
	mockInviteCounter := other.NewMockInviteCounter(ctrl)

	mockUserCounter.EXPECT().CountUsers(gomock.Any()).Return(0, errors.New("user count error"))

	service := NewAdminService(mockUserCounter, mockEventCounter, mockInviteCounter)

	_, err := service.GetAdminStats(t.Context())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAdminService_GetStats_EventCountError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCounter := other.NewMockUserCounter(ctrl)
	mockEventCounter := other.NewMockEventCounter(ctrl)
	mockInviteCounter := other.NewMockInviteCounter(ctrl)

	mockUserCounter.EXPECT().CountUsers(gomock.Any()).Return(10, nil)
	mockEventCounter.EXPECT().CountEvents(gomock.Any()).Return(0, errors.New("event count error"))

	service := NewAdminService(mockUserCounter, mockEventCounter, mockInviteCounter)

	_, err := service.GetAdminStats(t.Context())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAdminService_GetStats_InviteCountError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCounter := other.NewMockUserCounter(ctrl)
	mockEventCounter := other.NewMockEventCounter(ctrl)
	mockInviteCounter := other.NewMockInviteCounter(ctrl)

	mockUserCounter.EXPECT().CountUsers(gomock.Any()).Return(10, nil)
	mockEventCounter.EXPECT().CountEvents(gomock.Any()).Return(5, nil)
	mockInviteCounter.EXPECT().CountInvites(gomock.Any()).Return(0, errors.New("invite count error"))

	service := NewAdminService(mockUserCounter, mockEventCounter, mockInviteCounter)

	_, err := service.GetAdminStats(t.Context())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAdminService_GetStats_ZeroCounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCounter := other.NewMockUserCounter(ctrl)
	mockEventCounter := other.NewMockEventCounter(ctrl)
	mockInviteCounter := other.NewMockInviteCounter(ctrl)

	mockUserCounter.EXPECT().CountUsers(gomock.Any()).Return(0, nil)
	mockEventCounter.EXPECT().CountEvents(gomock.Any()).Return(0, nil)
	mockInviteCounter.EXPECT().CountInvites(gomock.Any()).Return(0, nil)

	service := NewAdminService(mockUserCounter, mockEventCounter, mockInviteCounter)

	stats, err := service.GetAdminStats(t.Context())
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
