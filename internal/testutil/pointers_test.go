package testutil_test

import (
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

func TestStringPtr(t *testing.T) {
	value := "test"
	ptr := testutil.StringPtr(value)

	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *ptr != value {
		t.Errorf("Expected %q, got %q", value, *ptr)
	}
}

func TestIntPtr(t *testing.T) {
	value := 42
	ptr := testutil.IntPtr(value)

	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *ptr != value {
		t.Errorf("Expected %d, got %d", value, *ptr)
	}
}

func TestInt64Ptr(t *testing.T) {
	value := int64(42)
	ptr := testutil.Int64Ptr(value)

	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *ptr != value {
		t.Errorf("Expected %d, got %d", value, *ptr)
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := testutil.BoolPtr(tt.value)
			if ptr == nil {
				t.Fatal("Expected non-nil pointer")
			}
			if *ptr != tt.value {
				t.Errorf("Expected %v, got %v", tt.value, *ptr)
			}
		})
	}
}

func TestTimePtr(t *testing.T) {
	value := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ptr := testutil.TimePtr(value)

	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if !ptr.Equal(value) {
		t.Errorf("Expected %v, got %v", value, *ptr)
	}
}

func TestFloat64Ptr(t *testing.T) {
	value := 3.14
	ptr := testutil.Float64Ptr(value)

	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *ptr != value {
		t.Errorf("Expected %f, got %f", value, *ptr)
	}
}
