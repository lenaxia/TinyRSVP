package rsvp

import (
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestPlusOnesValidator_ValidatePlusOnes(t *testing.T) {
	tests := []struct {
		name     string
		plusOnes int
		response models.RSVPResponse
		invite   *models.Invite
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid within limit",
			plusOnes: 2,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  false,
		},
		{
			name:     "zero plus ones",
			plusOnes: 0,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  false,
		},
		{
			name:     "at maximum limit",
			plusOnes: 5,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  false,
		},
		{
			name:     "exceeds limit",
			plusOnes: 6,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  true,
			errMsg:   "can bring up to 5 guest(s)",
		},
		{
			name:     "negative plus ones",
			plusOnes: -1,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  true,
			errMsg:   "cannot be negative",
		},
		{
			name:     "no response with plus ones",
			plusOnes: 2,
			response: models.RSVPResponseNo,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  true,
			errMsg:   "Cannot bring guests when declining",
		},
		{
			name:     "no response with zero plus ones",
			plusOnes: 0,
			response: models.RSVPResponseNo,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  false,
		},
		{
			name:     "maybe response with plus ones",
			plusOnes: 2,
			response: models.RSVPResponseMaybe,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  false,
		},
		{
			name:     "zero max allowed with zero plus ones",
			plusOnes: 0,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 0},
			wantErr:  false,
		},
		{
			name:     "zero max allowed with one plus one",
			plusOnes: 1,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 0},
			wantErr:  true,
			errMsg:   "can bring up to 0 guest(s)",
		},
		{
			name:     "maximum allowed plus ones (10)",
			plusOnes: 10,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 10},
			wantErr:  false,
		},
		{
			name:     "exceeds maximum allowed plus ones",
			plusOnes: 11,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 10},
			wantErr:  true,
			errMsg:   "can bring up to 10 guest(s)",
		},
		{
			name:     "maybe response with zero plus ones",
			plusOnes: 0,
			response: models.RSVPResponseMaybe,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  false,
		},
		{
			name:     "yes response at exact limit",
			plusOnes: 3,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 3},
			wantErr:  false,
		},
		{
			name:     "large negative value",
			plusOnes: -100,
			response: models.RSVPResponseYes,
			invite:   &models.Invite{MaxPlusOnes: 5},
			wantErr:  true,
			errMsg:   "cannot be negative",
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePlusOnes(tt.plusOnes, tt.response, tt.invite)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlusOnes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestPlusOnesValidator_ValidatePlusOnes_NilInvite(t *testing.T) {
	validator := NewValidator()

	err := validator.ValidatePlusOnes(1, models.RSVPResponseYes, nil)
	if err == nil {
		t.Error("ValidatePlusOnes() expected error for nil invite, got nil")
	}
}

func TestPlusOnesValidator_ValidatePlusOnes_EmptyResponse(t *testing.T) {
	validator := NewValidator()
	invite := &models.Invite{MaxPlusOnes: 5}

	err := validator.ValidatePlusOnes(1, models.RSVPResponse(""), invite)
	if err == nil {
		t.Error("ValidatePlusOnes() expected error for empty response, got nil")
	}
}

func TestPlusOnesValidator_ValidatePlusOnes_InvalidResponse(t *testing.T) {
	validator := NewValidator()
	invite := &models.Invite{MaxPlusOnes: 5}

	err := validator.ValidatePlusOnes(1, models.RSVPResponse("invalid"), invite)
	if err == nil {
		t.Error("ValidatePlusOnes() expected error for invalid response, got nil")
	}
}
