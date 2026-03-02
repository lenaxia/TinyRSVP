package auth

import (
	"testing"
)

func TestValidateReturnURL_CreativeAttacks(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "uppercase HTTP in query",
			url:     "/events?url=HTTP://evil.com",
			wantErr: true,
		},
		{
			name:    "uppercase HTTPS in fragment",
			url:     "/events#HTTPS://evil.com",
			wantErr: true,
		},
		{
			name:    "null byte after double slash",
			url:     "//\x00evil.com",
			wantErr: true,
		},
		{
			name:    "null byte in scheme",
			url:     "/HTTP\x00://evil.com",
			wantErr: true,
		},
		{
			name:    "null byte splitting scheme",
			url:     "/h\x00ttp://evil.com",
			wantErr: true,
		},
		{
			name:    "javascript no equals",
			url:     "/events?javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "tab in scheme",
			url:     "/HTT\tP://evil.com",
			wantErr: true,
		},
		{
			name:    "newline in scheme",
			url:     "/JAVA\nSCRIPT:alert(1)",
			wantErr: true,
		},
		{
			name:    "multiple slashes",
			url:     "///evil.com",
			wantErr: true,
		},
		{
			name:    "encoded null byte",
			url:     "//%00evil.com",
			wantErr: true,
		},
		{
			name:    "encoded tab",
			url:     "//%09evil.com",
			wantErr: true,
		},
		{
			name:    "space before scheme",
			url:     "/ http://evil.com",
			wantErr: true,
		},
		{
			name:    "unicode lookalike colon",
			url:     "/http\u003Aevil.com",
			wantErr: true,
		},
		{
			name:    "combining characters - POTENTIAL BYPASS",
			url:     "/http\u0301://evil.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateReturnURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReturnURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
			if err != nil && result != "/" {
				t.Errorf("ValidateReturnURL(%q) should return / on error, got %q", tt.url, result)
			}
		})
	}
}
