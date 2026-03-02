package auth

import (
	"strings"
	"testing"
)

// TestValidateReturnURL_PenetrationTesting attempts to bypass the validation
// using creative attack payloads
func TestValidateReturnURL_PenetrationTesting(t *testing.T) {
	attackPayloads := []struct {
		name    string
		payload string
		reason  string
	}{
		// Unicode slash tricks
		{"unicode division slash", "/\u2044evil.com", "Unicode U+2044 looks like /"},
		{"unicode fraction slash", "/\u2215evil.com", "Unicode U+2215 looks like /"},
		{"unicode fullwidth slash", "/\uFF0Fevil.com", "Fullwidth solidus U+FF0F"},

		// Null byte injection
		{"null byte", "/\x00evil.com", "Null byte can truncate strings"},
		{"null byte encoded", "/events\x00?return=http://evil.com", "Null byte to hide malicious part"},

		// Triple slash
		{"triple slash", "///evil.com", "Three slashes might bypass // check"},

		// Valid prefix with fragment/query tricks
		{"query parameter injection", "/events?return=http://evil.com", "Valid path with malicious query"},
		{"fragment injection", "/events#http://evil.com", "Valid path with malicious fragment"},

		// @ sign confusion
		{"at sign redirect", "/@evil.com", "@ sign can cause URL confusion"},
		{"at sign with path", "/path@evil.com", "@ sign mid-path"},

		// Semicolon tricks
		{"semicolon parameter", "/;url=http://evil.com", "Semicolon parameter pollution"},

		// Dot tricks
		{"dot confusion", "/.evil.com", "Dot after slash"},
		{"double dot", "/../evil.com", "Path traversal attempt"},

		// Mixed encoding
		{"double encoded slash", "%252F%252Fevil.com", "Double URL encoding"},
		{"hex encoded slash", "%2F%2Fevil.com", "Hex encoded //"},
		{"mixed case scheme", "/HTTp://evil.com", "Mixed case to bypass checks"},

		// CRLF injection
		{"CRLF location header", "/%0d%0aLocation:%20http://evil.com", "Header injection"},
		{"CRLF with space", "/\r\nLocation: http://evil.com", "Raw CRLF"},

		// Space tricks
		{"leading space", " /events", "Space before slash"},
		{"trailing space", "/events ", "Space after path"},
		{"space in path", "/ events", "Space within path"},

		// Backslash variations
		{"backslash only", "\\evil.com", "Single backslash"},
		{"backslash double", "\\\\evil.com", "Double backslash"},
		{"forward then backslash", "/\\\\evil.com", "Mixed slashes"},

		// Very long URLs (DoS)
		{"extremely long URL", "/" + strings.Repeat("a", 100000), "DoS via memory exhaustion"},

		// File protocol
		{"file protocol", "file:///etc/passwd", "File protocol access"},
		{"file with encoded slash", "file%3A%2F%2F%2Fetc%2Fpasswd", "Encoded file protocol"},

		// Malformed URLs
		{"incomplete scheme", "http:/evil.com", "Missing second slash"},
		{"scheme without colon", "http//evil.com", "Scheme without colon"},

		// Path traversal
		{"path traversal up", "/events/../../etc/passwd", "Directory traversal"},
		{"path traversal encoded", "/events/%2e%2e/%2e%2e/etc/passwd", "Encoded traversal"},

		// Very long domains
		{"long domain", "//" + strings.Repeat("evil.", 100) + "com", "Long domain bypass"},

		// Control characters
		{"vertical tab", "/\vevil.com", "Vertical tab character"},
		{"form feed", "/\fevil.com", "Form feed character"},
		{"zero width space", "/\u200Bevil.com", "Zero-width space U+200B"},
		{"zero width non-joiner", "/\u200Cevil.com", "Zero-width non-joiner U+200C"},

		// IDN homograph attack
		{"cyrillic a", "/еvil.com", "Cyrillic 'e' looks like Latin 'e'"},

		// Case variations
		{"uppercase scheme", "/HTTP://evil.com", "Uppercase scheme"},
		{"mixed case protocol", "/hTTp://evil.com", "Mixed case protocol"},

		// Protocol-less with authority
		{"authority without protocol", "/evil.com:80/path", "Port number confusion"},

		// Punycode
		{"punycode domain", "//xn--evil-coma", "IDN encoded domain"},
	}

	for _, tt := range attackPayloads {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateReturnURL(tt.payload)

			// Security requirement: Either reject (err != nil) or sanitize to safe value
			if err == nil {
				// If accepted, must be safe (starts with / and no external host)
				if !strings.HasPrefix(result, "/") {
					t.Errorf("SECURITY BYPASS: Payload %q bypassed validation, resulted in %q (reason: %s)",
						tt.payload, result, tt.reason)
				}

				// Additional safety check: result should not contain obvious attack patterns
				lowerResult := strings.ToLower(result)
				dangerousPatterns := []string{"http:", "https:", "//", "\\", "\n", "\r", "\t", "javascript:", "data:", "file:"}
				for _, pattern := range dangerousPatterns {
					if strings.Contains(lowerResult, pattern) {
						t.Errorf("SECURITY BYPASS: Result contains dangerous pattern %q. Payload: %q, Result: %q (reason: %s)",
							pattern, tt.payload, result, tt.reason)
					}
				}
			} else {
				// If rejected, must default to safe value
				if result != "/" {
					t.Errorf("Rejected payload should default to /, got %q for payload %q", result, tt.payload)
				}
			}
		})
	}
}

// TestValidateReturnURL_NormalizationAttacks tests for Unicode normalization vulnerabilities
func TestValidateReturnURL_NormalizationAttacks(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{"NFC normalized", "/caf\u00e9"},       // é as single character
		{"NFD normalized", "/cafe\u0301"},      // é as e + combining accent
		{"NFKC compatible", "/\uff0fevil.com"}, // Fullwidth solidus
		{"NFKD compatible", "/\u2044evil.com"}, // Fraction slash
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateReturnURL(tt.payload)

			// These should either be rejected or kept as-is (not transformed into attack vectors)
			if err != nil && result != "/" {
				t.Errorf("Rejected payload should default to /, got %q", result)
			}

			if err == nil {
				// If accepted, ensure it's still safe
				if !strings.HasPrefix(result, "/") || strings.Contains(result, "//") {
					t.Errorf("SECURITY ISSUE: Normalization attack may have succeeded. Input: %q, Output: %q", tt.payload, result)
				}
			}
		})
	}
}

// TestValidateReturnURL_RaceCondition checks for race conditions
func TestValidateReturnURL_RaceCondition(t *testing.T) {
	// Run validation concurrently to check for race conditions
	done := make(chan bool)
	payloads := []string{
		"/events",
		"//evil.com",
		"http://evil.com",
		"/valid/path",
		"javascript:alert(1)",
	}

	for i := 0; i < 100; i++ {
		go func() {
			for _, payload := range payloads {
				result, err := ValidateReturnURL(payload)

				// Verify result is always consistent
				if payload == "/events" {
					if err != nil || result != "/events" {
						t.Errorf("Race condition detected: /events should always be valid")
					}
				} else if strings.HasPrefix(payload, "//") || strings.Contains(payload, "http") || strings.Contains(payload, "javascript") {
					if err == nil && result != "/" {
						t.Errorf("Race condition detected: malicious payload %q was not rejected", payload)
					}
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
