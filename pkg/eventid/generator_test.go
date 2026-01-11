package eventid

import (
	"testing"
)

func TestGenerateEventID(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{
			name:    "generates 10 character ID",
			wantLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := GenerateEventID()
			if err != nil {
				t.Errorf("GenerateEventID() error = %v", err)
				return
			}

			if len(id) != tt.wantLen {
				t.Errorf("GenerateEventID() length = %d, want %d", len(id), tt.wantLen)
			}

			// Verify it only contains base62 characters
			for _, c := range id {
				if !isBase62Char(c) {
					t.Errorf("GenerateEventID() contains invalid character: %c", c)
				}
			}
		})
	}
}

func TestGenerateEventID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id, err := GenerateEventID()
		if err != nil {
			t.Fatalf("GenerateEventID() error = %v", err)
		}

		if ids[id] {
			t.Errorf("GenerateEventID() generated duplicate ID: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}
}

func TestGenerateEventID_NoSequentialPattern(t *testing.T) {
	// Generate multiple IDs and ensure they don't follow a sequential pattern
	id1, err := GenerateEventID()
	if err != nil {
		t.Fatalf("GenerateEventID() error = %v", err)
	}

	id2, err := GenerateEventID()
	if err != nil {
		t.Fatalf("GenerateEventID() error = %v", err)
	}

	// IDs should be different
	if id1 == id2 {
		t.Errorf("GenerateEventID() generated identical IDs")
	}

	// IDs should not be sequential (simple check)
	if isSequential(id1, id2) {
		t.Errorf("GenerateEventID() appears to generate sequential IDs: %s, %s", id1, id2)
	}
}

func TestValidateEventID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid 10 character ID",
			id:      "aBcD123456",
			wantErr: false,
		},
		{
			name:    "too short",
			id:      "abc123",
			wantErr: true,
		},
		{
			name:    "too long",
			id:      "aBcD1234567890",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "contains invalid characters",
			id:      "aBcD123-56",
			wantErr: true,
		},
		{
			name:    "contains spaces",
			id:      "aBcD 12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEventID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEventID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions for tests
func isBase62Char(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z')
}

func isSequential(id1, id2 string) bool {
	// Simple check: if IDs differ by only 1 in the last character
	if len(id1) != len(id2) {
		return false
	}

	differences := 0
	for i := range id1 {
		if id1[i] != id2[i] {
			differences++
		}
	}

	// If only one character differs, might be sequential
	return differences == 1
}
