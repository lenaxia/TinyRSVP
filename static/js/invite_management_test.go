package js

import (
	"os"
	"strings"
	"testing"
)

func getInviteManagementJSPath() string {
	return "invite_management.js"
}

func TestInviteManagementJSExists(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Errorf("invite_management.js does not exist at %s", jsPath)
	}
}

func TestInviteManagementJSStructure(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	tests := []struct {
		name     string
		pattern  string
		required bool
	}{
		{"initInviteManagement function", "initInviteManagement", true},
		{"handleImportSubmit function", "handleImportSubmit", true},
		{"handleCreateSubmit function", "handleCreateSubmit", true},
		{"getEventIdFromURL function", "getEventIdFromURL", true},
		{"getCSRFToken function", "getCSRFToken", true},
		{"showFeedback function", "showFeedback", true},
		{"Modal usage", "Modal", true},
		{"fetch API usage", "fetch", true},
		{"FormData usage", "FormData", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(jsContent, tt.pattern) {
				if tt.required {
					t.Errorf("Required pattern '%s' not found in invite_management.js", tt.pattern)
				}
			}
		})
	}
}

func TestInviteManagementJSEventListeners(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	requiredListeners := []string{
		"submit",
		"addEventListener",
	}

	for _, listener := range requiredListeners {
		if !strings.Contains(jsContent, listener) {
			t.Errorf("JavaScript should contain '%s' event listener", listener)
		}
	}
}

func TestInviteManagementJSAPIEndpoints(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	requiredEndpoints := []string{
		"/api/events/",
		"/invites/import",
		"/invites",
		"/invites/",
	}

	foundCount := 0
	for _, endpoint := range requiredEndpoints {
		if strings.Contains(jsContent, endpoint) {
			foundCount++
		}
	}

	if foundCount < 2 {
		t.Errorf("JavaScript should reference invite API endpoints (found %d)", foundCount)
	}
}

func TestInviteManagementJSErrorHandling(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "catch") {
		t.Error("JavaScript should have error handling with catch blocks")
	}

	if !strings.Contains(jsContent, "error") {
		t.Error("JavaScript should handle error responses")
	}
}

func TestInviteManagementJSFormValidation(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	validationChecks := []string{
		".csv",
		"name",
		"email",
	}

	for _, check := range validationChecks {
		if !strings.Contains(jsContent, check) {
			t.Errorf("JavaScript should validate '%s'", check)
		}
	}
}

func TestInviteManagementJSNoConsoleLog(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	if strings.Contains(jsContent, "console.log") {
		t.Error("Production JavaScript should not contain console.log statements")
	}
}

func TestInviteManagementJSFileSize(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	info, err := os.Stat(jsPath)
	if err != nil {
		t.Fatalf("Failed to stat invite_management.js: %v", err)
	}

	maxSize := int64(25 * 1024)
	if info.Size() > maxSize {
		t.Errorf("invite_management.js is too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}
}

func TestInviteManagementJSValidSyntax(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	openBraces := strings.Count(jsContent, "{")
	closeBraces := strings.Count(jsContent, "}")

	if openBraces != closeBraces {
		t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
	}

	openParens := strings.Count(jsContent, "(")
	closeParens := strings.Count(jsContent, ")")

	if openParens != closeParens {
		t.Errorf("Mismatched parentheses: %d open, %d close", openParens, closeParens)
	}
}

func TestInviteManagementJSCSRFProtection(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "X-CSRF-Token") {
		t.Error("JavaScript should include CSRF token in requests")
	}
}

func TestInviteManagementJSUserFeedback(t *testing.T) {
	jsPath := getInviteManagementJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read invite_management.js: %v", err)
	}

	jsContent := string(content)

	feedbackTypes := []string{
		"success",
		"error",
	}

	for _, feedbackType := range feedbackTypes {
		if !strings.Contains(jsContent, feedbackType) {
			t.Errorf("JavaScript should provide '%s' feedback to users", feedbackType)
		}
	}
}
