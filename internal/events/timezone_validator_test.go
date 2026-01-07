package events

import (
	"testing"
)

func TestTimezoneValidator_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		want     bool
	}{
		{"valid US timezone", "America/Los_Angeles", true},
		{"valid EU timezone", "Europe/London", true},
		{"valid Asia timezone", "Asia/Tokyo", true},
		{"invalid timezone", "Invalid/Timezone", false},
		{"empty timezone", "", false},
		{"UTC", "UTC", true},
		{"valid Australia timezone", "Australia/Sydney", true},
		{"invalid format", "US/Pacific", true},
		{"completely invalid", "NotATimezone", false},
		{"numeric only", "123", false},
	}

	validator := NewTimezoneValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.IsValid(tt.timezone)
			if got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.timezone, got, tt.want)
			}
		})
	}
}

func TestTimezoneValidator_GetLocation(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{"valid timezone", "America/Los_Angeles", false},
		{"invalid timezone", "Invalid/Timezone", true},
		{"empty timezone", "", true},
		{"UTC", "UTC", false},
		{"valid Europe timezone", "Europe/Paris", false},
	}

	validator := NewTimezoneValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := validator.GetLocation(tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocation(%q) error = %v, wantErr %v", tt.timezone, err, tt.wantErr)
				return
			}
			if !tt.wantErr && loc == nil {
				t.Errorf("GetLocation(%q) returned nil location", tt.timezone)
			}
			if !tt.wantErr && loc.String() != tt.timezone {
				t.Errorf("GetLocation(%q) location.String() = %q, want %q", tt.timezone, loc.String(), tt.timezone)
			}
		})
	}
}
