package web

import (
	"html/template"
	"strings"
	"testing"
	"time"
)

// ConfirmationQuestion mirrors the Question sub-struct used in the template.
type ConfirmationQuestion struct {
	Text string
}

// ConfirmationAnswer mirrors the Answer sub-struct used in the template.
type ConfirmationAnswer struct {
	AnswerText   string
	AnswerOption string
	// Legacy field kept for tests that pass display text directly.
	AnswerDisplay string
}

// ConfirmationAnswerWithQuestion mirrors AnswerWithQuestion used by the template.
type ConfirmationAnswerWithQuestion struct {
	Question ConfirmationQuestion
	Answer   ConfirmationAnswer
}

type ConfirmationEvent struct {
	Title       string
	Description string
	Location    string
	StartTime   time.Time
	EndTime     *time.Time
	Timezone    string
}

type ConfirmationRSVP struct {
	Response string
	PlusOnes int
}

type ConfirmationData struct {
	ActivePage           string
	Event                ConfirmationEvent
	RSVP                 ConfirmationRSVP
	AnswersWithQuestions []ConfirmationAnswerWithQuestion
	Token                string
	ErrorMessage         string
	CanUpdate            bool
	LocalStartTime       string
	LocalEndTime         string
}

func getConfirmationTemplate() (*template.Template, error) {
	return template.New("confirmation.html").Funcs(testFuncMap()).ParseFiles("confirmation.html")
}

func TestConfirmationTemplateExists(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse confirmation.html: %v", err)
	}
	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestConfirmationTemplateRenders(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:       "Birthday Party",
			Description: "Join us for a celebration",
			Location:    "123 Main St",
			StartTime:   time.Now(),
			Timezone:    "America/Los_Angeles",
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
			PlusOnes: 2,
		},
		Token: "test-token-123",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	if html == "" {
		t.Fatal("Template rendered empty string")
	}
}

func TestConfirmationTemplateStructure(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"confirmation",
		"confirmation-success",
		"confirmation-summary",
		"confirmation-details",
		"confirmation-actions",
	}

	for _, element := range requiredElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template missing required element: %s", element)
		}
	}
}

func TestConfirmationTemplateSuccessMessage(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	successElements := []string{
		"confirmation-success",
		"RSVP Confirmed",
		"Thank you",
	}

	for _, element := range successElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template should display success element: %s", element)
		}
	}
}

func TestConfirmationTemplateEventDetails(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:       "Birthday Party",
			Description: "Join us for a celebration",
			Location:    "123 Main St",
			StartTime:   time.Now(),
			Timezone:    "America/Los_Angeles",
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	eventDetails := []string{
		"Birthday Party",
		"Join us for a celebration",
		"123 Main St",
		"America/Los_Angeles",
	}

	for _, detail := range eventDetails {
		if !strings.Contains(html, detail) {
			t.Errorf("Template should display event detail: %s", detail)
		}
	}
}

func TestConfirmationTemplateRSVPSummary(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
			PlusOnes: 2,
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	rsvpDetails := []string{
		"Your Response",
		"YES",
		"2",
	}

	for _, detail := range rsvpDetails {
		if !strings.Contains(html, detail) {
			t.Errorf("Template should display RSVP detail: %s", detail)
		}
	}
}

func TestConfirmationTemplateResponseStatus(t *testing.T) {
	responses := []struct {
		response string
		expected string
	}{
		{"yes", "response-yes"},
		{"no", "response-no"},
		{"maybe", "response-maybe"},
	}

	for _, tc := range responses {
		t.Run(tc.response, func(t *testing.T) {
			tmpl, err := getConfirmationTemplate()
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			data := ConfirmationData{
				Event: ConfirmationEvent{
					Title:     "Test Event",
					StartTime: time.Now(),
				},
				RSVP: ConfirmationRSVP{
					Response: tc.response,
				},
				Token: "test-token",
			}

			var buf strings.Builder
			err = tmpl.Execute(&buf, data)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, tc.expected) {
				t.Errorf("Template should include response class: %s", tc.expected)
			}
		})
	}
}

func TestConfirmationTemplateAnswers(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		AnswersWithQuestions: []ConfirmationAnswerWithQuestion{
			{
				Question: ConfirmationQuestion{Text: "Dietary restrictions?"},
				Answer:   ConfirmationAnswer{AnswerText: "Vegetarian"},
			},
			{
				Question: ConfirmationQuestion{Text: "T-shirt size?"},
				Answer:   ConfirmationAnswer{AnswerText: "Medium"},
			},
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	answerDetails := []string{
		"Your Answers",
		"Dietary restrictions?",
		"Vegetarian",
		"T-shirt size?",
		"Medium",
	}

	for _, detail := range answerDetails {
		if !strings.Contains(html, detail) {
			t.Errorf("Template should display answer detail: %s", detail)
		}
	}
}

func TestConfirmationTemplateNoAnswers(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		AnswersWithQuestions: []ConfirmationAnswerWithQuestion{},
		Token:                "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, "Your Answers") {
		t.Error("Template should not display answers section when no answers")
	}
}

func TestConfirmationTemplateActions(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token-123",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	actionElements := []string{
		"confirmation-actions",
		"calendar-download",
		"update-rsvp",
		"/api/calendar/test-token-123",
		"/rsvp/test-token-123",
		"Add to Calendar",
		"Update RSVP",
	}

	for _, element := range actionElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template should include action element: %s", element)
		}
	}
}

func TestConfirmationTemplateAccessibility(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	accessibilityFeatures := []string{
		"<main",
		"role=",
		"aria-",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(html, feature) {
			t.Errorf("Template should include accessibility feature: %s", feature)
		}
	}
}

func TestConfirmationTemplateSemanticHTML(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	semanticElements := []string{
		"<main",
		"<section",
		"<h1",
		"<h2",
	}

	for _, element := range semanticElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template should use semantic HTML element: %s", element)
		}
	}
}

func TestConfirmationTemplateLinksCSS(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test-token",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	cssFiles := []string{
		"/static/css/variables.css",
		"/static/css/typography.css",
		"/static/css/colors.css",
		"/static/css/spacing.css",
		"/static/css/grid.css",
		"/static/css/buttons.css",
		"/static/css/confirmation.css",
	}

	for _, cssFile := range cssFiles {
		if !strings.Contains(html, cssFile) {
			t.Errorf("Template should link to CSS file: %s", cssFile)
		}
	}
}
