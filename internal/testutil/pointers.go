package testutil

import "time"

// StringPtr returns a pointer to the given string value.
// Useful for populating optional string fields in test data.
//
// Example:
//
//	invite := &models.Invite{
//	    Email: testutil.StringPtr("test@example.com"),
//	}
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int value.
//
// Example:
//
//	event := &models.Event{
//	    Capacity: testutil.IntPtr(100),
//	}
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to the given int64 value.
//
// Example:
//
//	invite := &models.Invite{
//	    EventID: testutil.Int64Ptr(123),
//	}
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr returns a pointer to the given bool value.
//
// Example:
//
//	event := &models.Event{
//	    AllowMaybeRSVP: testutil.BoolPtr(true),
//	}
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to the given time.Time value.
//
// Example:
//
//	event := &models.Event{
//	    RSVPDeadline: testutil.TimePtr(time.Now().Add(24 * time.Hour)),
//	}
func TimePtr(t time.Time) *time.Time {
	return &t
}

// Float64Ptr returns a pointer to the given float64 value.
//
// Example:
//
//	data := &SomeModel{
//	    Score: testutil.Float64Ptr(3.14),
//	}
func Float64Ptr(f float64) *float64 {
	return &f
}
