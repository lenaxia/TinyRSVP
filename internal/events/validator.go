package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type Validator interface {
	ValidateCreate(ctx context.Context, event *models.Event) error
	ValidateUpdate(ctx context.Context, event *models.Event) error
	ValidateStateTransition(from, to models.EventStatus) error
}

type validator struct {
	tzValidator TimezoneValidator
}

func NewValidator(tzValidator TimezoneValidator) Validator {
	return &validator{
		tzValidator: tzValidator,
	}
}

func (v *validator) ValidateCreate(ctx context.Context, event *models.Event) error {
	if err := v.validateTitle(event.Title); err != nil {
		return err
	}

	if err := v.validateDescription(event.Description); err != nil {
		return err
	}

	if err := v.validateTimezone(event.Timezone); err != nil {
		return err
	}

	if err := v.validateStartTime(event.StartTime); err != nil {
		return err
	}

	if err := v.validateEndTime(event.StartTime, event.EndTime); err != nil {
		return err
	}

	if err := v.validateRSVPDeadline(event.StartTime, event.RSVPDeadline); err != nil {
		return err
	}

	if err := v.validateMaxPlusOnes(event.MaxPlusOnes); err != nil {
		return err
	}

	if err := v.validateLocation(event.Location); err != nil {
		return err
	}

	return nil
}

func (v *validator) ValidateUpdate(ctx context.Context, event *models.Event) error {
	if event.Status == models.EventStatusCancelled {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot update cancelled event",
		}
	}

	if event.Status == models.EventStatusArchived {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot update archived event",
		}
	}

	if event.Status == models.EventStatusCompleted {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot update completed event",
		}
	}

	if err := v.validateTitle(event.Title); err != nil {
		return err
	}

	if err := v.validateDescription(event.Description); err != nil {
		return err
	}

	if err := v.validateTimezone(event.Timezone); err != nil {
		return err
	}

	if err := v.validateEndTime(event.StartTime, event.EndTime); err != nil {
		return err
	}

	if err := v.validateRSVPDeadline(event.StartTime, event.RSVPDeadline); err != nil {
		return err
	}

	if err := v.validateMaxPlusOnes(event.MaxPlusOnes); err != nil {
		return err
	}

	if err := v.validateLocation(event.Location); err != nil {
		return err
	}

	return nil
}

func (v *validator) ValidateStateTransition(from, to models.EventStatus) error {
	if from == to {
		return &models.ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("event is already in %s status", to),
		}
	}

	if from == models.EventStatusArchived {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot transition from archived status",
		}
	}

	if to == models.EventStatusDraft {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot revert to draft status",
		}
	}

	validTransitions := map[models.EventStatus][]models.EventStatus{
		models.EventStatusDraft: {
			models.EventStatusPublished,
			models.EventStatusCancelled,
		},
		models.EventStatusPublished: {
			models.EventStatusCancelled,
			models.EventStatusArchived,
		},
		models.EventStatusCancelled: {
			models.EventStatusArchived,
		},
		models.EventStatusCompleted: {
			models.EventStatusArchived,
		},
	}

	allowedStates, exists := validTransitions[from]
	if !exists {
		return &models.ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("no valid transitions from %s status", from),
		}
	}

	for _, allowed := range allowedStates {
		if to == allowed {
			return nil
		}
	}

	return &models.ValidationError{
		Field:   "status",
		Message: fmt.Sprintf("cannot transition from %s to %s", from, to),
	}
}

func (v *validator) validateTitle(title string) error {
	if title != strings.TrimSpace(title) {
		return &models.ValidationError{
			Field:   "title",
			Message: "title cannot have leading or trailing whitespace",
		}
	}

	if len(title) < 3 || len(title) > 200 {
		return &models.ValidationError{
			Field:   "title",
			Message: "title must be between 3 and 200 characters",
		}
	}

	return nil
}

func (v *validator) validateDescription(description *string) error {
	if description != nil && len(*description) > 5000 {
		return &models.ValidationError{
			Field:   "description",
			Message: "description cannot exceed 5000 characters",
		}
	}
	return nil
}

func (v *validator) validateTimezone(timezone string) error {
	if timezone == "" {
		return &models.ValidationError{
			Field:   "timezone",
			Message: "timezone is required",
		}
	}

	if !v.tzValidator.IsValid(timezone) {
		return &models.ValidationError{
			Field:   "timezone",
			Message: "invalid timezone",
		}
	}

	return nil
}

func (v *validator) validateStartTime(startTime time.Time) error {
	if startTime.Before(time.Now()) {
		return &models.ValidationError{
			Field:   "start_time",
			Message: "start time must be in the future",
		}
	}
	return nil
}

func (v *validator) validateEndTime(startTime time.Time, endTime *time.Time) error {
	if endTime == nil {
		return nil
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return &models.ValidationError{
			Field:   "end_time",
			Message: "end time must be after start time",
		}
	}

	maxDuration := 7 * 24 * time.Hour
	duration := endTime.Sub(startTime)
	if duration > maxDuration {
		return &models.ValidationError{
			Field:   "end_time",
			Message: "end time must be within 7 days of start time",
		}
	}

	return nil
}

func (v *validator) validateRSVPDeadline(startTime time.Time, rsvpDeadline *time.Time) error {
	if rsvpDeadline == nil {
		return nil
	}

	if rsvpDeadline.Before(time.Now()) {
		return &models.ValidationError{
			Field:   "rsvp_deadline",
			Message: "RSVP deadline must be in the future",
		}
	}

	if !rsvpDeadline.Before(startTime) {
		return &models.ValidationError{
			Field:   "rsvp_deadline",
			Message: "RSVP deadline must be before event start time",
		}
	}

	return nil
}

func (v *validator) validateMaxPlusOnes(maxPlusOnes int) error {
	if maxPlusOnes < 0 || maxPlusOnes > 10 {
		return &models.ValidationError{
			Field:   "max_plus_ones",
			Message: fmt.Sprintf("max plus ones must be between 0 and 10"),
		}
	}
	return nil
}

func (v *validator) validateLocation(location *string) error {
	if location != nil && len(*location) > 500 {
		return &models.ValidationError{
			Field:   "location",
			Message: "location cannot exceed 500 characters",
		}
	}
	return nil
}
