package rsvp

import (
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type PlusOnesValidator interface {
	ValidatePlusOnes(plusOnes int, response models.RSVPResponse, invite *models.Invite) error
}

type validator struct{}

func NewValidator() PlusOnesValidator {
	return &validator{}
}

func (v *validator) ValidatePlusOnes(plusOnes int, response models.RSVPResponse, invite *models.Invite) error {
	if invite == nil {
		return &models.ValidationError{
			Field:   "invite",
			Message: "invite cannot be nil",
		}
	}

	if !response.Valid() {
		return &models.ValidationError{
			Field:   "response",
			Message: "response must be yes, no, or maybe",
		}
	}

	if response == models.RSVPResponseNo && plusOnes > 0 {
		return &models.ValidationError{
			Field:   "plus_ones",
			Message: "Cannot bring guests when declining",
		}
	}

	if plusOnes < 0 {
		return &models.ValidationError{
			Field:   "plus_ones",
			Message: "plus_ones cannot be negative",
		}
	}

	if plusOnes > invite.MaxPlusOnes {
		return &models.ValidationError{
			Field:   "plus_ones",
			Message: fmt.Sprintf("you can bring up to %d guest(s)", invite.MaxPlusOnes),
		}
	}

	return nil
}
