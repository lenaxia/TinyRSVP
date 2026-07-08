package models

import (
	"errors"
	"testing"
)

func TestComponentType_String(t *testing.T) {
	tests := []struct {
		name string
		ct   ComponentType
		want string
	}{
		{"TextBox", ComponentTypeTextBox, "TextBox"},
		{"Image", ComponentTypeImage, "Image"},
		{"Background", ComponentTypeBackground, "Background"},
		{"Overlay", ComponentTypeOverlay, "Overlay"},
		{"Container", ComponentTypeContainer, "Container"},
		{"Divider", ComponentTypeDivider, "Divider"},
		{"custom value", ComponentType("Custom"), "Custom"},
		{"empty", ComponentType(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.String(); got != tt.want {
				t.Errorf("ComponentType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionMode_String(t *testing.T) {
	tests := []struct {
		name string
		pm   PositionMode
		want string
	}{
		{"absolute", PositionModeAbsolute, "absolute"},
		{"relative", PositionModeRelative, "relative"},
		{"flex", PositionModeFlex, "flex"},
		{"custom value", PositionMode("custom"), "custom"},
		{"empty", PositionMode(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pm.String(); got != tt.want {
				t.Errorf("PositionMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLayoutMode_String(t *testing.T) {
	tests := []struct {
		name string
		lm   LayoutMode
		want string
	}{
		{"flexbox", LayoutModeFlexbox, "flexbox"},
		{"grid", LayoutModeGrid, "grid"},
		{"absolute", LayoutModeAbsolute, "absolute"},
		{"custom value", LayoutMode("custom"), "custom"},
		{"empty", LayoutMode(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lm.String(); got != tt.want {
				t.Errorf("LayoutMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionConflictError(t *testing.T) {
	tests := []struct {
		name     string
		err      *VersionConflictError
		wantMsg  string
		wantType bool
	}{
		{
			name: "event version conflict",
			err: &VersionConflictError{
				ResourceType: "Event",
				ResourceID:   42,
				Expected:     3,
				Actual:       5,
			},
			wantMsg:  "Event 42 version conflict: expected 3, got 5",
			wantType: true,
		},
		{
			name: "template version conflict",
			err: &VersionConflictError{
				ResourceType: "Template",
				ResourceID:   7,
				Expected:     1,
				Actual:       2,
			},
			wantMsg:  "Template 7 version conflict: expected 1, got 2",
			wantType: true,
		},
		{
			name: "version conflict with matching versions",
			err: &VersionConflictError{
				ResourceType: "Invite",
				ResourceID:   100,
				Expected:     2,
				Actual:       2,
			},
			wantMsg:  "Invite 100 version conflict: expected 2, got 2",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("VersionConflictError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *VersionConflictError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestPermissionDeniedError(t *testing.T) {
	tests := []struct {
		name     string
		err      *PermissionDeniedError
		wantMsg  string
		wantType bool
	}{
		{
			name: "with string ID",
			err: &PermissionDeniedError{
				Action:   "delete",
				Resource: "event",
				ID:       "abc-123",
			},
			wantMsg:  "permission denied: cannot delete event abc-123",
			wantType: true,
		},
		{
			name: "with int ID",
			err: &PermissionDeniedError{
				Action:   "update",
				Resource: "user",
				ID:       42,
			},
			wantMsg:  "permission denied: cannot update user 42",
			wantType: true,
		},
		{
			name: "without ID (nil)",
			err: &PermissionDeniedError{
				Action:   "create",
				Resource: "event",
				ID:       nil,
			},
			wantMsg:  "permission denied: cannot create event",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("PermissionDeniedError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *PermissionDeniedError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      *UnauthorizedError
		wantMsg  string
		wantType bool
	}{
		{
			name: "with message",
			err: &UnauthorizedError{
				Message: "invalid credentials",
			},
			wantMsg:  "invalid credentials",
			wantType: true,
		},
		{
			name: "with custom message",
			err: &UnauthorizedError{
				Message: "session expired, please log in again",
			},
			wantMsg:  "session expired, please log in again",
			wantType: true,
		},
		{
			name:     "empty message defaults to unauthorized",
			err:      &UnauthorizedError{},
			wantMsg:  "unauthorized",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("UnauthorizedError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *UnauthorizedError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestForbiddenError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ForbiddenError
		wantMsg  string
		wantType bool
	}{
		{
			name: "with message",
			err: &ForbiddenError{
				Message: "insufficient permissions to perform this action",
			},
			wantMsg:  "insufficient permissions to perform this action",
			wantType: true,
		},
		{
			name: "with short message",
			err: &ForbiddenError{
				Message: "admin only",
			},
			wantMsg:  "admin only",
			wantType: true,
		},
		{
			name:     "empty message defaults to forbidden",
			err:      &ForbiddenError{},
			wantMsg:  "forbidden",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ForbiddenError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *ForbiddenError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestDashboardStats_CalculateResponseRate(t *testing.T) {
	tests := []struct {
		name             string
		stats            DashboardStats
		wantResponseRate int
	}{
		{
			name: "100% response rate",
			stats: DashboardStats{
				TotalInvites: 10,
				TotalRSVPs:   10,
			},
			wantResponseRate: 100,
		},
		{
			name: "50% response rate",
			stats: DashboardStats{
				TotalInvites: 100,
				TotalRSVPs:   50,
			},
			wantResponseRate: 50,
		},
		{
			name: "partial response rate rounds down",
			stats: DashboardStats{
				TotalInvites: 3,
				TotalRSVPs:   2,
			},
			wantResponseRate: 66,
		},
		{
			name: "zero invites yields zero rate",
			stats: DashboardStats{
				TotalInvites: 0,
				TotalRSVPs:   5,
			},
			wantResponseRate: 0,
		},
		{
			name: "zero invites and zero rsvps",
			stats: DashboardStats{
				TotalInvites: 0,
				TotalRSVPs:   0,
			},
			wantResponseRate: 0,
		},
		{
			name: "no rsvps yet",
			stats: DashboardStats{
				TotalInvites: 25,
				TotalRSVPs:   0,
			},
			wantResponseRate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.stats
			s.CalculateResponseRate()
			if s.ResponseRate != tt.wantResponseRate {
				t.Errorf("CalculateResponseRate() ResponseRate = %v, want %v", s.ResponseRate, tt.wantResponseRate)
			}
		})
	}
}

func TestPreferenceQuestion_Text(t *testing.T) {
	tests := []struct {
		name         string
		questionText string
		want         string
	}{
		{"non-empty text", "What is your meal preference?", "What is your meal preference?"},
		{"short text", "Diet?", "Diet?"},
		{"empty text", "", ""},
		{"text with special chars", "Are you bringing a +1? (yes/no)", "Are you bringing a +1? (yes/no)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &PreferenceQuestion{QuestionText: tt.questionText}
			if got := q.Text(); got != tt.want {
				t.Errorf("PreferenceQuestion.Text() = %v, want %v", got, tt.want)
			}
		})
	}
}
