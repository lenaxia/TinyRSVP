package events

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}

func TestEventValidator_ValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		event   *models.Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: &models.Event{
				Title:       "Birthday Party",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 2,
			},
			wantErr: false,
		},
		{
			name: "valid event with all fields",
			event: &models.Event{
				Title:        "Conference",
				Description:  stringPtr("Annual tech conference"),
				StartTime:    time.Now().Add(48 * time.Hour),
				EndTime:      timePtr(time.Now().Add(50 * time.Hour)),
				Timezone:     "America/New_York",
				Location:     stringPtr("Convention Center"),
				MaxPlusOnes:  5,
				RSVPDeadline: timePtr(time.Now().Add(24 * time.Hour)),
			},
			wantErr: false,
		},
		{
			name: "title too short",
			event: &models.Event{
				Title:     "AB",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "title must be between 3 and 200 characters",
		},
		{
			name: "title too long",
			event: &models.Event{
				Title:     strings.Repeat("A", 201),
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "title must be between 3 and 200 characters",
		},
		{
			name: "missing title",
			event: &models.Event{
				Title:     "",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "title must be between 3 and 200 characters",
		},
		{
			name: "title with leading whitespace",
			event: &models.Event{
				Title:     "  Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "title cannot have leading or trailing whitespace",
		},
		{
			name: "title with trailing whitespace",
			event: &models.Event{
				Title:     "Event  ",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "title cannot have leading or trailing whitespace",
		},
		{
			name: "description too long",
			event: &models.Event{
				Title:       "Event",
				Description: stringPtr(strings.Repeat("A", 5001)),
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "description cannot exceed 5000 characters",
		},
		{
			name: "start time in past",
			event: &models.Event{
				Title:     "Past Event",
				StartTime: time.Now().Add(-24 * time.Hour),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "start time must be in the future",
		},
		{
			name: "start time very close to now",
			event: &models.Event{
				Title:     "Soon Event",
				StartTime: time.Now().Add(30 * time.Second),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: false,
		},
		{
			name: "invalid timezone",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "Invalid/Zone",
			},
			wantErr: true,
			errMsg:  "invalid timezone",
		},
		{
			name: "missing timezone",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "",
			},
			wantErr: true,
			errMsg:  "timezone is required",
		},
		{
			name: "end time before start time",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				EndTime:   timePtr(time.Now().Add(12 * time.Hour)),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "end time must be after start time",
		},
		{
			name: "end time same as start time",
			event: func() *models.Event {
				startTime := time.Now().Add(24 * time.Hour)
				return &models.Event{
					Title:     "Event",
					StartTime: startTime,
					EndTime:   timePtr(startTime),
					Timezone:  "America/Los_Angeles",
				}
			}(),
			wantErr: true,
			errMsg:  "end time must be after start time",
		},
		{
			name: "end time more than 7 days after start",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				EndTime:   timePtr(time.Now().Add(24*time.Hour + 8*24*time.Hour)),
				Timezone:  "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "end time must be within 7 days of start time",
		},
		{
			name: "end time exactly 7 days after start",
			event: func() *models.Event {
				startTime := time.Now().Add(24 * time.Hour)
				return &models.Event{
					Title:     "Event",
					StartTime: startTime,
					EndTime:   timePtr(startTime.Add(7 * 24 * time.Hour)),
					Timezone:  "America/Los_Angeles",
				}
			}(),
			wantErr: false,
		},
		{
			name: "RSVP deadline after start time",
			event: &models.Event{
				Title:        "Event",
				StartTime:    time.Now().Add(24 * time.Hour),
				RSVPDeadline: timePtr(time.Now().Add(48 * time.Hour)),
				Timezone:     "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "RSVP deadline must be before event start time",
		},
		{
			name: "RSVP deadline same as start time",
			event: &models.Event{
				Title:        "Event",
				StartTime:    time.Now().Add(24 * time.Hour),
				RSVPDeadline: timePtr(time.Now().Add(24 * time.Hour)),
				Timezone:     "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "RSVP deadline must be before event start time",
		},
		{
			name: "RSVP deadline in past",
			event: &models.Event{
				Title:        "Event",
				StartTime:    time.Now().Add(24 * time.Hour),
				RSVPDeadline: timePtr(time.Now().Add(-1 * time.Hour)),
				Timezone:     "America/Los_Angeles",
			},
			wantErr: true,
			errMsg:  "RSVP deadline must be in the future",
		},
		{
			name: "max plus ones negative",
			event: &models.Event{
				Title:       "Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: -1,
			},
			wantErr: true,
			errMsg:  "max plus ones must be between 0 and 10",
		},
		{
			name: "max plus ones over limit",
			event: &models.Event{
				Title:       "Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 11,
			},
			wantErr: true,
			errMsg:  "max plus ones must be between 0 and 10",
		},
		{
			name: "max plus ones at upper limit",
			event: &models.Event{
				Title:       "Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 10,
			},
			wantErr: false,
		},
		{
			name: "max plus ones zero",
			event: &models.Event{
				Title:       "Event",
				StartTime:   time.Now().Add(24 * time.Hour),
				Timezone:    "America/Los_Angeles",
				MaxPlusOnes: 0,
			},
			wantErr: false,
		},
		{
			name: "location too long",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Location:  stringPtr(strings.Repeat("A", 501)),
			},
			wantErr: true,
			errMsg:  "location cannot exceed 500 characters",
		},
		{
			name: "location at limit",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Location:  stringPtr(strings.Repeat("A", 500)),
			},
			wantErr: false,
		},
	}

	validator := NewValidator(NewTimezoneValidator())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreate(context.Background(), tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("ValidateCreate() error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestEventValidator_ValidateStateTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    models.EventStatus
		to      models.EventStatus
		wantErr bool
		errMsg  string
	}{
		{
			name:    "draft to published",
			from:    models.EventStatusDraft,
			to:      models.EventStatusPublished,
			wantErr: false,
		},
		{
			name:    "draft to cancelled",
			from:    models.EventStatusDraft,
			to:      models.EventStatusCancelled,
			wantErr: false,
		},
		{
			name:    "published to cancelled",
			from:    models.EventStatusPublished,
			to:      models.EventStatusCancelled,
			wantErr: false,
		},
		{
			name:    "published to archived",
			from:    models.EventStatusPublished,
			to:      models.EventStatusArchived,
			wantErr: false,
		},
		{
			name:    "cancelled to archived",
			from:    models.EventStatusCancelled,
			to:      models.EventStatusArchived,
			wantErr: false,
		},
		{
			name:    "archived to published",
			from:    models.EventStatusArchived,
			to:      models.EventStatusPublished,
			wantErr: true,
			errMsg:  "cannot transition from archived",
		},
		{
			name:    "archived to cancelled",
			from:    models.EventStatusArchived,
			to:      models.EventStatusCancelled,
			wantErr: true,
			errMsg:  "cannot transition from archived",
		},
		{
			name:    "archived to draft",
			from:    models.EventStatusArchived,
			to:      models.EventStatusDraft,
			wantErr: true,
			errMsg:  "cannot transition from archived",
		},
		{
			name:    "published to draft",
			from:    models.EventStatusPublished,
			to:      models.EventStatusDraft,
			wantErr: true,
			errMsg:  "cannot revert to draft",
		},
		{
			name:    "cancelled to published",
			from:    models.EventStatusCancelled,
			to:      models.EventStatusPublished,
			wantErr: true,
			errMsg:  "cannot transition from cancelled to published",
		},
		{
			name:    "cancelled to draft",
			from:    models.EventStatusCancelled,
			to:      models.EventStatusDraft,
			wantErr: true,
			errMsg:  "cannot revert to draft",
		},
		{
			name:    "same state transition",
			from:    models.EventStatusPublished,
			to:      models.EventStatusPublished,
			wantErr: true,
			errMsg:  "event is already in",
		},
	}

	validator := NewValidator(NewTimezoneValidator())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStateTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStateTransition(%v, %v) error = %v, wantErr %v",
					tt.from, tt.to, err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("ValidateStateTransition() error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestEventValidator_ValidateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		event   *models.Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid draft event update",
			event: &models.Event{
				Title:       "Updated Event",
				StartTime:   time.Now().Add(48 * time.Hour),
				Timezone:    "America/Los_Angeles",
				Status:      models.EventStatusDraft,
				MaxPlusOnes: 3,
			},
			wantErr: false,
		},
		{
			name: "valid published event update non-date fields",
			event: &models.Event{
				Title:       "Updated Title",
				Description: stringPtr("Updated description"),
				StartTime:   time.Now().Add(48 * time.Hour),
				Timezone:    "America/Los_Angeles",
				Status:      models.EventStatusPublished,
				Location:    stringPtr("New Location"),
				MaxPlusOnes: 5,
			},
			wantErr: false,
		},
		{
			name: "update cancelled event",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusCancelled,
			},
			wantErr: true,
			errMsg:  "cannot update cancelled event",
		},
		{
			name: "update archived event",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusArchived,
			},
			wantErr: true,
			errMsg:  "cannot update archived event",
		},
		{
			name: "invalid title in update",
			event: &models.Event{
				Title:     "AB",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "America/Los_Angeles",
				Status:    models.EventStatusDraft,
			},
			wantErr: true,
			errMsg:  "title must be between 3 and 200 characters",
		},
		{
			name: "invalid timezone in update",
			event: &models.Event{
				Title:     "Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "Invalid/Zone",
				Status:    models.EventStatusDraft,
			},
			wantErr: true,
			errMsg:  "invalid timezone",
		},
	}

	validator := NewValidator(NewTimezoneValidator())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdate(context.Background(), tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("ValidateUpdate() error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
