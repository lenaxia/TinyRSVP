package models

import (
	"strings"
	"testing"
	"time"
)

func TestInvite_Validate(t *testing.T) {
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	pastTime := time.Now().Add(-24 * time.Hour)
	validTokenHash := strings.Repeat("a", 43)
	email := "test@example.com"
	name := "Test User"

	tests := []struct {
		name    string
		invite  *Invite
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid draft invite",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid sent invite with email",
			invite: &Invite{
				EventID:     1,
				Email:       &email,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusSent,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid invite with name and email",
			invite: &Invite{
				EventID:     1,
				Name:        &name,
				Email:       &email,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 0,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "invalid - missing event_id",
			invite: &Invite{
				EventID:     0,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "event_id must be positive",
		},
		{
			name: "invalid - negative event_id",
			invite: &Invite{
				EventID:     -1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "event_id must be positive",
		},
		{
			name: "invalid - missing token_hash",
			invite: &Invite{
				EventID:     1,
				TokenHash:   "",
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "token_hash is required",
		},
		{
			name: "invalid - token_hash too short",
			invite: &Invite{
				EventID:     1,
				TokenHash:   "short",
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "token_hash must be 43 characters",
		},
		{
			name: "invalid - token_hash too long",
			invite: &Invite{
				EventID:     1,
				TokenHash:   strings.Repeat("a", 45),
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "token_hash must be 43 characters",
		},
		{
			name: "invalid - negative max_plus_ones",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: -1,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "max_plus_ones must be between 0 and 10",
		},
		{
			name: "invalid - max_plus_ones too high",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 11,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "max_plus_ones must be between 0 and 10",
		},
		{
			name: "valid - max_plus_ones at boundary (0)",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 0,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid - max_plus_ones at boundary (10)",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 10,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "invalid - unknown status",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      "invalid",
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid - missing expires_at",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   time.Time{},
			},
			wantErr: true,
			errMsg:  "expires_at is required",
		},
		{
			name: "invalid - expires_at in past",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   pastTime,
			},
			wantErr: true,
			errMsg:  "expires_at must be in the future",
		},
		{
			name: "invalid - sent status without email",
			invite: &Invite{
				EventID:     1,
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusSent,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "email is required for sent invites",
		},
		{
			name: "invalid - name too long",
			invite: &Invite{
				EventID:     1,
				Name:        stringPtr(strings.Repeat("a", 101)),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "name must not exceed 100 characters",
		},
		{
			name: "valid - name at boundary (100 chars)",
			invite: &Invite{
				EventID:     1,
				Name:        stringPtr(strings.Repeat("a", 100)),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "invalid - email too long",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr(strings.Repeat("a", 256)),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "email must not exceed 255 characters",
		},
		{
			name: "valid - email at boundary (255 chars)",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr(strings.Repeat("a", 240) + "@example.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "invalid - email format (no @)",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("notanemail"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "invalid - email format (no domain)",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("user@"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "invalid - email format (no local part)",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("@example.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "invalid - email format (spaces)",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("user name@example.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "valid - email with subdomain",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("user@mail.example.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid - email with plus addressing",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("user+tag@example.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid - email with dots",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("first.last@example.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid - email with numbers",
			invite: &Invite{
				EventID:     1,
				Email:       stringPtr("user123@example456.com"),
				TokenHash:   validTokenHash,
				MaxPlusOnes: 2,
				Status:      InviteStatusDraft,
				ExpiresAt:   futureTime,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.invite.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestInvite_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  InviteStatus
		toStatus    InviteStatus
		wantErr     bool
		errContains string
	}{
		{
			name:       "draft to sent - valid",
			fromStatus: InviteStatusDraft,
			toStatus:   InviteStatusSent,
			wantErr:    false,
		},
		{
			name:       "draft to revoked - valid",
			fromStatus: InviteStatusDraft,
			toStatus:   InviteStatusRevoked,
			wantErr:    false,
		},
		{
			name:        "draft to viewed - invalid",
			fromStatus:  InviteStatusDraft,
			toStatus:    InviteStatusViewed,
			wantErr:     true,
			errContains: "cannot transition from draft to viewed",
		},
		{
			name:        "draft to responded - invalid",
			fromStatus:  InviteStatusDraft,
			toStatus:    InviteStatusResponded,
			wantErr:     true,
			errContains: "cannot transition from draft to responded",
		},
		{
			name:       "sent to viewed - valid",
			fromStatus: InviteStatusSent,
			toStatus:   InviteStatusViewed,
			wantErr:    false,
		},
		{
			name:       "sent to revoked - valid",
			fromStatus: InviteStatusSent,
			toStatus:   InviteStatusRevoked,
			wantErr:    false,
		},
		{
			name:        "sent to draft - invalid",
			fromStatus:  InviteStatusSent,
			toStatus:    InviteStatusDraft,
			wantErr:     true,
			errContains: "cannot transition from sent to draft",
		},
		{
			name:        "sent to responded - invalid",
			fromStatus:  InviteStatusSent,
			toStatus:    InviteStatusResponded,
			wantErr:     true,
			errContains: "cannot transition from sent to responded",
		},
		{
			name:       "viewed to responded - valid",
			fromStatus: InviteStatusViewed,
			toStatus:   InviteStatusResponded,
			wantErr:    false,
		},
		{
			name:       "viewed to revoked - valid",
			fromStatus: InviteStatusViewed,
			toStatus:   InviteStatusRevoked,
			wantErr:    false,
		},
		{
			name:        "viewed to draft - invalid",
			fromStatus:  InviteStatusViewed,
			toStatus:    InviteStatusDraft,
			wantErr:     true,
			errContains: "cannot transition from viewed to draft",
		},
		{
			name:        "viewed to sent - invalid",
			fromStatus:  InviteStatusViewed,
			toStatus:    InviteStatusSent,
			wantErr:     true,
			errContains: "cannot transition from viewed to sent",
		},
		{
			name:        "responded to any - invalid (terminal)",
			fromStatus:  InviteStatusResponded,
			toStatus:    InviteStatusDraft,
			wantErr:     true,
			errContains: "cannot transition from responded",
		},
		{
			name:        "responded to revoked - invalid (terminal)",
			fromStatus:  InviteStatusResponded,
			toStatus:    InviteStatusRevoked,
			wantErr:     true,
			errContains: "cannot transition from responded",
		},
		{
			name:        "revoked to any - invalid (terminal)",
			fromStatus:  InviteStatusRevoked,
			toStatus:    InviteStatusDraft,
			wantErr:     true,
			errContains: "cannot transition from revoked",
		},
		{
			name:        "revoked to sent - invalid (terminal)",
			fromStatus:  InviteStatusRevoked,
			toStatus:    InviteStatusSent,
			wantErr:     true,
			errContains: "cannot transition from revoked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invite := &Invite{Status: tt.fromStatus}
			err := invite.CanTransitionTo(tt.toStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanTransitionTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CanTransitionTo() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestIsValidInviteStatus(t *testing.T) {
	tests := []struct {
		name   string
		status InviteStatus
		want   bool
	}{
		{"draft is valid", InviteStatusDraft, true},
		{"sent is valid", InviteStatusSent, true},
		{"viewed is valid", InviteStatusViewed, true},
		{"responded is valid", InviteStatusResponded, true},
		{"revoked is valid", InviteStatusRevoked, true},
		{"empty is invalid", InviteStatus(""), false},
		{"unknown is invalid", InviteStatus("unknown"), false},
		{"random is invalid", InviteStatus("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidInviteStatus(tt.status); got != tt.want {
				t.Errorf("isValidInviteStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
