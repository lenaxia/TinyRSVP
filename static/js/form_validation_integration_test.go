package js

import (
	"os"
	"strings"
	"testing"
)

func TestFormValidationIntegrationWithEventForm(t *testing.T) {
	eventFormPath := "../../templates/web/event_form.html"
	content, err := os.ReadFile(eventFormPath)
	if err != nil {
		t.Skipf("Skipping integration test: event_form.html not found: %v", err)
		return
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "form_validation.js") {
		t.Error("event_form.html should include form_validation.js script")
	}

	if !strings.Contains(htmlContent, "novalidate") {
		t.Error("event_form.html should have novalidate attribute for progressive enhancement")
	}
}

func TestFormValidationIntegrationWithRSVPPage(t *testing.T) {
	rsvpPagePath := "../../templates/web/rsvp_page.html"
	content, err := os.ReadFile(rsvpPagePath)
	if err != nil {
		t.Skipf("Skipping integration test: rsvp_page.html not found: %v", err)
		return
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "form_validation.js") {
		t.Error("rsvp_page.html should include form_validation.js script")
	}

	if !strings.Contains(htmlContent, "novalidate") {
		t.Error("rsvp_page.html should have novalidate attribute for progressive enhancement")
	}
}

func TestFormValidationErrorMessagesMatchCSS(t *testing.T) {
	cssPath := "../../static/css/forms.css"
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Skipf("Skipping integration test: forms.css not found: %v", err)
		return
	}

	cssContent := string(content)

	requiredClasses := []string{
		".form-error",
		".form-success",
		".error",
		".success",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(cssContent, class) {
			t.Errorf("forms.css should contain %s class for form validation", class)
		}
	}
}

func TestFormValidationAccessibility(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	accessibilityFeatures := []string{
		"aria-invalid",
		"aria-describedby",
		"role",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(jsContent, feature) {
			t.Errorf("form_validation.js should implement %s for accessibility", feature)
		}
	}
}

func TestFormValidationHandlesAllInputTypes(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	inputTypes := []string{
		"email",
		"datetime-local",
		"number",
		"text",
		"textarea",
		"select",
		"radio",
	}

	for _, inputType := range inputTypes {
		if !strings.Contains(jsContent, inputType) {
			t.Errorf("form_validation.js should handle %s input type", inputType)
		}
	}
}

func TestFormValidationPreventSubmitOnError(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "preventDefault") {
		t.Error("form_validation.js should prevent form submission on validation errors")
	}

	if !strings.Contains(jsContent, "submit") {
		t.Error("form_validation.js should handle form submit event")
	}
}

func TestFormValidationFocusFirstError(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "focus") {
		t.Error("form_validation.js should focus first error field")
	}

	if !strings.Contains(jsContent, "scrollIntoView") {
		t.Error("form_validation.js should scroll to first error field")
	}
}

func TestFormValidationRealTimeValidation(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "blur") {
		t.Error("form_validation.js should validate on blur event")
	}

	if !strings.Contains(jsContent, "input") {
		t.Error("form_validation.js should validate on input event for fields with errors")
	}
}

func TestFormValidationDateRangeValidation(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "validateDateRange") || !strings.Contains(jsContent, "validateDateTimeRange") {
		t.Error("form_validation.js should validate date ranges (end after start)")
	}

	if !strings.Contains(jsContent, "start_time") {
		t.Error("form_validation.js should handle start_time field")
	}

	if !strings.Contains(jsContent, "end_time") {
		t.Error("form_validation.js should handle end_time field")
	}
}

func TestFormValidationRSVPDeadlineValidation(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "validateRSVPDeadline") {
		t.Error("form_validation.js should validate RSVP deadline")
	}

	if !strings.Contains(jsContent, "rsvp_deadline") {
		t.Error("form_validation.js should handle rsvp_deadline field")
	}
}

func TestFormValidationMaxLengthValidation(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "maxlength") {
		t.Error("form_validation.js should validate maxlength attribute")
	}
}

func TestFormValidationMinMaxNumberValidation(t *testing.T) {
	jsPath := getJSPath()
	content, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read form_validation.js: %v", err)
	}

	jsContent := string(content)

	if !strings.Contains(jsContent, "validateNumber") {
		t.Error("form_validation.js should validate number ranges")
	}

	if !strings.Contains(jsContent, "min") {
		t.Error("form_validation.js should validate min attribute")
	}

	if !strings.Contains(jsContent, "max") {
		t.Error("form_validation.js should validate max attribute")
	}
}
