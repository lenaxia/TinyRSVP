package web

import (
	"html/template"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRSVPPageTemplateExists(t *testing.T) {
	_, err := os.Stat("rsvp_page.html")
	if err != nil {
		t.Fatalf("rsvp_page.html template does not exist: %v", err)
	}
}

func TestRSVPPageTemplateNotEmpty(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	if len(content) == 0 {
		t.Error("rsvp_page.html is empty")
	}
}

func TestRSVPPageTemplateValidHTML(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE declaration")
	}

	if !strings.Contains(html, "<html") {
		t.Error("Missing html tag")
	}

	if !strings.Contains(html, "<head>") {
		t.Error("Missing head tag")
	}

	if !strings.Contains(html, "<body") {
		t.Error("Missing body tag")
	}
}

func TestRSVPPageTemplateHasMetaTags(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	requiredMeta := []string{
		`<meta charset="UTF-8">`,
		`<meta name="viewport"`,
	}

	for _, meta := range requiredMeta {
		if !strings.Contains(html, meta) {
			t.Errorf("Missing required meta tag: %s", meta)
		}
	}
}

func TestRSVPPageTemplateHasStylesheets(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	requiredCSS := []string{
		"variables.css",
		"typography.css",
		"colors.css",
		"spacing.css",
		"buttons.css",
		"forms.css",
		"rsvp_page.css",
	}

	for _, css := range requiredCSS {
		if !strings.Contains(html, css) {
			t.Errorf("Missing required stylesheet: %s", css)
		}
	}
}

func TestRSVPPageTemplateHasEventDetails(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	requiredElements := []string{
		"event-details",
		"event-title",
		"event-info",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Missing required element: %s", elem)
		}
	}
}

func TestRSVPPageTemplateHasRSVPForm(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	requiredElements := []string{
		"rsvp-form",
		"response-options",
		`type="radio"`,
		`name="response"`,
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Missing required form element: %s", elem)
		}
	}
}

func TestRSVPPageTemplateHasResponseOptions(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	requiredOptions := []string{
		`value="yes"`,
		`value="no"`,
		`value="maybe"`,
	}

	for _, option := range requiredOptions {
		if !strings.Contains(html, option) {
			t.Errorf("Missing required response option: %s", option)
		}
	}
}

func TestRSVPPageTemplateHasPlusOnesSelector(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	requiredElements := []string{
		"plus-ones-selector",
		`name="plus_ones"`,
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Missing required plus-ones element: %s", elem)
		}
	}
}

func TestRSVPPageTemplateHasQuestionsSection(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, "preference-questions") {
		t.Error("Missing preference-questions section")
	}

	if !strings.Contains(html, "{{range .Questions}}") && !strings.Contains(html, "{{ range .Questions }}") {
		t.Error("Missing template range for questions")
	}
}

func TestRSVPPageTemplateHasSubmitButton(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, `type="submit"`) {
		t.Error("Missing submit button")
	}
}

func TestRSVPPageTemplateHasErrorHandling(t *testing.T) {
	content, err := os.ReadFile("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_page.html: %v", err)
	}

	html := string(content)

	if !strings.Contains(html, ".ErrorMessage") {
		t.Error("Missing error message handling")
	}
}

func TestRSVPPageTemplateCanParse(t *testing.T) {
	tmpl, err := template.ParseFiles("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	if tmpl == nil {
		t.Error("Template is nil after parsing")
	}
}

func TestRSVPPageTemplateRendersWithValidData(t *testing.T) {
	tmpl, err := template.ParseFiles("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	now := time.Now()
	startTime := now.Add(24 * time.Hour)

	data := &handlers.RSVPPageData{
		Event: &models.Event{
			ID:          1,
			Title:       "Test Event",
			Description: strPtr("Test Description"),
			Location:    strPtr("Test Location"),
			StartTime:   startTime,
			Timezone:    "America/Los_Angeles",
			MaxPlusOnes: 2,
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   strPtr("test@example.com"),
		},
		Questions:      []*handlers.QuestionWithOptions{},
		Token:          "test-token",
		DeadlinePassed: false,
		EventPassed:    false,
		LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
		TimeUntilEvent: "1 day",
		CanUpdate:      false,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Template rendered empty output")
	}

	if !strings.Contains(output, "Test Event") {
		t.Error("Template output missing event title")
	}
}

func TestRSVPPageTemplateRendersWithExistingRSVP(t *testing.T) {
	tmpl, err := template.ParseFiles("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	now := time.Now()
	startTime := now.Add(24 * time.Hour)

	data := &handlers.RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Test Event",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   strPtr("test@example.com"),
		},
		ExistingRSVP: &models.RSVP{
			ID:       1,
			InviteID: 1,
			Response: models.RSVPResponseYes,
			PlusOnes: 1,
		},
		Questions:      []*handlers.QuestionWithOptions{},
		Token:          "test-token",
		DeadlinePassed: false,
		EventPassed:    false,
		LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
		TimeUntilEvent: "1 day",
		CanUpdate:      true,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Template rendered empty output")
	}
}

func TestRSVPPageTemplateRendersWithQuestions(t *testing.T) {
	tmpl, err := template.ParseFiles("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	now := time.Now()
	startTime := now.Add(24 * time.Hour)

	data := &handlers.RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Test Event",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   strPtr("test@example.com"),
		},
		Questions: []*handlers.QuestionWithOptions{
			{
				PreferenceQuestion: &models.PreferenceQuestion{
					ID:           1,
					EventID:      1,
					QuestionText: "Dietary restrictions?",
					QuestionType: "text",
					Required:     true,
				},
				ParsedOptions: []string{},
			},
		},
		Token:          "test-token",
		DeadlinePassed: false,
		EventPassed:    false,
		LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
		TimeUntilEvent: "1 day",
		CanUpdate:      false,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Template rendered empty output")
	}

	if !strings.Contains(output, "Dietary restrictions?") {
		t.Error("Template output missing question text")
	}
}

func TestRSVPPageTemplateRendersErrorState(t *testing.T) {
	tmpl, err := template.ParseFiles("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := &handlers.RSVPPageData{
		ErrorMessage: "This invite has expired",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Template rendered empty output")
	}

	if !strings.Contains(output, "This invite has expired") {
		t.Error("Template output missing error message")
	}
}

func strPtr(s string) *string {
	return &s
}
