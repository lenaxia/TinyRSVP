package handlers

import "testing"

func TestResolveScheme(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"https", "https", "https"},
		{"http", "http", "http"},
		{"empty defaults to https", "", "https"},
		{"invalid scheme defaults to https", "ftp", "https"},
		{"javascript blocked", "javascript", "https"},
		{"mixed case rejected", "HTTPS", "https"},
		{"with port rejected", "https://", "https"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveScheme(tt.input)
			if result != tt.expected {
				t.Errorf("resolveScheme(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
