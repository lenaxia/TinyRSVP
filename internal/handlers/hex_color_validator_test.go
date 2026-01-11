package handlers

import (
	"testing"
)

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{
			name:  "valid uppercase hex",
			color: "#FF5733",
			want:  true,
		},
		{
			name:  "valid lowercase hex",
			color: "#ff5733",
			want:  true,
		},
		{
			name:  "valid mixed case hex",
			color: "#Ff5733",
			want:  true,
		},
		{
			name:  "invalid - too short",
			color: "#FFF",
			want:  false,
		},
		{
			name:  "invalid - too long",
			color: "#FF5733AA",
			want:  false,
		},
		{
			name:  "invalid - no hash",
			color: "FF5733",
			want:  false,
		},
		{
			name:  "invalid - invalid characters",
			color: "#GGGGGG",
			want:  false,
		},
		{
			name:  "invalid - empty string",
			color: "",
			want:  false,
		},
		{
			name:  "invalid - only hash",
			color: "#",
			want:  false,
		},
		{
			name:  "invalid - spaces",
			color: "#FF 5733",
			want:  false,
		},
		{
			name:  "valid - all zeros",
			color: "#000000",
			want:  true,
		},
		{
			name:  "valid - all Fs",
			color: "#FFFFFF",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidHexColor(tt.color)
			if got != tt.want {
				t.Errorf("isValidHexColor(%q) = %v, want %v", tt.color, got, tt.want)
			}
		})
	}
}
