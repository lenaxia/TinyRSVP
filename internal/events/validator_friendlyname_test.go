package events

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestValidator_ValidateFriendlyName(t *testing.T) {
	validator := NewValidator(NewTimezoneValidator())

	baseEvent := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 0,
	}

	tests := []struct {
		name         string
		friendlyName *string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "valid friendly name",
			friendlyName: stringPtr("summer-party-2026"),
			wantErr:      false,
		},
		{
			name:         "valid with numbers",
			friendlyName: stringPtr("event-123"),
			wantErr:      false,
		},
		{
			name:         "valid all lowercase",
			friendlyName: stringPtr("myevent"),
			wantErr:      false,
		},
		{
			name:         "nil friendly name is valid",
			friendlyName: nil,
			wantErr:      false,
		},
		{
			name:         "too short",
			friendlyName: stringPtr("ab"),
			wantErr:      true,
			errMessage:   "must be between 3 and 100 characters",
		},
		{
			name:         "too long",
			friendlyName: stringPtr("a" + string(make([]byte, 100))),
			wantErr:      true,
			errMessage:   "must be between 3 and 100 characters",
		},
		{
			name:         "contains uppercase",
			friendlyName: stringPtr("Summer-Party"),
			wantErr:      true,
			errMessage:   "must be lowercase",
		},
		{
			name:         "contains spaces",
			friendlyName: stringPtr("summer party"),
			wantErr:      true,
			errMessage:   "can only contain lowercase letters, numbers, and hyphens",
		},
		{
			name:         "contains special characters",
			friendlyName: stringPtr("summer_party"),
			wantErr:      true,
			errMessage:   "can only contain lowercase letters, numbers, and hyphens",
		},
		{
			name:         "starts with hyphen",
			friendlyName: stringPtr("-summer-party"),
			wantErr:      true,
			errMessage:   "cannot start or end with a hyphen",
		},
		{
			name:         "ends with hyphen",
			friendlyName: stringPtr("summer-party-"),
			wantErr:      true,
			errMessage:   "cannot start or end with a hyphen",
		},
		{
			name:         "consecutive hyphens",
			friendlyName: stringPtr("summer--party"),
			wantErr:      true,
			errMessage:   "cannot contain consecutive hyphens",
		},
		{
			name:         "leading whitespace",
			friendlyName: stringPtr(" summer-party"),
			wantErr:      true,
			errMessage:   "cannot have leading or trailing whitespace",
		},
		{
			name:         "trailing whitespace",
			friendlyName: stringPtr("summer-party "),
			wantErr:      true,
			errMessage:   "cannot have leading or trailing whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &models.Event{
				FriendlyName: tt.friendlyName,
				Title:        baseEvent.Title,
				StartTime:    baseEvent.StartTime,
				Timezone:     baseEvent.Timezone,
				MaxPlusOnes:  baseEvent.MaxPlusOnes,
			}

			err := validator.ValidateCreate(context.Background(), event)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				valErr, ok := err.(*models.ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}

				if valErr.Field != "friendly_name" {
					t.Errorf("Expected error field 'friendly_name', got '%s'", valErr.Field)
				}

				if tt.errMessage != "" && !contains(valErr.Message, tt.errMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errMessage, valErr.Message)
				}
			}
		})
	}
}

func TestValidator_ValidateFriendlyName_InUpdate(t *testing.T) {
	validator := NewValidator(NewTimezoneValidator())

	baseEvent := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		MaxPlusOnes: 0,
		Status:      models.EventStatusDraft,
	}

	tests := []struct {
		name         string
		friendlyName *string
		wantErr      bool
	}{
		{
			name:         "valid friendly name in update",
			friendlyName: stringPtr("updated-event-name"),
			wantErr:      false,
		},
		{
			name:         "invalid friendly name in update",
			friendlyName: stringPtr("Invalid-Name"),
			wantErr:      true,
		},
		{
			name:         "nil friendly name in update",
			friendlyName: nil,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &models.Event{
				FriendlyName: tt.friendlyName,
				Title:        baseEvent.Title,
				StartTime:    baseEvent.StartTime,
				Timezone:     baseEvent.Timezone,
				MaxPlusOnes:  baseEvent.MaxPlusOnes,
				Status:       baseEvent.Status,
			}

			err := validator.ValidateUpdate(context.Background(), event)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
