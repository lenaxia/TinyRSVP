package events

import (
	"testing"
	"time"
)

func TestTimezoneValidator_AmericaLosAngeles(t *testing.T) {
	validator := NewTimezoneValidator()

	t.Run("America/Los_Angeles is valid", func(t *testing.T) {
		tz := "America/Los_Angeles"
		
		loc, err := time.LoadLocation(tz)
		if err != nil {
			t.Fatalf("time.LoadLocation failed for %s: %v", tz, err)
		}
		t.Logf("Successfully loaded timezone: %s", loc.String())

		if !validator.IsValid(tz) {
			t.Errorf("Validator rejected valid timezone: %s", tz)
		}

		location, err := validator.GetLocation(tz)
		if err != nil {
			t.Errorf("GetLocation failed for %s: %v", tz, err)
		}
		if location == nil {
			t.Error("GetLocation returned nil location")
		}
	})

	t.Run("all common US timezones are valid", func(t *testing.T) {
		timezones := []string{
			"America/Los_Angeles",
			"America/Denver",
			"America/Chicago",
			"America/New_York",
			"UTC",
		}

		for _, tz := range timezones {
			if !validator.IsValid(tz) {
				t.Errorf("Validator rejected valid timezone: %s", tz)
			}
		}
	})
}
