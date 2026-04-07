package web

import (
	"strings"
	"testing"
	"time"
)

func TestConfirmationTemplateIntegration(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	endTime := time.Now().Add(2 * time.Hour)
	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:       "Summer BBQ",
			Description: "Annual summer barbecue party",
			Location:    "Central Park",
			StartTime:   time.Now(),
			EndTime:     &endTime,
			Timezone:    "America/New_York",
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
			PlusOnes: 2,
		},
		AnswersWithQuestions: []ConfirmationAnswerWithQuestion{
			{
				Question: ConfirmationQuestion{Text: "Dietary restrictions?"},
				Answer:   ConfirmationAnswer{AnswerText: "Vegetarian"},
			},
			{
				Question: ConfirmationQuestion{Text: "T-shirt size?"},
				Answer:   ConfirmationAnswer{AnswerText: "Large"},
			},
		},
		Token: "abc123xyz",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredContent := []string{
		"Summer BBQ",
		"Annual summer barbecue party",
		"Central Park",
		"America/New_York",
		"YES",
		"2",
		"Dietary restrictions?",
		"Vegetarian",
		"T-shirt size?",
		"Large",
		"/api/calendar/abc123xyz",
		"/rsvp/abc123xyz",
	}

	for _, content := range requiredContent {
		if !strings.Contains(html, content) {
			t.Errorf("Template should contain: %s", content)
		}
	}
}

func TestConfirmationTemplateWithMinimalData(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Simple Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "no",
		},
		Token: "token123",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Simple Event") {
		t.Error("Template should display event title")
	}

	if !strings.Contains(html, "NO") {
		t.Error("Template should display response status")
	}

	if strings.Contains(html, "Guests:") {
		t.Error("Template should not show guests when PlusOnes is 0")
	}

	if strings.Contains(html, "Notes:") {
		t.Error("Template should not show notes when empty")
	}

	if strings.Contains(html, "Your Answers") {
		t.Error("Template should not show answers section when no answers")
	}
}

func TestConfirmationTemplateResponseTypes(t *testing.T) {
	responses := []struct {
		response      string
		expectedText  string
		expectedClass string
	}{
		{"yes", "YES", "response-yes"},
		{"no", "NO", "response-no"},
		{"maybe", "MAYBE", "response-maybe"},
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
				Token: "test",
			}

			var buf strings.Builder
			err = tmpl.Execute(&buf, data)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, tc.expectedText) {
				t.Errorf("Template should display response text: %s", tc.expectedText)
			}

			if !strings.Contains(html, tc.expectedClass) {
				t.Errorf("Template should include response class: %s", tc.expectedClass)
			}
		})
	}
}

func TestConfirmationTemplateAccessibilityIntegration(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Accessible Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	accessibilityFeatures := []string{
		`role="main"`,
		`role="status"`,
		`aria-live="polite"`,
		`aria-label`,
		`aria-labelledby`,
		`aria-hidden="true"`,
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(html, feature) {
			t.Errorf("Template should include accessibility feature: %s", feature)
		}
	}
}

func TestConfirmationTemplateSemanticHTMLIntegration(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Semantic Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test",
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
		"<time",
	}

	for _, element := range semanticElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template should use semantic element: %s", element)
		}
	}
}

func TestConfirmationTemplateTimeFormatting(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC)

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Time Test Event",
			StartTime: startTime,
			EndTime:   &endTime,
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, `datetime="2026-06-15T14:30:00Z"`) {
		t.Error("Template should include ISO 8601 datetime attribute for start time")
	}

	if !strings.Contains(html, `datetime="2026-06-15T18:00:00Z"`) {
		t.Error("Template should include ISO 8601 datetime attribute for end time")
	}
}

func TestConfirmationTemplateMultipleAnswers(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Survey Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		AnswersWithQuestions: []ConfirmationAnswerWithQuestion{
			{Question: ConfirmationQuestion{Text: "Question 1"}, Answer: ConfirmationAnswer{AnswerText: "Answer 1"}},
			{Question: ConfirmationQuestion{Text: "Question 2"}, Answer: ConfirmationAnswer{AnswerText: "Answer 2"}},
			{Question: ConfirmationQuestion{Text: "Question 3"}, Answer: ConfirmationAnswer{AnswerText: "Answer 3"}},
			{Question: ConfirmationQuestion{Text: "Question 4"}, Answer: ConfirmationAnswer{AnswerText: "Answer 4"}},
			{Question: ConfirmationQuestion{Text: "Question 5"}, Answer: ConfirmationAnswer{AnswerText: "Answer 5"}},
		},
		Token: "test",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	for i := 1; i <= 5; i++ {
		question := "Question " + string(rune('0'+i))
		answer := "Answer " + string(rune('0'+i))

		if !strings.Contains(html, question) {
			t.Errorf("Template should display question: %s", question)
		}

		if !strings.Contains(html, answer) {
			t.Errorf("Template should display answer: %s", answer)
		}
	}
}

func TestConfirmationTemplateActionLinks(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Link Test Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "unique-token-456",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	expectedLinks := []string{
		`href="/api/calendar/unique-token-456"`,
		`href="/rsvp/unique-token-456"`,
	}

	for _, link := range expectedLinks {
		if !strings.Contains(html, link) {
			t.Errorf("Template should include link: %s", link)
		}
	}

	linkClasses := []string{
		"calendar-download",
		"update-rsvp",
	}

	for _, class := range linkClasses {
		if !strings.Contains(html, class) {
			t.Errorf("Template should include link class: %s", class)
		}
	}
}

func TestConfirmationTemplateValidHTML(t *testing.T) {
	tmpl, err := getConfirmationTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := ConfirmationData{
		Event: ConfirmationEvent{
			Title:     "Valid HTML Event",
			StartTime: time.Now(),
		},
		RSVP: ConfirmationRSVP{
			Response: "yes",
		},
		Token: "test",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredHTMLElements := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"<head>",
		"<meta charset=\"UTF-8\">",
		"<meta name=\"viewport\"",
		"<title>",
		"</head>",
		"<body>",
		"</body>",
		"</html>",
	}

	for _, element := range requiredHTMLElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template should include HTML element: %s", element)
		}
	}
}
