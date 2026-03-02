package auth

import (
	"strings"
	"testing"
)

func TestValidateReturnURL_Security(t *testing.T) {
	tests := []struct {
		name      string
		returnURL string
		wantURL   string
		wantErr   bool
	}{
		{"empty defaults to root", "", "/", false},
		{"relative path allowed", "/events", "/events", false},
		{"relative with query", "/events?page=2", "/events?page=2", false},
		{"nested path allowed", "/events/123/edit", "/events/123/edit", false},

		{"http blocked", "http://evil.com", "/", true},
		{"https blocked", "https://evil.com", "/", true},
		{"ftp blocked", "ftp://evil.com", "/", true},

		{"protocol relative blocked", "//evil.com", "/", true},
		{"protocol relative path blocked", "//evil.com/path", "/", true},

		{"javascript blocked", "javascript:alert(1)", "/", true},
		{"data URL blocked", "data:text/html,<script>alert(1)</script>", "/", true},

		{"encoded http blocked", "%68%74%74%70%3A%2F%2Fevil.com", "/", true},
		{"url encoded blocked", "http%3A%2F%2Fevil.com", "/", true},

		{"backslash redirect blocked", "\\evil.com", "/", true},
		{"mixed slashes blocked", "/\\evil.com", "/", true},

		{"tab injection blocked", "/\tevil.com", "/", true},
		{"newline injection blocked", "/\nevil.com", "/", true},
		{"carriage return blocked", "/\revil.com", "/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := ValidateReturnURL(tt.returnURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReturnURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("ValidateReturnURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestValidateReturnURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		returnURL string
		wantURL   string
		wantErr   bool
	}{
		{"very long path allowed", "/" + strings.Repeat("a", 1000), "/" + strings.Repeat("a", 1000), false},
		{"path with dots allowed", "/events/../events", "/events/../events", false},
		{"path with special chars", "/events?q=test&sort=asc#section", "/events?q=test&sort=asc#section", false},
		{"percent encoded path allowed", "/events%2F123", "/events%2F123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := ValidateReturnURL(tt.returnURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReturnURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("ValidateReturnURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
