package web

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRSVPPageFullIntegration(t *testing.T) {
	tmpl, err := template.ParseFiles("rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	t.Run("renders complete RSVP page with all elements", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)
		deadline := now.Add(24 * time.Hour)

		data := &handlers.RSVPPageData{
			Event: &models.Event{
				ID:             1,
				Title:          "Summer BBQ Party",
				Description:    strPtr("Join us for a fun summer BBQ!"),
				Location:       strPtr("123 Main St, Anytown, USA"),
				StartTime:      startTime,
				Timezone:       "America/Los_Angeles",
				RSVPDeadline:   &deadline,
				MaxPlusOnes:    3,
				AllowMaybeRSVP: true,
			},
			Invite: &models.Invite{
				ID:          1,
				EventID:     1,
				Name:        strPtr("John Doe"),
				Email:       strPtr("john@example.com"),
				MaxPlusOnes: 2,
			},
			Questions: []*handlers.QuestionWithOptions{
				{
					PreferenceQuestion: &models.PreferenceQuestion{
						ID:           1,
						EventID:      1,
						QuestionText: "Any dietary restrictions?",
						QuestionType: "text",
						Required:     false,
					},
					ParsedOptions: []string{},
				},
				{
					PreferenceQuestion: &models.PreferenceQuestion{
						ID:           2,
						EventID:      1,
						QuestionText: "Meal preference?",
						QuestionType: "single_choice",
						Required:     true,
					},
					ParsedOptions: []string{"Beef", "Chicken", "Vegetarian"},
				},
			},
			Token:          "test-token-123",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			TimeUntilEvent: "2 days",
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		requiredElements := []struct {
			name    string
			content string
		}{
			{"event title", "Summer BBQ Party"},
			{"event description", "Join us for a fun summer BBQ!"},
			{"event location", "123 Main St, Anytown, USA"},
			{"invited guest name", "John Doe"},
			{"invited guest email", "john@example.com"},
			{"response yes option", `value="yes"`},
			{"response no option", `value="no"`},
			{"response maybe option", `value="maybe"`},
			{"plus ones input", `name="plus_ones"`},
			{"text question", "Any dietary restrictions?"},
			{"choice question", "Meal preference?"},
			{"submit button", `type="submit"`},
			{"form action", `/rsvp/test-token-123`},
		}

		for _, elem := range requiredElements {
			if !strings.Contains(html, elem.content) {
				t.Errorf("Missing %s: expected to find %q", elem.name, elem.content)
			}
		}
	})

	t.Run("renders with existing RSVP", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

		data := &handlers.RSVPPageData{
			Event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
				Timezone:  "America/Los_Angeles",
			},
			Invite: &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
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
			TimeUntilEvent: "2 days",
			CanUpdate:      true,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, "You've Already Responded") {
			t.Error("Missing existing RSVP notice")
		}

		if !strings.Contains(html, "Update RSVP") {
			t.Error("Should show 'Update RSVP' button text when existing RSVP present")
		}

		if !strings.Contains(html, `checked`) {
			t.Error("Existing response should be pre-selected")
		}
	})

	t.Run("renders error state", func(t *testing.T) {
		data := &handlers.RSVPPageData{
			ErrorMessage: "This invite has expired",
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, "This invite has expired") {
			t.Error("Missing error message")
		}

		if !strings.Contains(html, "alert-error") {
			t.Error("Missing error alert styling")
		}

		if strings.Contains(html, "rsvp-form") {
			t.Error("Should not show form when error present")
		}
	})

	t.Run("renders deadline passed state", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

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
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: true,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, "RSVP Deadline Passed") {
			t.Error("Missing deadline passed warning")
		}

		if strings.Contains(html, `<form`) {
			t.Error("Should not show form when deadline passed")
		}
	})

	t.Run("renders event passed state", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(-24 * time.Hour)

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
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    true,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, "Event Has Passed") {
			t.Error("Missing event passed warning")
		}

		if strings.Contains(html, `<form`) {
			t.Error("Should not show form when event passed")
		}
	})

	t.Run("CSS classes are properly applied", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

		data := &handlers.RSVPPageData{
			Event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
				Timezone:  "America/Los_Angeles",
			},
			Invite: &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			TimeUntilEvent: "2 days",
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		requiredClasses := []string{
			"rsvp-page",
			"rsvp-container",
			"event-details",
			"event-title",
			"event-info",
			"rsvp-form",
			"response-options",
			"response-option",
			"plus-ones-selector",
			"plus-ones-controls",
			"plus-ones-btn",
			"plus-ones-value",
			"rsvp-actions",
			"btn",
			"btn-primary",
		}

		for _, class := range requiredClasses {
			if !strings.Contains(html, class) {
				t.Errorf("Missing CSS class: %s", class)
			}
		}
	})

	t.Run("form validation attributes present", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

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
			},
			Questions: []*handlers.QuestionWithOptions{
				{
					PreferenceQuestion: &models.PreferenceQuestion{
						ID:           1,
						EventID:      1,
						QuestionText: "Required question?",
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
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		validationAttributes := []string{
			`required`,
			`aria-required="true"`,
			`maxlength="500"`,
			`novalidate`,
		}

		for _, attr := range validationAttributes {
			if !strings.Contains(html, attr) {
				t.Errorf("Missing validation attribute: %s", attr)
			}
		}
	})

	t.Run("accessibility attributes present", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

		data := &handlers.RSVPPageData{
			Event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
				Timezone:  "America/Los_Angeles",
			},
			Invite: &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		a11yAttributes := []string{
			`role="main"`,
			`aria-label`,
			`aria-hidden="true"`,
		}

		for _, attr := range a11yAttributes {
			if !strings.Contains(html, attr) {
				t.Errorf("Missing accessibility attribute: %s", attr)
			}
		}
	})

	t.Run("JavaScript functionality present", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

		data := &handlers.RSVPPageData{
			Event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
				Timezone:  "America/Los_Angeles",
			},
			Invite: &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		jsFeatures := []string{
			"DOMContentLoaded",
			"updatePlusOnesState",
			"addEventListener",
			"querySelector",
		}

		for _, feature := range jsFeatures {
			if !strings.Contains(html, feature) {
				t.Errorf("Missing JavaScript feature: %s", feature)
			}
		}
	})

	t.Run("works without JavaScript", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

		data := &handlers.RSVPPageData{
			Event: &models.Event{
				ID:        1,
				Title:     "Test Event",
				StartTime: startTime,
				Timezone:  "America/Los_Angeles",
			},
			Invite: &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, `method="POST"`) {
			t.Error("Form should use POST method for no-JS support")
		}

		if !strings.Contains(html, `type="radio"`) {
			t.Error("Should use native radio inputs for no-JS support")
		}

		if !strings.Contains(html, `type="submit"`) {
			t.Error("Should have submit button for no-JS support")
		}
	})

	t.Run("CSS file serves correctly", func(t *testing.T) {
		handler := http.FileServer(http.Dir("../../static/css"))
		req := httptest.NewRequest("GET", "/rsvp_page.css", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if len(body) == 0 {
			t.Error("CSS file is empty")
		}

		if !strings.Contains(body, ".rsvp-page") {
			t.Error("CSS file missing .rsvp-page class")
		}
	})

	t.Run("renders multiple choice questions correctly", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

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
			},
			Questions: []*handlers.QuestionWithOptions{
				{
					PreferenceQuestion: &models.PreferenceQuestion{
						ID:           1,
						EventID:      1,
						QuestionText: "Select activities",
						QuestionType: "multiple_choice",
						Required:     false,
					},
					ParsedOptions: []string{"Swimming", "Hiking", "Biking"},
				},
			},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, `type="checkbox"`) {
			t.Error("Multiple choice questions should use checkboxes")
		}

		options := []string{"Swimming", "Hiking", "Biking"}
		for _, option := range options {
			if !strings.Contains(html, option) {
				t.Errorf("Missing option: %s", option)
			}
		}
	})

	t.Run("mobile viewport meta tag present", func(t *testing.T) {
		now := time.Now()
		startTime := now.Add(48 * time.Hour)

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
			},
			Questions:      []*handlers.QuestionWithOptions{},
			Token:          "test-token",
			DeadlinePassed: false,
			EventPassed:    false,
			LocalStartTime: startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"),
			CanUpdate:      false,
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, `name="viewport"`) {
			t.Error("Missing viewport meta tag for mobile")
		}

		if !strings.Contains(html, `width=device-width`) {
			t.Error("Viewport should set width=device-width")
		}
	})
}
